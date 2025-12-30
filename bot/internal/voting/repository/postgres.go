package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matperez/group-photo-reaction-bot/bot/internal/voting"
)

// PostgresRepository реализует Repository с использованием PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository создает новый PostgreSQL репозиторий.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateSession создает новую сессию голосования.
func (r *PostgresRepository) CreateSession(chatID int64, messageID int, photoFileID string, facesCount int, duration time.Duration) (*voting.Session, error) {
	ctx := context.Background()

	// Проверяем, не создана ли уже сессия для этого сообщения
	existing, exists := r.GetSessionByMessage(chatID, messageID)
	if exists && existing != nil {
		return existing, nil
	}

	// Создаем новую сессию
	sessionID := fmt.Sprintf("%d:%d:%s", chatID, messageID, photoFileID)
	now := time.Now()
	endTime := now.Add(duration)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Вставляем сессию
	_, err = tx.ExecContext(ctx, `
		INSERT INTO sessions (id, chat_id, message_id, photo_file_id, faces_count, start_time, end_time, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, sessionID, chatID, messageID, photoFileID, facesCount, now, endTime, true, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert session: %w", err)
	}

	// Обновляем или создаем запись в chats
	_, err = tx.ExecContext(ctx, `
		INSERT INTO chats (chat_id, last_active, sessions_count, created_at, updated_at)
		VALUES ($1, $2, 1, $3, $4)
		ON CONFLICT(chat_id) DO UPDATE SET
			last_active = $2,
			sessions_count = chats.sessions_count + 1,
			updated_at = $4
	`, chatID, now, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	session := voting.NewSession(sessionID, chatID, messageID, photoFileID, facesCount, duration)
	return session, nil
}

// GetSession возвращает сессию по ID.
func (r *PostgresRepository) GetSession(sessionID string) (*voting.Session, bool) {
	ctx := context.Background()

	var s struct {
		ID          string
		ChatID      int64
		MessageID   int
		PhotoFileID string
		FacesCount  int
		StartTime   time.Time
		EndTime     time.Time
		IsActive    bool
	}

	err := r.db.QueryRowContext(ctx, `
		SELECT id, chat_id, message_id, photo_file_id, faces_count, start_time, end_time, is_active
		FROM sessions
		WHERE id = $1
	`, sessionID).Scan(&s.ID, &s.ChatID, &s.MessageID, &s.PhotoFileID, &s.FacesCount, &s.StartTime, &s.EndTime, &s.IsActive)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	// Загружаем голоса
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, face_index
		FROM votes
		WHERE session_id = $1
	`, sessionID)
	if err != nil {
		return nil, false
	}
	defer rows.Close()

	votes := make(map[int64]int)
	for rows.Next() {
		var userID int64
		var faceIndex int
		if err := rows.Scan(&userID, &faceIndex); err != nil {
			return nil, false
		}
		votes[userID] = faceIndex
	}

	session := &voting.Session{
		ID:          s.ID,
		ChatID:      s.ChatID,
		MessageID:   s.MessageID,
		PhotoFileID: s.PhotoFileID,
		FacesCount:  s.FacesCount,
		StartTime:   s.StartTime,
		EndTime:     s.EndTime,
		IsActive:    s.IsActive,
	}
	session.SetVotes(votes)

	return session, true
}

// GetSessionByMessage возвращает сессию по chatID и messageID.
func (r *PostgresRepository) GetSessionByMessage(chatID int64, messageID int) (*voting.Session, bool) {
	ctx := context.Background()

	var sessionID string
	err := r.db.QueryRowContext(ctx, `
		SELECT id
		FROM sessions
		WHERE chat_id = $1 AND message_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, chatID, messageID).Scan(&sessionID)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	return r.GetSession(sessionID)
}

// CanCreateSession проверяет, можно ли создать новую сессию в чате.
func (r *PostgresRepository) CanCreateSession(chatID int64, minInterval time.Duration) bool {
	ctx := context.Background()

	var lastActive sql.NullTime
	err := r.db.QueryRowContext(ctx, `
		SELECT last_active
		FROM chats
		WHERE chat_id = $1
	`, chatID).Scan(&lastActive)
	if err == sql.ErrNoRows {
		return true
	}
	if err != nil {
		return false
	}

	if !lastActive.Valid {
		return true
	}

	return time.Since(lastActive.Time) >= minInterval
}

// DeleteSession удаляет сессию из хранилища.
func (r *PostgresRepository) DeleteSession(sessionID string) {
	ctx := context.Background()
	r.db.ExecContext(ctx, `DELETE FROM sessions WHERE id = $1`, sessionID)
}

// CleanupExpired удаляет истекшие сессии старше указанного времени.
func (r *PostgresRepository) CleanupExpired(olderThan time.Duration) {
	ctx := context.Background()
	cutoff := time.Now().Add(-olderThan)
	r.db.ExecContext(ctx, `
		DELETE FROM sessions
		WHERE is_active = false AND end_time < $1
	`, cutoff)
}

// Vote регистрирует голос в сессии.
func (r *PostgresRepository) Vote(sessionID string, userID int64, faceIndex int) error {
	ctx := context.Background()

	// Проверяем, что сессия существует и активна
	var isActive bool
	var facesCount int
	err := r.db.QueryRowContext(ctx, `
		SELECT is_active, faces_count
		FROM sessions
		WHERE id = $1 AND end_time > $2
	`, sessionID, time.Now()).Scan(&isActive, &facesCount)
	if err == sql.ErrNoRows {
		return fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return fmt.Errorf("failed to check session: %w", err)
	}

	if !isActive {
		return fmt.Errorf("session is not active")
	}

	if faceIndex < 0 || faceIndex >= facesCount {
		return fmt.Errorf("invalid face index")
	}

	// Пытаемся вставить голос (UNIQUE constraint предотвратит дубликаты)
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO votes (session_id, user_id, face_index, created_at)
		VALUES ($1, $2, $3, $4)
	`, sessionID, userID, faceIndex, time.Now())
	if err != nil {
		// Проверяем, это дубликат голоса
		if err.Error() == "pq: duplicate key value violates unique constraint \"votes_session_id_user_id_key\"" {
			return fmt.Errorf("user already voted")
		}
		return fmt.Errorf("failed to insert vote: %w", err)
	}

	return nil
}

// GetSessions возвращает список сессий с опциональными фильтрами.
func (r *PostgresRepository) GetSessions(ctx context.Context, limit, offset int, chatID *int64, isActive *bool) ([]*voting.Session, error) {
	query := `SELECT id, chat_id, message_id, photo_file_id, faces_count, start_time, end_time, is_active FROM sessions WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if chatID != nil {
		query += fmt.Sprintf(` AND chat_id = $%d`, argNum)
		args = append(args, *chatID)
		argNum++
	}

	if isActive != nil {
		query += fmt.Sprintf(` AND is_active = $%d`, argNum)
		args = append(args, *isActive)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argNum, argNum+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*voting.Session
	for rows.Next() {
		var s struct {
			ID          string
			ChatID      int64
			MessageID   int
			PhotoFileID string
			FacesCount  int
			StartTime   time.Time
			EndTime     time.Time
			IsActive    bool
		}

		if err := rows.Scan(&s.ID, &s.ChatID, &s.MessageID, &s.PhotoFileID, &s.FacesCount, &s.StartTime, &s.EndTime, &s.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Загружаем голоса
		voteRows, err := r.db.QueryContext(ctx, `SELECT user_id, face_index FROM votes WHERE session_id = $1`, s.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to query votes: %w", err)
		}

		votes := make(map[int64]int)
		for voteRows.Next() {
			var userID int64
			var faceIndex int
			if err := voteRows.Scan(&userID, &faceIndex); err != nil {
				voteRows.Close()
				return nil, fmt.Errorf("failed to scan vote: %w", err)
			}
			votes[userID] = faceIndex
		}
		voteRows.Close()

		session := &voting.Session{
			ID:          s.ID,
			ChatID:      s.ChatID,
			MessageID:   s.MessageID,
			PhotoFileID: s.PhotoFileID,
			FacesCount:  s.FacesCount,
			StartTime:   s.StartTime,
			EndTime:     s.EndTime,
			IsActive:    s.IsActive,
		}
		session.SetVotes(votes)
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetSessionsCount возвращает общее количество сессий с учетом фильтров.
func (r *PostgresRepository) GetSessionsCount(ctx context.Context, chatID *int64, isActive *bool) (int, error) {
	query := `SELECT COUNT(*) FROM sessions WHERE 1=1`
	args := []interface{}{}
	argNum := 1

	if chatID != nil {
		query += fmt.Sprintf(` AND chat_id = $%d`, argNum)
		args = append(args, *chatID)
		argNum++
	}

	if isActive != nil {
		query += fmt.Sprintf(` AND is_active = $%d`, argNum)
		args = append(args, *isActive)
		argNum++
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return count, nil
}

// GetChats возвращает список чатов.
func (r *PostgresRepository) GetChats(ctx context.Context, limit, offset int) ([]ChatInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT chat_id, last_active, sessions_count, created_at, updated_at
		FROM chats
		ORDER BY last_active DESC NULLS LAST, created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query chats: %w", err)
	}
	defer rows.Close()

	var chats []ChatInfo
	for rows.Next() {
		var c ChatInfo
		var lastActive sql.NullTime
		if err := rows.Scan(&c.ChatID, &lastActive, &c.SessionsCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}
		if lastActive.Valid {
			c.LastActive = &lastActive.Time
		}
		chats = append(chats, c)
	}

	return chats, nil
}

// GetChatInfo возвращает информацию о чате.
func (r *PostgresRepository) GetChatInfo(ctx context.Context, chatID int64) (*ChatInfo, error) {
	var c ChatInfo
	var lastActive sql.NullTime

	err := r.db.QueryRowContext(ctx, `
		SELECT chat_id, last_active, sessions_count, created_at, updated_at
		FROM chats
		WHERE chat_id = $1
	`, chatID).Scan(&c.ChatID, &lastActive, &c.SessionsCount, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chat not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query chat: %w", err)
	}

	if lastActive.Valid {
		c.LastActive = &lastActive.Time
	}

	return &c, nil
}

// GetStats возвращает общую статистику.
func (r *PostgresRepository) GetStats(ctx context.Context) (*Stats, error) {
	var stats Stats

	// Общее количество сессий
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&stats.TotalSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	// Активные сессии
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE is_active = true`).Scan(&stats.ActiveSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}

	// Общее количество голосов
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM votes`).Scan(&stats.TotalVotes)
	if err != nil {
		return nil, fmt.Errorf("failed to count votes: %w", err)
	}

	// Общее количество чатов
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chats`).Scan(&stats.TotalChats)
	if err != nil {
		return nil, fmt.Errorf("failed to count chats: %w", err)
	}

	// Среднее количество голосов на сессию
	if stats.TotalSessions > 0 {
		stats.AvgVotesPerSession = float64(stats.TotalVotes) / float64(stats.TotalSessions)
	}

	return &stats, nil
}

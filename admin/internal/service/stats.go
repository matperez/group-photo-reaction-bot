package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Stats содержит общую статистику.
type Stats struct {
	TotalSessions      int
	ActiveSessions     int
	TotalVotes         int
	TotalChats         int
	AvgVotesPerSession float64
}

// ChatInfo содержит информацию о чате.
type ChatInfo struct {
	ChatID        int64
	LastActive    *time.Time
	SessionsCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Service предоставляет методы для работы со статистикой.
type Service struct {
	db *sql.DB
}

// NewService создает новый сервис статистики.
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
	}
}

// GetStats возвращает общую статистику.
func (s *Service) GetStats(ctx context.Context) (*Stats, error) {
	var stats Stats

	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions`).Scan(&stats.TotalSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sessions WHERE is_active = true`).Scan(&stats.ActiveSessions)
	if err != nil {
		return nil, fmt.Errorf("failed to count active sessions: %w", err)
	}

	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM votes`).Scan(&stats.TotalVotes)
	if err != nil {
		return nil, fmt.Errorf("failed to count votes: %w", err)
	}

	err = s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM chats`).Scan(&stats.TotalChats)
	if err != nil {
		return nil, fmt.Errorf("failed to count chats: %w", err)
	}

	if stats.TotalSessions > 0 {
		stats.AvgVotesPerSession = float64(stats.TotalVotes) / float64(stats.TotalSessions)
	}

	return &stats, nil
}

// GetChats возвращает список чатов.
func (s *Service) GetChats(ctx context.Context, limit, offset int) ([]ChatInfo, error) {
	rows, err := s.db.QueryContext(ctx, `
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
func (s *Service) GetChatInfo(ctx context.Context, chatID int64) (*ChatInfo, error) {
	var c ChatInfo
	var lastActive sql.NullTime

	err := s.db.QueryRowContext(ctx, `
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

// Session представляет сессию голосования для админки.
type Session struct {
	ID          string
	ChatID      int64
	MessageID   int
	PhotoFileID string
	FacesCount  int
	Votes       map[int64]int
	StartTime   time.Time
	EndTime     time.Time
	IsActive    bool
}

// GetSessions возвращает список сессий.
func (s *Service) GetSessions(ctx context.Context, limit, offset int, chatID *int64, isActive *bool) ([]*Session, error) {
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

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*Session
	for rows.Next() {
		var sess struct {
			ID          string
			ChatID      int64
			MessageID   int
			PhotoFileID string
			FacesCount  int
			StartTime   time.Time
			EndTime     time.Time
			IsActive    bool
		}

		if err := rows.Scan(&sess.ID, &sess.ChatID, &sess.MessageID, &sess.PhotoFileID, &sess.FacesCount, &sess.StartTime, &sess.EndTime, &sess.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}

		// Загружаем голоса
		voteRows, err := s.db.QueryContext(ctx, `SELECT user_id, face_index FROM votes WHERE session_id = $1`, sess.ID)
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

		session := &Session{
			ID:          sess.ID,
			ChatID:      sess.ChatID,
			MessageID:   sess.MessageID,
			PhotoFileID: sess.PhotoFileID,
			FacesCount:  sess.FacesCount,
			StartTime:   sess.StartTime,
			EndTime:     sess.EndTime,
			IsActive:    sess.IsActive,
			Votes:       votes,
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// GetSession возвращает сессию по ID.
func (s *Service) GetSession(ctx context.Context, sessionID string) (*Session, bool) {
	var sess struct {
		ID          string
		ChatID      int64
		MessageID   int
		PhotoFileID string
		FacesCount  int
		StartTime   time.Time
		EndTime     time.Time
		IsActive    bool
	}

	err := s.db.QueryRowContext(ctx, `
		SELECT id, chat_id, message_id, photo_file_id, faces_count, start_time, end_time, is_active
		FROM sessions
		WHERE id = $1
	`, sessionID).Scan(&sess.ID, &sess.ChatID, &sess.MessageID, &sess.PhotoFileID, &sess.FacesCount, &sess.StartTime, &sess.EndTime, &sess.IsActive)
	if err == sql.ErrNoRows {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	// Загружаем голоса
	voteRows, err := s.db.QueryContext(ctx, `SELECT user_id, face_index FROM votes WHERE session_id = $1`, sessionID)
	if err != nil {
		return nil, false
	}
	defer voteRows.Close()

	votes := make(map[int64]int)
	for voteRows.Next() {
		var userID int64
		var faceIndex int
		if err := voteRows.Scan(&userID, &faceIndex); err != nil {
			return nil, false
		}
		votes[userID] = faceIndex
	}

	session := &Session{
		ID:          sess.ID,
		ChatID:      sess.ChatID,
		MessageID:   sess.MessageID,
		PhotoFileID: sess.PhotoFileID,
		FacesCount:  sess.FacesCount,
		StartTime:   sess.StartTime,
		EndTime:     sess.EndTime,
		IsActive:    sess.IsActive,
		Votes:       votes,
	}

	return session, true
}

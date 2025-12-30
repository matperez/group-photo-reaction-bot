package voting

import (
	"fmt"
	"sync"
	"time"
)

// StoreInterface определяет интерфейс для хранилища сессий голосований.
type StoreInterface interface {
	CreateSession(chatID int64, messageID int, photoFileID string, facesCount int, duration time.Duration) (*Session, error)
	GetSession(sessionID string) (*Session, bool)
	GetSessionByMessage(chatID int64, messageID int) (*Session, bool)
	CanCreateSession(chatID int64, minInterval time.Duration) bool
	DeleteSession(sessionID string)
	CleanupExpired(olderThan time.Duration)
}

// Store представляет in-memory хранилище сессий голосований.
type Store struct {
	sessions        map[string]*Session // sessionID -> session
	messageSessions map[int64]string    // chatID:messageID -> sessionID (для предотвращения дубликатов)
	chatLastActive  map[int64]time.Time // chatID -> last active time (для rate limiting)
	mu              sync.RWMutex
}

// Убеждаемся, что Store реализует StoreInterface
var _ StoreInterface = (*Store)(nil)

// NewStore создает новое хранилище.
func NewStore() *Store {
	return &Store{
		sessions:        make(map[string]*Session),
		messageSessions: make(map[int64]string),
		chatLastActive:  make(map[int64]time.Time),
	}
}

// CreateSession создает новую сессию голосования.
func (s *Store) CreateSession(chatID int64, messageID int, photoFileID string, facesCount int, duration time.Duration) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Проверяем, не создана ли уже сессия для этого сообщения
	key := s.messageKey(chatID, messageID)
	if existingID, exists := s.messageSessions[key]; exists {
		return s.sessions[existingID], nil
	}

	// Создаем новую сессию
	sessionID := fmt.Sprintf("%d:%d:%s", chatID, messageID, photoFileID)
	session := NewSession(sessionID, chatID, messageID, photoFileID, facesCount, duration)

	s.sessions[sessionID] = session
	s.messageSessions[key] = sessionID
	s.chatLastActive[chatID] = time.Now()

	return session, nil
}

// GetSession возвращает сессию по ID.
func (s *Store) GetSession(sessionID string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[sessionID]
	return session, exists
}

// GetSessionByMessage возвращает сессию по chatID и messageID.
func (s *Store) GetSessionByMessage(chatID int64, messageID int) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.messageKey(chatID, messageID)
	sessionID, exists := s.messageSessions[key]
	if !exists {
		return nil, false
	}

	session, exists := s.sessions[sessionID]
	return session, exists
}

// CanCreateSession проверяет, можно ли создать новую сессию в чате (rate limiting).
// Возвращает true, если прошло достаточно времени с последней сессии.
func (s *Store) CanCreateSession(chatID int64, minInterval time.Duration) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lastActive, exists := s.chatLastActive[chatID]
	if !exists {
		return true
	}

	return time.Since(lastActive) >= minInterval
}

// DeleteSession удаляет сессию из хранилища.
func (s *Store) DeleteSession(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return
	}

	// Удаляем из всех мапов
	delete(s.sessions, sessionID)
	key := s.messageKey(session.ChatID, session.MessageID)
	delete(s.messageSessions, key)
}

// CleanupExpired удаляет истекшие сессии старше указанного времени.
func (s *Store) CleanupExpired(olderThan time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for sessionID, session := range s.sessions {
		if !session.IsActive && now.Sub(session.EndTime) > olderThan {
			s.DeleteSession(sessionID)
		}
	}
}

// messageKey создает ключ для мапы сообщений.
func (s *Store) messageKey(chatID int64, messageID int) int64 {
	// Используем комбинацию chatID и messageID
	// Для простоты используем хеш-функцию
	return chatID*1000000 + int64(messageID)
}

package voting

import (
	"sync"
	"time"
)

// Session представляет сессию голосования.
type Session struct {
	ID           string
	ChatID       int64
	MessageID    int
	PhotoFileID  string
	FacesCount   int
	Votes        map[int64]int // userID -> faceIndex (0-based)
	StartTime    time.Time
	EndTime      time.Time
	IsActive     bool
	mu           sync.RWMutex
}

// NewSession создает новую сессию голосования.
func NewSession(id string, chatID int64, messageID int, photoFileID string, facesCount int, duration time.Duration) *Session {
	now := time.Now()
	return &Session{
		ID:          id,
		ChatID:      chatID,
		MessageID:   messageID,
		PhotoFileID: photoFileID,
		FacesCount:  facesCount,
		Votes:       make(map[int64]int),
		StartTime:   now,
		EndTime:     now.Add(duration),
		IsActive:    true,
	}
}

// Vote регистрирует голос пользователя.
// Возвращает true, если голос успешно зарегистрирован, false если пользователь уже голосовал.
func (s *Session) Vote(userID int64, faceIndex int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.IsActive {
		return false
	}

	if time.Now().After(s.EndTime) {
		s.IsActive = false
		return false
	}

	// Проверяем, что пользователь еще не голосовал
	if _, exists := s.Votes[userID]; exists {
		return false
	}

	// Проверяем валидность индекса лица
	if faceIndex < 0 || faceIndex >= s.FacesCount {
		return false
	}

	s.Votes[userID] = faceIndex
	return true
}

// GetResults возвращает результаты голосования.
func (s *Session) GetResults() map[int]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make(map[int]int)
	for i := 0; i < s.FacesCount; i++ {
		results[i] = 0
	}

	for _, faceIndex := range s.Votes {
		results[faceIndex]++
	}

	return results
}

// GetWinner возвращает индекс победителя и количество голосов.
// Если несколько лиц имеют одинаковое количество голосов, возвращается первое.
func (s *Session) GetWinner() (faceIndex int, votes int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := s.GetResults()
	maxVotes := 0
	winnerIndex := 0

	for idx, voteCount := range results {
		if voteCount > maxVotes {
			maxVotes = voteCount
			winnerIndex = idx
		}
	}

	return winnerIndex, maxVotes
}

// IsExpired проверяет, истекла ли сессия.
func (s *Session) IsExpired() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Now().After(s.EndTime)
}

// Deactivate деактивирует сессию.
func (s *Session) Deactivate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IsActive = false
}


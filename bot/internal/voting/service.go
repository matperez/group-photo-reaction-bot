package voting

import (
	"context"
	"fmt"
	"time"
)

// ResultCallback вызывается при завершении голосования.
type ResultCallback func(session *Session) error

// Service управляет жизненным циклом голосований.
type Service struct {
	store           StoreInterface
	resultCallback  ResultCallback
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewService создает новый сервис голосований.
func NewService(store StoreInterface, resultCallback ResultCallback) *Service {
	return &Service{
		store:           store,
		resultCallback:  resultCallback,
		cleanupInterval: 1 * time.Hour,
		stopCleanup:     make(chan struct{}),
	}
}

// Start запускает фоновые процессы сервиса (cleanup, таймеры).
func (s *Service) Start(ctx context.Context) {
	// Запускаем cleanup в фоне
	go s.cleanupLoop(ctx)

	// Обрабатываем активные сессии и запускаем таймеры
	go s.processActiveSessions(ctx)
}

// Stop останавливает фоновые процессы.
func (s *Service) Stop() {
	close(s.stopCleanup)
}

// CreateVoting создает новое голосование и запускает таймер завершения.
func (s *Service) CreateVoting(chatID int64, messageID int, photoFileID string, facesCount int, duration time.Duration) (*Session, error) {
	session, err := s.store.CreateSession(chatID, messageID, photoFileID, facesCount, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Запускаем таймер завершения
	go s.startVotingTimer(session, duration)

	return session, nil
}

// Store возвращает хранилище сессий (для обратной совместимости).
// Внимание: возвращает интерфейс, может быть nil для репозитория.
func (s *Service) Store() StoreInterface {
	return s.store
}

// SetResultCallback устанавливает callback для публикации результатов.
func (s *Service) SetResultCallback(callback ResultCallback) {
	s.resultCallback = callback
}

// Vote регистрирует голос в сессии.
func (s *Service) Vote(sessionID string, userID int64, faceIndex int) (bool, error) {
	// Проверяем, является ли store репозиторием с методом Vote
	if repo, ok := s.store.(interface {
		Vote(sessionID string, userID int64, faceIndex int) error
	}); ok {
		// Используем метод репозитория
		if err := repo.Vote(sessionID, userID, faceIndex); err != nil {
			return false, err
		}
		return true, nil
	}

	// Fallback на старую логику для in-memory store
	session, exists := s.store.GetSession(sessionID)
	if !exists {
		return false, fmt.Errorf("session not found: %s", sessionID)
	}

	success := session.Vote(userID, faceIndex)
	return success, nil
}

// startVotingTimer запускает таймер для завершения голосования.
func (s *Service) startVotingTimer(session *Session, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	<-timer.C

	// Деактивируем сессию
	session.Deactivate()

	// Вызываем callback для публикации результатов
	if s.resultCallback != nil {
		if err := s.resultCallback(session); err != nil {
			// Логируем ошибку, но не паникуем
			// В реальном приложении здесь должен быть логгер
			_ = err
		}
	}
}

// cleanupLoop периодически очищает истекшие сессии.
func (s *Service) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCleanup:
			return
		case <-ticker.C:
			s.store.CleanupExpired(24 * time.Hour)
		}
	}
}

// processActiveSessions обрабатывает активные сессии и проверяет их статус.
func (s *Service) processActiveSessions(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.checkExpiredSessions()
		}
	}
}

// checkExpiredSessions проверяет и завершает истекшие сессии.
func (s *Service) checkExpiredSessions() {
	// Получаем все активные сессии и проверяем их
	// В реальной реализации здесь нужен метод для получения всех активных сессий
	// Для упрощения MVP оставляем основную логику в таймерах
}

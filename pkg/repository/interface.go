package repository

import (
	"context"
	"time"

	"github.com/matperez/group-photo-reaction-bot/bot/internal/voting"
)

// Repository определяет интерфейс для работы с хранилищем голосований.
// Расширяет интерфейс voting.StoreInterface для работы с БД.
type Repository interface {
	voting.StoreInterface

	// GetSessions возвращает список сессий с опциональными фильтрами.
	GetSessions(ctx context.Context, limit, offset int, chatID *int64, isActive *bool) ([]*voting.Session, error)

	// GetSessionsCount возвращает общее количество сессий с учетом фильтров.
	GetSessionsCount(ctx context.Context, chatID *int64, isActive *bool) (int, error)

	// GetChats возвращает список чатов.
	GetChats(ctx context.Context, limit, offset int) ([]ChatInfo, error)

	// GetChatInfo возвращает информацию о чате.
	GetChatInfo(ctx context.Context, chatID int64) (*ChatInfo, error)

	// GetStats возвращает общую статистику.
	GetStats(ctx context.Context) (*Stats, error)
}

// ChatInfo содержит информацию о чате.
type ChatInfo struct {
	ChatID       int64
	LastActive   *time.Time
	SessionsCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Stats содержит общую статистику.
type Stats struct {
	TotalSessions  int
	ActiveSessions int
	TotalVotes     int
	TotalChats     int
	AvgVotesPerSession float64
}


package auth

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Service предоставляет методы для авторизации.
type Service struct {
	db     *sql.DB
	secret string
}

// NewService создает новый сервис авторизации.
func NewService(db *sql.DB, secret string) *Service {
	return &Service{
		db:     db,
		secret: secret,
	}
}

// Login проверяет логин и пароль, возвращает JWT токен.
func (s *Service) Login(username, password string) (string, int, error) {
	var userID int
	var passwordHash string

	err := s.db.QueryRow(`
		SELECT id, password_hash
		FROM admin_users
		WHERE username = $1
	`, username).Scan(&userID, &passwordHash)
	if err == sql.ErrNoRows {
		return "", 0, fmt.Errorf("invalid credentials")
	}
	if err != nil {
		return "", 0, fmt.Errorf("failed to query user: %w", err)
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return "", 0, fmt.Errorf("invalid credentials")
	}

	// Генерируем токен
	token, err := GenerateToken(userID, username, s.secret)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, userID, nil
}

// EnsureAdminUser создает администратора, если его еще нет.
func (s *Service) EnsureAdminUser(username, password string) error {
	// Проверяем, существует ли пользователь
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM admin_users WHERE username = $1`, username).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check user: %w", err)
	}

	if count > 0 {
		return nil // Пользователь уже существует
	}

	// Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Создаем пользователя
	_, err = s.db.Exec(`
		INSERT INTO admin_users (username, password_hash, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, username, string(hash))
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

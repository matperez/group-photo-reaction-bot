package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

// Open открывает соединение с PostgreSQL БД и применяет миграции.
func Open(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настраиваем goose для использования PostgreSQL
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	// Применяем миграции
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Migrate применяет все миграции с помощью goose.
func Migrate(db *sql.DB) error {
	// Применяем миграции из папки migrations в корне проекта
	migrationsDir := "../../migrations"
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}

// Close закрывает соединение с БД.
func Close(db *sql.DB) error {
	if db == nil {
		return nil
	}
	return db.Close()
}

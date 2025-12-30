package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed migrations
var migrationsFS embed.FS

// Open открывает соединение с SQLite БД и применяет миграции.
func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=1&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Проверяем соединение
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Применяем миграции
	if err := Migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Migrate применяет все миграции из папки migrations.
func Migrate(db *sql.DB) error {
	// Читаем файлы миграций
	migrations, err := fs.ReadDir(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Сортируем по имени (001, 002, ...)
	for _, entry := range migrations {
		if entry.IsDir() {
			continue
		}

		migrationPath := entry.Name()
		migrationSQL, err := migrationsFS.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		// Выполняем миграцию в транзакции
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(migrationSQL)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", entry.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", entry.Name(), err)
		}
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


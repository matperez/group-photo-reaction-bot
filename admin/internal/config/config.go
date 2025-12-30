package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config содержит конфигурацию админки.
type Config struct {
	Port      int
	Username  string
	Password  string
	JWTSecret string
	DBPath    string
}

// Load загружает конфигурацию из переменных окружения.
func Load() (*Config, error) {
	portStr := os.Getenv("ADMIN_PORT")
	port := 8080
	if portStr != "" {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ADMIN_PORT: %w", err)
		}
		port = p
	}

	username := os.Getenv("ADMIN_USERNAME")
	if username == "" {
		username = "admin"
	}

	password := os.Getenv("ADMIN_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("ADMIN_PASSWORD environment variable is required")
	}

	jwtSecret := os.Getenv("ADMIN_JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("ADMIN_JWT_SECRET environment variable is required")
	}

	dbPath := os.Getenv("ADMIN_DB_PATH")
	if dbPath == "" {
		dbPath = "bot.db"
	}

	return &Config{
		Port:      port,
		Username:  username,
		Password:  password,
		JWTSecret: jwtSecret,
		DBPath:    dbPath,
	}, nil
}

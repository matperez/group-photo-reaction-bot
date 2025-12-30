package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config содержит конфигурацию приложения.
type Config struct {
	BotToken       string
	MinFaces       int
	MaxFaces       int
	VotingText     string
	VotingDuration time.Duration
}

// Load загружает конфигурацию из переменных окружения и файла .env.
func Load() (*Config, error) {
	// Загружаем .env файл, если он существует (игнорируем ошибку)
	_ = godotenv.Load()

	cfg := &Config{
		MinFaces:       4,
		MaxFaces:       6,
		VotingText:     "кто тут у нас самый улыбчивый",
		VotingDuration: 5 * time.Minute,
	}

	// Загружаем токен бота (обязательный параметр)
	cfg.BotToken = os.Getenv("BOT_TOKEN")
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN environment variable is required")
	}

	// Загружаем минимальное количество лиц
	if minFacesStr := os.Getenv("MIN_FACES"); minFacesStr != "" {
		minFaces, err := strconv.Atoi(minFacesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid MIN_FACES: %w", err)
		}
		cfg.MinFaces = minFaces
	}

	// Загружаем максимальное количество лиц
	if maxFacesStr := os.Getenv("MAX_FACES"); maxFacesStr != "" {
		maxFaces, err := strconv.Atoi(maxFacesStr)
		if err != nil {
			return nil, fmt.Errorf("invalid MAX_FACES: %w", err)
		}
		cfg.MaxFaces = maxFaces
	}

	// Загружаем текст голосования
	if votingText := os.Getenv("VOTING_TEXT"); votingText != "" {
		cfg.VotingText = votingText
	}

	// Загружаем длительность голосования
	if votingDurationStr := os.Getenv("VOTING_DURATION"); votingDurationStr != "" {
		votingDuration, err := time.ParseDuration(votingDurationStr)
		if err != nil {
			return nil, fmt.Errorf("invalid VOTING_DURATION: %w", err)
		}
		cfg.VotingDuration = votingDuration
	}

	// Валидация конфигурации
	if cfg.MinFaces < 1 {
		return nil, fmt.Errorf("MIN_FACES must be at least 1")
	}
	if cfg.MaxFaces < cfg.MinFaces {
		return nil, fmt.Errorf("MAX_FACES must be >= MIN_FACES")
	}

	return cfg, nil
}


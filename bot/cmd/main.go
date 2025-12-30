package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/matperez/group-photo-reaction-bot/bot/internal/bot"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/config"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/db"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/face"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/voting"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/voting/repository"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Инициализируем детектор лиц
	// Каскадный классификатор встроен в бинарник, внешние файлы не требуются
	detector, err := face.NewLocalDetector()
	if err != nil {
		log.Fatalf("Failed to create face detector: %v", err)
	}

	// Подключаемся к БД
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "botuser"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "botpass"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "botdb"
	}
	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		dbSSLMode = "disable"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	database, err := db.Open(connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(database); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Создаем PostgreSQL репозиторий
	store := repository.NewPostgresRepository(database)

	// Создаем временный callback (будет заменен после создания bot service)
	var botService *bot.Service
	tempCallback := func(session *voting.Session) error {
		if botService != nil {
			return botService.PublishResults(session)
		}
		log.Printf("Results callback called before bot service initialization for session %s", session.ID)
		return nil
	}

	// Создаем сервис голосований
	votingService := voting.NewService(store, tempCallback)

	// Создаем сервис бота
	botService, err = bot.NewService(cfg, detector, votingService)
	if err != nil {
		log.Fatalf("Failed to create bot service: %v", err)
	}

	// Обновляем callback в voting service для использования bot service
	resultCallback := func(session *voting.Session) error {
		return botService.PublishResults(session)
	}
	votingService.SetResultCallback(resultCallback)

	// Запускаем фоновые процессы сервиса голосований
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	votingService.Start(ctx)

	// Обрабатываем сигналы для graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Запускаем бота в отдельной горутине
	go func() {
		botService.Start()
	}()

	// Ждем сигнала завершения
	<-sigChan
	log.Println("Shutting down...")

	// Останавливаем сервисы
	votingService.Stop()
	botService.Stop()

	log.Println("Bot stopped")
}

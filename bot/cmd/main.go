package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/matperez/group-photo-reaction-bot/bot/internal/bot"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/config"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/face"
	"github.com/matperez/group-photo-reaction-bot/bot/internal/voting"
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

	// Создаем хранилище голосований
	store := voting.NewStore()

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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/matperez/group-photo-reaction-bot/admin/internal/api"
	"github.com/matperez/group-photo-reaction-bot/admin/internal/auth"
	"github.com/matperez/group-photo-reaction-bot/admin/internal/config"
	"github.com/matperez/group-photo-reaction-bot/admin/internal/service"
	"github.com/pressly/goose/v3"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключаемся к PostgreSQL БД
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

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Применяем миграции
	if err := migrate(database); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Создаем сервис авторизации
	authService := auth.NewService(database, cfg.JWTSecret)

	// Создаем администратора, если его еще нет
	if err := authService.EnsureAdminUser(cfg.Username, cfg.Password); err != nil {
		log.Fatalf("Failed to ensure admin user: %v", err)
	}

	// Создаем сервис статистики
	statsService := service.NewService(database)

	// Создаем handlers
	authHandler := api.NewAuthHandler(authService)
	sessionsHandler := api.NewSessionsHandler(statsService)
	statsHandler := api.NewStatsHandler(statsService)
	chatsHandler := api.NewChatsHandler(statsService)

	// Настраиваем роутинг
	mux := http.NewServeMux()

	// Публичные endpoints
	mux.HandleFunc("/api/auth/login", authHandler.Login)

	// Защищенные endpoints
	mux.HandleFunc("/api/sessions", sessionsHandler.List)
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		sessionsHandler.Get(w, r)
	})
	mux.HandleFunc("/api/stats", statsHandler.Get)
	mux.HandleFunc("/api/chats", chatsHandler.List)
	mux.HandleFunc("/api/chats/", func(w http.ResponseWriter, r *http.Request) {
		chatsHandler.Get(w, r)
	})

	// Применяем middleware
	handler := auth.CORSMiddleware(auth.Middleware(cfg.JWTSecret)(mux))

	// Создаем HTTP сервер
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: handler,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Admin API server starting on port %d", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Ожидаем сигнал завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("Server stopped")
}

// migrate применяет миграции с помощью goose.
func migrate(db *sql.DB) error {
	// Настраиваем goose для использования PostgreSQL
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Применяем миграции из папки migrations в корне проекта
	migrationsDir := "../../migrations"
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

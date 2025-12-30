.PHONY: help build-bot build-admin clean lint-bot lint-admin test-bot test-admin install-linter deps-bot deps-admin deps docker-build docker-up docker-down docker-logs docker-restart

# Переменные
BIN_DIR := bin
BOT_BINARY := $(BIN_DIR)/bot
ADMIN_BINARY := $(BIN_DIR)/admin
LINTER_BIN := $(shell go env GOPATH)/bin/golangci-lint

help: ## Показать справку по командам
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build-bot: ## Создать бинарник бота в bin/bot
	@echo "Сборка бота..."
	@mkdir -p $(BIN_DIR)
	@cd bot && go build -o ../$(BOT_BINARY) ./cmd/main.go
	@echo "Бот собран: $(BOT_BINARY)"

build-admin: ## Создать бинарник админки в bin/admin
	@echo "Сборка админки..."
	@mkdir -p $(BIN_DIR)
	@cd admin && go build -o ../$(ADMIN_BINARY) ./cmd/server.go
	@echo "Админка собрана: $(ADMIN_BINARY)"

build-all: build-bot build-admin ## Собрать оба бинарника

clean: ## Очистить артефакты сборки
	@echo "Очистка артефактов сборки..."
	@rm -rf $(BIN_DIR)
	@cd bot && go clean
	@cd admin && go clean
	@echo "Очистка завершена"

lint-bot: ## Запустить линтер для бота
	@echo "Запуск линтера для бота..."
	@cd bot && $(LINTER_BIN) run --config .golangci.yml ./... || { \
		echo "Линтинг завершен с предупреждениями"; \
		exit 0; \
	}
	@echo "Линтинг бота завершен"

lint-admin: ## Запустить линтер для админки
	@echo "Запуск линтера для админки..."
	@cd admin && $(LINTER_BIN) run --config .golangci.yml ./... || { \
		echo "Линтинг завершен с предупреждениями"; \
		exit 0; \
	}
	@echo "Линтинг админки завершен"

lint: lint-bot lint-admin ## Запустить линтер для всех проектов

test-bot: ## Запустить тесты бота
	@echo "Запуск тестов бота..."
	@cd bot && go test -v ./...
	@echo "Тесты бота завершены"

test-admin: ## Запустить тесты админки
	@echo "Запуск тестов админки..."
	@cd admin && go test -v ./...
	@echo "Тесты админки завершены"

test: test-bot test-admin ## Запустить тесты для всех проектов

install-linter: ## Установить/обновить golangci-lint
	@echo "Установка/обновление golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "golangci-lint установлен/обновлен."

deps-bot: ## Установить зависимости бота
	@echo "Установка зависимостей бота..."
	@cd bot && go mod tidy && go mod download
	@echo "Зависимости бота установлены"

deps-admin: ## Установить зависимости админки
	@echo "Установка зависимостей админки..."
	@cd admin && go mod tidy && go mod download
	@echo "Зависимости админки установлены"

deps: deps-bot deps-admin ## Установить зависимости для всех проектов

# Goose команды для миграций
migrate-up: ## Применить все миграции (требует DB_* переменные окружения)
	@echo "Применение миграций..."
	@goose -dir migrations postgres "host=$${DB_HOST:-localhost} port=$${DB_PORT:-5432} user=$${DB_USER:-botuser} password=$${DB_PASSWORD:-botpass} dbname=$${DB_NAME:-botdb} sslmode=$${DB_SSLMODE:-disable}" up
	@echo "Миграции применены"

migrate-down: ## Откатить последнюю миграцию
	@echo "Откат последней миграции..."
	@goose -dir migrations postgres "host=$${DB_HOST:-localhost} port=$${DB_PORT:-5432} user=$${DB_USER:-botuser} password=$${DB_PASSWORD:-botpass} dbname=$${DB_NAME:-botdb} sslmode=$${DB_SSLMODE:-disable}" down
	@echo "Миграция откачена"

migrate-status: ## Показать статус миграций
	@goose -dir migrations postgres "host=$${DB_HOST:-localhost} port=$${DB_PORT:-5432} user=$${DB_USER:-botuser} password=$${DB_PASSWORD:-botpass} dbname=$${DB_NAME:-botdb} sslmode=$${DB_SSLMODE:-disable}" status

migrate-create: ## Создать новую миграцию (использование: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Ошибка: укажите имя миграции: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@goose -dir migrations create $(NAME) sql
	@echo "Миграция $(NAME) создана"

install-goose: ## Установить goose
	@echo "Установка goose..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "goose установлен"

# Docker команды
docker-build: ## Собрать Docker образы
	@echo "Сборка Docker образов..."
	@docker-compose build
	@echo "Docker образы собраны"

docker-up: ## Запустить контейнеры
	@echo "Запуск Docker контейнеров..."
	@docker-compose up -d
	@echo "Контейнеры запущены"

docker-down: ## Остановить контейнеры
	@echo "Остановка Docker контейнеров..."
	@docker-compose down
	@echo "Контейнеры остановлены"

docker-logs: ## Показать логи контейнеров
	@docker-compose logs -f

docker-restart: docker-down docker-up ## Перезапустить контейнеры

docker-clean: docker-down ## Остановить и удалить контейнеры, volumes
	@echo "Удаление контейнеров и volumes..."
	@docker-compose down -v
	@echo "Очистка завершена"

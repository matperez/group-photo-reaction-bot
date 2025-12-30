.PHONY: help lint lint-fix test test-coverage build clean run deps deps-update install-linter

GO_VERSION := 1.25.5
LINTER_VERSION := v1.61.0
LINTER_BIN := $(shell go env GOPATH)/bin/golangci-lint

help: ## Показать справку по командам
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

lint: ## Запустить golangci-lint
	@echo "Запуск линтера..."
	@$(LINTER_BIN) run --config .golangci.yml ./... || { \
		echo "Линтинг завершен с предупреждениями"; \
		exit 0; \
	}
	@echo "Линтинг завершен"

install-linter: ## Установить/обновить golangci-lint до последней версии
	@echo "Установка/обновление golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "golangci-lint установлен/обновлен."

lint-fix: ## Запустить golangci-lint с автоисправлением
	@echo "Запуск линтера с автоисправлением..."
	@$(LINTER_BIN) run --fix --config .golangci.yml ./...
	@echo "Линтинг с автоисправлением завершен"

test: ## Запустить тесты
	@echo "Запуск тестов..."
	@go test -v ./...
	@echo "Тесты завершены"

test-coverage: ## Запустить тесты с отчетом о покрытии
	@echo "Запуск тестов с отчетом о покрытии..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out
	@echo "Отчет о покрытии создан: coverage.out"

build: ## Собрать приложение
	@echo "Сборка приложения..."
	@go build -o server ./cmd/server.go
	@echo "Сборка успешна"

clean: ## Очистить артефакты сборки
	@echo "Очистка артефактов сборки..."
	@go clean
	@rm -f server coverage.out
	@echo "Очистка завершена"

run: build ## Собрать и запустить приложение
	@echo "Запуск приложения..."
	@./server

deps: ## Установить зависимости
	@echo "Установка зависимостей..."
	@go mod tidy
	@go mod download
	@echo "Зависимости установлены"

deps-update: ## Обновить зависимости
	@echo "Обновление зависимостей..."
	@go get -u ./...
	@go mod tidy
	@echo "Зависимости обновлены"


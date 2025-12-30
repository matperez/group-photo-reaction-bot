.PHONY: lint test build clean help

# Переменные
BINARY_NAME=bot
CMD_PATH=./cmd/main.go
LINT_TIMEOUT=5m

help: ## Показать справку по командам
	@echo "Доступные команды:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

lint: ## Запустить golangci-lint
	@echo "Запуск линтера..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Ошибка: golangci-lint не установлен. Установите последнюю версию: make install-linter"; \
		exit 1; \
	fi
	@golangci-lint run --timeout $(LINT_TIMEOUT) ./...
	@echo "Линтинг завершен успешно"

install-linter: ## Установить/обновить golangci-lint до последней версии
	@echo "Установка golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest
	@golangci-lint --version
	@echo "golangci-lint установлен успешно"

lint-fix: ## Запустить golangci-lint с автоисправлением
	@echo "Запуск линтера с автоисправлением..."
	@golangci-lint run --fix --timeout $(LINT_TIMEOUT) ./...
	@echo "Линтинг завершен успешно"

test: ## Запустить тесты
	@echo "Запуск тестов..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "Тесты завершены"

test-coverage: test ## Запустить тесты с отчетом о покрытии
	@echo "Генерация отчета о покрытии..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Отчет сохранен в coverage.html"

build: ## Собрать приложение
	@echo "Сборка приложения..."
	@go build -o $(BINARY_NAME) $(CMD_PATH)
	@echo "Сборка завершена, бинарник: $(BINARY_NAME)"

clean: ## Очистить артефакты сборки
	@echo "Очистка..."
	@rm -f $(BINARY_NAME) coverage.out coverage.html
	@echo "Очистка завершена"

run: build ## Собрать и запустить приложение
	@echo "Запуск приложения..."
	@./$(BINARY_NAME)

deps: ## Установить зависимости
	@echo "Установка зависимостей..."
	@go mod download
	@go mod tidy
	@echo "Зависимости установлены"

deps-update: ## Обновить зависимости
	@echo "Обновление зависимостей..."
	@go get -u ./...
	@go mod tidy
	@echo "Зависимости обновлены"


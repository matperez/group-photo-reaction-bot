# Telegram Bot для голосования по групповым фото

MVP Telegram-бот для автоматического обнаружения групповых фото и создания голосований.

## Функционал

- Автоматическое обнаружение групповых фото (4-6 лиц, настраиваемо)
- Визуальное выделение лиц на фото
- Предложение создания голосования
- Сбор голосов в течение заданного времени (по умолчанию 5 минут)
- Публикация результатов голосования

## Требования

- **Go 1.25.5** (последняя стабильная версия)
- golangci-lint последней версии (для линтинга) - см. установку ниже
- Файл каскадного классификатора для pigo (facefinder)

### Установка golangci-lint

Для работы с Go 1.25.5 требуется golangci-lint последней версии (1.62.0 или выше):

```bash
# Автоматическая установка через Makefile
make install-linter

# Или вручную
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest

# Или через Homebrew (macOS)
brew install golangci-lint
brew upgrade golangci-lint
```

## Установка

1. Клонируйте репозиторий
2. Установите зависимости:
```bash
make deps
```

3. Скачайте каскадный классификатор для pigo:
```bash
# Скачайте facefinder из репозитория pigo
# https://github.com/esimov/pigo/tree/master/cascade
# И поместите в каталог cascade/
```

## Конфигурация

Настройки задаются через переменные окружения:

- `BOT_TOKEN` - токен Telegram бота (обязательно)
- `CASCADE_PATH` - путь к файлу каскадного классификатора (по умолчанию: `cascade/facefinder`)
- `MIN_FACES` - минимальное количество лиц (по умолчанию: 4)
- `MAX_FACES` - максимальное количество лиц (по умолчанию: 6)
- `VOTING_TEXT` - текст голосования (по умолчанию: "кто тут у нас самый улыбчивый")
- `VOTING_DURATION` - длительность голосования (по умолчанию: 5m)

Пример `.env` файла:
```
BOT_TOKEN=your_bot_token_here
CASCADE_PATH=cascade/facefinder
MIN_FACES=4
MAX_FACES=6
VOTING_TEXT=кто тут у нас самый улыбчивый
VOTING_DURATION=5m
```

## Использование

### Запуск бота

```bash
make run
```

или

```bash
go run cmd/main.go
```

### Линтинг

```bash
make lint
```

**Примечание:** typecheck может выдавать предупреждения из-за несовместимости версий Go между golangci-lint и проектом. Это не критично, остальные линтеры работают корректно.

### Тесты

```bash
make test
```

### Сборка

```bash
make build
```

## Доступные команды Make

- `make help` - показать справку по командам
- `make lint` - запустить линтер
- `make lint-fix` - запустить линтер с автоисправлением
- `make test` - запустить тесты
- `make test-coverage` - запустить тесты с отчетом о покрытии
- `make build` - собрать приложение
- `make run` - собрать и запустить приложение
- `make clean` - очистить артефакты сборки
- `make deps` - установить зависимости
- `make deps-update` - обновить зависимости

## Структура проекта

```
bot/
├── cmd/
│   └── main.go              # Точка входа
├── internal/
│   ├── bot/                 # Обработчики Telegram
│   ├── config/              # Конфигурация
│   ├── face/                # Детекция и разметка лиц
│   └── voting/              # Логика голосований
├── pkg/
│   └── image/               # Утилиты для работы с изображениями
├── .golangci.yml            # Конфигурация линтера
├── Makefile                 # Команды для разработки
└── go.mod                   # Зависимости
```

## Разработка

Проект использует:
- [telebot](https://github.com/tucnak/telebot) - библиотека для работы с Telegram Bot API
- [pigo](https://github.com/esimov/pigo) - библиотека для детекции лиц
- [golangci-lint](https://golangci-lint.run/) - линтер для Go

## Лицензия

MIT


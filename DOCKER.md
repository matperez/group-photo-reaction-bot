# Запуск проекта в Docker

## Быстрый старт

1. **Создайте файл `.env` в корне проекта:**
```bash
cp .env.example .env
# Отредактируйте .env и укажите свои значения
```

2. **Соберите и запустите контейнеры:**
```bash
make docker-build
make docker-up
```

Или одной командой:
```bash
docker-compose up -d --build
```

3. **Проверьте статус:**
```bash
docker-compose ps
```

4. **Просмотр логов:**
```bash
make docker-logs
# или для конкретного сервиса:
docker-compose logs -f bot
docker-compose logs -f admin
```

## Доступ к сервисам

- **Админка API:** http://localhost:8080
- **Бот:** работает в фоне, обрабатывает сообщения из Telegram

## Команды управления

```bash
# Остановить контейнеры
make docker-down

# Перезапустить
make docker-restart

# Остановить и удалить volumes (БД будет удалена!)
make docker-clean
```

## Переменные окружения

Все переменные настраиваются через файл `.env`:

- `BOT_TOKEN` - токен Telegram бота (обязательно)
- `MIN_FACES` / `MAX_FACES` - диапазон количества лиц
- `VOTING_TEXT` - текст голосования
- `VOTING_DURATION` - длительность голосования
- `ADMIN_PORT` - порт админки (по умолчанию 8080)
- `ADMIN_USERNAME` - логин администратора
- `ADMIN_PASSWORD` - пароль администратора (обязательно)
- `ADMIN_JWT_SECRET` - секретный ключ для JWT (обязательно)

## Персистентность данных

База данных SQLite хранится в Docker volume `db_data`. Данные сохраняются между перезапусками контейнеров.

Для полного удаления данных:
```bash
make docker-clean
```

## Разработка

Для разработки с hot-reload можно использовать `docker-compose.dev.yml` (требует дополнительной настройки).

## Troubleshooting

### Проблемы с правами доступа к БД
Если возникают проблемы с доступом к БД, проверьте права на volume:
```bash
docker-compose exec bot ls -la /data
```

### Проверка здоровья сервисов
```bash
docker-compose ps
# Должны быть статусы "healthy" или "Up"
```

### Пересборка после изменений
```bash
make docker-build
make docker-restart
```


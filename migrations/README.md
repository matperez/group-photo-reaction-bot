# Миграции базы данных

Этот проект использует [goose](https://github.com/pressly/goose) для управления миграциями базы данных PostgreSQL.

## Структура папок

```
migrations/
├── README.md                          # Этот файл
├── 20241230160000_init.up.sql         # Применение: создание начальной схемы
└── 20241230160000_init.down.sql       # Откат: удаление схемы
```

Миграции находятся в папке `migrations/` в **корне проекта**. Каждая миграция состоит из двух файлов:

- `YYYYMMDDHHMMSS_description.up.sql` - миграция для применения (создание/изменение схемы)
- `YYYYMMDDHHMMSS_description.down.sql` - миграция для отката (откат изменений)

**Важно:** Миграции общие для бота и админки, так как оба сервиса работают с одной БД.

## Установка goose

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Использование

### Автоматическое применение

Миграции применяются **автоматически** при запуске:
- Бота (`bot/cmd/main.go`)
- Админки (`admin/cmd/server.go`)

Оба сервиса используют функцию `Migrate()` из `bot/internal/db/postgres.go`, которая вызывает `goose.Up()`.

### Ручное применение миграций

Для ручного управления миграциями используйте команды Makefile из корня проекта:

```bash
# Применить все миграции
make migrate-up

# Откатить последнюю миграцию
make migrate-down

# Показать статус миграций
make migrate-status

# Создать новую миграцию
make migrate-create NAME=add_new_table
```

Или напрямую через goose:

```bash
# Из корня проекта
goose -dir migrations postgres "host=localhost port=5432 user=botuser password=botpass dbname=botdb sslmode=disable" up
```

### Откат миграций

```bash
# Откатить последнюю миграцию
goose -dir migrations postgres "host=localhost port=5432 user=botuser password=botpass dbname=botdb sslmode=disable" down

# Откатить все миграции
goose -dir migrations postgres "host=localhost port=5432 user=botuser password=botpass dbname=botdb sslmode=disable" down-to 0
```

### Проверка статуса

```bash
goose -dir migrations postgres "host=localhost port=5432 user=botuser password=botpass dbname=botdb sslmode=disable" status
```

### Создание новой миграции

```bash
# Создать новую миграцию
goose -dir migrations create migration_name sql

# Это создаст два файла:
# - YYYYMMDDHHMMSS_migration_name.up.sql
# - YYYYMMDDHHMMSS_migration_name.down.sql
```

## Формат миграций

### Up миграция (применение)

```sql
-- Пример: создание таблицы
CREATE TABLE IF NOT EXISTS example (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Down миграция (откат)

```sql
-- Пример: удаление таблицы
DROP TABLE IF EXISTS example;
```

## Важные замечания

1. **Порядок миграций**: Миграции применяются в порядке их временных меток (YYYYMMDDHHMMSS)
2. **Идемпотентность**: Используйте `IF NOT EXISTS` и `IF EXISTS` для безопасных миграций
3. **Транзакции**: Goose автоматически оборачивает каждую миграцию в транзакцию
4. **Откат**: Всегда создавайте down-миграции для возможности отката

## Текущие миграции

- `20241230160000_init` - Начальная схема БД (sessions, votes, chats, admin_users, admin_sessions)

## Интеграция в код

Миграции применяются автоматически при:
- Запуске бота (`bot/cmd/main.go`)
- Запуске админки (`admin/cmd/server.go`)

Оба сервиса используют функцию `Migrate()` из `bot/internal/db/postgres.go`, которая вызывает `goose.Up()`.


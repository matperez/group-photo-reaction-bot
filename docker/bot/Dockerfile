# Многостадийная сборка для бота
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# Копируем go.mod и go.sum для кэширования зависимостей
COPY bot/go.mod bot/go.sum ./
RUN go mod download

# Копируем исходный код (миграции уже включены)
COPY bot/ ./

# Собираем бинарник
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /build/bot .

# Создаем директорию для БД
RUN mkdir -p /data && chmod 777 /data

EXPOSE 8080

CMD ["./bot"]


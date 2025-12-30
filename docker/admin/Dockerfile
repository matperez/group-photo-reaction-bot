# Многостадийная сборка для админки
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# Копируем go.mod и go.sum для кэширования зависимостей
COPY admin/go.mod admin/go.sum ./
RUN go mod download

# Копируем исходный код
COPY admin/ ./

# Собираем бинарник
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o admin ./cmd/server.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Копируем бинарник из builder
COPY --from=builder /build/admin .

# Создаем директорию для БД
RUN mkdir -p /data && chmod 777 /data

EXPOSE 8080

CMD ["./admin"]


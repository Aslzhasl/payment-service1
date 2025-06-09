# Stage 1: Build the Go binary
FROM golang:1.24 as builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o payment-service .

# Stage 2: Минимальный образ для запуска
FROM alpine:latest

WORKDIR /app

# Копируем бинарник и .env
COPY --from=builder /app/payment-service .
COPY .env .

# Экспонируем порт
EXPOSE 8080

# Запускаем приложение
ENTRYPOINT ["./payment-service"]
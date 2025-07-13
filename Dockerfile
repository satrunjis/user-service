# Этап сборки
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# 1. Копируем только файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 2. Копируем ТОЛЬКО файлы, необходимые для генерации Swagger
# Укажите здесь все директории/файлы с аннотациями
COPY ./internal/handler ./internal/handler
COPY ./internal/domain ./internal/domain
COPY ./cmd/main.go ./cmd/main.go

# 3. Генерируем Swagger (слой будет пересобираться при изменении указанных файлов)
RUN swag init -g cmd/main.go --output ./docs

# 4. Копируем остальной код
COPY . .

# 5. Собираем приложение
RUN go build -o user-service ./cmd/main.go

# Финальный образ
FROM alpine:latest
WORKDIR /app
RUN mkdir cmd
COPY --from=builder /app/user-service ./cmd/
COPY --from=builder /app/docs ./docs
RUN if [ -f config.yaml ]; then cp config.yaml . ; fi
COPY docker-compose.yaml .
EXPOSE 8080
CMD ["./cmd/user-service"]
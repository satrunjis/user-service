FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY ./internal/handler ./internal/handler
COPY ./internal/domain ./internal/domain
COPY ./cmd/main.go ./cmd/main.go

RUN swag init -g cmd/main.go --output ./docs

COPY . .

RUN go build -o user-service ./cmd/main.go

FROM alpine:3.22.1
RUN adduser -S appuser

WORKDIR /app

COPY --from=builder --chown=appuser:appuser /app/user-service ./cmd/
COPY --from=builder --chown=appuser:appuser /app/docs ./docs

RUN if [ -f config.yaml ]; then cp config.yaml . ; fi && \
    chown appuser:appuser config.yaml 2>/dev/null || true

COPY --chown=appuser:appuser docker-compose.yaml .

RUN chmod 550 ./cmd/user-service

USER appuser

USER appuser
EXPOSE 8080
CMD ["./cmd/user-service"]
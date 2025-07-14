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

FROM alpine:latest
WORKDIR /app
RUN mkdir cmd
COPY --from=builder /app/user-service ./cmd/
COPY --from=builder /app/docs ./docs
RUN if [ -f config.yaml ]; then cp config.yaml . ; fi
COPY docker-compose.yaml .
EXPOSE 8080
CMD ["./cmd/user-service"]
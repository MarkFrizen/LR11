# Этап 1: Зависимости (кэшируется при неизменном go.mod)
FROM golang:1.22-alpine AS deps
WORKDIR /app
COPY go.mod ./
RUN go mod download 2>/dev/null || true

# Этап 2: Сборка
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY --from=deps /go /go
COPY . .
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o /main .

# Этап 3: Минимальный образ
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /main .
EXPOSE 8080
CMD ["./main"]

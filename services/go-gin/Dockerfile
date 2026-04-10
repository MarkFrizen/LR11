# Этап 1: Зависимости (кэшируется при неизменном go.mod)
FROM golang:1.26-alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Этап 2: Сборка
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY --from=deps /go /go
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

# Этап 3: Минимальный образ
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /main .
EXPOSE 8080
CMD ["./main"]

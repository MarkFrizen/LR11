# Этап 1: Сборка
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /main .

# Этап 2: Минимальный образ
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /main .

EXPOSE 8080

CMD ["./main"]

# Этап 1: Сборка
FROM rust:1.88-slim AS builder

WORKDIR /app

# Копируем весь код и собираем
COPY Cargo.toml ./
COPY src ./src
RUN cargo fetch && cargo build --release

# Этап 2: Минимальный образ
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends libssl3 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/target/release/rust-actix .

EXPOSE 8082

CMD ["./rust-actix"]

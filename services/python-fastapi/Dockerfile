# Этап 1: Компиляция Rust-крейта
FROM rust:1.88-slim AS rust-builder
WORKDIR /app/rust_extension
COPY rust_extension/Cargo.toml ./
COPY rust_extension/src ./src
RUN cargo build --release

# Этап 2: Сборка Python-колеса через maturin
FROM python:3.12-slim AS wheel-builder
WORKDIR /app/rust_extension

RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc && \
    rm -rf /var/lib/apt/lists/* && \
    pip install --no-cache-dir maturin

COPY rust_extension/Cargo.toml ./
COPY rust_extension/src ./src
COPY rust_extension/pyproject.toml ./
# Копируем скомпилированные артефакты из rust-builder (кэш крейтов)
COPY --from=rust-builder /app/rust_extension/target ./target

RUN maturin build --release --out dist --manylinux off

# Этап 3: Минимальный runtime
FROM python:3.12-slim AS runtime
WORKDIR /app

# Копируем только готовое колесо и устанавливаем
COPY --from=wheel-builder /app/rust_extension/dist ./dist
RUN pip install --no-cache-dir dist/*.whl

# Копируем Python-приложение
COPY python_app/main.py .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]

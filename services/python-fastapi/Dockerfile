# Этап 1: Сборка Python-колеса с Rust-расширением через maturin
FROM python:3.12-slim AS wheel-builder
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc curl libc6-dev pkg-config && \
    rm -rf /var/lib/apt/lists/* && \
    curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y && \
    pip install --no-cache-dir maturin

ENV PATH="/root/.cargo/bin:${PATH}"

COPY rust_extension/Cargo.toml ./
COPY rust_extension/src ./src
COPY rust_extension/pyproject.toml ./

RUN maturin build --release --out dist --manylinux off

# Этап 2: Минимальный runtime
FROM python:3.12-slim AS runtime
WORKDIR /app

# Устанавливаем колесо
COPY --from=wheel-builder /app/dist ./dist
RUN pip install --no-cache-dir dist/*.whl

# Устанавливаем зависимости приложения
RUN pip install --no-cache-dir fastapi uvicorn

# Копируем Python-приложение
COPY python_app/ ./python_app/

EXPOSE 8000

CMD ["uvicorn", "python_app.main:app", "--host", "0.0.0.0", "--port", "8000"]

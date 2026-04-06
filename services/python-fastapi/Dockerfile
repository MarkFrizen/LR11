# Этап 1: Установка зависимостей (билдер)
FROM python:3.12-slim AS builder

WORKDIR /app

# Build-зависимости для компиляции C-расширений (не попадут в финальный образ)
RUN apt-get update && \
    apt-get install -y --no-install-recommends gcc && \
    rm -rf /var/lib/apt/lists/*

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Этап 2: Минимальный runtime (без gcc и build-зависимостей)
FROM python:3.12-slim AS runtime

WORKDIR /app

# Копируем установленные пакеты из билдера
COPY --from=builder /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY --from=builder /usr/local/bin /usr/local/bin

# Копируем код приложения
COPY . .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]

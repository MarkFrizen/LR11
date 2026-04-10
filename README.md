# Лабораторная работа №11
### Студент: Фризен Марк Владимирович
### Группа: 221331
### Вариант 5
Задания варианта:
Средней сложности:
5.Создать docker-compose.yml, поднимающий Python, Go и Rust сервисы
7.Использовать volume для обмена данными между контейнерами.
9.Ограничить ресурсы (CPU, память) для контейнеров.
Повышенной сложности:
5.Оптимизировать слои Docker-образов для максимального кэширования.
7.Создать multi-stage сборку для Python-приложения с Rust-расширением.

# LR11

## Запуск сервисов

### Все сервисы через Docker Compose
```bash
docker compose up --build
```

### Запуск тестов

#### Go (Gin)
```bash
cd services/go-gin
go test ./... -v
```

#### Rust (Actix)
```bash
cd services/rust-actix
cargo test
```

#### Python (FastAPI)
```bash
cd services/python-fastapi
pip install -r requirements-test.txt
pytest test_main.py -v
```

## Структура проекта
```
LR11/
├── docker-compose.yml          # Конфигурация Docker Compose
├── PROMPT_LOG.md               # Логи промптов для Qwen Code
├── README.md                   # Этот файл
└── services/
    ├── go-gin/                 # Go сервис на Gin (порт 8080)
    │   ├── main.go             # Точка входа, маршрутизация
    │   ├── handlers.go         # HTTP обработчики, Handler struct
    │   ├── handlers_test.go    # Тесты (unit + mock)
    │   ├── storage.go          # Интерфейсы LogStorage, TimeFormatter
    │   ├── file_storage.go     # FileLogStorage реализация
    │   ├── go.mod
    │   ├── go.sum
    │   └── Dockerfile
    ├── python-fastapi/         # Python сервис на FastAPI (порт 8000)
    │   ├── python_app/
    │   │   ├── main.py         # Точка входа, lifespan, DI
    │   │   ├── routes.py       # Роуты через create_routes()
    │   │   ├── storage.py      # Абстрактный класс LogStorage
    │   │   └── file_storage.py # FileLogStorage реализация
    │   ├── rust_extension/     # Rust-расширение (pyo3 + maturin)
    │   │   ├── Cargo.toml
    │   │   ├── pyproject.toml
    │   │   └── src/lib.rs
    │   ├── test_main.py        # Тесты (unit + mock)
    │   ├── requirements-test.txt
    │   └── Dockerfile
    └── rust-actix/             # Rust сервис на Actix-web (порт 8082)
        ├── src/
        │   ├── main.rs         # Точка входа, AppState, маршруты
        │   ├── storage.rs      # Трейт LogStorage
        │   ├── file_storage.rs # FileLogStorage реализация
        │   └── storage_test.rs # Тесты (unit + mock)
        ├── Cargo.toml
        └── Dockerfile
```

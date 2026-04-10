import os
from contextlib import asynccontextmanager

from fastapi import FastAPI

from .storage import LogStorage
from .file_storage import FileLogStorage
from .routes import setup_routes

LOG_FILE = os.environ.get("LOG_FILE", "/data/log.txt")

# Создаём storage сразу — для тестов и прода
_storage = FileLogStorage(LOG_FILE)


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Переиспользуем уже созданный storage
    app.state.log_storage = _storage
    yield


app = FastAPI(title="Python FastAPI + Rust Extension", lifespan=lifespan)

# Регистрируем роуты с общим storage
setup_routes(app, _storage)

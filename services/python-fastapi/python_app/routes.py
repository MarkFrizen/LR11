import os
from datetime import datetime, timezone

from fastapi import FastAPI
from fastapi.responses import JSONResponse

from rust_ext import fibonacci, rust_hello

from .storage import LogStorage


def create_routes(storage: LogStorage):
    """Создаёт роуты с инъек зависимостей (DIP)"""

    async def root():
        return {"service": "python-fastapi", "status": "running"}

    async def health():
        return {"status": "healthy"}

    async def hello_rust(name: str = "World"):
        return {"message": rust_hello(name), "source": "rust_extension"}

    async def calc_fibonacci(n: int):
        if n < 0:
            return JSONResponse(status_code=400, content={"error": "n must be >= 0"})
        return {"n": n, "fibonacci": fibonacci(n)}

    async def write_time():
        try:
            now = datetime.now(timezone.utc).isoformat()
            source = os.environ.get("HOSTNAME", "unknown")
            line = f"[{source}] {now}\n"
            storage.write_line(line)
            return {"message": "time written", "time": now}
        except Exception as e:
            return JSONResponse(status_code=500, content={"error": str(e)})

    async def read_log():
        try:
            content = storage.read_all()
            return {"content": content}
        except Exception as e:
            return JSONResponse(status_code=500, content={"error": str(e)})

    return root, health, hello_rust, calc_fibonacci, write_time, read_log


def setup_routes(app: FastAPI, storage: LogStorage):
    """Регистрирует роуты в приложении"""
    root, health, hello_rust, calc_fibonacci, write_time, read_log = create_routes(storage)

    app.get("/")(root)
    app.get("/health")(health)
    app.get("/rust_hello")(hello_rust)
    app.get("/fibonacci/{n}")(calc_fibonacci)
    app.get("/write")(write_time)
    app.get("/read")(read_log)

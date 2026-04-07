import fcntl
import os
from datetime import datetime, timezone

from fastapi import FastAPI
from fastapi.responses import JSONResponse

from rust_ext import fibonacci, rust_hello

app = FastAPI(title="Python FastAPI + Rust Extension")

LOG_FILE = "/data/log.txt"


@app.get("/")
async def root():
    return {"service": "python-fastapi", "status": "running"}


@app.get("/health")
async def health():
    return {"status": "healthy"}


@app.get("/rust_hello")
async def hello_rust(name: str = "World"):
    """Проверка вызова Rust-функции из Python"""
    return {"message": rust_hello(name), "source": "rust_extension"}


@app.get("/fibonacci/{n}")
async def calc_fibonacci(n: int):
    """Вычисление Фибоначчи через Rust-расширение (pyo3)"""
    if n < 0:
        return JSONResponse(status_code=400, content={"error": "n must be >= 0"})
    return {"n": n, "fibonacci": fibonacci(n)}


@app.get("/write")
async def write_time():
    try:
        now = datetime.now(timezone.utc).isoformat()
        source = os.environ.get("HOSTNAME", "unknown")
        with open(LOG_FILE, "a") as f:
            fcntl.flock(f, fcntl.LOCK_EX)
            try:
                f.write(f"[{source}] {now}\n")
            finally:
                fcntl.flock(f, fcntl.LOCK_UN)
        return {"message": "time written", "time": now}
    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})


@app.get("/read")
async def read_log():
    try:
        if not os.path.exists(LOG_FILE):
            return {"content": ""}
        with open(LOG_FILE, "r") as f:
            fcntl.flock(f, fcntl.LOCK_SH)
            try:
                return {"content": f.read()}
            finally:
                fcntl.flock(f, fcntl.LOCK_UN)
    except Exception as e:
        return JSONResponse(status_code=500, content={"error": str(e)})

import fcntl
import os
from datetime import datetime, timezone

from fastapi import FastAPI
from fastapi.responses import JSONResponse

app = FastAPI(title="Python FastAPI Service")

LOG_FILE = "/data/log.txt"


@app.get("/")
async def root():
    return {"service": "python-fastapi", "status": "running"}


@app.get("/health")
async def health():
    return {"status": "healthy"}


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

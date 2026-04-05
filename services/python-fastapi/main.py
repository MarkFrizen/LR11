from fastapi import FastAPI

app = FastAPI(title="Python FastAPI Service")


@app.get("/")
async def root():
    return {"service": "python-fastapi", "status": "running"}


@app.get("/health")
async def health():
    return {"status": "healthy"}

"""Файловая реализация LogStorage (SRP — только файловые операции)"""
import fcntl
import os
from pathlib import Path

from .storage import LogStorage


class FileLogStorage(LogStorage):
    """Хранилище логов в файле с блокировками"""

    def __init__(self, filepath: str):
        self._filepath = Path(filepath)

    def write_line(self, line: str) -> None:
        with open(self._filepath, "a") as f:
            fcntl.flock(f, fcntl.LOCK_EX)
            try:
                f.write(line)
            finally:
                fcntl.flock(f, fcntl.LOCK_UN)

    def read_all(self) -> str:
        if not self._filepath.exists():
            return ""

        with open(self._filepath, "r") as f:
            fcntl.flock(f, fcntl.LOCK_SH)
            try:
                return f.read()
            finally:
                fcntl.flock(f, fcntl.LOCK_UN)

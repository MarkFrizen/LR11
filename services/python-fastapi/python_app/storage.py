"""Абстракции для работы с хранилищем логов (SRP, DIP)"""
from abc import ABC, abstractmethod


class LogStorage(ABC):
    """Интерфейс для чтения/записи логов (Dependency Inversion)"""

    @abstractmethod
    def write_line(self, line: str) -> None:
        """Записать строку в лог"""
        ...

    @abstractmethod
    def read_all(self) -> str:
        """Прочитать всё содержимое лога"""
        ...

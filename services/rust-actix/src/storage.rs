use std::io;

/// Интерфейс для работы с хранилищем логов (SRP, DIP)
pub trait LogStorage: Send + Sync {
    fn write_line(&self, line: &str) -> Result<(), io::Error>;
    fn read_all(&self) -> Result<String, io::Error>;
}

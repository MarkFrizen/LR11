use crate::file_storage::FileLogStorage;
use crate::storage::LogStorage;
use std::fs;
use std::io;
use std::sync::{Arc, Mutex};

fn create_test_storage(path: &str) -> FileLogStorage {
    let _ = fs::remove_file(path);
    FileLogStorage::new(path)
}

fn cleanup(path: &str) {
    let _ = fs::remove_file(path);
}

#[test]
fn test_file_storage_write_and_read() {
    let path = "/tmp/test_storage_basic.log";
    let storage = create_test_storage(path);

    storage.write_line("[test] line1\n").unwrap();
    storage.write_line("[test] line2\n").unwrap();

    let content = storage.read_all().unwrap();
    assert!(content.contains("[test] line1\n"));
    assert!(content.contains("[test] line2\n"));

    cleanup(path);
}

#[test]
fn test_file_storage_read_nonexistent() {
    let storage = FileLogStorage::new("/tmp/nonexistent_file_12345.log");
    let result = storage.read_all();
    assert!(result.is_ok());
    assert_eq!(result.unwrap(), "");
}

#[test]
fn test_file_storage_append_mode() {
    let path = "/tmp/test_storage_append.log";
    let _ = fs::remove_file(path);

    let storage = create_test_storage(path);

    storage.write_line("first\n").unwrap();
    let content1 = storage.read_all().unwrap();
    assert_eq!(content1, "first\n");

    storage.write_line("second\n").unwrap();
    let content2 = storage.read_all().unwrap();
    assert_eq!(content2, "first\nsecond\n");

    cleanup(path);
}

#[test]
fn test_file_storage_empty_write() {
    let path = "/tmp/test_storage_empty.log";
    let storage = create_test_storage(path);

    storage.write_line("").unwrap();
    let content = storage.read_all().unwrap();
    assert_eq!(content, "");

    cleanup(path);
}

// Mock-реализация LogStorage для тестирования endpoint-ов
struct MockLogStorage {
    write_error: Arc<Mutex<Option<String>>>,
    read_content: Arc<Mutex<Option<String>>>,
    read_error: Arc<Mutex<Option<String>>>,
    written_lines: Arc<Mutex<Vec<String>>>,
}

impl MockLogStorage {
    fn new() -> Self {
        Self {
            write_error: Arc::new(Mutex::new(None)),
            read_content: Arc::new(Mutex::new(None)),
            read_error: Arc::new(Mutex::new(None)),
            written_lines: Arc::new(Mutex::new(Vec::new())),
        }
    }

    fn with_write_error(error: &str) -> Self {
        let s = Self::new();
        *s.write_error.lock().unwrap() = Some(error.to_string());
        s
    }

    fn with_read_content(content: &str) -> Self {
        let s = Self::new();
        *s.read_content.lock().unwrap() = Some(content.to_string());
        s
    }

    fn with_read_error(error: &str) -> Self {
        let s = Self::new();
        *s.read_error.lock().unwrap() = Some(error.to_string());
        s
    }
}

impl LogStorage for MockLogStorage {
    fn write_line(&self, line: &str) -> Result<(), io::Error> {
        if let Some(err) = self.write_error.lock().unwrap().as_ref() {
            return Err(io::Error::new(io::ErrorKind::Other, err.clone()));
        }
        let mut lines = self.written_lines.lock().unwrap();
        lines.push(line.to_string());
        Ok(())
    }

    fn read_all(&self) -> Result<String, io::Error> {
        if let Some(err) = self.read_error.lock().unwrap().as_ref() {
            return Err(io::Error::new(io::ErrorKind::Other, err.clone()));
        }
        if let Some(content) = self.read_content.lock().unwrap().as_ref() {
            return Ok(content.clone());
        }
        Ok(String::new())
    }
}

#[test]
fn test_mock_storage_write_success() {
    let mock = MockLogStorage::new();
    mock.write_line("[test] data\n").unwrap();

    let lines = mock.written_lines.lock().unwrap();
    assert_eq!(lines.len(), 1);
    assert_eq!(lines[0], "[test] data\n");
}

#[test]
fn test_mock_storage_write_error() {
    let mock = MockLogStorage::with_write_error("disk full");

    let result = mock.write_line("[test] data\n");
    assert!(result.is_err());
    assert!(result.unwrap_err().to_string().contains("disk full"));
}

#[test]
fn test_mock_storage_read_success() {
    let mock = MockLogStorage::with_read_content("[host] 2026-04-09T12:00:00Z\n");

    let content = mock.read_all().unwrap();
    assert_eq!(content, "[host] 2026-04-09T12:00:00Z\n");
}

#[test]
fn test_mock_storage_read_error() {
    let mock = MockLogStorage::with_read_error("permission denied");

    let result = mock.read_all();
    assert!(result.is_err());
    assert!(result.unwrap_err().to_string().contains("permission denied"));
}

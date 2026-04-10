import os
import tempfile
from unittest.mock import patch, MagicMock

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

# Mock rust_ext module
import sys
sys.modules['rust_ext'] = MagicMock()

from python_app.routes import create_routes
from python_app.file_storage import FileLogStorage
from python_app.storage import LogStorage


def create_test_app(storage: LogStorage) -> FastAPI:
    """Создаёт FastAPI приложение с инъек зависимостей"""
    app = FastAPI(title="Test App")
    root, health, hello_rust, calc_fibonacci, write_time, read_log = create_routes(storage)

    app.get("/")(root)
    app.get("/health")(health)
    app.get("/rust_hello")(hello_rust)
    app.get("/fibonacci/{n}")(calc_fibonacci)
    app.get("/write")(write_time)
    app.get("/read")(read_log)

    return app


@pytest.fixture
def temp_log_file():
    """Создаёт временный файл лога"""
    with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.log') as f:
        temp_path = f.name
    yield temp_path
    if os.path.exists(temp_path):
        os.unlink(temp_path)


@pytest.fixture
def file_storage_client(temp_log_file):
    """Клиент с реальным файловым хранилищем"""
    storage = FileLogStorage(temp_log_file)
    app = create_test_app(storage)
    with TestClient(app) as c:
        yield c, storage


@pytest.fixture
def mock_storage_client():
    """Клиент с моком хранилища"""
    mock_storage = MagicMock(spec=LogStorage)
    app = create_test_app(mock_storage)
    with TestClient(app) as c:
        yield c, mock_storage


class TestRootEndpoint:
    def test_root_returns_json(self, file_storage_client):
        client, _ = file_storage_client
        response = client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["service"] == "python-fastapi"
        assert data["status"] == "running"


class TestHealthEndpoint:
    def test_health_returns_healthy(self, file_storage_client):
        client, _ = file_storage_client
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"


class TestRustHelloEndpoint:
    def test_rust_hello_default(self, file_storage_client):
        client, _ = file_storage_client
        with patch('python_app.routes.rust_hello') as mock_hello:
            mock_hello.return_value = "Hello from Rust, World!"
            response = client.get("/rust_hello")
            assert response.status_code == 200
            data = response.json()
            assert "message" in data
            assert data["source"] == "rust_extension"
            mock_hello.assert_called_once_with("World")

    def test_rust_hello_with_name(self, file_storage_client):
        client, _ = file_storage_client
        with patch('python_app.routes.rust_hello') as mock_hello:
            mock_hello.return_value = "Hello from Rust, Mark!"
            response = client.get("/rust_hello?name=Mark")
            assert response.status_code == 200
            data = response.json()
            assert "message" in data
            mock_hello.assert_called_once_with("Mark")


class TestFibonacciEndpoint:
    def test_fibonacci_valid(self, file_storage_client):
        client, _ = file_storage_client
        with patch('python_app.routes.fibonacci') as mock_fib:
            mock_fib.return_value = 55
            response = client.get("/fibonacci/10")
            assert response.status_code == 200
            data = response.json()
            assert data["n"] == 10
            assert data["fibonacci"] == 55
            mock_fib.assert_called_once_with(10)

    def test_fibonacci_zero(self, file_storage_client):
        client, _ = file_storage_client
        with patch('python_app.routes.fibonacci') as mock_fib:
            mock_fib.return_value = 0
            response = client.get("/fibonacci/0")
            assert response.status_code == 200
            data = response.json()
            assert data["n"] == 0
            assert data["fibonacci"] == 0

    def test_fibonacci_negative(self, file_storage_client):
        client, _ = file_storage_client
        response = client.get("/fibonacci/-5")
        assert response.status_code == 400
        data = response.json()
        assert "error" in data


class TestWriteAndReadEndpoint:
    def test_write_time_creates_file(self, file_storage_client, temp_log_file):
        client, _ = file_storage_client
        response = client.get("/write")
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "time written"
        assert "time" in data

    def test_read_empty_when_no_file(self, file_storage_client, temp_log_file):
        client, storage = file_storage_client
        os.unlink(temp_log_file)
        response = client.get("/read")
        assert response.status_code == 200
        data = response.json()
        assert data["content"] == ""

    def test_read_existing_content(self, file_storage_client, temp_log_file):
        client, storage = file_storage_client
        storage.write_line("[test] 2026-04-09T12:00:00+00:00\n")

        response = client.get("/read")
        assert response.status_code == 200
        data = response.json()
        assert "[test]" in data["content"]
        assert "2026-04-09T12:00:00+00:00" in data["content"]

    def test_write_and_read_integration(self, file_storage_client, temp_log_file):
        client, _ = file_storage_client
        write_response = client.get("/write")
        assert write_response.status_code == 200

        read_response = client.get("/read")
        assert read_response.status_code == 200
        data = read_response.json()
        assert len(data["content"]) > 0


class TestFileLogStorage:
    """Тесты для файлового хранилища (SRP)"""

    def test_write_and_read(self, temp_log_file):
        storage = FileLogStorage(temp_log_file)

        storage.write_line("[test] line1\n")
        storage.write_line("[test] line2\n")

        content = storage.read_all()
        assert "[test] line1\n" in content
        assert "[test] line2\n" in content

    def test_read_nonexistent_file(self):
        storage = FileLogStorage("/nonexistent/path/file.log")
        assert storage.read_all() == ""


class TestWriteReadWithMockStorage:
    """Тесты endpoint-ов с моком storage (DIP проверка)"""

    def test_write_uses_storage(self, mock_storage_client):
        client, mock_storage = mock_storage_client
        mock_storage.write_line.return_value = None

        response = client.get("/write")
        assert response.status_code == 200

        mock_storage.write_line.assert_called_once()
        call_args = mock_storage.write_line.call_args[0][0]
        assert "[" in call_args
        assert "T" in call_args

    def test_write_storage_error(self, mock_storage_client):
        client, mock_storage = mock_storage_client
        mock_storage.write_line.side_effect = IOError("disk full")

        response = client.get("/write")
        assert response.status_code == 500
        assert "disk full" in response.json()["error"]

    def test_read_uses_storage(self, mock_storage_client):
        client, mock_storage = mock_storage_client
        mock_storage.read_all.return_value = "[host] 2026-04-09T12:00:00Z\n"

        response = client.get("/read")
        assert response.status_code == 200
        data = response.json()
        assert "[host]" in data["content"]

        mock_storage.read_all.assert_called_once()

    def test_read_storage_error(self, mock_storage_client):
        client, mock_storage = mock_storage_client
        mock_storage.read_all.side_effect = IOError("permission denied")

        response = client.get("/read")
        assert response.status_code == 500
        assert "permission denied" in response.json()["error"]

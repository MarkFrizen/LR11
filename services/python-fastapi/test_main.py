import os
import sys
import tempfile
from datetime import datetime, timezone
from unittest.mock import patch, MagicMock

import pytest
from fastapi.testclient import TestClient

# Mock rust_ext module since we can't build it without maturin
sys.modules['rust_ext'] = MagicMock()

# Now import the app
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'python_app'))
from main import app


@pytest.fixture
def client():
    return TestClient(app)


@pytest.fixture
def temp_log_file():
    with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.log') as f:
        temp_path = f.name
    
    import main as app_main
    app_main.LOG_FILE = temp_path
    yield temp_path
    
    if os.path.exists(temp_path):
        os.unlink(temp_path)


class TestRootEndpoint:
    def test_root_returns_json(self, client):
        response = client.get("/")
        assert response.status_code == 200
        data = response.json()
        assert data["service"] == "python-fastapi"
        assert data["status"] == "running"


class TestHealthEndpoint:
    def test_health_returns_healthy(self, client):
        response = client.get("/health")
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"


class TestRustHelloEndpoint:
    def test_rust_hello_default(self, client):
        with patch('main.rust_hello') as mock_hello:
            mock_hello.return_value = "Hello from Rust, World!"
            response = client.get("/rust_hello")
            assert response.status_code == 200
            data = response.json()
            assert "message" in data
            assert data["source"] == "rust_extension"
            mock_hello.assert_called_once_with("World")

    def test_rust_hello_with_name(self, client):
        with patch('main.rust_hello') as mock_hello:
            mock_hello.return_value = "Hello from Rust, Mark!"
            response = client.get("/rust_hello?name=Mark")
            assert response.status_code == 200
            data = response.json()
            assert "message" in data
            mock_hello.assert_called_once_with("Mark")


class TestFibonacciEndpoint:
    def test_fibonacci_valid(self, client):
        with patch('main.fibonacci') as mock_fib:
            mock_fib.return_value = 55
            response = client.get("/fibonacci/10")
            assert response.status_code == 200
            data = response.json()
            assert data["n"] == 10
            assert data["fibonacci"] == 55
            mock_fib.assert_called_once_with(10)

    def test_fibonacci_zero(self, client):
        with patch('main.fibonacci') as mock_fib:
            mock_fib.return_value = 0
            response = client.get("/fibonacci/0")
            assert response.status_code == 200
            data = response.json()
            assert data["n"] == 0
            assert data["fibonacci"] == 0

    def test_fibonacci_negative(self, client):
        response = client.get("/fibonacci/-5")
        assert response.status_code == 400
        data = response.json()
        assert "error" in data


class TestWriteAndReadEndpoint:
    def test_write_time_creates_file(self, client, temp_log_file):
        response = client.get("/write")
        assert response.status_code == 200
        data = response.json()
        assert data["message"] == "time written"
        assert "time" in data
        
        # Проверяем что файл создан
        assert os.path.exists(temp_log_file)

    def test_read_empty_when_no_file(self, client, temp_log_file):
        os.unlink(temp_log_file)
        response = client.get("/read")
        assert response.status_code == 200
        data = response.json()
        assert data["content"] == ""

    def test_read_existing_content(self, client, temp_log_file):
        with open(temp_log_file, 'w') as f:
            f.write("[test] 2026-04-09T12:00:00+00:00\n")
        
        response = client.get("/read")
        assert response.status_code == 200
        data = response.json()
        assert "[test]" in data["content"]
        assert "2026-04-09T12:00:00+00:00" in data["content"]

    def test_write_and_read_integration(self, client, temp_log_file):
        # Пишем данные
        write_response = client.get("/write")
        assert write_response.status_code == 200
        
        # Читаем и проверяем что данные есть
        read_response = client.get("/read")
        assert read_response.status_code == 200
        data = read_response.json()
        assert len(data["content"]) > 0

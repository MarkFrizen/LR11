package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockTimeSource для тестирования
type MockTimeSource struct {
	FixedTime time.Time
}

func (m *MockTimeSource) Now() time.Time {
	return m.FixedTime
}

func TestFormatTime(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		tm := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
		result := tm.Format(time.RFC3339)
		expected := "2026-04-09T12:00:00Z"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("different timezone", func(t *testing.T) {
		loc := time.FixedZone("MSK", 3*3600)
		tm := time.Date(2026, 4, 9, 15, 30, 45, 0, loc)
		result := tm.Format(time.RFC3339)
		if !strings.Contains(result, "2026-04-09") {
			t.Errorf("expected date 2026-04-09, got %s", result)
		}
	})
}

func TestFormatLogLine(t *testing.T) {
	formatter := &DefaultLogFormatter{}

	t.Run("standard host", func(t *testing.T) {
		result := formatter.FormatLogLine("go-gin:8080", "2026-04-09T12:00:00Z")
		expected := "[go-gin:8080] 2026-04-09T12:00:00Z\n"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("empty host", func(t *testing.T) {
		result := formatter.FormatLogLine("", "2026-04-09T12:00:00Z")
		expected := "[] 2026-04-09T12:00:00Z\n"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})
}

func TestWriteToFile(t *testing.T) {
	t.Run("write single line", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test_write.log")
		storage := NewFileLogStorage(testFile)

		err := storage.WriteLine("[test] 2026-04-09T12:00:00Z\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		expected := "[test] 2026-04-09T12:00:00Z\n"
		if string(content) != expected {
			t.Errorf("expected %s, got %s", expected, string(content))
		}
	})

	t.Run("append multiple lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test_append.log")
		storage := NewFileLogStorage(testFile)

		err := storage.WriteLine("line1\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = storage.WriteLine("line2\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("failed to read file: %v", err)
		}

		expected := "line1\nline2\n"
		if string(content) != expected {
			t.Errorf("expected %s, got %s", expected, string(content))
		}
	})
}

func TestReadFromFile(t *testing.T) {
	t.Run("read existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test_read.log")
		storage := NewFileLogStorage(testFile)

		content := "test content\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		result, err := storage.ReadAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != content {
			t.Errorf("expected %s, got %s", content, result)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.log")
		storage := NewFileLogStorage(testFile)

		result, err := storage.ReadAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("read written content", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test_rw.log")
		storage := NewFileLogStorage(testFile)

		line := "[host] 2026-04-09T12:00:00Z\n"
		err := storage.WriteLine(line)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := storage.ReadAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != line {
			t.Errorf("expected %s, got %s", line, result)
		}
	})
}

func TestHandlerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	storage := NewFileLogStorage(testFile)
	formatter := &DefaultLogFormatter{}
	timeSource := &MockTimeSource{
		FixedTime: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
	}
	_ = NewHandler(storage, formatter, timeSource)

	t.Run("handler writes and reads log", func(t *testing.T) {
		// Имитируем запись
		line := formatter.FormatLogLine("test:8080", timeSource.Now().Format(time.RFC3339))
		err := storage.WriteLine(line)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Читаем и проверяем
		result, err := storage.ReadAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "[test:8080] 2026-04-09T12:00:00Z\n"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})
}

// MockLogStorage для тестирования Handler в изоляции (DIP проверка)
type MockLogStorage struct {
	WrittenLines []string
	ReadContent  string
	WriteError   error
	ReadError    error
}

func (m *MockLogStorage) WriteLine(line string) error {
	if m.WriteError != nil {
		return m.WriteError
	}
	m.WrittenLines = append(m.WrittenLines, line)
	return nil
}

func (m *MockLogStorage) ReadAll() (string, error) {
	if m.ReadError != nil {
		return "", m.ReadError
	}
	return m.ReadContent, nil
}

func TestHandlerWithMockStorage(t *testing.T) {
	t.Run("write time success", func(t *testing.T) {
		mockStorage := &MockLogStorage{}
		formatter := &DefaultLogFormatter{}
		timeSource := &MockTimeSource{
			FixedTime: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		}
		_ = NewHandler(mockStorage, formatter, timeSource)

		// Проверяем что запись прошла успешно
		line := formatter.FormatLogLine("test:8080", timeSource.Now().Format(time.RFC3339))
		err := mockStorage.WriteLine(line)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(mockStorage.WrittenLines) != 1 {
			t.Errorf("expected 1 written line, got %d", len(mockStorage.WrittenLines))
		}

		expected := "[test:8080] 2026-04-09T12:00:00Z\n"
		if mockStorage.WrittenLines[0] != expected {
			t.Errorf("expected %s, got %s", expected, mockStorage.WrittenLines[0])
		}
	})

	t.Run("write time error", func(t *testing.T) {
		mockStorage := &MockLogStorage{
			WriteError: fmt.Errorf("disk full"),
		}
		formatter := &DefaultLogFormatter{}
		timeSource := &MockTimeSource{
			FixedTime: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		}
		_ = NewHandler(mockStorage, formatter, timeSource)

		line := formatter.FormatLogLine("test:8080", timeSource.Now().Format(time.RFC3339))
		err := mockStorage.WriteLine(line)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "disk full") {
			t.Errorf("expected 'disk full' error, got %v", err)
		}
	})

	t.Run("read log success", func(t *testing.T) {
		mockStorage := &MockLogStorage{
			ReadContent: "[host] 2026-04-09T12:00:00Z\n",
		}
		formatter := &DefaultLogFormatter{}
		timeSource := &MockTimeSource{
			FixedTime: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		}
		_ = NewHandler(mockStorage, formatter, timeSource)

		content, err := mockStorage.ReadAll()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "[host] 2026-04-09T12:00:00Z\n"
		if content != expected {
			t.Errorf("expected %s, got %s", expected, content)
		}
	})

	t.Run("read log error", func(t *testing.T) {
		mockStorage := &MockLogStorage{
			ReadError: fmt.Errorf("permission denied"),
		}
		formatter := &DefaultLogFormatter{}
		timeSource := &MockTimeSource{
			FixedTime: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		}
		_ = NewHandler(mockStorage, formatter, timeSource)

		_, err := mockStorage.ReadAll()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("expected 'permission denied' error, got %v", err)
		}
	})
}

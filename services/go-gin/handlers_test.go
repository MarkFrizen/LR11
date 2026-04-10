package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	t.Run("valid time", func(t *testing.T) {
		tm := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
		result := formatTime(tm)
		expected := "2026-04-09T12:00:00Z"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("different timezone", func(t *testing.T) {
		loc := time.FixedZone("MSK", 3*3600)
		tm := time.Date(2026, 4, 9, 15, 30, 45, 0, loc)
		result := formatTime(tm)
		if !strings.Contains(result, "2026-04-09") {
			t.Errorf("expected date 2026-04-09, got %s", result)
		}
	})
}

func TestFormatLogLine(t *testing.T) {
	t.Run("standard host", func(t *testing.T) {
		result := formatLogLine("go-gin:8080", "2026-04-09T12:00:00Z")
		expected := "[go-gin:8080] 2026-04-09T12:00:00Z\n"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})

	t.Run("empty host", func(t *testing.T) {
		result := formatLogLine("", "2026-04-09T12:00:00Z")
		expected := "[] 2026-04-09T12:00:00Z\n"
		if result != expected {
			t.Errorf("expected %s, got %s", expected, result)
		}
	})
}

func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	t.Run("write single line", func(t *testing.T) {
		err := writeToFile(testFile, "[test] 2026-04-09T12:00:00Z\n")
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
		err := writeToFile(testFile, "line1\n")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		err = writeToFile(testFile, "line2\n")
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
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.log")

	t.Run("read existing file", func(t *testing.T) {
		content := "test content\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		result, err := readFromFile(testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != content {
			t.Errorf("expected %s, got %s", content, result)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		result, err := readFromFile(filepath.Join(tmpDir, "nonexistent.log"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("read written content", func(t *testing.T) {
		line := "[host] 2026-04-09T12:00:00Z\n"
		err := writeToFile(testFile, line)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result, err := readFromFile(testFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != line {
			t.Errorf("expected %s, got %s", line, result)
		}
	})
}

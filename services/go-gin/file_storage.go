package main

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

// FileLogStorage реализует LogStorage с использованием файловой системы
type FileLogStorage struct {
	filepath string
}

// NewFileLogStorage создаёт новое хранилище логов в файле
func NewFileLogStorage(filepath string) *FileLogStorage {
	return &FileLogStorage{filepath: filepath}
}

func (s *FileLogStorage) WriteLine(line string) error {
	f, err := os.OpenFile(s.filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock exclusive: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	if _, err := fmt.Fprint(f, line); err != nil {
		return fmt.Errorf("write line: %w", err)
	}

	return nil
}

func (s *FileLogStorage) ReadAll() (string, error) {
	f, err := os.Open(s.filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return "", fmt.Errorf("flock shared: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, err := io.ReadAll(f)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read file: %w", err)
	}

	return string(data), nil
}

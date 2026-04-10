package main

import (
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
)

const logFile = "/data/log.txt"

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

func formatLogLine(host, timestamp string) string {
	return fmt.Sprintf("[%s] %s\n", host, timestamp)
}

func writeToFile(filepath string, line string) error {
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("flock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	if _, err := fmt.Fprint(f, line); err != nil {
		return err
	}

	return nil
}

func readFromFile(filepath string) (string, error) {
	f, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		return "", fmt.Errorf("flock: %w", err)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	data, err := io.ReadAll(f)
	if err != nil && err != io.EOF {
		return "", err
	}

	return string(data), nil
}

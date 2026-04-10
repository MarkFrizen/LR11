package main

import (
	"fmt"
	"time"
)

// LogFormatter форматирует лог-запись
type LogFormatter interface {
	FormatLogLine(host, timestamp string) string
}

// DefaultLogFormatter реализует стандартное форматирование
type DefaultLogFormatter struct{}

func (f *DefaultLogFormatter) FormatLogLine(host, timestamp string) string {
	return fmt.Sprintf("[%s] %s\n", host, timestamp)
}

// Handler содержит зависимости для HTTP обработчиков
type Handler struct {
	storage    LogStorage
	formatter  LogFormatter
	timeSource TimeSource
}

// TimeSource определяет интерфейс для получения времени
type TimeSource interface {
	Now() time.Time
}

// RealTimeSource возвращает реальное время
type RealTimeSource struct{}

func (rts *RealTimeSource) Now() time.Time {
	return time.Now()
}

// NewHandler создаёт новый обработчик с зависимостями
func NewHandler(storage LogStorage, formatter LogFormatter, timeSource TimeSource) *Handler {
	return &Handler{
		storage:    storage,
		formatter:  formatter,
		timeSource: timeSource,
	}
}

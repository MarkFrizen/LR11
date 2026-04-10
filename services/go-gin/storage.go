package main

// LogStorage определяет интерфейс для работы с логом
type LogStorage interface {
	WriteLine(line string) error
	ReadAll() (string, error)
}

// TimeFormatter определяет интерфейс для форматирования времени
type TimeFormatter interface {
	Format() string
}

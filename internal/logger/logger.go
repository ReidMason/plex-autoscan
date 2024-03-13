package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type SLogger struct {
	logger slog.Logger
}

func NewLogger() *SLogger {
	opts := &slog.HandlerOptions{
		// AddSource: true,
	}

	file, err := os.OpenFile("data/log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Failed to open log file", err)
		panic(err)
	}

	multiWriter := io.MultiWriter(file, os.Stdout)

	logger := slog.New(slog.NewJSONHandler(multiWriter, opts))

	return &SLogger{logger: *logger}
}

func (l *SLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *SLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *SLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

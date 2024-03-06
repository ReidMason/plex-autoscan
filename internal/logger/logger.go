package logger

import (
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
	return &SLogger{logger: *slog.New(slog.NewJSONHandler(os.Stdout, nil))}
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

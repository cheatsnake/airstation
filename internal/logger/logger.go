package logger

import (
	"log/slog"
)

func New() *slog.Logger {
	logger := slog.New(slog.Default().Handler())
	return logger
}

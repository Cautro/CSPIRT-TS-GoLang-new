package logger

import (
	"io"
	"log/slog"
	"os"
)

func New() (*slog.Logger, *os.File, error) {
	logFile, err := os.OpenFile("cpirt/data/logs.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, err
	}

	writer := io.MultiWriter(os.Stdout, logFile)

	handler := slog.NewTextHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	return slog.New(handler), logFile, nil
}
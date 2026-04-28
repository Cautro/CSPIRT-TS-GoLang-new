package storage

import (
	"cspirt/internal/logger"
	"log/slog"
)

func writeLog(entry logger.LogEntry) {
	if err := logger.Write(entry); err != nil {
		slog.Error("failed to write log", "error", err)
	}
}

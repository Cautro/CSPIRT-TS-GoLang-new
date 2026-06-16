package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

const (
	logQueueSize  = 2048
	logBufferSize = 64 * 1024
	flushInterval = 500 * time.Millisecond
)

type LogEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Action  string `json:"action"`
	Login   string `json:"login"`
	Role    string `json:"role"`
	Class   string `json:"class"`
	Message string `json:"message"`
}

type request struct {
	entry LogEntry
	sync  chan error
}

var (
	initOnce  sync.Once
	initErr   error
	started   atomic.Bool
	requestCh chan request
)

func Write(entry LogEntry) error {
	if err := ensureWriter(); err != nil {
		return err
	}

	entry.Time = time.Now().Format(time.RFC3339)
	requestCh <- request{entry: entry}

	return nil
}

func WriteSafe(entry LogEntry) {
	if err := Write(entry); err != nil {
		slog.Error("failed to write log", "error", err)
	}
}

func Sync() error {
	if !started.Load() {
		return nil
	}

	syncCh := make(chan error, 1)
	requestCh <- request{sync: syncCh}
	return <-syncCh
}

func ensureWriter() error {
	initOnce.Do(func() {
		initErr = startWriter()
		if initErr == nil {
			started.Store(true)
		}
	})

	return initErr
}

func startWriter() error {
	logDir := "data"
	logPath := filepath.Join(logDir, "log.log")

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	requestCh = make(chan request, logQueueSize)
	go runWriter(file, bufio.NewWriterSize(file, logBufferSize))

	return nil
}

func runWriter(file *os.File, writer *bufio.Writer) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case req := <-requestCh:
			if req.sync != nil {
				req.sync <- flush(writer, file)
				close(req.sync)
				continue
			}

			if err := writeEntry(writer, req.entry); err != nil {
				slog.Error("failed to write log", "error", err)
			}
		case <-ticker.C:
			if err := writer.Flush(); err != nil {
				slog.Error("failed to flush log buffer", "error", err)
			}
		}
	}
}

func writeEntry(writer *bufio.Writer, entry LogEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal log entry: %w", err)
	}

	if _, err := writer.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write log: %w", err)
	}

	return nil
}

func flush(writer *bufio.Writer, file *os.File) error {
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush log buffer: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("sync log file: %w", err)
	}

	return nil
}

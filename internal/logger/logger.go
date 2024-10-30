package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/grid-stream-org/batcher/internal/config"
)

var (
	once   sync.Once
	logger *slog.Logger
)

func InitLogger(cfg *config.LoggerConfig, outputWriter io.Writer) {
	once.Do(func() { initLogger(cfg, outputWriter) })
}

func initLogger(cfg *config.LoggerConfig, outputWriter io.Writer) {
	var level slog.Level
	switch cfg.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var output io.Writer
	if outputWriter != nil {
		output = outputWriter
	} else {
		switch cfg.Output {
		case "stdout":
			output = os.Stdout
		case "stderr":
			output = os.Stderr
		default:
			var err error
			output, err = os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("Failed to open log file: %v", err)
			}
		}
	}

	var handler slog.Handler
	switch cfg.Format {
	case "json":
		handler = slog.NewJSONHandler(output, &slog.HandlerOptions{Level: level})
	case "text":
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{Level: level})
	default:
		handler = slog.NewTextHandler(output, &slog.HandlerOptions{Level: level})
	}

	logger = slog.New(handler)
}

func Logger() *slog.Logger {
	if logger == nil {
		log.Fatalf("Logger not initialized. Call logger.InitLogger() before using the logger.")
	}
	return logger
}

func Reset() {
	once = sync.Once{}
	logger = nil
}

package logger

import (
	"io"
	"log/slog"
	"os"

	"github.com/grid-stream-org/batcher/internal/config"
)

func Init(cfg *config.LoggerConfig, outputWriter io.Writer) (*slog.Logger, error) {
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
				return nil, err
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

	return slog.New(handler), nil
}

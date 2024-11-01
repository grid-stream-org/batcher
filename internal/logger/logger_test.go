package logger_test

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/logger"
)

func TestLogger_NotNilAfterInitialization(t *testing.T) {
	cfg := &config.LoggerConfig{
		Level:  "DEBUG",
		Format: "text",
		Output: "stdout",
	}
	log, err := logger.Init(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, log, "Logger should not be nil after initialization")
}

func TestLogger_LoggingAtDifferentLevels(t *testing.T) {

	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer w.Close()

	cfg := &config.LoggerConfig{
		Level:  "DEBUG",
		Format: "text",
		Output: "",
	}
	log, err := logger.Init(cfg, w)
	require.NoError(t, err)

	log.Debug("Debug message")
	log.Info("Info message")
	log.Warn("Warn message")
	log.Error("Error message")

	w.Close()
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	output := string(out)

	assert.Contains(t, output, "Debug message")
	assert.Contains(t, output, "Info message")
	assert.Contains(t, output, "Warn message")
	assert.Contains(t, output, "Error message")
}

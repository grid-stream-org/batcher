package logger

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_NotNilAfterInitialization(t *testing.T) {
	cfg := &Config{
		Level:  "DEBUG",
		Format: "text",
		Output: "stdout",
	}
	log, err := New(cfg, nil)
	require.NoError(t, err)
	require.NotNil(t, log, "Logger should not be nil after initialization")
}

func TestLogger_LoggingAtDifferentLevels(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer w.Close()

	cfg := &Config{
		Level:  "DEBUG",
		Format: "text",
		Output: "",
	}
	log, err := New(cfg, w)
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

func TestLogger_ValidateConfig(t *testing.T) {
	tests := []struct {
		name        string
		cfg         Config
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			cfg: Config{
				Level:  "INFO",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "valid config with lowercase level",
			cfg: Config{
				Level:  "debug",
				Format: "text",
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			cfg: Config{
				Level:  "INVALID",
				Format: "json",
			},
			wantErr:     true,
			errContains: "invalid log level: INVALID",
		},
		{
			name: "invalid format",
			cfg: Config{
				Level:  "INFO",
				Format: "yaml",
			},
			wantErr:     true,
			errContains: "invalid log format: yaml",
		},
		{
			name: "empty level",
			cfg: Config{
				Level:  "",
				Format: "json",
			},
			wantErr:     true,
			errContains: "invalid log level:",
		},
		{
			name: "empty format",
			cfg: Config{
				Level:  "INFO",
				Format: "",
			},
			wantErr:     true,
			errContains: "invalid log format:",
		},
		{
			name:        "all empty",
			cfg:         Config{},
			wantErr:     true,
			errContains: "invalid log level:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

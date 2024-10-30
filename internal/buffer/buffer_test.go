package buffer

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuffer(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	tests := []struct {
		name    string
		cfg     config.BufferConfig
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: config.BufferConfig{
				Duration: 100 * time.Millisecond,
				Offset:   10 * time.Millisecond,
				Capacity: 10,
			},
			wantErr: false,
		},
		{
			name: "invalid config - duration <= offset",
			cfg: config.BufferConfig{
				Duration: 10 * time.Millisecond,
				Offset:   10 * time.Millisecond,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBuffer(context.Background(), &tt.cfg, testLogger)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, b)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, b)
				b.Stop()
			}
		})
	}
}

func TestBuffer_Add(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	b, err := NewBuffer(context.Background(), &config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
		Capacity: 10,
	}, testLogger)
	require.NoError(t, err)
	defer b.Stop()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid data",
			data:    []byte("test data"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := b.Add(tt.data)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuffer_Flush(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	b, err := NewBuffer(context.Background(), &config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
		Capacity: 10,
	}, testLogger)
	require.NoError(t, err)
	defer b.Stop()

	t.Run("flush empty buffer", func(t *testing.T) {
		b.Flush()
		select {
		case data := <-b.FlushedData():
			t.Fatalf("Expected no data, but got: %v", data)
		case <-time.After(50 * time.Millisecond):
			// Test passes; no data was flushed
		}
	})

	t.Run("flush with data", func(t *testing.T) {
		testData := []byte("test data")
		err := b.Add(testData)
		require.NoError(t, err)

		b.Flush()

		select {
		case data := <-b.FlushedData():
			assert.Len(t, data, 1)
			assert.Equal(t, testData, data[0])
		case <-time.After(50 * time.Millisecond):
			t.Fatal("Expected flushed data, but none received")
		}
	})
}

func TestBuffer_AutoFlush(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	duration := 100 * time.Millisecond
	b, err := NewBuffer(context.Background(), &config.BufferConfig{
		Duration: duration,
		Offset:   10 * time.Millisecond,
		Capacity: 10,
	}, testLogger)
	require.NoError(t, err)
	defer b.Stop()

	testData := []byte("test data")
	err = b.Add(testData)
	require.NoError(t, err)

	select {
	case data := <-b.FlushedData():
		assert.Len(t, data, 1)
		assert.Equal(t, testData, data[0])
	case <-time.After(duration * 2):
		t.Fatal("Auto flush timeout")
	}
}

func TestBuffer_Stop(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	ctx := context.Background()
	b, err := NewBuffer(ctx, &config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
		Capacity: 10,
	}, testLogger)
	require.NoError(t, err)

	err = b.Add([]byte("test data"))
	require.NoError(t, err)

	b.Stop()

	// Drain any remaining data
	for range b.FlushedData() {
		// Optionally, you can check the data here
	}

	// Verify the channel is closed
	_, ok := <-b.FlushedData()
	assert.False(t, ok, "Expected the flushed data channel to be closed")
}

func TestBuffer_ConcurrentAccess(t *testing.T) {
	testLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	b, err := NewBuffer(context.Background(), &config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
		Capacity: 10,
	}, testLogger)
	require.NoError(t, err)
	defer b.Stop()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				err := b.Add([]byte(fmt.Sprintf("data-%d-%d", id, j)))
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	b.Flush()

	select {
	case data := <-b.FlushedData():
		assert.Len(t, data, numGoroutines*numOperations)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected flushed data, but none received")
	}
}

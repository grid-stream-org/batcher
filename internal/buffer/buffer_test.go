package buffer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuffer(t *testing.T) {
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
			b, err := NewBuffer(context.Background(), tt.cfg)
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
	b, err := NewBuffer(context.Background(), config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
	})
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
		{
			name:    "nil data",
			data:    nil,
			wantErr: true,
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
	b, err := NewBuffer(context.Background(), config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
	})
	require.NoError(t, err)
	defer b.Stop()

	t.Run("flush empty buffer", func(t *testing.T) {
		success := b.Flush()
		assert.True(t, success)
	})

	t.Run("flush with data", func(t *testing.T) {
		testData := []byte("test data")
		err := b.Add(testData)
		require.NoError(t, err)

		success := b.Flush()
		assert.True(t, success)

		select {
		case data := <-b.FlushedData():
			assert.Len(t, data, 1)
			assert.Equal(t, testData, data[0])
		default:
			t.Fatal("expected flushed data")
		}
	})
}

func TestBuffer_AutoFlush(t *testing.T) {
	duration := 100 * time.Millisecond
	b, err := NewBuffer(context.Background(), config.BufferConfig{
		Duration: duration,
		Offset:   10 * time.Millisecond,
	})
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
		t.Fatal("auto flush timeout")
	}
}

func TestBuffer_Stop(t *testing.T) {
	ctx := context.Background()
	b, err := NewBuffer(ctx, config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
	})
	require.NoError(t, err)

	err = b.Add([]byte("test data"))
	require.NoError(t, err)

	b.Stop()

	select {
	case _, ok := <-b.FlushedData():
		if ok {
			for range b.FlushedData() {
			}
		}
	default:
	}

	_, ok := <-b.FlushedData()
	assert.False(t, ok, "channel should be closed")
}

func TestBuffer_ConcurrentAccess(t *testing.T) {
	b, err := NewBuffer(context.Background(), config.BufferConfig{
		Duration: 100 * time.Millisecond,
		Offset:   10 * time.Millisecond,
	})
	require.NoError(t, err)
	defer b.Stop()

	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				err := b.Add([]byte(fmt.Sprintf("data-%d-%d", id, j)))
				assert.NoError(t, err)
			}
			done <- true
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	success := b.Flush()
	assert.True(t, success)

	select {
	case data := <-b.FlushedData():
		assert.Len(t, data, numGoroutines*numOperations)
	default:
		t.Fatal("expected flushed data")
	}
}

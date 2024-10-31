package buffer

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"github.com/grid-stream-org/batcher/internal/config"
)

var (
	ErrDurationGreater = errors.New("duration must be greater than offset")
)

type Metrics struct {
	MessagesCount int64
	FlushCount    int64
	LastFlushTime time.Time
}

type Buffer struct {
	mu      sync.Mutex
	data    [][]byte
	flushCh chan [][]byte
	ticker  *time.Ticker
	wg      sync.WaitGroup
	log     *slog.Logger
	metrics Metrics
}

func New(ctx context.Context, cfg *config.BufferConfig, log *slog.Logger) (*Buffer, error) {
	if cfg.Duration <= cfg.Offset {
		return nil, ErrDurationGreater
	}

	b := &Buffer{
		data:    make([][]byte, 0),
		flushCh: make(chan [][]byte, cfg.Capacity),
		ticker:  time.NewTicker(cfg.Duration - cfg.Offset),
		log:     log.With("component", "buffer"),
	}

	b.wg.Add(1)
	go b.autoFlush(ctx)

	return b, nil
}

func (b *Buffer) Add(data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.data = append(b.data, data)
	atomic.AddInt64(&b.metrics.MessagesCount, 1)

	b.log.Debug("Data added to buffer", "buffer_size", len(b.data))
	return nil
}

func (b *Buffer) flush() {
	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		return
	}

	data := b.data
	b.data = make([][]byte, 0)
	b.mu.Unlock()

	// Blocking send to ensure no data loss
	b.flushCh <- data

	atomic.AddInt64(&b.metrics.FlushCount, 1)
	b.metrics.LastFlushTime = time.Now()

	b.log.Info("Buffer flushed", "records", len(data))
}

func (b *Buffer) autoFlush(ctx context.Context) {
	defer b.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.ticker.C:
			b.flush()
		}
	}
}

func (b *Buffer) Stop() {
	b.ticker.Stop()
	b.flush() // Final flush
	close(b.flushCh)
	b.wg.Wait()
	b.log.Info("Buffer stopped")
}

func (b *Buffer) FlushedData() <-chan [][]byte {
	return b.flushCh
}

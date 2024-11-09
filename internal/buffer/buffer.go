package buffer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/metrics"
)

type Buffer struct {
	mu      sync.Mutex
	data    [][]byte
	flushCh chan [][]byte
	ticker  *time.Ticker
	wg      sync.WaitGroup
	log     *slog.Logger
}

func New(ctx context.Context, cfg *config.BufferConfig, log *slog.Logger) (*Buffer, error) {
	if cfg.Duration <= cfg.Offset {
		return nil, errors.New("duration must be greater than offset")
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

func (b *Buffer) Add(data []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.data = append(b.data, data)

	metrics.Local.Gauge(metrics.BufferSize).WithLabelValues().Set(float64(len(b.data)))
	b.log.Info("Record added to buffer", "buffer_size", len(b.data))
}

func (b *Buffer) flush() {
	b.log.Info("Flushing buffer")
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

	metrics.Local.Counter(metrics.FlushCount).WithLabelValues().Inc()
	metrics.Local.Gauge(metrics.LastFlushTime).WithLabelValues().SetToCurrentTime()
	metrics.Local.Gauge(metrics.BufferSize).WithLabelValues().Set(0)

	b.log.Info("Buffer flushed", "records", len(data))
}

func (b *Buffer) autoFlush(ctx context.Context) {
	b.log.Info("Starting auto flush")
	defer b.wg.Done()

	for {
		select {
		case <-ctx.Done():
			b.log.Info("Finishing auto flush")
			return
		case <-b.ticker.C:
			b.flush()
		}
	}
}

func (b *Buffer) Stop() {
	b.log.Info("Stopping buffer")
	b.ticker.Stop()
	b.flush() // Final flush
	close(b.flushCh)
	b.wg.Wait()
	b.log.Info("Buffer stopped")
}

func (b *Buffer) FlushedData() <-chan [][]byte {
	return b.flushCh
}

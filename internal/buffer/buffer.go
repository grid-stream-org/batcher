package buffer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
)

type Buffer struct {
	mu      sync.Mutex
	data    [][]byte
	flushCh chan [][]byte
	ticker  *time.Ticker
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	log     *slog.Logger
}

func NewBuffer(ctx context.Context, cfg *config.BufferConfig, log *slog.Logger) (*Buffer, error) {
	if cfg.Duration <= cfg.Offset {
		return nil, fmt.Errorf("duration must be greater than offset")
	}
	ctx, cancel := context.WithCancel(ctx)
	duration := cfg.Duration - cfg.Offset

	b := &Buffer{
		data:    make([][]byte, 0),
		flushCh: make(chan [][]byte, cfg.Capacity),
		ticker:  time.NewTicker(duration),
		cancel:  cancel,
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
	b.log.Debug("Data added to buffer", "current_length", len(b.data))
	return nil
}

func (b *Buffer) Flush() {
	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		return
	}
	data := b.data
	b.data = make([][]byte, 0)
	b.mu.Unlock()
	b.log.Info("Buffer flushed", "records_flushed", len(data))
	b.flushCh <- data

}

func (b *Buffer) autoFlush(ctx context.Context) {
	defer func() {
		b.ticker.Stop()
		b.Flush()
		b.wg.Done()
	}()

	for {
		select {
		case <-ctx.Done():
			b.log.Info("Context canceled; stopping autoFlush")
			return
		case <-b.ticker.C:
			b.log.Debug("Timer tick; flushing buffer")
			b.Flush()
		}
	}
}

func (b *Buffer) FlushedData() <-chan [][]byte {
	return b.flushCh
}

func (b *Buffer) Stop() {
	b.cancel()
	b.wg.Wait()
	close(b.flushCh)
	b.log.Info("Buffer stopped")
}

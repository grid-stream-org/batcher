package buffer

import (
	"context"
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
}

func NewBuffer(ctx context.Context, cfg config.BufferConfig) (*Buffer, error) {
	ctx, cancel := context.WithCancel(ctx)
	duration := cfg.Duration - cfg.Offset

	b := &Buffer{
		data:    make([][]byte, 0),
		flushCh: make(chan [][]byte, 1),
		ticker:  time.NewTicker(duration),
		cancel:  cancel,
	}

	b.wg.Add(1)
	go b.autoFlush(ctx)
	return b, nil
}

func (b *Buffer) Add(data []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, data)
	return nil
}

func (b *Buffer) Flush() bool {
	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		return true
	}

	data := b.data
	b.data = make([][]byte, 0)
	b.mu.Unlock()

	select {
	case b.flushCh <- data:
		return true
	default:
		return false
	}
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
			return
		case <-b.ticker.C:
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
}

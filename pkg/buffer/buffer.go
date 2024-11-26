package buffer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Buffer[T any] struct {
	mu        sync.Mutex
	data      []T
	flushFunc func(ctx context.Context, data []T) error
	ticker    *time.Ticker
	wg        sync.WaitGroup
	log       *slog.Logger
}

type Option[T any] func(*Buffer[T]) error

func New[T any](ctx context.Context, duration time.Duration, offset time.Duration, log *slog.Logger, opts ...Option[T]) (*Buffer[T], error) {
	buf := &Buffer[T]{
		data:      make([]T, 0),
		flushFunc: func(context.Context, []T) error { return nil },
		ticker:    time.NewTicker(duration - offset),
		log:       log.With("component", "buffer"),
	}

	for _, opt := range opts {
		if err := opt(buf); err != nil {
			return nil, err
		}
	}

	return buf, nil
}

func WithFlushFunc[T any](f func(ctx context.Context, data []T) error) Option[T] {
	return func(b *Buffer[T]) error {
		if f == nil {
			return errors.New("flushFunc cannot be nil")
		}
		b.flushFunc = f
		return nil
	}
}

func (b *Buffer[T]) Add(data T) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.data = append(b.data, data)
	b.log.Debug("record added to buffer", "buffer_size", len(b.data))
}

func (b *Buffer[T]) flush(ctx context.Context) error {
	b.log.Debug("flushing buffer")
	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		b.log.Debug("skipping; nothing to flush")
		return nil
	}

	data := b.data
	b.data = make([]T, 0)
	b.mu.Unlock()

	if err := b.flushFunc(ctx, data); err != nil {
		return errors.WithStack(err)
	}

	b.log.Debug("buffer flushed", "records", len(data))
	return nil
}

func (b *Buffer[T]) AutoFlush(ctx context.Context) {
	b.log.Info("starting auto flush")
	defer b.wg.Done()

	for {
		select {
		case <-ctx.Done():
			b.log.Debug("finishing auto flush")
			return
		case <-b.ticker.C:
			b.flush(ctx)
		}
	}
}

func (b *Buffer[T]) Stop() {
	b.log.Debug("stopping buffer")
	b.ticker.Stop()
	b.flush(context.Background()) // Final flush
	b.wg.Wait()
	b.log.Debug("buffer stopped")
}

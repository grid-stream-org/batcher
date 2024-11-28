package buffer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/pkg/errors"
)

type Buffer[T any] interface {
	Add(data T)
	Start(ctx context.Context, duration time.Duration)
	Stop()
	Flush(ctx context.Context) error
}

type buffer[T any] struct {
	mu        sync.Mutex
	data      []T
	flushFunc func(ctx context.Context, data []T) error
	log       *slog.Logger
	cancel    context.CancelFunc
}

type Option[T any] func(*buffer[T]) error

func New[T any](log *slog.Logger, opts ...Option[T]) (Buffer[T], error) {
	buf := &buffer[T]{
		data:      make([]T, 0),
		flushFunc: func(context.Context, []T) error { return nil },
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
	return func(b *buffer[T]) error {
		if f == nil {
			return errors.New("flushFunc cannot be nil")
		}
		b.flushFunc = f
		return nil
	}
}

func (b *buffer[T]) Add(data T) {
	b.mu.Lock()
	b.data = append(b.data, data)
	b.mu.Unlock()
	b.log.Debug("record added to buffer", "buffer_size", len(b.data))
}

func (b *buffer[T]) Start(ctx context.Context, duration time.Duration) {
	b.mu.Lock()
	if b.cancel != nil {
		b.mu.Unlock()
		b.log.Warn("buffer already started")
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	b.mu.Unlock()

	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				b.log.Debug("context canceled, exiting auto flush")
				if err := b.Flush(context.Background()); err != nil {
					b.log.Error("failed to flush buffer during shutdown", "error", err)
				}
				return
			case <-ticker.C:
				if err := b.Flush(ctx); err != nil {
					b.log.Error("failed to flush buffer", "error", err)
				}
			}
		}
	}()
}

func (b *buffer[T]) Stop() {
	b.mu.Lock()
	if b.cancel == nil {
		b.mu.Unlock()
		b.log.Warn("buffer not running or already stopped")
		return
	}
	b.cancel()
	b.cancel = nil
	b.mu.Unlock()
	b.log.Debug("buffer stopped")
}

func (b *buffer[T]) Flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		b.log.Debug("nothing to flush")
		return nil
	}

	data := b.data
	b.data = nil
	b.mu.Unlock()

	if err := b.flushFunc(ctx, data); err != nil {
		return errors.WithStack(err)
	}

	b.log.Debug("buffer flushed", "records", len(data))
	return nil
}

func (b *buffer[T]) Data() []T {
	return b.data
}

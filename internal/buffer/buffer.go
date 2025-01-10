package buffer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/pkg/stats"
	"github.com/pkg/errors"
)

type FlushOutcome struct {
	Outcomes   []outcome.Outcome
	AvgOutputs []outcome.AggregateAverageOutput
}

type FlushFunc func(ctx context.Context, data *FlushOutcome) error

type Buffer struct {
	cfg       *config.Buffer
	mu        sync.Mutex
	data      []outcome.Outcome
	avgCache  *stats.AvgCache
	flushFunc FlushFunc
	log       *slog.Logger
	cancel    context.CancelFunc
}

func New(cfg *config.Buffer, flushFunc FlushFunc, log *slog.Logger) *Buffer {
	buf := &Buffer{
		cfg:       cfg,
		data:      make([]outcome.Outcome, 0),
		flushFunc: flushFunc,
		log:       log.With("component", "buffer"),
		avgCache:  stats.NewAvgCache(),
		cancel:    nil,
	}
	log.Info("buffer initialized", "start_time", cfg.StartTime, "interval", cfg.Interval, "offset", cfg.Offset)
	return buf
}

func (b *Buffer) Add(data *outcome.Outcome) {
	b.mu.Lock()
	b.data = append(b.data, *data)
	b.mu.Unlock()
	b.avgCache.Add(data.ProjectID, data.TotalOutput)
	b.log.Debug("record added to buffer", "buffer_size", len(b.data))
}

func (b *Buffer) Start(ctx context.Context) {
	b.mu.Lock()
	if b.cancel != nil {
		b.mu.Unlock()
		b.log.Warn("buffer already started")
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	b.mu.Unlock()

	go b.autoFlush(ctx)
}

func (b *Buffer) autoFlush(ctx context.Context) {
	for {
		elapsed := time.Since(b.cfg.StartTime)
		nextFlush := b.cfg.StartTime.Add(elapsed - (elapsed % b.cfg.Interval) + b.cfg.Interval + b.cfg.Offset)
		timer := time.NewTimer(time.Until(nextFlush))

		select {
		case <-ctx.Done():
			timer.Stop()
			b.log.Debug("context canceled, exiting auto flush")
			if err := b.Flush(context.Background()); err != nil {
				b.log.Error("failed to flush buffer during shutdown", "error", err)
			}
			return
		case <-timer.C:
			if err := b.Flush(ctx); err != nil {
				b.log.Error("failed to flush buffer", "error", err)
			}
		}
		timer.Stop()
	}
}

func (b *Buffer) Stop() {
	b.mu.Lock()
	if b.cancel == nil {
		b.mu.Unlock()
		b.log.Warn("buffer not running or already stopped")
		return
	}
	b.cancel()
	b.cancel = nil
	b.mu.Unlock()
	b.avgCache.Close()
	b.log.Debug("buffer stopped")
}

func (b *Buffer) Flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		b.log.Debug("nothing to flush")
		return nil
	}

	avgOutputs := []outcome.AggregateAverageOutput{}
	for projID, runningAvg := range b.avgCache.GetAll() {
		avg := outcome.AggregateAverageOutput{
			ProjectID:     projID,
			AverageOutput: runningAvg.Get(),
			Timestamp:     time.Now(),
		}
		avgOutputs = append(avgOutputs, avg)
	}

	data := &FlushOutcome{
		Outcomes:   b.data,
		AvgOutputs: avgOutputs,
	}

	b.data = nil

	b.mu.Unlock()

	if err := b.flushFunc(ctx, data); err != nil {
		return errors.WithStack(err)
	}

	b.avgCache.Reset()
	b.log.Debug("buffer flushed", "outcomes", len(data.Outcomes), "average_outputs", len(data.AvgOutputs))
	return nil
}

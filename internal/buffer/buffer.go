package buffer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/grid-stream-org/batcher/pkg/validator"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type FlushOutcome struct {
	Outcomes   []outcome.Outcome     `json:"outcomes"`
	AvgOutputs []types.AverageOutput `json:"average_outputs"`
}

type FlushFunc func(ctx context.Context, data *FlushOutcome) error

type Buffer struct {
	cfg       *config.Buffer
	mu        sync.Mutex
	data      []outcome.Outcome
	vc        validator.ValidatorClient
	avgCache  *AvgCache
	flushFunc FlushFunc
	log       *slog.Logger
	cancel    context.CancelFunc
}

func New(cfg *config.Buffer, flushFunc FlushFunc, vc validator.ValidatorClient, log *slog.Logger) *Buffer {
	buf := &Buffer{
		cfg:       cfg,
		data:      make([]outcome.Outcome, 0),
		vc:        vc,
		flushFunc: flushFunc,
		log:       log.With("component", "buffer"),
		avgCache:  NewAvgCache(cfg.StartTime, cfg.StartTime.Add(cfg.Interval)),
		cancel:    nil,
	}
	log.Info("buffer initialized", "start_time", cfg.StartTime.Format(time.RFC3339), "interval", cfg.Interval, "offset", cfg.Offset)
	return buf
}

func (b *Buffer) Add(ctx context.Context, data *outcome.Outcome) {
	b.mu.Lock()
	b.data = append(b.data, *data)
	b.mu.Unlock()
	if exists := b.avgCache.Add(data.ProjectID, data.TotalOutput); !exists {
		if err := b.vc.NotifyProject(ctx, data.ProjectID); err != nil {
			b.log.Error("failed to notify validator of new project", "project_id", data.ProjectID, "error", err)
		}
	}
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
	b.log.Debug("buffer stopped")
}

func (b *Buffer) Flush(ctx context.Context) error {
	b.mu.Lock()
	totalStart := time.Now()
	if len(b.data) == 0 {
		b.mu.Unlock()
		b.log.Info("nothing to flush")
		return nil
	}

	b.log.Debug("starting flush", "data_length", len(b.data))
	outcomes := b.data
	b.data = make([]outcome.Outcome, 0, len(b.data))
	b.mu.Unlock()

	g, ctx := errgroup.WithContext(ctx)
	var validatorTime time.Duration
	var flushTime time.Duration
	var avgOutputs []types.AverageOutput

	g.Go(func() error {
		validateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		validatorStart := time.Now()
		if err := b.vc.SendAverages(validateCtx, b.avgCache.GetProtoOutputs()); err != nil {
			b.log.Error("failed to send averages", "error", err)
			return errors.WithStack(err)
		}
		validatorTime = time.Since(validatorStart)
		return nil
	})

	g.Go(func() error {
		flushCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		avgOutputs = b.avgCache.GetOutputs()

		data := &FlushOutcome{
			Outcomes:   outcomes,
			AvgOutputs: avgOutputs,
		}

		flushStart := time.Now()
		if err := b.flushFunc(flushCtx, data); err != nil {
			b.log.Error("failed to flush data", "error", err)
			return errors.WithStack(err)
		}
		flushTime = time.Since(flushStart)
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	b.avgCache.Reset()

	totalTime := time.Since(totalStart)
	b.log.Info("buffer flushed", "outcomes", len(outcomes),
		"average_outputs", len(avgOutputs),
		"validator_ms", validatorTime.Milliseconds(),
		"flush_ms", flushTime.Milliseconds(),
		"total_ms", totalTime.Milliseconds())
	return nil
}

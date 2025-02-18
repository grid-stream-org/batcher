package buffer

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/grid-stream-org/go-commons/pkg/validator"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
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
	done      chan struct{}
}

func New(ctx context.Context, cfg *config.Buffer, flushFunc FlushFunc, log *slog.Logger) (*Buffer, error) {
	vc, err := validator.New(ctx, cfg.Validator, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	buf := &Buffer{
		cfg:       cfg,
		data:      make([]outcome.Outcome, 0),
		vc:        vc,
		flushFunc: flushFunc,
		log:       log.With("component", "buffer"),
		avgCache:  NewAvgCache(cfg.StartTime, cfg.StartTime.Add(cfg.Interval)),
		done:      make(chan struct{}),
	}
	log.Info("buffer initialized", "start_time", cfg.StartTime.Format(time.RFC3339), "interval", cfg.Interval, "offset", cfg.Offset)
	return buf, nil
}

func (b *Buffer) Add(ctx context.Context, data *outcome.Outcome) {
	b.mu.Lock()
	b.data = append(b.data, *data)
	b.mu.Unlock()
	b.avgCache.Add(data)
	b.log.Debug("record added to buffer", "buffer_size", len(b.data))
}

func (b *Buffer) Start(ctx context.Context) {
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
			b.log.Debug("context canceled, performing final flush")
			if err := b.Flush(context.Background()); err != nil {
				b.log.Error("failed to flush buffer during shutdown", "error", err)
			}
			close(b.done)
			return
		case <-timer.C:
			if err := b.Flush(ctx); err != nil {
				b.log.Error("failed to flush buffer", "error", err)
			}
		}
		timer.Stop()
	}
}

func (b *Buffer) Stop() error {
	<-b.done
	// Close validator connection
	if err := b.vc.Close(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (b *Buffer) Flush(parentCtx context.Context) error {
	timeoutCtx, timeoutCancel := context.WithTimeout(parentCtx, b.cfg.Offset)
	defer timeoutCancel()

	b.mu.Lock()
	if len(b.data) == 0 {
		b.mu.Unlock()
		b.log.Info("nothing to flush")
		return nil
	}

	totalStart := time.Now()
	b.log.Debug("starting flush", "data_length", len(b.data))
	outcomes := b.data
	b.data = make([]outcome.Outcome, 0, len(b.data))
	b.mu.Unlock()

	var validatorTime time.Duration
	var flushTime time.Duration
	var avgOutputs []types.AverageOutput
	var validatorErr, flushErr error
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		validatorStart := time.Now()
		if err := b.vc.SendAverages(timeoutCtx, b.avgCache.GetProtoOutputs()); err != nil {
			validatorErr = errors.WithStack(err)
		}
		validatorTime = time.Since(validatorStart)
	}()

	go func() {
		defer wg.Done()
		avgOutputs = b.avgCache.GetOutputs()
		data := &FlushOutcome{
			Outcomes:   outcomes,
			AvgOutputs: avgOutputs,
		}
		flushStart := time.Now()
		if err := b.flushFunc(timeoutCtx, data); err != nil {
			flushErr = errors.WithStack(err)
		}
		flushTime = time.Since(flushStart)
	}()

	wg.Wait()

	b.avgCache.Reset()

	if validatorErr != nil || flushErr != nil {
		return errors.WithStack(multierr.Combine(validatorErr, flushErr))
	}

	totalTime := time.Since(totalStart)
	b.log.Info("buffer flushed",
		"outcomes", len(outcomes),
		"average_outputs", len(avgOutputs),
		"validator_ms", validatorTime.Milliseconds(),
		"flush_ms", flushTime.Milliseconds(),
		"total_ms", totalTime.Milliseconds())

	return nil
}

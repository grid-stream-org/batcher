package destination

import (
	"context"
	"log/slog"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/stats"
	"github.com/grid-stream-org/batcher/pkg/buffer"
	"github.com/pkg/errors"
)

type FlushFunc = func(ctx context.Context, data []outcome.Outcome) error

type microbatchDestination struct {
	buf        buffer.Buffer[outcome.Outcome]
	childFlush FlushFunc
	avgCache   *stats.AvgCache
	log        *slog.Logger
}

func newMicrobatchDestination(
	ctx context.Context,
	cfg *config.Destination,
	childFlush FlushFunc,
	log *slog.Logger,
) (Destination, error) {
	d := &microbatchDestination{
		childFlush: childFlush,
		avgCache:   stats.NewAvgCache(cfg.Buffer.Duration - cfg.Buffer.Offset),
		log:        log.With("component", "microbatch_destination"),
	}

	buf, err := buffer.New(log, buffer.WithFlushFunc(d.flushFunc))
	if err != nil {
		return nil, err
	}

	d.buf = buf
	buf.Start(ctx, cfg.Buffer.Duration-cfg.Buffer.Offset)

	return d, nil
}

func (d *microbatchDestination) Add(data any) error {
	outcome, ok := data.(outcome.Outcome)
	if !ok {
		return errors.Errorf("expected Outcome, got %T", data)
	}
	d.avgCache.Add(outcome.ProjectID(), outcome.CurrentOutput())
	d.buf.Add(outcome)
	return nil
}

func (d *microbatchDestination) Close() error {
	d.buf.Stop()
	d.log.Info("microbatch destination closed")
	return nil
}

func (d *microbatchDestination) flushFunc(ctx context.Context, data []outcome.Outcome) error {
	d.log.Info("starting flush",
		"original_outcomes", len(data),
		"averages_to_append", d.avgCache.Size(),
	)

	finalData := make([]outcome.Outcome, 0, len(data)+d.avgCache.Size())
	finalData = append(finalData, data...)

	start := time.Now()
	for projID, avg := range d.avgCache.GetAll() {
		avgOutput := avg.Get()

		d.log.Debug("calculating average outcome", "project_id", projID, "avg_output", avgOutput)

		avgData := outcome.AggregateAverageOutput{
			ProjectID:     projID,
			AverageOutput: avgOutput,
			Timestamp:     time.Now(),
		}

		outData := map[string][]any{"aggregate_average_outputs": {avgData}}
		o := outcome.NewAvgOutcome(projID, outData, avgOutput, time.Since(start))
		finalData = append(finalData, o)
	}

	d.log.Info("flush complete",
		"total_outcomes", len(finalData),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return d.childFlush(ctx, finalData)
}

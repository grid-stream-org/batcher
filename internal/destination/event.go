package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/go-commons/pkg/bqclient"
	"github.com/pkg/errors"
)

type eventDestination struct {
	client bqclient.BQClient
	buf    *buffer.Buffer
	log    *slog.Logger
}

func newEventDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	client, err := bqclient.New(ctx, cfg.Database)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &eventDestination{
		client: client,
		log:    log.With("component", "event_destination"),
	}

	buf, err := buffer.New(ctx, cfg.Buffer, d.flushFunc, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d.buf = buf
	d.buf.Start(ctx)
	return d, nil
}

func (d *eventDestination) Add(ctx context.Context, data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}
	d.buf.Add(ctx, outcome)
	return nil
}

func (d *eventDestination) Close() error {
	if err := d.buf.Stop(); err != nil {
		return errors.WithStack(err)
	}

	if err := d.client.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("event destination closed")
	return nil
}

func (d *eventDestination) flushFunc(ctx context.Context, data *buffer.FlushOutcome) error {
	if len(data.Outcomes) == 0 {
		d.log.Debug("no outcomes to flush")
		return nil
	}

	if err := d.client.StreamPut(ctx, "project_averages", data.AvgOutputs); err != nil {
		return errors.WithStack(err)
	}

	d.log.Debug("successfully flushed data to bigquery", "avg_records", len(data.AvgOutputs))
	return nil
}

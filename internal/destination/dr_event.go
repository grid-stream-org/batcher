package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/grid-stream-org/go-commons/pkg/bqclient"
	"github.com/pkg/errors"
)

type drEventDestination struct {
	client bqclient.BQClient
	buf    *buffer.Buffer
	log    *slog.Logger
}

/*
Honestly just keep everything how it is and make this a new streaming destination. maybe move validator client in here though and rename the destiantion?
*/

func newDREventDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	client, err := bqclient.New(ctx, cfg.Database)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &drEventDestination{
		client: client,
		log:    log.With("component", "dr_event_destination"),
	}

	buf, err := buffer.New(ctx, cfg.Buffer, d.flushFunc, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d.buf = buf
	d.buf.Start(ctx)
	return d, nil
}

func (d *drEventDestination) Add(ctx context.Context, data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}
	d.buf.Add(ctx, outcome)
	return nil
}

func (d *drEventDestination) Close() error {
	d.buf.Stop()

	if err := d.client.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("dr_event destination closed")
	return nil
}

func (d *drEventDestination) flushFunc(ctx context.Context, data *buffer.FlushOutcome) error {
	if len(data.Outcomes) == 0 {
		d.log.Debug("no outcomes to flush")
		return nil
	}

	derCount := 0
	for _, outcome := range data.Outcomes {
		derCount += len(outcome.Data)
	}
	if derCount == 0 {
		d.log.Debug("no DER data to flush")
		return nil
	}

	derData := make([]types.RealTimeDERData, 0, derCount)
	for _, outcome := range data.Outcomes {
		derData = append(derData, outcome.Data...)
	}

	if err := d.client.Put(ctx, "der_data", derData); err != nil {
		return errors.WithStack(err)
	}

	if err := d.client.Put(ctx, "project_averages", data.AvgOutputs); err != nil {
		return errors.WithStack(err)
	}

	d.log.Debug("successfully flushed data to BigQuery", "der_records", len(derData), "avg_records", len(data.AvgOutputs))
	return nil
}

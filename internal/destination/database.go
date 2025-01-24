package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/internal/types"
	"github.com/grid-stream-org/batcher/pkg/bqclient"
	"github.com/grid-stream-org/batcher/pkg/validator"
	"github.com/pkg/errors"
)

type databaseDestination struct {
	client bqclient.BQClient
	buf    *buffer.Buffer
	log    *slog.Logger
}

func newDatabaseDestination(ctx context.Context, cfg *config.Destination, vc validator.ValidatorClient, log *slog.Logger) (Destination, error) {
	client, err := bqclient.New(ctx, cfg.Database)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &databaseDestination{
		client: client,
		log:    log.With("component", "database_destination"),
	}
	d.log.Info("bigquery client initialized", "project_id", cfg.Database.ProjectID, "dataset_id", cfg.Database.DatasetID)

	d.buf = buffer.New(cfg.Buffer, d.flushFunc, vc, log)
	d.buf.Start(ctx)
	return d, nil
}

func (d *databaseDestination) Add(ctx context.Context, data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}
	d.buf.Add(ctx, outcome)
	return nil
}

func (d *databaseDestination) Close() error {
	d.buf.Stop()

	if err := d.client.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("database destination closed")
	return nil
}

func (d *databaseDestination) flushFunc(ctx context.Context, data *buffer.FlushOutcome) error {
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

	input := map[string][]any{
		"der_data":         make([]any, len(derData)),
		"project_averages": make([]any, len(data.AvgOutputs)),
	}

	for i := range derData {
		input["der_data"][i] = derData[i]
	}
	for i := range data.AvgOutputs {
		input["project_averages"][i] = data.AvgOutputs[i]
	}

	if err := d.client.PutAll(ctx, input); err != nil {
		return errors.Wrap(err, "failed to write data to BigQuery")
	}

	d.log.Debug("successfully flushed data to BigQuery", "der_records", len(derData), "avg_records", len(data.AvgOutputs))
	return nil
}

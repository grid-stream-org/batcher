package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
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

	d.buf = buffer.New(cfg.Buffer, d.flushFunc, vc, log)
	return d, errors.WithStack(err)
}

func (d *databaseDestination) Add(data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}
	d.buf.Add(outcome)
	return nil
}

func (d *databaseDestination) Close() error {
	if err := d.client.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.buf.Stop()
	d.log.Info("database destination closed")
	return nil
}

func (d *databaseDestination) flushFunc(ctx context.Context, data *buffer.FlushOutcome) error {

	return nil
}

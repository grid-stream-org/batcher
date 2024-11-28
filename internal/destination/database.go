package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/batcher/pkg/bqclient"
	"github.com/pkg/errors"
)

type databaseDestination struct {
	client         bqclient.BQClient
	microbatchDest Destination
	log            *slog.Logger
}

func newDatabaseDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	// TODO make public config
	client, err := bqclient.New(ctx, cfg.Database.ProjectID, cfg.Database.DatasetID, cfg.Database.CredsPath, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &databaseDestination{
		client: client,
		log:    log.With("component", "database_destination"),
	}

	d.microbatchDest, err = newMicrobatchDestination(ctx, cfg, d.flushFunc, log)
	return d, errors.WithStack(err)
}

func (d *databaseDestination) Add(data any) error {
	return d.microbatchDest.Add(data)
}

func (d *databaseDestination) Close() error {
	if err := d.client.Close(); err != nil {
		return errors.WithStack(err)
	}

	if err := d.microbatchDest.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("database destination closed")
	return nil
}

func (d *databaseDestination) flushFunc(ctx context.Context, data []outcome.Outcome) error {
	// TODO
	return nil
}

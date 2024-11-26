package destination

import (
	"context"
	"log/slog"
	"time"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/pkg/bqclient"
	"github.com/grid-stream-org/batcher/pkg/buffer"
	"github.com/pkg/errors"
)

type databaseOutcome = bqclient.PutInput

type databaseDestination struct {
	client bqclient.BQClient
	buf    *buffer.Buffer[databaseOutcome]
	log    *slog.Logger
}

func newDatabaseDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (*databaseDestination, error) {
	client, err := bqclient.New(ctx, cfg.Database.ProjectID, cfg.Database.DatasetID, cfg.Database.CredsPath, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &databaseDestination{
		client: client,
		log:    log.With("component", "database_destination"),
	}
	buf, err := buffer.New(ctx, cfg.Buffer.Duration, cfg.Buffer.Offset, log, buffer.WithFlushFunc(d.onFlush))
	if err != nil {
		return nil, err
	}
	d.buf = buf
	go buf.AutoFlush(ctx)
	return d, nil
}

func (d *databaseDestination) Add(data any) error {
	outcome, ok := data.(databaseOutcome)
	if !ok {
		return errors.Errorf("expected databaseOutcome, got %T", data)
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

func (d *databaseDestination) onFlush(ctx context.Context, data []databaseOutcome) error {
	tCtx, tCancel := context.WithTimeout(ctx, 5*time.Second)
	defer tCancel()
	return d.client.PutAll(tCtx, data)
}

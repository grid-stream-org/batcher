package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/go-commons/pkg/bqclient"
	"github.com/pkg/errors"
)

type streamDestination struct {
	client bqclient.BQClient
	log    *slog.Logger
}

func newStreamDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	client, err := bqclient.New(ctx, cfg.Database)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &streamDestination{
		client: client,
		log:    log.With("component", "stream_destination"),
	}

	return d, nil
}

func (d *streamDestination) Add(ctx context.Context, data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}

	if err := d.client.StreamPut(ctx, "der_data", outcome.Data); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (d *streamDestination) Close() error {
	if err := d.client.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("stream destination closed")
	return nil
}

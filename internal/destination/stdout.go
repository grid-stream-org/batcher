package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

type stdoutDestination struct {
	microbatchDest Destination
	log            *slog.Logger
}

func newStdoutDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	d := &stdoutDestination{
		log: log.With("component", "stdout_destination"),
	}

	microbatchDest, err := newMicrobatchDestination(ctx, cfg, d.flushFunc, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d.microbatchDest = microbatchDest
	return d, errors.WithStack(err)
}

func (d *stdoutDestination) Add(data any) error {
	d.microbatchDest.Add(data)
	return nil
}

func (d *stdoutDestination) Close() error {
	if err := d.microbatchDest.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("stdout destination closed")
	return nil
}

func (d *stdoutDestination) flushFunc(ctx context.Context, data []outcome.Outcome) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println(string(out))
	return nil
}

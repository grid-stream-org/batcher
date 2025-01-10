package destination

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

type stdoutDestination struct {
	buf *buffer.Buffer
	log *slog.Logger
}

func newStdoutDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	d := &stdoutDestination{
		log: log.With("component", "stdout_destination"),
	}
	d.buf = buffer.New(cfg.Buffer, d.flushFunc, log)
	d.buf.Start(ctx)
	return d, nil
}

func (d *stdoutDestination) Add(data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}
	d.buf.Add(outcome)
	return nil
}

func (d *stdoutDestination) Close() error {
	d.buf.Stop()
	d.log.Info("stdout destination closed")
	return nil
}

func (d *stdoutDestination) flushFunc(ctx context.Context, data *buffer.FlushOutcome) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println(string(out))
	return nil
}

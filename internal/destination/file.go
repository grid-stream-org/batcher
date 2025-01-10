package destination

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

type fileDestination struct {
	file    *os.File
	encoder *json.Encoder
	buf     *buffer.Buffer
	log     *slog.Logger
}

func newFileDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &fileDestination{
		file:    file,
		encoder: json.NewEncoder(file),
		log:     log.With("component", "file_destination"),
	}

	d.buf = buffer.New(cfg.Buffer, d.flushFunc, log)
	d.encoder.SetIndent("", "	")
	d.buf.Start(ctx)
	return d, nil
}

func (d *fileDestination) Add(data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}
	d.buf.Add(outcome)
	return nil
}

func (d *fileDestination) Close() error {
	if err := d.file.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.buf.Stop()
	d.log.Info("file destination closed")
	return nil
}

func (d *fileDestination) flushFunc(_ context.Context, data *buffer.FlushOutcome) error {
	return d.encoder.Encode(data)
}

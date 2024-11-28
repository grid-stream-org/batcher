package destination

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

type fileDestination struct {
	file           *os.File
	encoder        *json.Encoder
	microbatchDest Destination
	log            *slog.Logger
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

	d.encoder.SetIndent("", "	")
	d.microbatchDest, err = newMicrobatchDestination(ctx, cfg, d.flushFunc, log)
	return d, errors.WithStack(err)
}

func (d *fileDestination) Add(data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *Outcome, got %T", data)
	}
	if err := d.encoder.Encode(outcome); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (d *fileDestination) Close() error {
	if err := d.file.Close(); err != nil {
		return errors.WithStack(err)
	}

	if err := d.microbatchDest.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("file destination closed")
	return nil
}

func (d *fileDestination) flushFunc(ctx context.Context, data []outcome.Outcome) error {
	return d.encoder.Encode(data)
}

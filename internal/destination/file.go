package destination

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/grid-stream-org/go-commons/pkg/validator"
	"github.com/pkg/errors"
)

type fileDestination struct {
	file    *os.File
	encoder *json.Encoder
	buf     *buffer.Buffer
	log     *slog.Logger
}

func newFileDestination(ctx context.Context, cfg *config.Destination, vc validator.ValidatorClient, log *slog.Logger) (Destination, error) {
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, errors.WithStack(err)
	}

	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	d := &fileDestination{
		file:    file,
		encoder: json.NewEncoder(file),
		log:     log.With("component", "file_destination"),
	}

	d.buf = buffer.New(cfg.Buffer, d.flushFunc, vc, log)
	d.encoder.SetIndent("", "	")
	d.buf.Start(ctx)
	return d, nil
}

func (d *fileDestination) Add(ctx context.Context, data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}
	d.buf.Add(ctx, outcome)
	return nil
}

func (d *fileDestination) Close() error {
	d.buf.Stop()

	if err := d.file.Close(); err != nil {
		return errors.WithStack(err)
	}

	d.log.Info("file destination closed")
	return nil
}

func (d *fileDestination) flushFunc(_ context.Context, data *buffer.FlushOutcome) error {
	return d.encoder.Encode(data)
}

package destination

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

type fileDestination struct {
	file    *os.File
	encoder *json.Encoder
	log     *slog.Logger
}

func newFileDestination(cfg *config.Destination, log *slog.Logger) (*fileDestination, error) {
	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	fd := &fileDestination{
		file:    file,
		encoder: json.NewEncoder(file),
		log:     log.With("component", "file_destination"),
	}
	fd.encoder.SetIndent("", "	")
	return fd, nil
}

func (fd *fileDestination) Add(data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *Outcome, got %T", data)
	}
	if err := fd.encoder.Encode(outcome); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (fd *fileDestination) Close() error {
	if err := fd.file.Close(); err != nil {
		return errors.WithStack(err)
	}
	fd.log.Info("file destination closed")
	return nil
}

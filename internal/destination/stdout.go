package destination

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

type stdoutDestination struct {
	writer io.Writer
	log    *slog.Logger
	mu     sync.Mutex
	enc    *json.Encoder
}

func newStdoutDestination(log *slog.Logger) (Destination, error) {
	d := &stdoutDestination{
		writer: os.Stdout,
		log:    log.With("component", "stdout_destination"),
	}

	d.enc = json.NewEncoder(d.writer)
	d.enc.SetIndent("", "  ")
	d.enc.SetEscapeHTML(false)

	d.log.Info("stdout destination initialized")
	return d, nil
}

func (d *stdoutDestination) Add(_ context.Context, data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *outcome.Outcome, got %T", data)
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.enc.Encode(outcome); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (d *stdoutDestination) Close() error {
	if f, ok := d.writer.(*os.File); ok {
		if err := f.Sync(); err != nil {
			return errors.WithStack(err)
		}
	}

	d.log.Info("stdout destination closed")
	return nil
}

package destination

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/outcome"
	"github.com/pkg/errors"
)

type stdoutDestination struct {
	log *slog.Logger
}

func newStdoutDestination(log *slog.Logger) (*stdoutDestination, error) {
	return &stdoutDestination{
		log: log.With("component", "stdout_destination"),
	}, nil
}

func (d *stdoutDestination) Add(data any) error {
	outcome, ok := data.(*outcome.Outcome)
	if !ok {
		return errors.Errorf("expected *Outcome, got %T", data)
	}
	output, err := json.MarshalIndent(outcome, "", "  ")
	if err != nil {
		return errors.WithStack(err)
	}

	fmt.Println(string(output))
	return nil
}

func (d *stdoutDestination) Close() error {
	d.log.Info("stdout destination closed")
	return nil
}

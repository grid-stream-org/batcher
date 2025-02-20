package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/pkg/errors"
)

func NewDestination(ctx context.Context, cfg *config.Destination, log *slog.Logger) (Destination, error) {
	switch cfg.Type {
	case "event":
		return newEventDestination(ctx, cfg, log)
	case "stdout":
		return newStdoutDestination(log)
	case "stream":
		return newStreamDestination(ctx, cfg, log)
	default:
		return nil, errors.Errorf("invalid destination type: %s", cfg.Type)
	}
}

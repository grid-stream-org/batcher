package destination

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/go-commons/pkg/validator"
	"github.com/pkg/errors"
)

func NewDestination(ctx context.Context, cfg *config.Destination, vc validator.ValidatorClient, log *slog.Logger) (Destination, error) {
	switch cfg.Type {
	case "database":
		return newDatabaseDestination(ctx, cfg, vc, log)
	case "file":
		return newFileDestination(ctx, cfg, vc, log)
	case "stdout":
		return newStdoutDestination(ctx, cfg, vc, log)
	default:
		return nil, errors.Errorf("invalid destination type: %s", cfg.Type)
	}
}

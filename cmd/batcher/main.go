package batcher

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/logger"
)

func main() {
	// Load config
	cfg := config.MustLoadConfig()

	// Initialize logger
	logger.InitLogger(&cfg.Logger, nil)
	log := logger.Logger()

	// Create context with cancellation for shutdown handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the buffer
	buf, err := buffer.NewBuffer(ctx, &cfg.Buffer, log)
	if err != nil {
		log.Error("Failed to initialize buffer", "error", err)
		os.Exit(1)
	}

	// Run service
	if err := run(ctx, buf, log); err != nil {
		log.Error("Run failed", "error", err)
		os.Exit(1)
	}

	// Handle graceful shutdown
	handleShutdown(cancel, buf, log)
}

func run(ctx context.Context, buf *buffer.Buffer, log *slog.Logger) error {
	log.Info("Application is running...")
	<-ctx.Done()

	return nil
}

func handleShutdown(cancel context.CancelFunc, buf *buffer.Buffer, log *slog.Logger) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	log.Info("Received signal, initiating shutdown", "signal", sig)
	cancel()
	buf.Stop()
	log.Info("Shutdown complete")
}

package batcher

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/logger"
	"github.com/grid-stream-org/batcher/internal/mqtt"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Set up signal handling with context
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}

	// Initialize logger
	log, err := logger.Init(&cfg.Logger, nil)
	if err != nil {
		return errors.Wrap(err, "initializing logger")
	}

	// Initialize the buffer
	buf, err := buffer.New(ctx, &cfg.Buffer, log)
	if err != nil {
		return errors.Wrap(err, "initializing buffer")
	}
	defer buf.Stop()

	// Initialize MQTT client
	mqtt, err := mqtt.NewClient(ctx, &cfg.MQTT, buf, log)
	if err != nil {
		return errors.Wrap(err, "initializing mqtt client")
	}
	defer mqtt.Stop()

	// Wait for shutdown signal
	log.Info("Application is running...")
	<-ctx.Done()

	log.Info("Shutting down...")
	return nil
}

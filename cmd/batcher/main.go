package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"

	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/logger"
	"github.com/grid-stream-org/batcher/internal/mqtt"
	"github.com/grid-stream-org/batcher/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Set up signal handling with context
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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

	// Initialize prometheus metrics
	metrics.InitMetricsProvider()
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)

	// Initialize the buffer
	buf, err := buffer.New(ctx, &cfg.Buffer, log)
	if err != nil {
		return errors.Wrap(err, "initializing buffer")
	}

	// Initialize MQTT client
	mqtt, err := mqtt.NewClient(ctx, &cfg.MQTT, buf, log)
	if err != nil {
		return errors.Wrap(err, "initializing mqtt client")
	}
	if err := mqtt.Subscribe(); err != nil {
		return errors.Wrap(err, "subscribing")
	}

	// Wait for shutdown signal
	log.Info("Application is running...")
	<-ctx.Done()

	log.Info("Shutting down...")
	buf.Stop()
	mqtt.Stop()
	log.Info("Successfully shut down")
	return nil
}

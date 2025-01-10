package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/grid-stream-org/batcher/internal/batcher"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/metrics"
	"github.com/grid-stream-org/batcher/pkg/logger"
	"github.com/grid-stream-org/batcher/pkg/sigctx"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/multierr"
)

func main() {
	log := logger.Default()
	exitCode := 0
	if err := run(); err != nil {
		exitCode = handleErrors(err, log)
	}
	log.Info("Done", "exitCode", exitCode)
	os.Exit(exitCode)
}

func run() (err error) {
	// Set up our signal handler
	ctx, cancel := sigctx.New(context.Background())
	defer cancel()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Initialize logger
	log, err := logger.New(cfg.Log, nil)
	if err != nil {
		return err
	}

	// Initialize Prometheus metrics
	metrics.InitMetricsProvider()
	http.Handle("/metrics", promhttp.Handler())
	go metricsListenAndServe(log)

	// Create batcher
	batcher, err := batcher.New(ctx, cfg, log)
	if err != nil {
		return err
	}

	// Check for timeout
	// Do not return the context cancellation error because we suppress them (to account for signals)
	if cfg.Batcher.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, cfg.Batcher.Timeout)
		defer cancel()
	}

	// Run batcher
	err = batcher.Run(ctx)

	// Check for timeout
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.Errorf("Batcher timed out after %s", cfg.Batcher.Timeout)
	}

	return err
}

func handleErrors(err error, log *slog.Logger) int {
	if err == nil {
		return 0
	}
	var exitCode int
	errs := []error{}
	// Filter and process errors
	for _, mErr := range multierr.Errors(err) {
		var sigErr *sigctx.SignalError
		if errors.As(mErr, &sigErr) {
			exitCode = sigErr.SigNum()
		} else if !errors.Is(mErr, context.Canceled) {
			errs = append(errs, mErr)
		}
	}
	// Log non-signal errors
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error("error occurred", "error", err, "stack", fmt.Sprintf("%+v", err))
		}
		if exitCode == 0 {
			exitCode = 255
		}
	}
	return exitCode
}

func metricsListenAndServe(log *slog.Logger) {
	if err := http.ListenAndServe(":2112", nil); err != nil {
		log.Warn("metrics server failed to start; metrics will not be collected", "reason", err)
	}
}

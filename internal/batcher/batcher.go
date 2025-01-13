package batcher

import (
	"context"
	"log/slog"

	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/destination"
	"github.com/grid-stream-org/batcher/internal/mqtt"
	"github.com/grid-stream-org/batcher/internal/task"
	"github.com/grid-stream-org/batcher/pkg/eventbus"
	"github.com/grid-stream-org/batcher/pkg/validator"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type Batcher struct {
	cfg  *config.Config
	dest destination.Destination
	tp   *task.TaskPool
	vc   validator.ValidatorClient
	mqtt *mqtt.Client
	eb   eventbus.EventBus
	log  *slog.Logger
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) (*Batcher, error) {
	vc, err := validator.New(ctx, cfg.Validator)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	log.Info("validator client created successfully")

	dest, err := destination.NewDestination(ctx, cfg.Destination, vc, log)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	eb := eventbus.New()
	mqtt, err := mqtt.NewClient(cfg.MQTT, eb, log)
	if err != nil {
		dest.Close() // best effort cleanup
		return nil, errors.WithStack(err)
	}

	tp := task.NewTaskPool(ctx, cfg.Pool, dest, log)

	return &Batcher{
		cfg:  cfg,
		dest: dest,
		tp:   tp,
		vc:   vc,
		mqtt: mqtt,
		eb:   eb,
		log:  log.With("component", "batcher"),
	}, nil
}

func (b *Batcher) Run(ctx context.Context) (err error) {
	b.log.Info("starting batcher")
	defer func() {
		if stopErr := b.Stop(ctx); stopErr != nil {
			err = multierr.Combine(err, stopErr)
		}
	}()

	// Start event listener
	go b.listen(ctx)

	// Start task pool
	b.tp.Start(ctx)

	// Connect to MQTT
	if err := b.mqtt.Connect(); err != nil {
		return errors.WithStack(err)
	}

	// Subscribe to topic
	if err := b.mqtt.Subscribe(); err != nil {
		return errors.WithStack(err)
	}

	// Wait for context cancellation
	<-ctx.Done()
	if ctx.Err() != nil {
		err = multierr.Combine(err, ctx.Err())
	}
	return err
}

func (b *Batcher) listen(ctx context.Context) {
	b.log.Debug("starting event listener")
	events := b.eb.Subscribe(b.cfg.Pool.Capacity)
	defer b.eb.Unsubscribe(events)

	for {
		select {
		case event := <-events:
			b.tp.Submit(event)
		case <-ctx.Done():
			return
		}
	}
}

func (b *Batcher) Stop(ctx context.Context) error {
	b.log.Info("shutting down batcher")

	// Stop MQTT client
	if err := b.mqtt.Stop(); err != nil {
		b.log.Error("failed to stop MQTT client", "error", err)
	}

	// Close event bus to stop listening
	b.eb.Close()

	// Stop task pool and wait for workers to finish
	b.tp.Wait()

	// Close the destination
	if err := b.dest.Close(); err != nil {
		return errors.WithStack(err)
	}

	// Close validator connection
	if err := b.vc.Close(); err != nil {
		return errors.WithStack(err)
	}

	b.log.Info("batcher shutdown complete")
	return nil
}

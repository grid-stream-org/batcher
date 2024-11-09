package db

import (
	"context"
	"log/slog"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

type BQClient struct {
	client *bigquery.Client
	cfg    config.BQConfig
	log    *slog.Logger
}

func NewBQClient(ctx context.Context, cfg config.BQConfig, log *slog.Logger) (*BQClient, error) {
	logger := log.With(
		"component", "bigquery",
		"project", cfg.ProjectID,
		"dataset", cfg.DatasetID,
	)

	logger.Info("creating bigquery client")
	bq, err := bigquery.NewClient(ctx, cfg.ProjectID, option.WithCredentialsFile(cfg.CredsPath))
	if err != nil {
		return nil, errors.Wrap(err, "creating bigquery client")
	}

	c := &BQClient{
		client: bq,
		cfg:    cfg,
		log:    logger,
	}

	c.log.Info("bigquery client created")
	return c, nil
}

func (c *BQClient) Put(ctx context.Context, table string, data any) error {
	if err := c.validateTable(table); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	c.log.Info("beginning insert", "table", table)
	if err := c.inserter(table).Put(ctx, data); err != nil {
		c.log.Error("failed to insert data", "table", table, "error", err)
		return errors.Wrap(err, "inserting rows")
	}
	c.log.Info("insert complete", "table", table)
	return nil
}

func (c *BQClient) PutAll(ctx context.Context, inputs map[string][]any) error {
	if len(inputs) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	for table, data := range inputs {
		if err := c.validateTable(table); err != nil {
			return err
		}

		if len(data) == 0 {
			c.log.Debug("skipping empty table data", "table", table)
			continue
		}
		c.log.Info("processing batch insert", "table", table, "record_count", len(data))

		if err := c.inserter(table).Put(ctx, data); err != nil {
			c.log.Error("batch insert failed", "table", table, "error", err)
			return errors.Wrapf(err, "batch insert failed for table %s", table)
		}
		c.log.Info("batch insert successful", "table", table, "record_count", len(data))
	}

	return nil
}

func (c *BQClient) validateTable(table string) error {
	if table == "" {
		return errors.New("table cannot be empty")
	}
	return nil
}

func (c *BQClient) inserter(table string) *bigquery.Inserter {
	inserter := c.client.Dataset(c.cfg.DatasetID).Table(table).Inserter()
	inserter.SkipInvalidRows = false
	inserter.IgnoreUnknownValues = false
	return inserter
}

func (c *BQClient) Close() error {
	c.log.Info("closing bigquery client")
	if err := c.client.Close(); err != nil {
		return errors.Wrap(err, "closing bigquery client")
	}
	c.log.Info("bigquery client closed")
	return nil
}

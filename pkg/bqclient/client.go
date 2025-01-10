package bqclient

import (
	"context"
	"log/slog"

	"cloud.google.com/go/bigquery"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

type BQClient interface {
	Put(ctx context.Context, table string, data any) error
	PutAll(ctx context.Context, inputs map[string][]any) error
	Inserter(table string) *bigquery.Inserter
	Close() error
}

type Config struct {
	ProjectID string
	DatasetID string
	CredsPath string
}

type bqClient struct {
	cfg    *Config
	client *bigquery.Client
	log    *slog.Logger
}

func New(ctx context.Context, cfg *Config, log *slog.Logger) (BQClient, error) {
	bq, err := bigquery.NewClient(ctx, cfg.ProjectID, option.WithCredentialsFile(cfg.CredsPath))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	c := &bqClient{
		cfg:    cfg,
		client: bq,
		log:    log.With("component", "bigquery", "project", cfg.ProjectID, "dataset", cfg.DatasetID),
	}
	c.log.Debug("bigquery client created")
	return c, nil
}

func (c *bqClient) Put(ctx context.Context, table string, data any) error {
	if table == "" {
		return errors.New("table cannot be empty")
	}

	c.log.Debug("beginning insert", "table", table)
	if err := c.Inserter(table).Put(ctx, data); err != nil {
		return errors.WithStack(err)
	}
	c.log.Debug("insert successful", "table", table)
	return nil
}

func (c *bqClient) PutAll(ctx context.Context, inputs map[string][]any) error {
	if len(inputs) == 0 {
		c.log.Debug("skipping; inputs is empty")
		return nil
	}

	for table, data := range inputs {
		if table == "" {
			return errors.New("table cannot be empty")
		}

		c.log.Debug("processing batch insert", "table", table, "record_count", len(data))
		if err := c.Inserter(table).Put(ctx, data); err != nil {
			return errors.WithStack(err)
		}
		c.log.Debug("batch insert successful", "table", table, "record_count", len(data))
	}
	return nil
}

func (c *bqClient) Inserter(table string) *bigquery.Inserter {
	inserter := c.client.Dataset(c.cfg.DatasetID).Table(table).Inserter()
	inserter.SkipInvalidRows = false
	inserter.IgnoreUnknownValues = false
	return inserter
}

func (c *bqClient) Close() error {
	c.log.Debug("closing bigquery client")
	if err := c.client.Close(); err != nil {
		return errors.WithStack(err)
	}
	c.log.Debug("bigquery client closed")
	return nil
}

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

type bqClient struct {
	projectID string
	datasetID string
	client    *bigquery.Client
	log       *slog.Logger
}

func New(ctx context.Context, projectID string, datasetID string, credsPath string, log *slog.Logger) (BQClient, error) {
	bq, err := bigquery.NewClient(ctx, projectID, option.WithCredentialsFile(credsPath))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	c := &bqClient{
		projectID: projectID,
		datasetID: datasetID,
		client:    bq,
		log:       log.With("component", "bigquery", "project", projectID, "dataset", datasetID),
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
	inserter := c.client.Dataset(c.datasetID).Table(table).Inserter()
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

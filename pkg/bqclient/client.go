package bqclient

import (
	"context"

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
	ProjectID string `koanf:"project_id" json:"project_id" envconfig:"project_id"`
	DatasetID string `koanf:"dataset_id" json:"dataset_id" envconfig:"dataset_id"`
	CredsPath string `koanf:"creds_path" json:"creds_path" envconfig:"creds_path"`
}

type bqClient struct {
	cfg    *Config
	client *bigquery.Client
}

func New(ctx context.Context, cfg *Config) (BQClient, error) {
	bq, err := bigquery.NewClient(ctx, cfg.ProjectID, option.WithCredentialsFile(cfg.CredsPath))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	c := &bqClient{
		cfg:    cfg,
		client: bq,
	}
	return c, nil
}

func (c *bqClient) Put(ctx context.Context, table string, data any) error {
	if table == "" {
		return errors.New("table cannot be empty")
	}

	if err := c.Inserter(table).Put(ctx, data); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *bqClient) PutAll(ctx context.Context, inputs map[string][]any) error {
	if len(inputs) == 0 {
		return errors.New("inputs cannot be empty")
	}

	for table, data := range inputs {
		if table == "" {
			return errors.New("table cannot be empty")
		}

		if err := c.Inserter(table).Put(ctx, data); err != nil {
			return errors.WithStack(err)
		}
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
	if err := c.client.Close(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *Config) Validate() error {
	if c == nil {
		return errors.New("databse configuration required")
	}
	if c.ProjectID == "" {
		return errors.New("database project ID required")
	}
	if c.DatasetID == "" {
		return errors.New("database dataset ID required")
	}
	if c.CredsPath == "" {
		return errors.New("databse creds path ID required")
	}
	return nil
}

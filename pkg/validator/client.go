package validator

import (
	"context"
	"crypto/tls"
	"fmt"
	"path/filepath"

	pb "github.com/grid-stream-org/grid-stream-protos/gen/validator/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type ValidatorClient interface {
	SendAverages(ctx context.Context, averages []*pb.AverageOutput) error
	NotifyProject(ctx context.Context, projectID string) error
	Close() error
}

type Config struct {
	Host      string     `koanf:"host" json:"host" envconfig:"host"`
	Port      int        `koanf:"port" json:"port" envconfig:"port"`
	TLSConfig *TLSConfig `koanf:"tls_config" json:"tls_config" envconfig:"tls_config"`
}

type TLSConfig struct {
	Enabled  bool   `koanf:"enabled" json:"enabled" envconfig:"enabled"`
	CertPath string `koanf:"cert_path" json:"cert_path" envconfig:"cert_path"`
	KeyPath  string `koanf:"key_path" json:"key_path" envconfig:"key_path"`
}

type validatorClient struct {
	cfg    *Config
	client pb.ValidatorServiceClient
	conn   *grpc.ClientConn
}

func (c *Config) Validate() error {
	if c.Port <= 0 {
		return errors.New("port must be greater than 0")
	}

	if c.TLSConfig != nil && c.TLSConfig.Enabled {
		if c.TLSConfig.CertPath == "" {
			return errors.New("cert_path required when tls is enabled")
		}
		if c.TLSConfig.KeyPath == "" {
			return errors.New("key_path required when tls is enabled")
		}
	}

	return nil
}

func New(ctx context.Context, cfg *Config) (ValidatorClient, error) {
	var opts []grpc.DialOption

	if cfg.TLSConfig != nil && cfg.TLSConfig.Enabled {
		cert, err := tls.LoadX509KeyPair(
			filepath.Clean(cfg.TLSConfig.CertPath),
			filepath.Clean(cfg.TLSConfig.KeyPath),
		)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		creds := credentials.NewTLS(&tls.Config{Certificates: []tls.Certificate{cert}})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	c := &validatorClient{
		cfg:    cfg,
		client: pb.NewValidatorServiceClient(conn),
		conn:   conn,
	}

	return c, nil
}

func (c *validatorClient) Close() error {
	return c.conn.Close()
}

func (c *validatorClient) SendAverages(ctx context.Context, averageOutputs []*pb.AverageOutput) error {
	req := &pb.ValidateAverageOutputsRequest{
		AverageOutputs: averageOutputs,
	}

	res, err := c.client.ValidateAverageOutputs(ctx, req)
	if err != nil {
		return errors.WithStack(err)
	}

	if !res.Success && len(res.Errors) > 0 {
		return &ValidationErrors{Errors: res.Errors}
	}
	return nil
}

func (c *validatorClient) NotifyProject(ctx context.Context, projectID string) error {
	req := &pb.NotifyProjectRequest{
		ProjectId: projectID,
	}

	res, err := c.client.NotifyProject(ctx, req)
	if err != nil {
		return errors.WithStack(err)
	}

	if !res.Acknowledged && len(res.Errors) > 0 {
		return &NotifyProjectErrors{Errors: res.Errors}
	}
	return nil
}

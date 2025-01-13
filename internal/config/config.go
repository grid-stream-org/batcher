package config

import (
	"path/filepath"
	"slices"
	"time"

	"github.com/grid-stream-org/batcher/pkg/bqclient"
	"github.com/grid-stream-org/batcher/pkg/logger"
	"github.com/grid-stream-org/batcher/pkg/validator"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
)

type Config struct {
	Batcher     *Batcher          `koanf:"batcher"`
	Pool        *Pool             `koanf:"pool"`
	Destination *Destination      `koanf:"destination"`
	Validator   *validator.Config `koanf:"validator"`
	MQTT        *MQTT             `koanf:"mqtt"`
	Log         *logger.Config    `koanf:"log"`
}

type Batcher struct {
	Timeout time.Duration `koanf:"timeout"`
}

type Pool struct {
	NumWorkers int `koanf:"num_workers"`
	Capacity   int `koanf:"capacity"`
}

type Destination struct {
	Type     string           `koanf:"type"`
	Path     string           `koanf:"path"`
	Buffer   *Buffer          `koanf:"buffer"`
	Database *bqclient.Config `koanf:"database"`
}

type Buffer struct {
	StartTime time.Time     `koanf:"start_time"`
	Interval  time.Duration `koanf:"interval"`
	Offset    time.Duration `koanf:"offset"`
}

type MQTT struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	QoS      int    `koanf:"qos"`
	MTLS     *MTLS  `koanf:"m_tls"`
}

type MTLS struct {
	Enabled  bool   `koanf:"enabled"`
	CertFile string `koanf:"cert_file"`
	KeyFile  string `koanf:"key_file"`
	CAFile   string `koanf:"ca_file"`
}

func Load() (*Config, error) {
	k := koanf.New(".")
	path := filepath.Join("configs", "config.json")
	if err := k.Load(file.Provider(path), json.Parser()); err != nil {
		return nil, errors.WithStack(err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, errors.WithStack(err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Pool.NumWorkers < 1 {
		c.Pool.NumWorkers = 1
	}
	if c.Pool.Capacity < 0 {
		c.Pool.Capacity = 0
	}

	if err := c.Destination.validate(); err != nil {
		return errors.WithStack(err)
	}
	if err := c.MQTT.validate(); err != nil {
		return errors.WithStack(err)
	}
	if err := c.Log.Validate(); err != nil {
		return errors.WithStack(err)
	}
	if err := c.Validator.Validate(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (d *Destination) validate() error {
	if d.Type == "" {
		return errors.New("destination type is required")
	}

	validTypes := []string{
		"file",
		"database",
		"stdout",
	}
	if !slices.Contains(validTypes, d.Type) {
		return errors.Errorf("invalid destination type: %s", d.Type)
	}

	if d.Type == "file" {
		if d.Path == "" {
			return errors.New("file path required when destination is 'file'")
		}
	}

	if d.Type == "database" {
		if err := d.Database.Validate(); err != nil {
			return errors.WithStack(err)
		}
	}

	if d.Buffer == nil {
		return errors.New("buffer configuration required when type is 'database'")
	}
	if d.Buffer.Interval <= 0 {
		return errors.New("buffer interval must be positive")
	}
	if d.Buffer.Offset < 0 {
		return errors.New("buffer offset cannot be negative")
	}
	if d.Buffer.Offset >= d.Buffer.Interval {
		return errors.New("buffer offset must be less than interval")
	}
	if d.Buffer.StartTime.IsZero() {
		return errors.New("buffer start time required")
	}

	return nil
}

func (m *MQTT) validate() error {
	if m.Port < 1 || m.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if m.QoS < 0 || m.QoS > 2 {
		return errors.New("qos must be between 0 and 2")
	}
	if m.MTLS.Enabled {
		if m.MTLS.CertFile == "" {
			return errors.New("cert_file required when mtls enabled")
		}
		if m.MTLS.KeyFile == "" {
			return errors.New("key_file required when mtls enabled")
		}
		if m.MTLS.CAFile == "" {
			return errors.New("ca_file required when mtls enabled")
		}
	}
	return nil
}

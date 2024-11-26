package config

import (
	"path/filepath"
	"time"

	"github.com/grid-stream-org/batcher/pkg/logger"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/pkg/errors"
)

type Config struct {
	Batcher     *Batcher       `koanf:"batcher"`
	Pool        *Pool          `koanf:"pool"`
	Destination *Destination   `koanf:"destination"`
	MQTT        *MQTT          `koanf:"mqtt"`
	Log         *logger.Config `koanf:"log"`
}

type Batcher struct {
	Timeout time.Duration `koanf:"timeout"`
}

type Pool struct {
	NumWorkers int `koanf:"num_workers"`
	Capacity   int `koanf:"capacity"`
}

type Destination struct {
	Type     string    `koanf:"type"`
	Path     string    `koanf:"path"`
	Buffer   *Buffer   `koanf:"buffer"`
	Database *Database `koanf:"database"`
}

type Buffer struct {
	Duration time.Duration `koanf:"duration"`
	Offset   time.Duration `koanf:"offset"`
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

type Database struct {
	ProjectID string `koanf:"project_id"`
	DatasetID string `koanf:"dataset_id"`
	CredsPath string `koanf:"creds_path"`
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
		return errors.Wrap(err, "destination config")
	}
	if err := c.MQTT.validate(); err != nil {
		return errors.Wrap(err, "mqtt config")
	}
	if err := c.Log.Validate(); err != nil {
		return errors.Wrap(err, "log config")
	}

	return nil
}

func (d *Destination) validate() error {
	if d.Type == "file" && d.Path == "" {
		return errors.New("file path required when type is 'file'")
	}
	if d.Type == "database" {
		if d.Database.ProjectID == "" {
			return errors.New("database config required when type is database")
		}
		if d.Buffer.Duration < 1 {
			return errors.New("buffer duration must be positive when buffer enabled")
		}
		if d.Buffer.Offset > d.Buffer.Duration {
			return errors.New("buffer duration must be greater than offset")
		}
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

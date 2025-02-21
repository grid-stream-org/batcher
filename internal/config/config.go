package config

import (
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/grid-stream-org/go-commons/pkg/bqclient"
	"github.com/grid-stream-org/go-commons/pkg/logger"
	"github.com/grid-stream-org/go-commons/pkg/validator"
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
	Type     string           `koanf:"type"`
	Buffer   *Buffer          `koanf:"buffer"`
	Database *bqclient.Config `koanf:"database"`
}

type Buffer struct {
	StartTime time.Time
	Interval  time.Duration     `koanf:"interval"`
	Offset    time.Duration     `koanf:"offset"`
	Validator *validator.Config `koanf:"validator"`
}

type MQTT struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
	QoS      int    `koanf:"qos"`
	Topic    string `koanf:"qos"`
}

func Load() (*Config, error) {
	k := koanf.New(".")

	path := os.Getenv("CONFIG_PATH")
	if path == "" {
		path = filepath.Join("configs", "config.json")
		logger.Default().Info("CONFIG_PATH not set, using default", "path", path)
	}
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

	if c.Batcher == nil {
		c.Batcher = &Batcher{Timeout: 0}
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

	return nil
}

func (d *Destination) validate() error {
	if d.Type == "" {
		return errors.New("destination type is required")
	}

	validTypes := []string{
		"event",
		"stdout",
		"stream",
	}
	if !slices.Contains(validTypes, d.Type) {
		return errors.Errorf("invalid destination type: %s", d.Type)
	}

	if d.Type == "event" {
		if err := d.Database.Validate(); err != nil {
			return errors.WithStack(err)
		}

		if err := d.Buffer.validate(); err != nil {
			return errors.WithStack(err)
		}
	}

	if d.Type == "stream" {
		if err := d.Database.Validate(); err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func (b *Buffer) validate() error {
	if b == nil {
		return errors.New("buffer configuration required")
	}

	startTime := os.Getenv("BUFFER_START_TIME")
	if startTime == "" {
		return errors.New("buffer start time not set in environment and is required")
	}

	t, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		return errors.WithStack(err)
	}

	b.StartTime = t

	if b.Interval <= 0 {
		return errors.New("buffer interval must be positive")
	}
	if b.Offset < 0 {
		return errors.New("buffer offset cannot be negative")
	}
	if b.Offset >= b.Interval {
		return errors.New("buffer offset must be less than interval")
	}
	if b.StartTime.IsZero() {
		return errors.New("buffer start time required")
	}

	if err := b.Validator.Validate(); err != nil {
		return errors.WithStack(err)
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

	if m.Topic == "" {
		return errors.New("topic is required")
	}

	return nil
}

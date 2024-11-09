package config

import (
	"log"
	"time"

	"github.com/pkg/errors"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MQTT       MQTTConfig   `envconfig:"MQTT"`
	BQ         BQConfig     `envconfig:"BQ"`
	Buffer     BufferConfig `envconfig:"BUFFER"`
	Logger     LoggerConfig `envconfig:"LOG"`
	NumWorkers int          `envconfig:"NUM_WORKERS" default:"2"`
}

type BufferConfig struct {
	Duration time.Duration `envconfig:"DURATION" default:"5m"`
	Offset   time.Duration `envconfig:"OFFSET" default:"1s"`
	Capacity int           `envconfig:"CAPACITY" default:"10"`
}

type MQTTConfig struct {
	Host          string `envconfig:"HOST" required:"true"`
	Port          int    `envconfig:"PORT" required:"true"`
	Username      string `envconfig:"USERNAME" required:"true"`
	Password      string `envconfig:"PASSWORD" required:"true"`
	QOS           int    `envconfig:"QOS" default:"1"`
	SkipVerify    bool   `envconfig:"SKIP_VERIFY" default:"false"`
	CertFile      string `envconfig:"CERT_FILE"`
	KeyFile       string `envconfig:"KEY_FILE"`
	CAFile        string `envconfig:"CA_FILE"`
	PartitionMode bool   `envconfig:"PARTITION_MODE" default:"false"`
	Partition     int    `envconfig:"PARTITION" default:"0"`
	NumPartitions int    `envconfig:"NUM_PARTITIONS" default:"1"`
}

type BQConfig struct {
	ProjectID string `envconfig:"PROJECT_ID" required:"true"`
	DatasetID string `envconfig:"DATASET_ID" required:"true"`
	CredsPath string `envconfig:"CREDENTIALS_PATH" required:"true"`
}

type LoggerConfig struct {
	Level  string `envconfig:"LEVEL" default:"INFO"`
	Format string `envconfig:"FORMAT" default:"text"`
	Output string `envconfig:"OUTPUT" default:"stdout"`
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env file not found, proceeding with environment variables")
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func validate(cfg *Config) error {
	if cfg.Buffer.Duration <= cfg.Buffer.Offset {
		return errors.New("buffer duration must be greater than offset")
	}

	if cfg.MQTT.PartitionMode && cfg.MQTT.NumPartitions <= 0 {
		return errors.New("num_partitions must be positive when partition_mode is enabled")
	}

	if cfg.MQTT.SkipVerify {
		if cfg.MQTT.CertFile != "" && cfg.MQTT.KeyFile == "" {
			return errors.New("key file required when cert file provided")
		}
		if cfg.MQTT.KeyFile != "" && cfg.MQTT.CertFile == "" {
			return errors.New("cert file required when key file provided")
		}
	}
	return nil
}

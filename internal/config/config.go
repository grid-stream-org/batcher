package config

import (
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MQTT       MQTTConfig   `envconfig:"MQTT"`
	DB         DBConfig     `envconfig:"DB"`
	Buffer     BufferConfig `envconfig:"BUFFER"`
	Logger     LoggerConfig `envconfig:"LOG"`
	NumWorkers int          `envconfig:"NUM_WORKERS" default:"2"`
	ProjectID  string       `envconfig:"PROJECT_ID" required:"true"`
}

type BufferConfig struct {
	Duration time.Duration `envconfig:"DURATION" default:"5m"`
	Offset   time.Duration `envconfig:"OFFSET" default:"1s"`
	Capacity int           `envconfig:"CAPACITY" default:"10"`
}

type MQTTConfig struct {
	Host     string `envconfig:"HOST" required:"true"`
	Port     int    `envconfig:"PORT" required:"true"`
	Username string `envconfig:"USERNAME" required:"true"`
	Password string `envconfig:"PASSWORD" required:"true"`
}

type DBConfig struct {
	ProjectID string `envconfig:"PROJECT_ID" required:"true"`
	DatasetID string `envconfig:"DATASET_ID" required:"true"`
	TableID   string `envconfig:"TABLE_ID" required:"true"`
	CredsPath string `envconfig:"CREDENTIALS_PATH" required:"true"`
}

type LoggerConfig struct {
	Level  string `envconfig:"LEVEL" default:"INFO"`
	Format string `envconfig:"FORMAT" default:"text"`
	Output string `envconfig:"OUTPUT" default:"stdout"`
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf(".env file not found, proceeding with environment variables")
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustLoadConfig() *Config {
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)

	}
	return cfg
}

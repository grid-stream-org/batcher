package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEnv(envVars map[string]string) func() {
	allEnvVars := []string{
		"NUM_WORKERS",
		"PROJECT_ID",
		"BUFFER_DURATION",
		"BUFFER_OFFSET",
		"MQTT_HOST",
		"MQTT_PORT",
		"MQTT_USERNAME",
		"MQTT_PASSWORD",
		"DB_PROJECT_ID",
		"DB_DATASET_ID",
		"DB_TABLE_ID",
		"DB_CREDENTIALS_PATH",
	}

	originalEnv := make(map[string]*string)

	for _, key := range allEnvVars {
		value, exists := os.LookupEnv(key)
		if exists {
			v := value
			originalEnv[key] = &v
			os.Unsetenv(key)
		} else {
			originalEnv[key] = nil
		}
	}

	for key, value := range envVars {
		os.Setenv(key, value)
	}

	return func() {
		for key, value := range originalEnv {
			if value != nil {
				os.Setenv(key, *value)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}

func TestLoadConfig_Success(t *testing.T) {
	teardown := setupEnv(map[string]string{
		"NUM_WORKERS":         "4",
		"PROJECT_ID":          "test-project",
		"BUFFER_DURATION":     "10m",
		"BUFFER_OFFSET":       "2s",
		"MQTT_HOST":           "localhost",
		"MQTT_PORT":           "1883",
		"MQTT_USERNAME":       "user",
		"MQTT_PASSWORD":       "pass",
		"DB_PROJECT_ID":       "db-project",
		"DB_DATASET_ID":       "db-dataset",
		"DB_TABLE_ID":         "db-table",
		"DB_CREDENTIALS_PATH": "/path/to/creds.json",
	})
	defer teardown()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 4, cfg.NumWorkers)
	assert.Equal(t, "test-project", cfg.ProjectID)
	assert.Equal(t, 10*time.Minute, cfg.Buffer.Duration)
	assert.Equal(t, 2*time.Second, cfg.Buffer.Offset)
	assert.Equal(t, "localhost", cfg.MQTT.Host)
	assert.Equal(t, 1883, cfg.MQTT.Port)
	assert.Equal(t, "user", cfg.MQTT.Username)
	assert.Equal(t, "pass", cfg.MQTT.Password)
	assert.Equal(t, "db-project", cfg.DB.ProjectID)
	assert.Equal(t, "db-dataset", cfg.DB.DatasetID)
	assert.Equal(t, "db-table", cfg.DB.TableID)
	assert.Equal(t, "/path/to/creds.json", cfg.DB.CredsPath)
}

func TestLoadConfig_MissingRequiredEnv(t *testing.T) {
	teardown := setupEnv(map[string]string{
		"NUM_WORKERS": "2",
	})
	defer teardown()

	_, err := LoadConfig()
	assert.Error(t, err)
}

func TestLoadConfig_DefaultValues(t *testing.T) {
	teardown := setupEnv(map[string]string{
		"PROJECT_ID":          "default-project",
		"MQTT_HOST":           "localhost",
		"MQTT_PORT":           "1883",
		"MQTT_USERNAME":       "user",
		"MQTT_PASSWORD":       "pass",
		"DB_PROJECT_ID":       "db-project",
		"DB_DATASET_ID":       "db-dataset",
		"DB_TABLE_ID":         "db-table",
		"DB_CREDENTIALS_PATH": "/path/to/creds.json",
	})
	defer teardown()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 2, cfg.NumWorkers)
	assert.Equal(t, 5*time.Minute, cfg.Buffer.Duration)
	assert.Equal(t, 1*time.Second, cfg.Buffer.Offset)
}

func TestLoadConfig_InvalidDuration(t *testing.T) {
	teardown := setupEnv(map[string]string{
		"PROJECT_ID":          "test-project",
		"MQTT_HOST":           "localhost",
		"MQTT_PORT":           "1883",
		"MQTT_USERNAME":       "user",
		"MQTT_PASSWORD":       "pass",
		"DB_PROJECT_ID":       "db-project",
		"DB_DATASET_ID":       "db-dataset",
		"DB_TABLE_ID":         "db-table",
		"DB_CREDENTIALS_PATH": "/path/to/creds.json",
		"BUFFER_DURATION":     "invalid",
	})
	defer teardown()

	_, err := LoadConfig()
	require.Error(t, err)
}

func TestLoadConfig_EmptyDuration(t *testing.T) {
	teardown := setupEnv(map[string]string{
		"PROJECT_ID":          "test-project",
		"MQTT_HOST":           "localhost",
		"MQTT_PORT":           "1883",
		"MQTT_USERNAME":       "user",
		"MQTT_PASSWORD":       "pass",
		"DB_PROJECT_ID":       "db-project",
		"DB_DATASET_ID":       "db-dataset",
		"DB_TABLE_ID":         "db-table",
		"DB_CREDENTIALS_PATH": "/path/to/creds.json",
		"BUFFER_DURATION":     "",
	})
	defer teardown()

	_, err := LoadConfig()
	require.Error(t, err)
}

func TestLoadConfig_OverrideDefaults(t *testing.T) {
	teardown := setupEnv(map[string]string{
		"PROJECT_ID":          "test-project",
		"NUM_WORKERS":         "10",
		"BUFFER_DURATION":     "15m",
		"BUFFER_OFFSET":       "5s",
		"MQTT_HOST":           "mqtt.example.com",
		"MQTT_PORT":           "8883",
		"MQTT_USERNAME":       "mqttuser",
		"MQTT_PASSWORD":       "mqttpass",
		"DB_PROJECT_ID":       "db-project",
		"DB_DATASET_ID":       "db-dataset",
		"DB_TABLE_ID":         "db-table",
		"DB_CREDENTIALS_PATH": "/path/to/creds.json",
	})
	defer teardown()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 10, cfg.NumWorkers)
	assert.Equal(t, 15*time.Minute, cfg.Buffer.Duration)
	assert.Equal(t, 5*time.Second, cfg.Buffer.Offset)
	assert.Equal(t, "mqtt.example.com", cfg.MQTT.Host)
	assert.Equal(t, 8883, cfg.MQTT.Port)
	assert.Equal(t, "mqttuser", cfg.MQTT.Username)
	assert.Equal(t, "mqttpass", cfg.MQTT.Password)
}

func TestLoadConfig_InvalidPort(t *testing.T) {
	teardown := setupEnv(map[string]string{
		"PROJECT_ID":          "test-project",
		"MQTT_HOST":           "localhost",
		"MQTT_PORT":           "invalid",
		"MQTT_USERNAME":       "user",
		"MQTT_PASSWORD":       "pass",
		"DB_PROJECT_ID":       "db-project",
		"DB_DATASET_ID":       "db-dataset",
		"DB_TABLE_ID":         "db-table",
		"DB_CREDENTIALS_PATH": "/path/to/creds.json",
	})
	defer teardown()

	_, err := LoadConfig()
	require.Error(t, err)
}

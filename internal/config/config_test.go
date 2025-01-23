package config

import (
	"testing"
	"time"

	"github.com/grid-stream-org/batcher/pkg/bqclient"
	"github.com/grid-stream-org/batcher/pkg/logger"
	"github.com/grid-stream-org/batcher/pkg/validator"
	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (s *ConfigTestSuite) newValidConfig() *Config {
	return &Config{
		Pool: &Pool{
			NumWorkers: 4,
			Capacity:   100,
		},
		Destination: &Destination{
			Type: "database",
			Database: &bqclient.Config{
				ProjectID: "test-project",
				DatasetID: "test-dataset",
				CredsPath: "test-creds.json",
			},
			Buffer: &Buffer{
				StartTime: time.Now(),
				Interval:  time.Minute,
				Offset:    time.Second * 30,
			},
		},
		MQTT: &MQTT{
			Host:     "localhost",
			Port:     1883,
			Username: "user",
			Password: "pass",
			QoS:      1,
		},
		Log: &logger.Config{
			Level:  "INFO",
			Format: "json",
		},
		Validator: &validator.Config{
			Host: "localhost",
			Port: 8080,
		},
	}
}

func (s *ConfigTestSuite) TestConfigValidation() {
	testCases := []struct {
		name        string
		modify      func(*Config)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			modify:      func(c *Config) {},
			expectError: false,
		},
		{
			name: "zero workers gets set to 1",
			modify: func(c *Config) {
				c.Pool.NumWorkers = 0
			},
			expectError: false,
		},
		{
			name: "negative capacity gets set to 0",
			modify: func(c *Config) {
				c.Pool.Capacity = -1
			},
			expectError: false,
		},
		{
			name: "invalid destination",
			modify: func(c *Config) {
				c.Destination.Type = ""
			},
			expectError: true,
			errorMsg:    "destination type is required",
		},
		{
			name: "invalid mqtt",
			modify: func(c *Config) {
				c.MQTT.Port = 0
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cfg := s.newValidConfig()
			tc.modify(cfg)
			err := cfg.Validate()
			if tc.expectError {
				s.Error(err)
				if tc.errorMsg != "" {
					s.Contains(err.Error(), tc.errorMsg)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ConfigTestSuite) TestDestinationValidation() {
	testCases := []struct {
		name        string
		modify      func(*Destination)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid database config",
			modify:      func(d *Destination) {},
			expectError: false,
		},
		{
			name: "empty type",
			modify: func(d *Destination) {
				d.Type = ""
			},
			expectError: true,
			errorMsg:    "destination type is required",
		},
		{
			name: "invalid type",
			modify: func(d *Destination) {
				d.Type = "invalid"
			},
			expectError: true,
			errorMsg:    "invalid destination type: invalid",
		},
		{
			name: "file type without path",
			modify: func(d *Destination) {
				d.Type = "file"
				d.Path = ""
			},
			expectError: true,
			errorMsg:    "file path required when destination is 'file'",
		},
		{
			name: "database type without buffer",
			modify: func(d *Destination) {
				d.Type = "database"
				d.Buffer = nil
			},
			expectError: true,
			errorMsg:    "buffer configuration required when type is 'database'",
		},
		{
			name: "buffer zero interval",
			modify: func(d *Destination) {
				d.Type = "database"
				d.Buffer.Interval = 0
			},
			expectError: true,
			errorMsg:    "buffer interval must be positive",
		},
		{
			name: "buffer negative offset",
			modify: func(d *Destination) {
				d.Type = "database"
				d.Buffer.Offset = -1 * time.Second
			},
			expectError: true,
			errorMsg:    "buffer offset cannot be negative",
		},
		{
			name: "buffer offset equals interval",
			modify: func(d *Destination) {
				d.Type = "database"
				d.Buffer.Interval = time.Second
				d.Buffer.Offset = time.Second
			},
			expectError: true,
			errorMsg:    "buffer offset must be less than interval",
		},
		{
			name: "buffer zero start time",
			modify: func(d *Destination) {
				d.Type = "database"
				d.Buffer.StartTime = time.Time{}
			},
			expectError: true,
			errorMsg:    "buffer start time required",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			dest := &Destination{
				Type: "database",
				Database: &bqclient.Config{
					ProjectID: "test-project",
					DatasetID: "test-dataset",
					CredsPath: "test-creds.json",
				},
				Buffer: &Buffer{
					StartTime: time.Now(),
					Interval:  time.Minute,
					Offset:    time.Second * 30,
				},
			}
			tc.modify(dest)
			err := dest.validate()
			if tc.expectError {
				s.Error(err)
				if tc.errorMsg != "" {
					s.Contains(err.Error(), tc.errorMsg)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ConfigTestSuite) TestMQTTValidation() {
	testCases := []struct {
		name        string
		modify      func(*MQTT)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid mqtt config",
			modify:      func(m *MQTT) {},
			expectError: false,
		},
		{
			name: "port too low",
			modify: func(m *MQTT) {
				m.Port = 0
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "port too high",
			modify: func(m *MQTT) {
				m.Port = 65536
			},
			expectError: true,
			errorMsg:    "port must be between 1 and 65535",
		},
		{
			name: "QoS too low",
			modify: func(m *MQTT) {
				m.QoS = -1
			},
			expectError: true,
			errorMsg:    "qos must be between 0 and 2",
		},
		{
			name: "QoS too high",
			modify: func(m *MQTT) {
				m.QoS = 3
			},
			expectError: true,
			errorMsg:    "qos must be between 0 and 2",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			mqtt := &MQTT{
				Host:     "localhost",
				Port:     1883,
				Username: "user",
				Password: "pass",
				QoS:      1,
			}
			tc.modify(mqtt)
			err := mqtt.validate()
			if tc.expectError {
				s.Error(err)
				if tc.errorMsg != "" {
					s.Contains(err.Error(), tc.errorMsg)
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

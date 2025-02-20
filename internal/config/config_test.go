package config

// import (
// 	"os"
// 	"testing"
// 	"time"

// 	"github.com/grid-stream-org/go-commons/pkg/bqclient"
// 	"github.com/grid-stream-org/go-commons/pkg/logger"
// 	"github.com/grid-stream-org/go-commons/pkg/validator"
// 	"github.com/stretchr/testify/suite"
// )

// type ConfigTestSuite struct {
// 	suite.Suite
// }

// func (s *ConfigTestSuite) SetupTest() {
// 	startTime := time.Now().Format(time.RFC3339)
// 	os.Setenv("BUFFER_START_TIME", startTime)
// }

// func (s *ConfigTestSuite) TearDownTest() {
// 	os.Unsetenv("BUFFER_START_TIME")
// }

// func (s *ConfigTestSuite) newValidConfig() *Config {
// 	return &Config{
// 		Batcher: &Batcher{
// 			Timeout: time.Minute * 5,
// 		},
// 		Pool: &Pool{
// 			NumWorkers: 4,
// 			Capacity:   100,
// 		},
// 		Destination: &Destination{
// 			Type: "event",
// 			Database: &bqclient.Config{
// 				ProjectID: "test-project",
// 				DatasetID: "test-dataset",
// 				CredsPath: "test-creds.json",
// 			},
// 			Buffer: &Buffer{
// 				Interval: time.Minute,
// 				Offset:   time.Second * 30,
// 				Validator: &validator.Config{
// 					Host: "localhost",
// 					Port: 8080,
// 				},
// 			},
// 		},
// 		MQTT: &MQTT{
// 			Host:     "localhost",
// 			Port:     1883,
// 			Username: "user",
// 			Password: "pass",
// 			QoS:      1,
// 		},
// 		Log: &logger.Config{
// 			Level:  "INFO",
// 			Format: "json",
// 		},
// 	}
// }

// func (s *ConfigTestSuite) TestConfigValidation() {
// 	testCases := []struct {
// 		name        string
// 		modify      func(*Config)
// 		expectError bool
// 		errorMsg    string
// 	}{
// 		{
// 			name:        "valid config",
// 			modify:      func(c *Config) {},
// 			expectError: false,
// 		},
// 		{
// 			name: "zero workers gets set to 1",
// 			modify: func(c *Config) {
// 				c.Pool.NumWorkers = 0
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "negative capacity gets set to 0",
// 			modify: func(c *Config) {
// 				c.Pool.Capacity = -1
// 			},
// 			expectError: false,
// 		},
// 		{
// 			name: "invalid destination",
// 			modify: func(c *Config) {
// 				c.Destination.Type = ""
// 			},
// 			expectError: true,
// 			errorMsg:    "destination type is required",
// 		},
// 		{
// 			name: "invalid mqtt",
// 			modify: func(c *Config) {
// 				c.MQTT.Port = 0
// 			},
// 			expectError: true,
// 			errorMsg:    "port must be between 1 and 65535",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			cfg := s.newValidConfig()
// 			tc.modify(cfg)
// 			err := cfg.Validate()
// 			if tc.expectError {
// 				s.Error(err)
// 				if tc.errorMsg != "" {
// 					s.Contains(err.Error(), tc.errorMsg)
// 				}
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }

// func (s *ConfigTestSuite) TestDestinationValidation() {
// 	testCases := []struct {
// 		name        string
// 		modify      func(*Destination)
// 		setupEnv    func()
// 		expectError bool
// 		errorMsg    string
// 	}{
// 		{
// 			name:        "valid database config",
// 			modify:      func(d *Destination) {},
// 			setupEnv:    func() {},
// 			expectError: false,
// 		},
// 		{
// 			name:   "missing start time env var",
// 			modify: func(d *Destination) {},
// 			setupEnv: func() {
// 				os.Unsetenv("BUFFER_START_TIME")
// 			},
// 			expectError: true,
// 			errorMsg:    "buffer start time not set in environment and is required",
// 		},
// 		{
// 			name:   "invalid start time format",
// 			modify: func(d *Destination) {},
// 			setupEnv: func() {
// 				os.Setenv("BUFFER_START_TIME", "invalid-time")
// 			},
// 			expectError: true,
// 			errorMsg:    "parsing time",
// 		},
// 		{
// 			name: "empty type",
// 			modify: func(d *Destination) {
// 				d.Type = ""
// 			},
// 			setupEnv:    func() {},
// 			expectError: true,
// 			errorMsg:    "destination type is required",
// 		},
// 		{
// 			name: "invalid type",
// 			modify: func(d *Destination) {
// 				d.Type = "invalid"
// 			},
// 			setupEnv:    func() {},
// 			expectError: true,
// 			errorMsg:    "invalid destination type: invalid",
// 		},
// 		{
// 			name: "stream type without database",
// 			modify: func(d *Destination) {
// 				d.Type = "stream"
// 				d.Database = nil
// 			},
// 			setupEnv:    func() {},
// 			expectError: true,
// 			errorMsg:    "database configuration required",
// 		},
// 		{
// 			name: "event type without buffer",
// 			modify: func(d *Destination) {
// 				d.Type = "event"
// 				d.Buffer = nil
// 			},
// 			setupEnv:    func() {},
// 			expectError: true,
// 			errorMsg:    "buffer configuration required",
// 		},
// 		{
// 			name: "buffer zero interval",
// 			modify: func(d *Destination) {
// 				d.Type = "event"
// 				d.Buffer.Interval = 0
// 			},
// 			setupEnv:    func() {},
// 			expectError: true,
// 			errorMsg:    "buffer interval must be positive",
// 		},
// 		{
// 			name: "buffer negative offset",
// 			modify: func(d *Destination) {
// 				d.Type = "event"
// 				d.Buffer.Offset = -1 * time.Second
// 			},
// 			setupEnv:    func() {},
// 			expectError: true,
// 			errorMsg:    "buffer offset cannot be negative",
// 		},
// 		{
// 			name: "buffer offset equals interval",
// 			modify: func(d *Destination) {
// 				d.Type = "event"
// 				d.Buffer.Interval = time.Second
// 				d.Buffer.Offset = time.Second
// 			},
// 			setupEnv:    func() {},
// 			expectError: true,
// 			errorMsg:    "buffer offset must be less than interval",
// 		},
// 		{
// 			name: "stdout type is valid",
// 			modify: func(d *Destination) {
// 				d.Type = "stdout"
// 				d.Buffer = nil
// 				d.Database = nil
// 			},
// 			setupEnv:    func() {},
// 			expectError: false,
// 		},
// 		{
// 			name: "stream type is valid",
// 			modify: func(d *Destination) {
// 				d.Type = "stream"
// 				d.Buffer = nil
// 			},
// 			setupEnv:    func() {},
// 			expectError: false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			tc.setupEnv()

// 			dest := &Destination{
// 				Type: "event",
// 				Database: &bqclient.Config{
// 					ProjectID: "test-project",
// 					DatasetID: "test-dataset",
// 					CredsPath: "test-creds.json",
// 				},
// 				Buffer: &Buffer{
// 					Interval: time.Minute,
// 					Offset:   time.Second * 30,
// 					Validator: &validator.Config{
// 						Host: "localhost",
// 						Port: 8080,
// 					},
// 				},
// 			}
// 			tc.modify(dest)
// 			err := dest.validate()
// 			if tc.expectError {
// 				s.Error(err)
// 				if tc.errorMsg != "" {
// 					s.Contains(err.Error(), tc.errorMsg)
// 				}
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }

// func (s *ConfigTestSuite) TestMQTTValidation() {
// 	testCases := []struct {
// 		name        string
// 		modify      func(*MQTT)
// 		expectError bool
// 		errorMsg    string
// 	}{
// 		{
// 			name:        "valid mqtt config",
// 			modify:      func(m *MQTT) {},
// 			expectError: false,
// 		},
// 		{
// 			name: "port too low",
// 			modify: func(m *MQTT) {
// 				m.Port = 0
// 			},
// 			expectError: true,
// 			errorMsg:    "port must be between 1 and 65535",
// 		},
// 		{
// 			name: "port too high",
// 			modify: func(m *MQTT) {
// 				m.Port = 65536
// 			},
// 			expectError: true,
// 			errorMsg:    "port must be between 1 and 65535",
// 		},
// 		{
// 			name: "QoS too low",
// 			modify: func(m *MQTT) {
// 				m.QoS = -1
// 			},
// 			expectError: true,
// 			errorMsg:    "qos must be between 0 and 2",
// 		},
// 		{
// 			name: "QoS too high",
// 			modify: func(m *MQTT) {
// 				m.QoS = 3
// 			},
// 			expectError: true,
// 			errorMsg:    "qos must be between 0 and 2",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		s.Run(tc.name, func() {
// 			mqtt := &MQTT{
// 				Host:     "localhost",
// 				Port:     1883,
// 				Username: "user",
// 				Password: "pass",
// 				QoS:      1,
// 			}
// 			tc.modify(mqtt)
// 			err := mqtt.validate()
// 			if tc.expectError {
// 				s.Error(err)
// 				if tc.errorMsg != "" {
// 					s.Contains(err.Error(), tc.errorMsg)
// 				}
// 			} else {
// 				s.NoError(err)
// 			}
// 		})
// 	}
// }

// func TestConfigSuite(t *testing.T) {
// 	suite.Run(t, new(ConfigTestSuite))
// }

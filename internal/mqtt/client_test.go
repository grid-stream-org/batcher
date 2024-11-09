package mqtt

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log/slog"
// 	"testing"
// 	"time"

// 	mqtt "github.com/eclipse/paho.mqtt.golang"
// 	"github.com/grid-stream-org/batcher/internal/buffer"
// 	"github.com/grid-stream-org/batcher/internal/config"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"
// )

// const testTimeout = 500 * time.Millisecond

// type MockMessage struct {
// 	payload []byte
// 	topic   string
// }

// func (m *MockMessage) Duplicate() bool   { return false }
// func (m *MockMessage) Qos() byte         { return 0 }
// func (m *MockMessage) Retained() bool    { return false }
// func (m *MockMessage) Topic() string     { return m.topic }
// func (m *MockMessage) MessageID() uint16 { return 0 }
// func (m *MockMessage) Payload() []byte   { return m.payload }
// func (m *MockMessage) Ack()              {}

// type MockMQTTClient struct {
// 	mock.Mock
// }

// func (m *MockMQTTClient) Connect() mqtt.Token {
// 	args := m.Called()
// 	return args.Get(0).(mqtt.Token)
// }

// func (m *MockMQTTClient) Disconnect(quiesce uint) {
// 	m.Called(quiesce)
// }

// func (m *MockMQTTClient) Subscribe(topic string, qos byte, callback mqtt.MessageHandler) mqtt.Token {
// 	args := m.Called(topic, qos, callback)
// 	return args.Get(0).(mqtt.Token)
// }

// func (m *MockMQTTClient) Unsubscribe(topics ...string) mqtt.Token {
// 	args := m.Called(topics)
// 	return args.Get(0).(mqtt.Token)
// }

// type MockToken struct {
// 	mock.Mock
// }

// func (m *MockToken) Wait() bool {
// 	args := m.Called()
// 	return args.Bool(0)
// }

// func (m *MockToken) WaitTimeout(duration time.Duration) bool {
// 	args := m.Called(duration)
// 	return args.Bool(0)
// }

// func (m *MockToken) Error() error {
// 	args := m.Called()
// 	return args.Error(0)
// }

// func setupTestLogger() *slog.Logger {
// 	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
// 		Level: slog.LevelDebug,
// 	}))
// }

// func setupTestBuffer(t *testing.T, ctx context.Context) *buffer.Buffer {
// 	buf, err := buffer.New(ctx, &config.BufferConfig{
// 		Duration: 100 * time.Millisecond,
// 		Offset:   10 * time.Millisecond,
// 		Capacity: 10,
// 	}, setupTestLogger())
// 	require.NoError(t, err)
// 	return buf
// }

// func TestNewClient(t *testing.T) {
// 	_, cancel := context.WithTimeout(context.Background(), testTimeout)
// 	defer cancel()

// 	tests := []struct {
// 		name string
// 		cfg  config.MQTTConfig
// 	}{
// 		{
// 			name: "successful connection",
// 			cfg: config.MQTTConfig{
// 				Host:     "test.mosquitto.org",
// 				Port:     8883,
// 				Username: "test",
// 				Password: "test",
// 				QOS:      1,
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			opts, err := createClientOptions(&tt.cfg, setupTestLogger())
// 			assert.NoError(t, err)

// 			assert.NotNil(t, opts)

// 			assert.True(t, opts.AutoReconnect)
// 			assert.True(t, opts.CleanSession)
// 			assert.Equal(t, tt.cfg.Username, opts.Username)
// 			assert.Equal(t, tt.cfg.Password, opts.Password)
// 			assert.Equal(t, int64(30), opts.KeepAlive)

// 			expectedBrokerURL := fmt.Sprintf("ssl://%s:%d", tt.cfg.Host, tt.cfg.Port)
// 			assert.Equal(t, expectedBrokerURL, opts.Servers[0].String())
// 		})
// 	}
// }

// func TestClient_HandleMessage(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
// 	defer cancel()

// 	buf := setupTestBuffer(t, ctx)
// 	defer buf.Stop()

// 	client := &Client{
// 		buffer: buf,
// 		log:    setupTestLogger(),
// 	}

// 	tests := []struct {
// 		name           string
// 		payload        []DER
// 		expectDropped  bool
// 		expectBuffered bool
// 	}{
// 		{
// 			name: "valid message",
// 			payload: []DER{
// 				{
// 					DerID:         "1",
// 					ProjectID:     "project1",
// 					Type:          "solar",
// 					CurrentOutput: 100,
// 				},
// 			},
// 			expectDropped:  false,
// 			expectBuffered: true,
// 		},
// 		{
// 			name:           "empty DER array",
// 			payload:        []DER{},
// 			expectDropped:  true,
// 			expectBuffered: false,
// 		},
// 		{
// 			name: "multiple DERs",
// 			payload: []DER{
// 				{
// 					DerID:     "1",
// 					ProjectID: "project1",
// 				},
// 				{
// 					DerID:     "2",
// 					ProjectID: "project1",
// 				},
// 			},
// 			expectDropped:  false,
// 			expectBuffered: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			payloadBytes, err := json.Marshal(tt.payload)
// 			require.NoError(t, err)

// 			msg := &MockMessage{
// 				payload: payloadBytes,
// 				topic:   "projects/test/data",
// 			}

// 			client.handleMessage(nil, msg)
// 		})
// 	}
// }

// func TestGetTopic(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		cfg      config.MQTTConfig
// 		expected string
// 	}{
// 		{
// 			name: "non-partitioned mode",
// 			cfg: config.MQTTConfig{
// 				PartitionMode: false,
// 			},
// 			expected: "projects/+/data",
// 		},
// 		{
// 			name: "partitioned mode",
// 			cfg: config.MQTTConfig{
// 				PartitionMode: true,
// 				Partition:     1,
// 			},
// 			expected: "projects/1/+/data",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			topic := getTopic(&tt.cfg)
// 			assert.Equal(t, tt.expected, topic)
// 		})
// 	}
// }

// func TestCreateClientOptions(t *testing.T) {
// 	logger := setupTestLogger()

// 	tests := []struct {
// 		name string
// 		cfg  config.MQTTConfig
// 	}{
// 		{
// 			name: "valid config",
// 			cfg: config.MQTTConfig{
// 				Host:     "test.mosquitto.org",
// 				Port:     8883,
// 				Username: "test",
// 				Password: "test",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			opts, err := createClientOptions(&tt.cfg, logger)
// 			assert.NoError(t, err)
// 			assert.NotNil(t, opts)
// 			assert.True(t, opts.AutoReconnect)
// 			assert.True(t, opts.CleanSession)
// 			assert.Equal(t, tt.cfg.Username, opts.Username)
// 			assert.Equal(t, tt.cfg.Password, opts.Password)
// 			assert.Equal(t, int64(30), opts.KeepAlive)
// 		})
// 	}
// }

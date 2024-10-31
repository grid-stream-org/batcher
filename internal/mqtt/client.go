package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
)

type DER struct {
	DerID             string    `json:"derId"`
	Type              string    `json:"type"`
	IsOnline          bool      `json:"isOnline"`
	Timestamp         time.Time `json:"Timestamp"`
	CurrentOutput     float64   `json:"currentOutput"`
	Units             string    `json:"units"`
	ProjectID         string    `json:"projectId"`
	UtilityID         string    `json:"utilityId"`
	IsStandalone      bool      `json:"isStandalone"`
	ConnectionStartAt string    `json:"connectionStartAt"`
	CurrentSoc        int       `json:"currentSoc"`
}

type Metrics struct {
	MessagesReceived int64
	MessagesDropped  int64
	ConnectedAt      time.Time
	LastMessageAt    time.Time
	ActiveProjects   map[string]time.Time // projectID -> last seen
}

type Client struct {
	client  mqtt.Client
	buffer  *buffer.Buffer
	topic   string
	cfg     *config.MQTTConfig
	log     *slog.Logger
	metrics *Metrics
}

func NewClient(ctx context.Context, cfg *config.MQTTConfig, buf *buffer.Buffer, log *slog.Logger) (*Client, error) {
	opts := createClientOptions(cfg, log)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, errors.Wrap(token.Error(), "connecting to broker")
	}

	m := &Client{
		client: client,
		buffer: buf,
		topic:  getTopic(cfg),
		cfg:    cfg,
		log:    log.With("component", "mqtt"),
		metrics: &Metrics{
			ConnectedAt:    time.Now(),
			ActiveProjects: make(map[string]time.Time),
		},
	}

	if err := m.subscribe(); err != nil {
		client.Disconnect(250)
		return nil, errors.Wrap(err, "subscribing to topics")
	}

	return m, nil
}

func createClientOptions(cfg *config.MQTTConfig, log *slog.Logger) *mqtt.ClientOptions {
	clientID := fmt.Sprintf("batcher-%s", uuid.NewString())
	brokerURL := fmt.Sprintf("ssl://%s:%d", cfg.Host, cfg.Port)

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetMaxReconnectInterval(5 * time.Minute).
		SetWriteTimeout(10 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Info("Connected to MQTT broker",
				"client_id", clientID,
				"broker", cfg.Host)
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Error("Lost connection to MQTT broker", "error", err)
		}).
		SetReconnectingHandler(func(c mqtt.Client, options *mqtt.ClientOptions) {
			log.Warn("Attempting to reconnect to MQTT broker")
		})

	return opts
}

func (c *Client) subscribe() error {
	token := c.client.Subscribe(c.topic, byte(c.cfg.QOS), c.handleMessage)
	if token.Wait() && token.Error() != nil {
		return errors.Wrapf(token.Error(), "subscribing to topic %s", c.topic)
	}

	c.log.Info("Subscribed to topic pattern",
		"topic", c.topic,
		"partition_mode", c.cfg.PartitionMode,
		"partition", c.cfg.Partition)

	return nil
}

func getTopic(cfg *config.MQTTConfig) string {
	if !cfg.PartitionMode {
		return "projects/+/data"
	}
	return fmt.Sprintf("projects/%d/+/data", cfg.Partition)
}

func (c *Client) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	atomic.AddInt64(&c.metrics.MessagesReceived, 1)
	c.metrics.LastMessageAt = time.Now()

	var ders []DER
	if err := json.Unmarshal(msg.Payload(), &ders); err != nil {
		atomic.AddInt64(&c.metrics.MessagesDropped, 1)
		c.log.Error("Failed to parse message payload",
			"error", err,
			"topic", msg.Topic())
		return
	}

	if len(ders) == 0 {
		atomic.AddInt64(&c.metrics.MessagesDropped, 1)
		c.log.Warn("Received empty DER array", "topic", msg.Topic())
		return
	}

	projectID := ders[0].ProjectID
	c.updateProject(projectID)

	if err := c.buffer.Add(msg.Payload()); err != nil {
		atomic.AddInt64(&c.metrics.MessagesDropped, 1)
		c.log.Error("Failed to buffer message",
			"project_id", projectID,
			"error", err)
		return
	}

	c.log.Debug("Message processed",
		"project_id", projectID,
		"der_count", len(ders),
		"bytes", len(msg.Payload()))
}

func (c *Client) updateProject(projectID string) {
	c.metrics.ActiveProjects[projectID] = time.Now()
	c.cleanupProjects()
}

func (c *Client) cleanupProjects() {
	threshold := time.Now().Add(-1 * time.Hour)
	for id, lastSeen := range c.metrics.ActiveProjects {
		if lastSeen.Before(threshold) {
			delete(c.metrics.ActiveProjects, id)
		}
	}
}

func (c *Client) Stop() {
	if token := c.client.Unsubscribe(c.topic); token.Wait() && token.Error() != nil {
		c.log.Error("Failed to unsubscribe", "error", token.Error())
	}

	c.client.Disconnect(250)
	c.log.Info("MQTT client stopped")
}

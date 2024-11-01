package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/grid-stream-org/batcher/internal/buffer"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
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

type Client struct {
	client mqtt.Client
	buffer *buffer.Buffer
	topic  string
	cfg    *config.MQTTConfig
	log    *slog.Logger
}

func NewClient(ctx context.Context, cfg *config.MQTTConfig, buf *buffer.Buffer, log *slog.Logger) (*Client, error) {
	m := &Client{
		buffer: buf,
		topic:  getTopic(cfg),
		cfg:    cfg,
		log:    log.With("component", "mqtt"),
	}

	opts, err := createClientOptions(cfg, log)
	if err != nil {
		return nil, errors.Wrap(err, "creating client options")
	}

	m.client = mqtt.NewClient(opts)
	m.log.Debug("Attempting to connect to MQTT broker...")

	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		log.Error("Error connecting to MQTT client", "error", err)
		return nil, errors.Wrap(token.Error(), "error connecting to broker")
	}

	if err := m.subscribe(); err != nil {
		m.client.Disconnect(250)
		return nil, errors.Wrap(err, "subscribing to topics")
	}
	return m, nil
}

func createClientOptions(cfg *config.MQTTConfig, log *slog.Logger) (*mqtt.ClientOptions, error) {
	clientID := fmt.Sprintf("batcher-%s", uuid.NewString())
	brokerURL := fmt.Sprintf("tls://%s:%d", cfg.Host, cfg.Port)

	log.Debug("Creating client options", "client_id", clientID, "broker", brokerURL)
	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetProtocolVersion(4).
		SetConnectTimeout(10 * time.Second).
		SetKeepAlive(30 * time.Second).
		SetWriteTimeout(10 * time.Second).
		SetOnConnectHandler(func(c mqtt.Client) {
			metrics.Local.Gauge(metrics.ConnectionStatus).WithLabelValues().Set(1)
			log.Info("Connected to MQTT broker", "client_id", clientID)
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			metrics.Local.Gauge(metrics.ConnectionStatus).WithLabelValues().Set(0)
			log.Error("Lost connection to MQTT broker", "error", err)
		}).
		SetReconnectingHandler(func(c mqtt.Client, options *mqtt.ClientOptions) {
			log.Warn("Attempting to reconnect to MQTT broker")
		})

	tlsCfg := &tls.Config{
		InsecureSkipVerify: cfg.SkipVerify,
	}

	if cfg.SkipVerify {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, errors.Wrap(err, "loading client certificates")
		}
		tlsCfg.Certificates = []tls.Certificate{cert}

		caCert, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, errors.Wrap(err, "loading CA certificate")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to parse CA certificate")
		}
		tlsCfg.RootCAs = caCertPool
	}
	opts.SetTLSConfig(tlsCfg)

	return opts, nil
}

func (m *Client) subscribe() error {
	m.log.Debug("Subscribing to topic", "topic", m.topic)
	token := m.client.Subscribe(m.topic, byte(m.cfg.QOS), m.handleMessage)
	if token.Wait() && token.Error() != nil {
		return errors.Wrapf(token.Error(), "subscribing to topic %s", m.topic)
	}

	m.log.Info("Successfully subscribed to topic", "topic", m.topic, "QOS Level", m.cfg.QOS)
	return nil
}

func getTopic(cfg *config.MQTTConfig) string {
	if cfg.PartitionMode {
		return fmt.Sprintf("projects/%d/+/data", cfg.Partition)
	}
	return "projects/+/data"
}

func (c *Client) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	c.log.Debug("Message received", "topic", msg.Topic(), "payload_size", len(msg.Payload()))

	var ders []DER
	if err := json.Unmarshal(msg.Payload(), &ders); err != nil {
		metrics.Local.Counter(metrics.MessagesDropped).With(prometheus.Labels{"topic": c.topic, "error": "parse_error"}).Inc()
		c.log.Error("Failed to parse message payload", "error", err, "topic", msg.Topic())
		return
	}

	if len(ders) == 0 {
		metrics.Local.Counter(metrics.MessagesDropped).With(prometheus.Labels{"topic": c.topic, "error": "empty_payload"}).Inc()
		c.log.Warn("Received empty DER array", "topic", msg.Topic())
		return
	}

	c.buffer.Add(msg.Payload())
	projectID := ders[0].ProjectID
	metrics.Local.Counter(metrics.MessagesReceived).With(prometheus.Labels{"topic": c.topic}).Inc()
	c.log.Debug("Message processed", "project_id", projectID, "der_count", len(ders), "payload_bytes", len(msg.Payload()))
}

func (c *Client) Stop() {
	c.log.Info("Stopping MQTT client", "topic", c.topic)
	if token := c.client.Unsubscribe(c.topic); token.Wait() && token.Error() != nil {
		c.log.Error("Failed to unsubscribe", "error", token.Error())
	}

	c.client.Disconnect(250)
	c.log.Info("MQTT client disconnected", "topic", c.topic)
}

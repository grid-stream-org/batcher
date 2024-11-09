package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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

// type DER struct {
// 	DerID             string    `json:"derId"`
// 	Type              string    `json:"type"`
// 	IsOnline          bool      `json:"isOnline"`
// 	Timestamp         time.Time `json:"Timestamp"`
// 	CurrentOutput     float64   `json:"currentOutput"`
// 	Units             string    `json:"units"`
// 	ProjectID         string    `json:"projectId"`
// 	UtilityID         string    `json:"utilityId"`
// 	IsStandalone      bool      `json:"isStandalone"`
// 	ConnectionStartAt string    `json:"connectionStartAt"`
// 	CurrentSoc        int       `json:"currentSoc"`
// }

type Client struct {
	client mqtt.Client
	buffer *buffer.Buffer
	topic  string
	cfg    *config.MQTTConfig
	log    *slog.Logger
}

func NewClient(ctx context.Context, cfg *config.MQTTConfig, buf *buffer.Buffer, log *slog.Logger) (*Client, error) {
	c := &Client{
		buffer: buf,
		topic:  getTopic(cfg),
		cfg:    cfg,
		log:    log.With("component", "mqtt"),
	}

	opts, err := createClientOptions(cfg, log)
	if err != nil {
		return nil, errors.Wrap(err, "creating client options")
	}

	c.client = mqtt.NewClient(opts)
	c.log.Info("attempting to connect to mqtt broker...")

	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		log.Error("error connecting to mqtt client", "error", err)
		return nil, errors.Wrap(token.Error(), "error connecting to broker")
	}
	return c, nil
}

func createClientOptions(cfg *config.MQTTConfig, log *slog.Logger) (*mqtt.ClientOptions, error) {
	clientID := fmt.Sprintf("batcher-%s", uuid.NewString())
	brokerURL := fmt.Sprintf("tls://%s:%d", cfg.Host, cfg.Port)

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
			log.Info("connected to mqtt broker", "client_id", clientID)
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			metrics.Local.Gauge(metrics.ConnectionStatus).WithLabelValues().Set(0)
			log.Error("lost connection to mqtt broker", "error", err)
		}).
		SetReconnectingHandler(func(c mqtt.Client, options *mqtt.ClientOptions) {
			log.Warn("attempting to reconnect to mqtt broker")
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
	log.Info("mqtt client created with options",
		"broker", opts.Servers[0].String(),
		"client_id", opts.ClientID,
		"clean_session", opts.CleanSession,
		"keep_alive", opts.KeepAlive,
		"username", opts.Username,
		"tls_enabled", opts.TLSConfig != nil,
		"auto_reconnect", opts.AutoReconnect,
		"protocol_version", opts.ProtocolVersion,
		"connect_timeout", opts.ConnectTimeout.String(),
		"keep_alive", opts.KeepAlive,
		"write_timeout", opts.WriteTimeout.String(),
	)

	return opts, nil
}

func (c *Client) Subscribe() error {
	c.log.Info("subscribing to topic", "topic", c.topic)
	token := c.client.Subscribe(c.topic, byte(c.cfg.QOS), c.handleMessage)
	if token.Wait() && token.Error() != nil {
		return errors.Wrapf(token.Error(), "subscribing to topic %s", c.topic)
	}

	c.log.Info("successfully subscribed to topic", "topic", c.topic, "qos Level", c.cfg.QOS)
	return nil
}

func getTopic(cfg *config.MQTTConfig) string {
	if cfg.PartitionMode {
		return fmt.Sprintf("projects/%d/+/data", cfg.Partition)
	}
	return "projects/+/data"
}

func (c *Client) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	pl := msg.Payload()

	if len(pl) == 0 {
		c.log.Warn("received empty DER array", "topic", msg.Topic())
		metrics.Local.Counter(metrics.MessagesDropped).With(prometheus.Labels{"topic": c.topic, "error": "empty_payload"}).Inc()
		return
	}

	c.buffer.Add(msg.Payload())
	c.log.Info("message received", "topic", msg.Topic(), "payload_size", len(pl), "der_count", len(pl))
	metrics.Local.Counter(metrics.MessagesReceived).With(prometheus.Labels{"topic": c.topic}).Inc()
}

func (c *Client) Stop() {
	c.log.Info("stopping mqtt client", "topic", c.topic)
	if token := c.client.Unsubscribe(c.topic); token.Wait() && token.Error() != nil {
		c.log.Error("failed to unsubscribe", "error", token.Error())
	}

	c.client.Disconnect(250)
	c.log.Info("mqtt client disconnected", "topic", c.topic)
}

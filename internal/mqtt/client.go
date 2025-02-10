package mqtt

import (
	"crypto/tls"
	"fmt"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/grid-stream-org/batcher/internal/config"
	"github.com/grid-stream-org/batcher/internal/task"
	"github.com/grid-stream-org/go-commons/pkg/eventbus"
	"github.com/pkg/errors"
)

type Client struct {
	client     mqtt.Client
	eventBus   eventbus.EventBus
	topic      string
	cfg        *config.MQTT
	subscribed bool
	log        *slog.Logger
}

func NewClient(cfg *config.MQTT, eb eventbus.EventBus, log *slog.Logger) (*Client, error) {
	c := &Client{
		eventBus:   eb,
		topic:      "$share/batcher/projects/#",
		cfg:        cfg,
		subscribed: false,
		log:        log.With("component", "mqtt_client"),
	}

	clientID := fmt.Sprintf("batcher-%s", uuid.NewString())
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("tls://%s:%d", cfg.Host, cfg.Port)).
		SetClientID(clientID).
		SetUsername(cfg.Username).
		SetPassword(cfg.Password).
		SetProtocolVersion(4).
		SetOnConnectHandler(func(c mqtt.Client) {
			log.Info("connected to mqtt broker", "client_id", clientID)
		}).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			log.Error("lost connection to mqtt broker", "error", err)
		}).
		SetReconnectingHandler(func(c mqtt.Client, options *mqtt.ClientOptions) {
			log.Warn("attempting to reconnect to mqtt broker")
		}).
		SetTLSConfig(&tls.Config{
			InsecureSkipVerify: true,
		})

	c.client = mqtt.NewClient(opts)
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
	return c, nil
}

func (c *Client) Connect() error {
	c.log.Debug("attempting to connect to mqtt broker")
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return errors.WithStack(token.Error())
	}
	return nil
}

func (c *Client) Subscribe() error {
	c.log.Debug("attempting to subscribe to topic", "topic", c.topic)
	token := c.client.Subscribe(c.topic, byte(c.cfg.QoS), c.handleMessage)
	if token.Wait() && token.Error() != nil {
		return errors.WithStack(token.Error())
	}
	c.subscribed = true
	c.log.Debug("successfully subscribed to topic", "topic", c.topic, "qos_level", c.cfg.QoS)
	return nil
}

func (c *Client) handleMessage(_ mqtt.Client, msg mqtt.Message) {
	pl := msg.Payload()
	c.eventBus.Publish(task.NewTask(pl))
	c.log.Debug("message received", "topic", msg.Topic(), "payload_size", len(pl))
}

func (c *Client) Stop() error {
	c.log.Debug("stopping mqtt client", "topic", c.topic)

	if c.subscribed {
		if token := c.client.Unsubscribe(c.topic); token.Wait() && token.Error() != nil {
			return errors.WithStack(token.Error())
		}
	}

	if c.client.IsConnected() {
		c.client.Disconnect(250)
	}

	c.log.Debug("mqtt client disconnected", "topic", c.topic)
	return nil
}

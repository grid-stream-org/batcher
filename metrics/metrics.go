package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const BasePath = "batcher_"

// Labels
const (
	TopicLabel = "topic"
	ErrorLabel = "error"
)

// Counters
const (
	MessagesReceived = BasePath + "messages_received_total"
	MessagesDropped  = BasePath + "messages_dropped_total"
	FlushCount       = BasePath + "flushes_total"
)

// Gauges
const (
	ConnectionStatus = BasePath + "connection_status"
	BufferSize       = BasePath + "buffer_messages"
	LastFlushTime    = BasePath + "last_flush_timestamp"
)

type Provider struct {
	counters map[string]*prometheus.CounterVec
	gauges   map[string]*prometheus.GaugeVec
}

var (
	Local Provider
	once  sync.Once
)

func InitMetricsProvider() {
	once.Do(func() {
		Local = Provider{
			counters: make(map[string]*prometheus.CounterVec),
			gauges:   make(map[string]*prometheus.GaugeVec),
		}

		// Counters
		Local.counters[MessagesReceived] = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: MessagesReceived,
				Help: "Total number of MQTT messages received",
			},
			[]string{TopicLabel},
		)

		Local.counters[MessagesDropped] = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: MessagesDropped,
				Help: "Total number of messages dropped",
			},
			[]string{TopicLabel, ErrorLabel},
		)

		Local.counters[FlushCount] = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: FlushCount,
				Help: "Total number of buffer flushes",
			},
			[]string{},
		)

		// Gauges
		Local.gauges[ConnectionStatus] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: ConnectionStatus,
				Help: "MQTT connection status (1=connected, 0=disconnected)",
			},
			[]string{},
		)

		Local.gauges[BufferSize] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: BufferSize,
				Help: "Current number of messages in buffer",
			},
			[]string{},
		)

		Local.gauges[LastFlushTime] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: LastFlushTime,
				Help: "Timestamp of last buffer flush",
			},
			[]string{},
		)
	})
}

func (p Provider) Counter(name string) *prometheus.CounterVec {
	return p.counters[name]
}

func (p Provider) Gauge(name string) *prometheus.GaugeVec {
	return p.gauges[name]
}

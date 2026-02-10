package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/infra"
)

// KafkaPublisher publishes events to Kafka using the infra layer connection manager.
type KafkaPublisher struct {
	conn   *infra.KafkaConnection
	topic  string
	source string
}

// KafkaConfig holds configuration for the Kafka event publisher.
type KafkaConfig struct {
	Brokers         []string
	Topic           string // Publisher-specific topic
	MaxMessageBytes int
	Timeout         time.Duration
	Compression     sarama.CompressionCodec

	// Reconnection settings (optional, defaults provided)
	ReconnectInterval    time.Duration
	MaxReconnectAttempts int
	ReconnectDelay       time.Duration
}

// NewKafkaPublisher creates a new Kafka event publisher with reconnection support.
func NewKafkaPublisher(config KafkaConfig, source string, logger gomon.Logger) (*KafkaPublisher, error) {
	infraCfg := infra.KafkaConfig{
		Brokers:              config.Brokers,
		MaxMessageBytes:      config.MaxMessageBytes,
		Timeout:              config.Timeout,
		Compression:          config.Compression,
		ReconnectInterval:    config.ReconnectInterval,
		MaxReconnectAttempts: config.MaxReconnectAttempts,
		ReconnectDelay:       config.ReconnectDelay,
	}

	conn, err := infra.NewKafkaConnection(context.Background(), infraCfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka connection: %w", err)
	}

	if config.Topic == "" {
		return nil, fmt.Errorf("topic must be specified in config")
	}

	return &KafkaPublisher{
		conn:   conn,
		topic:  config.Topic,
		source: source,
	}, nil
}

// NewKafkaPublisherWithConn creates a new Kafka event publisher using an existing connection.
// This is useful when the connection is managed externally (e.g., in main.go).
func NewKafkaPublisherWithConn(conn *infra.KafkaConnection, topic string, source string) *KafkaPublisher {
	return &KafkaPublisher{
		conn:   conn,
		topic:  topic,
		source: source,
	}
}

// Publish publishes an event to Kafka with automatic reconnection handling.
func (k *KafkaPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	// Convert to CloudEvents format
	ce := toCloudEventData(eventType, eventData, k.source)

	data, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: k.topic,
		Key:   sarama.StringEncoder(ce.ID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("ce_type"), Value: []byte(ce.Type)},
			{Key: []byte("ce_source"), Value: []byte(ce.Source)},
			{Key: []byte("ce_id"), Value: []byte(ce.ID)},
			{Key: []byte("ce_specversion"), Value: []byte(ce.SpecVersion)},
		},
	}

	err = k.conn.PublishWithContext(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka connection.
func (k *KafkaPublisher) Close() error {
	return k.conn.Close()
}

// IsConnected returns true if the Kafka connection is active.
func (k *KafkaPublisher) IsConnected() bool {
	return k.conn.IsConnected()
}

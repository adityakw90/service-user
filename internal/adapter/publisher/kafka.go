package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/adityakw90/service-user/internal/core/domain/event"
)

// KafkaPublisher publishes events to Kafka using Sarama SyncProducer.
type KafkaPublisher struct {
	producer sarama.SyncProducer
	config   KafkaConfig
	source   string
}

// KafkaConfig holds configuration for the Kafka event publisher.
type KafkaConfig struct {
	Brokers         []string
	Topic           string
	MaxMessageBytes int
	Timeout         time.Duration
	Compression     sarama.CompressionCodec
}

// NewKafkaPublisher creates a new Kafka event publisher.
func NewKafkaPublisher(config KafkaConfig, source string) (*KafkaPublisher, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Compression = config.Compression

	if config.MaxMessageBytes > 0 {
		saramaConfig.Producer.MaxMessageBytes = config.MaxMessageBytes
	}

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	return &KafkaPublisher{
		producer: producer,
		config:   config,
		source:   source,
	}, nil
}

// Publish publishes an event to Kafka.
func (k *KafkaPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	// Convert to CloudEvents format
	ce := toCloudEventData(eventType, eventData, k.source)

	data, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: k.config.Topic,
		Key:   sarama.StringEncoder(ce.ID),
		Value: sarama.ByteEncoder(data),
		Headers: []sarama.RecordHeader{
			{Key: []byte("ce_type"), Value: []byte(ce.Type)},
			{Key: []byte("ce_source"), Value: []byte(ce.Source)},
			{Key: []byte("ce_id"), Value: []byte(ce.ID)},
			{Key: []byte("ce_specversion"), Value: []byte(ce.SpecVersion)},
		},
	}

	_, _, err = k.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	return nil
}

// Close closes the Kafka producer connection.
func (k *KafkaPublisher) Close() error {
	return k.producer.Close()
}

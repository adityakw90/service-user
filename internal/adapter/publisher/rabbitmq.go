package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/infra"
)

// RabbitMQPublisher publishes events to RabbitMQ using the infra layer connection manager.
type RabbitMQPublisher struct {
	conn   *infra.RabbitMQConnection
	source string
}

// RabbitMQConfig holds configuration for the RabbitMQ event publisher.
type RabbitMQConfig struct {
	URL              string
	Exchange         string
	ExchangeType     string
	RoutingKeyPrefix string
	Durable          bool

	// Reconnection settings (optional, defaults provided)
	ReconnectInterval    time.Duration
	MaxReconnectAttempts int
	ReconnectDelay       time.Duration
}

// NewRabbitMQPublisher creates a new RabbitMQ event publisher with reconnection support.
func NewRabbitMQPublisher(config RabbitMQConfig, source string, logger gomon.Logger) (*RabbitMQPublisher, error) {
	infraCfg := infra.RabbitMQConfig{
		URL:                  config.URL,
		Exchange:             config.Exchange,
		ExchangeType:         config.ExchangeType,
		RoutingKeyPrefix:     config.RoutingKeyPrefix,
		Durable:              config.Durable,
		ReconnectInterval:    config.ReconnectInterval,
		MaxReconnectAttempts: config.MaxReconnectAttempts,
		ReconnectDelay:       config.ReconnectDelay,
	}

	conn, err := infra.NewRabbitMQConnection(context.Background(), infraCfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ connection: %w", err)
	}

	return &RabbitMQPublisher{
		conn:   conn,
		source: source,
	}, nil
}

// NewRabbitMQPublisherWithConn creates a new RabbitMQ event publisher using an existing connection.
// This is useful when the connection is managed externally (e.g., in main.go).
func NewRabbitMQPublisherWithConn(conn *infra.RabbitMQConnection, source string) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		conn:   conn,
		source: source,
	}
}

// Publish publishes an event to RabbitMQ with automatic reconnection handling.
func (r *RabbitMQPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	// Convert to CloudEvents format
	ce := toCloudEventData(eventType, eventData, r.source)

	body, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud event: %w", err)
	}

	routingKey := r.conn.GetRoutingKey(string(eventType))

	err = r.conn.PublishWithContext(
		ctx,
		routingKey,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers: amqp.Table{
				"ce_type":        ce.Type,
				"ce_source":      ce.Source,
				"ce_id":          ce.ID,
				"ce_specversion": ce.SpecVersion,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	return nil
}

// Close closes the RabbitMQ connection.
func (r *RabbitMQPublisher) Close() error {
	return r.conn.Close()
}

// IsConnected returns true if the RabbitMQ connection is active.
func (r *RabbitMQPublisher) IsConnected() bool {
	return r.conn.IsConnected()
}

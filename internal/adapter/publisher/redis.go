package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
)

// RedisPubSubPublisher publishes events to Redis Pub/Sub.
type RedisPubSubPublisher struct {
	client  *redis.Client
	channel string
	source  string
}

// NewRedisPubSubPublisher creates a new Redis Pub/Sub publisher.
func NewRedisPubSubPublisher(client *redis.Client, channel, source string) portEvent.EventPublisher {
	return &RedisPubSubPublisher{
		client:  client,
		channel: channel,
		source:  source,
	}
}

// Publish publishes an event to Redis Pub/Sub.
func (p *RedisPubSubPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	// Convert to CloudEvents JSON
	ce := toCloudEventData(eventType, eventData, p.source)

	data, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	// Publish to Redis Pub/Sub
	return p.client.Publish(ctx, p.channel, data).Err()
}

// Close closes the publisher connection.
func (p *RedisPubSubPublisher) Close() error {
	return nil // Redis client is managed externally
}

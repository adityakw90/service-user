package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/redis/go-redis/v9"
)

const (
	// RedisStreamBackend is the backend name for metrics.
	RedisStreamBackend = "redis_stream"
	// DefaultMaxLen is the approximate maximum length of the stream before trimming.
	DefaultMaxLen int64 = 100000
)

// RedisPublisher publishes events to Redis Streams.
// Uses XADD with MAXLEN for automatic trimming, providing at-least-once semantics.
type RedisPublisher struct {
	client  *redis.Client
	config  RedisPublisherConfig
	logger  gomon.Logger
	metrics *streamMetrics
}

type RedisPublisherConfig struct {
	Stream string
	Source string
	MaxLen int64
}

// streamMetrics holds atomic metrics for the Redis Stream publisher.
type streamMetrics struct {
	successTotal int64
	failureTotal int64
}

// NewRedisPublisher creates a new Redis Stream publisher.
// The client is managed externally and should be closed by the caller.
// Logger must not be nil.
func NewRedisPublisher(client *redis.Client, config RedisPublisherConfig, logger gomon.Logger) (portEvent.EventPublisher, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if config.Source == "" {
		config.Source = "service-user"
	}
	if config.Stream == "" {
		config.Stream = "service-user.events"
	}
	if config.MaxLen == 0 {
		config.MaxLen = DefaultMaxLen
	}

	return &RedisPublisher{
		client:  client,
		config:  config,
		logger:  logger,
		metrics: &streamMetrics{},
	}, nil
}

// Publish publishes an event to Redis Stream using XADD.
func (p *RedisPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	startTime := time.Now()

	// Convert to CloudEvents format
	ce := toCloudEventData(eventType, eventData, p.config.Source)

	data, err := json.Marshal(ce)
	if err != nil {
		p.logger.Error("marshal event", map[string]interface{}{
			"backend":    RedisStreamBackend,
			"event_type": eventType,
			"error":      err.Error(),
		})
		return fmt.Errorf("marshal event: %w", err)
	}

	// XADD with MAXLEN for automatic trimming
	args := &redis.XAddArgs{
		Stream: p.config.Stream,
		MaxLen: p.config.MaxLen,
		ID:     "*",
		Values: map[string]interface{}{
			"type":   ce.Type,
			"source": ce.Source,
			"id":     ce.ID,
			"time":   ce.Time,
			"data":   string(data),
		},
	}

	err = p.client.XAdd(ctx, args).Err()
	if err != nil {
		p.emitFailure(eventType, err)
		p.logger.Error("publish to redis stream failed", map[string]interface{}{
			"backend":    RedisStreamBackend,
			"stream":     p.config.Stream,
			"event_type": eventType,
			"error":      err.Error(),
		})
		return fmt.Errorf("xadd: %w", err)
	}

	p.emitSuccess(eventType, time.Since(startTime))
	p.logger.Debug("published to redis stream", map[string]interface{}{
		"backend":    RedisStreamBackend,
		"stream":     p.config.Stream,
		"event_type": eventType,
		"event_id":   ce.ID,
	})

	return nil
}

// Close closes the publisher connection.
// The Redis client is managed externally, so this is a no-op.
func (p *RedisPublisher) Close() error {
	return nil
}

// emitSuccess records a successful publish event with latency.
func (p *RedisPublisher) emitSuccess(eventType event.EventType, latency time.Duration) {
	// TODO: Integrate with proper metrics system
	p.metrics.successTotal++
	_ = latency
}

// emitFailure records a failed publish event.
func (p *RedisPublisher) emitFailure(eventType event.EventType, err error) {
	// TODO: Integrate with proper metrics system
	p.metrics.failureTotal++
	_ = eventType
	_ = err
}

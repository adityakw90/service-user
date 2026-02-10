package infra

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/trace"
)

// RabbitMQConfig holds configuration for RabbitMQ connection.
type RabbitMQConfig struct {
	URL              string
	Exchange         string
	ExchangeType     string
	RoutingKeyPrefix string
	Durable          bool

	// Reconnection settings
	ReconnectInterval    time.Duration
	MaxReconnectAttempts int
	ReconnectDelay       time.Duration
}

// RabbitMQConnection manages a RabbitMQ connection with automatic reconnection.
// Each publish operation creates a new channel, providing better isolation
// for concurrent operations.
type RabbitMQConnection struct {
	config        RabbitMQConfig
	conn          *amqp.Connection
	connMu        sync.RWMutex // Protects connection only
	closed        atomic.Bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	reconnectChan chan struct{}
	logger        gomon.Logger
}

// NoopLogger is a no-op logger implementation.
type NoopLogger struct{}

func (l *NoopLogger) SetLogLevel(level string)                            {}
func (l *NoopLogger) Debug(message string, fields map[string]interface{}) {}
func (l *NoopLogger) Info(message string, fields map[string]interface{})  {}
func (l *NoopLogger) Warn(message string, fields map[string]interface{})  {}
func (l *NoopLogger) Error(message string, fields map[string]interface{}) {}
func (l *NoopLogger) Fatal(message string, fields map[string]interface{}) {}
func (l *NoopLogger) WithSpanContext(span trace.SpanContext) gomon.Logger { return l }
func (l *NoopLogger) Sync() error                                         { return nil }

// NewRabbitMQConnection creates a new RabbitMQ connection with reconnection support.
func NewRabbitMQConnection(ctx context.Context, cfg RabbitMQConfig, logger gomon.Logger) (*RabbitMQConnection, error) {
	if logger == nil {
		logger = &NoopLogger{}
	}

	// Set defaults
	if cfg.ReconnectInterval == 0 {
		cfg.ReconnectInterval = 5 * time.Second
	}
	if cfg.MaxReconnectAttempts == 0 {
		cfg.MaxReconnectAttempts = 0 // 0 means infinite retries
	}
	if cfg.ReconnectDelay == 0 {
		cfg.ReconnectDelay = 1 * time.Second
	}

	ctx, cancel := context.WithCancel(ctx)

	r := &RabbitMQConnection{
		config:        cfg,
		ctx:           ctx,
		cancel:        cancel,
		reconnectChan: make(chan struct{}, 1),
		logger:        logger,
	}

	// Initial connection
	if err := r.connect(ctx); err != nil {
		cancel()
		return nil, err
	}

	// Start connection monitor
	r.wg.Add(1)
	go r.monitorConnection()

	return r, nil
}

// connect establishes a new connection and declares the exchange.
// The exchange is declared once during connection using a temporary channel.
func (r *RabbitMQConnection) connect(ctx context.Context) error {
	r.connMu.Lock()
	defer r.connMu.Unlock()

	var lastErr error

	// Attempt connection with retry
	for attempt := 0; r.config.MaxReconnectAttempts == 0 || attempt < r.config.MaxReconnectAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.config.ReconnectDelay * time.Duration(min(attempt, 5))):
				// Exponential backoff (capped at 5x)
			}
		}

		// Dial connection
		conn, err := amqp.DialConfig(r.config.URL, amqp.Config{
			Heartbeat: 10 * time.Second,
			Locale:    "en_US",
		})
		if err != nil {
			lastErr = fmt.Errorf("failed to dial RabbitMQ (attempt %d): %w", attempt+1, err)
			r.logger.Error("rabbitmq connection failed", map[string]interface{}{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
			continue
		}

		// Open temporary channel to declare exchange
		tempCh, err := conn.Channel()
		if err != nil {
			conn.Close()
			lastErr = fmt.Errorf("failed to open channel (attempt %d): %w", attempt+1, err)
			r.logger.Error("rabbitmq channel open failed", map[string]interface{}{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
			continue
		}

		// Declare exchange
		err = tempCh.ExchangeDeclare(
			r.config.Exchange,
			r.config.ExchangeType,
			r.config.Durable,
			false, // auto-delete
			false, // internal
			false, // no-wait
			nil,   // arguments
		)
		// Close temp channel immediately after exchange declaration
		tempCh.Close()

		if err != nil {
			conn.Close()
			lastErr = fmt.Errorf("failed to declare exchange (attempt %d): %w", attempt+1, err)
			r.logger.Error("rabbitmq exchange declare failed", map[string]interface{}{
				"attempt":  attempt + 1,
				"exchange": r.config.Exchange,
				"error":    err.Error(),
			})
			continue
		}

		// Connection successful - update state
		if r.conn != nil {
			r.conn.Close()
		}
		r.conn = conn

		r.logger.Info("rabbitmq connected successfully", map[string]interface{}{
			"url":      r.config.URL,
			"exchange": r.config.Exchange,
		})

		// Setup close listener
		go r.waitForClose(conn)

		return nil
	}

	return lastErr
}

// waitForClose monitors the connection for close events.
func (r *RabbitMQConnection) waitForClose(conn *amqp.Connection) {
	err := <-conn.NotifyClose(make(chan *amqp.Error, 1))
	if err != nil {
		r.logger.Warn("rabbitmq connection closed", map[string]interface{}{
			"error": err.Error(),
		})
		select {
		case r.reconnectChan <- struct{}{}:
		default:
		}
	}
}

// monitorConnection handles reconnection logic.
func (r *RabbitMQConnection) monitorConnection() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.config.ReconnectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.reconnectChan:
			r.logger.Info("rabbitmq reconnection triggered", nil)
			if err := r.connect(r.ctx); err != nil {
				r.logger.Error("rabbitmq reconnection failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		case <-ticker.C:
			// Periodic health check
			if !r.IsConnected() {
				r.logger.Warn("rabbitmq connection check failed, attempting reconnect", nil)
				if err := r.connect(r.ctx); err != nil {
					r.logger.Error("rabbitmq reconnection failed", map[string]interface{}{
						"error": err.Error(),
					})
				}
			}
		}
	}
}

// PublishWithContext publishes a message with context support.
// Each publish creates a new channel, publishes, then closes it.
// This provides better isolation for concurrent publish operations.
func (r *RabbitMQConnection) PublishWithContext(ctx context.Context, routingKey string, publishing amqp.Publishing) error {
	if r.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	// Get connection with read lock
	r.connMu.RLock()
	conn := r.conn
	r.connMu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return fmt.Errorf("no active connection")
	}

	// Create a new channel for this publish
	ch, err := conn.Channel()
	if err != nil {
		r.logger.Error("failed to open channel for publish", map[string]interface{}{
			"error":       err.Error(),
			"routing_key": routingKey,
		})
		r.triggerReconnect()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Ensure channel is closed when done
	defer func() {
		if closeErr := ch.Close(); closeErr != nil {
			r.logger.Warn("error closing channel after publish", map[string]interface{}{
				"error": closeErr.Error(),
			})
		}
	}()

	// Publish with the channel
	err = ch.PublishWithContext(
		ctx,
		r.config.Exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		publishing,
	)

	if err != nil {
		r.logger.Error("rabbitmq publish failed", map[string]interface{}{
			"error":       err.Error(),
			"routing_key": routingKey,
		})
		r.triggerReconnect()
		return fmt.Errorf("publish failed: %w", err)
	}

	return nil
}

// triggerReconnect safely triggers reconnection.
func (r *RabbitMQConnection) triggerReconnect() {
	select {
	case r.reconnectChan <- struct{}{}:
	default:
	}
}

// IsConnected returns true if the connection is active.
func (r *RabbitMQConnection) IsConnected() bool {
	if r.closed.Load() {
		return false
	}

	r.connMu.RLock()
	defer r.connMu.RUnlock()

	if r.conn == nil {
		return false
	}

	// Check if connection is still open
	return !r.conn.IsClosed()
}

// Close closes the RabbitMQ connection and stops reconnection attempts.
func (r *RabbitMQConnection) Close() error {
	if !r.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	r.cancel()
	r.wg.Wait()

	r.connMu.Lock()
	defer r.connMu.Unlock()

	var errs []error

	if r.conn != nil && !r.conn.IsClosed() {
		if err := r.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("connection close: %w", err))
		}
		r.conn = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	r.logger.Info("rabbitmq connection closed", nil)
	return nil
}

// GetRoutingKey returns the full routing key for an event type.
func (r *RabbitMQConnection) GetRoutingKey(eventType string) string {
	return r.config.RoutingKeyPrefix + eventType
}

// Config returns the RabbitMQ configuration.
func (r *RabbitMQConnection) Config() RabbitMQConfig {
	return r.config
}

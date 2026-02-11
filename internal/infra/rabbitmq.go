package infra

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQConfig holds configuration for RabbitMQ connection.
type RabbitMQConfig struct {
	Host                 string
	Port                 int
	User                 string
	Password             string
	Vhost                string
	ReconnectInterval    time.Duration
	ReconnectMaxAttempts int
}

// RabbitMQConnection manages a RabbitMQ connection with automatic reconnection.
// Each publish operation creates a new channel, providing better isolation
// for concurrent operations.
// Exchanges and routing keys are managed by the adapter layer.
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

// NewRabbitMQConnection creates a new RabbitMQ connection with reconnection support.
func NewRabbitMQConnection(ctx context.Context, cfg RabbitMQConfig, logger gomon.Logger) (*RabbitMQConnection, error) {
	if logger == nil {
		logger = &NoopLogger{}
	}

	// Set defaults
	if cfg.ReconnectInterval == 0 {
		cfg.ReconnectInterval = 1 * time.Second
	}
	if cfg.ReconnectMaxAttempts == 0 {
		cfg.ReconnectMaxAttempts = 0 // 0 means infinite retries
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

// connect establishes a new connection.
func (r *RabbitMQConnection) connect(ctx context.Context) error {
	r.connMu.Lock()
	defer r.connMu.Unlock()

	var lastErr error

	// Attempt connection with retry
	for attempt := 0; r.config.ReconnectMaxAttempts == 0 || attempt < r.config.ReconnectMaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.config.ReconnectInterval * time.Duration(min(attempt, 5))):
				// Exponential backoff (capped at 5x)
			}
		}

		// Dial connection
		rabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%d/%s", r.config.User, r.config.Password, r.config.Host, r.config.Port, r.config.Vhost)
		conn, err := amqp.DialConfig(rabbitUrl, amqp.Config{
			Heartbeat: 10 * time.Second,
			Locale:    "en_US",
		})
		if err != nil {
			lastErr = fmt.Errorf("failed to dial RabbitMQ (attempt %d): %w", attempt+1, err)
			r.logger.Error("rabbitmq connection failed", map[string]any{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
			continue
		}

		// Connection successful - update state
		if r.conn != nil {
			r.conn.Close()
		}
		r.conn = conn

		r.logger.Info("rabbitmq connected successfully", map[string]any{
			"host":  r.config.Host,
			"port":  r.config.Port,
			"vhost": r.config.Vhost,
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
		r.logger.Warn("rabbitmq connection closed", map[string]any{
			"error": err.Error(),
		})
		select {
		case r.reconnectChan <- struct{}{}:
		default:
		}
	}
}

// monitorConnection handles reconnection logic.
// Relies on waitForClose() to detect connection failures via NotifyClose.
// Publish operations also trigger reconnection on failure.
func (r *RabbitMQConnection) monitorConnection() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.reconnectChan:
			r.logger.Info("rabbitmq reconnection triggered", nil)
			if err := r.connect(r.ctx); err != nil {
				r.logger.Error("rabbitmq reconnection failed", map[string]any{
					"error": err.Error(),
				})
			}
		}
	}
}

// PublishWithContext publishes a message with context support.
// The exchange, routing key, and publishing properties are provided by the caller.
// This is a fire-and-forget method - for at-least-once delivery, use PublishWithConfirm.
func (r *RabbitMQConnection) PublishWithContext(ctx context.Context, exchange, routingKey string, publishing amqp.Publishing) error {
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
		r.logger.Error("failed to open channel for publish", map[string]any{
			"error":       err.Error(),
			"exchange":    exchange,
			"routing_key": routingKey,
		})
		r.triggerReconnect()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Ensure channel is closed when done
	defer func() {
		if closeErr := ch.Close(); closeErr != nil {
			r.logger.Warn("error closing channel after publish", map[string]any{
				"error": closeErr.Error(),
			})
		}
	}()

	// Publish with the channel
	err = ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false, // mandatory
		false, // immediate
		publishing,
	)

	if err != nil {
		r.logger.Error("rabbitmq publish failed", map[string]any{
			"error":       err.Error(),
			"exchange":    exchange,
			"routing_key": routingKey,
		})
		r.triggerReconnect()
		return fmt.Errorf("publish failed: %w", err)
	}

	return nil
}

// PublishWithConfirm publishes a message with publisher confirms enabled.
// This implements "at least once" delivery semantics by:
// 1. Enabling confirm mode on the channel
// 2. Waiting for broker ACK before returning
// 3. Retrying on NACK or timeout with exponential backoff
//
// Parameters:
// - ctx: Context for timeout/cancellation
// - exchange: Target exchange name
// - routingKey: Routing key for message routing
// - publishing: Message to publish
// - maxRetries: Maximum number of retry attempts (use 0 for single attempt)
// - retryInterval: Initial retry interval (will apply exponential backoff)
// - confirmTimeout: Time to wait for broker confirmation
func (r *RabbitMQConnection) PublishWithConfirm(
	ctx context.Context,
	exchange, routingKey string,
	publishing amqp.Publishing,
	maxRetries int,
	retryInterval time.Duration,
	confirmTimeout time.Duration,
) error {
	if r.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	// Set defaults
	if maxRetries <= 0 {
		maxRetries = 5
	}
	if retryInterval <= 0 {
		retryInterval = 1 * time.Second
	}
	if confirmTimeout <= 0 {
		confirmTimeout = 30 * time.Second
	}

	var lastErr error

	// Retry loop with exponential backoff
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff (capped at 5x)
			backoffDuration := retryInterval * time.Duration(min(attempt, 5))

			r.logger.Info("rabbitmq publish retry", map[string]any{
				"attempt":     attempt + 1,
				"max_retries": maxRetries,
				"backoff":     backoffDuration.String(),
				"last_error":  lastErr.Error(),
			})

			select {
			case <-ctx.Done():
				return fmt.Errorf("publish canceled during backoff: %w", ctx.Err())
			case <-time.After(backoffDuration):
			}
		}

		// Get connection with read lock
		r.connMu.RLock()
		conn := r.conn
		r.connMu.RUnlock()

		if conn == nil || conn.IsClosed() {
			lastErr = fmt.Errorf("no active connection (attempt %d)", attempt+1)
			r.triggerReconnect()
			continue
		}

		// Create a new channel for this publish
		ch, err := conn.Channel()
		if err != nil {
			lastErr = fmt.Errorf("failed to open channel (attempt %d): %w", attempt+1, err)
			r.logger.Error("failed to open channel for publish with confirm", map[string]any{
				"attempt":      attempt + 1,
				"error":        err.Error(),
				"exchange":     exchange,
				"routing_key":  routingKey,
			})
			r.triggerReconnect()
			continue
		}

		// Enable confirm mode
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			lastErr = fmt.Errorf("failed to enable confirm mode (attempt %d): %w", attempt+1, err)
			r.logger.Error("failed to enable confirm mode", map[string]any{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
			continue
		}

		// Set up confirmation channel
		confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

		// Publish with the channel
		publishErr := ch.PublishWithContext(
			ctx,
			exchange,
			routingKey,
			false, // mandatory
			false, // immediate
			publishing,
		)

		if publishErr != nil {
			ch.Close()
			lastErr = fmt.Errorf("publish failed (attempt %d): %w", attempt+1, publishErr)
			r.logger.Error("rabbitmq publish with confirm failed", map[string]any{
				"attempt":      attempt + 1,
				"error":        publishErr.Error(),
				"exchange":     exchange,
				"routing_key":  routingKey,
			})
			r.triggerReconnect()
			continue
		}

		// Wait for confirmation with timeout
		select {
		case confirm := <-confirmCh:
			ch.Close()
			if confirm.Ack {
				// Success - broker acknowledged the message
				if attempt > 0 {
					r.logger.Info("rabbitmq publish succeeded after retry", map[string]any{
						"attempt":      attempt + 1,
						"exchange":     exchange,
						"routing_key":  routingKey,
						"delivery_tag": confirm.DeliveryTag,
					})
				}
				return nil
			}
			// NACK - broker rejected the message
			lastErr = fmt.Errorf("broker NACKed message (attempt %d, tag: %d)", attempt+1, confirm.DeliveryTag)
			r.logger.Warn("rabbitmq broker returned NACK", map[string]any{
				"attempt":      attempt + 1,
				"delivery_tag": confirm.DeliveryTag,
				"exchange":     exchange,
				"routing_key":  routingKey,
			})

		case <-time.After(confirmTimeout):
			ch.Close()
			lastErr = fmt.Errorf("confirmation timeout after %v (attempt %d)", confirmTimeout, attempt+1)
			r.logger.Warn("rabbitmq publish confirmation timeout", map[string]any{
				"attempt":      attempt + 1,
				"timeout":      confirmTimeout.String(),
				"exchange":     exchange,
				"routing_key":  routingKey,
			})

		case <-ctx.Done():
			ch.Close()
			return fmt.Errorf("publish canceled while waiting for confirmation: %w", ctx.Err())
		}
	}

	return fmt.Errorf("publish failed after %d attempts: %w", maxRetries, lastErr)
}

// DeclareExchange declares an exchange on the RabbitMQ server.
// This is typically called by the adapter layer during initialization.
func (r *RabbitMQConnection) DeclareExchange(exchange, exchangeType string, durable bool) error {
	if r.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	r.connMu.RLock()
	conn := r.conn
	r.connMu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return fmt.Errorf("no active connection")
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		exchange,
		exchangeType,
		durable,
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	r.logger.Info("rabbitmq exchange declared", map[string]any{
		"exchange": exchange,
		"type":     exchangeType,
		"durable":  durable,
	})

	return nil
}

// QueueConfig holds configuration for declaring a queue.
type QueueConfig struct {
	Name          string
	Durable       bool
	AutoDelete    bool
	Exclusive     bool
	Args          amqp.Table
}

// DeclareQueue declares a queue on the RabbitMQ server.
// This ensures messages are stored even when no consumers are running.
func (r *RabbitMQConnection) DeclareQueue(cfg QueueConfig) error {
	if r.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	r.connMu.RLock()
	conn := r.conn
	r.connMu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return fmt.Errorf("no active connection")
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		cfg.Name,
		cfg.Durable,
		cfg.AutoDelete,
		cfg.Exclusive,
		false, // no-wait
		cfg.Args,
	)

	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	r.logger.Info("rabbitmq queue declared", map[string]any{
		"queue":      cfg.Name,
		"durable":    cfg.Durable,
		"auto_delete": cfg.AutoDelete,
		"exclusive":  cfg.Exclusive,
	})

	return nil
}

// BindQueue binds a queue to an exchange with a routing key pattern.
// This enables messages sent to the exchange with matching routing keys to be
// delivered to the queue.
func (r *RabbitMQConnection) BindQueue(queue, exchange, routingKey string, args amqp.Table) error {
	if r.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	r.connMu.RLock()
	conn := r.conn
	r.connMu.RUnlock()

	if conn == nil || conn.IsClosed() {
		return fmt.Errorf("no active connection")
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	err = ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false, // no-wait
		args,
	)

	if err != nil {
		return fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	r.logger.Info("rabbitmq queue bound to exchange", map[string]any{
		"queue":       queue,
		"exchange":    exchange,
		"routing_key": routingKey,
	})

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

// Config returns the RabbitMQ configuration.
func (r *RabbitMQConnection) Config() RabbitMQConfig {
	return r.config
}

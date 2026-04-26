package infra

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	config        RabbitConfig
	conn          *amqp.Connection
	connMu        sync.RWMutex // Protects connection only
	closed        atomic.Bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	reconnectChan chan struct{}
	logger        gomon.Logger
}

type RabbitConfig struct {
	Host                 string
	Port                 int
	User                 string
	Password             string
	Vhost                string
	ReconnectInterval    time.Duration
	ReconnectMaxAttempts int
}

func NewRabbitConnection(ctx context.Context, cfg RabbitConfig, logger gomon.Logger) (*Rabbit, error) {
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

	r := &Rabbit{
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
func (r *Rabbit) connect(ctx context.Context) error {
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
func (r *Rabbit) waitForClose(conn *amqp.Connection) {
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
func (r *Rabbit) monitorConnection() {
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

// triggerReconnect safely triggers reconnection.
func (r *Rabbit) triggerReconnect() {
	select {
	case r.reconnectChan <- struct{}{}:
	default:
	}
}

// get connection
func (r *Rabbit) getConnection() (*amqp.Connection, error) {
	// get connection with read lock
	r.connMu.RLock()
	conn := r.conn
	r.connMu.RUnlock()

	// check connection is not nil or closed
	if conn == nil || conn.IsClosed() {
		return nil, fmt.Errorf("no active connection")
	}

	return conn, nil
}

// Get channel connection from rabbitmq connection
func (r *Rabbit) getChannel() (*amqp.Channel, error) {
	conn, err := r.getConnection()
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		// triger reconnect, because connection might be unhealthy
		r.triggerReconnect()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	return ch, nil
}

// Build amqp publishing message
func (r *Rabbit) buildAmqpPublishing(
	contentType string,
	headers map[string]string,
	body []byte,
	deliveryMode *uint8,
	priority *uint8,
	timestamp *time.Time,
	expiration *time.Duration,
) amqp.Publishing {
	amqpTimestamp := time.Now()
	if timestamp != nil {
		amqpTimestamp = *timestamp
	}

	publishing := amqp.Publishing{
		ContentType: contentType,
		Body:        body,
		Timestamp:   amqpTimestamp,
	}

	if headers != nil {
		amqpHeaders := make(amqp.Table, len(headers))
		for k, v := range headers {
			amqpHeaders[k] = v
		}
		publishing.Headers = amqpHeaders
	}

	if deliveryMode != nil {
		switch *deliveryMode {
		case amqp.Transient:
			publishing.DeliveryMode = amqp.Transient
		case amqp.Persistent:
			publishing.DeliveryMode = amqp.Persistent
		}
	}

	if priority != nil {
		publishing.Priority = *priority
	}

	if expiration != nil {
		publishing.Expiration = strconv.FormatInt(expiration.Milliseconds(), 10)
	}

	return publishing
}

// DeclareExchange declares an exchange on the RabbitMQ server.
// This is typically called by the adapter layer during initialization.
func (r *Rabbit) DeclareExchange(exchange string, exchangeType string, durable bool) error {
	if r.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	ch, err := r.getChannel()
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

// publish message
func (r *Rabbit) Publish(
	ctx context.Context,
	exchange string,
	routingKey string,
	contentType string,
	headers map[string]string,
	body []byte,
	deliveryMode *uint8,
	priority *uint8,
	timestamp *time.Time,
	expiration *time.Duration,
) error {
	if r.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	// Create a new channel for this publish
	ch, err := r.getChannel()
	if err != nil {
		r.logger.Error("failed to publish", map[string]any{
			"error":       err.Error(),
			"exchange":    exchange,
			"routing_key": routingKey,
		})
		return fmt.Errorf("failed to publish: %w", err)
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
		r.buildAmqpPublishing(
			contentType,
			headers,
			body,
			deliveryMode,
			priority,
			timestamp,
			expiration,
		),
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

// publish message with confirm
func (r *Rabbit) PublishWithConfirm(
	ctx context.Context,
	exchange string,
	routingKey string,
	contentType string,
	headers map[string]string,
	body []byte,
	deliveryMode *uint8,
	priority *uint8,
	timestamp *time.Time,
	expiration *time.Duration,
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

	// Get the AmqpPublishing once
	publishMsg := r.buildAmqpPublishing(
		contentType,
		headers,
		body,
		deliveryMode,
		priority,
		timestamp,
		expiration,
	)

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

		// Try to get channel
		ch, err := r.getChannel()
		if err != nil {
			lastErr = fmt.Errorf("failed to publish (attempt %d): %w", attempt+1, err)
			r.logger.Error("failed to publish", map[string]any{
				"attempt":     attempt + 1,
				"error":       err.Error(),
				"exchange":    exchange,
				"routing_key": routingKey,
			})
			continue // Retry
		}

		// Enable publisher confirms
		if err := ch.Confirm(false); err != nil {
			ch.Close()
			lastErr = fmt.Errorf("failed to enable confirm mode (attempt %d): %w", attempt+1, err)
			r.logger.Error("failed to enable confirm mode", map[string]any{
				"attempt": attempt + 1,
				"error":   err.Error(),
			})
			continue
		}

		// Set up confirmation listener
		confirmCh := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

		// Publish message
		publishErr := ch.PublishWithContext(
			ctx,
			exchange,
			routingKey,
			false, // mandatory
			false, // immediate
			publishMsg,
		)

		if publishErr != nil {
			ch.Close()
			lastErr = fmt.Errorf("publish failed (attempt %d): %w", attempt+1, publishErr)
			r.logger.Error("rabbitmq publish with confirm failed", map[string]any{
				"attempt":     attempt + 1,
				"error":       publishErr.Error(),
				"exchange":    exchange,
				"routing_key": routingKey,
			})
			r.triggerReconnect()
			continue // Retry
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
				"attempt":     attempt + 1,
				"timeout":     confirmTimeout.String(),
				"exchange":    exchange,
				"routing_key": routingKey,
			})

		case <-ctx.Done():
			ch.Close()
			return fmt.Errorf("publish canceled while waiting for confirmation: %w", ctx.Err())
		}
	}

	// All retries failed
	return fmt.Errorf("publish failed after %d attempts: %v", maxRetries, lastErr)
}

// Close closes the RabbitMQ connection and stops reconnection attempts.
func (r *Rabbit) Close() error {
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

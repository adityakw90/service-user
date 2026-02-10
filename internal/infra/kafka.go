package infra

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	gomon "github.com/adityakw90/go-monitoring"
)

// KafkaConfig holds configuration for Kafka connection.
type KafkaConfig struct {
	Brokers         []string
	MaxMessageBytes int
	Timeout         time.Duration
	Compression     sarama.CompressionCodec

	// Reconnection settings
	ReconnectMaxAttempts int
	ReconnectInterval    time.Duration
}

// KafkaConnection manages a Kafka producer connection with automatic reconnection.
// Each publish operation uses the shared producer for efficiency.
// Topics are specified per-message, allowing a single connection to publish to multiple topics.
type KafkaConnection struct {
	config        KafkaConfig
	producer      sarama.SyncProducer
	producerMu    sync.RWMutex // Protects producer only
	closed        atomic.Bool
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	reconnectChan chan struct{}
	logger        gomon.Logger
}

// NewKafkaConnection creates a new Kafka connection with reconnection support.
func NewKafkaConnection(ctx context.Context, cfg KafkaConfig, logger gomon.Logger) (*KafkaConnection, error) {
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
	if cfg.Compression == sarama.CompressionNone {
		cfg.Compression = sarama.CompressionSnappy
	}

	ctx, cancel := context.WithCancel(ctx)

	k := &KafkaConnection{
		config:        cfg,
		ctx:           ctx,
		cancel:        cancel,
		reconnectChan: make(chan struct{}, 1),
		logger:        logger,
	}

	// Initial connection
	if err := k.connect(ctx); err != nil {
		cancel()
		return nil, err
	}

	// Start connection monitor
	k.wg.Add(1)
	go k.monitorConnection()

	return k, nil
}

// connect establishes a new Kafka producer connection.
func (k *KafkaConnection) connect(ctx context.Context) error {
	k.producerMu.Lock()
	defer k.producerMu.Unlock()

	var lastErr error

	// Attempt connection with retry
	for attempt := 0; k.config.ReconnectMaxAttempts == 0 || attempt < k.config.ReconnectMaxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(k.config.ReconnectInterval * time.Duration(min(attempt, 5))):
				// Exponential backoff (capped at 5x)
			}
		}

		// Create Sarama config
		saramaConfig := sarama.NewConfig()
		saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
		saramaConfig.Producer.Retry.Max = 5
		saramaConfig.Producer.Return.Successes = true
		saramaConfig.Producer.Compression = k.config.Compression

		if k.config.MaxMessageBytes > 0 {
			saramaConfig.Producer.MaxMessageBytes = k.config.MaxMessageBytes
		}

		// Set client ID for tracking
		saramaConfig.ClientID = fmt.Sprintf("service-user-%d", attempt)

		// Create producer
		producer, err := sarama.NewSyncProducer(k.config.Brokers, saramaConfig)
		if err != nil {
			lastErr = fmt.Errorf("failed to create Kafka producer (attempt %d): %w", attempt+1, err)
			k.logger.Error("kafka producer creation failed", map[string]any{
				"attempt": attempt + 1,
				"brokers": k.config.Brokers,
				"error":   err.Error(),
			})
			continue
		}

		// Connection successful - update state
		if k.producer != nil {
			k.producer.Close()
		}
		k.producer = producer

		k.logger.Info("kafka connected successfully", map[string]any{
			"brokers": k.config.Brokers,
		})

		return nil
	}

	return lastErr
}

// monitorConnection handles reconnection logic.
// Sarama's internal health checking handles connection state detection.
// Publish operations trigger reconnection on failure.
func (k *KafkaConnection) monitorConnection() {
	defer k.wg.Done()

	for {
		select {
		case <-k.ctx.Done():
			return
		case <-k.reconnectChan:
			k.logger.Info("kafka reconnection triggered", nil)
			if err := k.connect(k.ctx); err != nil {
				k.logger.Error("kafka reconnection failed", map[string]any{
					"error": err.Error(),
				})
			}
		}
	}
}

// PublishWithContext publishes a message with context support.
// The message topic, key, value, and headers are provided by the caller.
func (k *KafkaConnection) PublishWithContext(ctx context.Context, msg *sarama.ProducerMessage) error {
	if k.closed.Load() {
		return fmt.Errorf("connection is closed")
	}

	// Get producer with read lock
	k.producerMu.RLock()
	producer := k.producer
	k.producerMu.RUnlock()

	if producer == nil {
		return fmt.Errorf("no active producer")
	}

	// Topic must be set by the caller
	if msg.Topic == "" {
		return fmt.Errorf("message topic must be set")
	}

	// Publish with timeout
	if k.config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, k.config.Timeout)
		defer cancel()
	}

	// Sarama's SendMessage is blocking, we need to run it in a goroutine
	// to respect context cancellation
	resultChan := make(chan error, 1)
	go func() {
		_, _, err := producer.SendMessage(msg)
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			k.logger.Error("kafka publish failed", map[string]any{
				"error": err.Error(),
				"topic": msg.Topic,
			})
			k.triggerReconnect()
			return fmt.Errorf("publish failed: %w", err)
		}
		return nil
	case <-ctx.Done():
		k.triggerReconnect()
		return fmt.Errorf("publish timeout: %w", ctx.Err())
	}
}

// triggerReconnect safely triggers reconnection.
func (k *KafkaConnection) triggerReconnect() {
	select {
	case k.reconnectChan <- struct{}{}:
	default:
	}
}

// IsConnected returns true if the producer is active.
func (k *KafkaConnection) IsConnected() bool {
	if k.closed.Load() {
		return false
	}

	k.producerMu.RLock()
	defer k.producerMu.RUnlock()

	return k.producer != nil
}

// Close closes the Kafka producer connection and stops reconnection attempts.
func (k *KafkaConnection) Close() error {
	if !k.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	k.cancel()
	k.wg.Wait()

	k.producerMu.Lock()
	defer k.producerMu.Unlock()

	var errs []error

	if k.producer != nil {
		if err := k.producer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("producer close: %w", err))
		}
		k.producer = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("close errors: %v", errs)
	}

	k.logger.Info("kafka connection closed", nil)
	return nil
}

// GetProducer returns the underlying Sarama SyncProducer.
// Use this for advanced operations not covered by PublishWithContext.
func (k *KafkaConnection) GetProducer() sarama.SyncProducer {
	k.producerMu.RLock()
	defer k.producerMu.RUnlock()
	return k.producer
}

// Config returns the Kafka configuration.
func (k *KafkaConnection) Config() KafkaConfig {
	return k.config
}

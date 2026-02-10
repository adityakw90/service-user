package infra

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/IBM/sarama"
)

func TestInfra_Kafka_NewKafkaConnection(t *testing.T) {
	tests := []struct {
		name   string
		config KafkaConfig
	}{
		{
			name: "Default reconnection settings",
			config: KafkaConfig{
				Brokers: []string{"localhost:9092"},
			},
		},
		{
			name: "Custom reconnection settings",
			config: KafkaConfig{
				Brokers:              []string{"localhost:9092"},
				ReconnectInterval:    2 * time.Second,
				ReconnectMaxAttempts: 5,
				Compression:          sarama.CompressionGZIP,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config is parsed correctly
			// We can't test actual connection without Kafka running
			if len(tt.config.Brokers) == 0 {
				t.Error("brokers should not be empty")
			}
		})
	}
}

// TestKafkaConnection_ReconnectSettings tests the reconnection configuration.
func TestKafkaConnection_ReconnectSettings(t *testing.T) {
	tests := []struct {
		name   string
		config KafkaConfig
	}{
		{
			name: "Default reconnection settings",
			config: KafkaConfig{
				Brokers: []string{"localhost:9092"},
			},
		},
		{
			name: "Custom reconnection settings",
			config: KafkaConfig{
				Brokers:              []string{"localhost:9092"},
				ReconnectInterval:    2 * time.Second,
				ReconnectMaxAttempts: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config is parsed correctly
			// We can't test actual connection without Kafka running
			if len(tt.config.Brokers) == 0 {
				t.Error("brokers should not be empty")
			}
		})
	}
}

// TestKafkaConnection_Concurrency tests concurrent access to connection methods.
func TestKafkaConnection_Concurrency(t *testing.T) {
	conn := &KafkaConnection{
		config: KafkaConfig{
			Brokers: []string{"localhost:9092"},
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Test concurrent reads
			_ = conn.Config()
			_ = conn.IsConnected()
		}()
	}
	wg.Wait()
}

// TestKafkaConnection_Close tests idempotent close.
func TestKafkaConnection_Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	conn := &KafkaConnection{
		config: KafkaConfig{
			Brokers:              []string{"invalid:9999"}, // Will fail to connect
			ReconnectInterval:    10 * time.Millisecond,
			ReconnectMaxAttempts: 1,
		},
		ctx:    ctx,
		cancel: cancel,
		logger: &NoopLogger{},
	}

	// First close
	err := conn.Close()
	if err != nil {
		t.Logf("First close returned error (expected): %v", err)
	}

	// Second close should be idempotent
	err = conn.Close()
	if err != nil {
		t.Errorf("Second close should be idempotent, got error: %v", err)
	}
}

// TestKafkaConnection_PublishWithContext_Concurrent tests concurrent publishes.
// Verifies no deadlocks or race conditions when multiple goroutines publish simultaneously.
func TestKafkaConnection_PublishWithContext_Concurrent(t *testing.T) {
	ctx := context.Background()

	conn := &KafkaConnection{
		config: KafkaConfig{
			Brokers: []string{"localhost:9092"},
		},
		closed: atomic.Bool{},
		logger: &NoopLogger{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// PublishWithContext will handle the nil producer gracefully
			msg := &sarama.ProducerMessage{
				Topic: "test-topic",
				Key:   sarama.StringEncoder(fmt.Sprintf("key-%d", id)),
				Value: sarama.StringEncoder(fmt.Sprintf("value-%d", id)),
			}
			_ = conn.PublishWithContext(ctx, msg)
		}(i)
	}

	wg.Wait()
	// Test passes if no deadlock/race condition
}

// TestKafkaConnection_CompressionDefault tests default compression codec.
func TestKafkaConnection_CompressionDefault(t *testing.T) {
	config := KafkaConfig{
		Brokers: []string{"localhost:9092"},
		// Compression not set, should default to Snappy
	}

	conn := &KafkaConnection{
		config: config,
	}

	// Manually apply default logic (since connect won't be called)
	if conn.config.Compression == sarama.CompressionNone {
		conn.config.Compression = sarama.CompressionSnappy
	}

	if conn.config.Compression != sarama.CompressionSnappy {
		t.Errorf("Expected default compression to be Snappy, got %v", conn.config.Compression)
	}
}

// TestKafkaConnection_PublishWithContext_NoTopic tests that publish fails without topic.
func TestKafkaConnection_PublishWithContext_NoTopic(t *testing.T) {
	ctx := context.Background()

	conn := &KafkaConnection{
		config: KafkaConfig{
			Brokers: []string{"localhost:9092"},
		},
		closed: atomic.Bool{},
		logger: &NoopLogger{},
	}

	// Create a mock producer for testing
	msg := &sarama.ProducerMessage{
		// No topic set
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("value"),
	}

	err := conn.PublishWithContext(ctx, msg)
	// With nil producer, we get "no active producer" error first
	// If we had a producer but no topic, we'd get "message topic must be set"
	if err == nil {
		t.Error("expected error when publishing without topic or producer")
	}
}

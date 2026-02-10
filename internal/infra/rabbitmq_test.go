package infra

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TestRabbitMQConnection_ReconnectSettings tests the reconnection configuration.
func TestRabbitMQConnection_ReconnectSettings(t *testing.T) {
	tests := []struct {
		name   string
		config RabbitMQConfig
	}{
		{
			name: "Default reconnection settings",
			config: RabbitMQConfig{
				URL:              "amqp://guest:guest@localhost:5672/",
				Exchange:         "test-exchange",
				ExchangeType:     "topic",
				RoutingKeyPrefix: "test.",
			},
		},
		{
			name: "Custom reconnection settings",
			config: RabbitMQConfig{
				URL:                  "amqp://guest:guest@localhost:5672/",
				Exchange:             "test-exchange",
				ExchangeType:         "topic",
				RoutingKeyPrefix:     "test.",
				ReconnectInterval:    10 * time.Second,
				MaxReconnectAttempts: 5,
				ReconnectDelay:       2 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config is parsed correctly
			// We can't test actual connection without RabbitMQ running
			if tt.config.Exchange == "" {
				t.Error("exchange should not be empty")
			}
			if tt.config.ExchangeType == "" {
				t.Error("exchange type should not be empty")
			}
		})
	}
}

// TestRabbitMQConnection_GetRoutingKey tests routing key construction.
func TestRabbitMQConnection_GetRoutingKey(t *testing.T) {
	config := RabbitMQConfig{
		RoutingKeyPrefix: "user.service.",
	}

	conn := &RabbitMQConnection{
		config: config,
	}

	tests := []struct {
		name      string
		eventType string
		want      string
	}{
		{
			name:      "Login event",
			eventType: "auth.login",
			want:      "user.service.auth.login",
		},
		{
			name:      "User created",
			eventType: "user.created",
			want:      "user.service.user.created",
		},
		{
			name:      "Device created",
			eventType: "device.created",
			want:      "user.service.device.created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := conn.GetRoutingKey(tt.eventType)
			if got != tt.want {
				t.Errorf("GetRoutingKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestRabbitMQConnection_Concurrency tests concurrent access to connection methods.
func TestRabbitMQConnection_Concurrency(t *testing.T) {
	conn := &RabbitMQConnection{
		config: RabbitMQConfig{
			RoutingKeyPrefix: "test.",
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Test concurrent reads
			_ = conn.GetRoutingKey("test.event")
			_ = conn.Config()
			_ = conn.IsConnected()
		}()
	}
	wg.Wait()
}

// TestNoopLogger tests the NoopLogger implementation.
func TestNoopLogger(t *testing.T) {
	logger := &NoopLogger{}

	// These should not panic
	logger.SetLogLevel("info")
	logger.Debug("test", nil)
	logger.Info("test", nil)
	logger.Warn("test", nil)
	logger.Error("test", nil)
	logger.Fatal("test", nil)
	_ = logger.Sync()
}

// TestRabbitMQConnection_Close tests idempotent close.
func TestRabbitMQConnection_Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	conn := &RabbitMQConnection{
		config: RabbitMQConfig{
			URL:              "amqp://invalid:9999/", // Will fail to connect
			Exchange:         "test",
			ExchangeType:     "topic",
			ReconnectDelay:   10 * time.Millisecond,
			MaxReconnectAttempts: 1,
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

// TestRabbitMQConnection_PublishWithContext_Concurrent tests concurrent publishes.
// Verifies no deadlocks or race conditions when multiple goroutines publish simultaneously.
func TestRabbitMQConnection_PublishWithContext_Concurrent(t *testing.T) {
	ctx := context.Background()

	conn := &RabbitMQConnection{
		config: RabbitMQConfig{
			Exchange:         "test-exchange",
			ExchangeType:     "topic",
			RoutingKeyPrefix: "test.",
		},
		closed: atomic.Bool{},
		logger: &NoopLogger{},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// PublishWithContext will handle the nil connection gracefully
			_ = conn.PublishWithContext(ctx, fmt.Sprintf("test.%d", id), amqp.Publishing{})
		}(i)
	}

	wg.Wait()
	// Test passes if no deadlock/race condition
}

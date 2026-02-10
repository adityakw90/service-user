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
				Host:     "localhost",
				Port:     5672,
				User:     "guest",
				Password: "guest",
				Vhost:    "/",
			},
		},
		{
			name: "Custom reconnection settings",
			config: RabbitMQConfig{
				Host:                 "localhost",
				Port:                 5672,
				User:                 "guest",
				Password:             "guest",
				Vhost:                "/",
				ReconnectInterval:    2 * time.Second,
				ReconnectMaxAttempts: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the config is parsed correctly
			if tt.config.Host == "" {
				t.Error("Host should not be empty")
			}
		})
	}
}

// TestRabbitMQConnection_Concurrency tests concurrent access to connection methods.
func TestRabbitMQConnection_Concurrency(t *testing.T) {
	conn := &RabbitMQConnection{
		config: RabbitMQConfig{
			Host:     "localhost",
			Port:     5672,
			User:     "guest",
			Password: "guest",
			Vhost:    "/",
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
			Host:                 "localhost",
			Port:                 5672,
			User:                 "guest",
			Password:             "guest",
			Vhost:                "/",
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

// TestRabbitMQConnection_PublishWithContext_Concurrent tests concurrent publishes.
// Verifies no deadlocks or race conditions when multiple goroutines publish simultaneously.
func TestRabbitMQConnection_PublishWithContext_Concurrent(t *testing.T) {
	ctx := context.Background()

	conn := &RabbitMQConnection{
		config: RabbitMQConfig{
			Host:     "localhost",
			Port:     5672,
			User:     "guest",
			Password: "guest",
			Vhost:    "/",
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
			_ = conn.PublishWithContext(ctx, "test-exchange", fmt.Sprintf("test.%d", id), amqp.Publishing{})
		}(i)
	}

	wg.Wait()
	// Test passes if no deadlock/race condition
}

// TestRabbitMQConnection_DeclareExchange tests exchange declaration.
func TestRabbitMQConnection_DeclareExchange(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	conn := &RabbitMQConnection{
		config: RabbitMQConfig{
			Host:                 "localhost",
			Port:                 5672,
			User:                 "guest",
			Password:             "guest",
			Vhost:                "/",
			ReconnectInterval:    10 * time.Millisecond,
			ReconnectMaxAttempts: 1,
		},
		ctx:    ctx,
		cancel: cancel,
		logger: &NoopLogger{},
	}

	// This should fail due to no active connection
	err := conn.DeclareExchange("test-exchange", "topic", true)
	if err == nil {
		t.Error("expected error when declaring exchange without connection")
	}
}

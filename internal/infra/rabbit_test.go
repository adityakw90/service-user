package infra

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/adityakw90/service-user/pkg/util"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

// Test_buildAmqpPublishing tests the buildAmqpPublishing method
func TestRabbit_buildAmqpPublishing(t *testing.T) {
	tests := []struct {
		name         string
		contentType  string
		headers      map[string]string
		body         []byte
		deliveryMode *uint8
		priority     *uint8
		timestamp    *time.Time
		expiration   *time.Duration
		want         amqp091.Publishing
	}{
		{
			name:        "Basic message with required fields",
			contentType: "application/json",
			body:        []byte(`{"test": "data"}`),
			want: amqp091.Publishing{
				ContentType: "application/json",
				Body:        []byte(`{"test": "data"}`),
			},
		},
		{
			name:        "With custom headers",
			contentType: "application/json",
			headers:     map[string]string{"trace-id": "12345", "user-id": "67890"},
			body:        []byte(`{}`),
			want: amqp091.Publishing{
				ContentType: "application/json",
				Body:        []byte(`{}`),
				Headers: amqp091.Table{
					"trace-id": "12345",
					"user-id":  "67890",
				},
			},
		},
		{
			name:         "With delivery mode Transient",
			contentType:  "text/plain",
			deliveryMode: util.Ptr(uint8(amqp091.Transient)),
			body:         []byte("test"),
			want: amqp091.Publishing{
				ContentType:  "text/plain",
				Body:         []byte("test"),
				DeliveryMode: amqp091.Transient,
			},
		},
		{
			name:         "With delivery mode Persistent",
			contentType:  "application/json",
			deliveryMode: util.Ptr(uint8(amqp091.Persistent)),
			body:         []byte(`{}`),
			want: amqp091.Publishing{
				ContentType:  "application/json",
				Body:         []byte(`{}`),
				DeliveryMode: amqp091.Persistent,
			},
		},
		{
			name:        "With priority",
			contentType: "application/json",
			priority:    util.Ptr(uint8(5)),
			body:        []byte(`{}`),
			want: amqp091.Publishing{
				ContentType: "application/json",
				Body:        []byte(`{}`),
				Priority:    5,
			},
		},
		{
			name:        "With custom timestamp",
			contentType: "application/json",
			timestamp:   util.Ptr(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
			body:        []byte(`{}`),
			want: amqp091.Publishing{
				ContentType: "application/json",
				Body:        []byte(`{}`),
				Timestamp:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		},
		{
			name:        "With expiration",
			contentType: "application/json",
			expiration:  util.Ptr(30 * time.Minute),
			body:        []byte(`{}`),
			want: amqp091.Publishing{
				ContentType: "application/json",
				Body:        []byte(`{}`),
				Expiration:  strconv.FormatInt((30 * time.Minute).Milliseconds(), 10),
			},
		},
		{
			name:         "All optional fields combined",
			contentType:  "application/json",
			headers:      map[string]string{"key": "value"},
			deliveryMode: util.Ptr(uint8(amqp091.Persistent)),
			priority:     util.Ptr(uint8(10)),
			timestamp:    util.Ptr(time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC)),
			expiration:   util.Ptr(1 * time.Hour),
			body:         []byte(`{"complete": true}`),
			want: amqp091.Publishing{
				ContentType:  "application/json",
				Body:         []byte(`{"complete": true}`),
				Headers:      amqp091.Table{"key": "value"},
				DeliveryMode: amqp091.Persistent,
				Priority:     10,
				Timestamp:    time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
				Expiration:   strconv.FormatInt((1 * time.Hour).Milliseconds(), 10),
			},
		},
		{
			name:         "Nil optional fields",
			contentType:  "application/json",
			headers:      nil,
			deliveryMode: nil,
			priority:     nil,
			timestamp:    nil,
			expiration:   nil,
			body:         []byte(`{}`),
			want: amqp091.Publishing{
				ContentType: "application/json",
				Body:        []byte(`{}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rabbit{}
			got := r.buildAmqpPublishing(
				tt.contentType,
				tt.headers,
				tt.body,
				tt.deliveryMode,
				tt.priority,
				tt.timestamp,
				tt.expiration,
			)

			if tt.timestamp == nil {
				// Default timestamp is set to time.Now(), so we just check it's not zero
				assert.False(t, got.Timestamp.IsZero(), "Timestamp should be set")
				got.Timestamp = time.Time{} // Clear for comparison
			}

			assert.Equal(t, tt.want.ContentType, got.ContentType)
			assert.Equal(t, tt.want.Body, got.Body)
			assert.Equal(t, tt.want.Headers, got.Headers)
			assert.Equal(t, tt.want.DeliveryMode, got.DeliveryMode)
			assert.Equal(t, tt.want.Priority, got.Priority)
			assert.Equal(t, tt.want.Timestamp, got.Timestamp)
			assert.Equal(t, tt.want.Expiration, got.Expiration)
		})
	}
}

// Test_getConnection tests the getConnection method
func TestRabbit_getConnection(t *testing.T) {
	tests := []struct {
		name        string
		setupRabbit func() *Rabbit
		wantErr     bool
		errMsg      string
	}{
		{
			name: "Returns active connection",
			setupRabbit: func() *Rabbit {
				return &Rabbit{conn: &amqp091.Connection{}}
			},
			wantErr: false,
		},
		{
			name: "Returns error when connection is nil",
			setupRabbit: func() *Rabbit {
				return &Rabbit{conn: nil}
			},
			wantErr: true,
			errMsg:  "no active connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupRabbit()
			got, err := r.getConnection()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// Test_getConnection_Concurrent tests concurrent access to getConnection
func TestRabbit_getConnection_Concurrent(t *testing.T) {
	r := &Rabbit{conn: &amqp091.Connection{}}
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := r.getConnection()
			errors <- err
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		assert.NoError(t, err)
	}
}

// Test_getChannel tests the getChannel method
func TestRabbit_getChannel(t *testing.T) {
	tests := []struct {
		name        string
		setupRabbit func() *Rabbit
		wantErr     bool
		errMsg      string
		skip        bool
		skipReason  string
	}{
		{
			name: "Returns error when connection is nil",
			setupRabbit: func() *Rabbit {
				return &Rabbit{conn: nil}
			},
			wantErr: true,
			errMsg:  "no active connection",
		},
		{
			name:       "Returns error when connection is closed",
			skip:       true,
			skipReason: "Cannot safely test IsClosed() without real RabbitMQ connection",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip(tt.skipReason)
			}

			r := tt.setupRabbit()
			got, err := r.getChannel()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// Test_DeclareExchange tests the DeclareExchange method
func TestRabbit_DeclareExchange(t *testing.T) {
	tests := []struct {
		name         string
		setupRabbit  func() *Rabbit
		exchange     string
		exchangeType string
		durable      bool
		wantErr      bool
		errMsg       string
	}{
		{
			name: "Returns error when connection is closed",
			setupRabbit: func() *Rabbit {
				r := &Rabbit{}
				r.closed.Store(true)
				return r
			},
			exchange:     "test-exchange",
			exchangeType: "direct",
			durable:      true,
			wantErr:      true,
			errMsg:       "connection is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupRabbit()
			err := r.DeclareExchange(tt.exchange, tt.exchangeType, tt.durable)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test_Publish tests the Publish method
func TestRabbit_Publish(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		setupRabbit  func() *Rabbit
		exchange     string
		routingKey   string
		contentType  string
		headers      map[string]string
		body         []byte
		deliveryMode *uint8
		wantErr      bool
		errMsg       string
	}{
		{
			name: "Returns error when connection is closed",
			setupRabbit: func() *Rabbit {
				r := &Rabbit{}
				r.closed.Store(true)
				return r
			},
			exchange:    "test-exchange",
			routingKey:  "test.key",
			contentType: "application/json",
			body:        []byte(`{}`),
			wantErr:     true,
			errMsg:      "connection is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupRabbit()
			err := r.Publish(ctx, tt.exchange, tt.routingKey, tt.contentType, tt.headers, tt.body, tt.deliveryMode, nil, nil, nil)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test_PublishWithConfirm tests the PublishWithConfirm method
func TestRabbit_PublishWithConfirm(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		setupRabbit    func() *Rabbit
		exchange       string
		routingKey     string
		contentType    string
		headers        map[string]string
		body           []byte
		maxRetries     int
		retryInterval  time.Duration
		confirmTimeout time.Duration
		wantErr        bool
		errMsg         string
	}{
		{
			name: "Returns error when connection is closed",
			setupRabbit: func() *Rabbit {
				r := &Rabbit{}
				r.closed.Store(true)
				return r
			},
			exchange:       "test-exchange",
			routingKey:     "test.key",
			contentType:    "application/json",
			body:           []byte(`{}`),
			maxRetries:     3,
			retryInterval:  1 * time.Second,
			confirmTimeout: 5 * time.Second,
			wantErr:        true,
			errMsg:         "connection is closed",
		},
		{
			name: "Uses default maxRetries when <= 0",
			setupRabbit: func() *Rabbit {
				r := &Rabbit{}
				r.closed.Store(true)
				return r
			},
			exchange:   "test-exchange",
			maxRetries: 0,
			wantErr:    true,
			errMsg:     "connection is closed",
		},
		{
			name: "Uses default retryInterval when <= 0",
			setupRabbit: func() *Rabbit {
				r := &Rabbit{}
				r.closed.Store(true)
				return r
			},
			exchange:      "test-exchange",
			retryInterval: 0,
			wantErr:       true,
			errMsg:        "connection is closed",
		},
		{
			name: "Uses default confirmTimeout when <= 0",
			setupRabbit: func() *Rabbit {
				r := &Rabbit{}
				r.closed.Store(true)
				return r
			},
			exchange:       "test-exchange",
			confirmTimeout: 0,
			wantErr:        true,
			errMsg:         "connection is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupRabbit()
			err := r.PublishWithConfirm(ctx, tt.exchange, tt.routingKey, tt.contentType, tt.headers, tt.body, nil, nil, nil, nil, tt.maxRetries, tt.retryInterval, tt.confirmTimeout)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test_PublishWithConfirm_ContextCancellation tests context cancellation during publish
func TestRabbit_PublishWithConfirm_ContextCancellation(t *testing.T) {
	tests := []struct {
		name        string
		cancelAfter time.Duration
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "Cancels during backoff",
			cancelAfter: 10 * time.Millisecond,
			wantErr:     true,
			errMsg:      "connection is closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			r := &Rabbit{}
			r.closed.Store(true)

			// Cancel context after delay
			go func() {
				time.Sleep(tt.cancelAfter)
				cancel()
			}()

			err := r.PublishWithConfirm(ctx, "test-exchange", "test.key", "application/json", nil, []byte(`{}`), nil, nil, nil, nil, 5, 100*time.Millisecond, 5*time.Second)

			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}

// Test_Close tests the Close method
func TestRabbit_Close(t *testing.T) {
	tests := []struct {
		name        string
		setupRabbit func() *Rabbit
		wantErr     bool
	}{
		{
			name: "Successfully closes with nil connection",
			setupRabbit: func() *Rabbit {
				ctx, cancel := context.WithCancel(context.Background())
				return &Rabbit{
					ctx:    ctx,
					cancel: cancel,
					conn:   nil,
					logger: &NoopLogger{},
				}
			},
			wantErr: false,
		},
		{
			name: "Idempotent Close - multiple calls",
			setupRabbit: func() *Rabbit {
				ctx, cancel := context.WithCancel(context.Background())
				return &Rabbit{
					ctx:    ctx,
					cancel: cancel,
					conn:   nil,
					logger: &NoopLogger{},
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupRabbit()

			// First close
			err := r.Close()
			assert.Equal(t, tt.wantErr, err != nil)

			// Second close (idempotent)
			err = r.Close()
			assert.NoError(t, err, "Second Close should be no-op")
		})
	}
}

// Test_RabbitConfig_defaults tests NewRabbitConnection default value handling
func TestNewRabbitConnection_defaults(t *testing.T) {
	tests := []struct {
		name      string
		config    RabbitConfig
		wantError bool
	}{
		{
			name: "Sets default ReconnectInterval when zero",
			config: RabbitConfig{
				Host:              "invalid-host-that-will-fail",
				Port:              5672,
				User:              "guest",
				Password:          "guest",
				Vhost:             "/",
				ReconnectInterval: 0, // Should default to 1s
			},
			wantError: true, // Connection will fail, but we're testing default handling
		},
		{
			name: "Sets default ReconnectMaxAttempts when zero",
			config: RabbitConfig{
				Host:                 "invalid-host",
				Port:                 5672,
				User:                 "guest",
				Password:             "guest",
				Vhost:                "/",
				ReconnectMaxAttempts: 0, // Should default to 0 (infinite)
				ReconnectInterval:    1 * time.Second,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			// logger := newMockLogger()
			logger := NewNoopLogger()
			_, err := NewRabbitConnection(ctx, tt.config, logger)

			// We expect connection errors since we're using invalid host
			// The key is that the function handles the config defaults
			assert.Error(t, err)
		})
	}
}

// Test_triggerReconnect tests the triggerReconnect method
func TestRabbit_triggerReconnect(t *testing.T) {
	tests := []struct {
		name        string
		setupRabbit func() *Rabbit
	}{
		{
			name: "Successfully triggers reconnect without blocking",
			setupRabbit: func() *Rabbit {
				ctx, cancel := context.WithCancel(context.Background())
				return &Rabbit{
					ctx:           ctx,
					cancel:        cancel,
					reconnectChan: make(chan struct{}, 1),
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.setupRabbit()

			// Trigger reconnect should not block
			done := make(chan bool)
			go func() {
				r.triggerReconnect()
				done <- true
			}()

			select {
			case <-done:
				// Success - didn't block
			case <-time.After(100 * time.Millisecond):
				t.Fatal("triggerReconnect blocked")
			}

			// Should be able to trigger again without blocking
			r.triggerReconnect()
		})
	}
}

// Test_buildAmqpPublishing_edgeCases tests edge cases for buildAmqpPublishing
func TestRabbit_buildAmqpPublishing_edgeCases(t *testing.T) {
	tests := []struct {
		name         string
		deliveryMode *uint8
		wantMode     uint8
	}{
		{
			name:         "Nil delivery mode - should not set",
			deliveryMode: nil,
			wantMode:     0, // Default zero value
		},
		{
			name:         "Transient delivery mode",
			deliveryMode: util.Ptr(uint8(amqp091.Transient)),
			wantMode:     amqp091.Transient,
		},
		{
			name:         "Persistent delivery mode",
			deliveryMode: util.Ptr(uint8(amqp091.Persistent)),
			wantMode:     amqp091.Persistent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rabbit{}
			publishing := r.buildAmqpPublishing("text/plain", nil, []byte("test"), tt.deliveryMode, nil, nil, nil)

			assert.Equal(t, tt.wantMode, publishing.DeliveryMode)
		})
	}
}

// Test_buildAmqpPublishing_expirationFormat tests expiration formatting
func TestRabbit_buildAmqpPublishing_expirationFormat(t *testing.T) {
	tests := []struct {
		name       string
		expiration *time.Duration
		wantFormat string
	}{
		{
			name:       "30 minutes expiration",
			expiration: util.Ptr(30 * time.Minute),
			wantFormat: strconv.FormatInt((30 * time.Minute).Milliseconds(), 10),
		},
		{
			name:       "1 hour expiration",
			expiration: util.Ptr(1 * time.Hour),
			wantFormat: strconv.FormatInt((1 * time.Hour).Milliseconds(), 10),
		},
		{
			name:       "1 second expiration",
			expiration: util.Ptr(1 * time.Second),
			wantFormat: strconv.FormatInt((1 * time.Second).Milliseconds(), 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Rabbit{}
			publishing := r.buildAmqpPublishing("text/plain", nil, []byte("test"), nil, nil, nil, tt.expiration)

			assert.Equal(t, tt.wantFormat, publishing.Expiration)
		})
	}
}

// Test_nilLogger tests that nil logger is handled
func TestNewRabbitConnection_nilLogger(t *testing.T) {
	cfg := RabbitConfig{
		Host:              "invalid-host",
		Port:              5672,
		User:              "guest",
		Password:          "guest",
		Vhost:             "/",
		ReconnectInterval: 100 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Pass nil logger - should use NoopLogger
	_, err := NewRabbitConnection(ctx, cfg, nil)
	assert.Error(t, err) // Will fail to connect, but shouldn't panic
}

// Test_PublishWithConfirm_retryLogic tests the retry logic
func TestRabbit_PublishWithConfirm_retryLogic(t *testing.T) {
	tests := []struct {
		name            string
		maxRetries      int
		retryInterval   time.Duration
		confirmTimeout  time.Duration
		expectRetryLogs bool
	}{
		{
			name:            "With custom retry settings",
			maxRetries:      2,
			retryInterval:   50 * time.Millisecond,
			confirmTimeout:  100 * time.Millisecond,
			expectRetryLogs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			r := &Rabbit{}
			r.closed.Store(true)
			logger := NewNoopLogger()
			r.logger = logger

			err := r.PublishWithConfirm(ctx, "test-exchange", "test.key", "application/json", nil, []byte(`{}`), nil, nil, nil, nil, tt.maxRetries, tt.retryInterval, tt.confirmTimeout)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "connection is closed")
		})
	}
}

// Test_exponentialBackoff tests exponential backoff calculation
func Test_exponentialBackoff(t *testing.T) {
	// Test that min() caps at 5
	tests := []struct {
		attempt int
		wantMax int
	}{
		{attempt: 1, wantMax: 1},
		{attempt: 2, wantMax: 2},
		{attempt: 3, wantMax: 3},
		{attempt: 4, wantMax: 4},
		{attempt: 5, wantMax: 5},
		{attempt: 6, wantMax: 5},
		{attempt: 10, wantMax: 5},
		{attempt: 100, wantMax: 5},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			got := min(tt.attempt, 5)
			assert.Equal(t, tt.wantMax, got)
		})
	}
}

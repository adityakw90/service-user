package publisher

import (
	"context"
	"fmt"
	"sync/atomic"

	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
)

// namedPublisher wraps a publisher with its backend name for metrics/logging.
type namedPublisher struct {
	publisher portEvent.EventPublisher
	backend   string
}

// MultiPublisher publishes events to multiple backend publishers.
// It provides fire-and-forget semantics where failure in one backend
// does not prevent publishing to others.
type MultiPublisher struct {
	publishers []namedPublisher
	logger     gomon.Logger
	metrics    *multiMetrics
}

// multiMetrics holds atomic metrics for the multi publisher.
type multiMetrics struct {
	successTotal atomic.Int64
	failureTotal atomic.Int64
}

// NewMultiPublisher creates a publisher that fans out to multiple backends.
// Logger must not be nil.
func NewMultiPublisher(logger gomon.Logger, publishers ...portEvent.EventPublisher) (portEvent.EventPublisher, error) {
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}
	if len(publishers) == 0 {
		return NewNoOpPublisher(), nil
	}
	if len(publishers) == 1 {
		return publishers[0], nil
	}

	named := make([]namedPublisher, len(publishers))
	for i, p := range publishers {
		named[i] = namedPublisher{
			publisher: p,
			backend:   inferBackendName(p),
		}
	}

	return &MultiPublisher{
		publishers: named,
		logger:     logger,
		metrics:    &multiMetrics{},
	}, nil
}

// inferBackendName attempts to infer the backend name from the publisher type.
func inferBackendName(p portEvent.EventPublisher) string {
	switch p.(type) {
	case *RedisPublisher:
		return RedisStreamBackend
	case *RabbitMQPublisher:
		return "rabbitmq"
	case *KafkaPublisher:
		return "kafka"
	case *HTTPPublisher:
		return "http"
	default:
		return "unknown"
	}
}

// Publish publishes to all backend publishers.
// Errors are logged and emitted as metrics but do not fail-fast.
// All backends are attempted regardless of individual failures.
func (m *MultiPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	var firstErr error

	for _, np := range m.publishers {
		err := np.publisher.Publish(ctx, eventType, eventData)
		if err != nil {
			m.emitFailure(np.backend, eventType)
			m.logger.Error("backend publish failed", map[string]interface{}{
				"backend":    np.backend,
				"event_type": eventType,
				"error":      err.Error(),
			})
			if firstErr == nil {
				firstErr = err
			}
		} else {
			m.emitSuccess(np.backend, eventType)
			m.logger.Debug("backend publish succeeded", map[string]interface{}{
				"backend":    np.backend,
				"event_type": eventType,
			})
		}
	}

	// Return error for observability, but only after trying all backends
	return firstErr
}

// Close closes all backend publishers.
func (m *MultiPublisher) Close() error {
	for _, np := range m.publishers {
		_ = np.publisher.Close()
	}
	return nil
}

// emitSuccess records a successful publish event.
func (m *MultiPublisher) emitSuccess(backend string, eventType event.EventType) {
	m.metrics.successTotal.Add(1)
	_ = backend
	_ = eventType
}

// emitFailure records a failed publish event.
func (m *MultiPublisher) emitFailure(backend string, eventType event.EventType) {
	m.metrics.failureTotal.Add(1)
	_ = backend
	_ = eventType
}

// GetMetrics returns the current metrics for observability.
func (m *MultiPublisher) GetMetrics() (success, failure int64) {
	return m.metrics.successTotal.Load(), m.metrics.failureTotal.Load()
}

// String returns a string representation of the multi publisher.
func (m *MultiPublisher) String() string {
	return fmt.Sprintf("MultiPublisher{backends=%d}", len(m.publishers))
}

// BackendNames returns the names of all backends.
func (m *MultiPublisher) BackendNames() []string {
	names := make([]string, len(m.publishers))
	for i, np := range m.publishers {
		names[i] = np.backend
	}
	return names
}

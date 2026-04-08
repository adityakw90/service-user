package observer

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"github.com/adityakw90/service-user/internal/core/port/observer"
)

// NoopObserver is a no-op observer for testing or when observers are disabled.
type NoopObserver[T any] struct{}

// NewNoopObserver creates a no-op observer.
func NewNoopObserver[T any]() observer.ServiceObserver[T] {
	return &NoopObserver[T]{}
}

// OnSignal is a no-op.
func (o *NoopObserver[T]) OnSignal(ctx context.Context, sig signal.SignalType, data T, err error) {
	// No-op - does nothing
}

func (o *NoopObserver[T]) DetachContext(ctx context.Context) context.Context {
	return context.Background()
}

package observer

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/signal"
)

func TestAdapter_Observer_NoopObserver_OnSignal_DoesNotPanic(t *testing.T) {
	noop := NewNoopObserver[signal.AuthSignal]()
	ctx := context.Background()

	// Should not panic
	noop.OnSignal(ctx, signal.SignalStart, signal.AuthSignal{
		Identifier: "test@example.com",
	}, nil)

	noop.OnSignal(ctx, signal.SignalFail, signal.AuthSignal{
		Identifier: "test@example.com",
	}, nil)
}

func TestAdapter_Observer_NoopObserver_Generic(t *testing.T) {
	// Test with different types - for now just AuthSignal
	// UserSignal will be added in Task 2
	authNoop := NewNoopObserver[signal.AuthSignal]()

	ctx := context.Background()

	// Should not panic with any signal type
	authNoop.OnSignal(ctx, signal.SignalSuccess, signal.AuthSignal{}, nil)
}

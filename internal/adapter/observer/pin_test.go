package observer

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/signal"
)

func TestAdapter_Observer_NewPinObserver(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewPinObserver(logger, tracer)

	if obs == nil {
		t.Fatal("NewPinObserver() returned nil")
	}
}

func TestAdapter_Observer_PinObserver_OnSignal_Success(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewPinObserver(logger, tracer)
	ctx := context.Background()

	success := true

	obs.OnSignal(ctx, signal.SignalSuccess, signal.PinSignal{
		UserUID:   "user-123",
		Operation: "verify",
		Success:   &success,
	}, nil)

	if len(logger.debugMessages) != 1 {
		t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
	}
}

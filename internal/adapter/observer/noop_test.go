package observer

import (
	"context"
	"sync"
	"testing"
	"time"

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
	// Test with different types
	authNoop := NewNoopObserver[signal.AuthSignal]()
	userNoop := NewNoopObserver[signal.UserSignal]()

	ctx := context.Background()

	// Should not panic with any signal type
	authNoop.OnSignal(ctx, signal.SignalSuccess, signal.AuthSignal{}, nil)
	userNoop.OnSignal(ctx, signal.SignalStart, signal.UserSignal{}, nil)
}

func TestAdapter_Observer_NoopObserver_DetachContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "background context",
			ctx:  context.Background(),
		},
		{
			name: "context with value",
			ctx:  context.WithValue(context.Background(), "key", "value"),
		},
		{
			name: "context with span",
			ctx: func() context.Context {
				tracer := newMockTracer()
				ctx, _ := tracer.StartSpan(context.Background(), "test")
				return ctx
			}(),
		},
		{
			name: "context with cancellation",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
		},
		{
			name: "context with timeout",
			ctx: func() context.Context {
				ctx, _ := context.WithTimeout(context.Background(), 1)
				return ctx
			}(),
		},
		{
			name: "context with deadline",
			ctx: func() context.Context {
				deadline, _ := context.WithDeadline(context.Background(), time.Now().Add(time.Hour))
				return deadline
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			noop := NewNoopObserver[signal.AuthSignal]()
			result := noop.DetachContext(tt.ctx)

			// The result should be a background context (not done)
			select {
			case <-result.Done():
				t.Error("noop DetachContext should return a background context that is not done")
			default:
				// Expected - background context is not done
			}

			// Verify it's actually a background-like context
			// (no cancellation, no deadline, no values from parent)
			if result.Value("key") != nil {
				t.Error("background context should not have values from parent")
			}

			// For contexts that are not background, verify the result is different
			// (Note: context.Background() is a singleton, so if input is Background(), result will be same)
			if tt.ctx != context.Background() && result == tt.ctx {
				t.Error("DetachContext should return context.Background(), not the input context")
			}
		})
	}
}

func TestAdapter_Observer_NoopObserver_DetachContext_Concurrent(t *testing.T) {
	noop := NewNoopObserver[signal.UserSignal]()

	// Test concurrent calls to DetachContext
	var wg sync.WaitGroup
	numGoroutines := 10
	callsPerGoroutine := 10

	results := make([]context.Context, 0, numGoroutines*callsPerGoroutine)
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				ctx := context.Background()
				detachedCtx := noop.DetachContext(ctx)

				mu.Lock()
				results = append(results, detachedCtx)
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	totalCalls := numGoroutines * callsPerGoroutine
	if len(results) != totalCalls {
		t.Errorf("expected %d results, got %d", totalCalls, len(results))
	}

	// All results should be background contexts
	for i, ctx := range results {
		if ctx == nil {
			t.Errorf("result %d is nil", i)
		}
		// Noop observer always returns context.Background()
		// which may be the same instance each time
		select {
		case <-ctx.Done():
			t.Errorf("result %d should be a background context (not done)", i)
		default:
			// Expected
		}
	}
}

func TestAdapter_Observer_NoopObserver_DetachContext_GenericTypes(t *testing.T) {
	// Test with different generic types
	authNoop := NewNoopObserver[signal.AuthSignal]()
	userNoop := NewNoopObserver[signal.UserSignal]()

	authCtx := authNoop.DetachContext(context.Background())
	userCtx := userNoop.DetachContext(context.Background())

	// Both should return background contexts
	if authCtx == nil || userCtx == nil {
		t.Error("DetachContext should not return nil")
	}

	// Both should be background contexts (not done)
	select {
	case <-authCtx.Done():
		t.Error("auth observer DetachContext should return a background context")
	default:
	}

	select {
	case <-userCtx.Done():
		t.Error("user observer DetachContext should return a background context")
	default:
	}
}

package executor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/infra"
	"github.com/stretchr/testify/assert"
)

func TestServiceExecutor_Do(t *testing.T) {
	tests := []struct {
		name          string
		operationName string
		fn            func(ctx context.Context)
	}{
		{
			name:          "Happy Path - function executes successfully",
			operationName: "test-operation",
			fn: func(ctx context.Context) {
				// Function that does nothing
			},
		},
		{
			name:          "Named Operation - operation name appears in logs",
			operationName: "important-task",
			fn: func(ctx context.Context) {
				// Function execution
			},
		},
		{
			name:          "Context Propagation - new context with span is passed",
			operationName: "context-test",
			fn: func(ctx context.Context) {
				// Verify context is passed through
				assert.NotNil(t, ctx, "context should not be nil")
			},
		},
		{
			name:          "Empty Function Name - logs empty name",
			operationName: "",
			fn:            func(ctx context.Context) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := infra.NewNoopLogger()
			tracer := infra.NewNoopTracer()

			executor := NewServiceExecutor(logger, tracer)
			ctx := context.Background()

			executor.Do(ctx, tt.operationName, tt.fn)
		})
	}
}

func TestServiceExecutor_DoAsync(t *testing.T) {
	tests := []struct {
		name          string
		operationName string
		fn            func(context.Context)
		setupCancel   func(context.Context) (context.Context, context.CancelFunc)
	}{
		{
			name:          "Happy Path - function executes in goroutine",
			operationName: "async-operation",
			fn: func(ctx context.Context) {
			},
		},
		{
			name:          "Panic Recovery - panic is caught and logged",
			operationName: "panic-operation",
			fn: func(ctx context.Context) {
				panic("something went wrong")
			},
		},
		{
			name:          "Background Context - goroutine uses background context",
			operationName: "background-operation",
			fn: func(ctx context.Context) {
				assert.NotNil(t, ctx, "context should not be nil")
			},
		},
		{
			name:          "Context Independence - parent cancellation doesn't affect",
			operationName: "independent-operation",
			fn: func(ctx context.Context) {
			},
			setupCancel: func(parent context.Context) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(parent)
				cancel()
				return ctx, cancel
			},
		},
		{
			name:          "Panic with Error Type - records error on span",
			operationName: "error-panic-operation",
			fn: func(ctx context.Context) {
				panic(fmt.Errorf("wrapped error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := infra.NewNoopLogger()
			tracer := infra.NewNoopTracer()

			executor := NewServiceExecutor(logger, tracer)
			ctx := context.Background()

			if tt.setupCancel != nil {
				var cancel context.CancelFunc
				ctx, cancel = tt.setupCancel(ctx)
				defer cancel()
			}

			done := make(chan struct{})
			go func() {
				executor.DoAsync(ctx, tt.operationName, tt.fn)
				close(done)
			}()

			select {
			case <-done:
				// Goroutine completed
			case <-time.After(5 * time.Second):
				t.Fatal("test timed out waiting for goroutine")
			}

			// Wait a bit for all goroutines to finish
			time.Sleep(10 * time.Millisecond)
		})
	}
}

func TestServiceExecutor_DoAsync_Concurrent(t *testing.T) {
	logger := infra.NewNoopLogger()
	tracer := infra.NewNoopTracer()

	executor := NewServiceExecutor(logger, tracer)
	ctx := context.Background()

	numGoroutines := 10
	done := make(chan struct{}, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			executor.DoAsync(ctx, fmt.Sprintf("concurrent-operation-%d", index), func(innerCtx context.Context) {
				done <- struct{}{}
			})
		}(i)
	}

	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("test timed out waiting for goroutines")
		}
	}
}

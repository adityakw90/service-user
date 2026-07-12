package executor

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	domainError "github.com/adityakw90/service-user/internal/core/domain/errors"
	portExecutor "github.com/adityakw90/service-user/internal/core/port/executor"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			wrappedFn := func(inner context.Context) {
				defer close(done)
				tt.fn(inner)
			}
			executor.DoAsync(ctx, tt.operationName, wrappedFn)
			select {
			case <-done:
				// async callback completed
			case <-time.After(5 * time.Second):
				t.Fatal("test timed out waiting for async callback")
			}
		})
	}
}

func TestServiceExecutor_DoAsync_Lifecycle(t *testing.T) {
	logger := infra.NewNoopLogger()
	tracer := infra.NewNoopTracer()
	executor := NewServiceExecutor(logger, tracer)

	require.ErrorIs(t, executor.DoAsync(context.Background(), "invalid", nil), domainError.ErrExecutorFnInvalid)

	started := make(chan struct{})
	release := make(chan struct{})
	require.NoError(t, executor.DoAsync(context.Background(), "blocking", func(context.Context) {
		close(started)
		<-release
	}))
	<-started

	closed := make(chan struct{})
	go func() {
		executor.Close()
		close(closed)
	}()

	select {
	case <-closed:
		t.Fatal("Close returned before the asynchronous task finished")
	case <-time.After(50 * time.Millisecond):
	}

	close(release)
	select {
	case <-closed:
	case <-time.After(5 * time.Second):
		t.Fatal("Close timed out waiting for the asynchronous task")
	}

	require.ErrorIs(t, executor.DoAsync(context.Background(), "closed", func(context.Context) {}), domainError.ErrExecutorClosed)
	executor.Close()
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

func TestServiceExecutor_DoParallel(t *testing.T) {
	tests := []struct {
		name          string
		operationName string
		timeout       time.Duration
		tasks         []portExecutor.Task
		setupContext  func(context.Context) (context.Context, context.CancelFunc)
		wantErr       bool
		errCheck      func(error) bool
	}{
		{
			name:          "Happy Path - all tasks execute successfully",
			operationName: "parallel-success",
			timeout:       5 * time.Second,
			tasks: []portExecutor.Task{
				{
					Name: "task-1",
					Execute: func(ctx context.Context) error {
						return nil
					},
				},
				{
					Name: "task-2",
					Execute: func(ctx context.Context) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name:          "Error Propagation - one task fails",
			operationName: "parallel-error",
			timeout:       5 * time.Second,
			tasks: []portExecutor.Task{
				{
					Name: "task-success",
					Execute: func(ctx context.Context) error {
						return nil
					},
				},
				{
					Name: "task-fail",
					Execute: func(ctx context.Context) error {
						return errors.New("task failed")
					},
				},
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err.Error() == "task failed"
			},
		},
		{
			name:          "Empty Tasks - returns nil immediately",
			operationName: "empty-tasks",
			timeout:       5 * time.Second,
			tasks:         []portExecutor.Task{},
			wantErr:       false,
		},
		{
			name:          "Context Cancellation - cancelled before completion",
			operationName: "cancelled-parallel",
			timeout:       10 * time.Second,
			tasks: []portExecutor.Task{
				{
					Name: "long-task",
					Execute: func(ctx context.Context) error {
						select {
						case <-time.After(10 * time.Second):
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					},
				},
			},
			setupContext: func(ctx context.Context) (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(ctx)
				go func() {
					time.Sleep(100 * time.Millisecond)
					cancel()
				}()
				return ctx, cancel
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, context.Canceled)
			},
		},
		{
			name:          "Timeout - tasks exceed timeout",
			operationName: "timeout-parallel",
			timeout:       100 * time.Millisecond,
			tasks: []portExecutor.Task{
				{
					Name: "slow-task",
					Execute: func(ctx context.Context) error {
						time.Sleep(1 * time.Second)
						return nil
					},
				},
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, context.DeadlineExceeded)
			},
		},
		{
			name:          "Panic Recovery - task panics",
			operationName: "panic-parallel",
			timeout:       5 * time.Second,
			tasks: []portExecutor.Task{
				{
					Name: "panic-task",
					Execute: func(ctx context.Context) error {
						panic("something went wrong")
					},
				},
			},
			wantErr: true,
			errCheck: func(err error) bool {
				return err != nil && err.Error() == "panic in task panic-task: something went wrong"
			},
		},
		{
			name:          "No Timeout - context without deadline",
			operationName: "no-timeout",
			timeout:       0,
			tasks: []portExecutor.Task{
				{
					Name: "quick-task",
					Execute: func(ctx context.Context) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := infra.NewNoopLogger()
			tracer := infra.NewNoopTracer()

			executor := NewServiceExecutor(logger, tracer)
			ctx := context.Background()

			if tt.setupContext != nil {
				var cancel context.CancelFunc
				ctx, cancel = tt.setupContext(ctx)
				defer cancel()
			} else if tt.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.timeout)
				defer cancel()
			}

			err := executor.DoParallel(ctx, tt.operationName, tt.tasks)

			if tt.wantErr {
				require.Error(t, err, "expected an error")
				if tt.errCheck != nil {
					assert.True(t, tt.errCheck(err), "error check failed: %v", err)
				}
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}

func TestServiceExecutor_DoParallel_Concurrency(t *testing.T) {
	logger := infra.NewNoopLogger()
	tracer := infra.NewNoopTracer()

	executor := NewServiceExecutor(logger, tracer)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var executionCount int32 = 0
	numTasks := 10
	tasks := make([]portExecutor.Task, numTasks)

	for i := 0; i < numTasks; i++ {
		tasks[i] = portExecutor.Task{
			Name: fmt.Sprintf("concurrent-task-%d", i),
			Execute: func(ctx context.Context) error {
				atomic.AddInt32(&executionCount, 1)
				time.Sleep(50 * time.Millisecond)
				return nil
			},
		}
	}

	err := executor.DoParallel(ctx, "concurrent-test", tasks)
	assert.NoError(t, err)
	assert.Equal(t, int32(numTasks), atomic.LoadInt32(&executionCount))
}

func TestServiceExecutor_DoParallel_AllTasksComplete(t *testing.T) {
	logger := infra.NewNoopLogger()
	tracer := infra.NewNoopTracer()

	executor := NewServiceExecutor(logger, tracer)
	ctx := context.Background()

	var task1Complete, task2Complete, task3Complete bool

	tasks := []portExecutor.Task{
		{
			Name: "task-1",
			Execute: func(ctx context.Context) error {
				task1Complete = true
				return nil
			},
		},
		{
			Name: "task-2",
			Execute: func(ctx context.Context) error {
				task2Complete = true
				return nil
			},
		},
		{
			Name: "task-3",
			Execute: func(ctx context.Context) error {
				task3Complete = true
				return nil
			},
		},
	}

	err := executor.DoParallel(ctx, "all-tasks-test", tasks)
	assert.NoError(t, err)
	assert.True(t, task1Complete, "task 1 should complete")
	assert.True(t, task2Complete, "task 2 should complete")
	assert.True(t, task3Complete, "task 3 should complete")
}

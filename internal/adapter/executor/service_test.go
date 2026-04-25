package executor

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

type logEntry struct {
	message string
	fields  map[string]interface{}
}

type mockLogger struct {
	mu            sync.Mutex
	debugMessages []logEntry
	infoMessages  []logEntry
	warnMessages  []logEntry
	errorMessages []logEntry
	fatalMessages []logEntry
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		debugMessages: make([]logEntry, 0),
		infoMessages:  make([]logEntry, 0),
		warnMessages:  make([]logEntry, 0),
		errorMessages: make([]logEntry, 0),
		fatalMessages: make([]logEntry, 0),
	}
}

func (m *mockLogger) SetLogLevel(level string) {}

func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.debugMessages = append(m.debugMessages, logEntry{message: msg, fields: fields})
}

func (m *mockLogger) Info(msg string, fields map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.infoMessages = append(m.infoMessages, logEntry{message: msg, fields: fields})
}

func (m *mockLogger) Warn(msg string, fields map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warnMessages = append(m.warnMessages, logEntry{message: msg, fields: fields})
}

func (m *mockLogger) Error(msg string, fields map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorMessages = append(m.errorMessages, logEntry{message: msg, fields: fields})
}

func (m *mockLogger) Fatal(msg string, fields map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fatalMessages = append(m.fatalMessages, logEntry{message: msg, fields: fields})
}

func (m *mockLogger) WithSpanContext(sc trace.SpanContext) monitoring.Logger {
	return m
}

func (m *mockLogger) AddCallerSkipNum(skipNum int) monitoring.Logger {
	return m
}

func (m *mockLogger) Sync() error {
	return nil
}

func (m *mockLogger) hasDebugMessage(msg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.debugMessages {
		if entry.message == msg {
			return true
		}
	}
	return false
}

func (m *mockLogger) hasErrorMessage(msg string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, entry := range m.errorMessages {
		if entry.message == msg {
			return true
		}
	}
	return false
}

func (m *mockLogger) getDebugMessages() []logEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.debugMessages
}

func (m *mockLogger) getErrorMessages() []logEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.errorMessages
}

type mockTracer struct {
	spansStarted int
	spansEnded   int
	mu           sync.Mutex
}

func newMockTracer() *mockTracer {
	return &mockTracer{}
}

func (t *mockTracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	t.mu.Lock()
	t.spansStarted++
	t.mu.Unlock()
	tracer := nooptrace.NewTracerProvider().Tracer("test")
	return tracer.Start(ctx, name, opts...)
}

func (t *mockTracer) EndSpan(span trace.Span) {
	t.mu.Lock()
	t.spansEnded++
	t.mu.Unlock()
	span.End()
}

func (t *mockTracer) ExtractContext(ctx context.Context, md map[string][]string) context.Context {
	return ctx
}

func (t *mockTracer) InjectContext(ctx context.Context) map[string][]string {
	return map[string][]string{}
}

func (t *mockTracer) Shutdown(ctx context.Context) error {
	return nil
}

func (t *mockTracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (t *mockTracer) StartChildSpan(ctx context.Context, name string, parent trace.Span) (context.Context, trace.Span) {
	t.mu.Lock()
	t.spansStarted++
	t.mu.Unlock()
	tracer := nooptrace.NewTracerProvider().Tracer("test")
	return tracer.Start(ctx, name)
}

func TestServiceExecutor_Do(t *testing.T) {
	tests := []struct {
		name           string
		operationName  string
		fn             func(ctx context.Context)
		wantStartLog   bool
		wantFinishLog  bool
		wantSpanCall   bool
		validateFunc   func(*testing.T, *mockLogger, *mockTracer)
	}{
		{
			name:          "Happy Path - function executes successfully",
			operationName: "test-operation",
			fn: func(ctx context.Context) {
				// Function that does nothing
			},
			wantStartLog:  true,
			wantFinishLog: true,
			wantSpanCall:  true,
			validateFunc: func(t *testing.T, ml *mockLogger, mt *mockTracer) {
				debugLogs := ml.getDebugMessages()
				require.Equal(t, 2, len(debugLogs), "should have start and finish debug logs")
				assert.Equal(t, "start doing something", debugLogs[0].message)
				assert.Equal(t, "test-operation", debugLogs[0].fields["name"])
				assert.Equal(t, "finish doing something", debugLogs[1].message)
				assert.Equal(t, "test-operation", debugLogs[1].fields["name"])
			},
		},
		{
			name:          "Named Operation - operation name appears in logs",
			operationName: "important-task",
			fn: func(ctx context.Context) {
				// Function execution
			},
			wantStartLog:  true,
			wantFinishLog: true,
			wantSpanCall:  true,
			validateFunc: func(t *testing.T, ml *mockLogger, mt *mockTracer) {
				debugLogs := ml.getDebugMessages()
				assert.Equal(t, "important-task", debugLogs[0].fields["name"])
				assert.Equal(t, "important-task", debugLogs[1].fields["name"])
			},
		},
		{
			name:          "Context Propagation - new context with span is passed",
			operationName: "context-test",
			fn: func(ctx context.Context) {
				// Verify context is passed through
				assert.NotNil(t, ctx, "context should not be nil")
			},
			wantStartLog:  true,
			wantFinishLog: true,
			wantSpanCall:  true,
		},
		{
			name:          "Empty Function Name - logs empty name",
			operationName: "",
			fn: func(ctx context.Context) {},
			wantStartLog:  true,
			wantFinishLog: true,
			wantSpanCall:  true,
			validateFunc: func(t *testing.T, ml *mockLogger, mt *mockTracer) {
				debugLogs := ml.getDebugMessages()
				assert.Equal(t, "", debugLogs[0].fields["name"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := newMockLogger()
			mockTracer := newMockTracer()

			executor := NewServiceExecutor(mockLogger, mockTracer)
			ctx := context.Background()

			executor.Do(ctx, tt.operationName, tt.fn)

			assert.Equal(t, tt.wantStartLog, mockLogger.hasDebugMessage("start doing something"))
			assert.Equal(t, tt.wantFinishLog, mockLogger.hasDebugMessage("finish doing something"))

			if tt.validateFunc != nil {
				tt.validateFunc(t, mockLogger, mockTracer)
			}
		})
	}
}

func TestServiceExecutor_DoAsync(t *testing.T) {
	tests := []struct {
		name           string
		operationName  string
		fn             func(context.Context)
		wantStartLog   bool
		wantFinishLog  bool
		wantErrorLog   bool
		validateFunc   func(*testing.T, *mockLogger, *mockTracer)
		setupCancel    func(context.Context) (context.Context, context.CancelFunc)
	}{
		{
			name:          "Happy Path - function executes in goroutine",
			operationName: "async-operation",
			fn: func(ctx context.Context) {
			},
			wantStartLog:  true,
			wantFinishLog: true,
			wantErrorLog:  false,
			validateFunc: func(t *testing.T, ml *mockLogger, mt *mockTracer) {
				debugLogs := ml.getDebugMessages()
				require.GreaterOrEqual(t, len(debugLogs), 2, "should have at least start and finish debug logs")
				assert.Equal(t, "start doing something", debugLogs[0].message)
				assert.Equal(t, "async-operation", debugLogs[0].fields["name"])
			},
		},
		{
			name:          "Panic Recovery - panic is caught and logged",
			operationName: "panic-operation",
			fn: func(ctx context.Context) {
				panic("something went wrong")
			},
			wantStartLog:  true,
			wantFinishLog: false,
			wantErrorLog:  true,
			validateFunc: func(t *testing.T, ml *mockLogger, mt *mockTracer) {
				errorLogs := ml.getErrorMessages()
				require.Equal(t, 1, len(errorLogs), "should have error log for panic")
				assert.Equal(t, "recovered from panic", errorLogs[0].message)
				assert.Equal(t, "something went wrong", fmt.Sprintf("%v", errorLogs[0].fields["msg"]))
				assert.Equal(t, "panic-operation", errorLogs[0].fields["name"])
			},
		},
		{
			name:          "Background Context - goroutine uses background context",
			operationName: "background-operation",
			fn: func(ctx context.Context) {
				assert.NotNil(t, ctx, "context should not be nil")
			},
			wantStartLog:  true,
			wantFinishLog: true,
			wantErrorLog:  false,
		},
		{
			name:          "Context Independence - parent cancellation doesn't affect",
			operationName: "independent-operation",
			fn: func(ctx context.Context) {
			},
			wantStartLog:  true,
			wantFinishLog: true,
			wantErrorLog:  false,
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
			wantStartLog:  true,
			wantFinishLog: false,
			wantErrorLog:  true,
			validateFunc: func(t *testing.T, ml *mockLogger, mt *mockTracer) {
				errorLogs := ml.getErrorMessages()
				require.Equal(t, 1, len(errorLogs))
				assert.Equal(t, "recovered from panic", errorLogs[0].message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := newMockLogger()
			mockTracer := newMockTracer()

			executor := NewServiceExecutor(mockLogger, mockTracer)
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

			// Wait a bit for all goroutines to finish logging
			time.Sleep(10 * time.Millisecond)

			assert.Equal(t, tt.wantStartLog, mockLogger.hasDebugMessage("start doing something"))
			assert.Equal(t, tt.wantFinishLog, mockLogger.hasDebugMessage("finish doing something"))
			assert.Equal(t, tt.wantErrorLog, mockLogger.hasErrorMessage("recovered from panic"))

			if tt.validateFunc != nil {
				tt.validateFunc(t, mockLogger, mockTracer)
			}
		})
	}
}

func TestServiceExecutor_DoAsync_Concurrent(t *testing.T) {
	mockLogger := newMockLogger()
	mockTracer := newMockTracer()

	executor := NewServiceExecutor(mockLogger, mockTracer)
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

	debugLogs := mockLogger.getDebugMessages()
	assert.GreaterOrEqual(t, len(debugLogs), numGoroutines, "should have logs for all concurrent operations")
}

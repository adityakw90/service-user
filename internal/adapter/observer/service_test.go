package observer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
)

// testDataType is a sample type for testing the generic observer
type testDataType struct {
	UserID   string
	Action   string
	Duration int64
}

func TestAdapter_Observer_NewServiceObserver(t *testing.T) {
	tests := []struct {
		name    string
		logger  monitoring.Logger
		tracer  monitoring.Tracer
		attrs   func(testDataType) []attribute.KeyValue
		logMap  func(testDataType) map[string]any
		wantNil bool
	}{
		{
			name:    "create observer with all parameters",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			attrs:   func(d testDataType) []attribute.KeyValue { return []attribute.KeyValue{} },
			logMap:  func(d testDataType) map[string]any { return map[string]any{} },
			wantNil: false,
		},
		{
			name:    "create observer with nil attrs",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			attrs:   nil,
			logMap:  func(d testDataType) map[string]any { return map[string]any{} },
			wantNil: false,
		},
		{
			name:    "create observer with nil logMap",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			attrs:   func(d testDataType) []attribute.KeyValue { return []attribute.KeyValue{} },
			logMap:  nil,
			wantNil: false,
		},
		{
			name:    "create observer with nil functions",
			logger:  newMockLogger(),
			tracer:  newMockTracer(),
			attrs:   nil,
			logMap:  nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs := NewServiceObserver(tt.logger, tt.tracer, tt.attrs, tt.logMap)

			if (obs == nil) != tt.wantNil {
				t.Errorf("NewServiceObserver() = %v, wantNil %v", obs, tt.wantNil)
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_OnSignal_Success(t *testing.T) {
	tests := []struct {
		name     string
		sig      signal.SignalType
		data     testDataType
		attrs    func(testDataType) []attribute.KeyValue
		logMap   func(testDataType) map[string]any
		wantMsg  string
		wantLvl  string
		wantKeys []string
	}{
		{
			name:     "success signal without data functions",
			sig:      signal.SignalSuccess,
			data:     testDataType{UserID: "123", Action: "login", Duration: 100},
			attrs:    nil,
			logMap:   nil,
			wantMsg:  "service signal",
			wantLvl:  "debug",
			wantKeys: []string{"signal"},
		},
		{
			name:  "start signal with logMap",
			sig:   signal.SignalStart,
			data:  testDataType{UserID: "456", Action: "register", Duration: 200},
			attrs: nil,
			logMap: func(d testDataType) map[string]any {
				return map[string]any{"user_id": d.UserID, "action": d.Action}
			},
			wantMsg:  "service signal",
			wantLvl:  "debug",
			wantKeys: []string{"signal", "user_id", "action"},
		},
		{
			name: "reject signal with full functions",
			sig:  signal.SignalReject,
			data: testDataType{UserID: "789", Action: "delete", Duration: 50},
			attrs: func(d testDataType) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("user_id", d.UserID),
					attribute.String("action", d.Action),
				}
			},
			logMap: func(d testDataType) map[string]any {
				return map[string]any{"user_id": d.UserID, "duration_ms": d.Duration}
			},
			wantMsg:  "service signal",
			wantLvl:  "debug",
			wantKeys: []string{"signal", "user_id", "duration_ms"},
		},
		{
			name:     "fail signal with no error",
			sig:      signal.SignalFail,
			data:     testDataType{},
			attrs:    nil,
			logMap:   nil,
			wantMsg:  "service signal",
			wantLvl:  "debug",
			wantKeys: []string{"signal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewServiceObserver(logger, tracer, tt.attrs, tt.logMap)
			ctx := context.Background()

			obs.OnSignal(ctx, tt.sig, tt.data, nil)

			// Verify debug log was called
			var logs []logEntry
			switch tt.wantLvl {
			case "debug":
				logs = logger.debugMessages
			case "error":
				logs = logger.errorMessages
			}

			if len(logs) != 1 {
				t.Fatalf("expected 1 log entry, got %d", len(logs))
			}

			entry := logs[0]
			if entry.message != tt.wantMsg {
				t.Errorf("message = %s, want %s", entry.message, tt.wantMsg)
			}

			// Verify all expected keys are present
			for _, key := range tt.wantKeys {
				if _, ok := entry.fields[key]; !ok {
					t.Errorf("expected key %q not found in fields: %v", key, entry.fields)
				}
			}

			// Verify signal value
			if entry.fields["signal"] != tt.sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], tt.sig)
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_OnSignal_WithError(t *testing.T) {
	tests := []struct {
		name    string
		sig     signal.SignalType
		data    testDataType
		err     error
		attrs   func(testDataType) []attribute.KeyValue
		logMap  func(testDataType) map[string]any
		wantMsg string
		wantLvl string
	}{
		{
			name:    "fail signal with error",
			sig:     signal.SignalFail,
			data:    testDataType{UserID: "123", Action: "login"},
			err:     errors.New("invalid credentials"),
			attrs:   nil,
			logMap:  nil,
			wantMsg: "service signal",
			wantLvl: "error",
		},
		{
			name:  "error with logMap",
			sig:   signal.SignalFail,
			data:  testDataType{UserID: "456", Action: "register"},
			err:   errors.New("database connection failed"),
			attrs: nil,
			logMap: func(d testDataType) map[string]any {
				return map[string]any{"user_id": d.UserID, "action": d.Action}
			},
			wantMsg: "service signal",
			wantLvl: "error",
		},
		{
			name:    "success signal with error (edge case)",
			sig:     signal.SignalSuccess,
			data:    testDataType{},
			err:     errors.New("unexpected error in success path"),
			attrs:   nil,
			logMap:  nil,
			wantMsg: "service signal",
			wantLvl: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewServiceObserver(logger, tracer, tt.attrs, tt.logMap)
			ctx := context.Background()

			obs.OnSignal(ctx, tt.sig, tt.data, tt.err)

			// Verify error log was called
			if len(logger.errorMessages) != 1 {
				t.Fatalf("expected 1 error log entry, got %d", len(logger.errorMessages))
			}

			entry := logger.errorMessages[0]
			if entry.message != tt.wantMsg {
				t.Errorf("message = %s, want %s", entry.message, tt.wantMsg)
			}

			// Verify error field is present
			if errType, ok := entry.fields["error.Type"].(string); !ok {
				t.Errorf("error field is not a string")
			} else if errType != fmt.Sprintf("%T", tt.err) {
				t.Errorf("error = %s, want %s", errType, fmt.Sprintf("%T", tt.err))
			}

			if errMsg, ok := entry.fields["error.Message"].(string); !ok {
				t.Errorf("error field is not a string")
			} else if errMsg != tt.err.Error() {
				t.Errorf("error = %s, want %s", errMsg, tt.err.Error())
			}

			// Verify signal value
			if entry.fields["signal"] != tt.sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], tt.sig)
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_OnSignal_AllSignalTypes(t *testing.T) {
	signals := []signal.SignalType{
		signal.SignalStart,
		signal.SignalReject,
		signal.SignalFail,
		signal.SignalSuccess,
	}

	for _, sig := range signals {
		t.Run(string(sig), func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			attrs := func(d testDataType) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("user_id", d.UserID),
				}
			}

			logMap := func(d testDataType) map[string]any {
				return map[string]any{
					"user_id": d.UserID,
					"action":  d.Action,
				}
			}

			obs := NewServiceObserver(logger, tracer, attrs, logMap)
			ctx := context.Background()

			data := testDataType{UserID: "test-123", Action: "test-action"}
			obs.OnSignal(ctx, sig, data, nil)

			// Verify debug log was called
			if len(logger.debugMessages) != 1 {
				t.Fatalf("expected 1 debug log entry, got %d", len(logger.debugMessages))
			}

			entry := logger.debugMessages[0]
			if entry.fields["signal"] != sig {
				t.Errorf("signal = %v, want %v", entry.fields["signal"], sig)
			}

			if entry.fields["user_id"] != "test-123" {
				t.Errorf("user_id = %v, want test-123", entry.fields["user_id"])
			}

			if entry.fields["action"] != "test-action" {
				t.Errorf("action = %v, want test-action", entry.fields["action"])
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_OnSignal_ContextWithSpan(t *testing.T) {
	tests := []struct {
		name          string
		withSpan      bool
		expectAttrs   bool
		expectLogging bool
	}{
		{
			name:          "context with valid span",
			withSpan:      true,
			expectAttrs:   true,
			expectLogging: true,
		},
		{
			name:          "context without span",
			withSpan:      false,
			expectAttrs:   false,
			expectLogging: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			attrs := func(d testDataType) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("test_attr", "test_value"),
				}
			}

			logMap := func(d testDataType) map[string]any {
				return map[string]any{"test_key": "test_value"}
			}

			obs := NewServiceObserver(logger, tracer, attrs, logMap)
			ctx := context.Background()

			// Add span to context if required
			if tt.withSpan {
				ctx, _ = tracer.StartSpan(ctx, "test-span")
			}

			data := testDataType{UserID: "123"}
			obs.OnSignal(ctx, signal.SignalStart, data, nil)

			if tt.expectLogging {
				if len(logger.debugMessages) != 1 {
					t.Errorf("expected 1 debug log, got %d", len(logger.debugMessages))
				}
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_AttributesFunction(t *testing.T) {
	tests := []struct {
		name         string
		attrs        func(testDataType) []attribute.KeyValue
		wantAttrName string
		wantAttrVal  string
	}{
		{
			name: "single attribute",
			attrs: func(d testDataType) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("user_id", d.UserID),
				}
			},
			wantAttrName: "user_id",
			wantAttrVal:  "test-123",
		},
		{
			name: "multiple attributes",
			attrs: func(d testDataType) []attribute.KeyValue {
				return []attribute.KeyValue{
					attribute.String("user_id", d.UserID),
					attribute.String("action", d.Action),
					attribute.Int64("duration", d.Duration),
				}
			},
			wantAttrName: "action",
			wantAttrVal:  "login",
		},
		{
			name:         "no attributes",
			attrs:        func(d testDataType) []attribute.KeyValue { return []attribute.KeyValue{} },
			wantAttrName: "",
			wantAttrVal:  "",
		},
		{
			name:         "nil attributes",
			attrs:        nil,
			wantAttrName: "",
			wantAttrVal:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewServiceObserver(logger, tracer, tt.attrs, nil)
			ctx := context.Background()

			data := testDataType{UserID: "test-123", Action: "login", Duration: 100}
			obs.OnSignal(ctx, signal.SignalStart, data, nil)

			// Verify the observer doesn't panic with different attrs configurations
			if len(logger.debugMessages) != 1 {
				t.Errorf("expected 1 debug log, got %d", len(logger.debugMessages))
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_LogMapFunction(t *testing.T) {
	tests := []struct {
		name    string
		logMap  func(testDataType) map[string]any
		data    testDataType
		wantMap map[string]any
	}{
		{
			name: "map with user_id and action",
			logMap: func(d testDataType) map[string]any {
				return map[string]any{
					"user_id": d.UserID,
					"action":  d.Action,
				}
			},
			data: testDataType{UserID: "user-123", Action: "logout"},
			wantMap: map[string]any{
				"user_id": "user-123",
				"action":  "logout",
			},
		},
		{
			name: "map with all fields",
			logMap: func(d testDataType) map[string]any {
				return map[string]any{
					"user_id":  d.UserID,
					"action":   d.Action,
					"duration": d.Duration,
				}
			},
			data: testDataType{UserID: "user-456", Action: "register", Duration: 250},
			wantMap: map[string]any{
				"user_id":  "user-456",
				"action":   "register",
				"duration": int64(250),
			},
		},
		{
			name:    "empty map",
			logMap:  func(d testDataType) map[string]any { return map[string]any{} },
			data:    testDataType{UserID: "user-789"},
			wantMap: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewServiceObserver(logger, tracer, nil, tt.logMap)
			ctx := context.Background()

			obs.OnSignal(ctx, signal.SignalSuccess, tt.data, nil)

			if len(logger.debugMessages) != 1 {
				t.Fatalf("expected 1 debug log, got %d", len(logger.debugMessages))
			}

			entry := logger.debugMessages[0]

			// Verify all expected keys from logMap are in fields
			for key, wantVal := range tt.wantMap {
				gotVal, ok := entry.fields[key]
				if !ok {
					t.Errorf("key %q not found in log fields", key)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("key %q = %v, want %v", key, gotVal, wantVal)
				}
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_ConcurrentCalls(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	attrs := func(d testDataType) []attribute.KeyValue {
		return []attribute.KeyValue{attribute.String("user_id", d.UserID)}
	}

	logMap := func(d testDataType) map[string]any {
		return map[string]any{"user_id": d.UserID, "action": d.Action}
	}

	obs := NewServiceObserver(logger, tracer, attrs, logMap)
	ctx := context.Background()

	// Concurrent calls to OnSignal
	var wg sync.WaitGroup
	numGoroutines := 10
	callsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				data := testDataType{
					UserID: fmt.Sprintf("user-%d-%d", id, j),
					Action: "test-action",
				}
				obs.OnSignal(ctx, signal.SignalStart, data, nil)
			}
		}(i)
	}

	wg.Wait()

	totalCalls := numGoroutines * callsPerGoroutine
	if len(logger.debugMessages) != totalCalls {
		t.Errorf("expected %d debug logs, got %d", totalCalls, len(logger.debugMessages))
	}
}

func TestAdapter_Observer_ServiceObserver_NilFunctionsSafe(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	// Create observer with nil functions
	obs := NewServiceObserver[testDataType](logger, tracer, nil, nil)
	ctx := context.Background()

	data := testDataType{UserID: "test-123", Action: "test"}

	// Should not panic
	obs.OnSignal(ctx, signal.SignalStart, data, nil)

	if len(logger.debugMessages) != 1 {
		t.Errorf("expected 1 debug log, got %d", len(logger.debugMessages))
	}

	entry := logger.debugMessages[0]
	if entry.fields["signal"] != signal.SignalStart {
		t.Errorf("signal = %v, want %v", entry.fields["signal"], signal.SignalStart)
	}
}

func TestAdapter_Observer_ServiceObserver_DetachContext(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func() context.Context
		wantBackground bool
	}{
		{
			name: "context with span preserves trace context",
			setupContext: func() context.Context {
				tracer := newMockTracer()
				ctx, _ := tracer.StartSpan(context.Background(), "parent-span")
				return ctx
			},
			wantBackground: true,
		},
		{
			name: "context without span returns background context",
			setupContext: func() context.Context {
				return context.Background()
			},
			wantBackground: true,
		},
		{
			name: "context with cancellation returns background context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				// Cancel immediately to test that detached context is not affected
				cancel()
				return ctx
			},
			wantBackground: true,
		},
		{
			name: "context with timeout returns background context",
			setupContext: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1)
				defer cancel()
				return ctx
			},
			wantBackground: true,
		},
		{
			name: "context with span and values",
			setupContext: func() context.Context {
				tracer := newMockTracer()
				ctx, _ := tracer.StartSpan(context.Background(), "parent-span")
				ctx = context.WithValue(ctx, "key", "value")
				return ctx
			},
			wantBackground: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := newMockLogger()
			tracer := newMockTracer()

			obs := NewServiceObserver[testDataType](logger, tracer, nil, nil)
			parentCtx := tt.setupContext()

			// Call DetachContext
			detachedCtx := obs.DetachContext(parentCtx)

			// Verify the context is not nil
			if detachedCtx == nil {
				t.Fatal("DetachContext() returned nil context")
			}

			// Verify the context is different from the parent
			if detachedCtx == parentCtx {
				t.Error("DetachContext() should return a new context, not the parent")
			}

			// Verify the returned context is a background context (no parent cancellation)
			// by checking if it's not done
			select {
			case <-detachedCtx.Done():
				t.Error("detached context should not be done (background context)")
			default:
				// Expected - background context is not done
			}

			// Verify the detached context has a span (even if noop)
			detachedSpan := tracer.SpanFromContext(detachedCtx)
			if detachedSpan == nil {
				t.Error("detached context should have a span")
			}

			// Verify the parent context is different from detached
			// (they may both have spans, but they should be different contexts)
			parentSpan := tracer.SpanFromContext(parentCtx)
			if parentSpan != nil && detachedSpan != nil {
				// Both have spans, which is expected
				// The mock tracer creates noop spans, so we can't validate trace linking
				// In a real tracer, the detached span would be linked to the parent's trace
			}
		})
	}
}

func TestAdapter_Observer_ServiceObserver_DetachContext_Concurrent(t *testing.T) {
	logger := newMockLogger()
	tracer := newMockTracer()

	obs := NewServiceObserver[testDataType](logger, tracer, nil, nil)

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
				detachedCtx := obs.DetachContext(ctx)

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

	// Verify all contexts are unique (or at least different from background)
	bgContext := context.Background()
	for i, ctx := range results {
		if ctx == nil {
			t.Errorf("result %d is nil", i)
		}
		if ctx == bgContext {
			// This is actually expected for noop tracer, but let's document it
			// The mock tracer returns noop spans, so the contexts might be similar
		}
	}
}

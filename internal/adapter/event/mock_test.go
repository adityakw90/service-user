package event

import (
	"context"
	"sync"

	"github.com/adityakw90/go-monitoring"
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

type mockTracer struct{}

func newMockTracer() monitoring.Tracer {
	return &mockTracer{}
}

func (t *mockTracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		return ctx, span
	}
	return nooptrace.NewTracerProvider().Tracer("test").Start(ctx, name, opts...)
}

func (t *mockTracer) EndSpan(span trace.Span) {}

func (t *mockTracer) ExtractContext(ctx context.Context, md map[string][]string) context.Context {
	return ctx
}

func (t *mockTracer) InjectContext(ctx context.Context) map[string][]string {
	return map[string][]string{
		"traceparent": []string{"00-1234567890abcdef-1234567890abcdef-01"},
	}
}

func (t *mockTracer) Shutdown(ctx context.Context) error {
	return nil
}

func (t *mockTracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (t *mockTracer) StartChildSpan(ctx context.Context, name string, parent trace.Span) (context.Context, trace.Span) {
	return nooptrace.NewTracerProvider().Tracer("test").Start(ctx, name)
}

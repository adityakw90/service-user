package resolver

import (
	"context"

	"github.com/adityakw90/go-monitoring"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// noOpTracer is a minimal tracer implementation for testing that satisfies the go-monitoring.Tracer interface
// It uses real OpenTelemetry noop implementations
type noOpTracer struct {
	tracer trace.Tracer
}

func newNoOpTracer() monitoring.Tracer {
	return &noOpTracer{
		tracer: trace.NewNoopTracerProvider().Tracer("test"),
	}
}

func (t *noOpTracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

func (t *noOpTracer) EndSpan(span trace.Span) {
	span.End()
}

func (t *noOpTracer) ExtractContext(ctx context.Context, md metadata.MD) context.Context {
	return ctx
}

func (t *noOpTracer) InjectContext(ctx context.Context) metadata.MD {
	return metadata.MD{}
}

func (t *noOpTracer) Shutdown(ctx context.Context) error {
	return nil
}

func (t *noOpTracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (t *noOpTracer) StartChildSpan(ctx context.Context, name string, parent trace.Span) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, trace.WithLinks(trace.Link{}))
}

// mockLogger is a mock implementation of portMonitoring.Logger for testing
type mockLogger struct{}

func (m *mockLogger) SetLogLevel(level string)                            {}
func (m *mockLogger) Debug(message string, fields map[string]interface{}) {}
func (m *mockLogger) Info(message string, fields map[string]interface{})  {}
func (m *mockLogger) Warn(message string, fields map[string]interface{})  {}
func (m *mockLogger) Error(message string, fields map[string]interface{}) {}
func (m *mockLogger) Fatal(message string, fields map[string]interface{}) {}
func (m *mockLogger) WithSpanContext(span trace.SpanContext) monitoring.Logger {
	return m
}
func (m *mockLogger) AddCallerSkipNum(skipNum int) monitoring.Logger {
	return m
}
func (m *mockLogger) Sync() error {
	return nil
}

// newMockRedis creates a new miniredis server and returns the redis client and cleanup function
func newMockRedis() (*redis.Client, func(), error) {
	s, err := miniredis.Run()
	if err != nil {
		return nil, nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cleanup := func() {
		client.Close()
		s.Close()
	}

	return client, cleanup, nil
}

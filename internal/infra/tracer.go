package infra

import (
	"context"

	gomon "github.com/adityakw90/go-monitoring"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/metadata"
)

// NoopTracer is a no-op tracer implementation.
type NoopTracer struct {
	tracer trace.Tracer
}

func NewNoopTracer() gomon.Tracer {
	return &NoopTracer{
		tracer: trace.NewNoopTracerProvider().Tracer("noop"),
	}
}

func (t *NoopTracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

func (t *NoopTracer) EndSpan(span trace.Span) {
	span.End()
}

func (t *NoopTracer) Shutdown(ctx context.Context) error {
	return nil
}

func (t *NoopTracer) StartChildSpan(ctx context.Context, name string, parent trace.Span) (context.Context, trace.Span) {
	return t.tracer.Start(trace.ContextWithSpan(ctx, parent), name, trace.WithLinks(trace.Link{
		SpanContext: parent.SpanContext(),
	}))
}

func (t *NoopTracer) SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func (t *NoopTracer) ExtractContext(ctx context.Context, md map[string][]string) context.Context {
	// no-op for noop tracer
	return ctx
}

func (t *NoopTracer) InjectContext(ctx context.Context) map[string][]string {
	return metadata.MD{}
}

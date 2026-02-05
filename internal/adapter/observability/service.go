package observability

import (
	"context"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ServiceObserver[T any] struct {
	logger monitoring.Logger
	tracer monitoring.Tracer
	attrs  func(T) []attribute.KeyValue
	logMap func(T) map[string]any
}

func (o *ServiceObserver[T]) OnSignal(
	ctx context.Context,
	signal signal.ServiceSignal,
	data T,
	err error,
) {
	span := trace.SpanFromContext(ctx)

	if span != nil {
		if o.attrs != nil {
			span.SetAttributes(o.attrs(data)...)
		}

		span.AddEvent(string(signal))
	}

	// logger auto attach span context
	l := o.logger.WithSpanContext(span.SpanContext())

	fields := map[string]any{
		"signal": signal,
	}

	if o.logMap != nil {
		for k, v := range o.logMap(data) {
			fields[k] = v
		}
	}

	if err != nil {
		fields["error"] = err.Error()
		l.Error("service signal", fields)
		return
	}

	l.Debug("service signal", fields)
}

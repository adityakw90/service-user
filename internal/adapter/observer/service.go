package observer

import (
	"context"
	"fmt"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type serviceObserver[T any] struct {
	logger monitoring.Logger
	tracer monitoring.Tracer
	attrs  func(T) []attribute.KeyValue
	logMap func(T) map[string]any
}

func NewServiceObserver[T any](
	logger monitoring.Logger,
	tracer monitoring.Tracer,
	attrs func(T) []attribute.KeyValue,
	logMap func(T) map[string]any,
) *serviceObserver[T] {
	return &serviceObserver[T]{
		logger: logger.AddCallerSkipNum(1),
		tracer: tracer,
		attrs:  attrs,
		logMap: logMap,
	}
}

func (o *serviceObserver[T]) OnSignal(
	ctx context.Context,
	signal signal.SignalType,
	data T,
	err error,
) {
	span := trace.SpanFromContext(ctx)

	if span != nil {
		attrs := []attribute.KeyValue{}
		if o.attrs != nil {
			attrs = o.attrs(data)
		}
		if err != nil {
			attrs = append(
				attrs,
				attribute.String("error.Type", fmt.Sprintf("%T", err)),
				attribute.String("error.Message", err.Error()),
			)
		}
		if len(attrs) > 0 {
			span.AddEvent(string(signal), trace.WithAttributes(attrs...))
		} else {
			span.AddEvent(string(signal))
		}
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
		fields["error.Type"] = fmt.Sprintf("%T", err)
		fields["error.Message"] = err.Error()
		l.Error("service signal", fields)
		return
	}

	l.Debug("service signal", fields)
}

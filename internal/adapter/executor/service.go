package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/go-monitoring"
)

type serviceExecutor struct {
	logger monitoring.Logger
	tracer monitoring.Tracer
}

func NewServiceExecutor(logger monitoring.Logger, tracer monitoring.Tracer) *serviceExecutor {
	if logger == nil {
		panic("executor logger must not be nil")
	}
	if tracer == nil {
		panic("executor tracer must not be nil")
	}
	return &serviceExecutor{
		logger: logger,
		tracer: tracer,
	}
}

func (s *serviceExecutor) Do(ctx context.Context, name string, fn func(ctx context.Context)) {
	newCtx, span := s.tracer.StartChildSpan(ctx, name, s.tracer.SpanFromContext(ctx))
	defer span.End()
	fn(newCtx)
}

func (s *serviceExecutor) DoAsync(ctx context.Context, name string, fn func(ctx context.Context)) {
	detached, cancel := context.WithTimeout(
		context.WithoutCancel(ctx),
		30*time.Second,
	)
	newCtx, span := s.tracer.StartChildSpan(
		detached,
		name,
		s.tracer.SpanFromContext(ctx),
	)
	logger := s.logger.WithSpanContext(span.SpanContext())

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("recovered from panic", map[string]any{
					"name": name,
					"msg":  r,
				})
				span.RecordError(fmt.Errorf("panic: %v", r))
			}
			span.End()
			cancel()
		}()
		fn(newCtx)
	}()
}

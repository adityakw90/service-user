package executor

import (
	"context"
	"fmt"

	"github.com/adityakw90/go-monitoring"
)

type serviceExecutor struct {
	logger monitoring.Logger
	tracer monitoring.Tracer
}

func NewServiceExecutor(logger monitoring.Logger, tracer monitoring.Tracer) *serviceExecutor {
	return &serviceExecutor{
		logger: logger,
		tracer: tracer,
	}
}

func (s *serviceExecutor) Do(ctx context.Context, name string, fn func(ctx context.Context)) {
	newCtx, span := s.tracer.StartChildSpan(ctx, name, s.tracer.SpanFromContext(ctx))
	defer span.End()
	logger := s.logger.WithSpanContext(span.SpanContext())

	logger.Debug("start doing something", map[string]any{
		"name": name,
	})

	fn(newCtx)

	logger.Debug("finish doing something", map[string]any{
		"name": name,
	})
}

func (s *serviceExecutor) DoAsync(ctx context.Context, name string, fn func(ctx context.Context)) {
	newCtx, span := s.tracer.StartChildSpan(
		context.Background(),
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
				span.RecordError(fmt.Errorf("%v", r))
			}
			span.End()
		}()
		logger.Debug("start doing something", map[string]any{
			"name": name,
		})
		fn(newCtx)
		logger.Debug("finish doing something", map[string]any{
			"name": name,
		})
	}()
}

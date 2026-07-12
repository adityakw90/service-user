package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/adityakw90/go-monitoring"

	domainError "github.com/adityakw90/service-user/internal/core/domain/errors"
	portExecutor "github.com/adityakw90/service-user/internal/core/port/executor"
)

type serviceExecutor struct {
	logger    monitoring.Logger
	tracer    monitoring.Tracer
	mu        sync.Mutex
	wg        sync.WaitGroup
	closed    bool
	closeOnce sync.Once
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

func (s *serviceExecutor) DoAsync(ctx context.Context, name string, fn func(ctx context.Context)) error {
	if fn == nil {
		return domainError.ErrExecutorFnInvalid
	}
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		return domainError.ErrExecutorClosed
	}

	s.wg.Add(1)
	s.mu.Unlock()

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
			s.wg.Done()
		}()
		fn(newCtx)
	}()

	return nil
}

func (s *serviceExecutor) Close() {
	s.closeOnce.Do(func() {
		s.mu.Lock()
		s.closed = true
		s.mu.Unlock()

		s.wg.Wait()
	})
}

func (s *serviceExecutor) DoParallel(ctx context.Context, name string, tasks []portExecutor.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	_, span := s.tracer.StartChildSpan(ctx, name, s.tracer.SpanFromContext(ctx))
	defer span.End()

	var wg sync.WaitGroup
	errChan := make(chan error, len(tasks))
	done := make(chan struct{})

	wg.Add(len(tasks))
	for _, task := range tasks {
		t := task
		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.WithSpanContext(span.SpanContext()).Error("recovered from panic", map[string]any{
						"task": t.Name,
						"msg":  r,
					})
					span.RecordError(fmt.Errorf("panic in task %s: %v", t.Name, r))
					errChan <- fmt.Errorf("panic in task %s: %v", t.Name, r)
				}
				wg.Done()
			}()

			taskCtx, taskSpan := s.tracer.StartChildSpan(ctx, t.Name, span)
			defer taskSpan.End()

			select {
			case <-taskCtx.Done():
				errChan <- taskCtx.Err()
				return
			default:
				if err := t.Execute(taskCtx); err != nil {
					taskSpan.RecordError(err)
					errChan <- err
					return
				}
				errChan <- nil
			}
		}()
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		close(errChan)
		for err := range errChan {
			if err != nil {
				return err
			}
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

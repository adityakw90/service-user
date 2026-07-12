package executor

import "context"

type Task struct {
	Name    string
	Execute func(ctx context.Context) error
}

type Executor interface {
	Do(ctx context.Context, name string, fn func(ctx context.Context))
	DoAsync(ctx context.Context, name string, fn func(ctx context.Context)) error
	DoParallel(ctx context.Context, name string, tasks []Task) error
}

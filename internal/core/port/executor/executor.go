package executor

import "context"

type Executor interface {
	Do(ctx context.Context, name string, fn func(ctx context.Context))
	DoAsync(ctx context.Context, name string, fn func(ctx context.Context))
}

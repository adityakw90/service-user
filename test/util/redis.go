package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/adityakw90/service-user/internal/config"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/redis/go-redis/v9"
)

func NewTestRedisConnection(t *testing.T, ctx context.Context, cfg *config.Config) (*redis.Client, error) {
	t.Helper()

	lockFile, err := AcquireLock("redis")
	if err != nil {
		return nil, fmt.Errorf("failed to lock redis: %w", err)
	}
	t.Cleanup(func() {
		ReleaseLock(lockFile)
	})

	client, err := infra.NewRedisConnection(ctx, &infra.RedisConfig{
		Host:              cfg.Redis.Host,
		Port:              cfg.Redis.Port,
		User:              cfg.Redis.User,
		Password:          cfg.Redis.Password,
		DB:                cfg.Redis.DB,
		PoolSize:          cfg.Redis.PoolSize,
		PoolTimeout:       cfg.Redis.PoolTimeout,
		ConnectionIdleMin: cfg.Redis.ConnectionIdleMin,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	if err := flushTestRedisDB(t, ctx, client); err != nil {
		return nil, fmt.Errorf("failed to flush redis: %w", err)
	}

	return client, nil
}

func flushTestRedisDB(t *testing.T, ctx context.Context, client *redis.Client) error {
	t.Helper()

	if err := client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush redis: %w", err)
	}

	return nil
}

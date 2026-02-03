package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	DB                int
	PoolSize          int
	PoolTimeout       time.Duration
	ConnectionIdleMin int
}

func NewRedisConnection(ctx context.Context, cfg *RedisConfig) (*redis.Client, error) {
	// Initialize the Redis client with the configuration
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username:     cfg.User,              // Set the username (for Redis ACLs, optional)
		Password:     cfg.Password,          // Set the password (for Redis ACLs or standard auth)
		DB:           cfg.DB,                // Use specified DB (default is 0)
		PoolSize:     cfg.PoolSize,          // Adjust based on expected concurrency
		PoolTimeout:  cfg.PoolTimeout,       // Set the pool timeout
		MinIdleConns: cfg.ConnectionIdleMin, // Maintain a minimum number of idle connections
	})

	// Ping the Redis server to ensure the connection is established
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

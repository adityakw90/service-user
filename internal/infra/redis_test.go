package infra

import (
	"context"
	"os"
	"testing"
	"time"
)

/*
this test using real redis server
*/

func getTestRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:              getEnv("REDIS_HOST", "localhost"),
		Port:              6379,
		User:              "",
		Password:          "",
		DB:                0,
		PoolSize:          10,
		PoolTimeout:       5 * time.Second,
		ConnectionIdleMin: 2,
	}
}

func TestRedisConnection_Connect(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name        string
		config      *RedisConfig
		wantPingErr bool
	}{
		{
			name:        "successful connection with default config",
			config:      getTestRedisConfig(),
			wantPingErr: false,
		},
		{
			name: "failed connection with invalid host",
			config: &RedisConfig{
				Host:              "invalid-host-that-does-not-exist",
				Port:              6379,
				PoolSize:          1,
				PoolTimeout:       1 * time.Second,
				ConnectionIdleMin: 0,
			},
			wantPingErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, err := NewRedisConnection(ctx, tt.config)
			if err != nil {
				if !tt.wantPingErr {
					t.Errorf("NewRedisConnection() error = %v", err)
				}
				return
			}
			defer client.Close()

			if tt.wantPingErr {
				t.Error("NewRedisConnection() expected error, got nil")
				return
			}

			if err := client.Ping(ctx).Err(); err != nil {
				t.Errorf("Ping() error = %v", err)
			}
		})
	}
}

func TestRedisConnection_ConnectionWithAuth(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name        string
		config      *RedisConfig
		wantPingErr bool
	}{
		{
			name: "connection with password",
			config: &RedisConfig{
				Host:              getEnv("REDIS_HOST", "localhost"),
				Port:              6379,
				User:              getEnv("REDIS_USER", ""),
				Password:          os.Getenv("REDIS_PASSWORD"),
				DB:                0,
				PoolSize:          10,
				PoolTimeout:       5 * time.Second,
				ConnectionIdleMin: 2,
			},
			wantPingErr: false,
		},
		{
			name: "connection without password when password required",
			config: &RedisConfig{
				Host:              getEnv("REDIS_HOST", "localhost"),
				Port:              6379,
				User:              "",
				Password:          "",
				DB:                0,
				PoolSize:          10,
				PoolTimeout:       5 * time.Second,
				ConnectionIdleMin: 2,
			},
			wantPingErr: true,
		},
	}

	// Skip if no password configured for auth tests
	if os.Getenv("REDIS_PASSWORD") == "" {
		t.Skip("No Redis password configured")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, err := NewRedisConnection(ctx, tt.config)
			if err != nil {
				if !tt.wantPingErr {
					t.Errorf("NewRedisConnection() error = %v", err)
				}
				return
			}
			defer client.Close()

			if tt.wantPingErr {
				t.Error("NewRedisConnection() expected error, got nil")
				return
			}

			if err := client.Ping(ctx).Err(); err != nil {
				t.Errorf("Ping() error = %v", err)
			}
		})
	}
}

func TestRedisConnection_PoolConfiguration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name      string
		config    *RedisConfig
		testKey   string
		testValue string
		wantValue bool
	}{
		{
			name: "set and get string value",
			config: &RedisConfig{
				Host:              getEnv("REDIS_HOST", "localhost"),
				Port:              6379,
				PoolSize:          5,
				PoolTimeout:       10 * time.Second,
				ConnectionIdleMin: 1,
			},
			testKey:   "test:pool:config:string",
			testValue: "hello-world",
			wantValue: true,
		},
		{
			name: "set and get numeric value",
			config: &RedisConfig{
				Host:              getEnv("REDIS_HOST", "localhost"),
				Port:              6379,
				PoolSize:          5,
				PoolTimeout:       10 * time.Second,
				ConnectionIdleMin: 1,
			},
			testKey:   "test:pool:config:number",
			testValue: "12345",
			wantValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, err := NewRedisConnection(ctx, tt.config)
			if err != nil {
				t.Fatalf("NewRedisConnection() error = %v", err)
			}
			defer client.Close()

			// Set value
			if err := client.Set(ctx, tt.testKey, tt.testValue, time.Minute).Err(); err != nil {
				t.Errorf("Set() error = %v", err)
			}

			// Get value
			got, err := client.Get(ctx, tt.testKey).Result()
			if err != nil {
				t.Errorf("Get() error = %v", err)
			}

			if tt.wantValue && got != tt.testValue {
				t.Errorf("Value mismatch: got %q, want %q", got, tt.testValue)
			}

			// Cleanup
			client.Del(ctx, tt.testKey)
		})
	}
}

func TestRedisConnection_ConcurrentOperations(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name   string
		config *RedisConfig
		numOps int
	}{
		{
			name: "5 concurrent operations",
			config: &RedisConfig{
				Host:              getEnv("REDIS_HOST", "localhost"),
				Port:              6379,
				PoolSize:          10,
				PoolTimeout:       5 * time.Second,
				ConnectionIdleMin: 2,
			},
			numOps: 5,
		},
		{
			name: "10 concurrent operations",
			config: &RedisConfig{
				Host:              getEnv("REDIS_HOST", "localhost"),
				Port:              6379,
				PoolSize:          15,
				PoolTimeout:       10 * time.Second,
				ConnectionIdleMin: 3,
			},
			numOps: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			client, err := NewRedisConnection(ctx, tt.config)
			if err != nil {
				t.Fatalf("NewRedisConnection() error = %v", err)
			}
			defer client.Close()

			done := make(chan error, tt.numOps)

			for i := 0; i < tt.numOps; i++ {
				go func(idx int) {
					key := "test:concurrent:" + string(rune('a'+idx))
					err := client.Set(ctx, key, idx, time.Minute).Err()
					if err != nil {
						done <- err
						return
					}
					_, err = client.Get(ctx, key).Result()
					if err != nil {
						done <- err
						return
					}
					done <- nil
				}(i)
			}

			for i := 0; i < tt.numOps; i++ {
				select {
				case err := <-done:
					if err != nil {
						t.Errorf("Concurrent operation failed: %v", err)
					}
				case <-time.After(5 * time.Second):
					t.Error("Timeout waiting for concurrent operations")
				}
			}
		})
	}
}

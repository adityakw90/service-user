package infra

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

/*
this test should use real postgre server
*/

func TestInfra_PostgreConnection(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name        string
		config      *PostgreConfig
		wantPingErr bool
	}{
		{
			name: "successful connection with default config",
			config: &PostgreConfig{
				Host:                  getEnv("DATABASE_HOST", "localhost"),
				Port:                  5432,
				User:                  getEnv("DATABASE_USER", "postgres"),
				Password:              getEnv("DATABASE_PASSWORD", "postgres"),
				Name:                  getEnv("DATABASE_NAME", "service_db"),
				SslMode:               "disable",
				Timezone:              "UTC",
				MinConns:              1,
				MinIdleConns:          1,
				MaxConns:              3,
				MaxConnIdleTime:       30 * time.Minute,
				MaxConnLifetime:       1 * time.Hour,
				MaxConnLifetimeJitter: 5 * time.Minute,
				HealthCheckPeriod:     1 * time.Minute,
			},
			wantPingErr: false,
		},
		{
			name: "failed connection with invalid host",
			config: &PostgreConfig{
				Host:                  "invalid-host-that-does-not-exist",
				Port:                  5432,
				User:                  "testuser",
				Password:              "testpass",
				Name:                  "testdb",
				SslMode:               "disable",
				MinConns:              1,
				MinIdleConns:          1,
				MaxConns:              1,
				MaxConnIdleTime:       1 * time.Minute,
				MaxConnLifetime:       5 * time.Minute,
				MaxConnLifetimeJitter: 1 * time.Minute,
				HealthCheckPeriod:     30 * time.Second,
			},
			wantPingErr: true,
		},
		{
			name: "failed connection with invalid credentials",
			config: &PostgreConfig{
				Host:                  getEnv("DATABASE_HOST", "localhost"),
				Port:                  5432,
				User:                  "invalid_user",
				Password:              "invalid_password",
				Name:                  getEnv("DATABASE_NAME", "service_db"),
				SslMode:               "disable",
				MinConns:              1,
				MinIdleConns:          1,
				MaxConns:              1,
				MaxConnIdleTime:       1 * time.Minute,
				MaxConnLifetime:       5 * time.Minute,
				MaxConnLifetimeJitter: 1 * time.Minute,
				HealthCheckPeriod:     30 * time.Second,
			},
			wantPingErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			pool, err := NewPostgreConnection(ctx, tt.config)
			if err != nil {
				if !tt.wantPingErr {
					t.Errorf("NewPostgreConnection() error = %v", err)
				}
				return
			}
			defer pool.Close()

			if tt.wantPingErr {
				t.Error("NewPostgreConnection() expected error, got nil")
				return
			}

			if err := pool.Ping(ctx); err != nil {
				t.Errorf("Ping() error = %v", err)
			}
		})
	}
}

func TestInfra_PostgreConnection_PoolConfiguration(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name       string
		config     *PostgreConfig
		query      string
		queryArg   any
		wantResult int
	}{
		{
			name: "basic query with single value",
			config: &PostgreConfig{
				Host:                  getEnv("DATABASE_HOST", "localhost"),
				Port:                  5432,
				User:                  getEnv("DATABASE_USER", "postgres"),
				Password:              getEnv("DATABASE_PASSWORD", "postgres"),
				Name:                  getEnv("DATABASE_NAME", "service_db"),
				SslMode:               "disable",
				Timezone:              "UTC",
				MinConns:              2,
				MinIdleConns:          1,
				MaxConns:              5,
				MaxConnIdleTime:       10 * time.Minute,
				MaxConnLifetime:       30 * time.Minute,
				MaxConnLifetimeJitter: 5 * time.Minute,
				HealthCheckPeriod:     5 * time.Minute,
			},
			query:      "SELECT $1::int",
			queryArg:   42,
			wantResult: 42,
		},
		{
			name: "current timestamp",
			config: &PostgreConfig{
				Host:                  getEnv("DATABASE_HOST", "localhost"),
				Port:                  5432,
				User:                  getEnv("DATABASE_USER", "postgres"),
				Password:              getEnv("DATABASE_PASSWORD", "postgres"),
				Name:                  getEnv("DATABASE_NAME", "service_db"),
				SslMode:               "disable",
				Timezone:              "UTC",
				MinConns:              1,
				MinIdleConns:          1,
				MaxConns:              2,
				MaxConnIdleTime:       30 * time.Minute,
				MaxConnLifetime:       1 * time.Hour,
				MaxConnLifetimeJitter: 5 * time.Minute,
				HealthCheckPeriod:     1 * time.Minute,
			},
			query:      "SELECT EXTRACT(YEAR FROM NOW())",
			queryArg:   nil,
			wantResult: time.Now().Year(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			pool, err := NewPostgreConnection(ctx, tt.config)
			if err != nil {
				t.Fatalf("NewPostgreConnection() error = %v", err)
			}
			defer pool.Close()

			var result int
			if tt.queryArg != nil {
				err = pool.QueryRow(ctx, tt.query, tt.queryArg).Scan(&result)
			} else {
				err = pool.QueryRow(ctx, tt.query).Scan(&result)
			}
			if err != nil {
				t.Errorf("QueryRow() error = %v", err)
			}
			if result != tt.wantResult {
				t.Errorf("Result = %d, want %d", result, tt.wantResult)
			}
		})
	}
}

func TestInfra_PostgreConnection_ConcurrentQueries(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") == "1" {
		t.Skip("Skipping integration test")
	}

	tests := []struct {
		name       string
		numQueries int
	}{
		{
			name:       "5 concurrent queries",
			numQueries: 5,
		},
		{
			name:       "10 concurrent queries",
			numQueries: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := &PostgreConfig{
				Host:                  getEnv("DATABASE_HOST", "localhost"),
				Port:                  5432,
				User:                  getEnv("DATABASE_USER", "postgres"),
				Password:              getEnv("DATABASE_PASSWORD", "postgres"),
				Name:                  getEnv("DATABASE_NAME", "service_db"),
				SslMode:               "disable",
				Timezone:              "UTC",
				MinConns:              1,
				MinIdleConns:          1,
				MaxConns:              5,
				MaxConnIdleTime:       30 * time.Minute,
				MaxConnLifetime:       1 * time.Hour,
				MaxConnLifetimeJitter: 5 * time.Minute,
				HealthCheckPeriod:     1 * time.Minute,
			}
			cfg.MaxConns = int32(tt.numQueries + 1)

			pool, err := NewPostgreConnection(ctx, cfg)
			if err != nil {
				t.Fatalf("NewPostgreConnection() error = %v", err)
			}
			defer pool.Close()

			done := make(chan error, tt.numQueries)

			for i := 0; i < tt.numQueries; i++ {
				go func(idx int) {
					var result int
					// Use explicit type cast for integer parameter
					err := pool.QueryRow(ctx, "SELECT $1::int", idx).Scan(&result)
					if err != nil {
						done <- err
						return
					}
					done <- nil
				}(i)
			}

			for i := 0; i < tt.numQueries; i++ {
				select {
				case err := <-done:
					if err != nil {
						t.Errorf("Concurrent query failed: %v", err)
					}
				case <-time.After(5 * time.Second):
					t.Error("Timeout waiting for concurrent queries")
				}
			}
		})
	}
}

func TestInfra_getPostgreQueryExecMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want pgx.QueryExecMode
	}{
		{
			name: "cache_statement mode",
			mode: "cache_statement",
			want: pgx.QueryExecModeCacheStatement,
		},
		{
			name: "cache_describe mode",
			mode: "cache_describe",
			want: pgx.QueryExecModeCacheDescribe,
		},
		{
			name: "simple_protocol mode",
			mode: "simple_protocol",
			want: pgx.QueryExecModeSimpleProtocol,
		},
		{
			name: "empty string defaults to cache_describe",
			mode: "",
			want: pgx.QueryExecModeCacheDescribe,
		},
		{
			name: "invalid mode defaults to cache_describe",
			mode: "invalid_mode",
			want: pgx.QueryExecModeCacheDescribe,
		},
		{
			name: "unknown mode defaults to cache_describe",
			mode: "unknown",
			want: pgx.QueryExecModeCacheDescribe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPostgreQueryExecMode(tt.mode)
			if got != tt.want {
				t.Errorf("getPostgreQueryExecMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

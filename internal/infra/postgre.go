package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgreConfig struct {
	Host                  string
	Port                  int
	User                  string
	Password              string
	Name                  string
	SslMode               string
	Timezone              string
	MinConns              int32
	MinIdleConns          int32
	MaxConns              int32
	MaxConnIdleTime       time.Duration
	MaxConnLifetime       time.Duration
	MaxConnLifetimeJitter time.Duration
	HealthCheckPeriod     time.Duration
}

func NewPostgreConnection(ctx context.Context, cfg *PostgreConfig) (*pgxpool.Pool, error) {
	// Build connection string
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&timezone=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SslMode, cfg.Timezone,
	)

	// Create connection config
	connConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	// Configure pool settings
	connConfig.MinConns = cfg.MinConns
	connConfig.MinIdleConns = cfg.MinIdleConns
	connConfig.MaxConns = cfg.MaxConns
	connConfig.MaxConnLifetime = cfg.MaxConnLifetime
	connConfig.MaxConnLifetimeJitter = cfg.MaxConnLifetimeJitter
	connConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	connConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Create pool with config
	pgxPool, err := pgxpool.NewWithConfig(ctx, connConfig)
	if err != nil {
		return nil, err
	}

	// Ping to verify connection
	if err := pgxPool.Ping(ctx); err != nil {
		return nil, err
	}

	return pgxPool, nil
}

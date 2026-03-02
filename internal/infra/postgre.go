package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
	QueryExecMode         string
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
	connConfig.ConnConfig.DefaultQueryExecMode = getPostgreQueryExecMode(cfg.QueryExecMode)

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

func getPostgreQueryExecMode(mode string) pgx.QueryExecMode {
	// QueryExecMode controls how queries are executed. Options:
	// - "cache_statement": Best performance, direct PostgreSQL only (may cause "prepared statement name already in use" errors)
	// - "cache_describe": Excellent performance, works with pgbouncer and direct PostgreSQL (recommended)
	// - "simple_protocol": Compatible with all proxies, slower performance
	// Default: "cache_describe"
	switch mode {
	case "cache_statement":
		// Best performance for direct PostgreSQL only
		// WARNING: May cause "prepared statement name already in use" errors with pgbouncer
		return pgx.QueryExecModeCacheStatement
	case "cache_describe":
		// Excellent performance, works with pgbouncer and direct PostgreSQL (recommended)
		return pgx.QueryExecModeCacheDescribe
	case "simple_protocol":
		// Compatible with all proxies, slower performance
		return pgx.QueryExecModeSimpleProtocol
	default:
		// Invalid value, fall back to recommended default
		// Default to "cache_describe" for best compatibility with pgbouncer and direct PostgreSQL
		return pgx.QueryExecModeCacheDescribe
	}
}

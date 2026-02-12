package testutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/adityakw90/service-user/internal/config"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewTestPostgreConnection(t *testing.T, ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	t.Helper()

	lockFile, err := AcquireLock("postgre")
	if err != nil {
		return nil, fmt.Errorf("failed to lock postgres: %w", err)
	}
	t.Cleanup(func() {
		ReleaseLock(lockFile)
	})

	client, err := infra.NewPostgreConnection(ctx, &infra.PostgreConfig{
		Host:                  cfg.Database.Host,
		Port:                  cfg.Database.Port,
		User:                  cfg.Database.User,
		Password:              cfg.Database.Password,
		Name:                  cfg.Database.Name,
		SslMode:               cfg.Database.SslMode,
		Timezone:              cfg.Database.Timezone,
		MinConns:              cfg.Database.MinConns,
		MinIdleConns:          cfg.Database.MinIdleConns,
		MaxConns:              cfg.Database.MaxConns,
		MaxConnIdleTime:       cfg.Database.MaxConnIdleTime,
		MaxConnLifetime:       cfg.Database.MaxConnLifetime,
		MaxConnLifetimeJitter: cfg.Database.MaxConnLifetimeJitter,
		HealthCheckPeriod:     cfg.Database.HealthCheckPeriod,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	t.Cleanup(func() {
		client.Close()
	})

	if err := truncateTestTablesPostgreDB(t, ctx, client); err != nil {
		return nil, fmt.Errorf("failed to truncate test tables: %w", err)
	}

	return client, nil
}

func truncateTestTablesPostgreDB(t *testing.T, ctx context.Context, client *pgxpool.Pool) error {
	t.Helper()

	// get all table names
	var tables []string
	err := client.QueryRow(ctx, `
		SELECT array_agg(tablename)
		FROM pg_tables
		WHERE schemaname = 'public' AND tablename != 'databasechangelog' AND tablename != 'databasechangeloglock'
	`).Scan(&tables)
	if err != nil {
		t.Fatalf("Failed to get table names: %v", err)
	}

	if len(tables) == 0 {
		t.Fatalf("No tables found")
	}

	tx, err := client.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// truncate all tables in a single statement
	// using CASCADE to handle foreign key constraints automatically
	query := fmt.Sprintf(`TRUNCATE TABLE "%s"`, tables[0])
	for i := 1; i < len(tables); i++ {
		query += fmt.Sprintf(`, "%s"`, tables[i])
	}
	query += " RESTART IDENTITY CASCADE"

	if _, err := tx.Exec(ctx, query); err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("Failed to commit transaction: %v", err)
	}
	return nil
}

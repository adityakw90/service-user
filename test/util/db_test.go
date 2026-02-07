package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestTestDB_TruncateTestTables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	cfg, err := LoadTestConfig(t)
	require.NoError(t, err)

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SslMode,
	)
	dbPool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)
	defer dbPool.Close()

	// ensure clean state
	TruncateTestTables(t, ctx, dbPool)

	// insert a test user
	t.Log("Inserting test user")
	user := &model.User{
		UID:       uuid.New().String(),
		Username:  "test_truncate",
		Email:     "test_truncate@example.com",
		Password:  "password",
		Status:    model.UserStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err = CreateTestUser(ctx, dbPool, user)
	require.NoError(t, err)

	// verify user exists
	t.Log("Verifying user exists before truncate")
	AssertUserCount(t, ctx, dbPool, 1)

	// truncate tables
	t.Log("Truncating tables")
	TruncateTestTables(t, ctx, dbPool)

	// verify user table is empty
	t.Log("Verifying user table is empty after truncate")
	AssertUserCount(t, ctx, dbPool, 0)
}

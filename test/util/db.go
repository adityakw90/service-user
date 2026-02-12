package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// func TruncateTestTables(t *testing.T, ctx context.Context, db *pgxpool.Pool) {
// 	t.Helper()

// 	// get all table names
// 	var tables []string
// 	err := db.QueryRow(ctx, `
// 		SELECT array_agg(tablename)
// 		FROM pg_tables
// 		WHERE schemaname = 'public' AND tablename != 'databasechangelog' AND tablename != 'databasechangeloglock'
// 	`).Scan(&tables)
// 	if err != nil {
// 		t.Fatalf("Failed to get table names: %v", err)
// 	}

// 	if len(tables) == 0 {
// 		return
// 	}

// 	tx, err := db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
// 	if err != nil {
// 		t.Fatalf("Failed to begin transaction: %v", err)
// 	}
// 	defer func() {
// 		_ = tx.Rollback(ctx)
// 	}()

// 	// truncate all tables in a single statement
// 	// using CASCADE to handle foreign key constraints automatically
// 	query := fmt.Sprintf(`TRUNCATE TABLE "%s"`, tables[0])
// 	for i := 1; i < len(tables); i++ {
// 		query += fmt.Sprintf(`, "%s"`, tables[i])
// 	}
// 	query += " CASCADE"

// 	if _, err := tx.Exec(ctx, query); err != nil {
// 		t.Fatalf("Failed to truncate tables: %v", err)
// 	}

// 	if err := tx.Commit(ctx); err != nil {
// 		t.Fatalf("Failed to commit transaction: %v", err)
// 	}
// }

// CreateTestUser inserts a user directly into the database for fixtures.
// This bypasses the service layer for faster test setup.
func CreateTestUser(ctx context.Context, db *pgxpool.Pool, user *model.User) error {
	query := `
		INSERT INTO "user" (uid, username, email, password, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	err := db.QueryRow(
		ctx,
		query,
		user.UID,
		user.Username,
		user.Email,
		user.Password,
		user.Status,
		time.Now(),
		time.Now(),
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

// SetupTestDatabase truncates all test tables in foreign key order.
// Use this to clean the database between tests.
func SetupTestDatabase(ctx context.Context, db pgxExecutor) error {
	// Delete in FK order to avoid constraint violations
	tables := []string{
		"user_device",
		"user_pin",
		"user_profile",
		"user_file",
		"user",
	}

	for _, table := range tables {
		if _, err := db.Exec(ctx, fmt.Sprintf(`DELETE FROM "%s"`, table)); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// GetTestUser retrieves a user by UID for assertions.
func GetTestUser(ctx context.Context, db pgxExecutor, uid string) (*model.User, error) {
	query := `
		SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at
		FROM "user"
		WHERE uid = $1`

	var user model.User
	err := db.QueryRow(ctx, query, uid).Scan(
		&user.ID,
		&user.UID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// AssertUserCount verifies the number of users in the database.
func AssertUserCount(t require.TestingT, ctx context.Context, db pgxExecutor, expected int) {
	var count int
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM "user"`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, expected, count)
}

// WaitForDatabase polls for database readiness.
// Useful for tests that run in containers where DB may not be immediately available.
func WaitForDatabase(ctx context.Context, dbURL string, maxAttempts int) error {
	var lastErr error

	for i := 0; i < maxAttempts; i++ {
		conn, err := pgx.Connect(ctx, dbURL)
		if err == nil {
			// Successfully connected
			conn.Close(ctx)
			return nil
		}
		lastErr = err

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return fmt.Errorf("database not ready after %d attempts: %w", maxAttempts, lastErr)
}

// CreateTestUserWithProfile creates both user and profile records.
func CreateTestUserWithProfile(ctx context.Context, db *pgxpool.Pool, user *model.User, profile *model.UserProfile) error {
	// First create the user
	if err := CreateTestUser(ctx, db, user); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Then create the profile
	query := `
		INSERT INTO "user_profile" (user_id, user_uid, first_name, last_name, bio, attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	err := db.QueryRow(
		ctx,
		query,
		user.ID,
		user.UID,
		profile.FirstName,
		profile.LastName,
		profile.Bio,
		profile.Attributes,
		time.Now(),
		time.Now(),
	).Scan(&profile.CreatedAt, &profile.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	profile.UserID = user.ID
	profile.UserUID = user.UID

	return nil
}

// pgxExecutor defines the interface for DB operations.
// This allows both pgx.Conn and pgxpool.Pool to be used.
type pgxExecutor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// UserToCreateParam converts a model.User to UserCreateParam for service calls.
func UserToCreateParam(user *model.User, password string) *params.UserCreateParam {
	return &params.UserCreateParam{
		Username: user.Username,
		Email:    user.Email,
		Password: password,
	}
}

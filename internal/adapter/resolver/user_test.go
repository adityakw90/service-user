package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/infra"
	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserResolver_FetchIDFromDB(t *testing.T) {
	tests := []struct {
		name      string
		uid       string
		setupMock func(*testing.T, pgxmock.PgxPoolIface)
		wantID    int64
		wantUID   string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "successful fetch",
			uid:  "user-uid-123",
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-123")
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE uid=$1").
					WithArgs("user-uid-123").
					WillReturnRows(rows)
			},
			wantID:  100,
			wantUID: "user-uid-123",
			wantErr: false,
		},
		{
			name: "not found - no rows",
			uid:  "nonexistent-uid",
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE uid=$1").
					WithArgs("nonexistent-uid").
					WillReturnRows(rows)
			},
			wantID:    0,
			wantUID:   "",
			wantErr:   true,
			wantErrIs: domainerrors.ErrUserNotFound,
		},
		{
			name: "database error",
			uid:  "error-uid",
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE uid=$1").
					WithArgs("error-uid").
					WillReturnError(errors.New("database connection error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(t, mockPool)

			r := &userResolver{
				db: mockPool,
			}

			got, err := r.fetchIDFromDB(context.Background(), tt.uid)

			if (err != nil) != tt.wantErr {
				t.Errorf("fetchIDFromDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("fetchIDFromDB() error = %v, wantErrIs %v", err, tt.wantErrIs)
			}

			if err == nil && got.id != tt.wantID {
				t.Errorf("fetchIDFromDB() id = %v, want %v", got.id, tt.wantID)
			}

			if err == nil && got.uid != tt.wantUID {
				t.Errorf("fetchIDFromDB() uid = %v, want %v", got.uid, tt.wantUID)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserResolver_FetchUIDFromDB(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(*testing.T, pgxmock.PgxPoolIface)
		wantID    int64
		wantUID   string
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "successful fetch",
			id:   100,
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-123")
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE id=$1").
					WithArgs(int64(100)).
					WillReturnRows(rows)
			},
			wantID:  100,
			wantUID: "user-uid-123",
			wantErr: false,
		},
		{
			name: "not found - no rows",
			id:   999,
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE id=$1").
					WithArgs(int64(999)).
					WillReturnRows(rows)
			},
			wantID:    0,
			wantUID:   "",
			wantErr:   true,
			wantErrIs: domainerrors.ErrUserNotFound,
		},
		{
			name: "database error",
			id:   500,
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE id=$1").
					WithArgs(int64(500)).
					WillReturnError(errors.New("database connection error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			tt.setupMock(t, mockPool)

			r := &userResolver{
				db: mockPool,
			}

			got, err := r.fetchUIDFromDB(context.Background(), tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("fetchUIDFromDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("fetchUIDFromDB() error = %v, wantErrIs %v", err, tt.wantErrIs)
			}

			if err == nil && got.id != tt.wantID {
				t.Errorf("fetchUIDFromDB() id = %v, want %v", got.id, tt.wantID)
			}

			if err == nil && got.uid != tt.wantUID {
				t.Errorf("fetchUIDFromDB() uid = %v, want %v", got.uid, tt.wantUID)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestNewUserResolver(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "valid creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			logger := infra.NewNoopLogger()
			tracer := infra.NewNoopTracer()

			got := NewUserResolver(mockPool, nil, "test", time.Minute, logger, tracer)

			assert.NotNil(t, got)
			_, ok := got.(*userResolver)
			assert.True(t, ok, "NewUserResolver() did not return *userResolver type")
		})
	}
}

func TestUserResolver_IDsByUIDs(t *testing.T) {
	tests := []struct {
		name           string
		userUIDs       []string
		setupDBMock    func(*testing.T, pgxmock.PgxPoolIface)
		wantErr        bool
		wantErrIs      error
		validateResult func(*testing.T, map[string]int64)
	}{
		{
			name:     "empty input returns empty map",
			userUIDs: []string{},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				// No DB calls expected for empty input
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[string]int64) {
				if len(result) != 0 {
					t.Errorf("expected empty map, got %d entries", len(result))
				}
			},
		},
		{
			name:     "single user fetch from DB",
			userUIDs: []string{"user-uid-1"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE uid=$1").
					WithArgs("user-uid-1").
					WillReturnRows(rows)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[string]int64) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result["user-uid-1"] != 100 {
					t.Errorf("expected id 100, got %d", result["user-uid-1"])
				}
			},
		},
		{
			name:     "single user fetch from DB",
			userUIDs: []string{"user-uid-1"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE uid=$1").
					WithArgs("user-uid-1").
					WillReturnRows(pgxmock.NewRows([]string{"id", "uid"}).
						AddRow(int64(100), "user-uid-1"))
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[string]int64) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result["user-uid-1"] != 100 {
					t.Errorf("expected id 100, got %d", result["user-uid-1"])
				}
			},
		},
		{
			name:     "user not found returns error",
			userUIDs: []string{"nonexistent-uid"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE uid=$1").
					WithArgs("nonexistent-uid").
					WillReturnRows(rows)
			},
			wantErr:   true,
			wantErrIs: domainerrors.ErrUserNotFound,
		},
		{
			name:     "database error",
			userUIDs: []string{"error-uid"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE uid=$1").
					WithArgs("error-uid").
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup pgxmock
			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup miniredis
			s := miniredis.RunT(t)
			redisClient := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer redisClient.Close()

			// Setup test expectations
			tt.setupDBMock(t, mockPool)

			logger := infra.NewNoopLogger()
			tracer := infra.NewNoopTracer()

			r := &userResolver{
				db:                 mockPool,
				redisClient:        redisClient,
				redisPrefix:        "user",
				redisCacheDuration: time.Minute,
				logger:             logger,
				tracer:             tracer,
			}

			got, err := r.IDsByUIDs(context.Background(), tt.userUIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("IDsByUIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("IDsByUIDs() error = %v, wantErrIs %v", err, tt.wantErrIs)
			}

			if tt.validateResult != nil && err == nil {
				tt.validateResult(t, got)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserResolver_UIDsByIDs(t *testing.T) {
	tests := []struct {
		name           string
		userIDs        []int64
		setupDBMock    func(*testing.T, pgxmock.PgxPoolIface)
		wantErr        bool
		wantErrIs      error
		validateResult func(*testing.T, map[int64]string)
	}{
		{
			name:    "empty input returns empty map",
			userIDs: []int64{},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				// No DB calls expected for empty input
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[int64]string) {
				if len(result) != 0 {
					t.Errorf("expected empty map, got %d entries", len(result))
				}
			},
		},
		{
			name:    "single user fetch from DB",
			userIDs: []int64{100},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE id=$1").
					WithArgs(int64(100)).
					WillReturnRows(rows)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[int64]string) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result[100] != "user-uid-1" {
					t.Errorf("expected uid user-uid-1, got %s", result[100])
				}
			},
		},
		{
			name:    "single user fetch from DB",
			userIDs: []int64{100},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE id=$1").
					WithArgs(int64(100)).
					WillReturnRows(pgxmock.NewRows([]string{"id", "uid"}).
						AddRow(int64(100), "user-uid-1"))
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[int64]string) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result[100] != "user-uid-1" {
					t.Errorf("expected uid user-uid-1, got %s", result[100])
				}
			},
		},
		{
			name:    "user not found returns error",
			userIDs: []int64{999},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE id=$1").
					WithArgs(int64(999)).
					WillReturnRows(rows)
			},
			wantErr:   true,
			wantErrIs: domainerrors.ErrUserNotFound,
		},
		{
			name:    "database error",
			userIDs: []int64{500},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM \"user\" WHERE id=$1").
					WithArgs(int64(500)).
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup pgxmock
			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup miniredis
			s := miniredis.RunT(t)
			redisClient := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer redisClient.Close()

			// Setup test expectations
			tt.setupDBMock(t, mockPool)

			logger := infra.NewNoopLogger()
			tracer := infra.NewNoopTracer()

			r := &userResolver{
				db:                 mockPool,
				redisClient:        redisClient,
				redisPrefix:        "user",
				redisCacheDuration: time.Minute,
				logger:             logger,
				tracer:             tracer,
			}

			got, err := r.UIDsByIDs(context.Background(), tt.userIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("UIDsByIDs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("UIDsByIDs() error = %v, wantErrIs %v", err, tt.wantErrIs)
			}

			if tt.validateResult != nil && err == nil {
				tt.validateResult(t, got)
			}

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserResolver_Invalidate(t *testing.T) {
	tests := []struct {
		name    string
		opts    []param.InvalidateOpt
		setup   func(s *miniredis.Miniredis)
		wantErr bool
		verify  func(s *miniredis.Miniredis)
	}{
		{
			name: "Happy Path - invalidate by UID",
			opts: []param.InvalidateOpt{
				param.WithUIDs("user-uid-1"),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:user-uid-1:id", "100")
				s.Set("user:id:100:uid", "user-uid-1")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				assert.False(t, s.Exists("user:user-uid-1:id"))
				assert.False(t, s.Exists("user:id:100:uid"))
			},
		},
		{
			name: "Happy Path - invalidate by ID",
			opts: []param.InvalidateOpt{
				param.WithIDs(100),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:user-uid-1:id", "100")
				s.Set("user:id:100:uid", "user-uid-1")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				assert.False(t, s.Exists("user:user-uid-1:id"))
				assert.False(t, s.Exists("user:id:100:uid"))
			},
		},
		{
			name: "Happy Path - invalidate multiple UIDs",
			opts: []param.InvalidateOpt{
				param.WithUIDs("user-uid-1", "user-uid-2"),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:user-uid-1:id", "100")
				s.Set("user:id:100:uid", "user-uid-1")
				s.Set("user:user-uid-2:id", "200")
				s.Set("user:id:200:uid", "user-uid-2")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				assert.False(t, s.Exists("user:user-uid-1:id"))
				assert.False(t, s.Exists("user:id:100:uid"))
				assert.False(t, s.Exists("user:user-uid-2:id"))
				assert.False(t, s.Exists("user:id:200:uid"))
			},
		},
		{
			name: "Happy Path - invalidate mixed UIDs and IDs",
			opts: []param.InvalidateOpt{
				param.WithUIDs("user-uid-1"),
				param.WithIDs(200),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:user-uid-1:id", "100")
				s.Set("user:id:100:uid", "user-uid-1")
				s.Set("user:user-uid-2:id", "200")
				s.Set("user:id:200:uid", "user-uid-2")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				assert.False(t, s.Exists("user:user-uid-1:id"))
				assert.False(t, s.Exists("user:id:100:uid"))
				assert.False(t, s.Exists("user:user-uid-2:id"))
				assert.False(t, s.Exists("user:id:200:uid"))
			},
		},
		{
			name: "Happy Path - empty options (no-op)",
			opts: []param.InvalidateOpt{},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:user-uid-1:id", "100")
				s.Set("user:id:100:uid", "user-uid-1")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				assert.True(t, s.Exists("user:user-uid-1:id"))
				assert.True(t, s.Exists("user:id:100:uid"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup miniredis
			s := miniredis.RunT(t)
			defer s.Close()

			redisClient := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer redisClient.Close()

			// Setup cache
			if tt.setup != nil {
				tt.setup(s)
			}

			logger := infra.NewNoopLogger()
			tracer := infra.NewNoopTracer()

			r := &userResolver{
				redisClient:        redisClient,
				redisPrefix:        "user",
				redisCacheDuration: time.Minute,
				logger:             logger,
				tracer:             tracer,
			}

			ctx := context.Background()

			// Execute
			err := r.Invalidate(ctx, tt.opts...)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify state
			if tt.verify != nil {
				tt.verify(s)
			}
		})
	}
}

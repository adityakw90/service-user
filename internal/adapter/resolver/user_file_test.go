package resolver

import (
	"context"
	"errors"
	"testing"
	"time"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFileResolver_FetchIDFromDB(t *testing.T) {
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
			uid:  "file-uid-123",
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "file-uid-123")
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE uid=$1").
					WithArgs("file-uid-123").
					WillReturnRows(rows)
			},
			wantID:  100,
			wantUID: "file-uid-123",
			wantErr: false,
		},
		{
			name: "not found - no rows",
			uid:  "nonexistent-uid",
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE uid=$1").
					WithArgs("nonexistent-uid").
					WillReturnRows(rows)
			},
			wantID:    0,
			wantUID:   "",
			wantErr:   true,
			wantErrIs: domainerrors.ErrFileNotFound,
		},
		{
			name: "database error",
			uid:  "error-uid",
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE uid=$1").
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

			r := &userFileResolver{
				db: mockPool,
			}

			got, err := r.fetchIDFromDB(tt.uid)

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

func TestUserFileResolver_FetchUIDFromDB(t *testing.T) {
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
					AddRow(int64(100), "file-uid-123")
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE id=$1").
					WithArgs(int64(100)).
					WillReturnRows(rows)
			},
			wantID:  100,
			wantUID: "file-uid-123",
			wantErr: false,
		},
		{
			name: "not found - no rows",
			id:   999,
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE id=$1").
					WithArgs(int64(999)).
					WillReturnRows(rows)
			},
			wantID:    0,
			wantUID:   "",
			wantErr:   true,
			wantErrIs: domainerrors.ErrFileNotFound,
		},
		{
			name: "database error",
			id:   500,
			setupMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE id=$1").
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

			r := &userFileResolver{
				db: mockPool,
			}

			got, err := r.fetchUIDFromDB(tt.id)

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

func TestNewUserFileResolver(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "valid creation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			logger := &mockLogger{}
			tracer := newNoOpTracer()

			got := NewUserFileResolver(mockPool, nil, "test", time.Minute, logger, tracer)

			if got == nil {
				t.Error("NewUserFileResolver() returned nil")
			}

			if _, ok := got.(*userFileResolver); !ok {
				t.Error("NewUserFileResolver() did not return *userFileResolver type")
			}
		})
	}
}

func TestUserFileResolver_IDsByUIDs(t *testing.T) {
	tests := []struct {
		name           string
		userFileUIDs   []string
		setupDBMock    func(*testing.T, pgxmock.PgxPoolIface)
		wantErr        bool
		wantErrIs      error
		validateResult func(*testing.T, map[string]int64)
	}{
		{
			name:         "empty input returns empty map",
			userFileUIDs: []string{},
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
			name:         "single file fetch from DB",
			userFileUIDs: []string{"file-uid-1"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "file-uid-1")
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE uid=$1").
					WithArgs("file-uid-1").
					WillReturnRows(rows)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[string]int64) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result["file-uid-1"] != 100 {
					t.Errorf("expected id 100, got %d", result["file-uid-1"])
				}
			},
		},
		{
			name:         "single file fetch from DB",
			userFileUIDs: []string{"file-uid-1"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE uid=$1").
					WithArgs("file-uid-1").
					WillReturnRows(pgxmock.NewRows([]string{"id", "uid"}).
						AddRow(int64(100), "file-uid-1"))
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[string]int64) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result["file-uid-1"] != 100 {
					t.Errorf("expected id 100, got %d", result["file-uid-1"])
				}
			},
		},
		{
			name:         "file not found returns error",
			userFileUIDs: []string{"nonexistent-uid"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE uid=$1").
					WithArgs("nonexistent-uid").
					WillReturnRows(rows)
			},
			wantErr:   true,
			wantErrIs: domainerrors.ErrFileNotFound,
		},
		{
			name:         "database error",
			userFileUIDs: []string{"error-uid"},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE uid=$1").
					WithArgs("error-uid").
					WillReturnError(errors.New("database error"))
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

			// Setup miniredis
			redisClient, redisCleanup, err := newMockRedis()
			require.NoError(t, err)
			defer redisCleanup()

			// Setup test expectations
			tt.setupDBMock(t, mockPool)

			logger := &mockLogger{}
			tracer := newNoOpTracer()

			r := &userFileResolver{
				db:                 mockPool,
				redisClient:        redisClient,
				redisPrefix:        "userfile",
				redisCacheDuration: time.Minute,
				logger:             logger,
				tracer:             tracer,
			}

			got, err := r.IDsByUIDs(context.Background(), tt.userFileUIDs)

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

func TestUserFileResolver_UIDsByIDs(t *testing.T) {
	tests := []struct {
		name           string
		userFileIDs    []int64
		setupDBMock    func(*testing.T, pgxmock.PgxPoolIface)
		wantErr        bool
		wantErrIs      error
		validateResult func(*testing.T, map[int64]string)
	}{
		{
			name:        "empty input returns empty map",
			userFileIDs: []int64{},
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
			name:        "single file fetch from DB",
			userFileIDs: []int64{100},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "file-uid-1")
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE id=$1").
					WithArgs(int64(100)).
					WillReturnRows(rows)
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[int64]string) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result[100] != "file-uid-1" {
					t.Errorf("expected uid file-uid-1, got %s", result[100])
				}
			},
		},
		{
			name:        "single file fetch from DB",
			userFileIDs: []int64{100},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE id=$1").
					WithArgs(int64(100)).
					WillReturnRows(pgxmock.NewRows([]string{"id", "uid"}).
						AddRow(int64(100), "file-uid-1"))
			},
			wantErr: false,
			validateResult: func(t *testing.T, result map[int64]string) {
				if len(result) != 1 {
					t.Errorf("expected 1 entry, got %d", len(result))
				}
				if result[100] != "file-uid-1" {
					t.Errorf("expected uid file-uid-1, got %s", result[100])
				}
			},
		},
		{
			name:        "file not found returns error",
			userFileIDs: []int64{999},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE id=$1").
					WithArgs(int64(999)).
					WillReturnRows(rows)
			},
			wantErr:   true,
			wantErrIs: domainerrors.ErrFileNotFound,
		},
		{
			name:        "database error",
			userFileIDs: []int64{500},
			setupDBMock: func(t *testing.T, mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("SELECT id, uid FROM user_file WHERE id=$1").
					WithArgs(int64(500)).
					WillReturnError(errors.New("database error"))
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

			// Setup miniredis
			redisClient, redisCleanup, err := newMockRedis()
			require.NoError(t, err)
			defer redisCleanup()

			// Setup test expectations
			tt.setupDBMock(t, mockPool)

			logger := &mockLogger{}
			tracer := newNoOpTracer()

			r := &userFileResolver{
				db:                 mockPool,
				redisClient:        redisClient,
				redisPrefix:        "userfile",
				redisCacheDuration: time.Minute,
				logger:             logger,
				tracer:             tracer,
			}

			got, err := r.UIDsByIDs(context.Background(), tt.userFileIDs)

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

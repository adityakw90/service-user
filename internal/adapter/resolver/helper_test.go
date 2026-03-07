package resolver

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
)

// testIdentity represents the database model for testing
type testIdentity struct {
	id  int64
	uid string
}

// TestMapperID_CacheHit tests mapperID when all items are in cache
func TestMapperID_CacheHit(t *testing.T) {
	tests := []struct {
		name    string
		uids    []string
		setup   func(s *miniredis.Miniredis)
		want    map[string]int64
		wantErr bool
	}{
		{
			name: "Happy Path - single item in cache",
			uids: []string{"user-uid-1"},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:uid:user-uid-1", "100")
			},
			want: map[string]int64{"user-uid-1": 100},
		},
		{
			name: "Happy Path - multiple items in cache",
			uids: []string{"user-uid-1", "user-uid-2", "user-uid-3"},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:uid:user-uid-1", "100")
				s.Set("user:uid:user-uid-2", "200")
				s.Set("user:uid:user-uid-3", "300")
			},
			want: map[string]int64{
				"user-uid-1": 100,
				"user-uid-2": 200,
				"user-uid-3": 300,
			},
		},
		{
			name:  "Happy Path - empty input returns empty map",
			uids:  []string{},
			setup: func(s *miniredis.Miniredis) {},
			want:  map[string]int64{},
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

			ctx := context.Background()
			logger := &mockLogger{}

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.uids,
				func(s string) int64 {
					id, _ := strconv.ParseInt(s, 10, 64)
					return id
				},
				func(uid string) string {
					return "user:uid:" + uid
				},
				func(uid string) (*testIdentity, error) {
					// Should not be called on cache hit
					return nil, errors.New("fetchFunc should not be called")
				},
				func(user *testIdentity) int64 {
					return user.id
				},
				time.Hour,
			)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMapperID_CacheMiss tests mapperID when items need to be fetched from DB
func TestMapperID_CacheMiss(t *testing.T) {
	tests := []struct {
		name    string
		uids    []string
		setup   func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface)
		want    map[string]int64
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - single item not in cache",
			uids: []string{"user-uid-1"},
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("user-uid-1").
					WillReturnRows(rows)
			},
			want: map[string]int64{"user-uid-1": 100},
		},
		{
			name: "Partial Cache Hit - some in cache, some not",
			uids: []string{"user-uid-1", "user-uid-2", "user-uid-3"},
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				s.Set("user:uid:user-uid-1", "100")
				s.Set("user:uid:user-uid-3", "300")
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(200), "user-uid-2")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("user-uid-2").
					WillReturnRows(rows)
			},
			want: map[string]int64{
				"user-uid-1": 100,
				"user-uid-2": 200,
				"user-uid-3": 300,
			},
		},
		{
			name: "Error - user not found",
			uids: []string{"nonexistent-uid"},
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("nonexistent-uid").
					WillReturnRows(rows)
			},
			wantErr: true,
			errMsg: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup miniredis and pgxmock
			s := miniredis.RunT(t)
			defer s.Close()

			redisClient := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer redisClient.Close()

			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup cache and mocks
			if tt.setup != nil {
				tt.setup(s, mockPool)
			}

			ctx := context.Background()
			logger := &mockLogger{}

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.uids,
				func(s string) int64 {
					id, _ := strconv.ParseInt(s, 10, 64)
					return id
				},
				func(uid string) string {
					return "user:uid:" + uid
				},
				func(uid string) (*testIdentity, error) {
					rows, err := mockPool.Query(context.Background(), `SELECT id, uid FROM "user" WHERE uid=$1`, uid)
					if err != nil {
						return nil, err
					}
					defer rows.Close()

					var id int64
					var u string
					if rows.Next() {
						if err := rows.Scan(&id, &u); err != nil {
							return nil, err
						}
						return &testIdentity{id: id, uid: u}, nil
					}
					return nil, domainerrors.ErrUserNotFound
				},
				func(id *testIdentity) int64 {
					return id.id
				},
				time.Hour,
			)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMapperID_DatabaseError tests error handling when database fails
func TestMapperID_DatabaseError(t *testing.T) {
	tests := []struct {
		name    string
		uids    []string
		setup   func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface)
		wantErr bool
	}{
		{
			name: "Error - database query fails",
			uids: []string{"user-uid-1"},
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("user-uid-1").
					WillReturnError(errors.New("database connection failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup miniredis and pgxmock
			s := miniredis.RunT(t)
			defer s.Close()

			redisClient := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer redisClient.Close()

			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup mocks
			if tt.setup != nil {
				tt.setup(s, mockPool)
			}

			ctx := context.Background()
			logger := &mockLogger{}

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.uids,
				func(s string) int64 {
					id, _ := strconv.ParseInt(s, 10, 64)
					return id
				},
				func(uid string) string {
					return "user:uid:" + uid
				},
				func(uid string) (*testIdentity, error) {
					rows, err := mockPool.Query(context.Background(), `SELECT id, uid FROM "user" WHERE uid=$1`, uid)
					if err != nil {
						return nil, err
					}
					defer rows.Close()

					var id int64
					var u string
					if rows.Next() {
						if err := rows.Scan(&id, &u); err != nil {
							return nil, err
						}
						return &testIdentity{id: id, uid: u}, nil
					}
					return nil, domainerrors.ErrUserNotFound
				},
				func(id *testIdentity) int64 {
					return id.id
				},
				time.Hour,
			)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

// TestMapperID_IDToUID tests mapperID for ID to UID mapping
func TestMapperID_IDToUID(t *testing.T) {
	tests := []struct {
		name    string
		ids     []int64
		setup   func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface)
		want    map[int64]string
		wantErr bool
	}{
		{
			name: "Happy Path - single item not in cache",
			ids:  []int64{100},
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE id=$1`).
					WithArgs(int64(100)).
					WillReturnRows(rows)
			},
			want: map[int64]string{100: "user-uid-1"},
		},
		{
			name: "Partial Cache Hit - some in cache, some not",
			ids:  []int64{100, 200, 300},
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				s.Set("user:id:100", "user-uid-1")
				s.Set("user:id:300", "user-uid-3")
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(200), "user-uid-2")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE id=$1`).
					WithArgs(int64(200)).
					WillReturnRows(rows)
			},
			want: map[int64]string{
				100: "user-uid-1",
				200: "user-uid-2",
				300: "user-uid-3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup miniredis and pgxmock
			s := miniredis.RunT(t)
			defer s.Close()

			redisClient := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer redisClient.Close()

			mockPool, err := pgxmock.NewPool(
				pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual),
			)
			require.NoError(t, err)
			defer mockPool.Close()

			// Setup cache and mocks
			if tt.setup != nil {
				tt.setup(s, mockPool)
			}

			ctx := context.Background()
			logger := &mockLogger{}

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.ids,
				func(s string) string { return s },
				func(id int64) string {
					return "user:id:" + strconv.FormatInt(id, 10)
				},
				func(id int64) (*testIdentity, error) {
					rows, err := mockPool.Query(context.Background(), `SELECT id, uid FROM "user" WHERE id=$1`, id)
					if err != nil {
						return nil, err
					}
					defer rows.Close()

					var i int64
					var u string
					if rows.Next() {
						if err := rows.Scan(&i, &u); err != nil {
							return nil, err
						}
						return &testIdentity{id: i, uid: u}, nil
					}
					return nil, domainerrors.ErrUserNotFound
				},
				func(id *testIdentity) string {
					return id.uid
				},
				time.Hour,
			)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

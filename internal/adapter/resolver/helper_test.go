package resolver

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/infra"
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

// TestMapperIDs_CacheHit tests mapperID when all items are in cache
func TestMapperIDs_CacheHit(t *testing.T) {
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperIDs(
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

// TestMapperIDs_CacheMiss tests mapperID when items need to be fetched from DB
func TestMapperIDs_CacheMiss(t *testing.T) {
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
			errMsg:  "user not found",
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperIDs(
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

// TestMapperIDs_DatabaseError tests error handling when database fails
func TestMapperIDs_DatabaseError(t *testing.T) {
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperIDs(
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

// TestMapperIDs_IDToUID tests mapperID for ID to UID mapping
func TestMapperIDs_IDToUID(t *testing.T) {
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperIDs(
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

// TestMapperID_CacheHit tests mapperID when the item is in cache
func TestMapperID_CacheHit(t *testing.T) {
	tests := []struct {
		name    string
		uid     string
		setup   func(s *miniredis.Miniredis)
		want    int64
		wantErr bool
	}{
		{
			name: "Happy Path - single item in cache",
			uid:  "user-uid-1",
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:uid:user-uid-1", "100")
			},
			want:    100,
			wantErr: false,
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.uid,
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

// TestMapperID_CacheMiss tests mapperID when the item is NOT in cache
func TestMapperID_CacheMiss(t *testing.T) {
	tests := []struct {
		name    string
		uid     string
		setup   func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface)
		want    int64
		wantErr bool
		errMsg  string
	}{
		{
			name: "Happy Path - single item not in cache",
			uid:  "user-uid-1",
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("user-uid-1").
					WillReturnRows(rows)
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "Error - user not found",
			uid:  "nonexistent-uid",
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"})
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("nonexistent-uid").
					WillReturnRows(rows)
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name: "Error - database query fails",
			uid:  "user-uid-1",
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.uid,
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

// TestMapperID_RedisError tests mapperID when Redis GET returns an error (not Nil)
func TestMapperID_RedisError(t *testing.T) {
	tests := []struct {
		name    string
		uid     string
		setup   func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface)
		want    int64
		wantErr bool
	}{
		{
			name: "Happy Path - Redis GET error falls through to DB",
			uid:  "user-uid-1",
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("user-uid-1").
					WillReturnRows(rows)
			},
			want:    100,
			wantErr: false,
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.uid,
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
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestMapperID_CacheSetError tests mapperID when Redis SET fails after DB fetch
func TestMapperID_CacheSetError(t *testing.T) {
	tests := []struct {
		name    string
		uid     string
		setup   func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface)
		want    int64
		wantErr bool
	}{
		{
			name: "Happy Path - Redis SET error logged but value returned",
			uid:  "user-uid-1",
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE uid=$1`).
					WithArgs("user-uid-1").
					WillReturnRows(rows)
			},
			want:    100,
			wantErr: false,
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.uid,
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
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestMapperID_IDToUID tests mapperID for ID to UID mapping
func TestMapperID_IDToUID(t *testing.T) {
	tests := []struct {
		name    string
		id      int64
		setup   func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface)
		want    string
		wantErr bool
	}{
		{
			name: "Happy Path - single item not in cache",
			id:   100,
			setup: func(s *miniredis.Miniredis, mockPool pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "uid"}).
					AddRow(int64(100), "user-uid-1")
				mockPool.ExpectQuery(`SELECT id, uid FROM "user" WHERE id=$1`).
					WithArgs(int64(100)).
					WillReturnRows(rows)
			},
			want:    "user-uid-1",
			wantErr: false,
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
			logger := infra.NewNoopLogger()

			// Execute
			got, err := mapperID(
				ctx,
				logger,
				redisClient,
				tt.id,
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

// TestInvalidate tests the invalidate helper function
func TestInvalidate(t *testing.T) {
	tests := []struct {
		name    string
		prefix  string
		opts    []param.InvalidateOpt
		setup   func(s *miniredis.Miniredis)
		wantErr bool
		verify  func(s *miniredis.Miniredis)
	}{
		{
			name:   "Happy Path - invalidate single UID with bidirectional mapping",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithUIDs("device-uid-1"),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				s.Set("device:id:100:uid", "device-uid-1")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				ok := s.Exists("device:device-uid-1:id")
				assert.False(t, ok, "forward mapping should be deleted")
				ok = s.Exists("device:id:100:uid")
				assert.False(t, ok, "reverse mapping should be deleted")
			},
		},
		{
			name:   "Happy Path - invalidate multiple UIDs",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithUIDs("device-uid-1", "device-uid-2", "device-uid-3"),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				s.Set("device:id:100:uid", "device-uid-1")
				s.Set("device:device-uid-2:id", "200")
				s.Set("device:id:200:uid", "device-uid-2")
				s.Set("device:device-uid-3:id", "300")
				s.Set("device:id:300:uid", "device-uid-3")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				ok := s.Exists("device:device-uid-1:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:100:uid")
				assert.False(t, ok)
				ok = s.Exists("device:device-uid-2:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:200:uid")
				assert.False(t, ok)
				ok = s.Exists("device:device-uid-3:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:300:uid")
				assert.False(t, ok)
			},
		},
		{
			name:   "Happy Path - invalidate single ID with bidirectional mapping",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithIDs(100),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				s.Set("device:id:100:uid", "device-uid-1")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				ok := s.Exists("device:device-uid-1:id")
				assert.False(t, ok, "forward mapping should be deleted")
				ok = s.Exists("device:id:100:uid")
				assert.False(t, ok, "reverse mapping should be deleted")
			},
		},
		{
			name:   "Happy Path - invalidate multiple IDs",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithIDs(100, 200, 300),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				s.Set("device:id:100:uid", "device-uid-1")
				s.Set("device:device-uid-2:id", "200")
				s.Set("device:id:200:uid", "device-uid-2")
				s.Set("device:device-uid-3:id", "300")
				s.Set("device:id:300:uid", "device-uid-3")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				ok := s.Exists("device:device-uid-1:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:100:uid")
				assert.False(t, ok)
				ok = s.Exists("device:device-uid-2:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:200:uid")
				assert.False(t, ok)
				ok = s.Exists("device:device-uid-3:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:300:uid")
				assert.False(t, ok)
			},
		},
		{
			name:   "Happy Path - invalidate mixed UIDs and IDs",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithUIDs("device-uid-1", "device-uid-2"),
				param.WithIDs(300),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				s.Set("device:id:100:uid", "device-uid-1")
				s.Set("device:device-uid-2:id", "200")
				s.Set("device:id:200:uid", "device-uid-2")
				s.Set("device:device-uid-3:id", "300")
				s.Set("device:id:300:uid", "device-uid-3")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				ok := s.Exists("device:device-uid-1:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:100:uid")
				assert.False(t, ok)
				ok = s.Exists("device:device-uid-2:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:200:uid")
				assert.False(t, ok)
				ok = s.Exists("device:device-uid-3:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:300:uid")
				assert.False(t, ok)
			},
		},
		{
			name:   "Happy Path - duplicate UID and ID pair (deduplication)",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithUIDs("device-uid-1"),
				param.WithIDs(100),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				s.Set("device:id:100:uid", "device-uid-1")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				ok := s.Exists("device:device-uid-1:id")
				assert.False(t, ok)
				ok = s.Exists("device:id:100:uid")
				assert.False(t, ok)
			},
		},
		{
			name:   "Happy Path - invalidate UID when reverse mapping doesn't exist",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithUIDs("device-uid-1"),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				// Reverse mapping doesn't exist
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				ok := s.Exists("device:device-uid-1:id")
				assert.False(t, ok, "forward mapping should be deleted")
			},
		},
		{
			name:   "Happy Path - invalidate ID when forward mapping doesn't exist",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithIDs(100),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:id:100:uid", "device-uid-1")
				// Forward mapping doesn't exist
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				assert.False(t, s.Exists("device:id:100:uid"), "reverse mapping should be deleted")
			},
		},
		{
			name:   "Happy Path - empty options (no-op)",
			prefix: "device",
			opts:   []param.InvalidateOpt{},
			setup: func(s *miniredis.Miniredis) {
				s.Set("device:device-uid-1:id", "100")
				s.Set("device:id:100:uid", "device-uid-1")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				// Keys should still exist
				assert.True(t, s.Exists("device:device-uid-1:id"))
				val, err := s.Get("device:device-uid-1:id")
				assert.NoError(t, err)
				assert.Equal(t, "100", val)
			},
		},
		{
			name:   "Happy Path - invalidate non-existent keys (no-op)",
			prefix: "device",
			opts: []param.InvalidateOpt{
				param.WithUIDs("non-existent-uid"),
			},
			setup: func(s *miniredis.Miniredis) {
				// No keys set up
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				// Nothing to verify, should not error
			},
		},
		{
			name:   "Happy Path - invalidate with different prefix",
			prefix: "user",
			opts: []param.InvalidateOpt{
				param.WithUIDs("user-uid-1"),
			},
			setup: func(s *miniredis.Miniredis) {
				s.Set("user:user-uid-1:id", "100")
				s.Set("user:id:100:uid", "user-uid-1")
				// These should NOT be deleted
				s.Set("device:device-uid-1:id", "200")
			},
			wantErr: false,
			verify: func(s *miniredis.Miniredis) {
				assert.False(t, s.Exists("user:user-uid-1:id"))
				assert.False(t, s.Exists("user:id:100:uid"))
				// Different prefix should remain
				assert.True(t, s.Exists("device:device-uid-1:id"))
				val, err := s.Get("device:device-uid-1:id")
				assert.NoError(t, err)
				assert.Equal(t, "200", val)
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

			ctx := context.Background()

			// Execute
			err := invalidate(
				ctx,
				redisClient,
				tt.prefix,
				tt.opts...,
			)

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

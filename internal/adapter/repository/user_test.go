package repository

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *model.User
		wantErr bool
	}{
		{
			name: "Create valid user",
			user: &model.User{
				UID:      "test-uid-001",
				Username: "testuser1",
				Email:    "test1@example.com",
				Password: "hashedpassword",
				Status:   model.UserStatusActive,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			// Expect the INSERT query with RETURNING id, created_at, updated_at
			// Database handles timestamps via DEFAULT NOW()
			dbTimestamp := time.Now()
			rows := pgxmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(int64(1), dbTimestamp, dbTimestamp)

			mockPool.ExpectQuery(`INSERT INTO "user"`).
				WithArgs(tt.user.UID, tt.user.Username, tt.user.Email, tt.user.Password, tt.user.Status).
				WillReturnRows(rows)

			created, err := repo.Create(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, tt.user.UID, created.UID)
			assert.Equal(t, tt.user.Username, created.Username)
			assert.Equal(t, tt.user.Email, created.Email)
			assert.NotZero(t, created.ID)
			assert.False(t, created.CreatedAt.IsZero())
			assert.False(t, created.UpdatedAt.IsZero())

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(mock pgxmock.PgxPoolIface, id int64)
		wantErr   bool
		wantUID   string
	}{
		{
			name: "Get existing active user by ID",
			id:   1,
			setupMock: func(mock pgxmock.PgxPoolIface, id int64) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(id), "test-uid", "testuser", "test@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE id = \$1 AND deleted_at IS NULL`).
					WithArgs(id).
					WillReturnRows(rows)
			},
			wantErr: false,
			wantUID: "test-uid",
		},
		{
			name:    "Get non-existent user by ID",
			id:      999999999,
			wantErr: true,
			setupMock: func(mock pgxmock.PgxPoolIface, id int64) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE id = \$1 AND deleted_at IS NULL`).
					WithArgs(id).
					WillReturnRows(rows)
			},
		},
		{
			name:    "ID exists but user is soft-deleted - should not find",
			id:      1,
			wantErr: true,
			setupMock: func(mock pgxmock.PgxPoolIface, id int64) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE id = \$1 AND deleted_at IS NULL`).
					WithArgs(id).
					WillReturnRows(rows)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.id)
			}

			result, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.id, result.ID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByUID(t *testing.T) {
	tests := []struct {
		name      string
		uid       string
		setupMock func(mock pgxmock.PgxPoolIface, uid string)
		wantErr   bool
	}{
		{
			name: "Get existing active user by UID",
			uid:  "test-uid-123",
			setupMock: func(mock pgxmock.PgxPoolIface, uid string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), uid, "testuser", "test@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE uid = \$1 AND deleted_at IS NULL`).
					WithArgs(uid).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Get non-existent user by UID",
			uid:  "non-existent-uid",
			setupMock: func(mock pgxmock.PgxPoolIface, uid string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE uid = \$1 AND deleted_at IS NULL`).
					WithArgs(uid).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name: "UID exists but user is soft-deleted - should not find",
			uid:  "deleted-uid-123",
			setupMock: func(mock pgxmock.PgxPoolIface, uid string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE uid = \$1 AND deleted_at IS NULL`).
					WithArgs(uid).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.uid)
			}

			result, err := repo.GetByUID(context.Background(), tt.uid)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.uid, result.UID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		setupMock func(mock pgxmock.PgxPoolIface, email string)
		wantErr   bool
	}{
		{
			name:  "Get existing active user by email",
			email: "test@example.com",
			setupMock: func(mock pgxmock.PgxPoolIface, email string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "test-uid", "testuser", email, "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE email = \$1 AND deleted_at IS NULL`).
					WithArgs(email).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:  "Get non-existent user by email",
			email: "nobody@example.com",
			setupMock: func(mock pgxmock.PgxPoolIface, email string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE email = \$1 AND deleted_at IS NULL`).
					WithArgs(email).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name:  "Email exists but user is soft-deleted - should not find",
			email: "deleted@example.com",
			setupMock: func(mock pgxmock.PgxPoolIface, email string) {
				// Simulate that the only user with this email is soft-deleted
				// The query filters by deleted_at IS NULL, so no rows returned
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE email = \$1 AND deleted_at IS NULL`).
					WithArgs(email).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.email)
			}

			result, err := repo.GetByEmail(context.Background(), tt.email)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.email, result.Email)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	tests := []struct {
		name      string
		username  string
		setupMock func(mock pgxmock.PgxPoolIface, username string)
		wantErr   bool
	}{
		{
			name:     "Get existing active user by username",
			username: "testuser",
			setupMock: func(mock pgxmock.PgxPoolIface, username string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "test-uid", username, "test@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE username = \$1 AND deleted_at IS NULL`).
					WithArgs(username).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:     "Get non-existent user by username",
			username: "nobody",
			setupMock: func(mock pgxmock.PgxPoolIface, username string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE username = \$1 AND deleted_at IS NULL`).
					WithArgs(username).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name:     "Username exists but user is soft-deleted - should not find",
			username: "deleteduser",
			setupMock: func(mock pgxmock.PgxPoolIface, username string) {
				// Simulate that the only user with this username is soft-deleted
				// The query filters by deleted_at IS NULL, so no rows returned
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"})
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE username = \$1 AND deleted_at IS NULL`).
					WithArgs(username).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.username)
			}

			result, err := repo.GetByUsername(context.Background(), tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.username, result.Username)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		user      *model.User
		setupMock func(mock pgxmock.PgxPoolIface, user *model.User)
		wantErr   bool
	}{
		{
			name: "Update username",
			user: &model.User{
				ID:        1,
				UID:       "test-uid",
				Username:  "new_username",
				Email:     "test@example.com",
				Password:  "hash",
				Status:    model.UserStatusActive,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface, user *model.User) {
				mock.ExpectExec(`UPDATE "user"`).
					WithArgs(user.Username, user.Email, user.Password, pgxmock.AnyArg(), pgxmock.AnyArg(), user.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.user)
			}

			err = repo.Update(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		user      *model.User
		setupMock func(mock pgxmock.PgxPoolIface, user *model.User)
		wantErr   bool
	}{
		{
			name: "Delete existing user",
			user: &model.User{
				ID:     1,
				UID:    "test-uid",
				Status: model.UserStatusActive,
			},
			setupMock: func(mock pgxmock.PgxPoolIface, user *model.User) {
				mock.ExpectExec(`UPDATE "user" SET deleted_at = \$1 WHERE id = \$2`).
					WithArgs(pgxmock.AnyArg(), user.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.user)
			}

			err = repo.Delete(context.Background(), tt.user)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.UserListFilterParam
		setupMock  func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "List all users with pagination",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(2))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				// Create rows with proper column type information for status (int32)
				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil).
					AddRow(int64(2), "uid2", "user2", "user2@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "List users with filter by username",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     &param.UserListFilterParam{Username: util.Ptr("testuser")},
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE username = \$1 AND deleted_at IS NULL`).
					WithArgs("testuser").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", *filter.Username, "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE username = \$1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
					WithArgs("testuser", 10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY id DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - uid",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "uid"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY uid DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - username",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "username"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY username DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - email",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "email"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY email DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - status",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "status"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY status DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - created_at",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "created_at"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - updated_at",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "updated_at"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY updated_at DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id; DROP TABLE user; --"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user.*`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "nonexistent"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user.*`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Nil OrderBy - should use default",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: nil},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "user" WHERE deleted_at IS NULL`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "username", "email", "password", "status", "created_at", "updated_at", "deleted_at"}).
					AddRow(int64(1), "uid1", "user1", "user1@example.com", "hash", model.UserStatusActive, time.Now(), time.Now(), nil)
				mock.ExpectQuery(`SELECT .+ FROM "user" WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.pagination, tt.filter)
			}

			result, err := repo.List(context.Background(), tt.pagination, tt.filter)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantCount, len(result.Items))

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserRepository_Create_DuplicateEmail(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewUserRepository(mock)

	user := &model.User{
		UID:      "test-uid",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   1,
	}

	// Mock duplicate email error
	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(user.UID, user.Username, user.Email, user.Password, user.Status).
		WillReturnError(&pgconn.PgError{
			Code:           "23505",
			ConstraintName: "idx_user_email_active",
		})

	_, err = repo.Create(context.Background(), user)
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
		return
	}

	customErr, ok := err.(*errors.CustomError)
	if !ok {
		t.Errorf("Expected CustomError, got %T", err)
		return
	}

	if customErr.Code != errors.ErrDuplicateEmail.Code {
		t.Errorf("Expected code %d, got %d", errors.ErrDuplicateEmail.Code, customErr.Code)
	}
}

func TestUserRepository_Create_DuplicateUsername(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewUserRepository(mock)

	user := &model.User{
		UID:      "test-uid",
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   1,
	}

	// Mock duplicate username error
	mock.ExpectQuery(`INSERT INTO "user"`).
		WithArgs(user.UID, user.Username, user.Email, user.Password, user.Status).
		WillReturnError(&pgconn.PgError{
			Code:           "23505",
			ConstraintName: "idx_user_username_active",
		})

	_, err = repo.Create(context.Background(), user)
	if err == nil {
		t.Error("Expected error for duplicate username, got nil")
		return
	}

	customErr, ok := err.(*errors.CustomError)
	if !ok {
		t.Errorf("Expected CustomError, got %T", err)
		return
	}

	if customErr.Code != errors.ErrDuplicateUsername.Code {
		t.Errorf("Expected code %d, got %d", errors.ErrDuplicateUsername.Code, customErr.Code)
	}
}

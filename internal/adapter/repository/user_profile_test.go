package repository

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		setupMock func(mock pgxmock.PgxPoolIface, userID int64)
		wantErr   bool
	}{
		{
			name:   "Get existing profile by user ID",
			userID: 1,
			setupMock: func(mock pgxmock.PgxPoolIface, userID int64) {
				rows := pgxmock.NewRows([]string{"user_id", "first_name", "last_name", "bio", "avatar_file_id", "attributes", "created_at", "updated_at"}).
					AddRow(userID, "John", "Doe", "Test bio", (*int64)(nil), map[string]any{"key": "value"}, time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_id, first_name, last_name, bio, avatar_file_id, attributes, created_at, updated_at FROM user_profile WHERE user_id = \$1`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:   "Get non-existent profile by user ID",
			userID: 999999999,
			setupMock: func(mock pgxmock.PgxPoolIface, userID int64) {
				rows := pgxmock.NewRows([]string{"user_id", "first_name", "last_name", "bio", "avatar_file_id", "attributes", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT user_id, first_name, last_name, bio, avatar_file_id, attributes, created_at, updated_at FROM user_profile WHERE user_id = \$1`).
					WithArgs(userID).
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

			repo := NewProfileRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.userID)
			}

			result, err := repo.GetByUserID(context.Background(), tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.userID, result.UserID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestProfileRepository_Create(t *testing.T) {
	tests := []struct {
		name     string
		profile  *model.UserProfile
		wantErr  bool
	}{
		{
			name: "Create valid profile",
			profile: &model.UserProfile{
				UserID:       1,
				FirstName:    "John",
				LastName:     "Doe",
				Bio:          "Test bio",
				AvatarFileID: (*int64)(nil),
				Attributes:   map[string]any{"key": "value"},
				CreatedAt:    time.Now().UTC(),
				UpdatedAt:    time.Now().UTC(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewProfileRepository(mockPool)

			mockPool.ExpectExec(`INSERT INTO user_profile`).
				WithArgs(
					tt.profile.UserID, tt.profile.FirstName, tt.profile.LastName, tt.profile.Bio,
					tt.profile.AvatarFileID, tt.profile.Attributes, pgxmock.AnyArg(), pgxmock.AnyArg(),
				).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))

			created, err := repo.Create(context.Background(), tt.profile)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, tt.profile.UserID, created.UserID)
			assert.Equal(t, tt.profile.FirstName, created.FirstName)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestProfileRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		profile   *model.UserProfile
		setupMock func(mock pgxmock.PgxPoolIface, profile *model.UserProfile)
		wantErr   bool
	}{
		{
			name: "Update profile",
			profile: &model.UserProfile{
				UserID:       1,
				FirstName:    "Jane",
				LastName:     "Smith",
				Bio:          "Updated bio",
				AvatarFileID: (*int64)(nil),
				Attributes:   map[string]any{"new": "attributes"},
				UpdatedAt:    time.Now().UTC(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface, profile *model.UserProfile) {
				mock.ExpectExec(`UPDATE user_profile SET first_name = \$1, last_name = \$2, bio = \$3, avatar_file_id = \$4, attributes = \$5, updated_at = \$6 WHERE user_id = \$7`).
					WithArgs(
						profile.FirstName, profile.LastName, profile.Bio,
						profile.AvatarFileID, profile.Attributes, profile.UpdatedAt, profile.UserID,
					).
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

			repo := NewProfileRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.profile)
			}

			err = repo.Update(context.Background(), tt.profile)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestProfileRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		profile   *model.UserProfile
		setupMock func(mock pgxmock.PgxPoolIface, profile *model.UserProfile)
		wantErr   bool
	}{
		{
			name: "Delete existing profile",
			profile: &model.UserProfile{
				UserID: 1,
			},
			setupMock: func(mock pgxmock.PgxPoolIface, profile *model.UserProfile) {
				mock.ExpectExec(`DELETE FROM user_profile WHERE user_id = \$1`).
					WithArgs(profile.UserID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewProfileRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.profile)
			}

			err = repo.Delete(context.Background(), tt.profile)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestProfileRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *params.PaginationParam
		filter     *params.UserProfileListFilterParam
		setupMock  func(mock pgxmock.PgxPoolIface, pagination *params.PaginationParam, filter *params.UserProfileListFilterParam)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "List all profiles with pagination",
			pagination: &params.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *params.PaginationParam, filter *params.UserProfileListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(2))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_profile`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "first_name", "last_name", "bio", "avatar_file_id", "attributes", "created_at", "updated_at"}).
					AddRow(int64(1), "John", "Doe", "Bio 1", (*int64)(nil), map[string]any{}, time.Now(), time.Now()).
					AddRow(int64(2), "Jane", "Smith", "Bio 2", (*int64)(nil), map[string]any{}, time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_id, first_name, last_name, bio, avatar_file_id, attributes, created_at, updated_at FROM user_profile`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewProfileRepository(mockPool)

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

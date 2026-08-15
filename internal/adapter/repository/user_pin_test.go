package repository

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPinRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		setupMock func(mock pgxmock.PgxPoolIface, userID int64)
		wantErr   bool
	}{
		{
			name:   "Get existing PIN by user ID",
			userID: 1,
			setupMock: func(mock pgxmock.PgxPoolIface, userID int64) {
				rows := pgxmock.NewRows([]string{"user_id", "uid", "code", "created_at", "updated_at"}).
					AddRow(userID, "user-uid", "hashedpin", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_pin.user_id, "user".uid, user_pin.code, user_pin.created_at, user_pin.updated_at FROM user_pin JOIN "user" ON "user".id = user_pin.user_id WHERE user_pin.user_id = \$1`).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:   "Get non-existent PIN by user ID",
			userID: 999999999,
			setupMock: func(mock pgxmock.PgxPoolIface, userID int64) {
				rows := pgxmock.NewRows([]string{"user_id", "uid", "code", "created_at", "updated_at"})
				mock.ExpectQuery(`SELECT user_pin.user_id, "user".uid, user_pin.code, user_pin.created_at, user_pin.updated_at FROM user_pin JOIN "user" ON "user".id = user_pin.user_id WHERE user_pin.user_id = \$1`).
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

			repo := NewPinRepository(mockPool, infra.NewNoopTracer(), nil)

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
			assert.Equal(t, "user-uid", result.UserUID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestPinRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		pin     *model.UserPin
		wantErr bool
	}{
		{
			name: "Create valid PIN",
			pin: &model.UserPin{
				UserID:    1,
				Code:      "hashedpin",
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewPinRepository(mockPool, infra.NewNoopTracer(), nil)

			mockPool.ExpectExec(`INSERT INTO user_pin`).
				WithArgs(tt.pin.UserID, tt.pin.Code, pgxmock.AnyArg(), pgxmock.AnyArg()).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))

			created, err := repo.Create(context.Background(), tt.pin)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, tt.pin.UserID, created.UserID)
			assert.Equal(t, tt.pin.Code, created.Code)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestPinRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		pin       *model.UserPin
		setupMock func(mock pgxmock.PgxPoolIface, pin *model.UserPin)
		wantErr   bool
	}{
		{
			name: "Update PIN code",
			pin: &model.UserPin{
				UserID:    1,
				Code:      "newhashedpin",
				UpdatedAt: time.Now().UTC(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface, pin *model.UserPin) {
				mock.ExpectExec(`UPDATE user_pin SET code = \$1, updated_at = \$2 WHERE user_id = \$3`).
					WithArgs(pin.Code, pin.UpdatedAt, pin.UserID).
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

			repo := NewPinRepository(mockPool, infra.NewNoopTracer(), nil)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.pin)
			}

			err = repo.Update(context.Background(), tt.pin)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestPinRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		pin       *model.UserPin
		setupMock func(mock pgxmock.PgxPoolIface, pin *model.UserPin)
		wantErr   bool
	}{
		{
			name: "Delete existing PIN",
			pin: &model.UserPin{
				UserID: 1,
			},
			setupMock: func(mock pgxmock.PgxPoolIface, pin *model.UserPin) {
				mock.ExpectExec(`DELETE FROM user_pin WHERE user_id = \$1`).
					WithArgs(pin.UserID).
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

			repo := NewPinRepository(mockPool, infra.NewNoopTracer(), nil)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.pin)
			}

			err = repo.Delete(context.Background(), tt.pin)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestPinRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.UserPinListFilterParam
		setupMock  func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "List all PINs with pagination",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(2))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_pin`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "code", "created_at", "updated_at"}).
					AddRow(int64(1), "hashedpin1", time.Now(), time.Now()).
					AddRow(int64(2), "hashedpin2", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_id, code, created_at, updated_at FROM user_pin`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - user_id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "user_id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_pin`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "code", "created_at", "updated_at"}).
					AddRow(int64(1), "hashedpin1", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_id, code, created_at, updated_at FROM user_pin`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - created_at",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "created_at"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_pin`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "code", "created_at", "updated_at"}).
					AddRow(int64(1), "hashedpin1", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_id, code, created_at, updated_at FROM user_pin`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - updated_at",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "updated_at"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_pin`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "code", "created_at", "updated_at"}).
					AddRow(int64(1), "hashedpin1", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_id, code, created_at, updated_at FROM user_pin`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "user_id; DROP TABLE user_pin; --"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_pin`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "invalid_column"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_pin`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Nil OrderBy - should use default",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: nil},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_pin`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "code", "created_at", "updated_at"}).
					AddRow(int64(1), "hashedpin1", time.Now(), time.Now())
				mock.ExpectQuery(`SELECT user_id, code, created_at, updated_at FROM user_pin`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
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

			repo := NewPinRepository(mockPool, infra.NewNoopTracer(), nil)

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

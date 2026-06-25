package repository

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserDeviceRepository_GetByUserIDAndDeviceID(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		deviceID  int64
		setupMock func(mock pgxmock.PgxPoolIface, userID, deviceID int64)
		wantErr   bool
	}{
		{
			name:     "Get existing user-device relationship",
			userID:   1,
			deviceID: 2,
			setupMock: func(mock pgxmock.PgxPoolIface, userID, deviceID int64) {
				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(userID, deviceID, "192.168.1.1", time.Now(), "session-123", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device WHERE user_id = \$1 AND device_id = \$2`).
					WithArgs(userID, deviceID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:     "Get non-existent user-device relationship",
			userID:   999999999,
			deviceID: 888888888,
			setupMock: func(mock pgxmock.PgxPoolIface, userID, deviceID int64) {
				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"})
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device WHERE user_id = \$1 AND device_id = \$2`).
					WithArgs(userID, deviceID).
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

			repo := NewUserDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.userID, tt.deviceID)
			}

			result, err := repo.GetByUserIDAndDeviceID(context.Background(), tt.userID, tt.deviceID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.userID, result.UserID)
			assert.Equal(t, tt.deviceID, result.DeviceID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserDeviceRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		userDev   *model.UserDevice
		wantErr   bool
	}{
		{
			name: "Create valid user-device relationship",
			userDev: &model.UserDevice{
				UserID:      1,
				DeviceID:    2,
				IPAddress:   "192.168.1.1",
				LastActiveAt: time.Now().UTC(),
				SessionID:   "test-session-123",
				CreatedAt:   time.Now().UTC(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewUserDeviceRepository(mockPool)

			mockPool.ExpectExec(`INSERT INTO user_device`).
				WithArgs(
					tt.userDev.UserID, tt.userDev.DeviceID, tt.userDev.IPAddress,
					tt.userDev.LastActiveAt, tt.userDev.SessionID, pgxmock.AnyArg(),
				).
				WillReturnResult(pgxmock.NewResult("INSERT", 1))

			created, err := repo.Create(context.Background(), tt.userDev)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, tt.userDev.UserID, created.UserID)
			assert.Equal(t, tt.userDev.DeviceID, created.DeviceID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserDeviceRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		userDev   *model.UserDevice
		setupMock func(mock pgxmock.PgxPoolIface, userDev *model.UserDevice)
		wantErr   bool
	}{
		{
			name: "Update user-device relationship",
			userDev: &model.UserDevice{
				UserID:      1,
				DeviceID:    2,
				IPAddress:   "10.0.0.1",
				LastActiveAt: time.Now().UTC(),
			},
			setupMock: func(mock pgxmock.PgxPoolIface, userDev *model.UserDevice) {
				mock.ExpectExec(`UPDATE user_device SET ip_address = \$1, last_active_at = \$2 WHERE user_id = \$3 AND device_id = \$4`).
					WithArgs(userDev.IPAddress, userDev.LastActiveAt, userDev.UserID, userDev.DeviceID).
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

			repo := NewUserDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.userDev)
			}

			err = repo.Update(context.Background(), tt.userDev)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserDeviceRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		userDev   *model.UserDevice
		setupMock func(mock pgxmock.PgxPoolIface, userDev *model.UserDevice)
		wantErr   bool
	}{
		{
			name: "Delete user-device relationship",
			userDev: &model.UserDevice{
				UserID:   1,
				DeviceID: 2,
			},
			setupMock: func(mock pgxmock.PgxPoolIface, userDev *model.UserDevice) {
				mock.ExpectExec(`DELETE FROM user_device WHERE user_id = \$1 AND device_id = \$2`).
					WithArgs(userDev.UserID, userDev.DeviceID).
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

			repo := NewUserDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.userDev)
			}

			err = repo.Delete(context.Background(), tt.userDev)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserDeviceRepository_Revoke(t *testing.T) {
	tests := []struct {
		name      string
		userID    int64
		deviceID  int64
		setupMock func(mock pgxmock.PgxPoolIface, userID, deviceID int64)
		wantErr   bool
	}{
		{
			name:     "Revoke user-device relationship",
			userID:   1,
			deviceID: 2,
			setupMock: func(mock pgxmock.PgxPoolIface, userID, deviceID int64) {
				mock.ExpectExec(`UPDATE user_device SET revoked_at = \$1 WHERE user_id = \$2 AND device_id = \$3`).
					WithArgs(pgxmock.AnyArg(), userID, deviceID).
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

			repo := NewUserDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.userID, tt.deviceID)
			}

			err = repo.Revoke(context.Background(), tt.userID, tt.deviceID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestUserDeviceRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.UserDeviceListFilterParam
		setupMock  func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "List all user-device relationships with pagination",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(2))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(int64(1), int64(2), "192.168.1.1", time.Now(), "session-1", (*time.Time)(nil), time.Now()).
					AddRow(int64(3), int64(4), "10.0.0.1", time.Now(), "session-2", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(int64(1), int64(2), "192.168.1.1", time.Now(), "session-1", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - user_id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "user_id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(int64(1), int64(2), "192.168.1.1", time.Now(), "session-1", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - device_id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "device_id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(int64(1), int64(2), "192.168.1.1", time.Now(), "session-1", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - last_active_at",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "last_active_at"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(int64(1), int64(2), "192.168.1.1", time.Now(), "session-1", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device`).
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
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(int64(1), int64(2), "192.168.1.1", time.Now(), "session-1", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device`).
					WithArgs(pgxmock.AnyArg(), pgxmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id; DROP TABLE user_device; --"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).WillReturnRows(countRows)
			},
			wantCount:  0,
			wantErr:    true,
		},
		{
			name:       "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "nonexistent"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).WillReturnRows(countRows)
			},
			wantCount:  0,
			wantErr:    true,
		},
		{
			name:       "Nil OrderBy - should use default",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: nil},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(int64(1))
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM user_device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"user_id", "device_id", "ip_address", "last_active_at", "session_id", "revoked_at", "created_at"}).
					AddRow(int64(1), int64(2), "192.168.1.1", time.Now(), "session-1", (*time.Time)(nil), time.Now())
				mock.ExpectQuery(`SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at FROM user_device`).
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

			repo := NewUserDeviceRepository(mockPool)

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

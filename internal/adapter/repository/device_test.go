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

func TestDeviceRepository_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		setupMock func(mock pgxmock.PgxPoolIface, id int64)
		wantErr   bool
	}{
		{
			name: "Get existing device by ID",
			id:   1,
			setupMock: func(mock pgxmock.PgxPoolIface, id int64) {
				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(id, "device-uid-001", "fp-abc123", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE id = \$1`).
					WithArgs(id).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Get non-existent device by ID",
			id:   999999999,
			setupMock: func(mock pgxmock.PgxPoolIface, id int64) {
				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"})
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE id = \$1`).
					WithArgs(id).
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

			repo := NewDeviceRepository(mockPool)

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

func TestDeviceRepository_GetByUID(t *testing.T) {
	tests := []struct {
		name      string
		uid       string
		setupMock func(mock pgxmock.PgxPoolIface, uid string)
		wantErr   bool
	}{
		{
			name: "Get existing device by UID",
			uid:  "device-uid-123",
			setupMock: func(mock pgxmock.PgxPoolIface, uid string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), uid, "fp-abc123", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE uid = \$1`).
					WithArgs(uid).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Get non-existent device by UID",
			uid:  "non-existent-uid",
			setupMock: func(mock pgxmock.PgxPoolIface, uid string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"})
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE uid = \$1`).
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

			repo := NewDeviceRepository(mockPool)

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

func TestDeviceRepository_GetByFingerprint(t *testing.T) {
	tests := []struct {
		name        string
		fingerprint string
		setupMock   func(mock pgxmock.PgxPoolIface, fingerprint string)
		wantErr     bool
	}{
		{
			name:        "Get existing device by fingerprint",
			fingerprint: "fp-abc123456",
			setupMock: func(mock pgxmock.PgxPoolIface, fingerprint string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "device-uid-001", fingerprint, "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE device_fingerprint = \$1`).
					WithArgs(fingerprint).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name:        "Get non-existent device by fingerprint",
			fingerprint: "fp-nonexistent",
			setupMock: func(mock pgxmock.PgxPoolIface, fingerprint string) {
				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"})
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE device_fingerprint = \$1`).
					WithArgs(fingerprint).
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

			repo := NewDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.fingerprint)
			}

			result, err := repo.GetByFingerprint(context.Background(), tt.fingerprint)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.fingerprint, result.DeviceFingerprint)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestDeviceRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		device  *model.Device
		wantErr bool
	}{
		{
			name: "Create valid device",
			device: &model.Device{
				UID:               "device-uid-001",
				DeviceFingerprint: "fp-abc123",
				DeviceName:        "iPhone 14",
				CreatedAt:         time.Now().UTC(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewDeviceRepository(mockPool)

			rows := pgxmock.NewRows([]string{"id"}).AddRow(int64(1))
			mockPool.ExpectQuery(`INSERT INTO device \(uid, device_fingerprint, device_name, created_at\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id`).
				WithArgs(tt.device.UID, tt.device.DeviceFingerprint, tt.device.DeviceName, pgxmock.AnyArg()).
				WillReturnRows(rows)

			created, err := repo.Create(context.Background(), tt.device)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, created)
			assert.Equal(t, tt.device.UID, created.UID)
			assert.Equal(t, tt.device.DeviceFingerprint, created.DeviceFingerprint)
			assert.NotZero(t, created.ID)

			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestDeviceRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		device    *model.Device
		setupMock func(mock pgxmock.PgxPoolIface, device *model.Device)
		wantErr   bool
	}{
		{
			name: "Update device name",
			device: &model.Device{
				ID:                1,
				UID:               "device-uid-001",
				DeviceFingerprint: "fp-abc123",
				DeviceName:        "iPhone 15 Pro",
			},
			setupMock: func(mock pgxmock.PgxPoolIface, device *model.Device) {
				mock.ExpectExec(`UPDATE device SET device_name = \$1 WHERE id = \$2`).
					WithArgs(device.DeviceName, device.ID).
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

			repo := NewDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.device)
			}

			err = repo.Update(context.Background(), tt.device)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestDeviceRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		device    *model.Device
		setupMock func(mock pgxmock.PgxPoolIface, device *model.Device)
		wantErr   bool
	}{
		{
			name: "Delete existing device",
			device: &model.Device{
				ID: 1,
			},
			setupMock: func(mock pgxmock.PgxPoolIface, device *model.Device) {
				mock.ExpectExec(`DELETE FROM device WHERE id = \$1`).
					WithArgs(device.ID).
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

			repo := NewDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.device)
			}

			err = repo.Delete(context.Background(), tt.device)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mockPool.ExpectationsWereMet())
		})
	}
}

func TestDeviceRepository_List(t *testing.T) {
	tests := []struct {
		name       string
		pagination *param.PaginationParam
		filter     *param.DeviceListFilterParam
		setupMock  func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "List all devices with pagination",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now()).
					AddRow(int64(2), "uid2", "fp2", "Samsung Galaxy", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "List devices with filter by device name",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     &param.DeviceListFilterParam{DeviceName: util.Ptr("iPhone 14")},
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device WHERE device_name = \$1`).
					WithArgs("iPhone 14").
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE device_name = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
					WithArgs("iPhone 14", 10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "List devices with filter by UIDs",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     &param.DeviceListFilterParam{Uids: []string{"uid1", "uid2"}},
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device WHERE uid = ANY\(\$1\)`).
					WithArgs(filter.Uids).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now()).
					AddRow(int64(2), "uid2", "fp2", "Samsung Galaxy", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE uid = ANY\(\$1\) ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
					WithArgs(filter.Uids, 10, 0).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - id",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device ORDER BY id DESC LIMIT \$1 OFFSET \$2`).
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
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device ORDER BY uid DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - device_fingerprint",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "device_fingerprint"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device ORDER BY device_fingerprint DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Valid OrderBy - device_name",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "device_name"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device ORDER BY device_name DESC LIMIT \$1 OFFSET \$2`).
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
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:       "Invalid OrderBy - SQL injection attempt",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "id; DROP TABLE device; --"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Invalid OrderBy - non-existent column",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "nonexistent"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
		{
			name:       "Nil OrderBy - should use default",
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: nil},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(1)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device`).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now())
				mock.ExpectQuery(`SELECT id, uid, device_fingerprint, device_name, created_at FROM device ORDER BY created_at DESC LIMIT \$1 OFFSET \$2`).
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

			repo := NewDeviceRepository(mockPool)

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

func TestDeviceRepository_ListByUserID(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		pagination *param.PaginationParam
		filter     *param.DeviceListFilterParam
		setupMock  func(mock pgxmock.PgxPoolIface, userID int64, pagination *param.PaginationParam, filter *param.DeviceListFilterParam)
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "List devices by user ID",
			userID:     1,
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1)},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, userID int64, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(2)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device d JOIN user_device ud ON d\.id = ud\.device_id WHERE ud\.user_id = \$1 AND ud\.revoked_at IS NULL`).
					WithArgs(userID).
					WillReturnRows(countRows)

				rows := pgxmock.NewRows([]string{"id", "uid", "device_fingerprint", "device_name", "created_at"}).
					AddRow(int64(1), "uid1", "fp1", "iPhone 14", time.Now()).
					AddRow(int64(2), "uid2", "fp2", "Samsung Galaxy", time.Now())
				mock.ExpectQuery(`SELECT d\.id, d\.uid, d\.device_fingerprint, d\.device_name, d\.created_at FROM device d JOIN user_device ud ON d\.id = ud\.device_id WHERE ud\.user_id = \$1 AND ud\.revoked_at IS NULL ORDER BY d\.created_at DESC LIMIT \$2 OFFSET \$3`).
					WithArgs(userID, 10, 0).
					WillReturnRows(rows)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name:       "Invalid OrderBy - non-existent column",
			userID:     1,
			pagination: &param.PaginationParam{Limit: util.Ptr(10), Page: util.Ptr(1), OrderBy: func() *string { s := "nonexistent"; return &s }()},
			filter:     nil,
			setupMock: func(mock pgxmock.PgxPoolIface, userID int64, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) {
				countRows := pgxmock.NewRows([]string{"count"}).AddRow(0)
				mock.ExpectQuery(`SELECT COUNT\(\*\) FROM device d JOIN user_device ud ON d\.id = ud\.device_id WHERE ud\.user_id = \$1 AND ud\.revoked_at IS NULL`).
					WithArgs(userID).
					WillReturnRows(countRows)
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool, err := pgxmock.NewPool()
			require.NoError(t, err)
			defer mockPool.Close()

			repo := NewDeviceRepository(mockPool)

			if tt.setupMock != nil {
				tt.setupMock(mockPool, tt.userID, tt.pagination, tt.filter)
			}

			result, err := repo.ListByUserID(context.Background(), tt.userID, tt.pagination, tt.filter)

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

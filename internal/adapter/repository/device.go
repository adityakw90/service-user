package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)

// deviceModel is the database model for device data.
type deviceModel struct {
	ID                int64
	UID               string
	DeviceFingerprint string
	DeviceName        string
	CreatedAt         time.Time
}

// toDomain converts a device model to a domain entity.
func (m *deviceModel) toDomain() *model.Device {
	return &model.Device{
		ID:                m.ID,
		UID:               m.UID,
		DeviceFingerprint: m.DeviceFingerprint,
		DeviceName:        m.DeviceName,
		CreatedAt:         m.CreatedAt,
	}
}

// DeviceRepository implements repository.DeviceRepository for PostgreSQL.
type DeviceRepository struct {
	db PostgrePool
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db PostgrePool) repository.DeviceRepository {
	return &DeviceRepository{db: db}
}

// GetByID retrieves a device by internal ID.
func (r *DeviceRepository) GetByID(ctx context.Context, id int64) (*model.Device, error) {
	query := `SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE id = $1`
	return r.scanDevice(r.db.QueryRow(ctx, query, id))
}

// GetByUID retrieves a device by public UID.
func (r *DeviceRepository) GetByUID(ctx context.Context, uid string) (*model.Device, error) {
	query := `SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE uid = $1`
	return r.scanDevice(r.db.QueryRow(ctx, query, uid))
}

// GetByFingerprint retrieves a device by fingerprint.
func (r *DeviceRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*model.Device, error) {
	query := `SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE device_fingerprint = $1`
	return r.scanDevice(r.db.QueryRow(ctx, query, fingerprint))
}

// Create adds a new device.
func (r *DeviceRepository) Create(ctx context.Context, device *model.Device) (*model.Device, error) {
	query := `INSERT INTO device (uid, device_fingerprint, device_name, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(ctx, query,
		device.UID, device.DeviceFingerprint, device.DeviceName, device.CreatedAt,
	).Scan(&device.ID)
	return device, err
}

// Update modifies an existing device.
func (r *DeviceRepository) Update(ctx context.Context, device *model.Device) error {
	query := `UPDATE device SET device_name = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, device.DeviceName, device.ID)
	return err
}

// Delete removes a device.
func (r *DeviceRepository) Delete(ctx context.Context, device *model.Device) error {
	query := `DELETE FROM device WHERE id = $1`
	_, err := r.db.Exec(ctx, query, device.ID)
	return err
}

// List retrieves all devices with pagination and filtering.
func (r *DeviceRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	limit := 10
	offset := 0
	page := 1
	if pagination != nil {
		if pagination.Limit != nil {
			limit = *pagination.Limit
		}
		if pagination.Page != nil {
			page = *pagination.Page
			offset = (page - 1) * limit
		}
	}
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter != nil {
		if len(filter.Uids) > 0 {
			conditions = append(conditions, fmt.Sprintf("uid = ANY($%d)", argIdx))
			args = append(args, filter.Uids)
			argIdx++
		}
		if filter.DeviceName != nil {
			conditions = append(conditions, fmt.Sprintf("device_name = $%d", argIdx))
			args = append(args, *filter.DeviceName)
			argIdx++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM device %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, uid, device_fingerprint, device_name, created_at
		FROM device
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices, err := r.scanRows(rows)
	if err != nil {
		return nil, err
	}

	// Convert []*Device to []Device
	deviceItems := make([]model.Device, len(devices))
	for i, d := range devices {
		deviceItems[i] = *d
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &model.Devices{
		Items: deviceItems,
		Meta: model.Meta{
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: totalPages,
		},
	}, nil
}

// ListByUserID lists all devices for a user.
func (r *DeviceRepository) ListByUserID(ctx context.Context, userID int64, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	limit := 10
	offset := 0
	page := 1
	if pagination != nil {
		if pagination.Limit != nil {
			limit = *pagination.Limit
		}
		if pagination.Page != nil {
			page = *pagination.Page
			offset = (page - 1) * limit
		}
	}
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("ud.user_id = $%d", argIdx))
	args = append(args, userID)
	argIdx++

	if filter != nil {
		if len(filter.Uids) > 0 {
			conditions = append(conditions, fmt.Sprintf("d.uid = ANY($%d)", argIdx))
			args = append(args, filter.Uids)
			argIdx++
		}
		if filter.DeviceName != nil {
			conditions = append(conditions, fmt.Sprintf("d.device_name = $%d", argIdx))
			args = append(args, *filter.DeviceName)
			argIdx++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM device d
		JOIN user_device ud ON d.id = ud.device_id
		%s
	`, whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Get paginated results - select only device columns
	query := fmt.Sprintf(`
		SELECT d.id, d.uid, d.device_fingerprint, d.device_name, d.created_at
		FROM device d
		JOIN user_device ud ON d.id = ud.device_id
		%s
		ORDER BY d.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices, err := r.scanRows(rows)
	if err != nil {
		return nil, err
	}

	// Convert []*Device to []Device
	deviceItems := make([]model.Device, len(devices))
	for i, d := range devices {
		deviceItems[i] = *d
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &model.Devices{
		Items: deviceItems,
		Meta: model.Meta{
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: totalPages,
		},
	}, nil
}

func (r *DeviceRepository) scanDevice(row pgx.Row) (*model.Device, error) {
	var m deviceModel
	err := row.Scan(
		&m.ID, &m.UID, &m.DeviceFingerprint, &m.DeviceName, &m.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrDeviceNotFound
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

func (r *DeviceRepository) scanRows(rows pgx.Rows) ([]*model.Device, error) {
	var devices []*model.Device
	for rows.Next() {
		var m deviceModel
		err := rows.Scan(
			&m.ID, &m.UID, &m.DeviceFingerprint, &m.DeviceName, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, m.toDomain())
	}
	return devices, nil
}

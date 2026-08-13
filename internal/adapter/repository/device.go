package repository

import (
	"context"
	"fmt"
	"strings"

	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)


// allowedOrderByDevice maps OrderBy string values to their typed enum for validation.
var allowedOrderByDevice = map[string]param.DeviceOrderBy{
	"id":                 param.OrderByDeviceID,
	"uid":                param.OrderByDeviceUID,
	"device_fingerprint": param.OrderByDeviceFingerprint,
	"device_name":        param.OrderByDeviceName,
	"created_at":         param.OrderByDeviceCreatedAt,
}

// DeviceRepository implements repository.DeviceRepository for PostgreSQL.
type DeviceRepository struct {
	db     PostgrePool
	tracer gomon.Tracer
	logger gomon.Logger
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db PostgrePool, tracer gomon.Tracer, logger gomon.Logger) repository.DeviceRepository {
	if db == nil {
		panic("db is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &DeviceRepository{
		db:     db,
		tracer: tracer,
		logger: logger,
	}
}

// GetByID retrieves a device by internal ID.
func (r *DeviceRepository) GetByID(ctx context.Context, id int64) (*model.Device, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.GetByID")
	defer span.End()

	query := `SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE id = $1`
	return r.scanDevice(r.db.QueryRow(newCtx, query, id))
}

// GetByUID retrieves a device by public UID.
func (r *DeviceRepository) GetByUID(ctx context.Context, uid string) (*model.Device, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.GetByUID")
	defer span.End()

	query := `SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE uid = $1`
	return r.scanDevice(r.db.QueryRow(newCtx, query, uid))
}

// GetByFingerprint retrieves a device by fingerprint.
func (r *DeviceRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*model.Device, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.GetByFingerprint")
	defer span.End()

	query := `SELECT id, uid, device_fingerprint, device_name, created_at FROM device WHERE device_fingerprint = $1`
	return r.scanDevice(r.db.QueryRow(newCtx, query, fingerprint))
}

// Create adds a new device.
func (r *DeviceRepository) Create(ctx context.Context, device *model.Device) (*model.Device, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.Create")
	defer span.End()

	query := `INSERT INTO device (uid, device_fingerprint, device_name, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(newCtx, query,
		device.UID, device.DeviceFingerprint, device.DeviceName, device.CreatedAt,
	).Scan(&device.ID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to create device", map[string]any{"error": err, "uid": device.UID})
	}
	return device, err
}

// Update modifies an existing device.
func (r *DeviceRepository) Update(ctx context.Context, device *model.Device) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.Update")
	defer span.End()

	query := `UPDATE device SET device_name = $1 WHERE id = $2`
	_, err := r.db.Exec(newCtx, query, device.DeviceName, device.ID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to update device", map[string]any{"error": err, "id": device.ID})
	}
	return err
}

// Delete removes a device.
func (r *DeviceRepository) Delete(ctx context.Context, device *model.Device) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.Delete")
	defer span.End()

	query := `DELETE FROM device WHERE id = $1`
	_, err := r.db.Exec(newCtx, query, device.ID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to delete device", map[string]any{"error": err, "id": device.ID})
	}
	return err
}

// List retrieves all devices with pagination and filtering.
func (r *DeviceRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) (*model.Devices, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.List")
	defer span.End()

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
	if err := r.db.QueryRow(newCtx, countQuery, args...).Scan(&total); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to count devices", map[string]any{"error": err})
		}
		return nil, err
	}

	// Get paginated results
	// Apply sorting
	orderByValue, err := validateOrderBy(pagination, "created_at", allowedOrderByDevice)
	if err != nil {
		return nil, err
	}

	// Build ORDER BY clause
	orderByClause := orderByValue
	if pagination != nil && pagination.Sort != nil && *pagination.Sort != "" {
		orderByClause += " " + *pagination.Sort
	} else {
		orderByClause += " DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, uid, device_fingerprint, device_name, created_at
		FROM device
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(newCtx, query, args...)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to list devices", map[string]any{"error": err})
		}
		return nil, err
	}
	defer rows.Close()

	devices, err := r.scanRows(rows)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to scan devices", map[string]any{"error": err})
		}
		return nil, err
	}

	// Convert []*Device to []Device
	deviceItems := make([]model.Device, len(devices))
	for i, d := range devices {
		deviceItems[i] = *d
	}

	return &model.Devices{
		Items: deviceItems,
		Meta:  buildMeta(total, page, limit),
	}, nil
}

// ListByUserID lists all devices for a user.
func (r *DeviceRepository) ListByUserID(ctx context.Context, userID int64, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) (*model.Devices, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.Device.ListByUserID")
	defer span.End()

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

	// Exclude revoked devices
	conditions = append(conditions, "ud.revoked_at IS NULL")

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
	if err := r.db.QueryRow(newCtx, countQuery, args...).Scan(&total); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to count user devices", map[string]any{"error": err, "userID": userID})
		}
		return nil, err
	}

	// Get paginated results - select only device columns
	// Apply sorting
	orderByValue, err := validateOrderBy(pagination, "created_at", allowedOrderByDevice)
	if err != nil {
		return nil, err
	}

	// Build ORDER BY clause with table alias
	orderByClause := fmt.Sprintf("d.%s", orderByValue)

	// Add sort direction if provided and not empty
	if pagination != nil && pagination.Sort != nil && *pagination.Sort != "" {
		orderByClause += " " + *pagination.Sort
	} else {
		orderByClause += " DESC"
	}

	query := fmt.Sprintf(`
		SELECT d.id, d.uid, d.device_fingerprint, d.device_name, d.created_at
		FROM device d
		JOIN user_device ud ON d.id = ud.device_id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(newCtx, query, args...)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to list user devices", map[string]any{"error": err, "userID": userID})
		}
		return nil, err
	}
	defer rows.Close()

	devices, err := r.scanRows(rows)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to scan user devices", map[string]any{"error": err, "userID": userID})
		}
		return nil, err
	}

	// Convert []*Device to []Device
	deviceItems := make([]model.Device, len(devices))
	for i, d := range devices {
		deviceItems[i] = *d
	}

	return &model.Devices{
		Items: deviceItems,
		Meta:  buildMeta(total, page, limit),
	}, nil
}

func (r *DeviceRepository) scanDevice(row pgx.Row) (*model.Device, error) {
	var m model.Device
	err := row.Scan(
		&m.ID, &m.UID, &m.DeviceFingerprint, &m.DeviceName, &m.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrDeviceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *DeviceRepository) scanRows(rows pgx.Rows) ([]*model.Device, error) {
	var devices []*model.Device
	for rows.Next() {
		var m model.Device
		err := rows.Scan(
			&m.ID, &m.UID, &m.DeviceFingerprint, &m.DeviceName, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, &m)
	}
	return devices, nil
}

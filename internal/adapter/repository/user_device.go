package repository

import (
	"context"
	"fmt"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)

// allowedOrderByUserDevice maps OrderBy string values to their typed enum for validation.
var allowedOrderByUserDevice = map[string]param.UserDeviceOrderBy{
	"id":             param.OrderByUserDeviceID,
	"user_id":        param.OrderByUserDeviceUserID,
	"device_id":      param.OrderByUserDeviceDeviceID,
	"last_active_at": param.OrderByUserDeviceLastActiveAt,
	"created_at":     param.OrderByUserDeviceCreatedAt,
}

// UserDeviceRepository implements repository.UserDeviceRepository for PostgreSQL.
type UserDeviceRepository struct {
	db     PostgrePool
	tracer gomon.Tracer
	logger gomon.Logger
}

// NewUserDeviceRepository creates a new UserDeviceRepository.
func NewUserDeviceRepository(db PostgrePool, tracer gomon.Tracer, logger gomon.Logger) repository.UserDeviceRepository {
	if db == nil {
		panic("db is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &UserDeviceRepository{
		db:     db,
		tracer: tracer,
		logger: logger,
	}
}

// GetByUserIDAndDeviceID finds a user-device relationship.
func (r *UserDeviceRepository) GetByUserIDAndDeviceID(ctx context.Context, userID, deviceID int64) (*model.UserDevice, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserDevice.GetByUserIDAndDeviceID")
	defer span.End()

	query := `
		SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at
		FROM user_device
		WHERE user_id = $1 AND device_id = $2
	`
	ud, err := r.scanUserDevice(r.db.QueryRow(newCtx, query, userID, deviceID))
	if err != nil && err != errors.ErrUserDeviceNotFound && r.logger != nil {
		r.logger.Error("failed to get user device", map[string]any{"error": err})
	}
	return ud, err
}

// Create adds a new user-device relationship.
func (r *UserDeviceRepository) Create(ctx context.Context, ud *model.UserDevice) (*model.UserDevice, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserDevice.Create")
	defer span.End()

	query := `
		INSERT INTO user_device (user_id, device_id, ip_address, last_active_at, session_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	// Handle empty string for ip_address - use 0.0.0.0 as default (inet type doesn't accept empty strings)
	ipAddress := ud.IPAddress
	if ipAddress == "" {
		ipAddress = "0.0.0.0"
	}
	_, err := r.db.Exec(newCtx, query,
		ud.UserID, ud.DeviceID, ipAddress, ud.LastActiveAt, ud.SessionID, ud.CreatedAt,
	)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to create user device", map[string]any{"error": err, "userID": ud.UserID, "deviceID": ud.DeviceID})
	}
	return ud, err
}

// Update modifies an existing relationship.
func (r *UserDeviceRepository) Update(ctx context.Context, ud *model.UserDevice) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserDevice.Update")
	defer span.End()

	query := `
		UPDATE user_device
		SET ip_address = $1, last_active_at = $2
		WHERE user_id = $3 AND device_id = $4
	`
	// Handle empty string for ip_address - keep existing value if empty
	ipAddress := ud.IPAddress
	if ipAddress == "" {
		// Don't update ip_address if empty, keep existing value
		query = `
			UPDATE user_device
			SET last_active_at = $1
			WHERE user_id = $2 AND device_id = $3
		`
		_, err := r.db.Exec(newCtx, query, ud.LastActiveAt, ud.UserID, ud.DeviceID)
		if err != nil && r.logger != nil {
			r.logger.Error("failed to update user device", map[string]any{"error": err, "userID": ud.UserID, "deviceID": ud.DeviceID})
		}
		return err
	}
	_, err := r.db.Exec(newCtx, query, ipAddress, ud.LastActiveAt, ud.UserID, ud.DeviceID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to update user device", map[string]any{"error": err, "userID": ud.UserID, "deviceID": ud.DeviceID})
	}
	return err
}

// UpdateSessionID updates the session ID for a user-device relationship.
func (r *UserDeviceRepository) UpdateSessionID(ctx context.Context, userID, deviceID int64, sessionID string) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserDevice.UpdateSessionID")
	defer span.End()

	query := `UPDATE user_device SET session_id = $1 WHERE user_id = $2 AND device_id = $3`
	_, err := r.db.Exec(newCtx, query, sessionID, userID, deviceID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to update user device session ID", map[string]any{"error": err, "userID": userID, "deviceID": deviceID})
	}
	return err
}

// Delete removes a user-device relationship.
func (r *UserDeviceRepository) Delete(ctx context.Context, ud *model.UserDevice) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserDevice.Delete")
	defer span.End()

	query := `DELETE FROM user_device WHERE user_id = $1 AND device_id = $2`
	_, err := r.db.Exec(newCtx, query, ud.UserID, ud.DeviceID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to delete user device", map[string]any{"error": err, "userID": ud.UserID, "deviceID": ud.DeviceID})
	}
	return err
}

// Revoke revokes access for a user-device pair.
func (r *UserDeviceRepository) Revoke(ctx context.Context, userID, deviceID int64) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserDevice.Revoke")
	defer span.End()

	query := `UPDATE user_device SET revoked_at = $1 WHERE user_id = $2 AND device_id = $3`
	_, err := r.db.Exec(newCtx, query, time.Now().UTC(), userID, deviceID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to revoke user device", map[string]any{"error": err, "userID": userID, "deviceID": deviceID})
	}
	return err
}

// List retrieves all user-device relationships with pagination and filtering.
func (r *UserDeviceRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) (*model.UserDevices, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserDevice.List")
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

	// Get total count
	countQuery := `SELECT COUNT(*) FROM user_device`
	var total int64
	if err := r.db.QueryRow(newCtx, countQuery).Scan(&total); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to count user devices", map[string]any{"error": err})
		}
		return nil, err
	}

	// Get paginated results
	// Apply sorting
	orderByValue, err := validateOrderBy(pagination, "created_at", allowedOrderByUserDevice)
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
		SELECT user_id, device_id, ip_address::text, last_active_at, session_id, revoked_at, created_at
		FROM user_device
		ORDER BY %s
		LIMIT $1 OFFSET $2
	`, orderByClause)
	rows, err := r.db.Query(newCtx, query, limit, offset)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to list user devices", map[string]any{"error": err})
		}
		return nil, err
	}
	defer rows.Close()

	var userDevices []*model.UserDevice
	for rows.Next() {
		var ud model.UserDevice
		err := rows.Scan(
			&ud.UserID, &ud.DeviceID, &ud.IPAddress,
			&ud.LastActiveAt, &ud.SessionID, &ud.RevokedAt, &ud.CreatedAt,
		)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("failed to scan user device", map[string]any{"error": err})
			}
			return nil, err
		}
		userDevices = append(userDevices, &ud)
	}

	// Convert []*UserDevice to []UserDevice
	udItems := make([]model.UserDevice, len(userDevices))
	for i, ud := range userDevices {
		udItems[i] = *ud
	}

	return &model.UserDevices{
		Items: udItems,
		Meta:  buildMeta(total, page, limit),
	}, nil
}

func (r *UserDeviceRepository) scanUserDevice(row pgx.Row) (*model.UserDevice, error) {
	var ud model.UserDevice
	err := row.Scan(
		&ud.UserID, &ud.DeviceID, &ud.IPAddress,
		&ud.LastActiveAt, &ud.SessionID, &ud.RevokedAt, &ud.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrUserDeviceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ud, nil
}

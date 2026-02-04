package repository

import (
	"context"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)

// UserDeviceRepository implements repository.UserDeviceRepository for PostgreSQL.
type UserDeviceRepository struct {
	db PostgrePool
}

// NewUserDeviceRepository creates a new UserDeviceRepository.
func NewUserDeviceRepository(db PostgrePool) repository.UserDeviceRepository {
	return &UserDeviceRepository{db: db}
}

// GetByUserIDAndDeviceID finds a user-device relationship.
func (r *UserDeviceRepository) GetByUserIDAndDeviceID(ctx context.Context, userID, deviceID int64) (*model.UserDevice, error) {
	query := `
		SELECT user_id, device_id, ip_address, last_active_at, revoked_at, created_at
		FROM user_device
		WHERE user_id = $1 AND device_id = $2
	`
	return r.scanUserDevice(r.db.QueryRow(ctx, query, userID, deviceID))
}

// Create adds a new user-device relationship.
func (r *UserDeviceRepository) Create(ctx context.Context, ud *model.UserDevice) (*model.UserDevice, error) {
	query := `
		INSERT INTO user_device (user_id, device_id, ip_address, last_active_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Exec(ctx, query,
		ud.UserID, ud.DeviceID, ud.IPAddress, ud.LastActiveAt, ud.CreatedAt,
	)
	return ud, err
}

// Update modifies an existing relationship.
func (r *UserDeviceRepository) Update(ctx context.Context, ud *model.UserDevice) error {
	query := `
		UPDATE user_device
		SET ip_address = $1, last_active_at = $2
		WHERE user_id = $3 AND device_id = $4
	`
	_, err := r.db.Exec(ctx, query, ud.IPAddress, ud.LastActiveAt, ud.UserID, ud.DeviceID)
	return err
}

// Delete removes a user-device relationship.
func (r *UserDeviceRepository) Delete(ctx context.Context, ud *model.UserDevice) error {
	query := `DELETE FROM user_device WHERE user_id = $1 AND device_id = $2`
	_, err := r.db.Exec(ctx, query, ud.UserID, ud.DeviceID)
	return err
}

// Revoke revokes access for a user-device pair.
func (r *UserDeviceRepository) Revoke(ctx context.Context, userID, deviceID int64) error {
	query := `UPDATE user_device SET revoked_at = $1 WHERE user_id = $2 AND device_id = $3`
	_, err := r.db.Exec(ctx, query, time.Now().UTC(), userID, deviceID)
	return err
}

// List retrieves all user-device relationships with pagination and filtering.
func (r *UserDeviceRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserDeviceListFilterParam) (*model.UserDevices, error) {
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
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, err
	}

	// Get paginated results
	query := `
		SELECT user_id, device_id, ip_address, last_active_at, revoked_at, created_at
		FROM user_device
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userDevices []*model.UserDevice
	for rows.Next() {
		var ud model.UserDevice
		err := rows.Scan(
			&ud.UserID, &ud.DeviceID, &ud.IPAddress,
			&ud.LastActiveAt, &ud.RevokedAt, &ud.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		userDevices = append(userDevices, &ud)
	}

	// Convert []*UserDevice to []UserDevice
	udItems := make([]model.UserDevice, len(userDevices))
	for i, ud := range userDevices {
		udItems[i] = *ud
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &model.UserDevices{
		Items: udItems,
		Meta: model.Meta{
			Total:  total,
			Page:   page,
			Limit:  limit,
			Pages:  totalPages,
		},
	}, nil
}

func (r *UserDeviceRepository) scanUserDevice(row pgx.Row) (*model.UserDevice, error) {
	var ud model.UserDevice
	err := row.Scan(
		&ud.UserID, &ud.DeviceID, &ud.IPAddress,
		&ud.LastActiveAt, &ud.RevokedAt, &ud.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrDeviceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ud, nil
}

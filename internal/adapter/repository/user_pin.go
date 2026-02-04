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

// pinModel is the database model for PIN data.
type pinModel struct {
	UserID    int64
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// toDomain converts a PIN model to a domain entity.
func (m *pinModel) toDomain() *model.UserPin {
	return &model.UserPin{
		UserID:    m.UserID,
		Code:      m.Code,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// PinRepository implements port.PinRepository for PostgreSQL.
type PinRepository struct {
	db PostgrePool
}

// NewPinRepository creates a new PinRepository.
func NewPinRepository(db PostgrePool) repository.UserPinRepository {
	return &PinRepository{db: db}
}

// GetByUserID retrieves a PIN by user ID.
func (r *PinRepository) GetByUserID(ctx context.Context, userID int64) (*model.UserPin, error) {
	query := `
		SELECT user_id, code, created_at, updated_at
		FROM user_pin
		WHERE user_id = $1
	`
	var m pinModel
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&m.UserID, &m.Code, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrPinNotSet
	}
	if err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

// Create adds a new PIN.
func (r *PinRepository) Create(ctx context.Context, pin *model.UserPin) (*model.UserPin, error) {
	query := `
		INSERT INTO user_pin (user_id, code, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, pin.UserID, pin.Code, pin.CreatedAt, pin.UpdatedAt)
	return pin, err
}

// Update modifies an existing PIN.
func (r *PinRepository) Update(ctx context.Context, pin *model.UserPin) error {
	query := `
		UPDATE user_pin
		SET code = $1, updated_at = $2
		WHERE user_id = $3
	`
	_, err := r.db.Exec(ctx, query, pin.Code, pin.UpdatedAt, pin.UserID)
	return err
}

// Delete removes a PIN.
func (r *PinRepository) Delete(ctx context.Context, pin *model.UserPin) error {
	query := `DELETE FROM user_pin WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, pin.UserID)
	return err
}

// List retrieves all PINs with pagination and filtering.
func (r *PinRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserPinListFilterParam) (*model.UserPins, error) {
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
	countQuery := `SELECT COUNT(*) FROM user_pin`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, err
	}

	// Get paginated results
	query := `
		SELECT user_id, code, created_at, updated_at
		FROM user_pin
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pins []*model.UserPin
	for rows.Next() {
		var m pinModel
		err := rows.Scan(&m.UserID, &m.Code, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		pins = append(pins, m.toDomain())
	}

	// Convert []*UserPin to []UserPin
	pinItems := make([]model.UserPin, len(pins))
	for i, p := range pins {
		pinItems[i] = *p
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &model.UserPins{
		Items: pinItems,
		Meta: model.Meta{
			Total:  total,
			Page:   page,
			Limit:  limit,
			Pages:  totalPages,
		},
	}, nil
}

package repository

import (
	"context"
	"fmt"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)

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
	var m model.UserPin
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&m.UserID, &m.Code, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrPinNotSet
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
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

// allowedOrderByUserPin maps OrderBy string values to their typed enum for validation.
var allowedOrderByUserPin = map[string]param.UserPinOrderBy{
	"user_id":    param.OrderByUserPinUserID,
	"created_at": param.OrderByUserPinCreatedAt,
	"updated_at": param.OrderByUserPinUpdatedAt,
}

// List retrieves all PINs with pagination and filtering.
func (r *PinRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) (*model.UserPins, error) {
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
	// Apply sorting
	orderByValue := validateOrderBy(pagination, "created_at", allowedOrderByUserPin)

	// Build ORDER BY clause
	orderByClause := orderByValue
	if pagination != nil && pagination.Sort != nil && *pagination.Sort != "" {
		orderByClause += " " + *pagination.Sort
	} else {
		orderByClause += " DESC"
	}

	query := fmt.Sprintf(`
		SELECT user_id, code, created_at, updated_at
		FROM user_pin
		ORDER BY %s
		LIMIT $1 OFFSET $2
	`, orderByClause)
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pins []*model.UserPin
	for rows.Next() {
		var m model.UserPin
		err := rows.Scan(&m.UserID, &m.Code, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		pins = append(pins, &m)
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
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: totalPages,
		},
	}, nil
}

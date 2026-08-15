package repository

import (
	"context"
	"fmt"

	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)

// allowedOrderByUserPin maps OrderBy string values to their typed enum for validation.
var allowedOrderByUserPin = map[string]param.UserPinOrderBy{
	"user_id":    param.OrderByUserPinUserID,
	"created_at": param.OrderByUserPinCreatedAt,
	"updated_at": param.OrderByUserPinUpdatedAt,
}

// PinRepository implements port.PinRepository for PostgreSQL.
type PinRepository struct {
	db     PostgrePool
	tracer gomon.Tracer
	logger gomon.Logger
}

// NewPinRepository creates a new PinRepository.
func NewPinRepository(db PostgrePool, tracer gomon.Tracer, logger gomon.Logger) repository.UserPinRepository {
	if db == nil {
		panic("db is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &PinRepository{
		db:     db,
		tracer: tracer,
		logger: logger,
	}
}

// GetByUserID retrieves a PIN by user ID.
func (r *PinRepository) GetByUserID(ctx context.Context, userID int64) (*model.UserPin, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserPin.GetByUserID")
	defer span.End()

	query := `
		SELECT user_pin.user_id, "user".uid, user_pin.code, user_pin.created_at, user_pin.updated_at
		FROM user_pin
		JOIN "user" ON "user".id = user_pin.user_id
		WHERE user_pin.user_id = $1
	`
	var m model.UserPin
	err := r.db.QueryRow(newCtx, query, userID).Scan(
		&m.UserID, &m.UserUID, &m.Code, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrPinNotSet
	}
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to get user PIN", map[string]any{"error": err, "userID": userID})
		}
		return nil, err
	}
	return &m, nil
}

// Create adds a new PIN.
func (r *PinRepository) Create(ctx context.Context, pin *model.UserPin) (*model.UserPin, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserPin.Create")
	defer span.End()

	query := `
		INSERT INTO user_pin (user_id, code, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.Exec(newCtx, query, pin.UserID, pin.Code, pin.CreatedAt, pin.UpdatedAt)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to create user PIN", map[string]any{"error": err, "userID": pin.UserID})
	}
	return pin, err
}

// Update modifies an existing PIN.
func (r *PinRepository) Update(ctx context.Context, pin *model.UserPin) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserPin.Update")
	defer span.End()

	query := `
		UPDATE user_pin
		SET code = $1, updated_at = $2
		WHERE user_id = $3
	`
	_, err := r.db.Exec(newCtx, query, pin.Code, pin.UpdatedAt, pin.UserID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to update user PIN", map[string]any{"error": err, "userID": pin.UserID})
	}
	return err
}

// Delete removes a PIN.
func (r *PinRepository) Delete(ctx context.Context, pin *model.UserPin) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserPin.Delete")
	defer span.End()

	query := `DELETE FROM user_pin WHERE user_id = $1`
	_, err := r.db.Exec(newCtx, query, pin.UserID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to delete user PIN", map[string]any{"error": err, "userID": pin.UserID})
	}
	return err
}

// List retrieves all PINs with pagination and filtering.
func (r *PinRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) (*model.UserPins, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserPin.List")
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
	countQuery := `SELECT COUNT(*) FROM user_pin`
	var total int64
	if err := r.db.QueryRow(newCtx, countQuery).Scan(&total); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to count user PINs", map[string]any{"error": err})
		}
		return nil, err
	}

	// Get paginated results
	// Apply sorting
	orderByValue, err := validateOrderBy(pagination, "created_at", allowedOrderByUserPin)
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
		SELECT user_id, code, created_at, updated_at
		FROM user_pin
		ORDER BY %s
		LIMIT $1 OFFSET $2
	`, orderByClause)
	rows, err := r.db.Query(newCtx, query, limit, offset)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to list user PINs", map[string]any{"error": err})
		}
		return nil, err
	}
	defer rows.Close()

	var pins []*model.UserPin
	for rows.Next() {
		var m model.UserPin
		err := rows.Scan(&m.UserID, &m.Code, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("failed to scan user PIN", map[string]any{"error": err})
			}
			return nil, err
		}
		pins = append(pins, &m)
	}

	// Convert []*UserPin to []UserPin
	pinItems := make([]model.UserPin, len(pins))
	for i, p := range pins {
		pinItems[i] = *p
	}

	return &model.UserPins{
		Items: pinItems,
		Meta:  buildMeta(total, page, limit),
	}, nil
}

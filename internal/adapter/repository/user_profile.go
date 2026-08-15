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

// allowedOrderByUserProfile maps OrderBy string values to their typed enum for validation.
var allowedOrderByUserProfile = map[string]param.UserProfileOrderBy{
	"user_id":    param.OrderByUserProfileID,
	"created_at": param.OrderByUserProfileCreatedAt,
	"updated_at": param.OrderByUserProfileUpdatedAt,
}

// ProfileRepository implements port.ProfileRepository for PostgreSQL.
type ProfileRepository struct {
	db     PostgrePool
	tracer gomon.Tracer
	logger gomon.Logger
}

// NewProfileRepository creates a new ProfileRepository.
func NewProfileRepository(db PostgrePool, tracer gomon.Tracer, logger gomon.Logger) repository.UserProfileRepository {
	if db == nil {
		panic("db is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &ProfileRepository{
		db:     db,
		tracer: tracer,
		logger: logger,
	}
}

// GetByUserID retrieves a profile by user ID.
func (r *ProfileRepository) GetByUserID(ctx context.Context, userID int64) (*model.UserProfile, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserProfile.GetByUserID")
	defer span.End()

	query := `
		SELECT user_id, first_name, last_name, bio, avatar_file_id, attributes, created_at, updated_at
		FROM user_profile
		WHERE user_id = $1
	`
	profile, err := r.scanProfile(r.db.QueryRow(newCtx, query, userID))
	if err != nil && err != errors.ErrProfileNotFound && r.logger != nil {
		r.logger.Error("failed to get user profile", map[string]any{"error": err, "userID": userID})
	}
	return profile, err
}

// Create adds a new profile.
func (r *ProfileRepository) Create(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserProfile.Create")
	defer span.End()

	query := `
		INSERT INTO user_profile (user_id, first_name, last_name, bio, avatar_file_id, attributes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(newCtx, query,
		profile.UserID, profile.FirstName, profile.LastName, profile.Bio,
		profile.AvatarFileID, profile.Attributes, profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to create user profile", map[string]any{"error": err, "userID": profile.UserID})
	}
	return profile, err
}

// Update modifies an existing profile.
func (r *ProfileRepository) Update(ctx context.Context, profile *model.UserProfile) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserProfile.Update")
	defer span.End()

	query := `
		UPDATE user_profile
		SET first_name = $1, last_name = $2, bio = $3, avatar_file_id = $4, attributes = $5, updated_at = $6
		WHERE user_id = $7
	`
	_, err := r.db.Exec(newCtx, query,
		profile.FirstName, profile.LastName, profile.Bio,
		profile.AvatarFileID, profile.Attributes, profile.UpdatedAt, profile.UserID,
	)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to update user profile", map[string]any{"error": err, "userID": profile.UserID})
	}
	return err
}

// Delete removes a profile.
func (r *ProfileRepository) Delete(ctx context.Context, profile *model.UserProfile) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserProfile.Delete")
	defer span.End()

	query := `DELETE FROM user_profile WHERE user_id = $1`
	_, err := r.db.Exec(newCtx, query, profile.UserID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to delete user profile", map[string]any{"error": err, "userID": profile.UserID})
	}
	return err
}

// List retrieves all profiles with pagination and filtering.
func (r *ProfileRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserProfileListFilterParam) (*model.UserProfiles, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.UserProfile.List")
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
	countQuery := `SELECT COUNT(*) FROM user_profile`
	var total int64
	if err := r.db.QueryRow(newCtx, countQuery).Scan(&total); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to count user profiles", map[string]any{"error": err})
		}
		return nil, err
	}

	// Get paginated results
	// Apply sorting
	orderByValue, err := validateOrderBy(pagination, "created_at", allowedOrderByUserProfile)
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
		SELECT user_id, first_name, last_name, bio, avatar_file_id, attributes, created_at, updated_at
		FROM user_profile
		ORDER BY %s
		LIMIT $1 OFFSET $2
	`, orderByClause)
	rows, err := r.db.Query(newCtx, query, limit, offset)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to list user profiles", map[string]any{"error": err})
		}
		return nil, err
	}
	defer rows.Close()

	var profiles []*model.UserProfile
	for rows.Next() {
		var m model.UserProfile
		err := rows.Scan(
			&m.UserID, &m.FirstName, &m.LastName, &m.Bio,
			&m.AvatarFileID, &m.Attributes, &m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			if r.logger != nil {
				r.logger.Error("failed to scan user profile", map[string]any{"error": err})
			}
			return nil, err
		}
		profiles = append(profiles, &m)
	}

	// Convert []*UserProfile to []UserProfile
	profileItems := make([]model.UserProfile, len(profiles))
	for i, p := range profiles {
		profileItems[i] = *p
	}

	return &model.UserProfiles{
		Items: profileItems,
		Meta:  buildMeta(total, page, limit),
	}, nil
}

func (r *ProfileRepository) scanProfile(row pgx.Row) (*model.UserProfile, error) {
	var m model.UserProfile
	err := row.Scan(
		&m.UserID, &m.FirstName, &m.LastName, &m.Bio,
		&m.AvatarFileID, &m.Attributes, &m.CreatedAt, &m.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrProfileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan profile: %w", err)
	}
	return &m, nil
}

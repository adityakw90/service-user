package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)

// allowedOrderByUser maps OrderBy string values to their typed enum for validation.
var allowedOrderByUser = map[string]param.UserOrderBy{
	"id":         param.OrderByUserID,
	"uid":        param.OrderByUserUID,
	"username":   param.OrderByUserUsername,
	"email":      param.OrderByUserEmail,
	"status":     param.OrderByUserStatus,
	"created_at": param.OrderByUserCreatedAt,
	"updated_at": param.OrderByUserUpdatedAt,
}

// UserRepository implements repository.UserRepository for PostgreSQL.
type UserRepository struct {
	db     PostgrePool
	tracer gomon.Tracer
	logger gomon.Logger
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db PostgrePool, tracer gomon.Tracer, logger gomon.Logger) repository.UserRepository {
	if db == nil {
		panic("db is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &UserRepository{
		db:     db,
		tracer: tracer,
		logger: logger,
	}
}

// Create adds a new user to the database.
func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.Create")
	defer span.End()

	query := `INSERT INTO "user" (uid, username, email, password, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at`
	err := r.db.QueryRow(newCtx, query,
		user.UID, user.Username, user.Email, user.Password,
		user.Status,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to create user", map[string]any{"error": err, "uid": user.UID})
		}
		return nil, HandlePgError(err)
	}

	return user, nil
}

// GetByID retrieves a user by internal ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.GetByID")
	defer span.End()

	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE id = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(newCtx, query, id))
}

// GetByUID retrieves a user by public UID.
func (r *UserRepository) GetByUID(ctx context.Context, uid string) (*model.User, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.GetByUID")
	defer span.End()

	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE uid = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(newCtx, query, uid))
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.GetByEmail")
	defer span.End()

	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE email = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(newCtx, query, email))
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.GetByUsername")
	defer span.End()

	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE username = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(newCtx, query, username))
}

// GetByPhone retrieves a user by phone.
// the phone currently not implemented now
// TODO : update to support phone
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.GetByPhone")
	defer span.End()

	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE phone = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(newCtx, query, phone))
}

// Update modifies an existing user.
// Database handles updated_at via DEFAULT NOW().
// TODO: Add DB trigger to auto-update updated_at on UPDATE, then remove app-level UpdatedAt
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.Update")
	defer span.End()

	query := `UPDATE "user" SET username = $1, email = $2, password = $3, status = $4, updated_at = $5 WHERE id = $6`
	_, err := r.db.Exec(newCtx, query,
		user.Username, user.Email, user.Password, user.Status, user.UpdatedAt, user.ID,
	)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to update user", map[string]any{"error": err, "id": user.ID})
	}
	return err
}

// Delete marks a user as deleted.
func (r *UserRepository) Delete(ctx context.Context, user *model.User) error {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.Delete")
	defer span.End()

	query := `UPDATE "user" SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.Exec(newCtx, query, time.Now().UTC(), user.ID)
	if err != nil && r.logger != nil {
		r.logger.Error("failed to delete user", map[string]any{"error": err, "id": user.ID})
	}
	return err
}

// List retrieves users with pagination and filtering.
func (r *UserRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserListFilterParam) (*model.Users, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "repository.User.List")
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

	if filter != nil && len(filter.Uids) > 0 {
		conditions = append(conditions, fmt.Sprintf("uid = ANY($%d)", argIdx))
		args = append(args, filter.Uids)
		argIdx++
	}
	if filter != nil && filter.Username != nil {
		conditions = append(conditions, fmt.Sprintf("username = $%d", argIdx))
		args = append(args, *filter.Username)
		argIdx++
	}
	if filter != nil && filter.Email != nil {
		conditions = append(conditions, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *filter.Email)
		argIdx++
	}
	if filter != nil && filter.Query != nil {
		conditions = append(conditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+*filter.Query+"%")
		argIdx++
	}
	conditions = append(conditions, "deleted_at IS NULL")

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM \"user\" %s", whereClause)
	var total int64
	if err := r.db.QueryRow(newCtx, countQuery, args...).Scan(&total); err != nil {
		if r.logger != nil {
			r.logger.Error("failed to count users", map[string]any{"error": err})
		}
		return nil, err
	}

	// Get paginated results
	// Apply sorting
	orderByValue, err := validateOrderBy(pagination, "created_at", allowedOrderByUser)
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
		SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at
		FROM "user"
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderByClause, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(newCtx, query, args...)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to list users", map[string]any{"error": err})
		}
		return nil, err
	}
	defer rows.Close()

	users, err := r.scanRows(rows)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("failed to scan users", map[string]any{"error": err})
		}
		return nil, err
	}

	// Convert []*User to []User
	userItems := make([]model.User, len(users))
	for i, u := range users {
		userItems[i] = *u
	}

	return &model.Users{
		Items: userItems,
		Meta:  buildMeta(total, page, limit),
	}, nil
}

// AddUserDevice associates a device with a user.
func (r *UserRepository) AddUserDevice(ctx context.Context, user *model.User, device *model.Device) error {
	return nil
}

func (r *UserRepository) scanUser(row pgx.Row) (*model.User, error) {
	var m model.User
	err := row.Scan(
		&m.ID, &m.UID, &m.Username, &m.Email, &m.Password,
		&m.Status, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *UserRepository) scanRows(rows pgx.Rows) ([]*model.User, error) {
	var users []*model.User
	for rows.Next() {
		var m model.User
		err := rows.Scan(
			&m.ID, &m.UID, &m.Username, &m.Email, &m.Password,
			&m.Status, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &m)
	}
	return users, nil
}

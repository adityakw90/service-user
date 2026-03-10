package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	db PostgrePool
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db PostgrePool) repository.UserRepository {
	return &UserRepository{db: db}
}

// Create adds a new user to the database.
func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	query := `INSERT INTO "user" (uid, username, email, password, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := r.db.QueryRow(ctx, query,
		user.UID, user.Username, user.Email, user.Password,
		user.Status, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		// Convert PostgreSQL errors to domain errors
		if domainErr := HandlePgError(err); domainErr != nil {
			return nil, domainErr
		}
		return nil, err
	}

	return user, nil
}

// GetByID retrieves a user by internal ID.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE id = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(ctx, query, id))
}

// GetByUID retrieves a user by public UID.
func (r *UserRepository) GetByUID(ctx context.Context, uid string) (*model.User, error) {
	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE uid = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(ctx, query, uid))
}

// GetByEmail retrieves a user by email.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE email = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(ctx, query, email))
}

// GetByUsername retrieves a user by username.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE username = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(ctx, query, username))
}

// GetByPhone retrieves a user by phone.
// the phone currently not implemented now
// TODO : update to support phone
func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	query := `SELECT id, uid, username, email, password, status, created_at, updated_at, deleted_at FROM "user" WHERE phone = $1 AND deleted_at IS NULL`
	return r.scanUser(r.db.QueryRow(ctx, query, phone))
}

// Update modifies an existing user.
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `UPDATE "user" SET username = $1, email = $2, password = $3, status = $4, updated_at = $5 WHERE id = $6`
	_, err := r.db.Exec(ctx, query,
		user.Username, user.Email, user.Password, user.Status, user.UpdatedAt, user.ID,
	)
	return err
}

// Delete marks a user as deleted.
func (r *UserRepository) Delete(ctx context.Context, user *model.User) error {
	query := `UPDATE "user" SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, time.Now().UTC(), user.ID)
	return err
}

// List retrieves users with pagination and filtering.
func (r *UserRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserListFilterParam) (*model.Users, error) {
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
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Get paginated results
	// Apply sorting
	orderByValue := validateOrderBy(pagination, "created_at", allowedOrderByUser)

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

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users, err := r.scanRows(rows)
	if err != nil {
		return nil, err
	}

	// Convert []*User to []User
	userItems := make([]model.User, len(users))
	for i, u := range users {
		userItems[i] = *u
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &model.Users{
		Items: userItems,
		Meta: model.Meta{
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: totalPages,
		},
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

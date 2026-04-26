package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	"github.com/jackc/pgx/v5"
)

// allowedOrderByUserFile maps OrderBy string values to their typed enum for validation.
var allowedOrderByUserFile = map[string]param.UserFileOrderBy{
	"id":         param.OrderByUserFileID,
	"uid":        param.OrderByUserFileUID,
	"user_id":    param.OrderByUserFileUserID,
	"file_type":  param.OrderByUserFileFileType,
	"file_name":  param.OrderByUserFileFileName,
	"created_at": param.OrderByUserFileCreatedAt,
}

// UserFileRepository implements repository.UserFileRepository for PostgreSQL.
type UserFileRepository struct {
	db PostgrePool
}

// NewUserFileRepository creates a new UserFileRepository.
func NewUserFileRepository(db PostgrePool) repository.UserFileRepository {
	return &UserFileRepository{db: db}
}

// GetByID retrieves a file by internal ID.
func (r *UserFileRepository) GetByID(ctx context.Context, id int64) (*model.UserFile, error) {
	query := `
		SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at
		FROM user_file
		WHERE id = $1
	`
	return r.scanFile(r.db.QueryRow(ctx, query, id))
}

// GetByUID retrieves a file by public UID.
func (r *UserFileRepository) GetByUID(ctx context.Context, uid string) (*model.UserFile, error) {
	query := `
		SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at
		FROM user_file
		WHERE uid = $1
	`
	return r.scanFile(r.db.QueryRow(ctx, query, uid))
}

// Create adds a new file to the database.
func (r *UserFileRepository) Create(ctx context.Context, file *model.UserFile) (*model.UserFile, error) {
	query := `
		INSERT INTO user_file (uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	return file, r.db.QueryRow(ctx, query,
		file.UID, file.UserID, file.FileType, file.FileName,
		file.FilePath, file.MimeType, file.SizeBytes, file.Visibility, file.CreatedAt,
	).Scan(&file.ID)
}

// Update modifies an existing file.
func (r *UserFileRepository) Update(ctx context.Context, file *model.UserFile) error {
	query := `
		UPDATE user_file
		SET file_type = $1, file_name = $2, file_path = $3, mime_type = $4, size_bytes = $5, visibility = $6
		WHERE id = $7
	`
	_, err := r.db.Exec(ctx, query,
		file.FileType, file.FileName, file.FilePath, file.MimeType, file.SizeBytes, file.Visibility, file.ID,
	)
	return err
}

// Delete removes a file from the database.
func (r *UserFileRepository) Delete(ctx context.Context, file *model.UserFile) error {
	query := `DELETE FROM user_file WHERE id = $1`
	_, err := r.db.Exec(ctx, query, file.ID)
	return err
}

// List retrieves files with pagination and filtering.
func (r *UserFileRepository) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) (*model.UserFiles, error) {
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
			placeholders := make([]string, len(filter.Uids))
			for i := range filter.Uids {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, filter.Uids[i])
				argIdx++
			}
			conditions = append(conditions, fmt.Sprintf("uid IN (%s)", strings.Join(placeholders, ", ")))
		}
		if len(filter.UserUid) > 0 {
			// First get user IDs from UIDs - use IN clause for simplicity
			// Build separate args for the subquery to avoid index confusion
			var userQueryArgs []interface{}
			placeholders := make([]string, len(filter.UserUid))
			for i := range filter.UserUid {
				placeholders[i] = fmt.Sprintf("$%d", i+1)
				userQueryArgs = append(userQueryArgs, filter.UserUid[i])
			}
			userQuery := fmt.Sprintf(`SELECT id FROM "user" WHERE uid IN (%s)`, strings.Join(placeholders, ", "))
			rows, err := r.db.Query(ctx, userQuery, userQueryArgs...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var userIDs []int64
			for rows.Next() {
				var userID int64
				if err := rows.Scan(&userID); err != nil {
					return nil, err
				}
				userIDs = append(userIDs, userID)
			}
			if rows.Err() != nil {
				return nil, rows.Err()
			}

			if len(userIDs) == 0 {
				return nil, errors.ErrUserNotFound
			}

			placeholders = make([]string, len(userIDs))
			for i := range userIDs {
				placeholders[i] = fmt.Sprintf("$%d", argIdx)
				args = append(args, userIDs[i])
				argIdx++
			}
			conditions = append(conditions, fmt.Sprintf("user_id IN (%s)", strings.Join(placeholders, ", ")))
		}
		if filter.FileType != nil {
			conditions = append(conditions, fmt.Sprintf("file_type = $%d", argIdx))
			args = append(args, *filter.FileType)
			argIdx++
		}
		if filter.Visibility != nil {
			conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIdx))
			args = append(args, *filter.Visibility)
			argIdx++
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM user_file %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Get paginated results
	// Apply sorting
	orderByValue := validateOrderBy(pagination, "created_at", allowedOrderByUserFile)

	// Build ORDER BY clause
	orderByClause := orderByValue
	if pagination != nil && pagination.Sort != nil && *pagination.Sort != "" {
		orderByClause += " " + *pagination.Sort
	} else {
		orderByClause += " DESC"
	}

	query := fmt.Sprintf(`
		SELECT id, uid, user_id, file_type, file_name, file_path, mime_type, size_bytes, visibility, created_at
		FROM user_file
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

	files, err := r.scanRows(rows)
	if err != nil {
		return nil, err
	}

	// Convert []*UserFile to []UserFile
	fileItems := make([]model.UserFile, len(files))
	for i, f := range files {
		fileItems[i] = *f
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &model.UserFiles{
		Items: fileItems,
		Meta: model.Meta{
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: totalPages,
		},
	}, nil
}

func (r *UserFileRepository) scanFile(row pgx.Row) (*model.UserFile, error) {
	var m model.UserFile
	err := row.Scan(
		&m.ID, &m.UID, &m.UserID, &m.FileType, &m.FileName,
		&m.FilePath, &m.MimeType, &m.SizeBytes, &m.Visibility, &m.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrFileNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *UserFileRepository) scanRows(rows pgx.Rows) ([]*model.UserFile, error) {
	var files []*model.UserFile
	for rows.Next() {
		var m model.UserFile
		err := rows.Scan(
			&m.ID, &m.UID, &m.UserID, &m.FileType, &m.FileName,
			&m.FilePath, &m.MimeType, &m.SizeBytes, &m.Visibility, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		files = append(files, &m)
	}
	return files, nil
}

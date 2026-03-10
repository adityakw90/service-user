package repository

import (
	stderrors "errors"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func validateOrderBy[T any](
	pagination *param.PaginationParam,
	defaultOrderBy string,
	allowedOrderBy map[string]T,
) string {
	if pagination != nil && pagination.OrderBy != nil {
		if _, ok := allowedOrderBy[*pagination.OrderBy]; ok {
			return *pagination.OrderBy
		}
	}
	return defaultOrderBy
}

// HandlePgError converts PostgreSQL errors to domain errors.
// Returns nil if the error is not a recognized PostgreSQL error.
func HandlePgError(err error) error {
	if err == nil {
		return nil
	}

	// Handle pgx.ErrNoRows
	if stderrors.Is(err, pgx.ErrNoRows) {
		return errors.ErrUserNotFound
	}

	// Handle PostgreSQL errors via pgconn.PgError
	var pgErr *pgconn.PgError
	if stderrors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			switch pgErr.ConstraintName {
			case "idx_user_email_active":
				return errors.ErrDuplicateEmail
			case "idx_user_username_active":
				return errors.ErrDuplicateUsername
			default:
				return errors.ErrResourceConflict
			}
		}
	}

	// Not a recognized PostgreSQL error, return as-is
	return nil
}

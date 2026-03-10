package repository

import (
	"errors"

	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	domainParam "github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/jackc/pgx/v5/pgconn"
)

func validateOrderBy[T any](
	pagination *domainParam.PaginationParam,
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

// HandlePgError converts PostgreSQL unique constraint violations to domain errors.
// Returns the original error if it's not a recognized unique violation.
func HandlePgError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			switch pgErr.ConstraintName {
			case "idx_user_email_active":
				return domainErrors.ErrDuplicateEmail
			case "idx_user_username_active":
				return domainErrors.ErrDuplicateUsername
			default:
				return domainErrors.ErrResourceConflict
			}
		}
	}

	return err
}

package repository

import (
	"errors"
	"fmt"

	domainError "github.com/adityakw90/service-user/internal/core/domain/errors"
	domainModel "github.com/adityakw90/service-user/internal/core/domain/model"
	domainParam "github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/jackc/pgx/v5/pgconn"
)

func validateOrderBy[T any](
	pagination *domainParam.PaginationParam,
	defaultOrderBy string,
	allowedOrderBy map[string]T,
) (string, error) {
	if pagination != nil && pagination.OrderBy != nil {
		if *pagination.OrderBy == "" {
			return defaultOrderBy, nil
		}
		if _, ok := allowedOrderBy[*pagination.OrderBy]; ok {
			return *pagination.OrderBy, nil
		}
		return "", domainError.NewCustomError(
			domainError.ErrCodeValidation,
			domainError.ErrMessageValidation,
			domainError.ErrorMap{
				"order_by": []string{fmt.Sprintf("invalid order by: %s", *pagination.OrderBy)},
			},
		)
	}
	return defaultOrderBy, nil
}

func buildMeta(total int64, page int, limit int) *domainModel.Meta {
	return &domainModel.Meta{
		Total: total,
		Page:  page,
		Limit: limit,
		Pages: calcTotalPages(total, limit),
	}
}

func calcTotalPages(total int64, limit int) int {
	/*
		Using ceiling division formula
		based on benchmark result,
		ceiling formula is ~6% faster across all scenarios
		compared to division + modulo check
	*/
	if total <= 0 || limit <= 0 {
		return 1
	}
	return int((total + int64(limit) - 1) / int64(limit))
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
				return domainError.ErrDuplicateEmail
			case "idx_user_username_active":
				return domainError.ErrDuplicateUsername
			default:
				return domainError.ErrResourceConflict
			}
		}
	}

	return err
}

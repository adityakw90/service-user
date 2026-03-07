package repository

import "github.com/adityakw90/service-user/internal/core/domain/param"

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

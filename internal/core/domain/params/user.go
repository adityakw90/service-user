package params

import "github.com/adityakw90/service-user/internal/core/domain/model"

type UserListFilterParam struct {
	Uids     []string
	Username *string
	Email    *string
	Exists   *bool // For existence check
	Status   *model.UserStatus
	Query    *string
}

type UserListParam struct {
	Pagination *PaginationParam
	Filter     *UserListFilterParam
}

type UserCreateParam struct {
	Username string
	Email    string
	Password string
}

type UserUpdateParam struct {
	Username *string
	Email    *string
	Password *string
	Status   *model.UserStatus
}

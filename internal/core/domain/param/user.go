package param

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

type UserChangePasswordParam struct {
	CurrentPassword string
	NewPassword     string
}

// UserOrderBy represents allowed OrderBy column values for User.
type UserOrderBy string

const (
	OrderByUserID        UserOrderBy = "id"
	OrderByUserUID       UserOrderBy = "uid"
	OrderByUserUsername  UserOrderBy = "username"
	OrderByUserEmail     UserOrderBy = "email"
	OrderByUserStatus    UserOrderBy = "status"
	OrderByUserCreatedAt UserOrderBy = "created_at"
	OrderByUserUpdatedAt UserOrderBy = "updated_at"
)

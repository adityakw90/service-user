package param

type UserPinUpdateParam struct {
	Code *string
}

type UserPinListFilterParam struct {
	UserUIDs []string
}

type UserPinListParam struct {
	Pagination *PaginationParam
	Filter     *UserPinListFilterParam
}

// UserPinOrderBy represents allowed OrderBy column values for UserPin.
type UserPinOrderBy string

const (
	OrderByUserPinUserID     UserPinOrderBy = "user_id"
	OrderByUserPinCreatedAt  UserPinOrderBy = "created_at"
	OrderByUserPinUpdatedAt  UserPinOrderBy = "updated_at"
)

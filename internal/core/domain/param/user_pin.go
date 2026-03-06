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

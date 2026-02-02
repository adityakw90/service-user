package params

type UserProfileUpdateParam struct {
	FirstName  *string
	LastName   *string
	Bio        *string
	Avatar     []byte
	Attributes map[string]any
}

type UserProfileListFilterParam struct {
	UserUIDs []string
}

type UserProfileListParam struct {
	Pagination *PaginationParam
	Filter     *UserProfileListFilterParam
}

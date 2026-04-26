package param

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

// UserProfileOrderBy represents allowed OrderBy column values for UserProfile.
type UserProfileOrderBy string

const (
	OrderByUserProfileID         UserProfileOrderBy = "user_id"
	OrderByUserProfileCreatedAt UserProfileOrderBy = "created_at"
	OrderByUserProfileUpdatedAt UserProfileOrderBy = "updated_at"
)

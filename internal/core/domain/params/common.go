package params

// Default values for pagination
const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100 // Maximum allowed items per page
	DefaultSort  = "asc"
)

type PaginationParam struct {
	Page    int
	Limit   int
	OrderBy string
	Sort    string
}

func NewPaginationParam(page int, limit int, orderBy string, sort string) *PaginationParam {
	if page < 1 {
		page = DefaultPage
	}
	if limit < 1 {
		limit = DefaultLimit
	} else if limit > MaxLimit {
		limit = MaxLimit
	}
	if sort == "" {
		sort = DefaultSort
	}
	return &PaginationParam{
		Page:    page,
		Limit:   limit,
		OrderBy: orderBy,
		Sort:    sort,
	}
}

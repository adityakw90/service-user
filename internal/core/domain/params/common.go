package params

type PaginationParam struct {
	Page    *int
	Limit   *int
	OrderBy *string
	Sort    *string
}

func NewPaginationParam(page int, limit int, orderBy string, sort string) *PaginationParam {
	return &PaginationParam{
		Page:    &page,
		Limit:   &limit,
		OrderBy: &orderBy,
		Sort:    &sort,
	}
}

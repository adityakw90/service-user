package params

type PaginationParam struct {
	Page    *int
	Limit   *int
	OrderBy *string
	Sort    *string
}

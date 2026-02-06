package request

import (
	"strings"

	common "github.com/adityakw90/service-user-proto/gen/go/common"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

type PaginationRequest struct {
	Page    int    `validate:"required,min=1"`
	Limit   int    `validate:"required,min=1,max=100"`
	Sort    string `validate:"omitempty,oneof=asc desc"`
	OrderBy string `validate:"omitempty"`
}

func (pr *PaginationRequest) ToPaginationParams() *params.PaginationParam {
	return &params.PaginationParam{
		Page:    &pr.Page,
		Limit:   &pr.Limit,
		Sort:    &pr.Sort,
		OrderBy: &pr.OrderBy,
	}
}

func PaginationRequestFromPb(req *common.Pagination) *PaginationRequest {
	return &PaginationRequest{
		Page:    int(req.Page),
		Limit:   int(req.Limit),
		Sort:    strings.TrimSpace(req.Sort),
		OrderBy: strings.TrimSpace(req.OrderBy),
	}
}

package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type UserFileService interface {
	Get(ctx context.Context, uid string) (*model.UserFile, error)
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error)
	Add(ctx context.Context, param params.UserFileCreateParam) (*model.UserFile, error)
	Update(ctx context.Context, uid string, param params.UserFileUpdateParam) error
	Delete(ctx context.Context, uid string) error
}

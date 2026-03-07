package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type UserFileService interface {
	Get(ctx context.Context, uid string) (*model.UserFile, error)
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) (*model.UserFiles, error)
	Add(ctx context.Context, param param.UserFileCreateParam) (*model.UserFile, error)
	Update(ctx context.Context, uid string, param param.UserFileUpdateParam) error
	Delete(ctx context.Context, uid string) error
}

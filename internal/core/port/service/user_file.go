package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

type UserFileService interface {
	Get(ctx context.Context, uid string) (*model.UserFile, error)
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.FileListFilterParam) (*model.Files, error)
	Add(ctx context.Context, param params.FileCreateParam) (*model.UserFile, error)
	Update(ctx context.Context, uid string, param params.FileUpdateParam) error
	Delete(ctx context.Context, uid string) error
}

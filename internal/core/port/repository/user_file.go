package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type UserFileRepository interface {
	Create(ctx context.Context, userFile *model.UserFile) (*model.UserFile, error)
	Update(ctx context.Context, userFile *model.UserFile) error
	Delete(ctx context.Context, userFile *model.UserFile) error
	GetByID(ctx context.Context, id int64) (*model.UserFile, error)
	GetByUID(ctx context.Context, uid string) (*model.UserFile, error)
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error)
}

package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

type UserFileRepository interface {
	Create(ctx context.Context, userFile *model.UserFile) (*model.UserFile, error)
	Update(ctx context.Context, userFile *model.UserFile) error
	Delete(ctx context.Context, userFile *model.UserFile) error
	GetByID(ctx context.Context, id int64) (*model.UserFile, error)
	GetByUID(ctx context.Context, uid string) (*model.UserFile, error)
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error)
	MapIDsByUIDs(ctx context.Context, userFileUIDs []string) (map[string]int64, error)
	MapUIDsByIDs(ctx context.Context, userFileIDs []int64) (map[int64]string, error)
}

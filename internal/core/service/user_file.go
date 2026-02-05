package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
)

type userFileService struct {
}

func NewUserFileService() portSvc.UserFileService {
	return &userFileService{}
}

func (s *userFileService) Get(ctx context.Context, uid string) (*model.UserFile, error) {
	return nil, nil
}

func (s *userFileService) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error) {
	return nil, nil
}

func (s *userFileService) Add(ctx context.Context, param params.UserFileCreateParam) (*model.UserFile, error) {
	return nil, nil
}

func (s *userFileService) Update(ctx context.Context, uid string, param params.UserFileUpdateParam) error {
	return nil
}

func (s *userFileService) Delete(ctx context.Context, uid string) error {
	return nil
}

package service

import (
	"context"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portSec "github.com/adityakw90/service-user/internal/core/port/security"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
)

type userFileService struct {
	userFileRepo repository.UserFileRepository
	userRepo     repository.UserRepository
	uidGen       portSec.UIDGenerator
}

func NewUserFileService(
	userFileRepo repository.UserFileRepository,
	userRepo repository.UserRepository,
	uidGen portSec.UIDGenerator,
) portSvc.UserFileService {
	return &userFileService{
		userFileRepo: userFileRepo,
		userRepo:     userRepo,
		uidGen:       uidGen,
	}
}

func (s *userFileService) Get(ctx context.Context, uid string) (*model.UserFile, error) {
	if uid == "" {
		return nil, domainerrors.ErrInvalidUID
	}

	file, err := s.userFileRepo.GetByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *userFileService) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error) {
	// Set defaults for pagination
	if pagination == nil {
		pagination = params.NewPaginationParam(1, 10, "created_at", "desc")
	}

	if filter == nil {
		filter = &params.UserFileListFilterParam{}
	}

	files, err := s.userFileRepo.List(ctx, pagination, filter)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func (s *userFileService) Add(ctx context.Context, param params.UserFileCreateParam) (*model.UserFile, error) {
	// Validate input
	if param.UserUID == "" {
		return nil, domainerrors.ErrInvalidUID
	}

	// Verify user exists
	user, err := s.userRepo.GetByUID(ctx, param.UserUID)
	if err != nil {
		return nil, err
	}

	// Create file record
	file := &model.UserFile{
		UID:        s.uidGen.New(),
		UserID:     user.ID,
		UserUID:    param.UserUID,
		FileType:   param.FileType,
		FileName:   param.FileName,
		FilePath:   param.FilePath,
		MimeType:   param.MimeType,
		SizeBytes:  param.SizeBytes,
		Visibility: param.Visibility,
	}

	file, err = s.userFileRepo.Create(ctx, file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *userFileService) Update(ctx context.Context, uid string, param params.UserFileUpdateParam) error {
	// Get file
	file, err := s.userFileRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// Update fields
	if param.FileName != nil {
		file.FileName = *param.FileName
	}
	if param.FilePath != nil {
		file.FilePath = *param.FilePath
	}
	if param.MimeType != nil {
		file.MimeType = *param.MimeType
	}
	if param.SizeBytes != nil {
		file.SizeBytes = *param.SizeBytes
	}
	if param.Visibility != nil {
		file.Visibility = *param.Visibility
	}

	// Save changes
	return s.userFileRepo.Update(ctx, file)
}

func (s *userFileService) Delete(ctx context.Context, uid string) error {
	// Get file
	file, err := s.userFileRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// Delete file
	return s.userFileRepo.Delete(ctx, file)
}

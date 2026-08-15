package service

import (
	"context"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portResolver "github.com/adityakw90/service-user/internal/core/port/resolver"
	portSec "github.com/adityakw90/service-user/internal/core/port/security"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
	"github.com/adityakw90/service-user/pkg/util"
)

type userFileService struct {
	userFileRepo   repository.UserFileRepository
	userRepo       repository.UserRepository
	userResolver   portResolver.UserResolver
	uidGen         portSec.UIDGenerator
	eventPublisher portEvent.EventPublisher
}

func NewUserFileService(
	userFileRepo repository.UserFileRepository,
	userRepo repository.UserRepository,
	userResolver portResolver.UserResolver,
	uidGen portSec.UIDGenerator,
	eventPublisher portEvent.EventPublisher,
) portSvc.UserFileService {
	return &userFileService{
		userFileRepo:   userFileRepo,
		userRepo:       userRepo,
		userResolver:   userResolver,
		uidGen:         uidGen,
		eventPublisher: eventPublisher,
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

	resUserUid, err := s.userResolver.UIDsByIDs(ctx, []int64{
		file.UserID,
	})
	if err != nil {
		return nil, err
	}
	file.UserUID = resUserUid[file.UserID]

	return file, nil
}

func (s *userFileService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserFileListFilterParam) (*model.UserFiles, error) {
	// Set defaults for pagination
	if pagination == nil {
		pagination = &param.PaginationParam{
			Page:    util.Ptr(1),
			Limit:   util.Ptr(10),
			Sort:    util.Ptr("asc"),
			OrderBy: util.Ptr("created_at"),
		}
	}

	if filter == nil {
		filter = &param.UserFileListFilterParam{}
	}

	files, err := s.userFileRepo.List(ctx, pagination, filter)
	if err != nil {
		return nil, err
	}

	if len(files.Items) > 0 {
		var listUserId []int64
		for _, item := range files.Items {
			listUserId = append(listUserId, item.UserID)
		}
		resUserUIDs, err := s.userResolver.UIDsByIDs(ctx, listUserId)
		if err != nil {
			return nil, err
		}
		for i, item := range files.Items {
			files.Items[i].UserUID = resUserUIDs[item.UserID]
		}
	}

	return files, nil
}

func (s *userFileService) Add(ctx context.Context, param param.UserFileCreateParam) (*model.UserFile, error) {
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

	// Publish user file created event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserFileCreated, Entity: event.NewUserFileEntity(file), Metadata: event.EventUserFileCreatedData{
		UserUID:  file.UserUID,
		FileName: file.FileName,
	}})
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (s *userFileService) Update(ctx context.Context, uid string, param param.UserFileUpdateParam) error {
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
	err = s.userFileRepo.Update(ctx, file)
	if err != nil {
		return err
	}

	// Publish user file updated event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserFileUpdated, Entity: event.NewUserFileEntity(file), Metadata: event.EventUserFileUpdatedData{
		UserUID: file.UserUID,
	}})
	if err != nil {
		return err
	}

	return nil
}

func (s *userFileService) Delete(ctx context.Context, uid string) error {
	// Get file
	file, err := s.userFileRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// Delete file
	err = s.userFileRepo.Delete(ctx, file)
	if err != nil {
		return err
	}

	// Publish user file deleted event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserFileDeleted, Entity: event.NewUserFileEntity(file), Metadata: event.EventUserFileDeletedData{
		UserUID: file.UserUID,
	}})
	if err != nil {
		return err
	}

	return nil
}

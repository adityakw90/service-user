package service

import (
	"context"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/adityakw90/service-user/internal/core/port/observer"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portResolver "github.com/adityakw90/service-user/internal/core/port/resolver"
	portSec "github.com/adityakw90/service-user/internal/core/port/security"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
	"github.com/adityakw90/service-user/pkg/util"
)

type userFileService struct {
	userFileRepo     repository.UserFileRepository
	userRepo         repository.UserRepository
	userResolver     portResolver.UserResolver
	uidGen           portSec.UIDGenerator
	userFileObserver observer.ServiceObserver[signal.UserFileSignal]
	eventPublisher   portEvent.EventPublisher
}

func NewUserFileService(
	userFileRepo repository.UserFileRepository,
	userRepo repository.UserRepository,
	userResolver portResolver.UserResolver,
	uidGen portSec.UIDGenerator,
	userFileObserver observer.ServiceObserver[signal.UserFileSignal],
	eventPublisher portEvent.EventPublisher,
) portSvc.UserFileService {
	if userFileObserver == nil {
		panic("userFileObserver is required")
	}
	return &userFileService{
		userFileRepo:     userFileRepo,
		userRepo:         userRepo,
		userResolver:     userResolver,
		uidGen:           uidGen,
		userFileObserver: userFileObserver,
		eventPublisher:   eventPublisher,
	}
}

func (s *userFileService) Get(ctx context.Context, uid string) (*model.UserFile, error) {
	s.userFileObserver.OnSignal(ctx, signal.SignalStart, signal.UserFileSignal{
		UID:       &uid,
		Operation: "get",
	}, nil)

	if uid == "" {
		s.userFileObserver.OnSignal(ctx, signal.SignalReject, signal.UserFileSignal{
			Operation: "get",
		}, domainerrors.ErrInvalidUID)
		return nil, domainerrors.ErrInvalidUID
	}

	file, err := s.userFileRepo.GetByUID(ctx, uid)
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}

	resUserUid, err := s.userResolver.UIDsByIDs(ctx, []int64{
		file.UserID,
	})
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}
	file.UserUID = resUserUid[file.UserID]

	s.userFileObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserFileSignal{
		UID:       &uid,
		UserUID:   &file.UserUID,
		FileName:  &file.FileName,
		FileType:  &file.FileType,
		FileSize:  &file.SizeBytes,
		Operation: "get",
	}, nil)

	return file, nil
}

func (s *userFileService) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error) {
	s.userFileObserver.OnSignal(ctx, signal.SignalStart, signal.UserFileSignal{
		Operation: "list",
	}, nil)

	// Set defaults for pagination
	if pagination == nil {
		pagination = &params.PaginationParam{
			Page:    util.Ptr(1),
			Limit:   util.Ptr(10),
			Sort:    util.Ptr("asc"),
			OrderBy: util.Ptr("created_at"),
		}
	}

	if filter == nil {
		filter = &params.UserFileListFilterParam{}
	}

	files, err := s.userFileRepo.List(ctx, pagination, filter)
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			Operation: "list",
		}, err)
		return nil, err
	}

	if len(files.Items) > 0 {
		var listUserId []int64
		for _, item := range files.Items {
			listUserId = append(listUserId, item.UserID)
		}
		resUserUIDs, err := s.userResolver.UIDsByIDs(ctx, listUserId)
		if err != nil {
			s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
				Operation: "list",
			}, err)
			return nil, err
		}
		for i, item := range files.Items {
			files.Items[i].UserUID = resUserUIDs[item.UserID]
		}
	}

	s.userFileObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserFileSignal{
		Operation: "list",
	}, nil)

	return files, nil
}

func (s *userFileService) Add(ctx context.Context, param params.UserFileCreateParam) (*model.UserFile, error) {
	s.userFileObserver.OnSignal(ctx, signal.SignalStart, signal.UserFileSignal{
		UserUID:   &param.UserUID,
		FileName:  &param.FileName,
		FileType:  &param.FileType,
		FileSize:  &param.SizeBytes,
		Operation: "add",
	}, nil)

	// Validate input
	if param.UserUID == "" {
		s.userFileObserver.OnSignal(ctx, signal.SignalReject, signal.UserFileSignal{
			Operation: "add",
		}, domainerrors.ErrInvalidUID)
		return nil, domainerrors.ErrInvalidUID
	}

	// Verify user exists
	user, err := s.userRepo.GetByUID(ctx, param.UserUID)
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UserUID:   &param.UserUID,
			Operation: "add",
		}, err)
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
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UserUID:   &param.UserUID,
			Operation: "add",
		}, err)
		return nil, err
	}

	// Publish user file created event

	err = s.eventPublisher.Publish(ctx, event.EventUserFileCreated, event.EventUserFileCreatedData{
		UserUID:  file.UserUID,
		FileUID:  file.UID,
		FileName: file.FileName,
	})
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UserUID:   &param.UserUID,
			Operation: "add",
		}, err)
		return nil, err
	}

	s.userFileObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserFileSignal{
		UID:       &file.UID,
		UserUID:   &file.UserUID,
		FileName:  &file.FileName,
		FileType:  &file.FileType,
		FileSize:  &file.SizeBytes,
		Operation: "add",
	}, nil)

	return file, nil
}

func (s *userFileService) Update(ctx context.Context, uid string, param params.UserFileUpdateParam) error {
	s.userFileObserver.OnSignal(ctx, signal.SignalStart, signal.UserFileSignal{
		UID:       &uid,
		Operation: "update",
	}, nil)

	// Get file
	file, err := s.userFileRepo.GetByUID(ctx, uid)
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "update",
		}, err)
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
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "update",
		}, err)
		return err
	}

	// Publish user file updated event
	err = s.eventPublisher.Publish(ctx, event.EventUserFileUpdated, event.EventUserFileUpdatedData{
		UserUID: file.UserUID,
		FileUID: uid,
	})
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "update",
		}, err)
		return err
	}

	s.userFileObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserFileSignal{
		UID:       &uid,
		UserUID:   &file.UserUID,
		FileName:  &file.FileName,
		FileType:  &file.FileType,
		FileSize:  &file.SizeBytes,
		Operation: "update",
	}, nil)

	return nil
}

func (s *userFileService) Delete(ctx context.Context, uid string) error {
	s.userFileObserver.OnSignal(ctx, signal.SignalStart, signal.UserFileSignal{
		UID:       &uid,
		Operation: "delete",
	}, nil)

	// Get file
	file, err := s.userFileRepo.GetByUID(ctx, uid)
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	// Delete file
	err = s.userFileRepo.Delete(ctx, file)
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	// Publish user file deleted event

	err = s.eventPublisher.Publish(ctx, event.EventUserFileDeleted, event.EventUserFileDeletedData{
		UserUID: file.UserUID,
		FileUID: uid,
	})
	if err != nil {
		s.userFileObserver.OnSignal(ctx, signal.SignalFail, signal.UserFileSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	s.userFileObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserFileSignal{
		UID:       &uid,
		UserUID:   &file.UserUID,
		FileName:  &file.FileName,
		FileType:  &file.FileType,
		FileSize:  &file.SizeBytes,
		Operation: "delete",
	}, nil)

	return nil
}

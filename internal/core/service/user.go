package service

import (
	"context"
	"errors"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/adityakw90/service-user/internal/core/port/observer"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portResolver "github.com/adityakw90/service-user/internal/core/port/resolver"
	portSec "github.com/adityakw90/service-user/internal/core/port/security"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
	"github.com/adityakw90/service-user/pkg/util"
)

type userService struct {
	userRepo       repository.UserRepository
	profileRepo    repository.UserProfileRepository
	pinRepo        repository.UserPinRepository
	deviceRepo     repository.DeviceRepository
	userDeviceRepo repository.UserDeviceRepository
	passwordHasher portSec.Hasher
	pinHasher      portSec.Hasher
	uidGen         portSec.UIDGenerator
	tokenWhitelist portSec.TokenStore
	userObserver   observer.ServiceObserver[signal.UserSignal]
	eventPublisher portEvent.EventPublisher
	resolvers      portResolver.ResolverProvider
}

func NewUserService(
	userRepo repository.UserRepository,
	profileRepo repository.UserProfileRepository,
	pinRepo repository.UserPinRepository,
	deviceRepo repository.DeviceRepository,
	userDeviceRepo repository.UserDeviceRepository,
	passwordHasher portSec.Hasher,
	pinHasher portSec.Hasher,
	uidGen portSec.UIDGenerator,
	tokenWhitelist portSec.TokenStore,
	userObserver observer.ServiceObserver[signal.UserSignal],
	eventPublisher portEvent.EventPublisher,
	resolvers portResolver.ResolverProvider,
) portSvc.UserService {
	if userObserver == nil {
		panic("userObserver is required")
	}
	return &userService{
		userRepo:       userRepo,
		profileRepo:    profileRepo,
		pinRepo:        pinRepo,
		deviceRepo:     deviceRepo,
		userDeviceRepo: userDeviceRepo,
		passwordHasher: passwordHasher,
		pinHasher:      pinHasher,
		uidGen:         uidGen,
		tokenWhitelist: tokenWhitelist,
		userObserver:   userObserver,
		eventPublisher: eventPublisher,
		resolvers:      resolvers,
	}
}

func (s *userService) Get(ctx context.Context, uid string) (*model.User, error) {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &uid,
		Operation: "get",
	}, nil)

	if uid == "" {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			Operation: "get",
		}, domainerrors.ErrInvalidUID)
		return nil, domainerrors.ErrInvalidUID
	}

	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "get",
			}, err)
			return nil, err
		}
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, domainerrors.ErrUserGetFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}

	if user.IsDeleted() {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &uid,
			Operation: "get",
		}, domainerrors.ErrUserDeleted)
		return nil, domainerrors.ErrUserDeleted
	}

	if !user.IsActive() {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &uid,
			Operation: "get",
		}, domainerrors.ErrUserInactive)
		return nil, domainerrors.ErrUserInactive
	}

	active := user.IsActive()
	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &uid,
		Username:  &user.Username,
		Email:     &user.Email,
		Status:    &user.Status,
		Active:    &active,
		Operation: "get",
	}, nil)

	return user, nil
}

func (s *userService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserListFilterParam) (*model.Users, error) {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		Operation: "list",
	}, nil)

	// Set defaults for pagination
	if pagination == nil {
		pagination = &param.PaginationParam{
			Page:    util.Ptr(1),
			Limit:   util.Ptr(10),
			OrderBy: util.Ptr("created_at"),
			Sort:    util.Ptr("desc"),
		}
	}

	users, err := s.userRepo.List(ctx, pagination, filter)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "list",
		}, err)
		return nil, err
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		Operation: "list",
	}, nil)

	return users, nil
}

func (s *userService) Create(ctx context.Context, createParam *param.UserCreateParam) (*model.User, error) {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		Username:  &createParam.Username,
		Email:     &createParam.Email,
		Operation: "create",
	}, nil)

	// Validate input
	if createParam.Username == "" {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			Operation: "create",
		}, domainerrors.ErrInvalidUsername)
		return nil, domainerrors.ErrInvalidUsername
	}
	if createParam.Email == "" {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			Operation: "create",
		}, domainerrors.ErrInvalidEmail)
		return nil, domainerrors.ErrInvalidEmail
	}
	if createParam.Password == "" {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			Operation: "create",
		}, domainerrors.ErrInvalidPassword)
		return nil, domainerrors.ErrInvalidPassword
	}

	// Check for duplicate email
	_, err := s.userRepo.GetByEmail(ctx, createParam.Email)
	if err == nil {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			Email:     &createParam.Email,
			Operation: "create",
		}, domainerrors.ErrDuplicateEmail)
		return nil, domainerrors.ErrDuplicateEmail
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "create",
		}, err)
		return nil, err
	}

	// Check for duplicate username
	_, err = s.userRepo.GetByUsername(ctx, createParam.Username)
	if err == nil {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			Username:  &createParam.Username,
			Operation: "create",
		}, domainerrors.ErrDuplicateUsername)
		return nil, domainerrors.ErrDuplicateUsername
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "create",
		}, err)
		return nil, err
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(createParam.Password)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "create",
		}, err)
		return nil, err
	}

	// Create user
	user := &model.User{
		UID:      s.uidGen.New(),
		Username: createParam.Username,
		Email:    createParam.Email,
		Password: hashedPassword,
		Status:   model.UserStatusActive,
	}

	user, err = s.userRepo.Create(ctx, user)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "create",
		}, err)
		return nil, err
	}

	// Invalidate cache for new user
	_ = s.resolvers.User().Invalidate(ctx, param.WithUIDs(user.UID))

	// Create empty profile
	profile := &model.UserProfile{
		UserID:  user.ID,
		UserUID: user.UID,
	}
	_, err = s.profileRepo.Create(ctx, profile)
	if err != nil {
		// Log error but don't fail user creation
	}

	// Publish user created event
	err = s.eventPublisher.Publish(ctx, event.EventUserCreated, event.EventUserCreatedData{
		UserUID:  user.UID,
		ActorUID: user.UID,
		Username: user.Username,
		Email:    user.Email,
		Status:   string(user.Status),
	})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "create",
		}, err)
	}

	active := user.IsActive()
	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &user.UID,
		Username:  &user.Username,
		Email:     &user.Email,
		Status:    &user.Status,
		Active:    &active,
		Operation: "create",
	}, nil)

	return user, nil
}

func (s *userService) Update(ctx context.Context, uid string, updateParam *param.UserUpdateParam) error {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &uid,
		Operation: "update",
	}, nil)

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, err)
			return err
		}
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &uid,
			Operation: "update",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &uid,
			Operation: "update",
		}, err)
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, err)
			return err
		}
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &uid,
			Operation: "update",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	// Check if user is deleted
	if user.IsDeleted() {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &uid,
			Operation: "update",
		}, domainerrors.ErrUserDeleted)
		return domainerrors.ErrUserDeleted
	}

	changesCount := 0

	// Update username if provided
	if updateParam.Username != nil {
		if *updateParam.Username == "" {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, domainerrors.ErrInvalidUsername)
			return domainerrors.ErrInvalidUsername
		}
		// Check for duplicate
		existing, err := s.userRepo.GetByUsername(ctx, *updateParam.Username)
		if err == nil && existing.UID != uid {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, domainerrors.ErrDuplicateUsername)
			return domainerrors.ErrDuplicateUsername
		} else if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, err)
			return err
		}
		user.Username = *updateParam.Username
		changesCount++
	}

	// Update email if provided
	if updateParam.Email != nil {
		if *updateParam.Email == "" {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, domainerrors.ErrInvalidEmail)
			return domainerrors.ErrInvalidEmail
		}
		// Check for duplicate
		existing, err := s.userRepo.GetByEmail(ctx, *updateParam.Email)
		if err == nil && existing.UID != uid {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, domainerrors.ErrDuplicateEmail)
			return domainerrors.ErrDuplicateEmail
		} else if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, err)
			return err
		}
		user.Email = *updateParam.Email
		changesCount++
	}

	// Update password if provided
	if updateParam.Password != nil {
		if *updateParam.Password == "" {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, domainerrors.ErrInvalidPassword)
			return domainerrors.ErrInvalidPassword
		}
		hashedPassword, err := s.passwordHasher.Hash(*updateParam.Password)
		if err != nil {
			s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
				UID:       &uid,
				Operation: "update",
			}, err)
			return err
		}
		user.Password = hashedPassword
		changesCount++
	}

	// Update status if provided
	if updateParam.Status != nil {
		user.Status = *updateParam.Status
		changesCount++
	}

	// Save changes
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &uid,
			Operation: "update",
		}, err)
		return err
	}

	// Invalidate cache for updated user
	_ = s.resolvers.User().Invalidate(ctx, param.WithUIDs(user.UID), param.WithIDs(user.ID))

	// Publish user updated event
	err = s.eventPublisher.Publish(ctx, event.EventUserUpdated, event.EventUserUpdatedData{
		UserUID:      uid,
		ActorUID:     uid,
		ChangesCount: changesCount,
	})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "update",
		}, err)
	}

	active := user.IsActive()
	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:          &uid,
		Username:     &user.Username,
		Email:        &user.Email,
		Status:       &user.Status,
		Active:       &active,
		Operation:    "update",
		ChangesCount: changesCount,
	}, nil)

	return nil
}

func (s *userService) Delete(ctx context.Context, uid string) error {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &uid,
		Operation: "delete",
	}, nil)

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return domainerrors.ErrUserDeleteFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	// Soft delete via repository
	err = s.userRepo.Delete(ctx, user)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	// Invalidate cache for deleted user
	_ = s.resolvers.User().Invalidate(ctx, param.WithUIDs(user.UID), param.WithIDs(user.ID))

	// Publish user deleted event
	err = s.eventPublisher.Publish(ctx, event.EventUserDeleted, event.EventUserDeletedData{
		UserUID:  uid,
		ActorUID: uid,
	})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			Operation: "delete",
		}, err)
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &uid,
		Username:  &user.Username,
		Email:     &user.Email,
		Status:    &user.Status,
		Operation: "delete",
	}, nil)

	return nil
}

func (s *userService) GetProfile(ctx context.Context, userUID string) (*model.UserProfile, error) {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &userUID,
		Operation: "get_profile",
	}, nil)

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &userUID,
			Operation: "get_profile",
		}, err)
		return nil, domainerrors.ErrUserGetFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "get_profile",
		}, err)
		return nil, err
	}

	// Get user first to verify existence
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "get_profile",
		}, err)
		return nil, err
	}

	// Get profile
	profile, err := s.profileRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "get_profile",
		}, domainerrors.ErrProfileNotFound)
		return nil, domainerrors.ErrProfileNotFound
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &userUID,
		Username:  &user.Username,
		Operation: "get_profile",
	}, nil)

	return profile, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userUID string, opts param.UserProfileUpdateParam) error {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &userUID,
		Operation: "update_profile",
	}, nil)

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &userUID,
				Operation: "update_profile",
			}, err)
			return err
		}
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &userUID,
			Operation: "update_profile",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "update_profile",
		}, err)
		return err
	}

	// Get user to verify existence
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &userUID,
				Operation: "update_profile",
			}, err)
			return err
		}
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "update_profile",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	// Get current profile, or create if not exists
	profile, err := s.profileRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrProfileNotFound) {
			// Create profile if it doesn't exist
			profile = &model.UserProfile{
				UserID:  user.ID,
				UserUID: user.UID,
			}
			profile, err = s.profileRepo.Create(ctx, profile)
			if err != nil {
				s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
					UID:       &userUID,
					Operation: "update_profile",
				}, err)
				return err
			}
		} else {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &userUID,
				Operation: "update_profile",
			}, domainerrors.ErrProfileNotFound)
			return domainerrors.ErrProfileNotFound
		}
	}

	// Update fields
	if opts.FirstName != nil {
		profile.FirstName = *opts.FirstName
	}
	if opts.LastName != nil {
		profile.LastName = *opts.LastName
	}
	if opts.Bio != nil {
		profile.Bio = *opts.Bio
	}
	if opts.Attributes != nil {
		profile.Attributes = opts.Attributes
	}
	// Handle avatar file if opts.Avatar is provided
	if len(opts.Avatar) > 0 {
		// Log that avatar was provided but not processed
		// Avatar handling requires UserFileService dependency to be added to userService
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "update_profile",
		}, errors.New("avatar file handling not yet implemented - requires UserFileService integration"))
		// Continue without updating avatar - this is a non-breaking change
	}

	// Save changes
	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "update_profile",
		}, err)
		return err
	}

	// Publish user update profile event
	err = s.eventPublisher.Publish(ctx, event.EventUserUpdateProfile, event.EventUserUpdateProfileData{
		UserUID:  userUID,
		ActorUID: userUID,
	})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "update_profile",
		}, err)
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &userUID,
		Username:  &user.Username,
		Operation: "update_profile",
	}, nil)

	return nil
}

func (s *userService) SetPin(ctx context.Context, userUID, pin string) error {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &userUID,
		Operation: "set_pin",
	}, nil)

	// Validate PIN
	if pin == "" {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "set_pin",
		}, domainerrors.ErrPinInvalid)
		return domainerrors.ErrPinInvalid
	}

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &userUID,
			Operation: "set_pin",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "set_pin",
		}, err)
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "set_pin",
		}, err)
		return err
	}

	// Check if user is deleted
	if user.IsDeleted() {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "set_pin",
		}, domainerrors.ErrUserDeleted)
		return domainerrors.ErrUserDeleted
	}

	// Hash PIN
	hashedPin, err := s.pinHasher.Hash(pin)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "set_pin",
		}, err)
		return err
	}

	// Check if PIN already exists
	existingPin, err := s.pinRepo.GetByUserID(ctx, user.ID)
	isNewPIN := errors.Is(err, domainerrors.ErrUserNotFound) || errors.Is(err, domainerrors.ErrPinNotSet)
	if err != nil && !isNewPIN {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "set_pin",
		}, err)
		return err
	}

	if isNewPIN {
		// Create new PIN
		userPin := &model.UserPin{
			UserID:  user.ID,
			UserUID: user.UID,
			Code:    hashedPin,
		}
		_, err = s.pinRepo.Create(ctx, userPin)
		if err != nil {
			s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
				UID:       &userUID,
				Operation: "set_pin",
			}, err)
			return err
		}
	} else {
		// Update existing PIN
		existingPin.Code = hashedPin
		err = s.pinRepo.Update(ctx, existingPin)
		if err != nil {
			s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
				UID:       &userUID,
				Operation: "set_pin",
			}, err)
			return err
		}
	}

	// Publish user update pin event
	if isNewPIN {
		err = s.eventPublisher.Publish(ctx, event.EventUserCreatePin, event.EventUserCreatePinData{
			UserUID:  userUID,
			ActorUID: userUID,
		})
		if err != nil {
			s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
				UID:       &userUID,
				Operation: "set_pin",
			}, err)
			return err
		}
	} else {
		err = s.eventPublisher.Publish(ctx, event.EventUserUpdatePin, event.EventUserUpdatePinData{
			UserUID:  userUID,
			ActorUID: userUID,
		})
		if err != nil {
			s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
				UID:       &userUID,
				Operation: "set_pin",
			}, err)
			return err
		}
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &userUID,
		Username:  &user.Username,
		Operation: "set_pin",
	}, nil)

	return nil
}

func (s *userService) ListDevice(ctx context.Context, userUID string, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) (*model.Devices, error) {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &userUID,
		Operation: "list_device",
	}, nil)

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &userUID,
			Operation: "list_device",
		}, err)
		return nil, domainerrors.ErrUserGetFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "list_device",
		}, err)
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "list_device",
		}, err)
		return nil, err
	}

	// Set defaults for pagination
	if pagination == nil {
		pagination = &param.PaginationParam{
			Page:    util.Ptr(1),
			Limit:   util.Ptr(10),
			OrderBy: util.Ptr("created_at"),
			Sort:    util.Ptr("desc"),
		}
	}
	if filter == nil {
		filter = &param.UserDeviceListFilterParam{}
	}

	// Convert UserDeviceListFilterParam to DeviceListFilterParam
	deviceFilter := &param.DeviceListFilterParam{
		Uids:       filter.DeviceUids,
		DeviceName: filter.DeviceName,
	}

	// Get devices
	devices, err := s.deviceRepo.ListByUserID(ctx, user.ID, pagination, deviceFilter)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "list_device",
		}, err)
		return nil, err
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &userUID,
		Username:  &user.Username,
		Operation: "list_device",
	}, nil)

	return devices, nil
}

func (s *userService) RevokeDevice(ctx context.Context, userUID, deviceUID string) error {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &userUID,
		Operation: "revoke_device",
	}, nil)

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return err
	}

	// Get device
	device, err := s.deviceRepo.GetByUID(ctx, deviceUID)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return err
	}

	// Get the user-device relationship to retrieve the current session ID
	userDevice, err := s.userDeviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, device.ID)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return err
	}

	// Remove the session from token whitelist before revoking the device
	if userDevice.SessionID != "" {
		if err := s.tokenWhitelist.Remove(ctx, userUID, userDevice.SessionID); err != nil {
			// Log error but don't fail - device will still be revoked
		}
	}

	// Revoke device
	err = s.userDeviceRepo.Revoke(ctx, user.ID, device.ID)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return err
	}

	// Publish device revoked event
	err = s.eventPublisher.Publish(ctx, event.EventDeviceDeleted, event.EventDeviceDeletedData{
		UserUID:   userUID,
		DeviceUID: deviceUID,
	})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return err
	}
	err = s.eventPublisher.Publish(ctx, event.EventUserRevokeDevice, event.EventUserRevokeDeviceData{
		UserUID:   userUID,
		ActorUID:  userUID,
		DeviceUID: deviceUID,
	})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "revoke_device",
		}, err)
		return err
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &userUID,
		Username:  &user.Username,
		Operation: "revoke_device",
	}, nil)

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userUID string, passwordParam *param.UserChangePasswordParam) error {
	s.userObserver.OnSignal(ctx, signal.SignalStart, signal.UserSignal{
		UID:       &userUID,
		Operation: "change_password",
	}, nil)

	// Validate input
	if passwordParam.CurrentPassword == "" {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, domainerrors.ErrInvalidCurrentPassword)
		return domainerrors.ErrInvalidCurrentPassword
	}
	if passwordParam.NewPassword == "" {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, domainerrors.ErrInvalidPassword)
		return domainerrors.ErrInvalidPassword
	}

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &userUID,
				Operation: "change_password",
			}, err)
			return err
		}
		s.userObserver.OnSignal(ctx, signal.SignalError, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, err)
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
				UID:       &userUID,
				Operation: "change_password",
			}, err)
			return err
		}
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, err)
		return domainerrors.ErrUserUpdateFailed
	}

	// Check if user is deleted
	if user.IsDeleted() {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, domainerrors.ErrUserDeleted)
		return domainerrors.ErrUserDeleted
	}

	// Verify current password
	if !s.passwordHasher.Compare(user.Password, passwordParam.CurrentPassword) {
		s.userObserver.OnSignal(ctx, signal.SignalReject, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, domainerrors.ErrInvalidCurrentPassword)
		return domainerrors.ErrInvalidCurrentPassword
	}

	// Hash new password
	hashedPassword, err := s.passwordHasher.Hash(passwordParam.NewPassword)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, err)
		return err
	}

	// Update password
	user.Password = hashedPassword
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, err)
		return err
	}

	// Publish user update password event
	err = s.eventPublisher.Publish(ctx, event.EventUserUpdatePassword, event.EventUserUpdatePasswordData{
		UserUID:  userUID,
		ActorUID: userUID,
	})
	if err != nil {
		s.userObserver.OnSignal(ctx, signal.SignalFail, signal.UserSignal{
			UID:       &userUID,
			Operation: "change_password",
		}, err)
		return err
	}

	s.userObserver.OnSignal(ctx, signal.SignalSuccess, signal.UserSignal{
		UID:       &userUID,
		Username:  &user.Username,
		Operation: "change_password",
	}, nil)

	return nil
}

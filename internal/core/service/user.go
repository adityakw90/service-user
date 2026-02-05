package service

import (
	"context"
	"errors"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portSec "github.com/adityakw90/service-user/internal/core/port/security"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
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
) portSvc.UserService {
	return &userService{
		userRepo:       userRepo,
		profileRepo:    profileRepo,
		pinRepo:        pinRepo,
		deviceRepo:     deviceRepo,
		userDeviceRepo: userDeviceRepo,
		passwordHasher: passwordHasher,
		pinHasher:      pinHasher,
		uidGen:         uidGen,
	}
}

func (s *userService) Get(ctx context.Context, uid string) (*model.User, error) {
	if uid == "" {
		return nil, domainerrors.ErrInvalidUID
	}

	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	if user.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	return user, nil
}

func (s *userService) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserListFilterParam) (*model.Users, error) {
	// Set defaults for pagination
	if pagination == nil {
		pagination = params.NewPaginationParam(1, 10, "created_at", "desc")
	}

	users, err := s.userRepo.List(ctx, pagination, filter)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *userService) Create(ctx context.Context, param *params.UserCreateParam) (*model.User, error) {
	// Validate input
	if param.Username == "" {
		return nil, domainerrors.ErrInvalidUsername
	}
	if param.Email == "" {
		return nil, domainerrors.ErrInvalidEmail
	}
	if param.Password == "" {
		return nil, domainerrors.ErrInvalidPassword
	}

	// Check for duplicate email
	_, err := s.userRepo.GetByEmail(ctx, param.Email)
	if err == nil {
		return nil, domainerrors.ErrDuplicateEmail
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	// Check for duplicate username
	_, err = s.userRepo.GetByUsername(ctx, param.Username)
	if err == nil {
		return nil, domainerrors.ErrDuplicateUsername
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(param.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		UID:      s.uidGen.New(),
		Username: param.Username,
		Email:    param.Email,
		Password: hashedPassword,
		Status:   model.UserStatusActive,
	}

	user, err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Create empty profile
	profile := &model.UserProfile{
		UserUID: user.UID,
	}
	_, err = s.profileRepo.Create(ctx, profile)
	if err != nil {
		// Log error but don't fail user creation
	}

	return user, nil
}

func (s *userService) Update(ctx context.Context, uid string, param *params.UserUpdateParam) error {
	// Get user
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return domainerrors.ErrUserDeleted
	}

	// Update username if provided
	if param.Username != nil {
		if *param.Username == "" {
			return domainerrors.ErrInvalidUsername
		}
		// Check for duplicate
		existing, err := s.userRepo.GetByUsername(ctx, *param.Username)
		if err == nil && existing.UID != uid {
			return domainerrors.ErrDuplicateUsername
		} else if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		user.Username = *param.Username
	}

	// Update email if provided
	if param.Email != nil {
		if *param.Email == "" {
			return domainerrors.ErrInvalidEmail
		}
		// Check for duplicate
		existing, err := s.userRepo.GetByEmail(ctx, *param.Email)
		if err == nil && existing.UID != uid {
			return domainerrors.ErrDuplicateEmail
		} else if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		user.Email = *param.Email
	}

	// Update password if provided
	if param.Password != nil {
		if *param.Password == "" {
			return domainerrors.ErrInvalidPassword
		}
		hashedPassword, err := s.passwordHasher.Hash(*param.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	// Update status if provided
	if param.Status != nil {
		user.Status = *param.Status
	}

	// Save changes
	return s.userRepo.Update(ctx, user)
}

func (s *userService) Delete(ctx context.Context, uid string) error {
	// Get user
	user, err := s.userRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// Soft delete via repository
	return s.userRepo.Delete(ctx, user)
}

func (s *userService) GetProfile(ctx context.Context, userUID string) (*model.UserProfile, error) {
	// Get user first to verify existence
	user, err := s.userRepo.GetByUID(ctx, userUID)
	if err != nil {
		return nil, err
	}

	// Get profile
	profile, err := s.profileRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return nil, domainerrors.ErrProfileNotFound
	}

	return profile, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userUID string, opts params.UserProfileUpdateParam) error {
	// Get user to verify existence
	user, err := s.userRepo.GetByUID(ctx, userUID)
	if err != nil {
		return err
	}

	// Get current profile
	profile, err := s.profileRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		return domainerrors.ErrProfileNotFound
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
	// TODO: Handle avatar file if opts.Avatar is provided

	// Save changes
	return s.profileRepo.Update(ctx, profile)
}

func (s *userService) SetPin(ctx context.Context, userUID, pin string) error {
	// Validate PIN
	if pin == "" {
		return domainerrors.ErrPinInvalid
	}

	// Get user
	user, err := s.userRepo.GetByUID(ctx, userUID)
	if err != nil {
		return err
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return domainerrors.ErrUserDeleted
	}

	// Hash PIN
	hashedPin, err := s.pinHasher.Hash(pin)
	if err != nil {
		return err
	}

	// Check if PIN already exists
	existingPin, err := s.pinRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			// Create new PIN
			userPin := &model.UserPin{
				UserUID: user.UID,
				Code:    hashedPin,
			}
			_, err = s.pinRepo.Create(ctx, userPin)
			return err
		}
		return err
	}

	// Update existing PIN
	existingPin.Code = hashedPin
	return s.pinRepo.Update(ctx, existingPin)
}

func (s *userService) ListDevice(ctx context.Context, userUID string, opts params.UserDeviceListFilterParam) (*model.Devices, error) {
	// Get user
	user, err := s.userRepo.GetByUID(ctx, userUID)
	if err != nil {
		return nil, err
	}

	// Set defaults for pagination
	pagination := params.NewPaginationParam(1, 10, "created_at", "desc")

	// Get devices
	devices, err := s.deviceRepo.ListByUserID(ctx, user.ID, pagination, &params.DeviceListFilterParam{})
	if err != nil {
		return nil, err
	}

	return devices, nil
}

func (s *userService) RevokeDevice(ctx context.Context, userUID, deviceUID string) error {
	// Get user
	user, err := s.userRepo.GetByUID(ctx, userUID)
	if err != nil {
		return err
	}

	// Get device
	device, err := s.deviceRepo.GetByUID(ctx, deviceUID)
	if err != nil {
		return err
	}

	// Revoke device
	return s.userDeviceRepo.Revoke(ctx, user.ID, device.ID)
}

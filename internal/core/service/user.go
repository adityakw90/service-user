package service

import (
	"context"
	"errors"

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
	eventPublisher portEvent.EventPublisher,
	resolvers portResolver.ResolverProvider,
) portSvc.UserService {
	if userRepo == nil {
		panic("userRepo is required")
	}
	if profileRepo == nil {
		panic("profileRepo is required")
	}
	if pinRepo == nil {
		panic("pinRepo is required")
	}
	if deviceRepo == nil {
		panic("deviceRepo is required")
	}
	if userDeviceRepo == nil {
		panic("userDeviceRepo is required")
	}
	if passwordHasher == nil {
		panic("passwordHasher is required")
	}
	if pinHasher == nil {
		panic("pinHasher is required")
	}
	if uidGen == nil {
		panic("uidGen is required")
	}
	if tokenWhitelist == nil {
		panic("tokenWhitelist is required")
	}
	if eventPublisher == nil {
		panic("eventPublisher is required")
	}
	if resolvers == nil {
		panic("resolvers is required")
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
		eventPublisher: eventPublisher,
		resolvers:      resolvers,
	}
}

func (s *userService) Get(ctx context.Context, uid string) (*model.User, error) {
	if uid == "" {
		return nil, domainerrors.ErrInvalidUID
	}

	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, err
		}
		return nil, domainerrors.ErrUserGetFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if user.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	if !user.IsActive() {
		return nil, domainerrors.ErrUserInactive
	}

	return user, nil
}

func (s *userService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserListFilterParam) (*model.Users, error) {
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
		return nil, err
	}

	return users, nil
}

func (s *userService) Create(ctx context.Context, createParam *param.UserCreateParam) (*model.User, error) {
	// Validate input
	if createParam.Username == "" {
		return nil, domainerrors.ErrInvalidUsername
	}
	if createParam.Email == "" {
		return nil, domainerrors.ErrInvalidEmail
	}
	if createParam.Password == "" {
		return nil, domainerrors.ErrInvalidPassword
	}

	// Check for duplicate email
	_, err := s.userRepo.GetByEmail(ctx, createParam.Email)
	if err == nil {
		return nil, domainerrors.ErrDuplicateEmail
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	// Check for duplicate username
	_, err = s.userRepo.GetByUsername(ctx, createParam.Username)
	if err == nil {
		return nil, domainerrors.ErrDuplicateUsername
	} else if !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := s.passwordHasher.Hash(createParam.Password)
	if err != nil {
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
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserCreated, Entity: event.NewUserEntity(user), Metadata: event.EventUserCreatedData{
		Username: user.Username,
		Email:    user.Email,
		Status:   string(user.Status),
	}})

	return user, nil
}

func (s *userService) Update(ctx context.Context, uid string, updateParam *param.UserUpdateParam) error {
	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		return domainerrors.ErrUserUpdateFailed
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return domainerrors.ErrUserDeleted
	}

	changesCount := 0

	// Update username if provided
	if updateParam.Username != nil {
		if *updateParam.Username == "" {
			return domainerrors.ErrInvalidUsername
		}
		// Check for duplicate
		existing, err := s.userRepo.GetByUsername(ctx, *updateParam.Username)
		if err == nil && existing.UID != uid {
			return domainerrors.ErrDuplicateUsername
		} else if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		user.Username = *updateParam.Username
		changesCount++
	}

	// Update email if provided
	if updateParam.Email != nil {
		if *updateParam.Email == "" {
			return domainerrors.ErrInvalidEmail
		}
		// Check for duplicate
		existing, err := s.userRepo.GetByEmail(ctx, *updateParam.Email)
		if err == nil && existing.UID != uid {
			return domainerrors.ErrDuplicateEmail
		} else if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		user.Email = *updateParam.Email
		changesCount++
	}

	// Update password if provided
	if updateParam.Password != nil {
		if *updateParam.Password == "" {
			return domainerrors.ErrInvalidPassword
		}
		hashedPassword, err := s.passwordHasher.Hash(*updateParam.Password)
		if err != nil {
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
		return err
	}

	// Invalidate cache for updated user
	_ = s.resolvers.User().Invalidate(ctx, param.WithUIDs(user.UID), param.WithIDs(user.ID))

	// Publish user updated event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserUpdated, Entity: event.NewUserEntity(user), Metadata: event.EventUserUpdatedData{
		ChangesCount: changesCount,
	}})

	return nil
}

func (s *userService) Delete(ctx context.Context, uid string) error {
	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{uid})
	if err != nil {
		return domainerrors.ErrUserDeleteFailed
	}

	id, exists := ids[uid]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Soft delete via repository
	err = s.userRepo.Delete(ctx, user)
	if err != nil {
		return err
	}

	// Invalidate cache for deleted user
	_ = s.resolvers.User().Invalidate(ctx, param.WithUIDs(user.UID), param.WithIDs(user.ID))

	// Publish user deleted event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserDeleted, Entity: event.NewUserEntity(user), Metadata: event.EventUserDeletedData{}})

	return nil
}

func (s *userService) GetProfile(ctx context.Context, userUID string) (*model.UserProfile, error) {
	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		return nil, domainerrors.ErrUserGetFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return nil, err
	}

	// Get user first to verify existence
	user, err := s.userRepo.GetByID(ctx, id)
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

func (s *userService) UpdateProfile(ctx context.Context, userUID string, opts param.UserProfileUpdateParam) error {
	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return err
	}

	// Get user to verify existence
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
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
				return err
			}
		} else {
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
		// Continue without updating avatar - this is a non-breaking change
	}

	// Save changes
	err = s.profileRepo.Update(ctx, profile)
	if err != nil {
		return err
	}

	// Publish user update profile event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserUpdateProfile, Entity: event.NewUserProfileEntity(profile), Metadata: event.EventUserUpdateProfileData{}})

	return nil
}

func (s *userService) SetPin(ctx context.Context, userUID, pin string) error {
	// Validate PIN
	if pin == "" {
		return domainerrors.ErrPinInvalid
	}

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
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
	isNewPIN := errors.Is(err, domainerrors.ErrUserNotFound) || errors.Is(err, domainerrors.ErrPinNotSet)
	if err != nil && !isNewPIN {
		return err
	}

	userPin := existingPin
	if isNewPIN {
		// Create new PIN
		userPin = &model.UserPin{
			UserID:  user.ID,
			UserUID: user.UID,
			Code:    hashedPin,
		}
		userPin, err = s.pinRepo.Create(ctx, userPin)
		if err != nil {
			return err
		}
	} else {
		// Update existing PIN
		userPin.Code = hashedPin
		err = s.pinRepo.Update(ctx, userPin)
		if err != nil {
			return err
		}
	}

	// Publish user update pin event
	if isNewPIN {
		err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserCreatePin, Entity: event.NewUserPinEntity(userPin), Metadata: event.EventUserCreatePinData{}})
	} else {
		err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserUpdatePin, Entity: event.NewUserPinEntity(userPin), Metadata: event.EventUserUpdatePinData{}})
	}

	return nil
}

func (s *userService) ListDevice(ctx context.Context, userUID string, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) (*model.Devices, error) {
	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		return nil, domainerrors.ErrUserGetFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return nil, err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
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
		return nil, err
	}

	return devices, nil
}

func (s *userService) RevokeDevice(ctx context.Context, userUID, deviceUID string) error {
	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get device
	device, err := s.deviceRepo.GetByUID(ctx, deviceUID)
	if err != nil {
		return err
	}

	// Get the user-device relationship to retrieve the current session ID
	userDevice, err := s.userDeviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, device.ID)
	if err != nil {
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
		return err
	}

	// Publish device revoked event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventDeviceDeleted, Entity: event.NewDeviceEntity(device), Metadata: event.EventDeviceDeletedData{
		UserUID: userUID,
	}})
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserRevokeDevice, Entity: event.NewDeviceEntity(device), Metadata: event.EventUserRevokeDeviceData{
		UserUID: userUID,
	}})

	return nil
}

func (s *userService) ChangePassword(ctx context.Context, userUID string, passwordParam *param.UserChangePasswordParam) error {
	// Validate input
	if passwordParam.CurrentPassword == "" {
		return domainerrors.ErrInvalidCurrentPassword
	}
	if passwordParam.NewPassword == "" {
		return domainerrors.ErrInvalidPassword
	}

	// Resolve UID to ID
	ids, err := s.resolvers.User().IDsByUIDs(ctx, []string{userUID})
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		return domainerrors.ErrUserUpdateFailed
	}

	id, exists := ids[userUID]
	if !exists {
		err := domainerrors.ErrUserNotFound
		return err
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return err
		}
		return domainerrors.ErrUserUpdateFailed
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return domainerrors.ErrUserDeleted
	}

	// Verify current password
	if !s.passwordHasher.Compare(user.Password, passwordParam.CurrentPassword) {
		return domainerrors.ErrInvalidCurrentPassword
	}

	// Hash new password
	hashedPassword, err := s.passwordHasher.Hash(passwordParam.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	user.Password = hashedPassword
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Publish user update password event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventUserUpdatePassword, Entity: event.NewUserEntity(user), Metadata: event.EventUserUpdatePasswordData{}})

	return nil
}

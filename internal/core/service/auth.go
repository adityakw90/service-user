package service

import (
	"context"
	"errors"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/port"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portSec "github.com/adityakw90/service-user/internal/core/port/security"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
)

type authService struct {
	userRepo       repository.UserRepository
	deviceRepo     repository.DeviceRepository
	userDeviceRepo repository.UserDeviceRepository
	pinRepo        repository.UserPinRepository
	passwordHasher portSec.Hasher
	pinHasher      portSec.Hasher
	tokenGen       portSec.TokenGenerator
	uidGen         portSec.UIDGenerator
	oauthProvider  port.OAuthProvider
	tokenWhitelist portSec.TokenStore
	tokenBlacklist portSec.TokenStore
	eventPublisher port.EventPublisher
}

func NewAuthService(
	userRepo repository.UserRepository,
	deviceRepo repository.DeviceRepository,
	userDeviceRepo repository.UserDeviceRepository,
	pinRepo repository.UserPinRepository,
	passwordHasher portSec.Hasher,
	pinHasher portSec.Hasher,
	tokenGen portSec.TokenGenerator,
	uidGen portSec.UIDGenerator,
	oauthProvider port.OAuthProvider,
	tokenWhitelist portSec.TokenStore,
	tokenBlacklist portSec.TokenStore,
	eventPublisher port.EventPublisher,
) portSvc.AuthService {
	return &authService{
		userRepo:       userRepo,
		deviceRepo:     deviceRepo,
		userDeviceRepo: userDeviceRepo,
		pinRepo:        pinRepo,
		passwordHasher: passwordHasher,
		pinHasher:      pinHasher,
		tokenGen:       tokenGen,
		uidGen:         uidGen,
		oauthProvider:  oauthProvider,
		tokenWhitelist: tokenWhitelist,
		tokenBlacklist: tokenBlacklist,
		eventPublisher: eventPublisher,
	}
}

func (s *authService) Authenticate(ctx context.Context, payload *params.AuthParams) (*model.Token, error) {
	var userDevice *model.UserDevice
	var device *model.Device

	user, err := s.findUser(ctx, payload.IdentifierType, payload.Identifier)

	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, domainerrors.ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is deleted
	if user.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	// Check if user is active
	if user.Status != model.UserStatusActive {
		return nil, domainerrors.ErrUserInactive
	}

	// check device
	if payload.DeviceFingerprint != "" {
		device, err = s.findOrCreateDevice(
			ctx,
			payload.DeviceName,
			payload.DeviceFingerprint,
		)
		if err != nil {
			// allow error and log this
		}
		if device != nil {
			userDevice, err = s.findOrCreateUserDevice(
				ctx,
				user,
				device,
				payload.DeviceIP,
			)
			if err != nil {
				// allow error and log this
			}
		}
	}

	// generate sessionID
	sid := s.uidGen.New()
	extaClaims := map[string]any{}
	if device != nil {
		deviceClaim := map[string]any{
			"uid":  device.UID,
			"name": device.DeviceName,
		}
		if userDevice != nil {
			deviceClaim["ip_address"] = userDevice.IPAddress
		}
		extaClaims["device"] = deviceClaim
	}
	if payload.Extra != nil {
		for k, v := range payload.Extra {
			extaClaims[k] = v
		}
	}

	// Generate tokens
	accessToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           model.TokenTypeAccess,
		Identifier:     payload.Identifier,
		IdentifierType: payload.IdentifierType,
		Extra:          extaClaims,
	})
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           model.TokenTypeRefresh,
		Identifier:     payload.Identifier,
		IdentifierType: payload.IdentifierType,
		Extra:          extaClaims,
	})
	if err != nil {
		return nil, err
	}

	return &model.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *authService) GoogleOAuth(ctx context.Context, redirectURI string) (string, error) {
	// Generate state parameter for OAuth flow
	state := s.uidGen.New()

	// Get authorization URL from OAuth provider
	authURL, err := s.oauthProvider.GetAuthorizationURL(ctx, redirectURI, state)
	if err != nil {
		return "", domainerrors.ErrOAuthExchangeFailed
	}

	return authURL, nil
}

func (s *authService) HandleGoogleOAuth(ctx context.Context, code, redirectURI string) (*model.Token, error) {
	// Exchange code for tokens
	tokens, err := s.oauthProvider.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		return nil, domainerrors.ErrOAuthExchangeFailed
	}

	// Get user info from OAuth provider
	userInfo, err := s.oauthProvider.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		return nil, domainerrors.ErrOAuthUserInfoFailed
	}

	// Find or create user
	user, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			// Create new user from OAuth info
			user = &model.User{
				UID:      s.uidGen.New(),
				Username: userInfo.DisplayName(),
				Email:    userInfo.Email,
				Status:   model.UserStatusActive,
			}
			user, err = s.userRepo.Create(ctx, user)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Generate session ID
	sid := s.uidGen.New()

	// Generate tokens
	accessToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           model.TokenTypeAccess,
		Identifier:     user.Email,
		IdentifierType: "email",
		Extra: map[string]any{
			"oauth_provider": "google",
		},
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           model.TokenTypeRefresh,
		Identifier:     user.Email,
		IdentifierType: "email",
		Extra: map[string]any{
			"oauth_provider": "google",
		},
	})
	if err != nil {
		return nil, err
	}

	// Add refresh token to whitelist
	if err := s.tokenWhitelist.Add(ctx, user.UID, sid); err != nil {
		// Log error but don't fail
	}

	// Publish auth event
	if s.eventPublisher != nil {
		// TODO: publish event
	}

	return &model.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.Token, error) {
	// Validate refresh token
	claims, err := s.tokenGen.ValidateToken(refreshToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrTokenExpired) {
			return nil, domainerrors.ErrTokenExpired
		}
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if it's a refresh token
	if !claims.IsRefresh() {
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if token is in whitelist
	allowed, err := s.tokenWhitelist.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil || !allowed {
		return nil, domainerrors.ErrTokenRevoked
	}

	// Get user to verify they exist and are active
	user, err := s.userRepo.GetByUID(ctx, claims.Uid)
	if err != nil {
		return nil, domainerrors.ErrTokenInvalid
	}

	if user.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	if user.Status != model.UserStatusActive {
		return nil, domainerrors.ErrUserInactive
	}

	// Generate new session ID
	newSid := s.uidGen.New()

	// Generate new tokens
	newAccessToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            claims.Uid,
		Sid:            newSid,
		Type:           model.TokenTypeAccess,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
		Extra:          claims.Extra,
	})
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            claims.Uid,
		Sid:            newSid,
		Type:           model.TokenTypeRefresh,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
		Extra:          claims.Extra,
	})
	if err != nil {
		return nil, err
	}

	// Add new refresh token to whitelist
	if err := s.tokenWhitelist.Add(ctx, claims.Uid, newSid); err != nil {
		// Log error but don't fail
	}

	// Remove old refresh token from whitelist
	if err := s.tokenWhitelist.Remove(ctx, claims.Uid, claims.Sid); err != nil {
		// Log error but don't fail
	}

	return &model.Token{
		Access:  newAccessToken,
		Refresh: newRefreshToken,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, accessToken string) (*model.TokenClaims, error) {
	// Validate token signature and expiration
	claims, err := s.tokenGen.ValidateToken(accessToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrTokenExpired) {
			return nil, domainerrors.ErrTokenExpired
		}
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if it's an access token
	if !claims.IsAccess() {
		return nil, domainerrors.ErrTokenInvalid
	}

	// TODO: Optionally check blacklist if needed

	return claims, nil
}

func (s *authService) RevokeToken(ctx context.Context, token string, tokenType string) error {
	// Validate token to extract claims
	claims, err := s.tokenGen.ValidateToken(token)
	if err != nil {
		return domainerrors.ErrTokenInvalid
	}

	// Remove from whitelist
	if err := s.tokenWhitelist.Remove(ctx, claims.Uid, claims.Sid); err != nil {
		return err
	}

	return nil
}

func (s *authService) VerifyPin(ctx context.Context, userUid string, pin string) (bool, error) {
	// Get user
	user, err := s.userRepo.GetByUID(ctx, userUid)
	if err != nil {
		return false, err
	}

	// Get PIN for user
	userPin, err := s.pinRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return false, domainerrors.ErrPinNotSet
		}
		return false, err
	}

	// Check if PIN is set
	if !userPin.IsSet() {
		return false, domainerrors.ErrPinNotSet
	}

	// Verify PIN hash
	if !s.pinHasher.Compare(pin, userPin.Code) {
		return false, domainerrors.ErrPinInvalid
	}

	return true, nil
}

func (s *authService) findUser(ctx context.Context, identifierType string, identifier string) (*model.User, error) {
	switch identifierType {
	case "email":
		return s.userRepo.GetByEmail(ctx, identifier)
	case "username":
		return s.userRepo.GetByUsername(ctx, identifier)
	default:
		// Try email first, then username
		user, err := s.userRepo.GetByEmail(ctx, identifier)
		if err != nil {
			user, err = s.userRepo.GetByUsername(ctx, identifier)
		}
		return user, err
	}
}

func (s *authService) findOrCreateDevice(ctx context.Context, name string, fingerprint string) (*model.Device, error) {
	device, err := s.deviceRepo.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		if err == domainerrors.ErrDeviceNotFound {
			return s.deviceRepo.Create(ctx, &model.Device{
				DeviceName:        name,
				DeviceFingerprint: fingerprint,
			})
		}
		return nil, err
	}
	return device, nil
}

func (s *authService) findOrCreateUserDevice(ctx context.Context, user *model.User, device *model.Device, ipAddress string) (*model.UserDevice, error) {
	userDevice, err := s.userDeviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, device.ID)
	if err != nil {
		if err == domainerrors.ErrUserDeviceNotFound {
			return s.userDeviceRepo.Create(ctx, &model.UserDevice{
				UserID:    user.ID,
				DeviceID:  device.ID,
				IPAddress: ipAddress,
			})
		}
		return nil, err
	}
	return userDevice, nil
}

package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	domainEvent "github.com/adityakw90/service-user/internal/core/domain/event"
	domainModel "github.com/adityakw90/service-user/internal/core/domain/model"
	domainParam "github.com/adityakw90/service-user/internal/core/domain/param"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	portExecutor "github.com/adityakw90/service-user/internal/core/port/executor"
	portOAuth "github.com/adityakw90/service-user/internal/core/port/oauth"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portSec "github.com/adityakw90/service-user/internal/core/port/security"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
	"github.com/adityakw90/service-user/pkg/util"
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
	oauthProvider  portOAuth.OAuthProvider
	tokenWhitelist portSec.TokenStore
	tokenBlacklist portSec.TokenStore
	executor       portExecutor.Executor
	eventPublisher portEvent.EventPublisher
	attemptTracker portSec.AttemptTracker
	rateLimiter    portSec.RateLimiter
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
	oauthProvider portOAuth.OAuthProvider,
	tokenWhitelist portSec.TokenStore,
	tokenBlacklist portSec.TokenStore,
	executor portExecutor.Executor,
	eventPublisher portEvent.EventPublisher,
	attemptTracker portSec.AttemptTracker,
	rateLimiter portSec.RateLimiter,
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
		executor:       executor,
		eventPublisher: eventPublisher,
		attemptTracker: attemptTracker,
		rateLimiter:    rateLimiter,
	}
}

func (s *authService) Authenticate(ctx context.Context, payload *domainParam.AuthParams) (*domainModel.Token, error) {
	var userDevice *domainModel.UserDevice
	var device *domainModel.Device

	// Check IP rate limiting first (if IP provided)
	if payload.DeviceIP != nil && *payload.DeviceIP != "" {
		allowed, err := s.rateLimiter.Acquire(ctx, *payload.DeviceIP)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, domainerrors.ErrRateLimitExceeded
		}
	}

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
	if user.Status != domainModel.UserStatusActive {
		return nil, domainerrors.ErrUserInactive
	}

	// Check if account is locked (before password verification)
	locked, err := s.attemptTracker.IsLocked(ctx, user.UID)
	if err != nil {
		return nil, err
	}
	if locked {
		// Publish login locked event
		s.executor.DoAsync(ctx, "auth.publish.locked", func(newCtx context.Context) error {
			eventMessage, errExc := domainEvent.NewMessage(
				domainEvent.EventLoginLocked,
				domainEvent.NewUserEntity(user),
				domainEvent.EventLoginLockedData{
					Identifier:     payload.Identifier,
					IdentifierType: payload.IdentifierType,
					FailureReason:  "Account is locked",
				},
			)
			if errExc != nil {
				return errExc
			}
			if errExc := s.eventPublisher.Publish(newCtx, eventMessage); errExc != nil {
				return errExc
			}
			return nil
		})
		return nil, domainerrors.ErrAccountLockedOut
	}

	// Verify password
	if !s.passwordHasher.Compare(user.Password, payload.Password) {
		// Track failed attempt
		_ = s.attemptTracker.Track(ctx, user.UID)

		// Publish login failed event
		s.executor.DoAsync(ctx, "auth.publish.failed", func(newCtx context.Context) error {
			eventMessage, errExc := domainEvent.NewMessage(
				domainEvent.EventLoginFailed,
				domainEvent.NewUserEntity(user),
				domainEvent.EventLoginFailedData{
					Identifier:     payload.Identifier,
					IdentifierType: payload.IdentifierType,
					FailureReason:  "invalid_credentials",
				},
			)
			if errExc != nil {
				return errExc
			}
			if errExc := s.eventPublisher.Publish(newCtx, eventMessage); errExc != nil {
				return errExc
			}
			return nil
		})

		return nil, domainerrors.ErrInvalidCredentials
	}

	// Reset failed attempts on successful authentication
	_ = s.attemptTracker.Reset(ctx, user.UID)

	// Generate session ID first - needed for device tracking
	sid := s.uidGen.New()

	// check device
	if payload.DeviceFingerprint != nil && payload.DeviceName != nil {
		device, _ = s.findOrCreateDevice(
			ctx,
			*payload.DeviceName,
			*payload.DeviceFingerprint,
		)
		if device != nil {
			var deviceIP string
			if payload.DeviceIP != nil {
				deviceIP = *payload.DeviceIP
			}
			userDevice, _ = s.findOrCreateUserDevice(
				ctx,
				user,
				device,
				deviceIP,
				sid, // Pass session ID to track this login
			)
		}
	}

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
		for k, v := range *payload.Extra {
			extaClaims[k] = v
		}
	}

	// Generate tokens
	accessToken, err := s.tokenGen.GenerateToken(&domainModel.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           domainModel.TokenTypeAccess,
		Identifier:     payload.Identifier,
		IdentifierType: payload.IdentifierType,
		Extra:          extaClaims,
	})
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.tokenGen.GenerateToken(&domainModel.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           domainModel.TokenTypeRefresh,
		Identifier:     payload.Identifier,
		IdentifierType: payload.IdentifierType,
		Extra:          extaClaims,
	})
	if err != nil {
		return nil, err
	}

	// Add refresh token to whitelist
	_ = s.tokenWhitelist.Add(ctx, user.UID, sid)

	s.executor.DoAsync(ctx, "auth.publish", func(newCtx context.Context) error {
		// Publish login event
		loginEventData := domainEvent.EventLoginData{
			Identifier:     payload.Identifier,
			IdentifierType: payload.IdentifierType,
		}
		if device != nil {
			loginEventData.DeviceUID = &device.UID
			loginEventData.DeviceName = &device.DeviceName
		}
		if userDevice != nil {
			loginEventData.IPAddress = &userDevice.IPAddress
		}

		eventMessage, err := domainEvent.NewMessage(
			domainEvent.EventLogin,
			domainEvent.NewUserEntity(user),
			&loginEventData,
		)
		if err != nil {
			return err
		}
		if err := s.eventPublisher.Publish(newCtx, eventMessage); err != nil {
			return err
		}
		return nil
	})

	return &domainModel.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *authService) GoogleOAuth(ctx context.Context, redirectURI string) (string, string, error) {
	// Generate state parameter for OAuth flow
	state := s.uidGen.New()

	// Get authorization URL from OAuth provider
	authURL, err := s.oauthProvider.GetAuthorizationURL(ctx, state, redirectURI)
	if err != nil {
		return "", "", err
	}

	return authURL, state, nil
}

func (s *authService) HandleGoogleOAuth(ctx context.Context, code, state, redirectURI string) (*domainModel.Token, error) {
	// Exchange code for tokens
	tokens, err := s.oauthProvider.ExchangeCode(ctx, code, state, redirectURI)
	if err != nil {
		return nil, err
	}

	// Get user info from OAuth provider
	userInfo, err := s.oauthProvider.GetUserInfo(ctx, tokens)
	if err != nil {
		return nil, err
	}

	// Find or create user
	user, err := s.userRepo.GetByEmail(ctx, userInfo.Email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			// generate password
			randPass, err := util.GenerateRandomPassword(8)
			if err != nil {
				return nil, err
			}
			hashedPassword, err := s.passwordHasher.Hash(randPass)
			if err != nil {
				return nil, err
			}
			// Create new user from OAuth info
			user = &domainModel.User{
				UID:      s.uidGen.New(),
				Username: s.generateUsername(ctx, userInfo),
				Password: hashedPassword,
				Email:    userInfo.Email,
				Status:   domainModel.UserStatusActive,
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
	accessToken, err := s.tokenGen.GenerateToken(&domainModel.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           domainModel.TokenTypeAccess,
		Identifier:     user.Email,
		IdentifierType: "email",
		Extra: map[string]any{
			"oauth_provider": "google",
		},
	})
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenGen.GenerateToken(&domainModel.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           domainModel.TokenTypeRefresh,
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

	// Publish OAuth login event
	s.executor.DoAsync(ctx, "auth.publish.oauth", func(newCtx context.Context) error {
		eventMessage, err := domainEvent.NewMessage(
			domainEvent.EventOAuthLogin,
			domainEvent.NewUserEntity(user),
			&domainEvent.EventOAuthLoginData{
				Provider: "google",
			},
		)
		if err != nil {
			return err
		}
		if err := s.eventPublisher.Publish(newCtx, eventMessage); err != nil {
			return err
		}
		return nil
	})

	return &domainModel.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*domainModel.Token, error) {
	// Validate refresh token
	claims, err := s.tokenGen.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Check if it's a refresh token
	if !claims.IsRefresh() {
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if token is in whitelist
	allowed, err := s.tokenWhitelist.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domainerrors.ErrTokenRevoked
	}

	// Get user to verify they exist and are active
	user, err := s.userRepo.GetByUID(ctx, claims.Uid)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, domainerrors.ErrTokenInvalid
		}
		return nil, err
	}

	if user.IsDeleted() {
		return nil, domainerrors.ErrUserDeleted
	}

	if user.Status != domainModel.UserStatusActive {
		return nil, domainerrors.ErrUserInactive
	}

	// Generate new session ID
	newSid := s.uidGen.New()

	// Update device session ID if this token is associated with a device
	if claims.Extra != nil {
		if deviceClaim, ok := claims.Extra["device"].(map[string]interface{}); ok {
			if deviceUIDStr, ok := deviceClaim["uid"].(string); ok {
				// Get device by UID
				device, err := s.deviceRepo.GetByUID(ctx, deviceUIDStr)
				if err == nil {
					// Update the session ID for this user-device pair
					if err := s.userDeviceRepo.UpdateSessionID(ctx, user.ID, device.ID, newSid); err != nil {
						// Log error but don't fail refresh
					}
				}
			}
		}
	}

	// Generate new tokens
	newAccessToken, err := s.tokenGen.GenerateToken(&domainModel.TokenClaims{
		Uid:            claims.Uid,
		Sid:            newSid,
		Type:           domainModel.TokenTypeAccess,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
		Extra:          claims.Extra,
	})
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.tokenGen.GenerateToken(&domainModel.TokenClaims{
		Uid:            claims.Uid,
		Sid:            newSid,
		Type:           domainModel.TokenTypeRefresh,
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

	// Remove old session from whitelist (single-use refresh token for security)
	if err := s.tokenWhitelist.Remove(ctx, claims.Uid, claims.Sid); err != nil {
		// Log error but don't fail
	}

	// Publish token refresh event
	s.executor.DoAsync(ctx, "auth.publish.refresh_token", func(newCtx context.Context) error {
		eventMessage, errExc := domainEvent.NewMessage(
			domainEvent.EventTokenRefresh,
			domainEvent.NewTokenEntity(claims),
			domainEvent.EventTokenRefreshData{
				Identifier:     claims.Identifier,
				IdentifierType: claims.IdentifierType,
			},
		)
		if errExc != nil {
			return errExc
		}
		if errExc := s.eventPublisher.Publish(newCtx, eventMessage); errExc != nil {
			return errExc
		}
		return nil
	})

	return &domainModel.Token{
		Access:  newAccessToken,
		Refresh: newRefreshToken,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, accessToken string) (*domainModel.TokenClaims, error) {
	// Validate token signature and expiration
	claims, err := s.tokenGen.ValidateToken(accessToken)
	if err != nil {
		return nil, err
	}

	// Check if it's an access token
	if !claims.IsAccess() {
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if token's session is in the whitelist (for device revocation support)
	allowed, err := s.tokenWhitelist.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domainerrors.ErrTokenRevoked
	}

	// Check if token's session is in the blacklist (for immediate revocation)
	blacklisted, err := s.tokenBlacklist.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil {
		return nil, err
	}
	if !blacklisted {
		return nil, domainerrors.ErrTokenRevoked
	}

	return claims, nil
}

func (s *authService) RevokeToken(ctx context.Context, token string, tokenType string) error {
	// Validate token to extract claims
	claims, err := s.tokenGen.ValidateToken(token)
	if err != nil {
		return err
	}

	// Remove from whitelist
	if err := s.tokenWhitelist.Remove(ctx, claims.Uid, claims.Sid); err != nil {
		return err
	}

	// Add to blacklist
	if err := s.tokenBlacklist.Add(ctx, claims.Uid, claims.Sid); err != nil {
		return err
	}

	// Publish revoke event
	s.executor.DoAsync(ctx, "auth.publish.revoke_token", func(newCtx context.Context) error {
		errExc := s.eventPublisher.Publish(
			newCtx,
			domainEvent.Message{
				Type:   domainEvent.EventRevokeToken,
				Entity: domainEvent.NewTokenEntity(claims),
				Metadata: domainEvent.EventRevokeTokenData{
					Identifier:     claims.Identifier,
					IdentifierType: claims.IdentifierType,
				},
			},
		)
		if errExc != nil {
			return errExc
		}
		return nil
	})

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
	if !s.pinHasher.Compare(userPin.Code, pin) {
		// Publish PIN verify failed event
		s.executor.DoAsync(ctx, "auth.publish.pin_fail", func(newCtx context.Context) error {
			err := s.eventPublisher.Publish(
				newCtx,
				domainEvent.Message{
					Type:   domainEvent.EventPINFail,
					Entity: domainEvent.NewUserPinEntity(userPin),
					Metadata: domainEvent.EventPinFailData{
						Reason: "invalid_pin",
					},
				},
			)
			if err != nil {
				return err
			}
			return err
		})

		return false, nil
	}

	// Publish PIN verify success event
	s.executor.DoAsync(ctx, "auth.publish.pin_verified", func(newCtx context.Context) error {
		errExc := s.eventPublisher.Publish(
			newCtx,
			domainEvent.Message{
				Type:   domainEvent.EventPINVerify,
				Entity: domainEvent.NewUserPinEntity(userPin),
				Metadata: domainEvent.EventPinVerifyData{
					Success: true,
					Reason:  "pin_verified",
				},
			},
		)
		if errExc != nil {
			return errExc
		}
		return nil
	})

	return true, nil
}

func (s *authService) findUser(ctx context.Context, identifierType string, identifier string) (*domainModel.User, error) {
	switch identifierType {
	case "email":
		return s.userRepo.GetByEmail(ctx, identifier)
	case "username":
		return s.userRepo.GetByUsername(ctx, identifier)
	default:
		return nil, domainerrors.ErrInvalidIdentifierType
	}
}

func (s *authService) findOrCreateDevice(ctx context.Context, name string, fingerprint string) (*domainModel.Device, error) {
	device, err := s.deviceRepo.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		if errors.Is(err, domainerrors.ErrDeviceNotFound) {
			return s.deviceRepo.Create(ctx, &domainModel.Device{
				UID:               s.uidGen.New(),
				DeviceName:        name,
				DeviceFingerprint: fingerprint,
				CreatedAt:         time.Now().UTC(),
			})
		}
		return nil, err
	}
	return device, nil
}

func (s *authService) findOrCreateUserDevice(ctx context.Context, user *domainModel.User, device *domainModel.Device, ipAddress, sessionID string) (*domainModel.UserDevice, error) {
	_, err := s.userDeviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, device.ID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserDeviceNotFound) {
			now := time.Now().UTC()
			return s.userDeviceRepo.Create(ctx, &domainModel.UserDevice{
				UserID:       user.ID,
				DeviceID:     device.ID,
				IPAddress:    ipAddress,
				LastActiveAt: now,
				SessionID:    sessionID,
				CreatedAt:    now,
			})
		}
		return nil, err
	}
	// Update existing device with new session ID
	if err := s.userDeviceRepo.UpdateSessionID(ctx, user.ID, device.ID, sessionID); err != nil {
		// Log error but don't fail
	}
	// Re-fetch to get the updated session ID
	userDevice, err := s.userDeviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, device.ID)
	if err != nil {
		return nil, err
	}
	return userDevice, nil
}

func (s *authService) generateUsername(ctx context.Context, userInfo *domainModel.OAuthUserInfo) string {
	local := strings.SplitN(userInfo.Email, "@", 2)[0]
	re := regexp.MustCompile(`[^a-zA-Z0-9_.]+`)
	username := re.ReplaceAllString(local, "_")
	if len(username) < 3 {
		username = "user_" + username
	}
	// Check uniqueness
	for {
		existing, err := s.userRepo.GetByUsername(ctx, username)
		if err == nil && existing != nil {
			// Collision: append 4-digit random suffix
			randSuffix := s.uidGen.New()
			username = username + "_" + randSuffix[len(randSuffix)-4:]
			continue
		}
		break
	}
	return username
}

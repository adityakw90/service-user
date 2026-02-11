package service

import (
	"context"
	"errors"
	"time"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	domainSignal "github.com/adityakw90/service-user/internal/core/domain/signal"
	"github.com/adityakw90/service-user/internal/core/port"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/adityakw90/service-user/internal/core/port/observer"
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
	eventPublisher portEvent.EventPublisher
	authObserver   observer.ServiceObserver[domainSignal.AuthSignal]
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
	oauthProvider port.OAuthProvider,
	tokenWhitelist portSec.TokenStore,
	tokenBlacklist portSec.TokenStore,
	eventPublisher portEvent.EventPublisher,
	authObserver observer.ServiceObserver[domainSignal.AuthSignal],
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
		eventPublisher: eventPublisher,
		authObserver:   authObserver,
		attemptTracker: attemptTracker,
		rateLimiter:    rateLimiter,
	}
}

func (s *authService) Authenticate(ctx context.Context, payload *params.AuthParams) (*model.Token, error) {
	s.authObserver.OnSignal(ctx, domainSignal.SignalStart, domainSignal.AuthSignal{
		Identifier:        payload.Identifier,
		IdentifierType:    payload.IdentifierType,
		DeviceFingerprint: payload.DeviceFingerprint,
		DeviceIP:          payload.DeviceIP,
		DeviceName:        payload.DeviceName,
		Extra:             payload.Extra,
	}, nil)
	var userDevice *model.UserDevice
	var device *model.Device

	// Check IP rate limiting first (if IP provided)
	if payload.DeviceIP != nil && *payload.DeviceIP != "" {
		allowed, err := s.rateLimiter.Acquire(ctx, *payload.DeviceIP)
		if err != nil {
			s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
				Identifier:     payload.Identifier,
				IdentifierType: payload.IdentifierType,
			}, err)
			return nil, err
		}
		if !allowed {
			s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
				Identifier:     payload.Identifier,
				IdentifierType: payload.IdentifierType,
			}, domainerrors.ErrRateLimitExceeded)
			return nil, domainerrors.ErrRateLimitExceeded
		}
	}

	user, err := s.findUser(ctx, payload.IdentifierType, payload.Identifier)

	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
				Identifier:     payload.Identifier,
				IdentifierType: payload.IdentifierType,
			}, domainerrors.ErrInvalidCredentials)
			return nil, domainerrors.ErrInvalidCredentials
		}
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			Identifier:     payload.Identifier,
			IdentifierType: payload.IdentifierType,
		}, err)
		return nil, err
	}

	// Check if user is deleted
	if user.IsDeleted() {
		deleted := true
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:      &user.UID,
			Email:    &user.Email,
			Username: &user.Username,
			Deleted:  &deleted,
		}, domainerrors.ErrUserDeleted)
		return nil, domainerrors.ErrUserDeleted
	}

	// Check if user is active
	if user.Status != model.UserStatusActive {
		active := false
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:    &user.UID,
			Email:  &user.Email,
			Status: &user.Status,
			Active: &active,
		}, domainerrors.ErrUserInactive)
		return nil, domainerrors.ErrUserInactive
	}

	// Check if account is locked (before password verification)
	locked, err := s.attemptTracker.IsLocked(ctx, user.UID)
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &user.UID,
			Email:          &user.Email,
			Username:       &user.Username,
			IdentifierType: payload.IdentifierType,
		}, err)
		return nil, err
	}
	if locked {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &user.UID,
			Email:          &user.Email,
			Username:       &user.Username,
			IdentifierType: payload.IdentifierType,
		}, domainerrors.ErrAccountLockedOut)

		// Publish login locked event
		s.eventPublisher.Publish(ctx, event.EventLoginLocked, event.EventLoginLockedData{
			Identifier:     payload.Identifier,
			IdentifierType: payload.IdentifierType,
			FailureReason:  "Account is locked",
		})

		return nil, domainerrors.ErrAccountLockedOut
	}

	// Verify password
	if !s.passwordHasher.Compare(user.Password, payload.Password) {
		// Track failed attempt
		_ = s.attemptTracker.Track(ctx, user.UID)

		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			Identifier:     payload.Identifier,
			IdentifierType: payload.IdentifierType,
		}, domainerrors.ErrInvalidCredentials)

		// Publish login failed event
		s.eventPublisher.Publish(ctx, event.EventLoginFailed, event.EventLoginFailedData{
			Identifier:     payload.Identifier,
			IdentifierType: payload.IdentifierType,
			FailureReason:  "invalid_credentials",
		})

		return nil, domainerrors.ErrInvalidCredentials
	}

	// Reset failed attempts on successful authentication
	if resetErr := s.attemptTracker.Reset(ctx, user.UID); resetErr != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &user.UID,
			Email:          &user.Email,
			Username:       &user.Username,
			IdentifierType: payload.IdentifierType,
		}, resetErr)
	}

	// Generate session ID first - needed for device tracking
	sid := s.uidGen.New()

	// check device
	if payload.DeviceFingerprint != nil {
		device, err = s.findOrCreateDevice(
			ctx,
			*payload.DeviceName,
			*payload.DeviceFingerprint,
		)
		if err != nil {
			// Log error but don't fail authentication - devices are optional
			s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
				DeviceFingerprint: payload.DeviceFingerprint,
				DeviceName:        payload.DeviceName,
				Extra:             &map[string]any{"context": "find_or_create_device"},
			}, err)
		}
		if device != nil {
			var deviceIP string
			if payload.DeviceIP != nil {
				deviceIP = *payload.DeviceIP
			}
			userDevice, err = s.findOrCreateUserDevice(
				ctx,
				user,
				device,
				deviceIP,
				sid, // Pass session ID to track this login
			)
			if err != nil {
				// Log error but don't fail authentication
				s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
					UID:      &user.UID,
					Email:    &user.Email,
					Username: &user.Username,
					Extra:    &map[string]any{"context": "find_or_create_user_device", "device_uid": device.UID},
				}, err)
			}
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
	accessToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            user.UID,
		Sid:            sid,
		Type:           model.TokenTypeAccess,
		Identifier:     payload.Identifier,
		IdentifierType: payload.IdentifierType,
		Extra:          extaClaims,
	})
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:      &user.UID,
			Email:    &user.Email,
			Username: &user.Username,
		}, err)
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
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:      &user.UID,
			Email:    &user.Email,
			Username: &user.Username,
		}, err)
		return nil, err
	}

	// Add refresh token to whitelist
	if err := s.tokenWhitelist.Add(ctx, user.UID, sid); err != nil {
		// Log error but don't fail authentication
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:      &user.UID,
			Email:    &user.Email,
			Username: &user.Username,
			Extra:    &map[string]any{"context": "Failed to add to whitelist"},
		}, err)
	}

	active := true
	deleted := false
	s.authObserver.OnSignal(ctx, domainSignal.SignalSuccess, domainSignal.AuthSignal{
		UID:            &user.UID,
		Email:          &user.Email,
		Username:       &user.Username,
		Status:         &user.Status,
		Active:         &active,
		Deleted:        &deleted,
		Identifier:     payload.Identifier,
		IdentifierType: payload.IdentifierType,
		Extra:          payload.Extra,
	}, nil)

	// Publish login event
	s.eventPublisher.Publish(ctx, event.EventLogin, event.EventLoginData{
		Identifier:     payload.Identifier,
		IdentifierType: payload.IdentifierType,
	})

	return &model.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *authService) GoogleOAuth(ctx context.Context, redirectURI string) (string, error) {
	s.authObserver.OnSignal(ctx, domainSignal.SignalStart, domainSignal.AuthSignal{
		IdentifierType: "oauth",
		Extra:          &map[string]any{"provider": "google", "redirect_uri": redirectURI},
	}, nil)

	// Generate state parameter for OAuth flow
	state := s.uidGen.New()

	// Get authorization URL from OAuth provider
	authURL, err := s.oauthProvider.GetAuthorizationURL(ctx, redirectURI, state)
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			IdentifierType: "oauth",
		}, domainerrors.ErrOAuthExchangeFailed)
		return "", domainerrors.ErrOAuthExchangeFailed
	}

	s.authObserver.OnSignal(ctx, domainSignal.SignalSuccess, domainSignal.AuthSignal{
		IdentifierType: "oauth",
		Extra:          &map[string]any{"provider": "google", "state": state},
	}, nil)

	return authURL, nil
}

func (s *authService) HandleGoogleOAuth(ctx context.Context, code, redirectURI string) (*model.Token, error) {
	s.authObserver.OnSignal(ctx, domainSignal.SignalStart, domainSignal.AuthSignal{
		IdentifierType: "oauth",
		Extra:          &map[string]any{"provider": "google"},
	}, nil)

	// Exchange code for tokens
	tokens, err := s.oauthProvider.ExchangeCode(ctx, code, redirectURI)
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			IdentifierType: "oauth",
		}, domainerrors.ErrOAuthExchangeFailed)
		return nil, domainerrors.ErrOAuthExchangeFailed
	}

	// Get user info from OAuth provider
	userInfo, err := s.oauthProvider.GetUserInfo(ctx, tokens.AccessToken)
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			IdentifierType: "oauth",
		}, domainerrors.ErrOAuthUserInfoFailed)
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
				s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
					Email:          &user.Email,
					IdentifierType: "oauth",
				}, err)
				return nil, err
			}
		} else {
			s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
				Email:          &userInfo.Email,
				IdentifierType: "oauth",
			}, err)
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
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &user.UID,
			Email:          &user.Email,
			IdentifierType: "oauth",
		}, err)
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
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &user.UID,
			Email:          &user.Email,
			IdentifierType: "oauth",
		}, err)
		return nil, err
	}

	// Add refresh token to whitelist
	if err := s.tokenWhitelist.Add(ctx, user.UID, sid); err != nil {
		// Log error but don't fail
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &user.UID,
			Email:          &user.Email,
			IdentifierType: "oauth",
			Extra:          &map[string]any{"context": "Failed to add to whitelist"},
		}, err)
	}

	// Publish OAuth login event
	s.eventPublisher.Publish(ctx, "auth.oauth_login", event.EventOAuthLoginData{
		UserUID:  user.UID,
		Provider: "google",
	})

	active := true
	deleted := false
	s.authObserver.OnSignal(ctx, domainSignal.SignalSuccess, domainSignal.AuthSignal{
		UID:            &user.UID,
		Email:          &user.Email,
		Username:       &user.Username,
		Status:         &user.Status,
		Active:         &active,
		Deleted:        &deleted,
		Identifier:     user.Email,
		IdentifierType: "oauth",
		Extra:          &map[string]any{"oauth_provider": "google"},
	}, nil)

	return &model.Token{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.Token, error) {
	s.authObserver.OnSignal(ctx, domainSignal.SignalStart, domainSignal.AuthSignal{
		IdentifierType: "refresh",
	}, nil)

	// Validate refresh token
	claims, err := s.tokenGen.ValidateToken(refreshToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrTokenExpired) {
			s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
				IdentifierType: "refresh",
			}, domainerrors.ErrTokenExpired)
			return nil, domainerrors.ErrTokenExpired
		}
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			IdentifierType: "refresh",
		}, domainerrors.ErrTokenInvalid)
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if it's a refresh token
	if !claims.IsRefresh() {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "refresh",
		}, domainerrors.ErrTokenInvalid)
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if token is in whitelist
	allowed, err := s.tokenWhitelist.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil || !allowed {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "refresh",
		}, domainerrors.ErrTokenRevoked)
		return nil, domainerrors.ErrTokenRevoked
	}

	// Get user to verify they exist and are active
	user, err := s.userRepo.GetByUID(ctx, claims.Uid)
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "refresh",
		}, domainerrors.ErrTokenInvalid)
		return nil, domainerrors.ErrTokenInvalid
	}

	if user.IsDeleted() {
		deleted := true
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:      &claims.Uid,
			Email:    &user.Email,
			Username: &user.Username,
			Deleted:  &deleted,
		}, domainerrors.ErrUserDeleted)
		return nil, domainerrors.ErrUserDeleted
	}

	if user.Status != model.UserStatusActive {
		active := false
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:    &claims.Uid,
			Email:  &user.Email,
			Status: &user.Status,
			Active: &active,
		}, domainerrors.ErrUserInactive)
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
	newAccessToken, err := s.tokenGen.GenerateToken(&model.TokenClaims{
		Uid:            claims.Uid,
		Sid:            newSid,
		Type:           model.TokenTypeAccess,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
		Extra:          claims.Extra,
	})
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			Email:          &user.Email,
			IdentifierType: "refresh",
		}, err)
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
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			Email:          &user.Email,
			IdentifierType: "refresh",
		}, err)
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
	s.eventPublisher.Publish(ctx, event.EventTokenRefresh, event.EventTokenRefreshData{
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
	})

	active := true
	deleted := false
	s.authObserver.OnSignal(ctx, domainSignal.SignalSuccess, domainSignal.AuthSignal{
		UID:            &claims.Uid,
		Email:          &user.Email,
		Username:       &user.Username,
		Status:         &user.Status,
		Active:         &active,
		Deleted:        &deleted,
		IdentifierType: "refresh",
	}, nil)

	return &model.Token{
		Access:  newAccessToken,
		Refresh: newRefreshToken,
	}, nil
}

func (s *authService) ValidateToken(ctx context.Context, accessToken string) (*model.TokenClaims, error) {
	s.authObserver.OnSignal(ctx, domainSignal.SignalStart, domainSignal.AuthSignal{
		IdentifierType: "validate",
	}, nil)

	// Validate token signature and expiration
	claims, err := s.tokenGen.ValidateToken(accessToken)
	if err != nil {
		if errors.Is(err, domainerrors.ErrTokenExpired) {
			s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
				IdentifierType: "validate",
			}, domainerrors.ErrTokenExpired)
			return nil, domainerrors.ErrTokenExpired
		}
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			IdentifierType: "validate",
		}, domainerrors.ErrTokenInvalid)
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if it's an access token
	if !claims.IsAccess() {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "validate",
		}, domainerrors.ErrTokenInvalid)
		return nil, domainerrors.ErrTokenInvalid
	}

	// Check if token's session is in the whitelist (for device revocation support)
	allowed, err := s.tokenWhitelist.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil || !allowed {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "validate",
		}, domainerrors.ErrTokenRevoked)
		return nil, domainerrors.ErrTokenRevoked
	}

	// Check if token's session is in the blacklist (for immediate revocation)
	blacklisted, err := s.tokenBlacklist.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil || !blacklisted {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "validate",
			Extra:          &map[string]any{"context": "token_blacklisted"},
		}, domainerrors.ErrTokenRevoked)
		return nil, domainerrors.ErrTokenRevoked
	}

	s.authObserver.OnSignal(ctx, domainSignal.SignalSuccess, domainSignal.AuthSignal{
		UID:            &claims.Uid,
		IdentifierType: "validate",
	}, nil)

	return claims, nil
}

func (s *authService) RevokeToken(ctx context.Context, token string, tokenType string) error {
	s.authObserver.OnSignal(ctx, domainSignal.SignalStart, domainSignal.AuthSignal{
		IdentifierType: "revoke",
	}, nil)

	// Validate token to extract claims
	claims, err := s.tokenGen.ValidateToken(token)
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			IdentifierType: "revoke",
		}, domainerrors.ErrTokenInvalid)
		return domainerrors.ErrTokenInvalid
	}

	// Remove from whitelist
	if err := s.tokenWhitelist.Remove(ctx, claims.Uid, claims.Sid); err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "revoke",
			Extra:          &map[string]any{"context": "Failed to remove from whitelist"},
		}, err)
		return err
	}

	// Add to blacklist
	if err := s.tokenBlacklist.Add(ctx, claims.Uid, claims.Sid); err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &claims.Uid,
			IdentifierType: "revoke",
			Extra:          &map[string]any{"context": "Failed to add to blacklist"},
		}, err)
		return err
	}

	// Publish revoke event
	s.eventPublisher.Publish(ctx, event.EventRevokeToken, event.EventRevokeTokenData{
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
	})

	s.authObserver.OnSignal(ctx, domainSignal.SignalSuccess, domainSignal.AuthSignal{
		UID:            &claims.Uid,
		IdentifierType: "revoke",
	}, nil)

	return nil
}

func (s *authService) VerifyPin(ctx context.Context, userUid string, pin string) (bool, error) {
	s.authObserver.OnSignal(ctx, domainSignal.SignalStart, domainSignal.AuthSignal{
		UID:            &userUid,
		IdentifierType: "verify_pin",
	}, nil)

	// Get user
	user, err := s.userRepo.GetByUID(ctx, userUid)
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &userUid,
			IdentifierType: "verify_pin",
		}, err)
		return false, err
	}

	// Get PIN for user
	userPin, err := s.pinRepo.GetByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
				UID:            &userUid,
				Email:          &user.Email,
				IdentifierType: "verify_pin",
			}, domainerrors.ErrPinNotSet)
			return false, domainerrors.ErrPinNotSet
		}
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &userUid,
			Email:          &user.Email,
			IdentifierType: "verify_pin",
		}, err)
		return false, err
	}

	// Check if PIN is set
	if !userPin.IsSet() {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &userUid,
			Email:          &user.Email,
			IdentifierType: "verify_pin",
		}, domainerrors.ErrPinNotSet)
		return false, domainerrors.ErrPinNotSet
	}

	// Verify PIN hash
	if !s.pinHasher.Compare(userPin.Code, pin) {
		s.authObserver.OnSignal(ctx, domainSignal.SignalReject, domainSignal.AuthSignal{
			UID:            &userUid,
			Email:          &user.Email,
			IdentifierType: "verify_pin",
		}, nil) // Invalid PIN, but not an error - just reject

		// Publish PIN verify failed event
		err = s.eventPublisher.Publish(ctx, event.EventPINFail, event.EventPinFailData{
			UserUID: userUid,
			Reason:  "invalid_pin",
		})
		if err != nil {
			s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
				UID:            &userUid,
				Email:          &user.Email,
				IdentifierType: "verify_pin",
			}, err)
			return false, err
		}

		return false, nil
	}

	// Publish PIN verify success event
	err = s.eventPublisher.Publish(ctx, event.EventPINVerify, event.EventPinVerifyData{
		UserUID: userUid,
		Success: true,
		Reason:  "pin_verified",
	})
	if err != nil {
		s.authObserver.OnSignal(ctx, domainSignal.SignalFail, domainSignal.AuthSignal{
			UID:            &userUid,
			Email:          &user.Email,
			IdentifierType: "verify_pin",
		}, err)
		return false, err
	}

	active := user.IsActive()
	s.authObserver.OnSignal(ctx, domainSignal.SignalSuccess, domainSignal.AuthSignal{
		UID:            &userUid,
		Email:          &user.Email,
		Username:       &user.Username,
		Active:         &active,
		IdentifierType: "verify_pin",
	}, nil)

	return true, nil
}

func (s *authService) findUser(ctx context.Context, identifierType string, identifier string) (*model.User, error) {
	switch identifierType {
	case "email":
		return s.userRepo.GetByEmail(ctx, identifier)
	case "username":
		return s.userRepo.GetByUsername(ctx, identifier)
	default:
		return nil, domainerrors.ErrInvalidIdentifierType
	}
}

func (s *authService) findOrCreateDevice(ctx context.Context, name string, fingerprint string) (*model.Device, error) {
	device, err := s.deviceRepo.GetByFingerprint(ctx, fingerprint)
	if err != nil {
		if errors.Is(err, domainerrors.ErrDeviceNotFound) {
			return s.deviceRepo.Create(ctx, &model.Device{
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

func (s *authService) findOrCreateUserDevice(ctx context.Context, user *model.User, device *model.Device, ipAddress, sessionID string) (*model.UserDevice, error) {
	userDevice, err := s.userDeviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, device.ID)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserDeviceNotFound) {
			now := time.Now().UTC()
			return s.userDeviceRepo.Create(ctx, &model.UserDevice{
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
	userDevice, err = s.userDeviceRepo.GetByUserIDAndDeviceID(ctx, user.ID, device.ID)
	if err != nil {
		return nil, err
	}
	return userDevice, nil
}

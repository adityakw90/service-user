package service

import (
	"context"
	"errors"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	domainEvent "github.com/adityakw90/service-user/internal/core/domain/event"
	domainModel "github.com/adityakw90/service-user/internal/core/domain/model"
	domainParam "github.com/adityakw90/service-user/internal/core/domain/param"
	domainSignal "github.com/adityakw90/service-user/internal/core/domain/signal"
	eventmocks "github.com/adityakw90/service-user/mocks/event"
	executormocks "github.com/adityakw90/service-user/mocks/executor"
	oauthmocks "github.com/adityakw90/service-user/mocks/oauth"
	observermocks "github.com/adityakw90/service-user/mocks/observer"
	repomocks "github.com/adityakw90/service-user/mocks/repository"
	securitymocks "github.com/adityakw90/service-user/mocks/security"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type authServiceMocks struct {
	userRepo       *repomocks.MockUserRepository
	deviceRepo     *repomocks.MockDeviceRepository
	userDeviceRepo *repomocks.MockUserDeviceRepository
	pinRepo        *repomocks.MockUserPinRepository
	passwordHasher *securitymocks.MockHasher
	pinHasher      *securitymocks.MockHasher
	tokenGen       *securitymocks.MockTokenGenerator
	uidGen         *securitymocks.MockUIDGenerator
	oauthProvider  *oauthmocks.MockOAuthProvider
	tokenWhitelist *securitymocks.MockTokenStore
	tokenBlacklist *securitymocks.MockTokenStore
	executor       *executormocks.MockExecutor
	eventPublisher *eventmocks.MockEventPublisher
	attemptTracker *securitymocks.MockAttemptTracker
	rateLimiter    *securitymocks.MockRateLimiter
	authObserver   *observermocks.MockServiceObserver[domainSignal.AuthSignal]
}

func newAuthServiceMocks(t *testing.T) authServiceMocks {
	return authServiceMocks{
		userRepo:       repomocks.NewMockUserRepository(t),
		deviceRepo:     repomocks.NewMockDeviceRepository(t),
		userDeviceRepo: repomocks.NewMockUserDeviceRepository(t),
		pinRepo:        repomocks.NewMockUserPinRepository(t),
		passwordHasher: securitymocks.NewMockHasher(t),
		pinHasher:      securitymocks.NewMockHasher(t),
		tokenGen:       securitymocks.NewMockTokenGenerator(t),
		uidGen:         securitymocks.NewMockUIDGenerator(t),
		oauthProvider:  oauthmocks.NewMockOAuthProvider(t),
		tokenWhitelist: securitymocks.NewMockTokenStore(t),
		tokenBlacklist: securitymocks.NewMockTokenStore(t),
		executor:       executormocks.NewMockExecutor(t),
		eventPublisher: eventmocks.NewMockEventPublisher(t),
		attemptTracker: securitymocks.NewMockAttemptTracker(t),
		rateLimiter:    securitymocks.NewMockRateLimiter(t),
		authObserver:   observermocks.NewMockServiceObserver[domainSignal.AuthSignal](t),
	}
}

// TestAuthService_GoogleOAuth tests the GoogleOAuth method.
func TestAuthService_GoogleOAuth(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*securitymocks.MockUIDGenerator, *oauthmocks.MockOAuthProvider)
		redirectURI string
		want        string
		state       string
		wantErr     bool
	}{
		{
			name: "Happy Path - returns authorization URL and state",
			setupMocks: func(uidGen *securitymocks.MockUIDGenerator, mockOAuth *oauthmocks.MockOAuthProvider) {
				uidGen.EXPECT().New().Return("state-123").Once()
				mockOAuth.EXPECT().GetAuthorizationURL(mock.Anything, "state-123", "http://localhost:8080/callback").Return("https://accounts.google.com/o/oauth2/v2/auth?state=state-123", nil).Once()
			},
			redirectURI: "http://localhost:8080/callback",
			want:        "https://accounts.google.com/o/oauth2/v2/auth?state=state-123",
			state:       "state-123",
			wantErr:     false,
		},
		{
			name: "Happy Path - custom state value",
			setupMocks: func(uidGen *securitymocks.MockUIDGenerator, mockOAuth *oauthmocks.MockOAuthProvider) {
				uidGen.EXPECT().New().Return("custom-state-456").Once()
				mockOAuth.EXPECT().GetAuthorizationURL(mock.Anything, "custom-state-456", "https://example.com/oauth/callback").Return("https://accounts.google.com/o/oauth2/v2/auth?state=custom-state-456", nil).Once()
			},
			redirectURI: "https://example.com/oauth/callback",
			want:        "https://accounts.google.com/o/oauth2/v2/auth?state=custom-state-456",
			state:       "custom-state-456",
			wantErr:     false,
		},
		{
			name: "OAuth provider returns error",
			setupMocks: func(uidGen *securitymocks.MockUIDGenerator, mockOAuth *oauthmocks.MockOAuthProvider) {
				uidGen.EXPECT().New().Return("state-error").Once()
				mockOAuth.EXPECT().GetAuthorizationURL(mock.Anything, "state-error", "http://localhost:8080/callback").Return("", context.Canceled).Once()
			},
			redirectURI: "http://localhost:8080/callback",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockOAuthProvider := oauthmocks.NewMockOAuthProvider(t)
			mockTokenGen := securitymocks.NewMockTokenGenerator(t)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)
			mockTokenBlacklist := securitymocks.NewMockTokenStore(t)
			mockExecutor := executormocks.NewMockExecutor(t)
			mockEventPublisher := eventmocks.NewMockEventPublisher(t)
			mockAuthObserver := observermocks.NewMockServiceObserver[domainSignal.AuthSignal](t)
			mockAuthObserver.EXPECT().OnSignal(mock.Anything, mock.Anything, mock.AnythingOfType("signal.AuthSignal"), mock.Anything).Maybe()

			if tt.setupMocks != nil {
				tt.setupMocks(mockUIDGen, mockOAuthProvider)
			}

			svc := NewAuthService(
				mockUserRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPinRepo,
				mockPasswordHasher,
				mockPinHasher,
				mockTokenGen,
				mockUIDGen,
				mockOAuthProvider,
				mockTokenWhitelist,
				mockTokenBlacklist,
				mockExecutor,
				mockEventPublisher,
				nil,
				nil,
			)

			got, state, err := svc.GoogleOAuth(context.Background(), tt.redirectURI)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			if tt.state != "" {
				assert.Equal(t, tt.state, state)
			}
		})
	}
}

// TestAuthService_generateUsername tests the generateUsername method.
func TestAuthService_generateUsername(t *testing.T) {
	tests := []struct {
		name       string
		userInfo   *domainModel.OAuthUserInfo
		setupMocks func(*repomocks.MockUserRepository, *securitymocks.MockUIDGenerator)
		want       string
	}{
		{
			name: "Happy Path - Simple Email",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "john.doe@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john.doe").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "john.doe",
		},
		{
			name: "Happy Path - Email with Numbers",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "user123@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user123").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user123",
		},
		{
			name: "Special Characters Normalization - Hyphens and Plus",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "john-doe+test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john_doe_test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "john_doe_test",
		},
		{
			name: "Special Characters Normalization - Exclamation Mark",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "user!test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_test",
		},
		{
			name: "Special Characters Normalization - Multiple Special Chars Sequence",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "user#$%test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_test",
		},
		{
			name: "Minimum Length Handling - Two Characters",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "ab@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_ab").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_ab",
		},
		{
			name: "Minimum Length Handling - Single Character",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "x@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_x").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_x",
		},
		{
			name: "Underscores and Dots Preserved",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "user_name.test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_name.test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_name.test",
		},
		{
			name: "Single Collision - Adds Suffix",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "john.doe@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				existingUser := &domainModel.User{Username: "john.doe"}
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john.doe").Return(existingUser, nil)
				mockUIDGen.EXPECT().New().Return("uid123456789")
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john.doe_6789").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "john.doe_6789",
		},
		{
			name: "Multiple Collisions - Keeps Adding Suffixes",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "john@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				existingUser := &domainModel.User{Username: "john"}
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john").Return(existingUser, nil).Once()
				mockUIDGen.EXPECT().New().Return("uid1111222333").Once()
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john_2333").Return(existingUser, nil).Once()
				mockUIDGen.EXPECT().New().Return("uid4444555666").Once()
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john_2333_5666").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			want: "john_2333_5666",
		},
		{
			name: "Short Username with Collision",
			userInfo: &domainModel.OAuthUserInfo{
				Email: "ab@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				existingUser := &domainModel.User{Username: "user_ab"}
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_ab").Return(existingUser, nil)
				mockUIDGen.EXPECT().New().Return("uid9876543210")
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_ab_3210").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_ab_3210",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockUIDGen)
			}

			svc := NewAuthService(
				mockUserRepo,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				mockUIDGen,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)

			got := svc.(*authService).generateUsername(context.Background(), tt.userInfo)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthService_Authenticate(t *testing.T) {
	testUser := createTestUser(1, "user-uid-123", "john_doe", "john@example.com", "hashed_password", domainModel.UserStatusActive)
	inactiveUser := createTestUser(2, "user-uid-456", "inactive_user", "inactive@example.com", "hashed_password", domainModel.UserStatusInactive)
	deletedUser := createDeletedUser(3, "user-uid-789", "deleted_user", "deleted@example.com")

	tests := []struct {
		name       string
		payload    *domainParam.AuthParams
		setupMocks func(m authServiceMocks)
		wantErr    error
	}{
		{
			name: "IP rate limit hit",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
				DeviceIP:       util.Ptr("192.168.1.1"),
			},
			setupMocks: func(m authServiceMocks) {
				m.rateLimiter.EXPECT().Acquire(mock.Anything, "192.168.1.1").Return(false, nil).Once()
			},
			wantErr: domainerrors.ErrRateLimitExceeded,
		},
		{
			name: "IP rate limit error",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
				DeviceIP:       util.Ptr("192.168.1.1"),
			},
			setupMocks: func(m authServiceMocks) {
				m.rateLimiter.EXPECT().Acquire(mock.Anything, "192.168.1.1").Return(false, errors.New("limiter error")).Once()
			},
			wantErr: errors.New("limiter error"),
		},
		{
			name: "User not found",
			payload: &domainParam.AuthParams{
				Identifier:     "notfound@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "notfound@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			wantErr: domainerrors.ErrInvalidCredentials,
		},
		{
			name: "Database error during findUser",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(nil, errors.New("db error")).Once()
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "User is deleted",
			payload: &domainParam.AuthParams{
				Identifier:     "deleted@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "deleted@example.com").Return(deletedUser, nil).Once()
			},
			wantErr: domainerrors.ErrUserDeleted,
		},
		{
			name: "User is inactive",
			payload: &domainParam.AuthParams{
				Identifier:     "inactive@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "inactive@example.com").Return(inactiveUser, nil).Once()
			},
			wantErr: domainerrors.ErrUserInactive,
		},
		{
			name: "Account locked out",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(true, nil).Once()

				// Mock executor synchronously executing event publisher call
				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.locked", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventLoginLocked &&
						msg.Entity.ID == testUser.UID &&
						msg.Metadata.(domainEvent.EventLoginLockedData).FailureReason == "Account is locked"
				})).Return(nil).Once()
			},
			wantErr: domainerrors.ErrAccountLockedOut,
		},
		{
			name: "Attempt tracker error during IsLocked",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(false, errors.New("tracker error")).Once()
			},
			wantErr: errors.New("tracker error"),
		},
		{
			name: "Invalid password credentials",
			payload: &domainParam.AuthParams{
				Identifier:     "john_doe",
				IdentifierType: "username",
				Password:       "wrong_password",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByUsername(mock.Anything, "john_doe").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(false, nil).Once()
				m.passwordHasher.EXPECT().Compare("hashed_password", "wrong_password").Return(false).Once()
				m.attemptTracker.EXPECT().Track(mock.Anything, testUser.UID).Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.failed", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventLoginFailed &&
						msg.Entity.ID == testUser.UID &&
						msg.Metadata.(domainEvent.EventLoginFailedData).FailureReason == "invalid_credentials"
				})).Return(nil).Once()
			},
			wantErr: domainerrors.ErrInvalidCredentials,
		},
		{
			name: "Token generator access token error",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(false, nil).Once()
				m.passwordHasher.EXPECT().Compare("hashed_password", "password123").Return(true).Once()
				m.attemptTracker.EXPECT().Reset(mock.Anything, testUser.UID).Return(nil).Once()
				m.uidGen.EXPECT().New().Return("session-id-123").Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == testUser.UID && claims.Sid == "session-id-123" && claims.Type == domainModel.TokenTypeAccess
				})).Return("", errors.New("access token err")).Once()
			},
			wantErr: errors.New("access token err"),
		},
		{
			name: "Token generator refresh token error",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(false, nil).Once()
				m.passwordHasher.EXPECT().Compare("hashed_password", "password123").Return(true).Once()
				m.attemptTracker.EXPECT().Reset(mock.Anything, testUser.UID).Return(nil).Once()
				m.uidGen.EXPECT().New().Return("session-id-123").Once()

				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("access-token", nil).Once()
				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == testUser.UID && claims.Sid == "session-id-123" && claims.Type == domainModel.TokenTypeRefresh
				})).Return("", errors.New("refresh token err")).Once()
			},
			wantErr: errors.New("refresh token err"),
		},
		{
			name: "Happy Path - successful authentication without device info",
			payload: &domainParam.AuthParams{
				Identifier:     "john@example.com",
				IdentifierType: "email",
				Password:       "password123",
				Extra:          &map[string]any{"custom": "value"},
			},
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(false, nil).Once()
				m.passwordHasher.EXPECT().Compare("hashed_password", "password123").Return(true).Once()
				m.attemptTracker.EXPECT().Reset(mock.Anything, testUser.UID).Return(nil).Once()
				m.uidGen.EXPECT().New().Return("session-id-123").Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == testUser.UID && claims.Sid == "session-id-123" && claims.Type == domainModel.TokenTypeAccess && claims.Extra["custom"] == "value"
				})).Return("access-token-123", nil).Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == testUser.UID && claims.Sid == "session-id-123" && claims.Type == domainModel.TokenTypeRefresh
				})).Return("refresh-token-123", nil).Once()

				m.tokenWhitelist.EXPECT().Add(mock.Anything, testUser.UID, "session-id-123").Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventLogin &&
						msg.Entity.ID == testUser.UID &&
						msg.Metadata.(*domainEvent.EventLoginData).Identifier == "john@example.com"
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Happy Path - successful authentication with new device registration",
			payload: &domainParam.AuthParams{
				Identifier:        "john@example.com",
				IdentifierType:    "email",
				Password:          "password123",
				DeviceFingerprint: util.Ptr("fingerprint-123"),
				DeviceName:        util.Ptr("Pixel 6"),
				DeviceIP:          util.Ptr("192.168.1.100"),
			},
			setupMocks: func(m authServiceMocks) {
				m.rateLimiter.EXPECT().Acquire(mock.Anything, "192.168.1.100").Return(true, nil).Once()
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(false, nil).Once()
				m.passwordHasher.EXPECT().Compare("hashed_password", "password123").Return(true).Once()
				m.attemptTracker.EXPECT().Reset(mock.Anything, testUser.UID).Return(nil).Once()

				// First session ID generation for login sid
				m.uidGen.EXPECT().New().Return("session-id-123").Once()

				// Device creation mocks
				m.deviceRepo.EXPECT().GetByFingerprint(mock.Anything, "fingerprint-123").Return(nil, domainerrors.ErrDeviceNotFound).Once()
				m.uidGen.EXPECT().New().Return("device-uid-456").Once() // uid for new device
				newDevice := &domainModel.Device{ID: 10, UID: "device-uid-456", DeviceName: "Pixel 6", DeviceFingerprint: "fingerprint-123"}
				m.deviceRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(d *domainModel.Device) bool {
					return d.DeviceName == "Pixel 6" && d.DeviceFingerprint == "fingerprint-123"
				})).Return(newDevice, nil).Once()

				m.userDeviceRepo.EXPECT().GetByUserIDAndDeviceID(mock.Anything, testUser.ID, newDevice.ID).Return(nil, domainerrors.ErrUserDeviceNotFound).Once()
				userDevice := &domainModel.UserDevice{UserID: testUser.ID, DeviceID: newDevice.ID, IPAddress: "192.168.1.100", SessionID: "session-id-123"}
				m.userDeviceRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ud *domainModel.UserDevice) bool {
					return ud.UserID == testUser.ID && ud.DeviceID == newDevice.ID && ud.SessionID == "session-id-123"
				})).Return(userDevice, nil).Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					deviceClaim := claims.Extra["device"].(map[string]any)
					return claims.Uid == testUser.UID &&
						claims.Sid == "session-id-123" &&
						deviceClaim["uid"] == "device-uid-456" &&
						deviceClaim["name"] == "Pixel 6" &&
						deviceClaim["ip_address"] == "192.168.1.100"
				})).Return("access-token-123", nil).Once()

				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("refresh-token-123", nil).Once()

				m.tokenWhitelist.EXPECT().Add(mock.Anything, testUser.UID, "session-id-123").Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					ld := msg.Metadata.(*domainEvent.EventLoginData)
					return msg.Type == domainEvent.EventLogin &&
						ld.Identifier == "john@example.com" &&
						*ld.DeviceUID == "device-uid-456" &&
						*ld.IPAddress == "192.168.1.100"
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Happy Path - successful authentication with existing device update",
			payload: &domainParam.AuthParams{
				Identifier:        "john@example.com",
				IdentifierType:    "email",
				Password:          "password123",
				DeviceFingerprint: util.Ptr("fingerprint-123"),
				DeviceName:        util.Ptr("Pixel 6"),
				DeviceIP:          util.Ptr("192.168.1.100"),
			},
			setupMocks: func(m authServiceMocks) {
				m.rateLimiter.EXPECT().Acquire(mock.Anything, "192.168.1.100").Return(true, nil).Once()
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()
				m.attemptTracker.EXPECT().IsLocked(mock.Anything, testUser.UID).Return(false, nil).Once()
				m.passwordHasher.EXPECT().Compare("hashed_password", "password123").Return(true).Once()
				m.attemptTracker.EXPECT().Reset(mock.Anything, testUser.UID).Return(nil).Once()
				m.uidGen.EXPECT().New().Return("session-id-123").Once()

				existingDevice := &domainModel.Device{ID: 10, UID: "device-uid-456", DeviceName: "Pixel 6", DeviceFingerprint: "fingerprint-123"}
				m.deviceRepo.EXPECT().GetByFingerprint(mock.Anything, "fingerprint-123").Return(existingDevice, nil).Once()

				m.userDeviceRepo.EXPECT().GetByUserIDAndDeviceID(mock.Anything, testUser.ID, existingDevice.ID).Return(&domainModel.UserDevice{UserID: testUser.ID, DeviceID: existingDevice.ID}, nil).Once()
				m.userDeviceRepo.EXPECT().UpdateSessionID(mock.Anything, testUser.ID, existingDevice.ID, "session-id-123").Return(nil).Once()

				updatedUserDevice := &domainModel.UserDevice{UserID: testUser.ID, DeviceID: existingDevice.ID, IPAddress: "192.168.1.100", SessionID: "session-id-123"}
				m.userDeviceRepo.EXPECT().GetByUserIDAndDeviceID(mock.Anything, testUser.ID, existingDevice.ID).Return(updatedUserDevice, nil).Once()

				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("access-token-123", nil).Once()
				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("refresh-token-123", nil).Once()

				m.tokenWhitelist.EXPECT().Add(mock.Anything, testUser.UID, "session-id-123").Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Unknown identifier type",
			payload: &domainParam.AuthParams{
				Identifier:     "john",
				IdentifierType: "phone",
				Password:       "password123",
			},
			setupMocks: func(m authServiceMocks) {},
			wantErr:    domainerrors.ErrInvalidIdentifierType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAuthServiceMocks(t)
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			svc := NewAuthService(
				m.userRepo,
				m.deviceRepo,
				m.userDeviceRepo,
				m.pinRepo,
				m.passwordHasher,
				m.pinHasher,
				m.tokenGen,
				m.uidGen,
				m.oauthProvider,
				m.tokenWhitelist,
				m.tokenBlacklist,
				m.executor,
				m.eventPublisher,
				m.attemptTracker,
				m.rateLimiter,
			)

			got, err := svc.Authenticate(context.Background(), tt.payload)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "access-token-123", got.Access)
			assert.Equal(t, "refresh-token-123", got.Refresh)
		})
	}
}

func TestAuthService_HandleGoogleOAuth(t *testing.T) {
	testUser := createTestUser(1, "user-uid-123", "john_doe", "john@example.com", "hashed_password", domainModel.UserStatusActive)
	oauthTokens := &domainModel.OAuthTokens{AccessToken: "oauth-access"}
	oauthUserInfo := &domainModel.OAuthUserInfo{Email: "john@example.com"}

	tests := []struct {
		name       string
		setupMocks func(m authServiceMocks)
		wantErr    error
	}{
		{
			name: "oauth provider exchange code error",
			setupMocks: func(m authServiceMocks) {
				m.oauthProvider.EXPECT().ExchangeCode(mock.Anything, "code", "state", "redirect").Return(nil, errors.New("exchange error")).Once()
			},
			wantErr: errors.New("exchange error"),
		},
		{
			name: "oauth provider user info error",
			setupMocks: func(m authServiceMocks) {
				m.oauthProvider.EXPECT().ExchangeCode(mock.Anything, "code", "state", "redirect").Return(oauthTokens, nil).Once()
				m.oauthProvider.EXPECT().GetUserInfo(mock.Anything, oauthTokens).Return(nil, errors.New("user info error")).Once()
			},
			wantErr: errors.New("user info error"),
		},
		{
			name: "Happy Path - logs in existing user",
			setupMocks: func(m authServiceMocks) {
				m.oauthProvider.EXPECT().ExchangeCode(mock.Anything, "code", "state", "redirect").Return(oauthTokens, nil).Once()
				m.oauthProvider.EXPECT().GetUserInfo(mock.Anything, oauthTokens).Return(oauthUserInfo, nil).Once()
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(testUser, nil).Once()

				m.uidGen.EXPECT().New().Return("session-id-123").Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == testUser.UID && claims.Sid == "session-id-123" && claims.Type == domainModel.TokenTypeAccess
				})).Return("access-token-123", nil).Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == testUser.UID && claims.Sid == "session-id-123" && claims.Type == domainModel.TokenTypeRefresh
				})).Return("refresh-token-123", nil).Once()

				m.tokenWhitelist.EXPECT().Add(mock.Anything, testUser.UID, "session-id-123").Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.oauth", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventOAuthLogin &&
						msg.Entity.ID == testUser.UID &&
						msg.Metadata.(*domainEvent.EventOAuthLoginData).Provider == "google"
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Happy Path - registers new user with Google OAuth info",
			setupMocks: func(m authServiceMocks) {
				m.oauthProvider.EXPECT().ExchangeCode(mock.Anything, "code", "state", "redirect").Return(oauthTokens, nil).Once()
				m.oauthProvider.EXPECT().GetUserInfo(mock.Anything, oauthTokens).Return(oauthUserInfo, nil).Once()

				// User not found -> start signup flow
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
				m.passwordHasher.EXPECT().Hash(mock.Anything).Return("hashed_random_password", nil).Once()

				m.uidGen.EXPECT().New().Return("new-user-uid").Once()                                                     // uid for user model
				m.userRepo.EXPECT().GetByUsername(mock.Anything, "john").Return(nil, domainerrors.ErrUserNotFound).Once() // generateUsername collision check

				newUser := &domainModel.User{UID: "new-user-uid", Username: "john", Email: "john@example.com", Password: "hashed_random_password", Status: domainModel.UserStatusActive}
				m.userRepo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *domainModel.User) bool {
					return u.UID == "new-user-uid" && u.Username == "john" && u.Email == "john@example.com"
				})).Return(newUser, nil).Once()

				m.uidGen.EXPECT().New().Return("session-id-456").Once() // session id for login

				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("access-token-123", nil).Once()
				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("refresh-token-123", nil).Once()

				m.tokenWhitelist.EXPECT().Add(mock.Anything, "new-user-uid", "session-id-456").Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.oauth", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Create user fails on new OAuth user signup",
			setupMocks: func(m authServiceMocks) {
				m.oauthProvider.EXPECT().ExchangeCode(mock.Anything, "code", "state", "redirect").Return(oauthTokens, nil).Once()
				m.oauthProvider.EXPECT().GetUserInfo(mock.Anything, oauthTokens).Return(oauthUserInfo, nil).Once()

				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
				m.passwordHasher.EXPECT().Hash(mock.Anything).Return("hashed_random_password", nil).Once()
				m.uidGen.EXPECT().New().Return("new-user-uid").Once()
				m.userRepo.EXPECT().GetByUsername(mock.Anything, "john").Return(nil, domainerrors.ErrUserNotFound).Once()

				m.userRepo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("db insert error")).Once()
			},
			wantErr: errors.New("db insert error"),
		},
		{
			name: "GetByEmail DB error",
			setupMocks: func(m authServiceMocks) {
				m.oauthProvider.EXPECT().ExchangeCode(mock.Anything, "code", "state", "redirect").Return(oauthTokens, nil).Once()
				m.oauthProvider.EXPECT().GetUserInfo(mock.Anything, oauthTokens).Return(oauthUserInfo, nil).Once()
				m.userRepo.EXPECT().GetByEmail(mock.Anything, "john@example.com").Return(nil, errors.New("db fetch error")).Once()
			},
			wantErr: errors.New("db fetch error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAuthServiceMocks(t)
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			svc := NewAuthService(
				m.userRepo,
				m.deviceRepo,
				m.userDeviceRepo,
				m.pinRepo,
				m.passwordHasher,
				m.pinHasher,
				m.tokenGen,
				m.uidGen,
				m.oauthProvider,
				m.tokenWhitelist,
				m.tokenBlacklist,
				m.executor,
				m.eventPublisher,
				m.attemptTracker,
				m.rateLimiter,
			)

			got, err := svc.HandleGoogleOAuth(context.Background(), "code", "state", "redirect")
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "access-token-123", got.Access)
			assert.Equal(t, "refresh-token-123", got.Refresh)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	testUser := createTestUser(1, "user-uid-123", "john_doe", "john@example.com", "hashed_password", domainModel.UserStatusActive)
	inactiveUser := createTestUser(2, "user-uid-456", "inactive_user", "inactive@example.com", "hashed_password", domainModel.UserStatusInactive)
	deletedUser := createDeletedUser(3, "user-uid-789", "deleted_user", "deleted@example.com")

	refreshClaims := &domainModel.TokenClaims{
		Uid:            "user-uid-123",
		Sid:            "old-session-123",
		Type:           domainModel.TokenTypeRefresh,
		Identifier:     "john@example.com",
		IdentifierType: "email",
	}

	tests := []struct {
		name         string
		refreshToken string
		setupMocks   func(m authServiceMocks)
		wantErr      error
	}{
		{
			name:         "token validation error",
			refreshToken: "invalid-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("invalid-token").Return(nil, errors.New("invalid signature")).Once()
			},
			wantErr: errors.New("invalid signature"),
		},
		{
			name:         "not a refresh token type",
			refreshToken: "access-token",
			setupMocks: func(m authServiceMocks) {
				accessClaims := &domainModel.TokenClaims{Type: domainModel.TokenTypeAccess}
				m.tokenGen.EXPECT().ValidateToken("access-token").Return(accessClaims, nil).Once()
			},
			wantErr: domainerrors.ErrTokenInvalid,
		},
		{
			name:         "token not whitelisted/revoked",
			refreshToken: "revoked-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("revoked-token").Return(refreshClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, refreshClaims.Uid, refreshClaims.Sid).Return(false, nil).Once()
			},
			wantErr: domainerrors.ErrTokenRevoked,
		},
		{
			name:         "database error during whitelist check",
			refreshToken: "refresh-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("refresh-token").Return(refreshClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, refreshClaims.Uid, refreshClaims.Sid).Return(false, errors.New("whitelist error")).Once()
			},
			wantErr: errors.New("whitelist error"),
		},
		{
			name:         "user not found",
			refreshToken: "refresh-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("refresh-token").Return(refreshClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, refreshClaims.Uid, refreshClaims.Sid).Return(true, nil).Once()
				m.userRepo.EXPECT().GetByUID(mock.Anything, refreshClaims.Uid).Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			wantErr: domainerrors.ErrTokenInvalid,
		},
		{
			name:         "user deleted",
			refreshToken: "refresh-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("refresh-token").Return(refreshClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, refreshClaims.Uid, refreshClaims.Sid).Return(true, nil).Once()
				m.userRepo.EXPECT().GetByUID(mock.Anything, refreshClaims.Uid).Return(deletedUser, nil).Once()
			},
			wantErr: domainerrors.ErrUserDeleted,
		},
		{
			name:         "user inactive",
			refreshToken: "refresh-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("refresh-token").Return(refreshClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, refreshClaims.Uid, refreshClaims.Sid).Return(true, nil).Once()
				m.userRepo.EXPECT().GetByUID(mock.Anything, refreshClaims.Uid).Return(inactiveUser, nil).Once()
			},
			wantErr: domainerrors.ErrUserInactive,
		},
		{
			name:         "Happy Path - refreshes tokens successfully without device",
			refreshToken: "refresh-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("refresh-token").Return(refreshClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, refreshClaims.Uid, refreshClaims.Sid).Return(true, nil).Once()
				m.userRepo.EXPECT().GetByUID(mock.Anything, refreshClaims.Uid).Return(testUser, nil).Once()

				m.uidGen.EXPECT().New().Return("new-session-id-789").Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == refreshClaims.Uid && claims.Sid == "new-session-id-789" && claims.Type == domainModel.TokenTypeAccess
				})).Return("new-access-token-123", nil).Once()

				m.tokenGen.EXPECT().GenerateToken(mock.MatchedBy(func(claims *domainModel.TokenClaims) bool {
					return claims.Uid == refreshClaims.Uid && claims.Sid == "new-session-id-789" && claims.Type == domainModel.TokenTypeRefresh
				})).Return("new-refresh-token-123", nil).Once()

				m.tokenWhitelist.EXPECT().Add(mock.Anything, refreshClaims.Uid, "new-session-id-789").Return(nil).Once()
				m.tokenWhitelist.EXPECT().Remove(mock.Anything, refreshClaims.Uid, refreshClaims.Sid).Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.refresh_token", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventTokenRefresh &&
						msg.Entity.ID == refreshClaims.Sid &&
						msg.Metadata.(domainEvent.EventTokenRefreshData).Identifier == "john@example.com"
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:         "Happy Path - refreshes tokens successfully with device update",
			refreshToken: "refresh-token",
			setupMocks: func(m authServiceMocks) {
				claimsWithDevice := &domainModel.TokenClaims{
					Uid:            "user-uid-123",
					Sid:            "old-session-123",
					Type:           domainModel.TokenTypeRefresh,
					Identifier:     "john@example.com",
					IdentifierType: "email",
					Extra:          map[string]any{"device": map[string]any{"uid": "device-uid-xyz"}},
				}
				m.tokenGen.EXPECT().ValidateToken("refresh-token").Return(claimsWithDevice, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, claimsWithDevice.Uid, claimsWithDevice.Sid).Return(true, nil).Once()
				m.userRepo.EXPECT().GetByUID(mock.Anything, claimsWithDevice.Uid).Return(testUser, nil).Once()

				m.uidGen.EXPECT().New().Return("new-session-id-789").Once()

				// Device lookup and session ID update mocks
				existingDevice := &domainModel.Device{ID: 15, UID: "device-uid-xyz"}
				m.deviceRepo.EXPECT().GetByUID(mock.Anything, "device-uid-xyz").Return(existingDevice, nil).Once()
				m.userDeviceRepo.EXPECT().UpdateSessionID(mock.Anything, testUser.ID, existingDevice.ID, "new-session-id-789").Return(nil).Once()

				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("new-access-token-123", nil).Once()
				m.tokenGen.EXPECT().GenerateToken(mock.Anything).Return("new-refresh-token-123", nil).Once()

				m.tokenWhitelist.EXPECT().Add(mock.Anything, claimsWithDevice.Uid, "new-session-id-789").Return(nil).Once()
				m.tokenWhitelist.EXPECT().Remove(mock.Anything, claimsWithDevice.Uid, claimsWithDevice.Sid).Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.refresh_token", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAuthServiceMocks(t)
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			svc := NewAuthService(
				m.userRepo,
				m.deviceRepo,
				m.userDeviceRepo,
				m.pinRepo,
				m.passwordHasher,
				m.pinHasher,
				m.tokenGen,
				m.uidGen,
				m.oauthProvider,
				m.tokenWhitelist,
				m.tokenBlacklist,
				m.executor,
				m.eventPublisher,
				m.attemptTracker,
				m.rateLimiter,
			)

			got, err := svc.RefreshToken(context.Background(), tt.refreshToken)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "new-access-token-123", got.Access)
			assert.Equal(t, "new-refresh-token-123", got.Refresh)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	validClaims := &domainModel.TokenClaims{
		Uid:  "user-uid-123",
		Sid:  "session-id-abc",
		Type: domainModel.TokenTypeAccess,
	}

	tests := []struct {
		name        string
		accessToken string
		setupMocks  func(m authServiceMocks)
		wantErr     error
	}{
		{
			name:        "validation error",
			accessToken: "invalid-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("invalid-token").Return(nil, errors.New("expired")).Once()
			},
			wantErr: errors.New("expired"),
		},
		{
			name:        "invalid token type (refresh instead of access)",
			accessToken: "refresh-token",
			setupMocks: func(m authServiceMocks) {
				refreshClaims := &domainModel.TokenClaims{Type: domainModel.TokenTypeRefresh}
				m.tokenGen.EXPECT().ValidateToken("refresh-token").Return(refreshClaims, nil).Once()
			},
			wantErr: domainerrors.ErrTokenInvalid,
		},
		{
			name:        "token not in whitelist (session revoked)",
			accessToken: "access-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("access-token").Return(validClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, validClaims.Uid, validClaims.Sid).Return(false, nil).Once()
			},
			wantErr: domainerrors.ErrTokenRevoked,
		},
		{
			name:        "token in blacklist (immediate revocation)",
			accessToken: "access-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("access-token").Return(validClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, validClaims.Uid, validClaims.Sid).Return(true, nil).Once()
				// blacklisted is true if IsAllowed returns FALSE (due to blacklist storage semantics)
				m.tokenBlacklist.EXPECT().IsAllowed(mock.Anything, validClaims.Uid, validClaims.Sid).Return(false, nil).Once()
			},
			wantErr: domainerrors.ErrTokenRevoked,
		},
		{
			name:        "Happy Path - valid token",
			accessToken: "access-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("access-token").Return(validClaims, nil).Once()
				m.tokenWhitelist.EXPECT().IsAllowed(mock.Anything, validClaims.Uid, validClaims.Sid).Return(true, nil).Once()
				m.tokenBlacklist.EXPECT().IsAllowed(mock.Anything, validClaims.Uid, validClaims.Sid).Return(true, nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAuthServiceMocks(t)
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			svc := NewAuthService(
				m.userRepo,
				m.deviceRepo,
				m.userDeviceRepo,
				m.pinRepo,
				m.passwordHasher,
				m.pinHasher,
				m.tokenGen,
				m.uidGen,
				m.oauthProvider,
				m.tokenWhitelist,
				m.tokenBlacklist,
				m.executor,
				m.eventPublisher,
				m.attemptTracker,
				m.rateLimiter,
			)

			got, err := svc.ValidateToken(context.Background(), tt.accessToken)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, validClaims, got)
		})
	}
}

func TestAuthService_RevokeToken(t *testing.T) {
	validClaims := &domainModel.TokenClaims{
		Uid:            "user-uid-123",
		Sid:            "session-id-abc",
		Type:           domainModel.TokenTypeAccess,
		Identifier:     "john@example.com",
		IdentifierType: "email",
	}

	tests := []struct {
		name       string
		token      string
		setupMocks func(m authServiceMocks)
		wantErr    error
	}{
		{
			name:  "validation error",
			token: "invalid-token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("invalid-token").Return(nil, errors.New("invalid signature")).Once()
			},
			wantErr: errors.New("invalid signature"),
		},
		{
			name:  "whitelist removal error",
			token: "token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("token").Return(validClaims, nil).Once()
				m.tokenWhitelist.EXPECT().Remove(mock.Anything, validClaims.Uid, validClaims.Sid).Return(errors.New("redis remove error")).Once()
			},
			wantErr: errors.New("redis remove error"),
		},
		{
			name:  "blacklist addition error",
			token: "token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("token").Return(validClaims, nil).Once()
				m.tokenWhitelist.EXPECT().Remove(mock.Anything, validClaims.Uid, validClaims.Sid).Return(nil).Once()
				m.tokenBlacklist.EXPECT().Add(mock.Anything, validClaims.Uid, validClaims.Sid).Return(errors.New("redis add error")).Once()
			},
			wantErr: errors.New("redis add error"),
		},
		{
			name:  "Happy Path - successfully revokes token",
			token: "token",
			setupMocks: func(m authServiceMocks) {
				m.tokenGen.EXPECT().ValidateToken("token").Return(validClaims, nil).Once()
				m.tokenWhitelist.EXPECT().Remove(mock.Anything, validClaims.Uid, validClaims.Sid).Return(nil).Once()
				m.tokenBlacklist.EXPECT().Add(mock.Anything, validClaims.Uid, validClaims.Sid).Return(nil).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.revoke_token", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventRevokeToken &&
						msg.Entity.ID == validClaims.Sid &&
						msg.Metadata.(domainEvent.EventRevokeTokenData).Identifier == "john@example.com"
				})).Return(nil).Once()
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAuthServiceMocks(t)
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			svc := NewAuthService(
				m.userRepo,
				m.deviceRepo,
				m.userDeviceRepo,
				m.pinRepo,
				m.passwordHasher,
				m.pinHasher,
				m.tokenGen,
				m.uidGen,
				m.oauthProvider,
				m.tokenWhitelist,
				m.tokenBlacklist,
				m.executor,
				m.eventPublisher,
				m.attemptTracker,
				m.rateLimiter,
			)

			err := svc.RevokeToken(context.Background(), tt.token, "access")
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAuthService_VerifyPin(t *testing.T) {
	testUser := createTestUser(1, "user-uid-123", "john_doe", "john@example.com", "hashed_password", domainModel.UserStatusActive)
	userPin := createUserPin(testUser.ID, testUser.UID, "hashed_pin_123")
	userPinNotSet := createUserPin(testUser.ID, testUser.UID, "")

	tests := []struct {
		name       string
		userUid    string
		pinCode    string
		setupMocks func(m authServiceMocks)
		want       bool
		wantErr    error
	}{
		{
			name:    "user not found",
			userUid: "unknown-user",
			pinCode: "123456",
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByUID(mock.Anything, "unknown-user").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			want:    false,
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name:    "pin not found / set for user",
			userUid: testUser.UID,
			pinCode: "123456",
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByUID(mock.Anything, testUser.UID).Return(testUser, nil).Once()
				m.pinRepo.EXPECT().GetByUserID(mock.Anything, testUser.ID).Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			want:    false,
			wantErr: domainerrors.ErrPinNotSet,
		},
		{
			name:    "pin set empty check fails",
			userUid: testUser.UID,
			pinCode: "123456",
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByUID(mock.Anything, testUser.UID).Return(testUser, nil).Once()
				m.pinRepo.EXPECT().GetByUserID(mock.Anything, testUser.ID).Return(userPinNotSet, nil).Once()
			},
			want:    false,
			wantErr: domainerrors.ErrPinNotSet,
		},
		{
			name:    "invalid pin comparison",
			userUid: testUser.UID,
			pinCode: "654321",
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByUID(mock.Anything, testUser.UID).Return(testUser, nil).Once()
				m.pinRepo.EXPECT().GetByUserID(mock.Anything, testUser.ID).Return(userPin, nil).Once()
				m.pinHasher.EXPECT().Compare("hashed_pin_123", "654321").Return(false).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.pin_fail", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventPINFail &&
						msg.Entity.ID == userPin.UserUID &&
						msg.Metadata.(domainEvent.EventPinFailData).Reason == "invalid_pin"
				})).Return(nil).Once()
			},
			want:    false,
			wantErr: nil,
		},
		{
			name:    "Happy Path - correct PIN code verified",
			userUid: testUser.UID,
			pinCode: "123456",
			setupMocks: func(m authServiceMocks) {
				m.userRepo.EXPECT().GetByUID(mock.Anything, testUser.UID).Return(testUser, nil).Once()
				m.pinRepo.EXPECT().GetByUserID(mock.Anything, testUser.ID).Return(userPin, nil).Once()
				m.pinHasher.EXPECT().Compare("hashed_pin_123", "123456").Return(true).Once()

				m.executor.EXPECT().DoAsync(mock.Anything, "auth.publish.pin_verified", mock.Anything).Run(func(ctx context.Context, name string, fn func(context.Context) error) {
					_ = fn(ctx)
				}).Return(nil).Once()

				m.eventPublisher.EXPECT().Publish(mock.Anything, mock.MatchedBy(func(msg domainEvent.Message) bool {
					return msg.Type == domainEvent.EventPINVerify &&
						msg.Entity.ID == userPin.UserUID &&
						msg.Metadata.(domainEvent.EventPinVerifyData).Success == true
				})).Return(nil).Once()
			},
			want:    true,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAuthServiceMocks(t)
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			svc := NewAuthService(
				m.userRepo,
				m.deviceRepo,
				m.userDeviceRepo,
				m.pinRepo,
				m.passwordHasher,
				m.pinHasher,
				m.tokenGen,
				m.uidGen,
				m.oauthProvider,
				m.tokenWhitelist,
				m.tokenBlacklist,
				m.executor,
				m.eventPublisher,
				m.attemptTracker,
				m.rateLimiter,
			)

			got, err := svc.VerifyPin(context.Background(), tt.userUid, tt.pinCode)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				assert.False(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

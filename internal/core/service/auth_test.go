package service

import (
	"context"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	eventmocks "github.com/adityakw90/service-user/mocks/event"
	executormocks "github.com/adityakw90/service-user/mocks/executor"
	oauthmocks "github.com/adityakw90/service-user/mocks/oauth"
	observermocks "github.com/adityakw90/service-user/mocks/observer"
	repomocks "github.com/adityakw90/service-user/mocks/repository"
	securitymocks "github.com/adityakw90/service-user/mocks/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupObserverAny allows any OnSignal calls on the observer (useful when not testing signal behavior)
func setupAuthObserverAny(t *testing.T, observer *observermocks.MockServiceObserver[signal.AuthSignal]) {
	// Allow any OnSignal call without checking parameters
	// Use Maybe() to make the expectation optional (can be called 0 or more times)
	// Note: Using EXPECT().OnSignal() pattern for better type safety
	observer.EXPECT().OnSignal(mock.Anything, mock.Anything, mock.AnythingOfType("signal.AuthSignal"), mock.Anything).Maybe()
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
			// Setup mocks using generated mocks
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
			mockAuthObserver := observermocks.NewMockServiceObserver[signal.AuthSignal](t)
			setupAuthObserverAny(t, mockAuthObserver) // Keep custom observer mock

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
				mockAuthObserver,
				nil, // attemptTracker
				nil, // rateLimiter
			)

			// Execute
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

func TestAuthService_generateUsername(t *testing.T) {
	tests := []struct {
		name       string
		userInfo   *model.OAuthUserInfo
		setupMocks func(*repomocks.MockUserRepository, *securitymocks.MockUIDGenerator)
		want       string
	}{
		{
			name: "Happy Path - Simple Email",
			userInfo: &model.OAuthUserInfo{
				Email: "john.doe@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john.doe").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "john.doe",
		},
		{
			name: "Happy Path - Email with Numbers",
			userInfo: &model.OAuthUserInfo{
				Email: "user123@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user123").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user123",
		},
		{
			name: "Special Characters Normalization - Hyphens and Plus",
			userInfo: &model.OAuthUserInfo{
				Email: "john-doe+test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john_doe_test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "john_doe_test",
		},
		{
			name: "Special Characters Normalization - Exclamation Mark",
			userInfo: &model.OAuthUserInfo{
				Email: "user!test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_test",
		},
		{
			name: "Special Characters Normalization - Multiple Special Chars Sequence",
			userInfo: &model.OAuthUserInfo{
				Email: "user#$%test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_test",
		},
		{
			name: "Minimum Length Handling - Two Characters",
			userInfo: &model.OAuthUserInfo{
				Email: "ab@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_ab").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_ab",
		},
		{
			name: "Minimum Length Handling - Single Character",
			userInfo: &model.OAuthUserInfo{
				Email: "x@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_x").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_x",
		},
		{
			name: "Underscores and Dots Preserved",
			userInfo: &model.OAuthUserInfo{
				Email: "user_name.test@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_name.test").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_name.test",
		},
		{
			name: "Single Collision - Adds Suffix",
			userInfo: &model.OAuthUserInfo{
				Email: "john.doe@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				// First call returns existing user
				existingUser := &model.User{Username: "john.doe"}
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john.doe").Return(existingUser, nil)
				// Generate suffix
				mockUIDGen.EXPECT().New().Return("uid123456789")
				// Second call with suffix finds no user
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john.doe_6789").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "john.doe_6789",
		},
		{
			name: "Multiple Collisions - Keeps Adding Suffixes",
			userInfo: &model.OAuthUserInfo{
				Email: "john@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				existingUser := &model.User{Username: "john"}
				// First collision
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john").Return(existingUser, nil).Once()
				mockUIDGen.EXPECT().New().Return("uid1111222333").Once()
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john_2333").Return(existingUser, nil).Once()
				// Second collision - username is now "john_2333", suffix appends to it
				mockUIDGen.EXPECT().New().Return("uid4444555666").Once()
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "john_2333_5666").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			want: "john_2333_5666",
		},
		{
			name: "Short Username with Collision",
			userInfo: &model.OAuthUserInfo{
				Email: "ab@example.com",
			},
			setupMocks: func(mockUserRepo *repomocks.MockUserRepository, mockUIDGen *securitymocks.MockUIDGenerator) {
				existingUser := &model.User{Username: "user_ab"}
				// First collision after adding prefix
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_ab").Return(existingUser, nil)
				mockUIDGen.EXPECT().New().Return("uid9876543210")
				mockUserRepo.EXPECT().GetByUsername(mock.Anything, "user_ab_3210").Return(nil, domainerrors.ErrUserNotFound)
			},
			want: "user_ab_3210",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)

			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockUIDGen)
			}

			svc := NewAuthService(
				mockUserRepo,
				nil, // deviceRepo
				nil, // userDeviceRepo
				nil, // pinRepo
				nil, // passwordHasher
				nil, // pinHasher
				nil, // tokenGen
				mockUIDGen,
				nil, // oauthProvider
				nil, // tokenWhitelist
				nil, // tokenBlacklist
				nil, // executor
				nil, // eventPublisher
				nil, // authObserver
				nil, // attemptTracker
				nil, // rateLimiter
			)

			got := svc.(*authService).generateUsername(context.Background(), tt.userInfo)
			assert.Equal(t, tt.want, got)
		})
	}
}

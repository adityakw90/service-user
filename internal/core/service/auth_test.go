package service

import (
	"context"
	"testing"

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

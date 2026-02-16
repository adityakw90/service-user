package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthService_GoogleOAuth tests the GoogleOAuth method.
func TestAuthService_GoogleOAuth(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockUIDGenerator, *MockOAuthProvider)
		redirectURI string
		want        string
		state       string
		wantErr     bool
	}{
		{
			name: "Happy Path - returns authorization URL and state",
			setupMocks: func(uidGen *MockUIDGenerator, mockOAuth *MockOAuthProvider) {
				uidGen.NewFunc = func() string {
					return "state-123"
				}
				mockOAuth.GetAuthorizationURLFunc = func(ctx context.Context, redirectURI, state string) (string, error) {
					return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state, nil
				}
			},
			redirectURI: "http://localhost:8080/callback",
			want:        "https://accounts.google.com/o/oauth2/v2/auth?state=state-123",
			state:       "state-123",
			wantErr:     false,
		},
		{
			name: "Happy Path - custom state value",
			setupMocks: func(uidGen *MockUIDGenerator, mockOAuth *MockOAuthProvider) {
				uidGen.NewFunc = func() string {
					return "custom-state-456"
				}
				mockOAuth.GetAuthorizationURLFunc = func(ctx context.Context, redirectURI, state string) (string, error) {
					return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state, nil
				}
			},
			redirectURI: "https://example.com/oauth/callback",
			want:        "https://accounts.google.com/o/oauth2/v2/auth?state=custom-state-456",
			state:       "custom-state-456",
			wantErr:     false,
		},
		{
			name: "OAuth provider returns error",
			setupMocks: func(uidGen *MockUIDGenerator, mockOAuth *MockOAuthProvider) {
				uidGen.NewFunc = func() string {
					return "state-error"
				}
				mockOAuth.GetAuthorizationURLFunc = func(ctx context.Context, redirectURI, state string) (string, error) {
					return "", context.Canceled
				}
			},
			redirectURI: "http://localhost:8080/callback",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()
			mockOAuthProvider := NewMockOAuthProvider()

			if tt.setupMocks != nil {
				tt.setupMocks(mockUIDGen, mockOAuthProvider)
			}

			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			svc := NewAuthService(
				mockUserRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPinRepo,
				mockHasher,
				mockHasher,
				mockTokenGen,
				mockUIDGen,
				mockOAuthProvider,
				mockTokenWhitelist,
				mockTokenBlacklist,
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

package service

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/errors"
	domainSignal "github.com/adityakw90/service-user/internal/core/domain/signal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_GoogleOAuth(t *testing.T) {
	tests := []struct {
		name   string
		setupMocks  func(*MockUIDGenerator, *MockOAuthProvider)
		redirectURI string
		want        string
		state        string
		wantErr     error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ug *MockUIDGenerator, oauth *MockOAuthProvider) {
				ug.NewFunc = func() string { return "state-123" }
				oauth.GetAuthorizationURLFunc = func(ctx context.Context, redirectURI, state string) (string, error) {
					return "https://accounts.google.com/o/oauth2/v2/auth?state=" + state, nil
				}
			},
			redirectURI: "http://localhost:8080/callback",
			want:       "https://accounts.google.com/o/oauth2/v2/auth?state=state-123",
			state:       "state-123",
			wantErr:     nil,
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
			mockUIDGen := tt.setupMocks(ug, nil)

			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockServiceObserver(ug, mockEventPublisher)

			svc := NewAuthService(
				mockUserRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPinRepo,
				mockHasher,
				mockHasher,
				mockTokenGen,
				mockTokenWhitelist,
				mockTokenBlacklist,
				mockUIDGen,
				nil, // rateLimiter
				mockOAuthProvider,
				mockTokenWhitelist,
				mockTokenBlacklist,
				mockEventPublisher,
				mockAuthObserver,
				nil, // attemptTracker
			)

			// Execute
			got, state, err := svc.GoogleOAuth(context.Background(), tt.redirectURI)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			if tt.state != "" {
				assert.Equal(t, tt.state, state)
			}
		})
	}
}

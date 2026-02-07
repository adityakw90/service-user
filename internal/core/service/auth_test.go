package service

import (
	"context"
	"errors"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_Authenticate(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockUserRepository, *MockDeviceRepository, *MockUserDeviceRepository, *MockHasher, *MockUIDGenerator, *MockTokenGenerator)
		input       *params.AuthParams
		want        *model.Token
		wantErr     error
		assertError func(t *testing.T, err error)
	}{
		{
			name: "Happy Path - email authentication",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				h.CompareFunc = func(hashed, plain string) bool { return true }
				ug.NewFunc = func() string { return "session-123" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "access_token", nil
					}
					return "refresh_token", nil
				}
			},
			input: createAuthParams("test@example.com", "email", "password", "", "", ""),
			want: &model.Token{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			wantErr: nil,
		},
		{
			name: "Happy Path - username authentication",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				ur.GetByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				h.CompareFunc = func(hashed, plain string) bool { return true }
				ug.NewFunc = func() string { return "session-123" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "access_token", nil
					}
					return "refresh_token", nil
				}
			},
			input: createAuthParams("testuser", "username", "password", "", "", ""),
			want: &model.Token{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			wantErr: nil,
		},
		{
			name: "Error - user not found returns invalid credentials",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				ur.GetByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			input:   createAuthParams("nonexistent@example.com", "email", "password", "", "", ""),
			wantErr: domainerrors.ErrInvalidCredentials,
		},
		{
			name: "Error - deleted user",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil
				}
				h.CompareFunc = func(hashed, plain string) bool { return true }
			},
			input:   createAuthParams("test@example.com", "email", "password", "", "", ""),
			wantErr: domainerrors.ErrUserDeleted,
		},
		{
			name: "Error - inactive user",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusInactive), nil
				}
				h.CompareFunc = func(hashed, plain string) bool { return true }
			},
			input:   createAuthParams("test@example.com", "email", "password", "", "", ""),
			wantErr: domainerrors.ErrUserInactive,
		},
		{
			name: "Happy Path - with device tracking new device",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				h.CompareFunc = func(hashed, plain string) bool { return true }
				dr.GetByFingerprintFunc = func(ctx context.Context, fingerprint string) (*model.Device, error) {
					return nil, domainerrors.ErrDeviceNotFound
				}
				dr.CreateFunc = func(ctx context.Context, device *model.Device) (*model.Device, error) {
					return createTestDevice(1, "device-uid", "iPhone", "fp123"), nil
				}
				udr.GetByUserIDAndDeviceIDFunc = func(ctx context.Context, userID int64, deviceID int64) (*model.UserDevice, error) {
					return nil, domainerrors.ErrUserDeviceNotFound
				}
				udr.CreateFunc = func(ctx context.Context, ud *model.UserDevice) (*model.UserDevice, error) {
					return createUserDevice(1, 1, "192.168.1.1"), nil
				}
				ug.NewFunc = func() string { return "session-123" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "access_token", nil
					}
					return "refresh_token", nil
				}
			},
			input: createAuthParams("test@example.com", "email", "password", "iPhone", "fp123", "192.168.1.1"),
			want: &model.Token{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			wantErr: nil,
		},
		{
			name: "Happy Path - with device tracking existing device",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				h.CompareFunc = func(hashed, plain string) bool { return true }
				dr.GetByFingerprintFunc = func(ctx context.Context, fingerprint string) (*model.Device, error) {
					return createTestDevice(1, "device-uid", "iPhone", "fp123"), nil
				}
				udr.GetByUserIDAndDeviceIDFunc = func(ctx context.Context, userID int64, deviceID int64) (*model.UserDevice, error) {
					return createUserDevice(1, 1, "192.168.1.1"), nil
				}
				ug.NewFunc = func() string { return "session-123" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "access_token", nil
					}
					return "refresh_token", nil
				}
			},
			input: createAuthParams("test@example.com", "email", "password", "iPhone", "fp123", "192.168.1.1"),
			want: &model.Token{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			wantErr: nil,
		},
		{
			name: "Happy Path - no device fingerprint",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository, h *MockHasher, ug *MockUIDGenerator, tg *MockTokenGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				h.CompareFunc = func(hashed, plain string) bool { return true }
				ug.NewFunc = func() string { return "session-123" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "access_token", nil
					}
					return "refresh_token", nil
				}
			},
			input: createAuthParams("test@example.com", "email", "password", "", "", ""),
			want: &model.Token{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()
			mockTokenGen := NewMockTokenGenerator()
			mockPinRepo := NewMockUserPinRepository()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockOAuthProvider := NewMockOAuthProvider()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockDeviceRepo, mockUserDeviceRepo, mockHasher, mockUIDGen, mockTokenGen)
			}

			// Create service
			svc := NewAuthService(
				mockUserRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPinRepo,
				mockHasher,
				mockHasher, // pin hasher same as password hasher for tests
				mockTokenGen,
				mockUIDGen,
				mockOAuthProvider,
				mockTokenWhitelist,
				mockTokenBlacklist,
				mockEventPublisher,
				mockAuthObserver,
			)

			// Execute
			got, err := svc.Authenticate(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				if tt.assertError != nil {
					tt.assertError(t, err)
				} else {
					assert.ErrorIs(t, err, tt.wantErr)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthService_GoogleOAuth(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockUIDGenerator, *MockOAuthProvider)
		redirectURI string
		want        string
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
			want:        "https://accounts.google.com/o/oauth2/v2/auth?state=state-123",
			wantErr:     nil,
		},
		{
			name: "Error - OAuth provider failure",
			setupMocks: func(ug *MockUIDGenerator, oauth *MockOAuthProvider) {
				ug.NewFunc = func() string { return "state-123" }
				oauth.GetAuthorizationURLFunc = func(ctx context.Context, redirectURI, state string) (string, error) {
					return "", errors.New("oauth error")
				}
			},
			redirectURI: "http://localhost:8080/callback",
			wantErr:     domainerrors.ErrOAuthExchangeFailed,
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
			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockOAuthProvider := NewMockOAuthProvider()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUIDGen, mockOAuthProvider)
			}

			// Create service
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
			)

			// Execute
			got, err := svc.GoogleOAuth(context.Background(), tt.redirectURI)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthService_HandleGoogleOAuth(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*MockUserRepository, *MockOAuthProvider, *MockUIDGenerator, *MockTokenGenerator, *MockTokenStore)
		code        string
		redirectURI string
		want        *model.Token
		wantErr     error
	}{
		{
			name: "Happy Path - new user creation",
			setupMocks: func(ur *MockUserRepository, oauth *MockOAuthProvider, ug *MockUIDGenerator, tg *MockTokenGenerator, tw *MockTokenStore) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				ur.CreateFunc = func(ctx context.Context, user *model.User) (*model.User, error) {
					return createTestUser(1, "new-uid", "testuser", "test@example.com", "", model.UserStatusActive), nil
				}
				oauth.ExchangeCodeFunc = func(ctx context.Context, code, redirectURI string) (*model.OAuthTokens, error) {
					return &model.OAuthTokens{AccessToken: "oauth_access"}, nil
				}
				oauth.GetUserInfoFunc = func(ctx context.Context, accessToken string) (*model.OAuthUserInfo, error) {
					return &model.OAuthUserInfo{Email: "test@example.com", FirstName: "Test", LastName: "User"}, nil
				}
				ug.NewFunc = func() string { return "session-123" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "access_token", nil
					}
					return "refresh_token", nil
				}
			},
			code:        "auth-code-123",
			redirectURI: "http://localhost:8080/callback",
			want: &model.Token{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			wantErr: nil,
		},
		{
			name: "Happy Path - existing user login",
			setupMocks: func(ur *MockUserRepository, oauth *MockOAuthProvider, ug *MockUIDGenerator, tg *MockTokenGenerator, tw *MockTokenStore) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(1, "existing-uid", "testuser", "test@example.com", "", model.UserStatusActive), nil
				}
				oauth.ExchangeCodeFunc = func(ctx context.Context, code, redirectURI string) (*model.OAuthTokens, error) {
					return &model.OAuthTokens{AccessToken: "oauth_access"}, nil
				}
				oauth.GetUserInfoFunc = func(ctx context.Context, accessToken string) (*model.OAuthUserInfo, error) {
					return &model.OAuthUserInfo{Email: "test@example.com", FirstName: "Test", LastName: "User"}, nil
				}
				ug.NewFunc = func() string { return "session-123" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "access_token", nil
					}
					return "refresh_token", nil
				}
			},
			code:        "auth-code-123",
			redirectURI: "http://localhost:8080/callback",
			want: &model.Token{
				Access:  "access_token",
				Refresh: "refresh_token",
			},
			wantErr: nil,
		},
		{
			name: "Error - token exchange failure",
			setupMocks: func(ur *MockUserRepository, oauth *MockOAuthProvider, ug *MockUIDGenerator, tg *MockTokenGenerator, tw *MockTokenStore) {
				oauth.ExchangeCodeFunc = func(ctx context.Context, code, redirectURI string) (*model.OAuthTokens, error) {
					return nil, errors.New("exchange failed")
				}
			},
			code:        "invalid-code",
			redirectURI: "http://localhost:8080/callback",
			wantErr:     domainerrors.ErrOAuthExchangeFailed,
		},
		{
			name: "Error - user info fetch failure",
			setupMocks: func(ur *MockUserRepository, oauth *MockOAuthProvider, ug *MockUIDGenerator, tg *MockTokenGenerator, tw *MockTokenStore) {
				oauth.ExchangeCodeFunc = func(ctx context.Context, code, redirectURI string) (*model.OAuthTokens, error) {
					return &model.OAuthTokens{AccessToken: "oauth_access"}, nil
				}
				oauth.GetUserInfoFunc = func(ctx context.Context, accessToken string) (*model.OAuthUserInfo, error) {
					return nil, errors.New("fetch failed")
				}
			},
			code:        "auth-code-123",
			redirectURI: "http://localhost:8080/callback",
			wantErr:     domainerrors.ErrOAuthUserInfoFailed,
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
			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockOAuthProvider := NewMockOAuthProvider()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockOAuthProvider, mockUIDGen, mockTokenGen, mockTokenWhitelist)
			}

			// Create service
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
			)

			// Execute
			got, err := svc.HandleGoogleOAuth(context.Background(), tt.code, tt.redirectURI)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockTokenGenerator, *MockUIDGenerator, *MockTokenStore)
		token      string
		want       *model.Token
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *MockUserRepository, tg *MockTokenGenerator, ug *MockUIDGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid:            "test-uid",
						Sid:            "old-session",
						Type:           model.TokenTypeRefresh,
						Identifier:     "test@example.com",
						IdentifierType: "email",
					}, nil
				}
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				tw.IsAllowedFunc = func(ctx context.Context, userUID string, tid string) (bool, error) {
					return true, nil
				}
				ug.NewFunc = func() string { return "new-session" }
				tg.GenerateTokenFunc = func(claims *model.TokenClaims) (string, error) {
					if claims.Type == model.TokenTypeAccess {
						return "new_access_token", nil
					}
					return "new_refresh_token", nil
				}
			},
			token: "valid_refresh_token",
			want: &model.Token{
				Access:  "new_access_token",
				Refresh: "new_refresh_token",
			},
			wantErr: nil,
		},
		{
			name: "Error - expired token",
			setupMocks: func(ur *MockUserRepository, tg *MockTokenGenerator, ug *MockUIDGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return nil, domainerrors.ErrTokenExpired
				}
			},
			token:   "expired_token",
			wantErr: domainerrors.ErrTokenExpired,
		},
		{
			name: "Error - invalid token type (access token instead of refresh)",
			setupMocks: func(ur *MockUserRepository, tg *MockTokenGenerator, ug *MockUIDGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid:            "test-uid",
						Sid:            "session",
						Type:           model.TokenTypeAccess,
						Identifier:     "test@example.com",
						IdentifierType: "email",
					}, nil
				}
			},
			token:   "access_token_instead_of_refresh",
			wantErr: domainerrors.ErrTokenInvalid,
		},
		{
			name: "Error - token not in whitelist (revoked)",
			setupMocks: func(ur *MockUserRepository, tg *MockTokenGenerator, ug *MockUIDGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid:            "test-uid",
						Sid:            "revoked-session",
						Type:           model.TokenTypeRefresh,
						Identifier:     "test@example.com",
						IdentifierType: "email",
					}, nil
				}
				tw.IsAllowedFunc = func(ctx context.Context, userUID string, tid string) (bool, error) {
					return false, nil
				}
			},
			token:   "revoked_token",
			wantErr: domainerrors.ErrTokenRevoked,
		},
		{
			name: "Error - user deleted",
			setupMocks: func(ur *MockUserRepository, tg *MockTokenGenerator, ug *MockUIDGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid:            "test-uid",
						Sid:            "session",
						Type:           model.TokenTypeRefresh,
						Identifier:     "test@example.com",
						IdentifierType: "email",
					}, nil
				}
				tw.IsAllowedFunc = func(ctx context.Context, userUID string, tid string) (bool, error) {
					return true, nil
				}
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil
				}
			},
			token:   "valid_token_but_user_deleted",
			wantErr: domainerrors.ErrUserDeleted,
		},
		{
			name: "Error - user inactive",
			setupMocks: func(ur *MockUserRepository, tg *MockTokenGenerator, ug *MockUIDGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid:            "test-uid",
						Sid:            "session",
						Type:           model.TokenTypeRefresh,
						Identifier:     "test@example.com",
						IdentifierType: "email",
					}, nil
				}
				tw.IsAllowedFunc = func(ctx context.Context, userUID string, tid string) (bool, error) {
					return true, nil
				}
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusInactive), nil
				}
			},
			token:   "valid_token_but_user_inactive",
			wantErr: domainerrors.ErrUserInactive,
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
			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockOAuthProvider := NewMockOAuthProvider()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockTokenGen, mockUIDGen, mockTokenWhitelist)
			}

			// Create service
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
			)

			// Execute
			got, err := svc.RefreshToken(context.Background(), tt.token)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	tests := []struct {
		name              string
		setupMocks        func(*MockTokenGenerator)
		setupTokenWhitelist func(*MockTokenStore)
		token             string
		want              *model.TokenClaims
		wantErr           error
	}{
		{
			name: "Happy Path - valid access token",
			setupMocks: func(tg *MockTokenGenerator) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid:            "test-uid",
						Sid:            "session-123",
						Type:           model.TokenTypeAccess,
						Identifier:     "test@example.com",
						IdentifierType: "email",
					}, nil
				}
			},
			setupTokenWhitelist: func(tw *MockTokenStore) {
				tw.IsAllowedFunc = func(ctx context.Context, userUID string, tid string) (bool, error) {
					return true, nil
				}
			},
			token: "valid_access_token",
			want: &model.TokenClaims{
				Uid:            "test-uid",
				Sid:            "session-123",
				Type:           model.TokenTypeAccess,
				Identifier:     "test@example.com",
				IdentifierType: "email",
			},
			wantErr: nil,
		},
		{
			name: "Error - expired token",
			setupMocks: func(tg *MockTokenGenerator) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return nil, domainerrors.ErrTokenExpired
				}
			},
			token:   "expired_token",
			wantErr: domainerrors.ErrTokenExpired,
		},
		{
			name: "Error - invalid token type (refresh token)",
			setupMocks: func(tg *MockTokenGenerator) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid:            "test-uid",
						Sid:            "session-123",
						Type:           model.TokenTypeRefresh,
						Identifier:     "test@example.com",
						IdentifierType: "email",
					}, nil
				}
			},
			token:   "refresh_token",
			wantErr: domainerrors.ErrTokenInvalid,
		},
		{
			name: "Error - invalid signature",
			setupMocks: func(tg *MockTokenGenerator) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return nil, domainerrors.ErrTokenInvalid
				}
			},
			token:   "invalid_token",
			wantErr: domainerrors.ErrTokenInvalid,
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
			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockOAuthProvider := NewMockOAuthProvider()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockTokenGen)
			}

			// Setup token whitelist expectations
			if tt.setupTokenWhitelist != nil {
				tt.setupTokenWhitelist(mockTokenWhitelist)
			}

			// Create service
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
			)

			// Execute
			got, err := svc.ValidateToken(context.Background(), tt.token)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthService_RevokeToken(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockTokenGenerator, *MockTokenStore)
		token      string
		tokenType  string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(tg *MockTokenGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return &model.TokenClaims{
						Uid: "test-uid",
						Sid: "session-123",
					}, nil
				}
				tw.RemoveFunc = func(ctx context.Context, userUID string, tid string) error {
					return nil
				}
			},
			token:     "valid_token",
			tokenType: "refresh",
			wantErr:   nil,
		},
		{
			name: "Error - invalid token",
			setupMocks: func(tg *MockTokenGenerator, tw *MockTokenStore) {
				tg.ValidateTokenFunc = func(token string) (*model.TokenClaims, error) {
					return nil, domainerrors.ErrTokenInvalid
				}
			},
			token:     "invalid_token",
			tokenType: "refresh",
			wantErr:   domainerrors.ErrTokenInvalid,
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
			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockOAuthProvider := NewMockOAuthProvider()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockTokenGen, mockTokenWhitelist)
			}

			// Create service
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
			)

			// Execute
			err := svc.RevokeToken(context.Background(), tt.token, tt.tokenType)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestAuthService_VerifyPin(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockUserPinRepository, *MockHasher)
		userUID    string
		pin        string
		want       bool
		wantErr    error
	}{
		{
			name: "Happy Path - valid PIN",
			setupMocks: func(ur *MockUserRepository, pr *MockUserPinRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserPin, error) {
					return createUserPin(1, "test-uid", "hashed_1234"), nil
				}
				h.CompareFunc = func(hashed, plain string) bool {
					// Note: The service implementation has arguments in wrong order
					return hashed == "hashed_1234" && plain == "1234"
				}
			},
			userUID: "test-uid",
			pin:     "1234",
			want:    true,
			wantErr: nil,
		},
		{
			name: "Invalid PIN returns false",
			setupMocks: func(ur *MockUserRepository, pr *MockUserPinRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserPin, error) {
					return createUserPin(1, "test-uid", "hashed_1234"), nil
				}
				h.CompareFunc = func(hashed, plain string) bool {
					return hashed == "hashed_1234" && plain == "9999" && false // Always return false for this test
				}
			},
			userUID: "test-uid",
			pin:     "9999",
			want:    false,
			wantErr: nil,
		},
		{
			name: "Error - PIN not set",
			setupMocks: func(ur *MockUserRepository, pr *MockUserPinRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserPin, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			userUID: "test-uid",
			pin:     "1234",
			want:    false,
			wantErr: domainerrors.ErrPinNotSet,
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository, pr *MockUserPinRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			userUID: "nonexistent-uid",
			pin:     "1234",
			want:    false,
			wantErr: domainerrors.ErrUserNotFound,
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
			mockTokenGen := NewMockTokenGenerator()
			mockTokenWhitelist := NewMockTokenStore()
			mockTokenBlacklist := NewMockTokenStore()
			mockOAuthProvider := NewMockOAuthProvider()
			mockEventPublisher := NewMockEventPublisher()
			mockAuthObserver := NewMockAuthObserver()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockPinRepo, mockHasher)
			}

			// Create service
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
			)

			// Execute
			got, err := svc.VerifyPin(context.Background(), tt.userUID, tt.pin)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

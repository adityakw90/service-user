package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	servicemocks "github.com/adityakw90/service-user/test/mocks/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authpb "github.com/adityakw90/service-user-proto/gen/go/auth"
)

// TestNewAuthHandler tests the NewAuthHandler constructor.
func TestNewAuthHandler(t *testing.T) {
	mockService := servicemocks.NewMockAuthService(t)
	v := validator.New()

	h := NewAuthHandler(mockService, v)

	assert.NotNil(t, h)
	assert.Equal(t, mockService, h.service)
	assert.Equal(t, v, h.validator)
}

// TestAuthHandler_GoogleOAuth tests the GoogleOAuth handler method.
func TestAuthHandler_GoogleOAuth(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*servicemocks.MockAuthService)
		input       *authpb.GoogleOAuthRequest
		want        *authpb.GoogleOAuthResponse
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "Happy Path - service returns URL",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().GoogleOAuth(mock.Anything, "http://localhost:8080/callback").Return(
					"https://accounts.google.com/o/oauth2/v2/auth?client_id=test&redirect_uri=http://localhost:8080/callback&response_type=code&scope=openid+email+profile&state=abc123&access_type=offline&prompt=consent",
					"abc123", nil).Once()
			},
			input: &authpb.GoogleOAuthRequest{
				RedirectUri: "http://localhost:8080/callback",
			},
			want: &authpb.GoogleOAuthResponse{
				AuthorizationUrl: "https://accounts.google.com/o/oauth2/v2/auth?client_id=test&redirect_uri=http://localhost:8080/callback&response_type=code&scope=openid+email+profile&state=abc123&access_type=offline&prompt=consent",
				State:            "abc123",
			},
			wantErr: false,
		},
		{
			name: "Happy Path - minimal URL",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().GoogleOAuth(mock.Anything, "http://localhost:8080/callback").Return(
					"https://accounts.google.com/o/oauth2/v2/auth?state=xyz", "xyz", nil).Once()
			},
			input: &authpb.GoogleOAuthRequest{
				RedirectUri: "http://localhost:8080/callback",
			},
			want: &authpb.GoogleOAuthResponse{
				AuthorizationUrl: "https://accounts.google.com/o/oauth2/v2/auth?state=xyz",
				State:            "xyz",
			},
			wantErr: false,
		},
		{
			name: "Service Error - internal error from service",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().GoogleOAuth(mock.Anything, "http://localhost:8080/callback").Return(
					"", "", errors.New("service unavailable")).Once()
			},
			input: &authpb.GoogleOAuthRequest{
				RedirectUri: "http://localhost:8080/callback",
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "internal server error",
		},
		{
			name:       "Invalid Input - empty redirect URI",
			setupMocks: func(m *servicemocks.MockAuthService) {},
			input: &authpb.GoogleOAuthRequest{
				RedirectUri: "",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			errContains: "validation error",
		},
		{
			name:       "Invalid Input - invalid URI format",
			setupMocks: func(m *servicemocks.MockAuthService) {},
			input: &authpb.GoogleOAuthRequest{
				RedirectUri: "not-a-valid-uri",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			errContains: "validation error",
		},
		{
			name: "Service returns context canceled error",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().GoogleOAuth(mock.Anything, "http://localhost:8080/callback").Return(
					"", "", context.Canceled).Once()
			},
			input: &authpb.GoogleOAuthRequest{
				RedirectUri: "http://localhost:8080/callback",
			},
			wantErr:     true,
			wantCode:    codes.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := servicemocks.NewMockAuthService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			// Execute
			got, err := handler.GoogleOAuth(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				// Convert error through middleware to simulate actual flow
				grpcErr := response.MakeErrorResponse(err)
				st, ok := status.FromError(grpcErr)
				require.True(t, ok, "error should be a gRPC status error after conversion")
				assert.Equal(t, tt.wantCode, st.Code())
				if tt.errContains != "" {
					assert.Contains(t, st.Message(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.AuthorizationUrl, got.AuthorizationUrl)
		})
	}
}

// TestAuthHandler_HandleGoogleOAuth tests the HandleGoogleOAuth handler method.
func TestAuthHandler_HandleGoogleOAuth(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*servicemocks.MockAuthService)
		input       *authpb.HandleGoogleOAuthRequest
		want        *authpb.Token
		wantErr     bool
		wantCode    codes.Code
		errContains string
	}{
		{
			name: "Happy Path - service returns tokens",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().HandleGoogleOAuth(mock.Anything, "valid-auth-code-abc", "test-state-abc", "http://localhost:8080/callback").Return(
					&model.Token{
						Access:  "new-access-token-123",
						Refresh: "new-refresh-token-456",
					}, nil).Once()
			},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "valid-auth-code-abc",
				State:        "test-state-abc",
				RedirectUri: "http://localhost:8080/callback",
			},
			want: &authpb.Token{
				AccessToken:  "new-access-token-123",
				RefreshToken: "new-refresh-token-456",
			},
			wantErr: false,
		},
		{
			name: "Happy Path - token exchange with different redirect URI",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().HandleGoogleOAuth(mock.Anything, "code-for-custom-redirect", "test-state-custom", "https://example.com/oauth/callback").Return(
					&model.Token{
						Access:  "access-from-custom-redirect",
						Refresh: "refresh-from-custom-redirect",
					}, nil).Once()
			},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "code-for-custom-redirect",
				State:        "test-state-custom",
				RedirectUri: "https://example.com/oauth/callback",
			},
			want: &authpb.Token{
				AccessToken:  "access-from-custom-redirect",
				RefreshToken: "refresh-from-custom-redirect",
			},
			wantErr: false,
		},
		{
			name: "Service Error - OAuth exchange failed",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().HandleGoogleOAuth(mock.Anything, "invalid-code", "test-state-error", "http://localhost:8080/callback").Return(
					nil, errors.New("invalid authorization code")).Once()
			},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "invalid-code",
				State:        "test-state-error",
				RedirectUri: "http://localhost:8080/callback",
			},
			wantErr:     true,
			wantCode:    codes.Internal,
			errContains: "internal server error",
		},
		{
			name:       "Invalid Input - empty code",
			setupMocks: func(m *servicemocks.MockAuthService) {},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "",
				State:        "",
				RedirectUri: "http://localhost:8080/callback",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			errContains: "validation error",
		},
		{
			name:       "Invalid Input - empty redirect URI",
			setupMocks: func(m *servicemocks.MockAuthService) {},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "valid-code",
				State:        "test-state",
				RedirectUri: "",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			errContains: "validation error",
		},
		{
			name:       "Invalid Input - both code and redirect URI empty",
			setupMocks: func(m *servicemocks.MockAuthService) {},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "",
				State:        "",
				RedirectUri: "",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
		},
		{
			name:       "Invalid Input - invalid URI format",
			setupMocks: func(m *servicemocks.MockAuthService) {},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "valid-code",
				State:        "test-state",
				RedirectUri: "not-a-valid-uri",
			},
			wantErr:     true,
			wantCode:    codes.InvalidArgument,
			errContains: "validation error",
		},
		{
			name: "Service returns context canceled error",
			setupMocks: func(m *servicemocks.MockAuthService) {
				m.EXPECT().HandleGoogleOAuth(mock.Anything, "valid-code", "test-state", "http://localhost:8080/callback").Return(
					nil, context.Canceled).Once()
			},
			input: &authpb.HandleGoogleOAuthRequest{
				Code:        "valid-code",
				State:        "test-state",
				RedirectUri: "http://localhost:8080/callback",
			},
			wantErr:     true,
			wantCode:    codes.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockService := servicemocks.NewMockAuthService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(mockService)
			}

			// Create handler
			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			// Execute
			got, err := handler.HandleGoogleOAuth(context.Background(), tt.input)

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				// Convert error through middleware to simulate actual flow
				grpcErr := response.MakeErrorResponse(err)
				st, ok := status.FromError(grpcErr)
				require.True(t, ok, "error should be a gRPC status error after conversion")
				assert.Equal(t, tt.wantCode, st.Code())
				if tt.errContains != "" {
					assert.Contains(t, st.Message(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.AccessToken, got.AccessToken)
			assert.Equal(t, tt.want.RefreshToken, got.RefreshToken)
		})
	}
}

// TestAuthHandler_GoogleOAuth_ValidURIVariations tests various valid URI formats.
func TestAuthHandler_GoogleOAuth_ValidURIVariations(t *testing.T) {
	tests := []struct {
		name        string
		redirectURI string
	}{
		{
			name:        "HTTP localhost",
			redirectURI: "http://localhost:8080/callback",
		},
		{
			name:        "HTTPS URL",
			redirectURI: "https://example.com/oauth/callback",
		},
		{
			name:        "URL with query parameters",
			redirectURI: "https://example.com/callback?param1=value1&param2=value2",
		},
		{
			name:        "URL with port",
			redirectURI: "https://example.com:8443/callback",
		},
		{
			name:        "URL with fragment",
			redirectURI: "https://example.com/callback#fragment",
		},
		{
			name:        "URL with path",
			redirectURI: "https://example.com/api/v1/oauth/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := servicemocks.NewMockAuthService(t)
			mockService.EXPECT().GoogleOAuth(mock.Anything, tt.redirectURI).Return(
				"https://accounts.google.com/o/oauth2/v2/auth?state=test", "test", nil).Once()

			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			req := &authpb.GoogleOAuthRequest{
				RedirectUri: tt.redirectURI,
			}

			got, err := handler.GoogleOAuth(context.Background(), req)

			require.NoError(t, err)
			assert.NotEmpty(t, got.AuthorizationUrl)
			assert.Equal(t, "test", got.State)
		})
	}
}

// TestAuthHandler_HandleGoogleOAuth_ValidURIVariations tests various valid URI formats for callback.
func TestAuthHandler_HandleGoogleOAuth_ValidURIVariations(t *testing.T) {
	tests := []struct {
		name        string
		redirectURI string
		code        string
		state       string
	}{
		{
			name:        "HTTP localhost",
			redirectURI: "http://localhost:8080/callback",
			code:        "auth-code-localhost",
			state:       "state-localhost",
		},
		{
			name:        "HTTPS URL",
			redirectURI: "https://example.com/oauth/callback",
			code:        "auth-code-https",
			state:       "state-https",
		},
		{
			name:        "URL with query parameters",
			redirectURI: "https://example.com/callback?param1=value1",
			code:        "auth-code-with-params",
			state:       "state-with-params",
		},
		{
			name:        "URL with port",
			redirectURI: "https://example.com:8443/callback",
			code:        "auth-code-with-port",
			state:       "state-with-port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := servicemocks.NewMockAuthService(t)
			mockService.EXPECT().HandleGoogleOAuth(mock.Anything, tt.code, tt.state, tt.redirectURI).Return(
				&model.Token{
					Access:  "test-access-token",
					Refresh: "test-refresh-token",
				}, nil).Once()

			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			req := &authpb.HandleGoogleOAuthRequest{
				Code:        tt.code,
				State:        tt.state,
				RedirectUri: tt.redirectURI,
			}

			got, err := handler.HandleGoogleOAuth(context.Background(), req)

			require.NoError(t, err)
			assert.NotEmpty(t, got.AccessToken)
			assert.NotEmpty(t, got.RefreshToken)
		})
	}
}

// TestAuthHandler_GoogleOAuth_InvalidInputVariations tests various invalid input scenarios.
func TestAuthHandler_GoogleOAuth_InvalidInputVariations(t *testing.T) {
	tests := []struct {
		name        string
		redirectURI string
		errField    string
	}{
		{
			name:     "Empty string",
			redirectURI: "",
			errField: "RedirectUri",
		},
		{
			name:     "Not a URI - plain text",
			redirectURI: "just-text",
			errField: "RedirectUri",
		},
		{
			name:     "Not a URI - missing scheme",
			redirectURI: "example.com/callback",
			errField: "RedirectUri",
		},
		{
			name:     "Whitespace only",
			redirectURI: "   ",
			errField: "RedirectUri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := servicemocks.NewMockAuthService(t)
			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			req := &authpb.GoogleOAuthRequest{
				RedirectUri: tt.redirectURI,
			}

			_, err := handler.GoogleOAuth(context.Background(), req)

			require.Error(t, err)
			// Convert error through middleware to simulate actual flow
			grpcErr := response.MakeErrorResponse(err)
			st, ok := status.FromError(grpcErr)
			require.True(t, ok, "error should be a gRPC status error after conversion")
			assert.Equal(t, codes.InvalidArgument, st.Code())
			assert.Contains(t, st.Message(), "validation error")
		})
	}
}

// TestAuthHandler_HandleGoogleOAuth_InvalidInputVariations tests various invalid input scenarios.
func TestAuthHandler_HandleGoogleOAuth_InvalidInputVariations(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		state        string
		redirectURI string
		errField    string
	}{
		{
			name:        "Empty code",
			code:        "",
			state:        "",
			redirectURI: "http://localhost:8080/callback",
			errField:    "Code",
		},
		{
			name:        "Whitespace code",
			code:        "   ",
			state:        "test-state",
			redirectURI: "http://localhost:8080/callback",
			errField:    "Code",
		},
		{
			name:        "Empty redirect URI",
			code:        "valid-code",
			state:        "test-state",
			redirectURI: "",
			errField:    "RedirectUri",
		},
		{
			name:        "Both empty",
			code:        "",
			state:        "",
			redirectURI: "",
			errField:    "",
		},
		{
			name:        "Invalid redirect URI",
			code:        "valid-code",
			state:        "test-state",
			redirectURI: "not-a-uri",
			errField:    "RedirectUri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := servicemocks.NewMockAuthService(t)
			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			req := &authpb.HandleGoogleOAuthRequest{
				Code:        tt.code,
				State:        tt.state,
				RedirectUri: tt.redirectURI,
			}

			_, err := handler.HandleGoogleOAuth(context.Background(), req)

			require.Error(t, err)
			// Convert error through middleware to simulate actual flow
			grpcErr := response.MakeErrorResponse(err)
			st, ok := status.FromError(grpcErr)
			require.True(t, ok, "error should be a gRPC status error after conversion")
			assert.Equal(t, codes.InvalidArgument, st.Code())
			assert.Contains(t, st.Message(), "validation error")
		})
	}
}

// TestAuthHandler_GoogleOAuth_WhitespaceHandling tests that whitespace in URI is accepted by validator.
func TestAuthHandler_GoogleOAuth_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name        string
		redirectURI string
		expectedURI string
	}{
		{
			name:        "Leading whitespace - validator accepts it",
			redirectURI: "  http://localhost:8080/callback",
			expectedURI: "  http://localhost:8080/callback",
		},
		{
			name:        "Trailing whitespace - validator accepts it",
			redirectURI: "http://localhost:8080/callback  ",
			expectedURI: "http://localhost:8080/callback  ",
		},
		{
			name:        "Both leading and trailing whitespace - validator accepts it",
			redirectURI: "  http://localhost:8080/callback  ",
			expectedURI: "  http://localhost:8080/callback  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := servicemocks.NewMockAuthService(t)
			mockService.EXPECT().GoogleOAuth(mock.Anything, tt.redirectURI).Return(
				"https://accounts.google.com/o/oauth2/v2/auth", "test", nil).Once()

			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			req := &authpb.GoogleOAuthRequest{
				RedirectUri: tt.redirectURI,
			}

			got, err := handler.GoogleOAuth(context.Background(), req)

			// The URI validator accepts leading/trailing whitespace
			require.NoError(t, err)
			assert.Equal(t, "test", got.State, "state should be returned")
		})
	}
}

// TestAuthHandler_HandleGoogleOAuth_WhitespaceHandling tests that whitespace in inputs is accepted by validator.
func TestAuthHandler_HandleGoogleOAuth_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		state       string
		redirectURI string
	}{
		{
			name:        "Code with leading whitespace - validator accepts it",
			code:        "  auth-code-123",
			state:        "test-state",
			redirectURI: "http://localhost:8080/callback",
		},
		{
			name:        "Code with trailing whitespace - validator accepts it",
			code:        "auth-code-456  ",
			state:        "test-state",
			redirectURI: "http://localhost:8080/callback",
		},
		{
			name:        "Redirect URI with leading whitespace - validator accepts it",
			code:        "auth-code-789",
			state:        "test-state",
			redirectURI: "  http://localhost:8080/callback",
		},
		{
			name:        "Valid code and redirect URI - passes",
			code:        "valid-code",
			state:        "test-state",
			redirectURI: "http://localhost:8080/callback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := servicemocks.NewMockAuthService(t)
			mockService.EXPECT().HandleGoogleOAuth(mock.Anything, tt.code, tt.state, tt.redirectURI).Return(
				&model.Token{
					Access:  "test-access",
					Refresh: "test-refresh",
				}, nil).Once()

			v := validator.New()
			handler := NewAuthHandler(mockService, v)

			req := &authpb.HandleGoogleOAuthRequest{
				Code:        tt.code,
				State:        tt.state,
				RedirectUri: tt.redirectURI,
			}

			_, err := handler.HandleGoogleOAuth(context.Background(), req)

			// The validator accepts leading/trailing whitespace
			require.NoError(t, err)
		})
	}
}

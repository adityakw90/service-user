package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestAdapter_Oauth_NewGoogleOAuth(t *testing.T) {
	tests := []struct {
		name        string
		config      *GoogleOAuthConfig
		checkFn     func(*GoogleOAuth)
		wantErr     bool
		wantErrType error
		wantErrMsg  string
	}{
		{
			name: "Happy Path",
			config: &GoogleOAuthConfig{
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			},
			checkFn: func(a *GoogleOAuth) {
				assert.Equal(t, "test-client-id", a.config.ClientID)
				assert.Equal(t, "test-client-secret", a.config.ClientSecret)
			},
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name: "Empty Client ID",
			config: &GoogleOAuthConfig{
				ClientID:     "",
				ClientSecret: "test-client-secret",
			},
			checkFn:     nil,
			wantErr:     true,
			wantErrType: domainErrors.ErrOAuthClientIDRequired,
		},
		{
			name: "Empty Client Secret",
			config: &GoogleOAuthConfig{
				ClientID:     "test-client-id",
				ClientSecret: "",
			},
			checkFn:     nil,
			wantErr:     true,
			wantErrType: domainErrors.ErrOAuthClientSecretRequired,
		},
		{
			name:        "nil config",
			config:      nil,
			checkFn:     nil,
			wantErr:     true,
			wantErrType: domainErrors.ErrOAuthClientIDRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGoogleOAuth(tt.config, nil)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.wantErrType, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, got)
				if tt.checkFn != nil {
					tt.checkFn(got)
				}
			}
		})
	}
}

func TestAdapter_Oauth_GetAuthorizationURL(t *testing.T) {
	tests := []struct {
		name        string
		state       string
		redirectURI string
		wantParams  map[string]string
		wantErr     bool
		wantErrType error
		wantErrMsg  string
	}{
		{
			name:        "Happy Path",
			state:       "test-state",
			redirectURI: "test-redirect-uri",
			wantParams: map[string]string{
				"client_id":     "test-client-id",
				"redirect_uri":  "test-redirect-uri",
				"response_type": "code",
				"scope":         "openid email profile",
				"state":         "test-state",
				"access_type":   "offline",
				"prompt":        "consent",
			},
			wantErr:    false,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			// Setup miniredis
			redisClient, redisCleanup, err := newMockRedis()
			require.NoError(t, err)
			defer redisCleanup()

			oauth, err := NewGoogleOAuth(
				&GoogleOAuthConfig{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
				},
				redisClient,
			)
			require.NoError(t, err)

			gotURL, err := oauth.GetAuthorizationURL(ctx, tt.state, tt.redirectURI)
			if tt.wantErr {
				require.Error(t, err)
				assert.Equal(t, tt.wantErrType, err)
			} else {
				require.NoError(t, err)
				assert.True(
					t,
					strings.HasPrefix(gotURL, "https://accounts.google.com/o/oauth2/auth?"),
					"URL should start with Google's OAuth endpoint",
				)
				// Parse and verify query parameters
				parsedURL, err := url.Parse(gotURL)
				require.NoError(t, err)

				queryParams := parsedURL.Query()
				// assert the code challenge and code challenge method are present
				assert.Equal(t, "S256", queryParams.Get("code_challenge_method"))
				assert.NotEmpty(t, queryParams.Get("code_challenge"))
				for key, wantValue := range tt.wantParams {
					gotValue := queryParams.Get(key)
					assert.Equal(t, wantValue, gotValue, "parameter %s should match", key)
				}
			}
		})
	}
}

func TestAdapter_Oauth_ExchangeCode(t *testing.T) {
	tests := []struct {
		name        string
		state       string
		redirectURI string
		tokenResp   map[string]any
		tokenStatus int
		setupRedis  func(t *testing.T, ctx context.Context, redisClient *redis.Client)
		wantErr     bool
		wantErrMsg  string
		verifyToken func(t *testing.T, got *model.OAuthTokens)
	}{
		{
			name:        "Happy Path",
			state:       "test-state",
			redirectURI: "http://localhost:8080/callback",
			tokenResp: map[string]any{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_in":    3600,
				"id_token":      "test-id-token",
				"token_type":    "Bearer",
			},
			tokenStatus: http.StatusOK,
			setupRedis: func(t *testing.T, ctx context.Context, redisClient *redis.Client) {
				err := redisClient.Set(ctx, "oauth:pkce:test-state", "mock-verifier", 10*time.Minute).Err()
				require.NoError(t, err)
			},
			wantErr: false,
			verifyToken: func(t *testing.T, got *model.OAuthTokens) {
				assert.Equal(t, "test-access-token", got.AccessToken)
				assert.Equal(t, "test-refresh-token", got.RefreshToken)
				assert.Equal(t, "test-id-token", got.IDToken)
				assert.Equal(t, "Bearer", got.TokenType)
				assert.Greater(t, got.ExpiresIn, 0)
			},
		},
		{
			name:        "Invalid Code - OAuth Server Error",
			state:       "test-state-error",
			redirectURI: "http://localhost:8080/callback",
			tokenResp: map[string]any{
				"error":             "invalid_grant",
				"error_description": "The authorization code is invalid",
			},
			tokenStatus: http.StatusBadRequest,
			setupRedis: func(t *testing.T, ctx context.Context, redisClient *redis.Client) {
				err := redisClient.Set(ctx, "oauth:pkce:test-state-error", "mock-verifier", 10*time.Minute).Err()
				require.NoError(t, err)
			},
			wantErr:    true,
			wantErrMsg: "failed to exchange code for tokens",
		},
		{
			name:        "Missing Challenge - Redis Returns Nil",
			state:       "test-state-missing",
			redirectURI: "http://localhost:8080/callback",
			tokenResp:   nil,
			tokenStatus: http.StatusOK,
			setupRedis: func(t *testing.T, ctx context.Context, redisClient *redis.Client) {
				// Don't set anything - simulate missing/expired challenge
			},
			wantErr:    true,
			wantErrMsg: "google code challenge not found",
		},
		{
			name:        "State Mismatch - Challenge Not Found",
			state:       "different-state",
			redirectURI: "http://localhost:8080/callback",
			tokenResp:   nil,
			tokenStatus: http.StatusOK,
			setupRedis: func(t *testing.T, ctx context.Context, redisClient *redis.Client) {
				// Set challenge for different state
				err := redisClient.Set(ctx, "oauth:pkce:original-state", "mock-verifier", 10*time.Minute).Err()
				require.NoError(t, err)
			},
			wantErr:    true,
			wantErrMsg: "google code challenge not found",
		},
		{
			name:        "Token Without ID Token",
			state:       "test-state-no-id",
			redirectURI: "http://localhost:8080/callback",
			tokenResp: map[string]any{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_in":    1800,
				"token_type":    "Bearer",
			},
			tokenStatus: http.StatusOK,
			setupRedis: func(t *testing.T, ctx context.Context, redisClient *redis.Client) {
				err := redisClient.Set(ctx, "oauth:pkce:test-state-no-id", "mock-verifier", 10*time.Minute).Err()
				require.NoError(t, err)
			},
			wantErr: false,
			verifyToken: func(t *testing.T, got *model.OAuthTokens) {
				assert.Equal(t, "test-access-token", got.AccessToken)
				assert.Equal(t, "test-refresh-token", got.RefreshToken)
				assert.Equal(t, "", got.IDToken)
				assert.Equal(t, "Bearer", got.TokenType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Setup mock HTTP server for token endpoint
			mockServer, cleanup := setupMockHTTPServer(t, &mockOAuthServerConfig{
				TokenResponse: tt.tokenResp,
				TokenStatus:   tt.tokenStatus,
			})
			defer cleanup()

			// Setup miniredis
			redisClient, redisCleanup, err := newMockRedis()
			require.NoError(t, err)
			defer redisCleanup()

			// Setup Redis state
			tt.setupRedis(t, ctx, redisClient)

			// Create OAuth adapter with custom endpoint
			oauth, err := NewGoogleOAuth(
				&GoogleOAuthConfig{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
					Endpoint: &oauth2.Endpoint{
						AuthURL:  "https://accounts.google.com/o/oauth2/auth",
						TokenURL: mockServer.URL + "/token",
					},
					HttpTimeout: 5 * time.Second,
				},
				redisClient,
			)
			require.NoError(t, err)

			// Execute ExchangeCode
			got, err := oauth.ExchangeCode(ctx, "auth-code", tt.state, tt.redirectURI)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMsg)
				assert.Nil(t, got)

				if strings.Contains(tt.wantErrMsg, "exchange code") {
					assert.True(t, errors.Is(err, domainErrors.ErrOAuthExchangeFailed), "error should wrap ErrOAuthExchangeFailed")
				} else if strings.Contains(tt.wantErrMsg, "code challenge") {
					assert.True(t, errors.Is(err, domainErrors.ErrOAuthCodeVerifierMissing), "error should wrap ErrOAuthCodeVerifierMissing")
				}

				cause := errors.Unwrap(err)
				require.NotNil(t, cause, "underlying cause should be retained")
				assert.Contains(t, cause.Error(), tt.wantErrMsg, "cause error should contain the failure description")
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				if tt.verifyToken != nil {
					tt.verifyToken(t, got)
				}
			}
		})
	}
}

func TestAdapter_Oauth_GetUserInfo(t *testing.T) {
	baseUserInfo := &googleUserResp{
		ID:            "google-12345",
		Email:         "test@example.com",
		VerifiedEmail: true,
		GivenName:     "Test",
		FamilyName:    "User",
		Name:          "Test User",
		Picture:       "http://example.com/pic.jpg",
		Locale:        "en",
	}

	tests := []struct {
		name           string
		userInfoResp   *googleUserResp
		userInfoStatus int
		accessToken    string
		wantErr        bool
		wantErrMsg     string
		verifyUser     func(t *testing.T, got *model.OAuthUserInfo)
	}{
		{
			name:           "Happy Path",
			userInfoResp:   baseUserInfo,
			userInfoStatus: http.StatusOK,
			accessToken:    "valid-token",
			wantErr:        false,
			verifyUser: func(t *testing.T, got *model.OAuthUserInfo) {
				assert.Equal(t, "google-12345", got.ProviderID)
				assert.Equal(t, "test@example.com", got.Email)
				assert.True(t, got.EmailVerified)
				assert.Equal(t, "Test", got.FirstName)
				assert.Equal(t, "User", got.LastName)
				assert.Equal(t, "Test User", got.FullName)
				assert.Equal(t, "http://example.com/pic.jpg", got.Picture)
				assert.Equal(t, "en", got.Locale)
			},
		},
		{
			name:           "Unauthorized - Invalid Access Token",
			userInfoResp:   nil,
			userInfoStatus: http.StatusUnauthorized,
			accessToken:    "invalid-token",
			wantErr:        true,
			wantErrMsg:     "userinfo error",
		},
		{
			name:           "Invalid Response - Malformed JSON",
			userInfoResp:   nil,
			userInfoStatus: http.StatusOK,
			accessToken:    "valid-token",
			wantErr:        true,
			wantErrMsg:     "failed to parse user info",
		},
		{
			name: "Partial User Info",
			userInfoResp: &googleUserResp{
				ID:            "google-67890",
				Email:         "partial@example.com",
				VerifiedEmail: false,
			},
			userInfoStatus: http.StatusOK,
			accessToken:    "valid-token",
			wantErr:        false,
			verifyUser: func(t *testing.T, got *model.OAuthUserInfo) {
				assert.Equal(t, "google-67890", got.ProviderID)
				assert.Equal(t, "partial@example.com", got.Email)
				assert.False(t, got.EmailVerified)
				assert.Equal(t, "", got.FirstName)
				assert.Equal(t, "", got.LastName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Create a mock HTTP server for user info endpoint
			userInfoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authHeader := r.Header.Get("Authorization")
				if !strings.HasPrefix(authHeader, "Bearer ") {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.userInfoStatus)

				if tt.userInfoResp != nil {
					json.NewEncoder(w).Encode(tt.userInfoResp)
				} else if tt.userInfoStatus == http.StatusOK {
					// Return malformed JSON for invalid response test
					w.Write([]byte("invalid json"))
				}
			}))
			defer userInfoServer.Close()

			// Setup miniredis
			redisClient, redisCleanup, err := newMockRedis()
			require.NoError(t, err)
			defer redisCleanup()

			// Create OAuth adapter
			oauth, err := NewGoogleOAuth(
				&GoogleOAuthConfig{
					ClientID:     "test-client-id",
					ClientSecret: "test-client-secret",
					HttpTimeout:  5 * time.Second,
				},
				redisClient,
			)
			require.NoError(t, err)

			// Create token with access token
			token := &model.OAuthTokens{
				AccessToken: tt.accessToken,
			}

			// Create a custom HTTP client that will redirect to our mock server
			// We need to override the httpClient field to intercept requests
			customClient := &http.Client{
				Timeout: 5 * time.Second,
				Transport: &mockTransport{
					serverURL: userInfoServer.URL,
				},
			}
			oauth.httpClient = customClient

			// Execute GetUserInfo
			got, err := oauth.GetUserInfo(ctx, token)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
				assert.True(t, errors.Is(err, domainErrors.ErrOAuthUserInfoFailed), "error should wrap ErrOAuthUserInfoFailed")

				cause := errors.Unwrap(err)
				require.NotNil(t, cause, "underlying cause should be retained")
				assert.Contains(t, cause.Error(), tt.wantErrMsg, "cause error should contain the failure description")
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				if tt.verifyUser != nil {
					tt.verifyUser(t, got)
				}
			}
		})
	}
}

// TestAdapter_Oauth_GetUserInfo_NetworkError tests network error handling.
func TestAdapter_Oauth_GetUserInfo_NetworkError(t *testing.T) {
	ctx := context.Background()

	// Setup miniredis
	redisClient, redisCleanup, err := newMockRedis()
	require.NoError(t, err)
	defer redisCleanup()

	// Create OAuth adapter
	oauth, err := NewGoogleOAuth(
		&GoogleOAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			HttpTimeout:  1 * time.Millisecond, // Very short timeout
		},
		redisClient,
	)
	require.NoError(t, err)

	// Create a custom HTTP client that will fail
	customClient := &http.Client{
		Timeout: 1 * time.Millisecond,
		Transport: &failingTransport{
			err: errors.New("network error"),
		},
	}
	oauth.httpClient = customClient

	token := &model.OAuthTokens{
		AccessToken: "valid-token",
	}

	// Execute GetUserInfo - should fail with network error
	_, err = oauth.GetUserInfo(ctx, token)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user info")
	assert.True(t, errors.Is(err, domainErrors.ErrOAuthUserInfoFailed), "error should wrap ErrOAuthUserInfoFailed")

	cause := errors.Unwrap(err)
	require.NotNil(t, cause, "underlying cause should be retained")
	assert.Contains(t, cause.Error(), "failed to get user info", "cause error should contain network failure description")
}

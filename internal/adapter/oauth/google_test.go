package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testClientID     = "test-client-id"
	testClientSecret = "test-client-secret"
	testRedirectURI  = "http://localhost:8080/callback"
)

// Test helper to create a test adapter
func newTestAdapter(scopes []string) *GoogleOAuthAdapter {
	config := GoogleOAuthConfig{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		RedirectURI:  testRedirectURI,
		Scopes:       scopes,
	}
	return NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)
}

// TestNewGoogleOAuthAdapter_DefaultScopes tests that default scopes are applied when none provided.
func TestNewGoogleOAuthAdapter_DefaultScopes(t *testing.T) {
	tests := []struct {
		name   string
		scopes []string
		want   []string
	}{
		{
			name:   "nil scopes defaults to openid, email, profile",
			scopes: nil,
			want:   []string{"openid", "email", "profile"},
		},
		{
			name:   "empty scopes is kept as-is (no default applied to empty slice)",
			scopes: []string{},
			want:   []string{}, // Empty slice is different from nil
		},
		{
			name:   "custom scopes are preserved",
			scopes: []string{"custom1", "custom2"},
			want:   []string{"custom1", "custom2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GoogleOAuthConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				Scopes:       tt.scopes,
			}
			adapter := NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)

			got := adapter.oauth2Config.Scopes
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGoogleOAuthAdapter_GetAuthorizationURL tests the GetAuthorizationURL method.
func TestGoogleOAuthAdapter_GetAuthorizationURL(t *testing.T) {
	tests := []struct {
		name        string
		scopes      []string
		redirectURI string
		state       string
		wantParams  map[string]string
	}{
		{
			name:        "Happy Path - valid redirect URI and state",
			scopes:      []string{"openid", "email", "profile"},
			redirectURI: testRedirectURI,
			state:       "random-state-123",
			wantParams: map[string]string{
				"client_id":     testClientID,
				"redirect_uri":  testRedirectURI,
				"response_type": "code",
				"scope":         "openid email profile",
				"state":         "random-state-123",
				"access_type":   "offline",
				"prompt":        "consent",
			},
		},
		{
			name:        "Empty Redirect URI",
			scopes:      []string{"openid", "email"},
			redirectURI: "",
			state:       "state-456",
			wantParams: map[string]string{
				"client_id":     testClientID,
				"redirect_uri":  "",
				"response_type": "code",
				"scope":         "openid email",
				"state":         "state-456",
				"access_type":   "offline",
				"prompt":        "consent",
			},
		},
		{
			name:        "With State",
			scopes:      []string{"openid", "email", "profile"},
			redirectURI: testRedirectURI,
			state:       "my-custom-state",
			wantParams: map[string]string{
				"client_id":     testClientID,
				"redirect_uri":  testRedirectURI,
				"response_type": "code",
				"scope":         "openid email profile",
				"state":         "my-custom-state",
				"access_type":   "offline",
				"prompt":        "consent",
			},
		},
		{
			name:        "Default Scopes",
			scopes:      nil,
			redirectURI: testRedirectURI,
			state:       "state-with-default-scopes",
			wantParams: map[string]string{
				"client_id":     testClientID,
				"redirect_uri":  testRedirectURI,
				"response_type": "code",
				"scope":         "openid email profile",
				"state":         "state-with-default-scopes",
				"access_type":   "offline",
				"prompt":        "consent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := newTestAdapter(tt.scopes)
			ctx := context.Background()

			gotURL, err := adapter.GetAuthorizationURL(ctx, tt.redirectURI, tt.state)

			require.NoError(t, err)
			require.NotNil(t, gotURL)

			// Verify URL starts with Google's OAuth endpoint
			// Note: oauth2 library uses /o/oauth2/auth (v1), not /o/oauth2/v2/auth
			assert.True(t, strings.HasPrefix(gotURL, "https://accounts.google.com/o/oauth2/auth?"),
				"URL should start with Google's OAuth endpoint")

			// Parse and verify query parameters
			parsedURL, err := url.Parse(gotURL)
			require.NoError(t, err)

			queryParams := parsedURL.Query()
			for key, wantValue := range tt.wantParams {
				gotValue := queryParams.Get(key)
				assert.Equal(t, wantValue, gotValue, "parameter %s should match", key)
			}
		})
	}
}

// TestGoogleOAuthAdapter_ExchangeCode_NetworkCall tests that ExchangeCode makes proper HTTP requests.
// Note: This test makes actual HTTP calls to Google's OAuth endpoint which will fail
// with test credentials. It verifies the method structure and error handling.
func TestGoogleOAuthAdapter_ExchangeCode_NetworkCall(t *testing.T) {
	adapter := newTestAdapter([]string{"openid", "email", "profile"})
	ctx := context.Background()

	// This will make an actual HTTP call to Google's token endpoint
	// It will fail because we're using test credentials
	gotTokens, err := adapter.ExchangeCode(ctx, "test-code", testRedirectURI)

	// Should return error due to invalid credentials
	require.Error(t, err)
	assert.Nil(t, gotTokens)
}

// TestGoogleOAuthAdapter_GetUserInfo_NetworkCall tests that GetUserInfo makes proper HTTP requests.
// Note: This test makes actual HTTP calls to Google's OAuth endpoint which will fail
// with test tokens. It verifies the method structure and error handling.
func TestGoogleOAuthAdapter_GetUserInfo_NetworkCall(t *testing.T) {
	adapter := newTestAdapter([]string{"openid", "email", "profile"})
	ctx := context.Background()

	// This will make an actual HTTP call to Google's userinfo endpoint
	// It will fail because we're using an invalid token
	gotUserInfo, err := adapter.GetUserInfo(ctx, "invalid-test-token")

	// Should return error due to invalid token
	require.Error(t, err)
	assert.Contains(t, err.Error(), "userinfo error")
	assert.Nil(t, gotUserInfo)
}

// Test helper structs
type tokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	IDToken          string `json:"id_token"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type userInfoResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// TestGoogleOAuthAdapter_ExchangeCode_Integration tests HTTP request format with a mock server.
func TestGoogleOAuthAdapter_ExchangeCode_Integration(t *testing.T) {
	t.Run("Happy Path - valid token response", func(t *testing.T) {
		// Create a test server that mocks Google's token endpoint
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify it's a POST request
			assert.Equal(t, http.MethodPost, r.Method)

			// Verify content type
			assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

			// Write response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&tokenResponse{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				ExpiresIn:    3600,
				IDToken:      "test-id-token",
				TokenType:    "Bearer",
			})
		}))
		defer server.Close()

		// Create adapter with custom token URL that points to our mock server
		config := GoogleOAuthConfig{
			ClientID:     testClientID,
			ClientSecret: testClientSecret,
			RedirectURI:  testRedirectURI,
			TokenURL:     server.URL,
		}
		adapter := NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)

		// Test ExchangeCode
		ctx := context.Background()
		gotTokens, err := adapter.ExchangeCode(ctx, "valid-code", testRedirectURI)

		require.NoError(t, err)
		assert.Equal(t, "test-access-token", gotTokens.AccessToken)
		assert.Equal(t, "test-refresh-token", gotTokens.RefreshToken)
		assert.GreaterOrEqual(t, gotTokens.ExpiresIn, 3590) // Allow small timing diff
		assert.Equal(t, "test-id-token", gotTokens.IDToken)
		assert.Equal(t, "Bearer", gotTokens.TokenType)
	})

	t.Run("OAuth Error - invalid_grant", func(t *testing.T) {
		// Create a test server that mocks Google's token endpoint with an error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify it's a POST request
			assert.Equal(t, http.MethodPost, r.Method)

			// Write error response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&tokenResponse{
				Error:            "invalid_grant",
				ErrorDescription: "Bad Request",
			})
		}))
		defer server.Close()

		// Create adapter with custom token URL that points to our mock server
		config := GoogleOAuthConfig{
			ClientID:     testClientID,
			ClientSecret: testClientSecret,
			RedirectURI:  testRedirectURI,
			TokenURL:     server.URL,
		}
		adapter := NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)

		// Test ExchangeCode
		ctx := context.Background()
		gotTokens, err := adapter.ExchangeCode(ctx, "invalid-code", testRedirectURI)

		require.Error(t, err)
		assert.Nil(t, gotTokens)
		assert.Contains(t, err.Error(), "failed to exchange code")
	})
}

// TestGoogleOAuthAdapter_GetUserInfo_Integration tests HTTP request format with a mock server.
func TestGoogleOAuthAdapter_GetUserInfo_Integration(t *testing.T) {
	tests := []struct {
		name           string
		accessToken    string
		mockResponse   *userInfoResponse
		mockStatusCode int
		wantUserInfo   *model.OAuthUserInfo
		wantErr        bool
	}{
		{
			name:        "Happy Path",
			accessToken: "valid-token",
			mockResponse: &userInfoResponse{
				ID:            "google-123",
				Email:         "test@example.com",
				VerifiedEmail: true,
				GivenName:     "Test",
				FamilyName:    "User",
				Name:          "Test User",
				Picture:       "http://example.com/pic.jpg",
				Locale:        "en",
			},
			mockStatusCode: http.StatusOK,
			wantUserInfo: &model.OAuthUserInfo{
				ProviderID:    "google-123",
				Email:         "test@example.com",
				EmailVerified: true,
				FirstName:     "Test",
				LastName:      "User",
				FullName:      "Test User",
				Picture:       "http://example.com/pic.jpg",
				Locale:        "en",
			},
			wantErr: false,
		},
		{
			name:           "Unauthorized",
			accessToken:    "invalid-token",
			mockResponse:   &userInfoResponse{},
			mockStatusCode: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server that mocks Google's userinfo endpoint
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify it's a GET request
				assert.Equal(t, http.MethodGet, r.Method)

				// Verify Authorization header
				expectedAuth := "Bearer " + tt.accessToken
				assert.Equal(t, expectedAuth, r.Header.Get("Authorization"))

				// Write response
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockResponse != nil {
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Create adapter with custom UserInfo URL that points to our mock server
			config := GoogleOAuthConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				RedirectURI:  testRedirectURI,
				UserInfoURL:  server.URL,
			}
			adapter := NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)

			// Test GetUserInfo
			ctx := context.Background()
			gotUserInfo, err := adapter.GetUserInfo(ctx, tt.accessToken)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, gotUserInfo)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantUserInfo.ProviderID, gotUserInfo.ProviderID)
				assert.Equal(t, tt.wantUserInfo.Email, gotUserInfo.Email)
				assert.Equal(t, tt.wantUserInfo.EmailVerified, gotUserInfo.EmailVerified)
			}
		})
	}
}

// TestGoogleOAuthAdapter_CustomURLInjection tests that custom URLs can be injected for testing.
func TestGoogleOAuthAdapter_CustomURLInjection(t *testing.T) {
	tests := []struct {
		name          string
		authURL       string
		tokenURL      string
		userInfoURL   string
		expectedAuth  string
		expectedToken string
	}{
		{
			name:          "Empty URLs use Google defaults",
			authURL:       "",
			tokenURL:      "",
			userInfoURL:   "",
			expectedAuth:  "https://accounts.google.com/o/oauth2/auth",
			expectedToken: "https://oauth2.googleapis.com/token",
		},
		{
			name:          "Custom URLs override defaults",
			authURL:       "http://localhost:8080/oauth/auth",
			tokenURL:      "http://localhost:8080/oauth/token",
			userInfoURL:   "http://localhost:8080/oauth/userinfo",
			expectedAuth:  "http://localhost:8080/oauth/auth",
			expectedToken: "http://localhost:8080/oauth/token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GoogleOAuthConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				RedirectURI:  testRedirectURI,
				AuthURL:      tt.authURL,
				TokenURL:     tt.tokenURL,
				UserInfoURL:  tt.userInfoURL,
			}
			adapter := NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)

			// Test GetAuthorizationURL uses correct auth URL
			ctx := context.Background()
			authURL, err := adapter.GetAuthorizationURL(ctx, testRedirectURI, "test-state")
			require.NoError(t, err)
			require.True(t, strings.HasPrefix(authURL, tt.expectedAuth+"?"),
				"URL should start with %s", tt.expectedAuth)

			// Test oauth2Config.Endpoint.TokenURL returns correct token URL
			assert.Equal(t, tt.expectedToken, adapter.oauth2Config.Endpoint.TokenURL)
		})
	}
}

// TestGoogleOAuthAdapter_HTTPClientInjection tests that a custom HTTP client can be injected.
func TestGoogleOAuthAdapter_HTTPClientInjection(t *testing.T) {
	tests := []struct {
		name      string
		httpClient *http.Client
	}{
		{
			name:      "Default HTTP client when nil",
			httpClient: nil,
		},
		{
			name:      "Custom HTTP client",
			httpClient: &http.Client{Timeout: 30 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GoogleOAuthConfig{
				ClientID:     testClientID,
				ClientSecret: testClientSecret,
				RedirectURI:  testRedirectURI,
				HTTPClient:   tt.httpClient,
			}
			adapter := NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)

			// Verify getHTTPClient returns correct client
			if tt.httpClient != nil {
				require.Equal(t, tt.httpClient, adapter.getHTTPClient())
			} else {
				// Should create a default client
				client := adapter.getHTTPClient()
				require.NotNil(t, client)
				require.Equal(t, 10*time.Second, client.Timeout)
			}
		})
	}
}

// TestGoogleOAuthAdapter_ExchangeCode_WithMockServer tests ExchangeCode with a proper mock server.
// Note: This test is now covered by TestGoogleOAuthAdapter_ExchangeCode_Integration
// which provides better coverage of both success and error cases.
// Keeping this test for explicit verification of mock server behavior.
func TestGoogleOAuthAdapter_ExchangeCode_WithMockServer(t *testing.T) {
	t.Run("OAuth Error - invalid_grant", func(t *testing.T) {
		// Create a mock server that returns an error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(&tokenResponse{
				Error:            "invalid_grant",
				ErrorDescription: "The authorization code is invalid",
			})
		}))
		defer server.Close()

		// Create adapter with custom token URL
		config := GoogleOAuthConfig{
			ClientID:     testClientID,
			ClientSecret: testClientSecret,
			RedirectURI:  testRedirectURI,
			TokenURL:     server.URL,
		}
		adapter := NewGoogleOAuthAdapter(config).(*GoogleOAuthAdapter)

		// Test ExchangeCode
		ctx := context.Background()
		tokens, err := adapter.ExchangeCode(ctx, "invalid-code", testRedirectURI)

		require.Error(t, err)
		assert.Nil(t, tokens)
		assert.Contains(t, err.Error(), "failed to exchange code")
	})
}

// TestGoogleOAuthAdapter_GetAuthorizationURL_oauth2Options tests that oauth2 options are applied.
func TestGoogleOAuthAdapter_GetAuthorizationURL_oauth2Options(t *testing.T) {
	adapter := newTestAdapter([]string{"openid", "email", "profile"})
	ctx := context.Background()

	// Get authorization URL
	authURL, err := adapter.GetAuthorizationURL(ctx, testRedirectURI, "test-state")
	require.NoError(t, err)

	// Parse URL to verify parameters
	parsedURL, err := url.Parse(authURL)
	require.NoError(t, err)

	queryParams := parsedURL.Query()

	// Verify oauth2 library's options are applied
	assert.Equal(t, "offline", queryParams.Get("access_type"),
		"oauth2.AccessTypeOffline should add access_type=offline")
	assert.Equal(t, "consent", queryParams.Get("prompt"),
		"oauth2.ApprovalForce should add prompt=consent")
}

// TestGetAuthorizationURLWithPKCE tests full PKCE flow with verifier storage
func TestGetAuthorizationURLWithPKCE(t *testing.T) {
	t.Run("returns URL with code challenge", func(t *testing.T) {
		s := miniredis.RunT(t)
		defer s.Close()

		config := oauth.GoogleOAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURI:  "http://localhost:8080/callback",
		}
		adapter := oauth.NewGoogleOAuthAdapter(config, s.Client())

		ctx := context.Background()
		state := "test-state-abc123"
		redirectURI := "http://localhost:8080/callback"

		authURL, err := adapter.GetAuthorizationURL(ctx, redirectURI, state)
		if err != nil {
			t.Fatalf("GetAuthorizationURL() error = %v", err)
		}

		// Verify URL contains PKCE parameters
		if !strings.Contains(authURL, "code_challenge=") {
			t.Error("GetAuthorizationURL() missing code_challenge parameter")
		}
		if !strings.Contains(authURL, "code_challenge_method=S256") {
			t.Error("GetAuthorizationURL() missing code_challenge_method parameter")
		}
		if !strings.Contains(authURL, "state="+state) {
			t.Error("GetAuthorizationURL() missing state parameter")
		}

		// Verify verifier was stored in Redis
		verifier, err := adapter.GetVerifier(ctx, state)
		if err != nil {
			t.Errorf("verifier not stored in Redis: %v", err)
		}
		if verifier == "" {
			t.Error("stored verifier is empty")
		}
	})
}

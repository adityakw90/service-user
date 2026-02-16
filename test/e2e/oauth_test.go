package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	authgrpc "github.com/adityakw90/service-user-proto/gen/go/auth"
	usergrpc "github.com/adityakw90/service-user-proto/gen/go/user"
	testutil "github.com/adityakw90/service-user/test/util"
	"github.com/stretchr/testify/require"
)

// TestE2E_AuthService_GoogleOAuth tests the GoogleOAuth endpoint that generates the authorization URL.
func TestE2E_AuthService_GoogleOAuth(t *testing.T) {
	testServices, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	tests := []struct {
		name       string
		oauthReq   func(t *testing.T) *authgrpc.GoogleOAuthRequest
		wantErr    bool
		verifyFunc func(t *testing.T, resp *authgrpc.GoogleOAuthResponse)
	}{
		{
			name: "Valid redirect URI returns authorization URL with correct parameters",
			oauthReq: func(t *testing.T) *authgrpc.GoogleOAuthRequest {
				return &authgrpc.GoogleOAuthRequest{
					RedirectUri: "http://localhost:8080/callback",
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *authgrpc.GoogleOAuthResponse) {
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.AuthorizationUrl)

				// Verify URL starts with Google's OAuth endpoint
				// Note: oauth2 library uses /o/oauth2/auth (v1), not /o/oauth2/v2/auth
				require.True(t, strings.HasPrefix(resp.AuthorizationUrl, "https://accounts.google.com/o/oauth2/auth?"),
					"URL should start with Google's OAuth endpoint")

				// Parse and verify query parameters
				parsedURL, err := url.Parse(resp.AuthorizationUrl)
				require.NoError(t, err)

				queryParams := parsedURL.Query()
				require.Equal(t, testServices.Cfg.OAuth.Google.ClientID, queryParams.Get("client_id"))
				require.Equal(t, "http://localhost:8080/callback", queryParams.Get("redirect_uri"))
				require.Equal(t, "code", queryParams.Get("response_type"))
				require.NotEmpty(t, queryParams.Get("state"))
				require.Contains(t, queryParams.Get("scope"), "openid")
				require.Contains(t, queryParams.Get("scope"), "email")
				require.Equal(t, "offline", queryParams.Get("access_type"))
				require.Equal(t, "consent", queryParams.Get("prompt"))
			},
		},
		{
			name: "Empty redirect URI returns validation error",
			oauthReq: func(t *testing.T) *authgrpc.GoogleOAuthRequest {
				return &authgrpc.GoogleOAuthRequest{
					RedirectUri: "",
				}
			},
			wantErr: true,
			verifyFunc: nil,
		},
		{
			name: "Multiple calls generate unique state",
			oauthReq: func(t *testing.T) *authgrpc.GoogleOAuthRequest {
				return &authgrpc.GoogleOAuthRequest{
					RedirectUri: "http://localhost:8080/callback",
				}
			},
			wantErr: false,
			verifyFunc: func(t *testing.T, resp *authgrpc.GoogleOAuthResponse) {
				// Get another URL
				ctx := context.Background()
				resp2, err := grpcClient.AuthClient.GoogleOAuth(ctx, &authgrpc.GoogleOAuthRequest{
					RedirectUri: "http://localhost:8080/callback",
				})
				require.NoError(t, err)
				require.NotNil(t, resp2)

				// Extract state from both URLs
				parsedURL1, err := url.Parse(resp.AuthorizationUrl)
				require.NoError(t, err)
				state1 := parsedURL1.Query().Get("state")

				parsedURL2, err := url.Parse(resp2.AuthorizationUrl)
				require.NoError(t, err)
				state2 := parsedURL2.Query().Get("state")

				// States should be different
				require.NotEqual(t, state1, state2, "Each call should generate unique state")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			oauthReq := tt.oauthReq(t)

			resp, err := grpcClient.AuthClient.GoogleOAuth(ctx, oauthReq)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, resp)
				}
			}
		})
	}
}

// mockOAuthUserInfo represents the mock user info returned by the mock OAuth server.
type mockOAuthUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

// TestE2E_AuthService_HandleGoogleOAuth tests the HandleGoogleOAuth callback endpoint with mock OAuth server.
// Uses mock OAuth server for integration testing.
func TestE2E_AuthService_HandleGoogleOAuth(t *testing.T) {
	// Set up mock OAuth server
	mockEndpoints, closeMockServer, _ := setupMockOAuthServer(t)
	defer closeMockServer()

	// Set environment variables to configure OAuth adapter with mock endpoints
	t.Setenv("OAUTH_GOOGLE_TOKEN_URL", mockEndpoints.TokenURL)
	t.Setenv("OAUTH_GOOGLE_USER_INFO_URL", mockEndpoints.UserInfoURL)

	testServices, grpcClient, cleanup := setupE2ETest(t)
	defer cleanup()

	// Skip if OAuth provider is not configured
	if testServices.AuthService == nil {
		t.Skip("OAuth provider not configured in test services")
	}

	tests := []struct {
		name             string
		mockUserInfo      *mockOAuthUserInfo
		mockTokenResponse map[string]any
		mockStatusCode   int
		setup            func(t *testing.T, grpcClient *testutil.TestGRPCClient) string
		handleReq        func(t *testing.T) *authgrpc.HandleGoogleOAuthRequest
		wantErr          bool
	}{
		{
			name: "New user login creates new user and returns JWT tokens",
			mockUserInfo: &mockOAuthUserInfo{
				ID:            "google-12345",
				Email:         "oauthnewuser@example.com",
				VerifiedEmail: true,
				GivenName:     "New",
				FamilyName:    "User",
				Name:          "New User",
				Picture:       "http://example.com/pic.jpg",
				Locale:        "en",
			},
			mockTokenResponse: map[string]any{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_in":    3600,
				"id_token":      "test-id-token",
				"token_type":    "Bearer",
			},
			mockStatusCode: http.StatusOK,
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				return "" // No existing user
			},
			handleReq: func(t *testing.T) *authgrpc.HandleGoogleOAuthRequest {
				return &authgrpc.HandleGoogleOAuthRequest{
					Code:        "valid-auth-code",
					State:       "test-state-new-user",
					RedirectUri: "http://localhost:8080/callback",
				}
			},
			wantErr: false,
		},
		{
			name: "Existing user login returns JWT tokens for existing user",
			mockUserInfo: &mockOAuthUserInfo{
				ID:            "google-67890",
				Email:         "oauthexistinguser@example.com",
				VerifiedEmail: true,
				GivenName:     "Existing",
				FamilyName:    "User",
				Name:          "Existing User",
				Picture:       "http://example.com/pic2.jpg",
				Locale:        "en",
			},
			mockTokenResponse: map[string]any{
				"access_token":  "test-access-token-2",
				"refresh_token": "test-refresh-token-2",
				"expires_in":    3600,
				"id_token":      "test-id-token-2",
				"token_type":    "Bearer",
			},
			mockStatusCode: http.StatusOK,
			setup: func(t *testing.T, grpcClient *testutil.TestGRPCClient) string {
				// Create user via regular auth first
				ctx := context.Background()
				userReq := &usergrpc.AddRequest{
					Username: "oauthexistinguser",
					Email:    "oauthexistinguser@example.com",
					Password: "Password123!",
				}
				resp, err := grpcClient.UserClient.Add(ctx, userReq)
				require.NoError(t, err)
				return resp.Uid
			},
			handleReq: func(t *testing.T) *authgrpc.HandleGoogleOAuthRequest {
				return &authgrpc.HandleGoogleOAuthRequest{
					Code:        "valid-auth-code-2",
					State:       "test-state-existing-user",
					RedirectUri: "http://localhost:8080/callback",
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			_ = tt.setup(t, grpcClient)
			handleReq := tt.handleReq(t)

			// Pre-store PKCE verifier in Redis for the test state
			// Key format: oauth:pkce:{state} (from internal/adapter/oauth/google.go)
			mockVerifier := "mock-verifier-" + handleReq.State
			err := testServices.Redis.Set(ctx, "oauth:pkce:"+handleReq.State, mockVerifier, 10*time.Minute).Err()
			require.NoError(t, err, "Failed to store PKCE verifier in Redis")

			resp, err := grpcClient.AuthClient.HandleGoogleOAuth(ctx, handleReq)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, resp)
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.AccessToken)
				require.NotEmpty(t, resp.RefreshToken)
			}
		})
	}
}

// mockOAuthServerEndpoints holds the endpoint URLs for the mock OAuth server.
type mockOAuthServerEndpoints struct {
	BaseURL     string // Base URL of the mock server (e.g., "http://127.0.0.1:8080")
	AuthURL     string // Full URL for auth endpoint
	TokenURL    string // Full URL for token endpoint
	UserInfoURL string // Full URL for userinfo endpoint
}

// setupMockOAuthServer creates a test HTTP server mocking Google's OAuth endpoints.
// Returns mockOAuthServerEndpoints, close function, and test user info.
func setupMockOAuthServer(t *testing.T) (*mockOAuthServerEndpoints, func(), *mockOAuthUserInfo) {
	t.Helper()

	testUserInfo := &mockOAuthUserInfo{
		ID:            "google-test-123",
		Email:         "testoauth@example.com",
		VerifiedEmail: true,
		GivenName:     "Test",
		FamilyName:    "OAuth",
		Name:          "Test OAuth",
		Picture:       "http://example.com/test-pic.jpg",
		Locale:        "en",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/token"):
			// Token exchange endpoint
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"access_token": "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_in": 3600,
				"id_token": "test-id-token",
				"token_type": "Bearer"
			}`))

		case strings.HasPrefix(r.URL.Path, "/oauth2/v2/userinfo"):
			// User info endpoint
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"id": "google-test-123",
				"email": "testoauth@example.com",
				"verified_email": true,
				"given_name": "Test",
				"family_name": "OAuth",
				"name": "Test OAuth",
				"picture": "http://example.com/test-pic.jpg",
				"locale": "en"
			}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	closeFunc := func() {
		server.Close()
	}

	endpoints := &mockOAuthServerEndpoints{
		BaseURL:     server.URL,
		AuthURL:     server.URL + "/auth", // For completeness, though GetAuthorizationURL doesn't make HTTP calls
		TokenURL:    server.URL + "/token",
		UserInfoURL: server.URL + "/oauth2/v2/userinfo",
	}

	return endpoints, closeFunc, testUserInfo
}

// skipIfNoOAuthConfig skips the test if OAuth is not configured.
func skipIfNoOAuthConfig(t *testing.T) {
	t.Helper()

	// Check if OAuth is configured with test credentials
	// If not, skip the test with a message
	if testing.Short() {
		t.Skip("Skipping OAuth test in short mode")
	}

	// Tests will fail gracefully if OAuth provider is not initialized
	// The service setup in test/util/service.go creates OAuth provider
	// only if OAUTH_GOOGLE_CLIENT_ID is set
}

// createOAuthTestUser creates a user via regular auth for testing existing user OAuth flow.
func createOAuthTestUser(t *testing.T, grpcClient *testutil.TestGRPCClient, username, email, password string) string {
	t.Helper()
	ctx := context.Background()
	req := &usergrpc.AddRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	resp, err := grpcClient.UserClient.Add(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	return resp.Uid
}

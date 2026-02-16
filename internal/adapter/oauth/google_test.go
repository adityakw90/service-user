package oauth

import (
	"context"
	"net/url"
	"strings"
	"testing"

	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				// assert teh code challenge and code challenge method are present
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

package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	portOAuth "github.com/adityakw90/service-user/internal/core/port/oauth"
	"github.com/redis/go-redis/v9"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Default Google OAuth UserInfo URL.
const (
	defaultGoogleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// GoogleOAuthConfig holds Google OAuth configuration.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	// Optional URL overrides for testing/custom endpoints
	AuthURL     string // If empty, defaults to Google's auth URL
	TokenURL    string // If empty, defaults to Google's token URL
	UserInfoURL string // If empty, defaults to Google's userinfo URL
	// Optional HTTP client injection for testing
	HTTPClient *http.Client // If nil, creates default client
}

// GoogleOAuthAdapter implements OAuthProvider for Google using golang.org/x/oauth2.
type GoogleOAuthAdapter struct {
	oauth2Config *oauth2.Config
	userInfoURL  string
	httpClient   *http.Client
	redis        redis.Cmdable
}

// NewGoogleOAuthAdapter creates a new Google OAuth adapter.
func NewGoogleOAuthAdapter(config GoogleOAuthConfig, redis redis.Cmdable) portOAuth.OAuthProvider {
	// Default scopes if none provided
	scopes := config.Scopes
	if scopes == nil {
		scopes = []string{
			"openid",
			"email",
			"profile",
		}
	}

	// Build oauth2 config with Google endpoints
	cfg := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURI,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}

	// Allow custom endpoint override for testing
	if config.AuthURL != "" || config.TokenURL != "" {
		authURL := google.Endpoint.AuthURL
		tokenURL := google.Endpoint.TokenURL

		if config.AuthURL != "" {
			authURL = config.AuthURL
		}
		if config.TokenURL != "" {
			tokenURL = config.TokenURL
		}

		cfg.Endpoint = oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		}
	}

	userInfoURL := config.UserInfoURL
	if userInfoURL == "" {
		userInfoURL = defaultGoogleUserInfoURL
	}

	return &GoogleOAuthAdapter{
		oauth2Config: cfg,
		userInfoURL:  userInfoURL,
		httpClient:   config.HTTPClient,
		redis:        redis,
	}
}

func (g *GoogleOAuthAdapter) getHTTPClient() *http.Client {
	if g.httpClient != nil {
		return g.httpClient
	}
	g.httpClient = &http.Client{Timeout: 10 * time.Second}
	return g.httpClient
}

// GetAuthorizationURL returns the OAuth authorization URL with state using oauth2 library.
func (g *GoogleOAuthAdapter) GetAuthorizationURL(ctx context.Context, redirectURI, state string) (string, error) {
	// Create a temporary config with the redirectURI override
	cfg := *g.oauth2Config // Copy the config
	cfg.RedirectURL = redirectURI

	// Use oauth2 library to generate auth URL with access_type=offline and prompt=consent
	// This ensures we get a refresh_token
	authURL := cfg.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
	)
	return authURL, nil
}

// ExchangeCode exchanges authorization code for tokens using oauth2 library.
func (g *GoogleOAuthAdapter) ExchangeCode(ctx context.Context, code, redirectURI string) (*model.OAuthTokens, error) {
	// Inject HTTP client into context if custom client is set
	if g.httpClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, g.httpClient)
	}

	// Use oauth2 library to exchange code
	token, err := g.oauth2Config.Exchange(ctx, code, oauth2.SetAuthURLParam("redirect_uri", redirectURI))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Calculate expires_in
	expiresIn := 0
	if !token.Expiry.IsZero() {
		expiresIn = int(time.Until(token.Expiry).Seconds())
	}

	// Extract id_token from extra fields
	idToken := ""
	if token.Extra("id_token") != nil {
		if idTokenStr, ok := token.Extra("id_token").(string); ok {
			idToken = idTokenStr
		}
	}

	return &model.OAuthTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    expiresIn,
		IDToken:      idToken,
		TokenType:    token.TokenType,
	}, nil
}

// GetUserInfo retrieves user information from Google.
// Note: The UserInfo endpoint is not part of the oauth2 library, so we keep the custom implementation.
func (g *GoogleOAuthAdapter) GetUserInfo(ctx context.Context, accessToken string) (*model.OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", g.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := g.getHTTPClient()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo error: %s", string(body))
	}

	var userResp struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Locale        string `json:"locale"`
	}

	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &model.OAuthUserInfo{
		ProviderID:    userResp.ID,
		Email:         userResp.Email,
		EmailVerified: userResp.VerifiedEmail,
		FirstName:     userResp.GivenName,
		LastName:      userResp.FamilyName,
		FullName:      userResp.Name,
		Picture:       userResp.Picture,
		Locale:        userResp.Locale,
	}, nil
}

package security

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/port"
)

// GoogleOAuthConfig holds Google OAuth configuration.
type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
}

// GoogleOAuthAdapter implements OAuthProvider for Google.
type GoogleOAuthAdapter struct {
	config GoogleOAuthConfig
}

// NewGoogleOAuthAdapter creates a new Google OAuth adapter.
func NewGoogleOAuthAdapter(config GoogleOAuthConfig) port.OAuthProvider {
	if config.Scopes == nil {
		config.Scopes = []string{
			"openid",
			"email",
			"profile",
		}
	}
	return &GoogleOAuthAdapter{config: config}
}

// GetAuthorizationURL returns the OAuth authorization URL with state.
func (g *GoogleOAuthAdapter) GetAuthorizationURL(ctx context.Context, redirectURI, state string) (string, error) {
	params := url.Values{
		"client_id":     {g.config.ClientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"scope":         {strings.Join(g.config.Scopes, " ")},
		"state":         {state},
		"access_type":   {"offline"},
		"prompt":        {"consent"},
	}

	return "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode(), nil
}

// ExchangeCode exchanges authorization code for tokens.
func (g *GoogleOAuthAdapter) ExchangeCode(ctx context.Context, code, redirectURI string) (*model.OAuthTokens, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {g.config.ClientID},
		"client_secret": {g.config.ClientSecret},
		"redirect_uri":  {redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oauth error: %s", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		IDToken      string `json:"id_token"`
		TokenType    string `json:"token_type"`
	}

	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &model.OAuthTokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		IDToken:      tokenResp.IDToken,
		TokenType:    tokenResp.TokenType,
	}, nil
}

// GetUserInfo retrieves user information from Google.
func (g *GoogleOAuthAdapter) GetUserInfo(ctx context.Context, accessToken string) (*model.OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
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

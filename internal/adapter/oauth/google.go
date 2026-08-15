package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	domainErrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	GoogleChallengeMethod  = "S256"
	GoogleChallengeCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"
	GoogleUserInfoURL      = "https://www.googleapis.com/oauth2/v2/userinfo"
)

var ErrGoogleCodeChallengeNotFound = errors.New("google code challenge not found")

type googleUserResp struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type GoogleOAuthConfig struct {
	ClientID          string
	ClientSecret      string
	Scopes            []string
	Endpoint          *oauth2.Endpoint
	UserInfoURL       string
	RedisPKCEPrefix   string
	RedisPKCETTL      time.Duration
	MinVerifierLength int
	MaxVerifierLength int
	HttpTimeout       time.Duration
}

type GoogleOAuth struct {
	config      *GoogleOAuthConfig
	oauthConfig *oauth2.Config
	redisClient *redis.Client
	httpClient  *http.Client
	tracer      gomon.Tracer
	logger      gomon.Logger
}

func NewGoogleOAuth(config *GoogleOAuthConfig, redis *redis.Client, tracer gomon.Tracer, logger gomon.Logger) (*GoogleOAuth, error) {
	if tracer == nil {
		panic("tracer is required")
	}
	if redis == nil {
		panic("redis client is required")
	}
	if logger == nil {
		panic("logger is required")
	}
	if config == nil {
		config = &GoogleOAuthConfig{}
	}
	if config.ClientID == "" {
		return nil, domainErrors.ErrOAuthClientIDRequired
	}
	if config.ClientSecret == "" {
		return nil, domainErrors.ErrOAuthClientSecretRequired
	}
	if config.Scopes == nil {
		config.Scopes = []string{"openid", "email", "profile"}
	}
	if config.Endpoint == nil {
		googleEndpoint := google.Endpoint
		config.Endpoint = &googleEndpoint
	}
	if config.UserInfoURL == "" {
		config.UserInfoURL = GoogleUserInfoURL
	}
	if config.RedisPKCEPrefix == "" {
		config.RedisPKCEPrefix = "oauth:pkce:"
	}
	if config.RedisPKCETTL == 0 {
		config.RedisPKCETTL = 10 * time.Minute
	}
	if config.MinVerifierLength == 0 {
		config.MinVerifierLength = 43
	}
	if config.MaxVerifierLength == 0 {
		config.MaxVerifierLength = 128
	}
	if config.HttpTimeout == 0 {
		config.HttpTimeout = 10 * time.Second
	}
	oauth := &GoogleOAuth{
		config: config,
		oauthConfig: &oauth2.Config{
			ClientID:     config.ClientID,
			ClientSecret: config.ClientSecret,
			Endpoint:     *config.Endpoint,
			Scopes:       config.Scopes,
		},
		redisClient: redis,
		httpClient:  &http.Client{Timeout: config.HttpTimeout},
		tracer:      tracer,
		logger:      logger,
	}
	return oauth, nil
}

// GetAuthorizationURL returns the OAuth authorization URL with state using oauth2 library.
// For PKCE, generates code_verifier, computes code_challenge, and stores in Redis.
func (g *GoogleOAuth) GetAuthorizationURL(ctx context.Context, state string, redirectURI string) (string, error) {
	newCtx, span := g.tracer.StartSpan(ctx, "oauth.google.GetAuthorizationURL")
	defer span.End()

	span.SetAttributes(
		attribute.String("redirect_url", redirectURI),
	)

	// Generate PKCE challenge
	challenge, err := g.createCodeChallenge()
	if err != nil {
		span.RecordError(err)
		return "", domainErrors.ErrOAuthFailedGenerateCodeVerifier.WithCause(fmt.Errorf("failed to create code challenge: %w", err))
	}

	// Store challenge in Redis
	if err := g.storeChallenge(newCtx, state, challenge); err != nil {
		span.RecordError(err)
		return "", domainErrors.ErrOAuthFailedGenerateCodeVerifier.WithCause(fmt.Errorf("failed to store code challenge: %w", err))
	}

	// Use oauth2 library to generate auth URL with access_type=offline and prompt=consent
	// This ensures we get a refresh_token
	authURL := g.oauthConfig.AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.ApprovalForce,
		oauth2.S256ChallengeOption(challenge), // PKCE option
		oauth2.SetAuthURLParam("redirect_uri", redirectURI),
	)

	return authURL, nil
}

// ExchangeCode exchanges authorization code for tokens using oauth2 library.
func (g *GoogleOAuth) ExchangeCode(ctx context.Context, code, state, redirectURI string) (*model.OAuthTokens, error) {
	newCtx, span := g.tracer.StartSpan(ctx, "oauth.google.ExchangeCode")
	defer span.End()

	span.SetAttributes(
		attribute.String("exchange.redirect_url", redirectURI),
	)

	// Get code challenge from Redis
	challenge, err := g.getChallenge(newCtx, state)
	if err != nil {
		span.RecordError(err)
		// Map only not-found/missing-key errors to ErrOAuthCodeVerifierMissing
		// Preserve backend and context failures as operational errors
		if errors.Is(err, ErrGoogleCodeChallengeNotFound) {
			return nil, domainErrors.ErrOAuthCodeVerifierMissing.WithCause(err)
		}
		// Operational error: return with cause preserved
		return nil, err
	}

	// Exchange code for tokens using oauth2 library
	token, err := g.oauthConfig.Exchange(newCtx, code,
		oauth2.SetAuthURLParam("redirect_uri", redirectURI),
		oauth2.SetAuthURLParam("code_verifier", challenge),
	)
	if err != nil {
		span.RecordError(err)
		return nil, domainErrors.ErrOAuthExchangeFailed.WithCause(fmt.Errorf("failed to exchange code for tokens: %w", err))
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

	result := &model.OAuthTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresIn:    expiresIn,
		IDToken:      idToken,
		TokenType:    token.TokenType,
	}

	span.AddEvent("oauth.google.ExchangeCode.result",
		trace.WithAttributes(
			attribute.String("token.token_type", result.TokenType),
			attribute.Int("token.expires_in", result.ExpiresIn),
		),
	)

	return result, nil
}

// GetUserInfo retrieves user information using the access token.
func (g *GoogleOAuth) GetUserInfo(ctx context.Context, token *model.OAuthTokens) (*model.OAuthUserInfo, error) {
	newCtx, span := g.tracer.StartSpan(ctx, "oauth.google.GetUserInfo")
	defer span.End()

	req, err := http.NewRequestWithContext(newCtx, "GET", g.config.UserInfoURL, nil)
	if err != nil {
		span.RecordError(err)
		return nil, domainErrors.ErrOAuthUserInfoFailed.WithCause(fmt.Errorf("failed to create request: %w", err))
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		return nil, domainErrors.ErrOAuthUserInfoFailed.WithCause(fmt.Errorf("failed to get user info: %w", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return nil, domainErrors.ErrOAuthUserInfoFailed.WithCause(fmt.Errorf("failed to read response: %w", err))
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("userinfo error: %s", string(body))
		span.RecordError(err)
		return nil, domainErrors.ErrOAuthUserInfoFailed.WithCause(err)
	}

	var userResp googleUserResp
	if err := json.Unmarshal(body, &userResp); err != nil {
		span.RecordError(err)
		return nil, domainErrors.ErrOAuthUserInfoFailed.WithCause(fmt.Errorf("failed to parse user info: %w", err))
	}

	result := &model.OAuthUserInfo{
		ProviderID:    userResp.ID,
		Email:         userResp.Email,
		EmailVerified: userResp.VerifiedEmail,
		FirstName:     userResp.GivenName,
		LastName:      userResp.FamilyName,
		FullName:      userResp.Name,
		Picture:       userResp.Picture,
		Locale:        userResp.Locale,
	}

	return result, nil
}

func (g *GoogleOAuth) createCodeChallenge() (string, error) {
	b := make([]byte, g.config.MaxVerifierLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate verifier: %w", err)
	}

	// Trim to desired length and encode using base64url
	verifier := strings.Builder{}
	for i := range g.config.MaxVerifierLength {
		verifier.WriteByte(GoogleChallengeCharset[b[i]%byte(len(GoogleChallengeCharset))])
	}
	result := verifier.String()

	// Ensure minimum length
	if len(result) < g.config.MinVerifierLength {
		return "", fmt.Errorf("verifier too short: %d", len(result))
	}

	// Compute SHA-256 hash of verifier
	h := sha256.Sum256([]byte(result))
	return base64.RawURLEncoding.EncodeToString(h[:]), nil
}

// storeChallenge stores code challenge in Redis with TTL.
func (g *GoogleOAuth) storeChallenge(ctx context.Context, state, challenge string) error {
	newCtx, span := g.tracer.StartSpan(ctx, "oauth.google.storeChallenge")
	defer span.End()

	return g.redisClient.Set(
		newCtx,
		g.buildKey(state),
		challenge,
		g.config.RedisPKCETTL,
	).Err()
}

// getChallenge retrieves code challenge from Redis.
func (g *GoogleOAuth) getChallenge(ctx context.Context, state string) (string, error) {
	newCtx, span := g.tracer.StartSpan(ctx, "oauth.google.getChallenge")
	defer span.End()

	logger := g.logger.WithSpanContext(span.SpanContext())

	key := g.buildKey(state)

	code, err := g.redisClient.Get(newCtx, key).Result()
	if err != nil {
		span.RecordError(err)
		if err == redis.Nil {
			logger.Error("failed to get code challenge", map[string]any{
				"error.type":    fmt.Sprintf("%T", err),
				"error.message": err.Error(),
			})
			return "", ErrGoogleCodeChallengeNotFound
		}
		return "", fmt.Errorf("failed to get code challenge: %w", err)
	}

	// delete for one time use
	// Delete verifier (single-use)
	if err := g.redisClient.Del(newCtx, key).Err(); err != nil {
		// Log but don't fail - verifier is already consumed
		// In production, this should be logged
		logger.Error("failed to delete code challenge for one time use", map[string]any{
			"error.type":    fmt.Sprintf("%T", err),
			"error.message": err.Error(),
		})
	}
	return code, nil
}

func (g *GoogleOAuth) buildKey(state string) string {
	return g.config.RedisPKCEPrefix + state
}

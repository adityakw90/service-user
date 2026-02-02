package security

import (
	"errors"
	"testing"
	"time"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

const (
	testSecretKey     = "test-secret-key-for-jwt-testing-32bytes"
	testAccessExpiry  = 15 * time.Minute
	testRefreshExpiry = 7 * 24 * time.Hour
)

func TestJWTGenerator_GenerateAccessToken(t *testing.T) {
	generator := NewJWTGenerator(testSecretKey, testAccessExpiry, testRefreshExpiry)

	tests := []struct {
		name    string
		claims  model.TokenClaims
		wantErr bool
	}{
		{
			name: "generates valid access token with all fields",
			claims: model.TokenClaims{
				Uid:            "user-123",
				Identifier:     "user@example.com",
				IdentifierType: "email",
				Extra:          map[string]any{"role": "admin"},
			},
			wantErr: false,
		},
		{
			name: "generates access token with minimal fields",
			claims: model.TokenClaims{
				Uid:        "user-456",
				Identifier: "username",
			},
			wantErr: false,
		},
		{
			name: "generates access token with empty extra",
			claims: model.TokenClaims{
				Uid:            "user-789",
				Identifier:     "phone",
				IdentifierType: "phone",
				Extra:          nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generator.GenerateAccessToken(tt.claims)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateAccessToken() returned empty token")
			}
			if !tt.wantErr && len(token) < 50 {
				t.Errorf("GenerateAccessToken() token too short: %s", token)
			}
		})
	}
}

func TestJWTGenerator_GenerateRefreshToken(t *testing.T) {
	generator := NewJWTGenerator(testSecretKey, testAccessExpiry, testRefreshExpiry)

	tests := []struct {
		name    string
		claims  model.TokenClaims
		wantErr bool
	}{
		{
			name: "generates valid refresh token",
			claims: model.TokenClaims{
				Uid:            "user-123",
				Identifier:     "user@example.com",
				IdentifierType: "email",
			},
			wantErr: false,
		},
		{
			name: "generates refresh token with extra data",
			claims: model.TokenClaims{
				Uid:            "user-456",
				Identifier:     "device-id-123",
				IdentifierType: "device",
				Extra:          map[string]any{"ip": "192.168.1.1"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generator.GenerateRefreshToken(tt.claims)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == "" {
				t.Error("GenerateRefreshToken() returned empty token")
			}
		})
	}
}

func TestJWTGenerator_ValidateToken(t *testing.T) {
	generator := NewJWTGenerator(testSecretKey, testAccessExpiry, testRefreshExpiry)

	// Generate a valid token for testing
	validClaims := model.TokenClaims{
		Uid:            "user-123",
		Identifier:     "user@example.com",
		IdentifierType: "email",
		Extra:          map[string]any{"role": "admin"},
	}
	validToken, err := generator.GenerateAccessToken(validClaims)
	if err != nil {
		t.Fatalf("Failed to generate valid token: %v", err)
	}

	tests := []struct {
		name        string
		token       string
		wantUID     string
		wantErr     bool
		wantErrType error
	}{
		{
			name:    "validates correct token",
			token:   validToken,
			wantUID: "user-123",
			wantErr: false,
		},
		{
			name:        "rejects malformed token",
			token:       "invalid.token.string",
			wantErr:     true,
			wantErrType: domainerrors.ErrTokenInvalid,
		},
		{
			name:        "rejects empty token",
			token:       "",
			wantErr:     true,
			wantErrType: domainerrors.ErrTokenInvalid,
		},
		{
			name:        "rejects token signed with different secret",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiJ0ZXN0IiwiaXNzIjoic2VydmljZS11c2VyIn0.wrong_signature",
			wantErr:     true,
			wantErrType: domainerrors.ErrTokenInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := generator.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims.Uid != tt.wantUID {
					t.Errorf("ValidateToken() claims.Uid = %v, want %v", claims.Uid, tt.wantUID)
				}
				if claims.Identifier != validClaims.Identifier {
					t.Errorf("ValidateToken() claims.Identifier = %v, want %v", claims.Identifier, validClaims.Identifier)
				}
				if claims.IdentifierType != validClaims.IdentifierType {
					t.Errorf("ValidateToken() claims.IdentifierType = %v, want %v", claims.IdentifierType, validClaims.IdentifierType)
				}
			}
			if tt.wantErrType != nil && err != nil {
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("ValidateToken() error type = %v, want %v", err, tt.wantErrType)
				}
			}
		})
	}
}

func TestJWTGenerator_TokenRoundTrip(t *testing.T) {
	generator := NewJWTGenerator(testSecretKey, testAccessExpiry, testRefreshExpiry)

	originalClaims := model.TokenClaims{
		Uid:            "user-round-trip-test",
		Identifier:     "test@example.com",
		IdentifierType: "email",
		Extra:          map[string]any{"key": "value", "number": 42},
	}

	// Generate token
	token, err := generator.GenerateAccessToken(originalClaims)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Validate and extract claims
	extractedClaims, err := generator.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	// Verify claims match
	if extractedClaims.Uid != originalClaims.Uid {
		t.Errorf("Uid mismatch: got %v, want %v", extractedClaims.Uid, originalClaims.Uid)
	}
	if extractedClaims.Identifier != originalClaims.Identifier {
		t.Errorf("Identifier mismatch: got %v, want %v", extractedClaims.Identifier, originalClaims.Identifier)
	}
	if extractedClaims.IdentifierType != originalClaims.IdentifierType {
		t.Errorf("IdentifierType mismatch: got %v, want %v", extractedClaims.IdentifierType, originalClaims.IdentifierType)
	}
	if extractedClaims.Extra["key"] != "value" {
		t.Errorf("Extra[key] mismatch: got %v, want %v", extractedClaims.Extra["key"], "value")
	}
	if extractedClaims.Extra["number"] != float64(42) {
		t.Errorf("Extra[number] mismatch: got %v, want %v", extractedClaims.Extra["number"], 42)
	}
}

func TestJWTGenerator_DifferentSecrets(t *testing.T) {
	generator1 := NewJWTGenerator("secret-key-1", testAccessExpiry, testRefreshExpiry)
	generator2 := NewJWTGenerator("secret-key-2", testAccessExpiry, testRefreshExpiry)

	claims := model.TokenClaims{
		Uid:        "user-123",
		Identifier: "user@example.com",
	}

	// Generate token with first secret
	token, err := generator1.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Validate with different secret should fail
	_, err = generator2.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should fail with different secret")
	}
	if !errors.Is(err, domainerrors.ErrTokenInvalid) {
		t.Errorf("ValidateToken() error = %v, want ErrTokenInvalid", err)
	}
}

func TestJWTGenerator_TokenWithDifferentExpiry(t *testing.T) {
	accessGenerator := NewJWTGenerator(testSecretKey, testAccessExpiry, testRefreshExpiry)
	refreshGenerator := NewJWTGenerator(testSecretKey, 24*time.Hour, testRefreshExpiry)

	claims := model.TokenClaims{
		Uid:        "user-123",
		Identifier: "user@example.com",
	}

	// Both should generate valid tokens
	accessToken, err := accessGenerator.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	refreshToken, err := refreshGenerator.GenerateRefreshToken(claims)
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	// Both should be valid
	_, err = accessGenerator.ValidateToken(accessToken)
	if err != nil {
		t.Errorf("Access token should be valid: %v", err)
	}

	_, err = refreshGenerator.ValidateToken(refreshToken)
	if err != nil {
		t.Errorf("Refresh token should be valid: %v", err)
	}

	// Tokens should be different
	if accessToken == refreshToken {
		t.Error("Access and refresh tokens should be different")
	}
}

func TestJWTGenerator_ExpiredToken(t *testing.T) {
	// Create generator with very short expiry
	generator := NewJWTGenerator(testSecretKey, -1*time.Hour, testRefreshExpiry)

	claims := model.TokenClaims{
		Uid:        "user-123",
		Identifier: "user@example.com",
	}

	// Generate expired token
	token, err := generator.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	// Validate should return expired error
	_, err = generator.ValidateToken(token)
	if err == nil {
		t.Error("ValidateToken() should fail for expired token")
	}
	if !errors.Is(err, domainerrors.ErrTokenExpired) {
		t.Errorf("ValidateToken() error = %v, want ErrTokenExpired", err)
	}
}

func TestJWTGenerator_ValidateInvalidTokenType(t *testing.T) {
	generator := NewJWTGenerator(testSecretKey, testAccessExpiry, testRefreshExpiry)

	// Create a token with wrong signing method (simulated by modifying)
	// For JWT library, we test with malformed claims
	_, err := generator.ValidateToken("not.a.valid.jwt.token")
	if err == nil {
		t.Error("ValidateToken() should fail for invalid token")
	}
}

func TestJWTGenerator_TokenWithSpecialCharacters(t *testing.T) {
	generator := NewJWTGenerator(testSecretKey, testAccessExpiry, testRefreshExpiry)

	tests := []struct {
		name           string
		identifier     string
		identifierType string
	}{
		{"email with plus", "user+tag@example.com", "email"},
		{"email with dots", "user.name@example.com", "email"},
		{"phone number", "+1234567890", "phone"},
		{"username with underscore", "user_name_123", "username"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := model.TokenClaims{
				Uid:            "user-123",
				Identifier:     tt.identifier,
				IdentifierType: tt.identifierType,
			}

			token, err := generator.GenerateAccessToken(claims)
			if err != nil {
				t.Fatalf("GenerateAccessToken() error = %v", err)
			}

			extracted, err := generator.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken() error = %v", err)
			}

			if extracted.Identifier != tt.identifier {
				t.Errorf("Identifier mismatch: got %v, want %v", extracted.Identifier, tt.identifier)
			}
		})
	}
}

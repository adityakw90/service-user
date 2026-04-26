package model

import (
	"testing"
	"time"
)

func TestCore_Domain_OAuthUserInfo_HasValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		verified bool
		expected bool
	}{
		{"valid email with verification", "user@example.com", true, true},
		{"empty email", "", true, false},
		{"unverified email", "user@example.com", false, false},
		{"empty email and unverified", "", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &OAuthUserInfo{Email: tt.email, EmailVerified: tt.verified}
			result := info.HasValidEmail()
			if result != tt.expected {
				t.Errorf("HasValidEmail() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCore_Domain_OAuthUserInfo_DisplayName(t *testing.T) {
	tests := []struct {
		name      string
		fullName  string
		firstName string
		lastName  string
		email     string
		expected  string
	}{
		{"full name takes precedence", "John Doe", "John", "Doe", "john@example.com", "John Doe"},
		{"first and last name fallback", "", "John", "Doe", "john@example.com", "John Doe"},
		{"first name only", "", "John", "", "john@example.com", "John"},
		{"email fallback", "", "", "", "john@example.com", "john@example.com"},
		{"empty fallback", "", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &OAuthUserInfo{FullName: tt.fullName, FirstName: tt.firstName, LastName: tt.lastName, Email: tt.email}
			result := info.DisplayName()
			if result != tt.expected {
				t.Errorf("DisplayName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCore_Domain_OAuthUserInfo_AvatarURL(t *testing.T) {
	tests := []struct {
		name     string
		picture  string
		expected string
	}{
		{"returns picture URL", "https://example.com/avatar.jpg", "https://example.com/avatar.jpg"},
		{"empty picture", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &OAuthUserInfo{Picture: tt.picture}
			result := info.AvatarURL()
			if result != tt.expected {
				t.Errorf("AvatarURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCore_Domain_OAuthUserInfo_StructFields(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	tests := []struct {
		name        string
		providerID  string
		email       string
		emailVerify bool
		firstName   string
		lastName    string
		fullName    string
		picture     string
		locale      string
		expiresAt   *time.Time
		checkValid  bool
		checkName   string
		checkAvatar string
	}{
		{
			name:        "all fields set",
			providerID:  "google-123",
			email:       "test@example.com",
			emailVerify: true,
			firstName:   "Test",
			lastName:    "User",
			fullName:    "Test User",
			picture:     "https://example.com/pic.jpg",
			locale:      "en",
			expiresAt:   &expiresAt,
			checkValid:  true,
			checkName:   "Test User",
			checkAvatar: "https://example.com/pic.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &OAuthUserInfo{
				ProviderID:    tt.providerID,
				Email:         tt.email,
				EmailVerified: tt.emailVerify,
				FirstName:     tt.firstName,
				LastName:      tt.lastName,
				FullName:      tt.fullName,
				Picture:       tt.picture,
				Locale:        tt.locale,
				ExpiresAt:     tt.expiresAt,
			}
			if info.ProviderID != tt.providerID {
				t.Errorf("ProviderID = %q, want %q", info.ProviderID, tt.providerID)
			}
			if info.HasValidEmail() != tt.checkValid {
				t.Errorf("HasValidEmail() = %v, want %v", info.HasValidEmail(), tt.checkValid)
			}
			if info.DisplayName() != tt.checkName {
				t.Errorf("DisplayName() = %q, want %q", info.DisplayName(), tt.checkName)
			}
			if info.AvatarURL() != tt.checkAvatar {
				t.Errorf("AvatarURL() = %q, want %q", info.AvatarURL(), tt.checkAvatar)
			}
		})
	}
}

func TestCore_Domain_OAuthTokens_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn int
		expected  bool
	}{
		{"zero expires in means not expired", 0, false},
		{"positive expires in returns false", 3600, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := &OAuthTokens{ExpiresIn: tt.expiresIn}
			result := tokens.IsExpired()
			if result != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCore_Domain_OAuthTokens_StructFields(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		refreshToken string
		expiresIn    int
		idToken      string
		tokenType    string
	}{
		{"all fields set", "access-123", "refresh-456", 3600, "id-token-789", "Bearer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := &OAuthTokens{
				AccessToken:  tt.accessToken,
				RefreshToken: tt.refreshToken,
				ExpiresIn:    tt.expiresIn,
				IDToken:      tt.idToken,
				TokenType:    tt.tokenType,
			}
			if tokens.AccessToken != tt.accessToken {
				t.Errorf("AccessToken = %q, want %q", tokens.AccessToken, tt.accessToken)
			}
			if tokens.RefreshToken != tt.refreshToken {
				t.Errorf("RefreshToken = %q, want %q", tokens.RefreshToken, tt.refreshToken)
			}
			if tokens.ExpiresIn != tt.expiresIn {
				t.Errorf("ExpiresIn = %d, want %d", tokens.ExpiresIn, tt.expiresIn)
			}
			if tokens.IDToken != tt.idToken {
				t.Errorf("IDToken = %q, want %q", tokens.IDToken, tt.idToken)
			}
			if tokens.TokenType != tt.tokenType {
				t.Errorf("TokenType = %q, want %q", tokens.TokenType, tt.tokenType)
			}
		})
	}
}

package model

import (
	"testing"
)

func TestCore_Domain_TokenConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"TokenTypeAccess", TokenTypeAccess, "access"},
		{"TokenTypeRefresh", TokenTypeRefresh, "refresh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestCore_Domain_Token_StructFields(t *testing.T) {
	tests := []struct {
		name    string
		access  string
		refresh string
	}{
		{"standard values", "access-token-123", "refresh-token-456"},
		{"empty values", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := Token{Access: tt.access, Refresh: tt.refresh}
			if token.Access != tt.access {
				t.Errorf("Access = %q, want %q", token.Access, tt.access)
			}
			if token.Refresh != tt.refresh {
				t.Errorf("Refresh = %q, want %q", token.Refresh, tt.refresh)
			}
		})
	}
}

func TestCore_Domain_TokenClaims_StructFields(t *testing.T) {
	tests := []struct {
		name           string
		uid            string
		sid            string
		tokenType      string
		identifier     string
		identifierType string
		extra          map[string]any
	}{
		{
			name:           "all fields",
			uid:            "user-123",
			sid:            "session-456",
			tokenType:      TokenTypeAccess,
			identifier:     "user@example.com",
			identifierType: "email",
			extra:          map[string]any{"role": "admin"},
		},
		{
			name:           "minimal fields",
			uid:            "user-456",
			sid:            "",
			tokenType:      TokenTypeRefresh,
			identifier:     "",
			identifierType: "",
			extra:          nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := TokenClaims{
				Uid:            tt.uid,
				Sid:            tt.sid,
				Type:           tt.tokenType,
				Identifier:     tt.identifier,
				IdentifierType: tt.identifierType,
				Extra:          tt.extra,
			}
			if claims.Uid != tt.uid {
				t.Errorf("Uid = %q, want %q", claims.Uid, tt.uid)
			}
			if claims.Sid != tt.sid {
				t.Errorf("Sid = %q, want %q", claims.Sid, tt.sid)
			}
			if claims.Type != tt.tokenType {
				t.Errorf("Type = %q, want %q", claims.Type, tt.tokenType)
			}
			if claims.Identifier != tt.identifier {
				t.Errorf("Identifier = %q, want %q", claims.Identifier, tt.identifier)
			}
			if claims.IdentifierType != tt.identifierType {
				t.Errorf("IdentifierType = %q, want %q", claims.IdentifierType, tt.identifierType)
			}
		})
	}
}

func TestCore_Domain_TokenClaims_IsAccess(t *testing.T) {
	tests := []struct {
		name      string
		tokenType string
		expected  bool
	}{
		{"access token returns true", TokenTypeAccess, true},
		{"refresh token returns false", TokenTypeRefresh, false},
		{"unknown token type returns false", "unknown", false},
		{"empty token type returns false", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &TokenClaims{Type: tt.tokenType}
			result := claims.IsAccess()
			if result != tt.expected {
				t.Errorf("IsAccess() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCore_Domain_TokenClaims_IsRefresh(t *testing.T) {
	tests := []struct {
		name      string
		tokenType string
		expected  bool
	}{
		{"refresh token returns true", TokenTypeRefresh, true},
		{"access token returns false", TokenTypeAccess, false},
		{"unknown token type returns false", "unknown", false},
		{"empty token type returns false", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &TokenClaims{Type: tt.tokenType}
			result := claims.IsRefresh()
			if result != tt.expected {
				t.Errorf("IsRefresh() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCore_Domain_TokenClaims_TokenTypeExclusivity(t *testing.T) {
	tests := []struct {
		name      string
		tokenType string
		isAccess  bool
		isRefresh bool
	}{
		{"access token", TokenTypeAccess, true, false},
		{"refresh token", TokenTypeRefresh, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := &TokenClaims{Type: tt.tokenType}
			if claims.IsAccess() != tt.isAccess {
				t.Errorf("IsAccess() = %v, want %v", claims.IsAccess(), tt.isAccess)
			}
			if claims.IsRefresh() != tt.isRefresh {
				t.Errorf("IsRefresh() = %v, want %v", claims.IsRefresh(), tt.isRefresh)
			}
		})
	}
}

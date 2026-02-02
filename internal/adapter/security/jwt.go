package security

import (
	"errors"
	"fmt"
	"time"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the claims in a JWT token.
type JWTClaims struct {
	Uid            string         `json:"uid"`
	Identifier     string         `json:"identifier"`
	IdentifierType string         `json:"identifier_type"`
	Extra          map[string]any `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// JWTGenerator generates JWT tokens.
type JWTGenerator struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTGenerator creates a new JWT token generator.
func NewJWTGenerator(secretKey string, accessExpiry, refreshExpiry time.Duration) portsec.TokenGenerator {
	return &JWTGenerator{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateAccessToken generates a new access token.
func (g *JWTGenerator) GenerateAccessToken(claims model.TokenClaims) (string, error) {
	now := time.Now()
	jwtClaims := JWTClaims{
		Uid:            claims.Uid,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
		Extra:          claims.Extra,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(g.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "service-user",
			Subject:   claims.Uid,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(g.secretKey)
}

// GenerateRefreshToken generates a new refresh token.
func (g *JWTGenerator) GenerateRefreshToken(claims model.TokenClaims) (string, error) {
	now := time.Now()
	jwtClaims := JWTClaims{
		Uid:            claims.Uid,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
		Extra:          claims.Extra,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(g.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "service-user",
			Subject:   claims.Uid,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(g.secretKey)
}

// ValidateToken validates a token and returns the claims.
func (g *JWTGenerator) ValidateToken(tokenString string) (*model.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return g.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, domainerrors.ErrTokenExpired
		}
		return nil, domainerrors.ErrTokenInvalid
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, domainerrors.ErrTokenInvalidClaim
	}

	return &model.TokenClaims{
		Uid:            claims.Uid,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
		Extra:          claims.Extra,
	}, nil
}

package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
)

type authService struct {
}

func NewAuthService() portSvc.AuthService {
	return &authService{}
}

func (s *authService) Authenticate(ctx context.Context, payload *params.AuthParams) (*model.Token, error) {
	return nil, nil
}

func (s *authService) GoogleOAuth(ctx context.Context, redirectURI string) (string, error) {
	return "", nil
}

func (s *authService) HandleGoogleOAuth(ctx context.Context, code, redirectURI string) (*model.Token, error) {
	return nil, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.Token, error) {
	return nil, nil
}

func (s *authService) ValidateToken(ctx context.Context, accessToken string) (*model.TokenClaims, error) {
	return nil, nil
}

func (s *authService) RevokeToken(ctx context.Context, token string, tokenType string) error {
	return nil
}

func (s *authService) VerifyPin(ctx context.Context, userUid string, pin string) (bool, error) {
	return false, nil
}

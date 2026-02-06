package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	auth "github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	portsvc "github.com/adityakw90/service-user/internal/core/port/service"
)

// AuthHandler implements the gRPC AuthService.
type AuthHandler struct {
	auth.UnimplementedAuthServiceServer
	service   portsvc.AuthService
	validator *validator.Validator
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(service portsvc.AuthService) *AuthHandler {
	return &AuthHandler{
		service:   service,
		validator: validator.New(),
	}
}

// Auth authenticates a user and returns tokens.
func (h *AuthHandler) Auth(ctx context.Context, req *auth.AuthRequest) (*auth.Token, error) {
	// Validate request using DTO
	dto := validator.AuthRequestDTO{
		Identifier:        req.Identifier,
		Password:          req.Password,
		IdentifierType:    req.IdentifierType,
		DeviceFingerprint: req.DeviceFingerprint,
		DeviceName:        req.DeviceName,
	}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	payload := &params.AuthParams{
		Identifier:        req.Identifier,
		IdentifierType:    req.IdentifierType,
		Password:          req.Password,
		DeviceFingerprint: req.DeviceFingerprint,
		DeviceName:        req.DeviceName,
	}

	result, err := h.service.Authenticate(ctx, payload)
	if err != nil {
		return nil, response.MapAuthError(err)
	}

	return &auth.Token{
		AccessToken:  result.Access,
		RefreshToken: result.Refresh,
	}, nil
}

// RefreshToken refreshes access token using a refresh token.
func (h *AuthHandler) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (*auth.Token, error) {
	dto := validator.RefreshTokenRequestDTO{RefreshToken: req.RefreshToken}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	result, err := h.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, response.MapAuthError(err)
	}

	return &auth.Token{
		AccessToken:  result.Access,
		RefreshToken: result.Refresh,
	}, nil
}

// ValidateToken validates an access token and returns claims.
func (h *AuthHandler) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	dto := validator.ValidateTokenRequestDTO{AccessToken: req.AccessToken}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	claims, err := h.service.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return nil, response.MapAuthError(err)
	}

	resp := &auth.ValidateTokenResponse{
		Uid:            claims.Uid,
		Identifier:     claims.Identifier,
		IdentifierType: claims.IdentifierType,
	}

	if claims.Extra != nil {
		structVal, _ := structpb.NewStruct(claims.Extra)
		resp.Claims = structVal
	}

	return resp, nil
}

// VerifyPin verifies a user's PIN.
func (h *AuthHandler) VerifyPin(ctx context.Context, req *auth.VerifyPinRequest) (*auth.VerifyPinResponse, error) {
	dto := validator.VerifyPinRequestDTO{Uid: req.Uid, Code: req.Code}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	valid, err := h.service.VerifyPin(ctx, req.Uid, req.Code)
	if err != nil {
		return nil, response.MapAuthError(err)
	}

	return &auth.VerifyPinResponse{Valid: valid}, nil
}

// GoogleOAuth initiates Google OAuth flow.
func (h *AuthHandler) GoogleOAuth(ctx context.Context, req *auth.GoogleOAuthRequest) (*auth.GoogleOAuthResponse, error) {
	dto := validator.GoogleOAuthRequestDTO{RedirectUri: req.RedirectUri}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	url, err := h.service.GoogleOAuth(ctx, req.RedirectUri)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to initiate OAuth: %v", err)
	}

	return &auth.GoogleOAuthResponse{AuthorizationUrl: url}, nil
}

// HandleGoogleOAuth handles the OAuth callback.
func (h *AuthHandler) HandleGoogleOAuth(ctx context.Context, req *auth.HandleGoogleOAuthRequest) (*auth.Token, error) {
	dto := validator.HandleGoogleOAuthRequestDTO{Code: req.Code, RedirectUri: req.RedirectUri}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	result, err := h.service.HandleGoogleOAuth(ctx, req.Code, req.RedirectUri)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to handle OAuth: %v", err)
	}

	return &auth.Token{
		AccessToken:  result.Access,
		RefreshToken: result.Refresh,
	}, nil
}

// RevokeToken revokes a token.
func (h *AuthHandler) RevokeToken(ctx context.Context, req *auth.RevokeTokenRequest) (*auth.RevokeTokenResponse, error) {
	dto := validator.RevokeTokenRequestDTO{Token: req.Token, TokenType: req.TokenType}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.service.RevokeToken(ctx, req.Token, req.TokenType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke token: %v", err)
	}

	return &auth.RevokeTokenResponse{Success: true}, nil
}

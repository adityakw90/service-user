package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	auth "github.com/adityakw90/service-user-proto/gen/go/auth"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
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
	r := request.AuthRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	payload := r.ToAuthParams()

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
	r := request.RefreshTokenRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
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
	r := request.ValidateTokenRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
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
	r := request.VerifyPinRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
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
	r := request.GoogleOAuthRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	url, _, err := h.service.GoogleOAuth(ctx, req.GetRedirectUri())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to initiate OAuth: %v", err)
	}

	// TODO: Include state in response once v0.1.1 is released
	return &auth.GoogleOAuthResponse{AuthorizationUrl: url}, nil
}

// HandleGoogleOAuth handles the OAuth callback.
func (h *AuthHandler) HandleGoogleOAuth(ctx context.Context, req *auth.HandleGoogleOAuthRequest) (*auth.Token, error) {
	r := request.HandleGoogleOAuthRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	result, err := h.service.HandleGoogleOAuth(ctx, req.GetCode(), req.GetState(), req.GetRedirectUri())
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
	r := request.RevokeTokenRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	err := h.service.RevokeToken(ctx, req.Token, req.TokenType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to revoke token: %v", err)
	}

	return &auth.RevokeTokenResponse{Success: true}, nil
}

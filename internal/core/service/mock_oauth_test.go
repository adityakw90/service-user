package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// MockOAuthProvider is a mock implementation of port.OAuthProvider.
type MockOAuthProvider struct {
	GetAuthorizationURLFunc func(ctx context.Context, redirectURI string, state string) (string, error)
	ExchangeCodeFunc        func(ctx context.Context, code string, state string, redirectURI string) (*model.OAuthTokens, error)
	GetUserInfoFunc         func(ctx context.Context, token *model.OAuthTokens) (*model.OAuthUserInfo, error)

	getAuthorizationURLCalls int
	exchangeCodeCalls        int
	getUserInfoCalls         int
}

func NewMockOAuthProvider() *MockOAuthProvider {
	return &MockOAuthProvider{}
}

func (m *MockOAuthProvider) GetAuthorizationURL(ctx context.Context, redirectURI string, state string) (string, error) {
	m.getAuthorizationURLCalls++
	if m.GetAuthorizationURLFunc != nil {
		return m.GetAuthorizationURLFunc(ctx, redirectURI, state)
	}
	return "https://oauth.example.com/auth?state=" + state, nil
}

func (m *MockOAuthProvider) ExchangeCode(ctx context.Context, code string, state string, redirectURI string) (*model.OAuthTokens, error) {
	m.exchangeCodeCalls++
	if m.ExchangeCodeFunc != nil {
		return m.ExchangeCodeFunc(ctx, code, state, redirectURI)
	}
	return &model.OAuthTokens{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		TokenType:    "Bearer",
	}, nil
}

func (m *MockOAuthProvider) GetUserInfo(ctx context.Context, token *model.OAuthTokens) (*model.OAuthUserInfo, error) {
	m.getUserInfoCalls++
	if m.GetUserInfoFunc != nil {
		return m.GetUserInfoFunc(ctx, token)
	}
	return &model.OAuthUserInfo{
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}, nil
}

// MockEventPublisher is a mock implementation of port.EventPublisher.
type MockEventPublisher struct {
	PublishFunc func(ctx context.Context, eventType event.EventType, eventData any) error
	CloseFunc   func() error

	publishCalls int
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{}
}

func (m *MockEventPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	m.publishCalls++
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, eventType, eventData)
	}
	return nil
}

func (m *MockEventPublisher) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

package service

import (
	"context"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// MockHasher is a mock implementation of security.Hasher.
type MockHasher struct {
	HashFunc    func(plain string) (string, error)
	CompareFunc func(hashed string, plain string) bool

	hashCalls    int
	compareCalls int
}

func NewMockHasher() *MockHasher {
	return &MockHasher{}
}

func (m *MockHasher) Hash(plain string) (string, error) {
	m.hashCalls++
	if m.HashFunc != nil {
		return m.HashFunc(plain)
	}
	return "hashed_" + plain, nil
}

func (m *MockHasher) Compare(hashed string, plain string) bool {
	m.compareCalls++
	if m.CompareFunc != nil {
		return m.CompareFunc(hashed, plain)
	}
	// Default: simple string comparison for "hashed_" + plain
	return hashed == "hashed_"+plain || hashed == plain
}

// MockTokenGenerator is a mock implementation of security.TokenGenerator.
type MockTokenGenerator struct {
	GenerateTokenFunc func(claims *model.TokenClaims) (string, error)
	ValidateTokenFunc func(token string) (*model.TokenClaims, error)

	generateCalls int
	validateCalls int
}

func NewMockTokenGenerator() *MockTokenGenerator {
	return &MockTokenGenerator{}
}

func (m *MockTokenGenerator) GenerateToken(claims *model.TokenClaims) (string, error) {
	m.generateCalls++
	if m.GenerateTokenFunc != nil {
		return m.GenerateTokenFunc(claims)
	}
	return "mock_token_" + claims.Uid + "_" + claims.Sid, nil
}

func (m *MockTokenGenerator) ValidateToken(token string) (*model.TokenClaims, error) {
	m.validateCalls++
	if m.ValidateTokenFunc != nil {
		return m.ValidateTokenFunc(token)
	}
	// Default: return valid claims for mock tokens
	if len(token) > 11 && token[:11] == "mock_token_" {
		return &model.TokenClaims{
			Uid:  "test-uid",
			Sid:  "test-sid",
			Type: model.TokenTypeAccess,
		}, nil
	}
	return nil, domainerrors.ErrTokenExpired
}

// MockUIDGenerator is a mock implementation of security.UIDGenerator.
type MockUIDGenerator struct {
	NewCalls int
	NewFunc  func() string
}

func NewMockUIDGenerator() *MockUIDGenerator {
	return &MockUIDGenerator{}
}

func (m *MockUIDGenerator) New() string {
	m.NewCalls++
	if m.NewFunc != nil {
		return m.NewFunc()
	}
	return "test-uid-" + string(rune(m.NewCalls))
}

// MockTokenStore is a mock implementation of security.TokenStore.
type MockTokenStore struct {
	AddFunc       func(ctx context.Context, userUID string, tid string) error
	RemoveFunc    func(ctx context.Context, userUID string, tid string) error
	RemoveAllFunc func(ctx context.Context, userUID string) error
	IsAllowedFunc func(ctx context.Context, userUID string, tid string) (bool, error)

	addCalls       int
	removeCalls    int
	removeAllCalls int
	isAllowedCalls int

	// Internal storage for mock behavior
	tokens map[string]map[string]bool // userUID -> sessionID -> allowed
}

func NewMockTokenStore() *MockTokenStore {
	return &MockTokenStore{
		tokens: make(map[string]map[string]bool),
	}
}

func (m *MockTokenStore) Add(ctx context.Context, userUID string, tid string) error {
	m.addCalls++
	if m.AddFunc != nil {
		return m.AddFunc(ctx, userUID, tid)
	}
	if m.tokens[userUID] == nil {
		m.tokens[userUID] = make(map[string]bool)
	}
	m.tokens[userUID][tid] = true
	return nil
}

func (m *MockTokenStore) Remove(ctx context.Context, userUID string, tid string) error {
	m.removeCalls++
	if m.RemoveFunc != nil {
		return m.RemoveFunc(ctx, userUID, tid)
	}
	if m.tokens[userUID] != nil {
		delete(m.tokens[userUID], tid)
	}
	return nil
}

func (m *MockTokenStore) RemoveAll(ctx context.Context, userUID string) error {
	m.removeAllCalls++
	if m.RemoveAllFunc != nil {
		return m.RemoveAllFunc(ctx, userUID)
	}
	delete(m.tokens, userUID)
	return nil
}

func (m *MockTokenStore) IsAllowed(ctx context.Context, userUID string, tid string) (bool, error) {
	m.isAllowedCalls++
	if m.IsAllowedFunc != nil {
		return m.IsAllowedFunc(ctx, userUID, tid)
	}
	if m.tokens[userUID] == nil {
		return false, nil
	}
	return m.tokens[userUID][tid], nil
}

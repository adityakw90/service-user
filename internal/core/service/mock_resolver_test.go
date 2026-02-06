package service

import "context"

// MockUserResolver is a mock implementation of resolver.UserResolver.
type MockUserResolver struct {
	IDsByUIDsFunc func(ctx context.Context, userUIDs []string) (map[string]int64, error)
	UIDsByIDsFunc func(ctx context.Context, userIDs []int64) (map[int64]string, error)

	idsByUIDsCalls int
	uidsByIDsCalls int
}

func NewMockUserResolver() *MockUserResolver {
	return &MockUserResolver{}
}

func (m *MockUserResolver) IDsByUIDs(ctx context.Context, userUIDs []string) (map[string]int64, error) {
	m.idsByUIDsCalls++
	if m.IDsByUIDsFunc != nil {
		return m.IDsByUIDsFunc(ctx, userUIDs)
	}
	return make(map[string]int64), nil
}

func (m *MockUserResolver) UIDsByIDs(ctx context.Context, userIDs []int64) (map[int64]string, error) {
	m.uidsByIDsCalls++
	if m.UIDsByIDsFunc != nil {
		return m.UIDsByIDsFunc(ctx, userIDs)
	}
	return make(map[int64]string), nil
}

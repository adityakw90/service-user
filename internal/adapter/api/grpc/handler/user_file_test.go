package handler

import (
	"testing"

	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	servicemocks "github.com/adityakw90/service-user/mocks/service"
	"github.com/stretchr/testify/assert"
)

// TestNewUserFileHandler tests the NewUserFileHandler constructor.
func TestNewUserFileHandler(t *testing.T) {
	mockService := servicemocks.NewMockUserFileService(t)
	v := validator.New()

	h := NewUserFileHandler(mockService, v)

	assert.NotNil(t, h)
	assert.Equal(t, mockService, h.service)
	assert.Equal(t, v, h.validator)
}

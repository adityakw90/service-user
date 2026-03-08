package handler

import (
	"testing"

	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	servicemocks "github.com/adityakw90/service-user/test/mocks/service"
	"github.com/stretchr/testify/assert"
)

// TestNewUserHandler tests the NewUserHandler constructor.
func TestNewUserHandler(t *testing.T) {
	mockService := servicemocks.NewMockUserService(t)
	v := validator.New()

	h := NewUserHandler(mockService, v)

	assert.NotNil(t, h)
	assert.Equal(t, mockService, h.service)
	assert.Equal(t, v, h.validator)
}

// TestNewUserHandler_ServiceNil tests that handler can be created with nil service.
func TestNewUserHandler_ServiceNil(t *testing.T) {
	v := validator.New()

	h := NewUserHandler(nil, v)

	assert.NotNil(t, h)
	assert.Nil(t, h.service)
	assert.Equal(t, v, h.validator)
}

// TestNewUserHandler_ValidatorNil tests that handler can be created with nil validator.
func TestNewUserHandler_ValidatorNil(t *testing.T) {
	mockService := servicemocks.NewMockUserService(t)

	h := NewUserHandler(mockService, nil)

	assert.NotNil(t, h)
	assert.Equal(t, mockService, h.service)
	assert.Nil(t, h.validator)
}


package handler

import (
	"testing"

	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	servicemocks "github.com/adityakw90/service-user/test/mocks/service"
	"github.com/stretchr/testify/assert"
)

// TestNewDeviceHandler tests the NewDeviceHandler constructor.
func TestNewDeviceHandler(t *testing.T) {
	mockService := servicemocks.NewMockDeviceService(t)
	v := validator.New()

	h := NewDeviceHandler(mockService, v)

	assert.NotNil(t, h)
	assert.Equal(t, mockService, h.service)
	assert.Equal(t, v, h.validator)
}

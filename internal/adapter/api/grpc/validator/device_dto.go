package validator

// DeviceGetRequestDTO represents validated device get request.
type DeviceGetRequestDTO struct {
	Uid string `validate:"required"`
}

// DeviceDeleteRequestDTO represents validated device delete request.
type DeviceDeleteRequestDTO struct {
	Uid string `validate:"required"`
}

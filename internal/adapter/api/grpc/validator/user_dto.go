package validator

// ListRequestDTO represents validated list request.
type ListRequestDTO struct {
	Page  int `validate:"omitempty,min=1"`
	Limit int `validate:"omitempty,min=1,max=100"`
}

// GetRequestDTO represents validated get request.
type GetRequestDTO struct {
	Uid string `validate:"required"`
}

// AddRequestDTO represents validated user creation request.
type AddRequestDTO struct {
	Username string `validate:"required,min=3,max=50,alphanum"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=128"`
}

// UpdateRequestDTO represents validated user update request.
type UpdateRequestDTO struct {
	Uid       string `validate:"required"`
	Username  string `validate:"omitempty,min=3,max=50,alphanum"`
	Email     string `validate:"omitempty,email"`
	Password  string `validate:"omitempty,min=8,max=128"`
	StatusPtr *int32 // Status is int32 in proto, validated separately
}

// DeleteRequestDTO represents validated delete request.
type DeleteRequestDTO struct {
	Uid string `validate:"required"`
}

// GetProfileRequestDTO represents validated get profile request.
type GetProfileRequestDTO struct {
	UserUid string `validate:"required"`
}

// UpdateProfileRequestDTO represents validated profile update request.
type UpdateProfileRequestDTO struct {
	UserUid   string `validate:"required"`
	FirstName string `validate:"omitempty,min=1,max=100"`
	LastName  string `validate:"omitempty,min=1,max=100"`
	Bio       string `validate:"omitempty,max=500"`
}

// UpdatePinRequestDTO represents validated PIN update request.
type UpdatePinRequestDTO struct {
	UserUid string `validate:"required"`
	PIN     string `validate:"required,pin"`
}

// ListDevicesRequestDTO represents validated list devices request.
type ListDevicesRequestDTO struct {
	UserUid    string `validate:"required"`
	DeviceName string `validate:"omitempty,max=100"`
	Revoked    *bool
}

// RevokeDeviceRequestDTO represents validated device revocation request.
type RevokeDeviceRequestDTO struct {
	UserUid   string `validate:"required"`
	DeviceUid string `validate:"required"`
}

// ChangePasswordRequestDTO represents validated password change request.
type ChangePasswordRequestDTO struct {
	Uid             string `validate:"required"`
	CurrentPassword string `validate:"required,min=8,max=128"`
	NewPassword     string `validate:"required,min=8,max=128"`
	ConfirmPassword string `validate:"required,min=8,max=128,eqfield=NewPassword"`
}

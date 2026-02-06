package validator

// AuthRequestDTO represents validated auth request data.
type AuthRequestDTO struct {
	Identifier        string `validate:"required"`
	Password          string `validate:"required"`
	IdentifierType    string `validate:"omitempty,oneof=email username phone"`
	DeviceFingerprint string `validate:"omitempty,max=255"`
	DeviceName        string `validate:"omitempty,max=100"`
}

// RefreshTokenRequestDTO represents validated refresh token request.
type RefreshTokenRequestDTO struct {
	RefreshToken string `validate:"required"`
}

// ValidateTokenRequestDTO represents validated token validation request.
type ValidateTokenRequestDTO struct {
	AccessToken string `validate:"required"`
}

// VerifyPinRequestDTO represents validated PIN verification request.
type VerifyPinRequestDTO struct {
	Uid  string `validate:"required"`
	Code string `validate:"required,pin"`
}

// GoogleOAuthRequestDTO represents validated Google OAuth request.
type GoogleOAuthRequestDTO struct {
	RedirectUri string `validate:"required,uri"`
}

// HandleGoogleOAuthRequestDTO represents validated OAuth callback request.
type HandleGoogleOAuthRequestDTO struct {
	Code        string `validate:"required"`
	RedirectUri string `validate:"omitempty,uri"`
}

// RevokeTokenRequestDTO represents validated token revocation request.
type RevokeTokenRequestDTO struct {
	Token     string `validate:"required"`
	TokenType string `validate:"omitempty,oneof=access refresh"`
}

// SetPasswordRequestDTO represents validated password creation request.
type SetPasswordRequestDTO struct {
	UserUID  string `validate:"required"`
	Password string `validate:"required,min=12,password_strength,no_repeated,mixed_chars"`
}

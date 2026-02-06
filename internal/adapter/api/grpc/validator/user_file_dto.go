package validator

// UserFileGetRequestDTO represents validated user file get request.
type UserFileGetRequestDTO struct {
	Uid string `validate:"required"`
}

// UserFileAddRequestDTO represents validated user file add request.
type UserFileAddRequestDTO struct {
	UserUid  string `validate:"required"`
	Name     string `validate:"required,min=1,max=255"`
	Filename string `validate:"required,min=1,max=255"`
	FileData []byte `validate:"max=10485760"` // 10MB max
}

// UserFileUpdateRequestDTO represents validated user file update request.
type UserFileUpdateRequestDTO struct {
	Uid      string `validate:"required"`
	Name     string `validate:"omitempty,min=1,max=255"`
	Filename string `validate:"omitempty,min=1,max=255"`
	FileData []byte `validate:"omitempty,max=10485760"` // 10MB max
}

// UserFileDeleteRequestDTO represents validated user file delete request.
type UserFileDeleteRequestDTO struct {
	Uid string `validate:"required"`
}

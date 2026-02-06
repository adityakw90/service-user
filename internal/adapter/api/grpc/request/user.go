package request

import (
	"strings"

	user "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

// UserGetRequest represents validated get request.
type UserGetRequest struct {
	Uid string `validate:"required"`
}

// UserGetRequestFromPb creates a UserGetRequest from protobuf.
func UserGetRequestFromPb(req *user.GetRequest) *UserGetRequest {
	return &UserGetRequest{
		Uid: strings.TrimSpace(req.Uid),
	}
}

type UserFilterRequest struct {
	Uids     []string `validate:"omitempty"`
	Username *string  `validate:"omitempty,min=3,max=50,username"`
	Email    *string  `validate:"omitempty,email"`
	Status   *int32   `validate:"omitempty"`
	Exists   *bool    `validate:"omitempty"`
	Query    *string  `validate:"omitempty"`
}

func (r *UserFilterRequest) ToUserFilterParams() *params.UserListFilterParam {
	var status model.UserStatus
	if r.Status != nil {
		status = model.UserStatus(*r.Status)
	}
	return &params.UserListFilterParam{
		Uids:     r.Uids,
		Username: r.Username,
		Email:    r.Email,
		Status:   &status,
		Exists:   r.Exists,
		Query:    r.Query,
	}
}

func UserFilterRequestFromPb(req *user.FilterRequest) *UserFilterRequest {
	payload := &UserFilterRequest{
		Uids: req.GetUids(),
	}

	if req.Username != nil {
		username := strings.TrimSpace(*req.Username)
		payload.Username = &username
	}

	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		payload.Email = &email
	}

	if req.Status != nil {
		status := int32(*req.Status)
		payload.Status = &status
	}

	if req.Exists != nil {
		exists := *req.Exists
		payload.Exists = &exists
	}

	if req.Query != nil {
		query := *req.Query
		payload.Query = &query
	}

	return payload
}

// UserListRequest represents validated list request.
type UserListRequest struct {
	Pagination *PaginationRequest
	Filter     *UserFilterRequest
}

func (r *UserListRequest) ToUserListParams() *params.UserListParam {
	return &params.UserListParam{
		Pagination: r.Pagination.ToPaginationParams(),
		Filter:     r.Filter.ToUserFilterParams(),
	}
}

// UserListRequestFromPb creates a UserListRequest from protobuf.
func UserListRequestFromPb(req *user.ListRequest) *UserListRequest {
	payload := &UserListRequest{}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = UserFilterRequestFromPb(req.GetFilter())
	}

	return payload
}

// UserAddRequest represents validated user creation request.
type UserAddRequest struct {
	Username string `validate:"required,min=3,max=50,username"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=128"`
}

// UserAddRequestFromPb creates a UserAddRequest from protobuf.
func UserAddRequestFromPb(req *user.AddRequest) *UserAddRequest {
	return &UserAddRequest{
		Username: strings.TrimSpace(req.Username),
		Email:    strings.TrimSpace(req.Email),
		Password: strings.TrimSpace(req.Password),
	}
}

// UserUpdateRequest represents validated user update request.
type UserUpdateRequest struct {
	Uid       string  `validate:"required"`
	Username  *string `validate:"omitempty,min=3,max=50,username"`
	Email     *string `validate:"omitempty,email"`
	Password  *string `validate:"omitempty,min=8,max=128"`
	StatusPtr *int32  // Status is int32 in proto, validated separately
}

// UserUpdateRequestFromPb creates a UserUpdateRequest from protobuf.
func UserUpdateRequestFromPb(req *user.UpdateRequest) *UserUpdateRequest {
	r := &UserUpdateRequest{
		Uid:       strings.TrimSpace(req.Uid),
		StatusPtr: req.Status,
	}

	if req.Username != nil {
		username := strings.TrimSpace(req.GetUsername())
		r.Username = &username
	}

	if req.Email != nil {
		email := strings.TrimSpace(req.GetEmail())
		r.Email = &email
	}

	if req.Password != nil {
		password := strings.TrimSpace(req.GetPassword())
		r.Password = &password
	}

	return r
}

// UserDeleteRequest represents validated delete request.
type UserDeleteRequest struct {
	Uid string `validate:"required"`
}

// UserDeleteRequestFromPb creates a UserDeleteRequest from protobuf.
func UserDeleteRequestFromPb(req *user.DeleteRequest) *UserDeleteRequest {
	return &UserDeleteRequest{
		Uid: strings.TrimSpace(req.Uid),
	}
}

// UserGetProfileRequest represents validated get profile request.
type UserGetProfileRequest struct {
	UserUid string `validate:"required"`
}

// UserGetProfileRequestFromPb creates a UserGetProfileRequest from protobuf.
func UserGetProfileRequestFromPb(req *user.GetProfileRequest) *UserGetProfileRequest {
	return &UserGetProfileRequest{
		UserUid: strings.TrimSpace(req.UserUid),
	}
}

// UserUpdateProfileRequest represents validated profile update request.
type UserUpdateProfileRequest struct {
	UserUid   string  `validate:"required"`
	FirstName *string `validate:"omitempty,min=1,max=100"`
	LastName  *string `validate:"omitempty,min=1,max=100"`
	Bio       *string `validate:"omitempty,max=500"`
}

// UserUpdateProfileRequestFromPb creates a UserUpdateProfileRequest from protobuf.
func UserUpdateProfileRequestFromPb(req *user.UpdateProfileRequest) *UserUpdateProfileRequest {
	r := &UserUpdateProfileRequest{
		UserUid: strings.TrimSpace(req.UserUid),
	}

	if req.FirstName != "" {
		firstName := strings.TrimSpace(req.FirstName)
		r.FirstName = &firstName
	}

	if req.LastName != "" {
		lastName := strings.TrimSpace(req.LastName)
		r.LastName = &lastName
	}

	if req.Bio != "" {
		bio := strings.TrimSpace(req.Bio)
		r.Bio = &bio
	}

	return r
}

// UserUpdatePinRequest represents validated PIN update request.
type UserUpdatePinRequest struct {
	UserUid string `validate:"required"`
	PIN     string `validate:"required,pin"`
}

// UserUpdatePinRequestFromPb creates a UserUpdatePinRequest from protobuf.
func UserUpdatePinRequestFromPb(req *user.UpdatePinRequest) *UserUpdatePinRequest {
	return &UserUpdatePinRequest{
		UserUid: strings.TrimSpace(req.UserUid),
		PIN:     strings.TrimSpace(req.Pin),
	}
}

type UserFilterDeviceRequest struct {
	DeviceUids []string
	DeviceName *string
	Revoked    *bool
}

func (r *UserFilterDeviceRequest) ToUserDeviceListFilterParams() *params.UserDeviceListFilterParam {
	return &params.UserDeviceListFilterParam{
		DeviceUids: r.DeviceUids,
		DeviceName: r.DeviceName,
		Revoked:    r.Revoked,
	}
}

func UserFilterDeviceRequestFromPb(req *user.FilterDeviceRequest) *UserFilterDeviceRequest {
	payload := &UserFilterDeviceRequest{
		DeviceUids: req.GetDeviceUids(),
	}

	if req.DeviceName != nil {
		deviceName := strings.TrimSpace(req.GetDeviceName())
		if deviceName != "" {
			payload.DeviceName = &deviceName
		}
	}

	if req.Revoked != nil {
		payload.Revoked = req.Revoked
	}

	return payload
}

// UserListDevicesRequest represents validated list devices request.
type UserListDevicesRequest struct {
	UserUid    string `validate:"required"`
	Pagination *PaginationRequest
	Filter     *UserFilterDeviceRequest
}

func (r *UserListDevicesRequest) ToUserDeviceListParam() *params.UserDeviceListParam {
	pagination := r.Pagination.ToPaginationParams()
	filter := r.Filter.ToUserDeviceListFilterParams()
	filter.UserUids = []string{r.UserUid}
	return &params.UserDeviceListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// UserListDevicesRequestFromPb creates a UserListDevicesRequest from protobuf.
func UserListDevicesRequestFromPb(req *user.ListDevicesRequest) *UserListDevicesRequest {
	payload := &UserListDevicesRequest{
		UserUid: strings.TrimSpace(req.UserUid),
	}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = UserFilterDeviceRequestFromPb(req.GetFilter())
	}

	return payload
}

// UserRevokeDeviceRequest represents validated device revocation request.
type UserRevokeDeviceRequest struct {
	UserUid   string `validate:"required"`
	DeviceUid string `validate:"required"`
}

// UserRevokeDeviceRequestFromPb creates a UserRevokeDeviceRequest from protobuf.
func UserRevokeDeviceRequestFromPb(req *user.RevokeDeviceRequest) *UserRevokeDeviceRequest {
	return &UserRevokeDeviceRequest{
		UserUid:   strings.TrimSpace(req.UserUid),
		DeviceUid: strings.TrimSpace(req.DeviceUid),
	}
}

// UserChangePasswordRequest represents validated password change request.
type UserChangePasswordRequest struct {
	Uid             string `validate:"required"`
	CurrentPassword string `validate:"required,min=8,max=128"`
	NewPassword     string `validate:"required,min=8,max=128"`
	ConfirmPassword string `validate:"required,min=8,max=128,eqfield=NewPassword"`
}

// UserChangePasswordRequestFromPb creates a UserChangePasswordRequest from protobuf.
func UserChangePasswordRequestFromPb(req *user.ChangePasswordRequest) *UserChangePasswordRequest {
	return &UserChangePasswordRequest{
		Uid:             strings.TrimSpace(req.Uid),
		CurrentPassword: strings.TrimSpace(req.CurrentPassword),
		NewPassword:     strings.TrimSpace(req.NewPassword),
		ConfirmPassword: strings.TrimSpace(req.ConfirmPassword),
	}
}

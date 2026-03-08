package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	common "github.com/adityakw90/service-user-proto/gen/go/common"
	user "github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	portsvc "github.com/adityakw90/service-user/internal/core/port/service"
)

// UserHandler implements the gRPC UserService.
type UserHandler struct {
	user.UnimplementedUserServiceServer
	service   portsvc.UserService
	validator *validator.Validator
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service portsvc.UserService, v *validator.Validator) *UserHandler {
	return &UserHandler{
		service:   service,
		validator: v,
	}
}

// Get retrieves a single user.
func (h *UserHandler) Get(ctx context.Context, req *user.GetRequest) (*user.User, error) {
	r := request.UserGetRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	u, err := h.service.Get(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	return response.ToProtoUser(u), nil
}

// List retrieves a list of users.
func (h *UserHandler) List(ctx context.Context, req *user.ListRequest) (*user.ListResponse, error) {
	r := request.UserListRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	p := r.ToUserListParams()

	result, err := h.service.List(ctx, p.Pagination, p.Filter)
	if err != nil {
		return nil, err
	}

	items := make([]*user.User, len(result.Items))
	for i, u := range result.Items {
		items[i] = response.ToProtoUser(&u)
	}

	meta := &common.Meta{
		Total: int64(len(result.Items)),
		Limit: int32(*p.Pagination.Limit),
	}

	return &user.ListResponse{
		Items: items,
		Meta:  meta,
	}, nil
}

// Add creates a new user.
func (h *UserHandler) Add(ctx context.Context, req *user.AddRequest) (*user.AddResponse, error) {
	r := request.UserAddRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	u, err := h.service.Create(ctx, &param.UserCreateParam{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &user.AddResponse{Uid: u.UID}, nil
}

// Update updates a user.
func (h *UserHandler) Update(ctx context.Context, req *user.UpdateRequest) (*common.Success, error) {
	r := request.UserUpdateRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	param := &param.UserUpdateParam{}
	if req.Username != nil {
		param.Username = req.Username
	}
	if req.Email != nil {
		param.Email = req.Email
	}
	if req.Password != nil {
		param.Password = req.Password
	}
	if req.Status != nil {
		status := model.UserStatus(*req.Status)
		param.Status = &status
	}

	if err := h.service.Update(ctx, req.Uid, param); err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

// Delete performs a soft delete on a user.
func (h *UserHandler) Delete(ctx context.Context, req *user.DeleteRequest) (*common.Success, error) {
	r := request.UserDeleteRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	if err := h.service.Delete(ctx, req.Uid); err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

// GetProfile retrieves a user's profile.
func (h *UserHandler) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.Profile, error) {
	r := request.UserGetProfileRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	p, err := h.service.GetProfile(ctx, req.UserUid)
	if err != nil {
		return nil, err
	}

	return response.ToProtoProfile(p), nil
}

// UpdateProfile updates a user's profile.
func (h *UserHandler) UpdateProfile(ctx context.Context, req *user.UpdateProfileRequest) (*common.Success, error) {
	r := request.UserUpdateProfileRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	opts := param.UserProfileUpdateParam{}
	if req.FirstName != "" {
		opts.FirstName = &req.FirstName
	}
	if req.LastName != "" {
		opts.LastName = &req.LastName
	}
	if req.Bio != "" {
		opts.Bio = &req.Bio
	}
	if len(req.Avatar) > 0 {
		opts.Avatar = req.Avatar
	}

	if err := h.service.UpdateProfile(ctx, req.UserUid, opts); err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

// UpdatePin sets or updates a user's PIN.
func (h *UserHandler) UpdatePin(ctx context.Context, req *user.UpdatePinRequest) (*common.Success, error) {
	r := request.UserUpdatePinRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	if err := h.service.SetPin(ctx, req.UserUid, req.Pin); err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

// ListDevice lists devices for a user.
func (h *UserHandler) ListDevice(ctx context.Context, req *user.ListDevicesRequest) (*user.ListDevicesResponse, error) {
	r := request.UserListDevicesRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	p := r.ToUserDeviceListParam()

	result, err := h.service.ListDevice(ctx, r.UserUid, p.Pagination, p.Filter)
	if err != nil {
		return nil, err
	}

	items := make([]*user.Device, len(result.Items))
	for i, d := range result.Items {
		items[i] = response.ToProtoDevice(&d)
	}

	meta := &common.Meta{
		Total: int64(len(result.Items)),
		Limit: 20,
	}

	return &user.ListDevicesResponse{
		Items: items,
		Meta:  meta,
	}, nil
}

// RevokeDevice revokes access for a device.
func (h *UserHandler) RevokeDevice(ctx context.Context, req *user.RevokeDeviceRequest) (*common.Success, error) {
	r := request.UserRevokeDeviceRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	if err := h.service.RevokeDevice(ctx, req.UserUid, req.DeviceUid); err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

// ChangePassword changes a user's password.
func (h *UserHandler) ChangePassword(ctx context.Context, req *user.ChangePasswordRequest) (*common.Success, error) {
	r := request.UserChangePasswordRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	if req.NewPassword != req.ConfirmPassword {
		return nil, status.Error(codes.InvalidArgument, "new password and confirm password do not match")
	}

	if err := h.service.ChangePassword(ctx, req.Uid, &param.UserChangePasswordParam{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		return nil, err
	}

	return &common.Success{Success: true}, nil
}

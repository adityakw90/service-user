package request

import (
	common "github.com/adityakw90/service-user-proto/gen/go/common"
	device "github.com/adityakw90/service-user-proto/gen/go/device"
	user "github.com/adityakw90/service-user-proto/gen/go/user"
	userFile "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/port"
)

func ToPortPagination(p *common.Pagination) *port.PaginationFilter {
	if p == nil {
		return nil
	}
	return &port.PaginationFilter{
		Limit:  int(p.Limit),
		Offset: int((p.Page - 1) * p.Limit),
	}
}

func ToPortUserFilter(f *user.FilterRequest) *port.UserListFilter {
	if f == nil {
		return nil
	}
	filter := &port.UserListFilter{
		UIDs:   f.Uids,
		Active: f.Active,
	}
	if f.Username != nil {
		filter.Username = f.Username
	}
	if f.Email != nil {
		filter.Email = f.Email
	}
	if f.Query != nil {
		filter.Query = f.Query
	}
	return filter
}

func ToPortDeviceFilter(f *user.FilterDeviceRequest) *port.DeviceListFilter {
	if f == nil {
		return nil
	}
	return &port.DeviceListFilter{
		DeviceUids: f.DeviceUids,
		DeviceName: f.DeviceName,
		Revoked:    f.Revoked,
	}
}

// Conversion functions for port/service interfaces

func ToPaginationParam(p *common.Pagination) *params.PaginationParam {
	if p == nil {
		return nil
	}
	return params.NewPaginationParam(int(p.Page), int(p.Limit), "", "")
}

func ToUserListFilterParam(f *user.FilterRequest) *params.UserListFilterParam {
	if f == nil {
		return nil
	}
	return &params.UserListFilterParam{
		Uids:     f.Uids,
		Username: f.Username,
		Email:    f.Email,
		Query:    f.Query,
	}
}

type UserCreateParam struct {
	Username string
	Email    string
	Password string
}

type UserUpdateParam struct {
	Username *string
	Email    *string
	Password *string
	Status   *int32
}

type UserProfileUpdateParam struct {
	FirstName  *string
	LastName   *string
	Bio        *string
	Avatar     []byte
	Attributes map[string]any
}

func ToUserDeviceListFilterParam(f *user.FilterDeviceRequest) params.UserDeviceListFilterParam {
	if f == nil {
		return params.UserDeviceListFilterParam{}
	}
	return params.UserDeviceListFilterParam{
		DeviceUids: f.DeviceUids,
		Revoked:    f.Revoked,
	}
}

func ToDeviceListFilterParam(f *device.FilterRequest) *params.DeviceListFilterParam {
	if f == nil {
		return &params.DeviceListFilterParam{}
	}
	return &params.DeviceListFilterParam{
		Uids:       f.Uids,
		DeviceName: f.DeviceFingerprint,
	}
}

func ToUserFileListFilterParam(f *userFile.FilterRequest) *params.UserFileListFilterParam {
	if f == nil {
		return &params.UserFileListFilterParam{}
	}
	var visibility *string
	if f.Public != nil {
		v := "public"
		if !*f.Public {
			v = "private"
		}
		visibility = &v
	}
	return &params.UserFileListFilterParam{
		Uids:       f.Uids,
		UserUid:    firstOrNil(f.UserUid),
		Visibility: visibility,
	}
}

func firstOrNil(userUids []string) *string {
	if len(userUids) == 0 {
		return nil
	}
	return &userUids[0]
}

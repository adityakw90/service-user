package param

type UserDeviceListFilterParam struct {
	UserUids   []string // user uids
	DeviceUids []string // device uids
	DeviceName *string
	IpAddress  *string
	Revoked    *bool
}

type UserDeviceListParam struct {
	Pagination *PaginationParam
	Filter     *UserDeviceListFilterParam
}

// UserDeviceOrderBy represents allowed OrderBy column values for UserDevice.
type UserDeviceOrderBy string

const (
	OrderByUserDeviceID           UserDeviceOrderBy = "id"
	OrderByUserDeviceUserID       UserDeviceOrderBy = "user_id"
	OrderByUserDeviceDeviceID     UserDeviceOrderBy = "device_id"
	OrderByUserDeviceLastActiveAt UserDeviceOrderBy = "last_active_at"
	OrderByUserDeviceCreatedAt    UserDeviceOrderBy = "created_at"
)

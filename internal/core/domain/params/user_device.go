package params

type UserDeviceListFilterParam struct {
	UserUids   []string // user uids
	DeviceUids []string // device uids
	IpAddress  *string
	Revoked    *bool
}

type UserDeviceListParam struct {
	Pagination *PaginationParam
	Filter     *UserDeviceListFilterParam
}

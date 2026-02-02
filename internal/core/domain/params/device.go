package params

type DeviceListFilterParam struct {
	Uids       []string // device uids
	DeviceName *string
}

type DeviceListParam struct {
	Pagination *PaginationParam
	Filter     *DeviceListFilterParam
}

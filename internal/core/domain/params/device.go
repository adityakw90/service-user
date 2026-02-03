package params

type DeviceListFilterParam struct {
	Ids        []int64
	Uids       []string // device uids
	DeviceName *string
}

type DeviceListParam struct {
	Pagination *PaginationParam
	Filter     *DeviceListFilterParam
}

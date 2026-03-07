package param

type DeviceListFilterParam struct {
	Ids               []int64
	Uids              []string // device uids
	DeviceName        *string
	DeviceFingerprint *string
}

type DeviceListParam struct {
	Pagination *PaginationParam
	Filter     *DeviceListFilterParam
}

// DeviceOrderBy represents allowed OrderBy column values for Device.
type DeviceOrderBy string

const (
	OrderByDeviceID          DeviceOrderBy = "id"
	OrderByDeviceUID         DeviceOrderBy = "uid"
	OrderByDeviceFingerprint DeviceOrderBy = "device_fingerprint"
	OrderByDeviceName        DeviceOrderBy = "device_name"
	OrderByDeviceCreatedAt   DeviceOrderBy = "created_at"
)

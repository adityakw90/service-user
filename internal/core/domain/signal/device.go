package signal

// DeviceSignal represents data for device service operations.
type DeviceSignal struct {
	UID     *string // Device UID
	UserUID *string // User owning device (for user-scoped operations)

	// Device data
	DeviceName  *string
	Fingerprint *string
	IPAddress   *string

	// Operation context
	Operation string // "get", "list", "delete"
}

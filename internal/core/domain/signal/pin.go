package signal

// PinSignal represents data for PIN service operations.
type PinSignal struct {
	UserUID   string // Always set for PIN operations
	Operation string // "set", "verify", "change", "reset"
	Success   *bool  // For verify operations
}

package security

// UIDGenerator is a port for generating unique IDs.
type UIDGenerator interface {
	New() string
}

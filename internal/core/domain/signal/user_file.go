package signal

// UserFileSignal represents data for user file service operations.
type UserFileSignal struct {
	UID     *string // File UID
	UserUID *string // Owner user UID

	// File data
	FileName *string
	FileType *string
	FileSize *int64
	Category *string

	// Operation context
	Operation string // "get", "list", "add", "update", "delete"
}

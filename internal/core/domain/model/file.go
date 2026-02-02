package model

import "time"

// Visibility constants for user files.
const (
	FileVisibilityPublic  = "public"
	FileVisibilityPrivate = "private"
)

// UserFile represents a file uploaded by a user.
type UserFile struct {
	ID         int64
	UID        string
	UserID     int64
	FileType   string
	FileName   string
	FilePath   string
	MimeType   string
	SizeBytes  int64
	Visibility string
	CreatedAt  time.Time
}

// IsPublic returns true if the file is publicly visible.
func (f *UserFile) IsPublic() bool {
	return f.Visibility == FileVisibilityPublic
}

// IsPrivate returns true if the file is privately visible.
func (f *UserFile) IsPrivate() bool {
	return f.Visibility == FileVisibilityPrivate
}

// Files contains the list of files and metadata for pagination.
type Files struct {
	Items []UserFile
	Meta  Meta
}

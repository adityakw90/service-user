package domain

import (
	"testing"
	"time"
)

func TestUserFile_IsPublic(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		want       bool
	}{
		{"public visibility", FileVisibilityPublic, true},
		{"private visibility", FileVisibilityPrivate, false},
		{"unknown visibility", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &UserFile{Visibility: tt.visibility}
			if got := f.IsPublic(); got != tt.want {
				t.Errorf("UserFile.IsPublic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserFile_IsPrivate(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		want       bool
	}{
		{"public visibility", FileVisibilityPublic, false},
		{"private visibility", FileVisibilityPrivate, true},
		{"unknown visibility", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &UserFile{Visibility: tt.visibility}
			if got := f.IsPrivate(); got != tt.want {
				t.Errorf("UserFile.IsPrivate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserFile_VisibilityConstants(t *testing.T) {
	tests := []struct {
		name  string
		got   string
		want  string
	}{
		{"FileVisibilityPublic is 'public'", FileVisibilityPublic, "public"},
		{"FileVisibilityPrivate is 'private'", FileVisibilityPrivate, "private"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("FileVisibility = %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestUserFile_Fields(t *testing.T) {
	tests := []struct {
		name      string
		id        int64
		uid       string
		userID    int64
		fileType  string
		fileName  string
		filePath  string
		mimeType  string
		sizeBytes int64
		checkField string
	}{
		{"ID is set", 1, "test-file-uid", 2, "avatar", "test.png", "/uploads/test.png", "image/png", 1024, "ID"},
		{"UID is set", 1, "test-file-uid", 2, "avatar", "test.png", "/uploads/test.png", "image/png", 1024, "UID"},
		{"UserID is set", 1, "test-file-uid", 2, "avatar", "test.png", "/uploads/test.png", "image/png", 1024, "UserID"},
		{"FileType is set", 1, "test-file-uid", 2, "avatar", "test.png", "/uploads/test.png", "image/png", 1024, "FileType"},
		{"SizeBytes is set", 1, "test-file-uid", 2, "avatar", "test.png", "/uploads/test.png", "image/png", 1024, "SizeBytes"},
		{"CreatedAt is set", 1, "test-file-uid", 2, "avatar", "test.png", "/uploads/test.png", "image/png", 1024, "CreatedAt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			f := &UserFile{
				ID:         tt.id,
				UID:        tt.uid,
				UserID:     tt.userID,
				FileType:   tt.fileType,
				FileName:   tt.fileName,
				FilePath:   tt.filePath,
				MimeType:   tt.mimeType,
				SizeBytes:  tt.sizeBytes,
				Visibility: FileVisibilityPrivate,
				CreatedAt:  now,
			}

			switch tt.checkField {
			case "ID":
				if f.ID != tt.id {
					t.Errorf("UserFile.ID = %v, want %v", f.ID, tt.id)
				}
			case "UID":
				if f.UID != tt.uid {
					t.Errorf("UserFile.UID = %v, want %v", f.UID, tt.uid)
				}
			case "UserID":
				if f.UserID != tt.userID {
					t.Errorf("UserFile.UserID = %v, want %v", f.UserID, tt.userID)
				}
			case "FileType":
				if f.FileType != tt.fileType {
					t.Errorf("UserFile.FileType = %v, want %v", f.FileType, tt.fileType)
				}
			case "SizeBytes":
				if f.SizeBytes != tt.sizeBytes {
					t.Errorf("UserFile.SizeBytes = %v, want %v", f.SizeBytes, tt.sizeBytes)
				}
			case "CreatedAt":
				if f.CreatedAt.IsZero() {
					t.Error("UserFile.CreatedAt is zero")
				}
			}
		})
	}
}

func TestUserFile_ZeroValue(t *testing.T) {
	tests := []struct {
		name       string
		got        interface{}
		want       interface{}
		fieldCheck string
	}{
		{"zero ID", int64(0), int64(0), "ID"},
		{"zero UID", "", "", "UID"},
		{"zero SizeBytes", int64(0), int64(0), "SizeBytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f UserFile
			switch tt.fieldCheck {
			case "ID":
				if f.ID != tt.want.(int64) {
					t.Errorf("zero UserFile.ID = %v, want %v", f.ID, tt.want)
				}
			case "UID":
				if f.UID != tt.want.(string) {
					t.Errorf("zero UserFile.UID = %v, want %v", f.UID, tt.want)
				}
			case "SizeBytes":
				if f.SizeBytes != tt.want.(int64) {
					t.Errorf("zero UserFile.SizeBytes = %v, want %v", f.SizeBytes, tt.want)
				}
			}
		})
	}
}

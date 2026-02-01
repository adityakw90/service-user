package domain

import (
	"testing"
	"time"
)

func TestUserFile_IsPublic(t *testing.T) {
	tests := []struct {
		name      string
		visibility string
		want      bool
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
		name      string
		visibility string
		want      bool
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
	if FileVisibilityPublic != "public" {
		t.Errorf("FileVisibilityPublic = %v, want 'public'", FileVisibilityPublic)
	}
	if FileVisibilityPrivate != "private" {
		t.Errorf("FileVisibilityPrivate = %v, want 'private'", FileVisibilityPrivate)
	}
}

func TestUserFile_Fields(t *testing.T) {
	now := time.Now().UTC()
	f := &UserFile{
		ID:         1,
		UID:        "test-file-uid",
		UserID:     2,
		FileType:   "avatar",
		FileName:   "test.png",
		FilePath:   "/uploads/test.png",
		MimeType:   "image/png",
		SizeBytes:  1024,
		Visibility: FileVisibilityPrivate,
		CreatedAt:  now,
	}

	if f.ID != 1 {
		t.Errorf("UserFile.ID = %v, want 1", f.ID)
	}
	if f.UID != "test-file-uid" {
		t.Errorf("UserFile.UID = %v, want 'test-file-uid'", f.UID)
	}
	if f.UserID != 2 {
		t.Errorf("UserFile.UserID = %v, want 2", f.UserID)
	}
	if f.FileType != "avatar" {
		t.Errorf("UserFile.FileType = %v, want 'avatar'", f.FileType)
	}
	if f.SizeBytes != 1024 {
		t.Errorf("UserFile.SizeBytes = %v, want 1024", f.SizeBytes)
	}
	if f.CreatedAt.IsZero() {
		t.Error("UserFile.CreatedAt is zero")
	}
}

func TestUserFile_ZeroValue(t *testing.T) {
	var f UserFile
	if f.ID != 0 {
		t.Errorf("zero UserFile.ID = %v, want 0", f.ID)
	}
	if f.UID != "" {
		t.Errorf("zero UserFile.UID = %v, want ''", f.UID)
	}
	if f.SizeBytes != 0 {
		t.Errorf("zero UserFile.SizeBytes = %v, want 0", f.SizeBytes)
	}
}

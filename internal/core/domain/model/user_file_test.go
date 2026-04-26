package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/* test completed by jojo */

func TestCore_Domain_UserFile(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func() *UserFile
		checkFn func(*UserFile)
	}{
		{
			name: "set correctly",
			setupFn: func() *UserFile {
				return &UserFile{
					ID:         1,
					UID:        "test-uid",
					UserID:     2,
					FileType:   "avatar",
					FileName:   "test.png",
					FilePath:   "/uploads/test.png",
					MimeType:   "image/png",
					SizeBytes:  1024,
					Visibility: FileVisibilityPrivate,
					CreatedAt:  time.Now().UTC(),
				}
			},
			checkFn: func(d *UserFile) {
				assert.Equal(t, int64(1), d.ID)
				assert.Equal(t, "test-uid", d.UID)
				assert.Equal(t, "avatar", d.FileType)
				assert.Equal(t, "test.png", d.FileName)
				assert.Equal(t, "/uploads/test.png", d.FilePath)
				assert.Equal(t, "image/png", d.MimeType)
				assert.Equal(t, int64(1024), d.SizeBytes)
				assert.Equal(t, FileVisibilityPrivate, d.Visibility)
				assert.NotZero(t, d.CreatedAt)
			},
		},
		{
			name: "Zero Value",
			setupFn: func() *UserFile {
				var d UserFile
				return &d
			},
			checkFn: func(d *UserFile) {
				assert.Equal(t, int64(0), d.ID)
				assert.Equal(t, "", d.UID)
				assert.Equal(t, int64(0), d.UserID)
				assert.Equal(t, "", d.FileType)
				assert.Equal(t, "", d.FileName)
				assert.Equal(t, "", d.FilePath)
				assert.Equal(t, "", d.MimeType)
				assert.Equal(t, int64(0), d.SizeBytes)
				assert.Equal(t, "", d.Visibility)
				assert.Zero(t, d.CreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := tt.setupFn()
			tt.checkFn(device)
		})
	}
}

func TestCore_Domain_UserFile_IsPublic(t *testing.T) {
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
			assert.Equal(t, tt.want, f.IsPublic())
		})
	}
}

func TestCore_Domain_UserFile_IsPrivate(t *testing.T) {
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
			assert.Equal(t, tt.want, f.IsPrivate())
		})
	}
}

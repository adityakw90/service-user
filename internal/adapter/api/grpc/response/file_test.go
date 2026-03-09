package response

import (
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/google/uuid"
	userFile "github.com/adityakw90/service-user-proto/gen/go/user_file"
)

func TestToProtoUserFile(t *testing.T) {
	now := time.Now()
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name  string
		input *model.UserFile
		want  *userFile.UserFile
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid user file",
			input: &model.UserFile{
				UID:        uid.String(),
				UserUID:    uuid.MustParse("00000000-0000-0000-0000-000000000002").String(),
				FileType:   "avatar",
				FileName:   "profile.jpg",
				FilePath:   "/uploads/avatar/profile.jpg",
				MimeType:   "image/jpeg",
				SizeBytes:  102400,
				Visibility: "public",
				CreatedAt:  now,
			},
			want: &userFile.UserFile{
				Uid:        "00000000-0000-0000-0000-000000000001",
				UserUid:    "00000000-0000-0000-0000-000000000002",
				FileType:   "avatar",
				FileName:   "profile.jpg",
				FilePath:   "/uploads/avatar/profile.jpg",
				MimeType:   "image/jpeg",
				SizeBytes:  102400,
				Visibility: "public",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoUserFile(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToProtoUserFile() = %v, want nil", got)
				}
				return
			}

			if got.Uid != tt.want.Uid {
				t.Errorf("ToProtoUserFile().Uid = %v, want %v", got.Uid, tt.want.Uid)
			}
			if got.FileName != tt.want.FileName {
				t.Errorf("ToProtoUserFile().FileName = %v, want %v", got.FileName, tt.want.FileName)
			}
			if got.SizeBytes != tt.want.SizeBytes {
				t.Errorf("ToProtoUserFile().SizeBytes = %v, want %v", got.SizeBytes, tt.want.SizeBytes)
			}
		})
	}
}

func TestToProtoFileList(t *testing.T) {
	uid := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name    string
		files   *model.UserFiles
		meta    *model.Meta
		wantLen int
	}{
		{
			name:  "Nil files",
			files: nil,
			meta:  &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
			wantLen: 0,
		},
		{
			name: "Empty list",
			files: &model.UserFiles{
				Items: []model.UserFile{},
			},
			meta:    &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
			wantLen: 0,
		},
		{
			name: "List with files",
			files: &model.UserFiles{
				Items: []model.UserFile{
					{UID: uid.String(), FileName: "file1.jpg", SizeBytes: 1024},
					{UID: uid.String(), FileName: "file2.jpg", SizeBytes: 2048},
				},
			},
			meta:    &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoFileList(tt.files, tt.meta)

			if len(got.Items) != tt.wantLen {
				t.Errorf("ToProtoFileList() len = %d, want %d", len(got.Items), tt.wantLen)
			}
		})
	}
}

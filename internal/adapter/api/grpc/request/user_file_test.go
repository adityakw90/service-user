package request

import (
	"testing"

	commonpb "github.com/adityakw90/service-user-proto/gen/go/common"
	userfilepb "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFileGetRequestFromPb(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantUid string
	}{
		{name: "trims surrounding whitespace", raw: "  some-file-uid  ", wantUid: "some-file-uid"},
		{name: "already trimmed value unchanged", raw: "file-uid-abc", wantUid: "file-uid-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFileGetRequestFromPb(&userfilepb.GetRequest{Uid: tt.raw})
			assert.Equal(t, tt.wantUid, got.Uid)
		})
	}
}

func TestUserFileDeleteRequestFromPb(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantUid string
	}{
		{name: "trims surrounding whitespace", raw: "  deleted-file-uid  ", wantUid: "deleted-file-uid"},
		{name: "already trimmed value unchanged", raw: "file-uid-xyz", wantUid: "file-uid-xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFileDeleteRequestFromPb(&userfilepb.DeleteRequest{Uid: tt.raw})
			assert.Equal(t, tt.wantUid, got.Uid)
		})
	}
}

func TestUserFileAddRequestFromPb(t *testing.T) {
	tests := []struct {
		name         string
		req          *userfilepb.AddRequest
		wantUserUid  string
		wantName     string
		wantFilename string
		wantFileData []byte
	}{
		{
			name: "trims all string fields and copies file data",
			req: &userfilepb.AddRequest{
				UserUid:  "  user-uid-123  ",
				Name:     "  My Image  ",
				Filename: "  photo.png  ",
				Filedata: []byte{1, 2, 3},
			},
			wantUserUid:  "user-uid-123",
			wantName:     "My Image",
			wantFilename: "photo.png",
			wantFileData: []byte{1, 2, 3},
		},
		{
			name: "already trimmed strings unchanged",
			req: &userfilepb.AddRequest{
				UserUid:  "user-uid-abc",
				Name:     "Image",
				Filename: "photo.jpg",
				Filedata: []byte{9, 9},
			},
			wantUserUid:  "user-uid-abc",
			wantName:     "Image",
			wantFilename: "photo.jpg",
			wantFileData: []byte{9, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFileAddRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUserUid, got.UserUid)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, tt.wantFilename, got.Filename)
			assert.Equal(t, tt.wantFileData, got.FileData)
		})
	}
}

func TestUserFileUpdateRequestFromPb(t *testing.T) {
	name := " Updated Name "
	filename := " photo_v2.png "
	bothSet := &userfilepb.UpdateRequest{
		Uid:      "  file-uid-xyz  ",
		Name:     &name,
		Filename: &filename,
		Filedata: []byte{4, 5},
	}
	onlyName := &userfilepb.UpdateRequest{
		Uid:  "  file-uid-xyz  ",
		Name: &name,
	}
	neither := &userfilepb.UpdateRequest{
		Uid:      "  file-uid-xyz  ",
		Filedata: []byte{0},
	}

	tests := []struct {
		name         string
		req          *userfilepb.UpdateRequest
		wantUid      string
		wantName     *string
		wantFilename *string
		wantFileData []byte
	}{
		{
			name:         "both optional fields populated and trimmed",
			req:          bothSet,
			wantUid:      "file-uid-xyz",
			wantName:     strPtr("Updated Name"),
			wantFilename: strPtr("photo_v2.png"),
			wantFileData: []byte{4, 5},
		},
		{
			name:         "only Name populated",
			req:          onlyName,
			wantUid:      "file-uid-xyz",
			wantName:     strPtr("Updated Name"),
			wantFilename: nil,
			wantFileData: nil,
		},
		{
			name:         "neither Name nor Filename populated",
			req:          neither,
			wantUid:      "file-uid-xyz",
			wantName:     nil,
			wantFilename: nil,
			wantFileData: []byte{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFileUpdateRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUid, got.Uid)
			assertOptionalString(t, tt.wantName, got.Name, "Name")
			assertOptionalString(t, tt.wantFilename, got.Filename, "Filename")
			assert.Equal(t, tt.wantFileData, got.FileData)
		})
	}
}

func TestUserFileFilterRequestFromPb(t *testing.T) {
	visibility := "public"
	filetype := " image/png "
	emptyFiletype := "   "

	tests := []struct {
		name           string
		req            *userfilepb.FilterRequest
		wantUids       []string
		wantUserUid    []string
		wantVisibility *string
		wantFileType   *string
	}{
		{
			name: "all fields populated and trimmed",
			req: &userfilepb.FilterRequest{
				Uids:       []string{"f1", "f2"},
				UserUid:    []string{"u1"},
				Visibility: &visibility,
				Filetype:   &filetype,
			},
			wantUids:       []string{"f1", "f2"},
			wantUserUid:    []string{"u1"},
			wantVisibility: strPtr("public"),
			wantFileType:   strPtr("image/png"),
		},
		{
			name:           "whitespace-only filetype becomes nil",
			req:            &userfilepb.FilterRequest{Filetype: &emptyFiletype},
			wantUids:       nil,
			wantUserUid:    nil,
			wantVisibility: nil,
			wantFileType:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFileFilterRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUids, got.Uids)
			assert.Equal(t, tt.wantUserUid, got.UserUid)
			assertOptionalString(t, tt.wantVisibility, got.Visibility, "Visibility")
			assertOptionalString(t, tt.wantFileType, got.FileType, "FileType")

			// Verify ToUserFileFilterParams forwards fields unchanged.
			params := got.ToUserFileFilterParams()
			require.NotNil(t, params)
			assert.Equal(t, got.Uids, params.Uids)
			assert.Equal(t, got.UserUid, params.UserUid)
			assert.Equal(t, got.Visibility, params.Visibility)
			assert.Equal(t, got.FileType, params.FileType)
		})
	}
}

func TestUserFileListRequestFromPb(t *testing.T) {
	filetype := "text/plain"

	tests := []struct {
		name         string
		req          *userfilepb.ListRequest
		wantPage     int
		wantLimit    int
		wantSort     string
		wantOrderBy  string
		wantFileType *string
	}{
		{
			name: "with pagination and filter",
			req: &userfilepb.ListRequest{
				Pagination: &commonpb.Pagination{
					Page:  5,
					Limit: 50,
					Sort:  "desc",
				},
				Filter: &userfilepb.FilterRequest{
					Filetype: &filetype,
				},
			},
			wantPage:     5,
			wantLimit:    50,
			wantSort:     "desc",
			wantOrderBy:  "",
			wantFileType: strPtr("text/plain"),
		},
		{
			name:         "nil pagination and filter apply defaults",
			req:          &userfilepb.ListRequest{},
			wantPage:     1,
			wantLimit:    10,
			wantSort:     "desc",
			wantOrderBy:  "created_at",
			wantFileType: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UserFileListRequestFromPb(tt.req)
			params := got.ToUserFileListParams()
			require.NotNil(t, params)
			require.NotNil(t, params.Pagination)
			assert.Equal(t, tt.wantPage, *params.Pagination.Page)
			assert.Equal(t, tt.wantLimit, *params.Pagination.Limit)
			assert.Equal(t, tt.wantSort, *params.Pagination.Sort)
			assert.Equal(t, tt.wantOrderBy, *params.Pagination.OrderBy)
			require.NotNil(t, params.Filter)
			assertOptionalString(t, tt.wantFileType, params.Filter.FileType, "FileType")
		})
	}
}

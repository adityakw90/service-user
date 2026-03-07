package request

import (
	"strings"

	userFile "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/pkg/util"
)

// UserFileGetRequest represents validated user file get request.
type UserFileGetRequest struct {
	Uid string `validate:"required"`
}

// UserFileGetRequestFromPb creates a UserFileGetRequest from protobuf.
func UserFileGetRequestFromPb(req *userFile.GetRequest) *UserFileGetRequest {
	return &UserFileGetRequest{
		Uid: strings.TrimSpace(req.Uid),
	}
}

type UserFileFilterRequest struct {
	Uids       []string `validate:"omitempty"`
	UserUid    []string `validate:"omitempty"`
	FileType   *string  `validate:"omitempty"`
	Visibility *string  `validate:"omitempty"`
}

func (r *UserFileFilterRequest) ToUserFileFilterParams() *param.UserFileListFilterParam {
	return &param.UserFileListFilterParam{
		Uids:       r.Uids,
		UserUid:    r.UserUid,
		FileType:   r.FileType,
		Visibility: r.Visibility,
	}
}

func UserFileFilterRequestFromPb(req *userFile.FilterRequest) *UserFileFilterRequest {
	payload := &UserFileFilterRequest{
		Uids:    req.GetUids(),
		UserUid: req.GetUserUid(),
	}

	if req.Visibility != nil {
		field := req.GetVisibility()
		payload.Visibility = &field
	}

	if req.Filetype != nil {
		field := strings.TrimSpace(req.GetFiletype())
		if field != "" {
			payload.FileType = &field
		}
	}

	return payload
}

// UserFileListRequest represents validated list request.
type UserFileListRequest struct {
	Pagination *PaginationRequest
	Filter     *UserFileFilterRequest
}

func (r *UserFileListRequest) ToUserFileListParams() *param.UserFileListParam {
	var pagination *param.PaginationParam
	if r.Pagination != nil {
		pagination = r.Pagination.ToPaginationParams()
	} else {
		// Default pagination
		page := 1
		limit := 10
		pagination = &param.PaginationParam{
			Page:    &page,
			Limit:   &limit,
			Sort:    util.Ptr("desc"),
			OrderBy: util.Ptr("created_at"),
		}
	}

	var filter *param.UserFileListFilterParam
	if r.Filter != nil {
		filter = r.Filter.ToUserFileFilterParams()
	} else {
		filter = &param.UserFileListFilterParam{}
	}

	return &param.UserFileListParam{
		Pagination: pagination,
		Filter:     filter,
	}
}

// UserFileListRequestFromPb creates a UserFileListRequest from protobuf.
func UserFileListRequestFromPb(req *userFile.ListRequest) *UserFileListRequest {
	payload := &UserFileListRequest{}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = UserFileFilterRequestFromPb(req.GetFilter())
	}

	return payload
}

// UserFileAddRequest represents validated user file add request.
type UserFileAddRequest struct {
	UserUid  string `validate:"required"`
	Name     string `validate:"required,min=1,max=255"`
	Filename string `validate:"required,min=1,max=255"`
	FileData []byte `validate:"max=10485760"` // 10MB max
}

// UserFileAddRequestFromPb creates a UserFileAddRequest from protobuf.
func UserFileAddRequestFromPb(req *userFile.AddRequest) *UserFileAddRequest {
	return &UserFileAddRequest{
		UserUid:  strings.TrimSpace(req.UserUid),
		Name:     strings.TrimSpace(req.Name),
		Filename: strings.TrimSpace(req.Filename),
		FileData: req.Filedata,
	}
}

// UserFileUpdateRequest represents validated user file update request.
type UserFileUpdateRequest struct {
	Uid      string  `validate:"required"`
	Name     *string `validate:"omitempty,min=1,max=255"`
	Filename *string `validate:"omitempty,min=1,max=255"`
	FileData []byte  `validate:"max=10485760"` // 10MB max
}

// UserFileUpdateRequestFromPb creates a UserFileUpdateRequest from protobuf.
func UserFileUpdateRequestFromPb(req *userFile.UpdateRequest) *UserFileUpdateRequest {
	r := &UserFileUpdateRequest{
		Uid:      strings.TrimSpace(req.Uid),
		FileData: req.GetFiledata(),
	}

	if req.Name != nil {
		name := strings.TrimSpace(req.GetName())
		r.Name = &name
	}

	if req.Filename != nil {
		filename := strings.TrimSpace(req.GetFilename())
		r.Filename = &filename
	}

	return r
}

// UserFileDeleteRequest represents validated user file delete request.
type UserFileDeleteRequest struct {
	Uid string `validate:"required"`
}

// UserFileDeleteRequestFromPb creates a UserFileDeleteRequest from protobuf.
func UserFileDeleteRequestFromPb(req *userFile.DeleteRequest) *UserFileDeleteRequest {
	return &UserFileDeleteRequest{
		Uid: strings.TrimSpace(req.Uid),
	}
}

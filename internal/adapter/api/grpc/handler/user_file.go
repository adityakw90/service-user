package handler

import (
	"context"
	"encoding/base64"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	common "github.com/adityakw90/service-user-proto/gen/go/common"
	userFile "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	portsvc "github.com/adityakw90/service-user/internal/core/port/service"
)

const (
	// MaxFileSize is the maximum file size in bytes (10MB)
	MaxFileSize = 10 * 1024 * 1024
	// FileType is the default file type for uploaded files
	FileType = "uploaded"
)

// UserFileHandler implements the gRPC UserFileService.
type UserFileHandler struct {
	userFile.UnimplementedUserFileServiceServer
	service   portsvc.UserFileService
	validator *validator.Validator
}

// NewUserFileHandler creates a new UserFileHandler.
func NewUserFileHandler(service portsvc.UserFileService) *UserFileHandler {
	return &UserFileHandler{
		service:   service,
		validator: validator.New(),
	}
}

// Get retrieves a single user file by UID.
func (h *UserFileHandler) Get(ctx context.Context, req *userFile.GetRequest) (*userFile.UserFile, error) {
	r := request.UserFileGetRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	f, err := h.service.Get(ctx, req.Uid)
	if err != nil {
		return nil, response.MapError(err)
	}

	return response.ToProtoUserFile(f), nil
}

// List retrieves a list of user files.
func (h *UserFileHandler) List(ctx context.Context, req *userFile.ListRequest) (*userFile.ListResponse, error) {
	r := request.UserFileListRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	p := r.ToUserFileListParams()

	result, err := h.service.List(ctx, p.Pagination, p.Filter)
	if err != nil {
		return nil, response.MapError(err)
	}

	items := make([]*userFile.UserFile, len(result.Items))
	for i, f := range result.Items {
		items[i] = response.ToProtoUserFile(&f)
	}

	meta := &common.Meta{
		Total: int64(len(result.Items)),
		Limit: int32(*p.Pagination.Limit),
	}

	return &userFile.ListResponse{
		Items: items,
		Meta:  meta,
	}, nil
}

// Add creates a new user file.
func (h *UserFileHandler) Add(ctx context.Context, req *userFile.AddRequest) (*userFile.AddResponse, error) {
	r := request.UserFileAddRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	// Convert filedata to base64 and store as file path
	// In production, this should be uploaded to S3 or similar storage service
	filePath := ""
	if len(req.Filedata) > 0 {
		if len(req.Filedata) > MaxFileSize {
			return nil, status.Error(codes.ResourceExhausted, "file size exceeds maximum limit")
		}
		filePath = "data:" + getMimeType(req.Filename) + ";base64," + base64.StdEncoding.EncodeToString(req.Filedata)
	}

	createParam := params.UserFileCreateParam{
		UserUID:    req.UserUid,
		FileType:   FileType,
		FileName:   req.Name,
		FilePath:   filePath,
		MimeType:   getMimeType(req.Filename),
		SizeBytes:  int64(len(req.Filedata)),
		Visibility: getVisibility(req.GetPublic()),
	}

	f, err := h.service.Add(ctx, createParam)
	if err != nil {
		return nil, response.MapError(err)
	}

	return &userFile.AddResponse{Uid: f.UID}, nil
}

// Update updates an existing user file.
func (h *UserFileHandler) Update(ctx context.Context, req *userFile.UpdateRequest) (*userFile.UpdateResponse, error) {
	r := request.UserFileUpdateRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	updateParam := params.UserFileUpdateParam{}
	if req.Name != nil {
		updateParam.FileName = req.Name
	}
	if req.Filename != nil {
		updateParam.FileName = req.Filename
	}
	if len(req.Filedata) > 0 {
		if len(req.Filedata) > MaxFileSize {
			return nil, status.Error(codes.ResourceExhausted, "file size exceeds maximum limit")
		}
		// Determine filename for MIME type detection
		filename := ""
		if req.Filename != nil {
			filename = *req.Filename
		} else if req.Name != nil {
			filename = *req.Name
		}
		mimeType := getMimeType(filename)
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		filePath := "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(req.Filedata)
		updateParam.FilePath = &filePath
		updateParam.SizeBytes = new(int64)
		*updateParam.SizeBytes = int64(len(req.Filedata))
	}
	if req.Public != nil {
		visibility := getVisibility(*req.Public)
		updateParam.Visibility = &visibility
	}

	if err := h.service.Update(ctx, req.Uid, updateParam); err != nil {
		return nil, response.MapError(err)
	}

	return &userFile.UpdateResponse{Success: true}, nil
}

// Delete deletes a user file by UID.
func (h *UserFileHandler) Delete(ctx context.Context, req *userFile.DeleteRequest) (*userFile.DeleteResponse, error) {
	r := request.UserFileDeleteRequestFromPb(req)
	if err := h.validator.Struct(r); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	if err := h.service.Delete(ctx, req.Uid); err != nil {
		return nil, response.MapError(err)
	}

	return &userFile.DeleteResponse{Success: true}, nil
}

// Helper functions

func getMimeType(filename string) string {
	// Simple MIME type detection based on extension
	// In production, use a proper MIME type library
	extension := ""
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			extension = filename[i+1:]
			break
		}
	}

	switch extension {
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "pdf":
		return "application/pdf"
	case "txt":
		return "text/plain"
	case "json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}

func getVisibility(public bool) string {
	if public {
		return model.FileVisibilityPublic
	}
	return model.FileVisibilityPrivate
}

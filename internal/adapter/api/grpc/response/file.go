package response

import (
	userFile "github.com/adityakw90/service-user-proto/gen/go/user_file"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// ToProtoUserFile converts domain UserFile to proto UserFile.
func ToProtoUserFile(f *model.UserFile) *userFile.UserFile {
	if f == nil {
		return nil
	}
	return &userFile.UserFile{
		Uid:        f.UID,
		UserUid:    f.UserUID,
		FileType:   f.FileType,
		FileName:   f.FileName,
		FilePath:   f.FilePath,
		MimeType:   f.MimeType,
		SizeBytes:  f.SizeBytes,
		Visibility: f.Visibility,
		CreatedAt:  toProtoTimestampPB(f.CreatedAt),
	}
}

// ToProtoFileList converts domain UserFiles to proto ListResponse.
func ToProtoFileList(files *model.UserFiles, meta *model.Meta) *userFile.ListResponse {
	if files == nil {
		return &userFile.ListResponse{Meta: ToProtoMeta(meta)}
	}
	items := make([]*userFile.UserFile, len(files.Items))
	for i, f := range files.Items {
		items[i] = ToProtoUserFile(&f)
	}

	return &userFile.ListResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}

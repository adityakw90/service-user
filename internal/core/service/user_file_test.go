package service

import (
	"context"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserFileService_Get(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserFileRepository, *MockUserResolver)
		uid        string
		want       *model.UserFile
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(fr *MockUserFileRepository, ur *MockUserResolver) {
				fr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.UserFile, error) {
					return createUserFile(1, "file-uid", "user-uid", "image"), nil
				}
				ur.UIDsByIDsFunc = func(ctx context.Context, userIDs []int64) (map[int64]string, error) {
					return map[int64]string{1: "user-uid"}, nil
				}
			},
			uid:  "file-uid",
			want: createUserFile(1, "file-uid", "user-uid", "image"),
		},
		{
			name:    "Error - empty UID",
			uid:     "",
			wantErr: domainerrors.ErrInvalidUID,
		},
		{
			name: "Error - file not found",
			setupMocks: func(fr *MockUserFileRepository, ur *MockUserResolver) {
				fr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.UserFile, error) {
					return nil, domainerrors.ErrFileNotFound
				}
			},
			uid:     "nonexistent-file",
			wantErr: domainerrors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserFileRepo := NewMockUserFileRepository()
			mockUserRepo := NewMockUserRepository()
			mockUserResolver := NewMockUserResolver()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFileRepo, mockUserResolver)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				NewMockUserFileObserver(),
			)

			// Execute
			got, err := svc.Get(context.Background(), tt.uid)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.UID, got.UID)
			assert.Equal(t, tt.want.UserUID, got.UserUID)
			assert.Equal(t, tt.want.FileName, got.FileName)
		})
	}
}

func TestUserFileService_List(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserFileRepository, *MockUserResolver)
		pagination *params.PaginationParam
		filter     *params.UserFileListFilterParam
		wantErr    error
		verifyFunc func(*testing.T, *model.UserFiles)
	}{
		{
			name: "Happy Path - default pagination",
			setupMocks: func(fr *MockUserFileRepository, ur *MockUserResolver) {
				fr.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.UserFileListFilterParam) (*model.UserFiles, error) {
					return &model.UserFiles{
						Items: []model.UserFile{
							*createUserFile(1, "file1", "user-uid", "image"),
							*createUserFile(2, "file2", "user-uid", "document"),
						},
					}, nil
				}
				ur.UIDsByIDsFunc = func(ctx context.Context, userIDs []int64) (map[int64]string, error) {
					return map[int64]string{1: "user-uid"}, nil
				}
			},
			pagination: nil,
			filter:     nil,
			wantErr:    nil,
			verifyFunc: func(t *testing.T, got *model.UserFiles) {
				require.Len(t, got.Items, 2)
				assert.Equal(t, "file1", got.Items[0].UID)
				assert.Equal(t, "user-uid", got.Items[0].UserUID)
				assert.Equal(t, "file2", got.Items[1].UID)
				assert.Equal(t, "user-uid", got.Items[1].UserUID)
			},
		},
		{
			name: "Happy Path - custom pagination",
			setupMocks: func(fr *MockUserFileRepository, ur *MockUserResolver) {
				fr.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.UserFileListFilterParam) (*model.UserFiles, error) {
					return &model.UserFiles{
						Items: []model.UserFile{},
					}, nil
				}
			},
			pagination: &params.PaginationParam{
				Page:    util.Ptr(2),
				Limit:   util.Ptr(20),
				Sort:    util.Ptr("desc"),
				OrderBy: util.Ptr("created_at"),
			},
			filter:  nil,
			wantErr: nil,
			verifyFunc: func(t *testing.T, got *model.UserFiles) {
				require.Len(t, got.Items, 0)
			},
		},
		{
			name: "Happy Path - with filters",
			setupMocks: func(fr *MockUserFileRepository, ur *MockUserResolver) {
				fr.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.UserFileListFilterParam) (*model.UserFiles, error) {
					return &model.UserFiles{Items: []model.UserFile{}}, nil
				}
			},
			pagination: &params.PaginationParam{
				Page:    util.Ptr(1),
				Limit:   util.Ptr(10),
				Sort:    util.Ptr("asc"),
				OrderBy: util.Ptr("created_at"),
			},
			filter: &params.UserFileListFilterParam{
				UserUid: []string{"user-123"},
			},
			wantErr: nil,
			verifyFunc: func(t *testing.T, got *model.UserFiles) {
				require.Len(t, got.Items, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserFileRepo := NewMockUserFileRepository()
			mockUserRepo := NewMockUserRepository()
			mockUserResolver := NewMockUserResolver()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFileRepo, mockUserResolver)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				NewMockUserFileObserver(),
			)

			// Execute
			got, err := svc.List(context.Background(), tt.pagination, tt.filter)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			if tt.verifyFunc != nil {
				tt.verifyFunc(t, got)
			}
		})
	}
}

func TestUserFileService_Add(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockUserFileRepository, *MockUIDGenerator)
		param      params.UserFileCreateParam
		want       *model.UserFile
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *MockUserRepository, fr *MockUserFileRepository, ug *MockUIDGenerator) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "user-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				ug.NewFunc = func() string { return "file-uid" }
				fr.CreateFunc = func(ctx context.Context, file *model.UserFile) (*model.UserFile, error) {
					file.ID = 1
					return file, nil
				}
			},
			param: params.UserFileCreateParam{
				UserUID:    "user-uid",
				FileType:   "image",
				FileName:   "photo.jpg",
				FilePath:   "/uploads/photo.jpg",
				MimeType:   "image/jpeg",
				SizeBytes:  1024,
				Visibility: model.FileVisibilityPrivate,
			},
			want: &model.UserFile{
				ID:         1,
				UID:        "file-uid",
				UserID:     1,
				UserUID:    "user-uid",
				FileType:   "image",
				FileName:   "photo.jpg",
				FilePath:   "/uploads/photo.jpg",
				MimeType:   "image/jpeg",
				SizeBytes:  1024,
				Visibility: model.FileVisibilityPrivate,
			},
		},
		{
			name: "Error - empty user UID",
			param: params.UserFileCreateParam{
				UserUID:    "",
				FileType:   "image",
				FileName:   "photo.jpg",
				FilePath:   "/uploads/photo.jpg",
				MimeType:   "image/jpeg",
				SizeBytes:  1024,
				Visibility: model.FileVisibilityPrivate,
			},
			wantErr: domainerrors.ErrInvalidUID,
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository, fr *MockUserFileRepository, ug *MockUIDGenerator) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			param: params.UserFileCreateParam{
				UserUID:    "nonexistent-uid",
				FileType:   "image",
				FileName:   "photo.jpg",
				FilePath:   "/uploads/photo.jpg",
				MimeType:   "image/jpeg",
				SizeBytes:  1024,
				Visibility: model.FileVisibilityPrivate,
			},
			wantErr: domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserFileRepo := NewMockUserFileRepository()
			mockUserRepo := NewMockUserRepository()
			mockUserResolver := NewMockUserResolver()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockUserFileRepo, mockUIDGen)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				NewMockUserFileObserver(),
			)

			// Execute
			got, err := svc.Add(context.Background(), tt.param)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.FileName, got.FileName)
		})
	}
}

func TestUserFileService_Update(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserFileRepository)
		uid        string
		param      params.UserFileUpdateParam
		wantErr    error
	}{
		{
			name: "Happy Path - update all fields",
			setupMocks: func(fr *MockUserFileRepository) {
				fr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.UserFile, error) {
					return createUserFile(1, "file-uid", "user-uid", "image"), nil
				}
				fr.UpdateFunc = func(ctx context.Context, file *model.UserFile) error {
					return nil
				}
			},
			uid: "file-uid",
			param: params.UserFileUpdateParam{
				FileName:   util.Ptr("newname.jpg"),
				FilePath:   util.Ptr("/uploads/newname.jpg"),
				MimeType:   util.Ptr("image/png"),
				SizeBytes:  util.Ptr(int64(2048)),
				Visibility: util.Ptr(model.FileVisibilityPublic),
			},
		},
		{
			name: "Happy Path - partial update",
			setupMocks: func(fr *MockUserFileRepository) {
				fr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.UserFile, error) {
					return createUserFile(1, "file-uid", "user-uid", "image"), nil
				}
				fr.UpdateFunc = func(ctx context.Context, file *model.UserFile) error {
					return nil
				}
			},
			uid: "file-uid",
			param: params.UserFileUpdateParam{
				FileName: util.Ptr("updated.jpg"),
			},
		},
		{
			name: "Error - file not found",
			setupMocks: func(fr *MockUserFileRepository) {
				fr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.UserFile, error) {
					return nil, domainerrors.ErrFileNotFound
				}
			},
			uid: "nonexistent-file",
			param: params.UserFileUpdateParam{
				FileName: util.Ptr("newname.jpg"),
			},
			wantErr: domainerrors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserFileRepo := NewMockUserFileRepository()
			mockUserRepo := NewMockUserRepository()
			mockUserResolver := NewMockUserResolver()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFileRepo)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				NewMockUserFileObserver(),
			)

			// Execute
			err := svc.Update(context.Background(), tt.uid, tt.param)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserFileService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserFileRepository)
		uid        string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(fr *MockUserFileRepository) {
				fr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.UserFile, error) {
					return createUserFile(1, "file-uid", "user-uid", "image"), nil
				}
				fr.DeleteFunc = func(ctx context.Context, file *model.UserFile) error {
					return nil
				}
			},
			uid: "file-uid",
		},
		{
			name: "Error - file not found",
			setupMocks: func(fr *MockUserFileRepository) {
				fr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.UserFile, error) {
					return nil, domainerrors.ErrFileNotFound
				}
			},
			uid:     "nonexistent-file",
			wantErr: domainerrors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserFileRepo := NewMockUserFileRepository()
			mockUserRepo := NewMockUserRepository()
			mockUserResolver := NewMockUserResolver()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFileRepo)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				NewMockUserFileObserver(),
			)

			// Execute
			err := svc.Delete(context.Background(), tt.uid)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

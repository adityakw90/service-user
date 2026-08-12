package service

import (
	"context"
	"errors"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	eventmocks "github.com/adityakw90/service-user/mocks/event"
	observermocks "github.com/adityakw90/service-user/mocks/observer"
	repomocks "github.com/adityakw90/service-user/mocks/repository"
	resolvermocks "github.com/adityakw90/service-user/mocks/resolver"
	securitymocks "github.com/adityakw90/service-user/mocks/security"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupObserverAny allows any OnSignal calls on the observer (useful when not testing signal behavior)
func setupUserFileObserverAny(t *testing.T, observer *observermocks.MockServiceObserver[signal.UserFileSignal]) {
	// Allow any OnSignal call without checking parameters
	// Use Maybe() to make the expectation optional (can be called 0 or more times)
	// Note: Using EXPECT().OnSignal() pattern for better type safety
	observer.EXPECT().OnSignal(mock.Anything, mock.Anything, mock.AnythingOfType("signal.UserFileSignal"), mock.Anything).Maybe()
}

func TestUserFileService_Get(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*repomocks.MockUserFileRepository, *resolvermocks.MockUserResolver)
		uid        string
		want       *model.UserFile
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ur *resolvermocks.MockUserResolver) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				ur.EXPECT().UIDsByIDs(mock.Anything, []int64{1}).Return(map[int64]string{1: "user-uid"}, nil).Once()
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
			setupMocks: func(fr *repomocks.MockUserFileRepository, ur *resolvermocks.MockUserResolver) {
				fr.EXPECT().GetByUID(mock.Anything, "nonexistent-file").Return(nil, domainerrors.ErrFileNotFound).Once()
			},
			uid:     "nonexistent-file",
			wantErr: domainerrors.ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserFileRepo := repomocks.NewMockUserFileRepository(t)
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockUserResolver := resolvermocks.NewMockUserResolver(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)

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
				func() *observermocks.MockServiceObserver[signal.UserFileSignal] {
					obs := observermocks.NewMockServiceObserver[signal.UserFileSignal](t)
					setupUserFileObserverAny(t, obs)
					return obs
				}(),
				eventmocks.NewMockEventPublisher(t),
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
		setupMocks func(*repomocks.MockUserFileRepository, *resolvermocks.MockUserResolver)
		pagination *param.PaginationParam
		filter     *param.UserFileListFilterParam
		wantErr    error
		verifyFunc func(*testing.T, *model.UserFiles)
	}{
		{
			name: "Happy Path - default pagination",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ur *resolvermocks.MockUserResolver) {
				fr.EXPECT().List(mock.Anything, mock.AnythingOfType("*param.PaginationParam"), mock.AnythingOfType("*param.UserFileListFilterParam")).Return(&model.UserFiles{
					Items: []model.UserFile{
						*createUserFile(1, "file1", "user-uid", "image"),
						*createUserFile(2, "file2", "user-uid", "document"),
					},
				}, nil).Once()
				// Both files have UserID 1, so UIDsByIDs will be called with [1, 1]
				ur.EXPECT().UIDsByIDs(mock.Anything, mock.MatchedBy(func(ids []int64) bool {
					return len(ids) == 2 && ids[0] == 1 && ids[1] == 1
				})).Return(map[int64]string{1: "user-uid"}, nil).Once()
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
			setupMocks: func(fr *repomocks.MockUserFileRepository, ur *resolvermocks.MockUserResolver) {
				fr.EXPECT().List(mock.Anything, mock.AnythingOfType("*param.PaginationParam"), mock.AnythingOfType("*param.UserFileListFilterParam")).Return(&model.UserFiles{
					Items: []model.UserFile{},
				}, nil).Once()
			},
			pagination: &param.PaginationParam{
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
			setupMocks: func(fr *repomocks.MockUserFileRepository, ur *resolvermocks.MockUserResolver) {
				fr.EXPECT().List(mock.Anything, mock.AnythingOfType("*param.PaginationParam"), mock.AnythingOfType("*param.UserFileListFilterParam")).Return(&model.UserFiles{Items: []model.UserFile{}}, nil).Once()
			},
			pagination: &param.PaginationParam{
				Page:    util.Ptr(1),
				Limit:   util.Ptr(10),
				Sort:    util.Ptr("asc"),
				OrderBy: util.Ptr("created_at"),
			},
			filter: &param.UserFileListFilterParam{
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
			// Setup mocks using generated mocks
			mockUserFileRepo := repomocks.NewMockUserFileRepository(t)
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockUserResolver := resolvermocks.NewMockUserResolver(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)

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
				func() *observermocks.MockServiceObserver[signal.UserFileSignal] {
					obs := observermocks.NewMockServiceObserver[signal.UserFileSignal](t)
					setupUserFileObserverAny(t, obs)
					return obs
				}(),
				eventmocks.NewMockEventPublisher(t),
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
		setupMocks func(*repomocks.MockUserRepository, *repomocks.MockUserFileRepository, *securitymocks.MockUIDGenerator, *eventmocks.MockEventPublisher)
		param      param.UserFileCreateParam
		want       *model.UserFile
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *repomocks.MockUserRepository, fr *repomocks.MockUserFileRepository, ug *securitymocks.MockUIDGenerator, ep *eventmocks.MockEventPublisher) {
				ur.EXPECT().GetByUID(mock.Anything, "user-uid").Return(createTestUser(1, "user-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				ug.EXPECT().New().Return("file-uid").Once()
				fr.EXPECT().Create(mock.Anything, mock.MatchedBy(func(f *model.UserFile) bool {
					return f.UserUID == "user-uid" && f.FileName == "photo.jpg"
				})).RunAndReturn(func(ctx context.Context, file *model.UserFile) (*model.UserFile, error) {
					file.ID = 1
					return file, nil
				}).Once()
				ep.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()
			},
			param: param.UserFileCreateParam{
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
			param: param.UserFileCreateParam{
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
			setupMocks: func(ur *repomocks.MockUserRepository, fr *repomocks.MockUserFileRepository, ug *securitymocks.MockUIDGenerator, ep *eventmocks.MockEventPublisher) {
				ur.EXPECT().GetByUID(mock.Anything, "nonexistent-uid").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			param: param.UserFileCreateParam{
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
			// Setup mocks using generated mocks
			mockUserFileRepo := repomocks.NewMockUserFileRepository(t)
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockUserResolver := resolvermocks.NewMockUserResolver(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockEventPublisher := eventmocks.NewMockEventPublisher(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockUserFileRepo, mockUIDGen, mockEventPublisher)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				func() *observermocks.MockServiceObserver[signal.UserFileSignal] {
					obs := observermocks.NewMockServiceObserver[signal.UserFileSignal](t)
					setupUserFileObserverAny(t, obs)
					return obs
				}(),
				mockEventPublisher,
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
		setupMocks func(*repomocks.MockUserFileRepository, *eventmocks.MockEventPublisher)
		uid        string
		param      param.UserFileUpdateParam
		wantErr    error
	}{
		{
			name: "Happy Path - update all fields",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				fr.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.UserFile")).Return(nil).Once()
				ep.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()
			},
			uid: "file-uid",
			param: param.UserFileUpdateParam{
				FileName:   util.Ptr("newname.jpg"),
				FilePath:   util.Ptr("/uploads/newname.jpg"),
				MimeType:   util.Ptr("image/png"),
				SizeBytes:  util.Ptr(int64(2048)),
				Visibility: util.Ptr(model.FileVisibilityPublic),
			},
		},
		{
			name: "Happy Path - partial update",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				fr.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.UserFile")).Return(nil).Once()
				ep.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()
			},
			uid: "file-uid",
			param: param.UserFileUpdateParam{
				FileName: util.Ptr("updated.jpg"),
			},
		},
		{
			name: "Error - file not found",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "nonexistent-file").Return(nil, domainerrors.ErrFileNotFound).Once()
			},
			uid: "nonexistent-file",
			param: param.UserFileUpdateParam{
				FileName: util.Ptr("newname.jpg"),
			},
			wantErr: domainerrors.ErrFileNotFound,
		},
		{
			name: "Error - repository update fails",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				fr.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.UserFile")).Return(errors.New("db error")).Once()
			},
			uid: "file-uid",
			param: param.UserFileUpdateParam{
				FileName: util.Ptr("updated.jpg"),
			},
			wantErr: errors.New("db error"),
		},
		{
			name: "Error - event publish fails",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				fr.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.UserFile")).Return(nil).Once()
				ep.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("event error")).Once()
			},
			uid: "file-uid",
			param: param.UserFileUpdateParam{
				FileName: util.Ptr("updated.jpg"),
			},
			wantErr: errors.New("event error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserFileRepo := repomocks.NewMockUserFileRepository(t)
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockUserResolver := resolvermocks.NewMockUserResolver(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockEventPublisher := eventmocks.NewMockEventPublisher(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFileRepo, mockEventPublisher)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				func() *observermocks.MockServiceObserver[signal.UserFileSignal] {
					obs := observermocks.NewMockServiceObserver[signal.UserFileSignal](t)
					setupUserFileObserverAny(t, obs)
					return obs
				}(),
				mockEventPublisher,
			)

			// Execute
			err := svc.Update(context.Background(), tt.uid, tt.param)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestUserFileService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*repomocks.MockUserFileRepository, *eventmocks.MockEventPublisher)
		uid        string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				fr.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*model.UserFile")).Return(nil).Once()
				ep.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil).Once()
			},
			uid: "file-uid",
		},
		{
			name: "Error - file not found",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "nonexistent-file").Return(nil, domainerrors.ErrFileNotFound).Once()
			},
			uid:     "nonexistent-file",
			wantErr: domainerrors.ErrFileNotFound,
		},
		{
			name: "Error - repository delete fails",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				fr.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*model.UserFile")).Return(errors.New("db error")).Once()
			},
			uid:     "file-uid",
			wantErr: errors.New("db error"),
		},
		{
			name: "Error - event publish fails",
			setupMocks: func(fr *repomocks.MockUserFileRepository, ep *eventmocks.MockEventPublisher) {
				fr.EXPECT().GetByUID(mock.Anything, "file-uid").Return(createUserFile(1, "file-uid", "user-uid", "image"), nil).Once()
				fr.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*model.UserFile")).Return(nil).Once()
				ep.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("event error")).Once()
			},
			uid:     "file-uid",
			wantErr: errors.New("event error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserFileRepo := repomocks.NewMockUserFileRepository(t)
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockUserResolver := resolvermocks.NewMockUserResolver(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockEventPublisher := eventmocks.NewMockEventPublisher(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserFileRepo, mockEventPublisher)
			}

			// Create service
			svc := NewUserFileService(
				mockUserFileRepo,
				mockUserRepo,
				mockUserResolver,
				mockUIDGen,
				func() *observermocks.MockServiceObserver[signal.UserFileSignal] {
					obs := observermocks.NewMockServiceObserver[signal.UserFileSignal](t)
					setupUserFileObserverAny(t, obs)
					return obs
				}(),
				mockEventPublisher,
			)

			// Execute
			err := svc.Delete(context.Background(), tt.uid)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
		})
	}
}

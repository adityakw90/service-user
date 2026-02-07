package service

import (
	"context"
	"errors"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function for UserStatus pointer
func userStatusPtr(s model.UserStatus) *model.UserStatus {
	return &s
}

func TestUserService_Get(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository)
		uid        string
		want       *model.User
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *MockUserRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil
				}
			},
			uid:  "test-uid",
			want: createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive),
		},
		{
			name:    "Error - empty UID",
			uid:     "",
			wantErr: domainerrors.ErrInvalidUID,
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			uid:     "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - deleted user",
			setupMocks: func(ur *MockUserRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil
				}
			},
			uid:     "deleted-uid",
			wantErr: domainerrors.ErrUserDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()
			mockObserver := NewMockUserObserver()
			mockTokenWhitelist := NewMockTokenStore()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
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
			assert.Equal(t, tt.want.Username, got.Username)
			assert.Equal(t, tt.want.Email, got.Email)
		})
	}
}

func TestUserService_List(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository)
		pagination *params.PaginationParam
		filter     *params.UserListFilterParam
		want       *model.Users
		wantErr    error
	}{
		{
			name: "Happy Path - default pagination",
			setupMocks: func(ur *MockUserRepository) {
				ur.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.UserListFilterParam) (*model.Users, error) {
					return &model.Users{
						Items: []model.User{
							*createTestUser(1, "uid1", "user1", "user1@example.com", "pass", model.UserStatusActive),
							*createTestUser(2, "uid2", "user2", "user2@example.com", "pass", model.UserStatusActive),
						},
						Meta: model.Meta{
							Page:  1,
							Limit: 10,
							Total: 2,
							Pages: 1,
						},
					}, nil
				}
			},
			pagination: nil,
			filter:     nil,
			want: &model.Users{
				Items: []model.User{
					*createTestUser(1, "uid1", "user1", "user1@example.com", "pass", model.UserStatusActive),
					*createTestUser(2, "uid2", "user2", "user2@example.com", "pass", model.UserStatusActive),
				},
			},
		},
		{
			name: "Happy Path - custom pagination",
			setupMocks: func(ur *MockUserRepository) {
				ur.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.UserListFilterParam) (*model.Users, error) {
					return &model.Users{
						Items: []model.User{},
						Meta:  model.Meta{Page: 2, Limit: 20},
					}, nil
				}
			},
			pagination: &params.PaginationParam{
				Page:    util.Ptr(2),
				Limit:   util.Ptr(20),
				OrderBy: util.Ptr("created_at"),
				Sort:    util.Ptr("desc"),
			},
			filter: nil,
			want: &model.Users{
				Items: []model.User{},
			},
		},
		{
			name: "Happy Path - with filters",
			setupMocks: func(ur *MockUserRepository) {
				ur.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.UserListFilterParam) (*model.Users, error) {
					return &model.Users{Items: []model.User{}}, nil
				}
			},
			pagination: &params.PaginationParam{
				Page:    util.Ptr(1),
				Limit:   util.Ptr(10),
				OrderBy: util.Ptr("created_at"),
				Sort:    util.Ptr("desc"),
			},
			filter: &params.UserListFilterParam{
				Status: userStatusPtr(model.UserStatusActive),
			},
			want: &model.Users{
				Items: []model.User{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()
			mockObserver := NewMockUserObserver()
			mockTokenWhitelist := NewMockTokenStore()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
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
		})
	}
}

func TestUserService_Create(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockUserProfileRepository, *MockHasher, *MockUIDGenerator)
		input      *params.UserCreateParam
		want       *model.User
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository, h *MockHasher, ug *MockUIDGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				ur.GetByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				h.HashFunc = func(plain string) (string, error) {
					return "hashed_password", nil
				}
				ug.NewFunc = func() string { return "new-uid" }
				ur.CreateFunc = func(ctx context.Context, user *model.User) (*model.User, error) {
					user.ID = 1
					return user, nil
				}
				pr.CreateFunc = func(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
					profile.UserID = 1
					return profile, nil
				}
			},
			input: createUserCreateParams("newuser", "newuser@example.com", "password123"),
			want:  createTestUser(1, "new-uid", "newuser", "newuser@example.com", "hashed_password", model.UserStatusActive),
		},
		{
			name:    "Error - empty username",
			input:   createUserCreateParams("", "test@example.com", "password123"),
			wantErr: domainerrors.ErrInvalidUsername,
		},
		{
			name:    "Error - empty email",
			input:   createUserCreateParams("testuser", "", "password123"),
			wantErr: domainerrors.ErrInvalidEmail,
		},
		{
			name:    "Error - empty password",
			input:   createUserCreateParams("testuser", "test@example.com", ""),
			wantErr: domainerrors.ErrInvalidPassword,
		},
		{
			name: "Error - duplicate email",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository, h *MockHasher, ug *MockUIDGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(1, "existing-uid", "existing", "existing@example.com", "pass", model.UserStatusActive), nil
				}
			},
			input:   createUserCreateParams("newuser", "existing@example.com", "password123"),
			wantErr: domainerrors.ErrDuplicateEmail,
		},
		{
			name: "Error - duplicate username",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository, h *MockHasher, ug *MockUIDGenerator) {
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				ur.GetByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return createTestUser(1, "existing-uid", "existing", "existing@example.com", "pass", model.UserStatusActive), nil
				}
			},
			input:   createUserCreateParams("existing", "newuser@example.com", "password123"),
			wantErr: domainerrors.ErrDuplicateUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockProfileRepo, mockHasher, mockUIDGen)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				NewMockUserObserver(),
			)

			// Execute
			got, err := svc.Create(context.Background(), tt.input)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Username, got.Username)
			assert.Equal(t, tt.want.Email, got.Email)
		})
	}
}

func TestUserService_Update(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockHasher)
		uid        string
		param      *params.UserUpdateParam
		wantErr    error
	}{
		{
			name: "Happy Path - update username",
			setupMocks: func(ur *MockUserRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "olduser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				ur.GetByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				ur.UpdateFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			uid:   "test-uid",
			param: createUserUpdateParams(util.Ptr("newuser"), nil, nil, nil),
		},
		{
			name: "Happy Path - update email",
			setupMocks: func(ur *MockUserRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "old@example.com", "pass", model.UserStatusActive), nil
				}
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				ur.UpdateFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			uid:   "test-uid",
			param: createUserUpdateParams(nil, util.Ptr("new@example.com"), nil, nil),
		},
		{
			name: "Happy Path - update password",
			setupMocks: func(ur *MockUserRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "oldpass", model.UserStatusActive), nil
				}
				h.HashFunc = func(plain string) (string, error) {
					return "new_hashed_pass", nil
				}
				ur.UpdateFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			uid:   "test-uid",
			param: createUserUpdateParams(nil, nil, util.Ptr("newpassword"), nil),
		},
		{
			name: "Happy Path - update status",
			setupMocks: func(ur *MockUserRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				ur.UpdateFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			uid:   "test-uid",
			param: createUserUpdateParams(nil, nil, nil, userStatusPtr(model.UserStatusInactive)),
		},
		{
			name: "Error - username conflict",
			setupMocks: func(ur *MockUserRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				ur.GetByUsernameFunc = func(ctx context.Context, username string) (*model.User, error) {
					return createTestUser(2, "other-uid", "existing", "other@example.com", "pass", model.UserStatusActive), nil
				}
			},
			uid:     "test-uid",
			param:   createUserUpdateParams(util.Ptr("existing"), nil, nil, nil),
			wantErr: domainerrors.ErrDuplicateUsername,
		},
		{
			name: "Error - email conflict",
			setupMocks: func(ur *MockUserRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				ur.GetByEmailFunc = func(ctx context.Context, email string) (*model.User, error) {
					return createTestUser(2, "other-uid", "other", "existing@example.com", "pass", model.UserStatusActive), nil
				}
			},
			uid:     "test-uid",
			param:   createUserUpdateParams(nil, util.Ptr("existing@example.com"), nil, nil),
			wantErr: domainerrors.ErrDuplicateEmail,
		},
		{
			name: "Error - deleted user",
			setupMocks: func(ur *MockUserRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil
				}
			},
			uid:     "deleted-uid",
			param:   createUserUpdateParams(util.Ptr("newname"), nil, nil, nil),
			wantErr: domainerrors.ErrUserDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockHasher)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				NewMockUserObserver(),
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

func TestUserService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository)
		uid        string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *MockUserRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				ur.DeleteFunc = func(ctx context.Context, user *model.User) error {
					return nil
				}
			},
			uid: "test-uid",
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			uid:     "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()
			mockObserver := NewMockUserObserver()
			mockTokenWhitelist := NewMockTokenStore()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
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

func TestUserService_GetProfile(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockUserProfileRepository)
		userUID    string
		want       *model.UserProfile
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserProfile, error) {
					return createTestProfile(1, "test-uid"), nil
				}
			},
			userUID: "test-uid",
			want:    createTestProfile(1, "test-uid"),
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			userUID: "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - profile not found",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserProfile, error) {
					return nil, errors.New("profile not found")
				}
			},
			userUID: "test-uid",
			wantErr: domainerrors.ErrProfileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockProfileRepo)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				NewMockUserObserver(),
			)

			// Execute
			got, err := svc.GetProfile(context.Background(), tt.userUID)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.FirstName, got.FirstName)
			assert.Equal(t, tt.want.LastName, got.LastName)
		})
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockUserProfileRepository)
		userUID    string
		opts       params.UserProfileUpdateParam
		wantErr    error
	}{
		{
			name: "Happy Path - update all fields",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserProfile, error) {
					return createTestProfile(userID, "test-uid"), nil
				}
				pr.UpdateFunc = func(ctx context.Context, profile *model.UserProfile) error {
					return nil
				}
			},
			userUID: "test-uid",
			opts: params.UserProfileUpdateParam{
				FirstName: util.Ptr("John"),
				LastName:  util.Ptr("Doe"),
				Bio:       util.Ptr("New bio"),
			},
		},
		{
			name: "Happy Path - partial update",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserProfile, error) {
					return createTestProfile(userID, "test-uid"), nil
				}
				pr.UpdateFunc = func(ctx context.Context, profile *model.UserProfile) error {
					return nil
				}
			},
			userUID: "test-uid",
			opts: params.UserProfileUpdateParam{
				Bio: util.Ptr("Updated bio only"),
			},
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			userUID: "nonexistent-uid",
			opts: params.UserProfileUpdateParam{
				FirstName: util.Ptr("John"),
			},
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - profile not found",
			setupMocks: func(ur *MockUserRepository, pr *MockUserProfileRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserProfile, error) {
					return nil, errors.New("not found")
				}
			},
			userUID: "test-uid",
			opts: params.UserProfileUpdateParam{
				FirstName: util.Ptr("John"),
			},
			wantErr: domainerrors.ErrProfileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockProfileRepo)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				NewMockUserObserver(),
			)

			// Execute
			err := svc.UpdateProfile(context.Background(), tt.userUID, tt.opts)

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

func TestUserService_SetPin(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockUserPinRepository, *MockHasher)
		userUID    string
		pin        string
		wantErr    error
	}{
		{
			name: "Happy Path - new PIN",
			setupMocks: func(ur *MockUserRepository, pr *MockUserPinRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserPin, error) {
					return nil, domainerrors.ErrUserNotFound
				}
				h.HashFunc = func(plain string) (string, error) {
					return "hashed_1234", nil
				}
				pr.CreateFunc = func(ctx context.Context, pin *model.UserPin) (*model.UserPin, error) {
					return pin, nil
				}
			},
			userUID: "test-uid",
			pin:     "1234",
		},
		{
			name: "Happy Path - update existing PIN",
			setupMocks: func(ur *MockUserRepository, pr *MockUserPinRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				pr.GetByUserIDFunc = func(ctx context.Context, userID int64) (*model.UserPin, error) {
					return createUserPin(userID, "test-uid", "old_hashed_pin"), nil
				}
				h.HashFunc = func(plain string) (string, error) {
					return "new_hashed_pin", nil
				}
				pr.UpdateFunc = func(ctx context.Context, pin *model.UserPin) error {
					return nil
				}
			},
			userUID: "test-uid",
			pin:     "9999",
		},
		{
			name:    "Error - empty PIN",
			userUID: "test-uid",
			pin:     "",
			wantErr: domainerrors.ErrPinInvalid,
		},
		{
			name: "Error - deleted user",
			setupMocks: func(ur *MockUserRepository, pr *MockUserPinRepository, h *MockHasher) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil
				}
			},
			userUID: "deleted-uid",
			pin:     "1234",
			wantErr: domainerrors.ErrUserDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockPinRepo, mockHasher)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				NewMockUserObserver(),
			)

			// Execute
			err := svc.SetPin(context.Background(), tt.userUID, tt.pin)

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

func TestUserService_ListDevice(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockDeviceRepository)
		userUID    string
		opts       params.UserDeviceListFilterParam
		want       *model.Devices
		wantErr    error
	}{
		{
			name: "Happy Path with defaults",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				dr.ListByUserIDFunc = func(ctx context.Context, userId int64, p *params.PaginationParam, f *params.DeviceListFilterParam) (*model.Devices, error) {
					return &model.Devices{
						Items: []model.Device{
							*createTestDevice(1, "device1", "iPhone", "fp1"),
						},
					}, nil
				}
			},
			userUID: "test-uid",
			opts:    params.UserDeviceListFilterParam{},
			want: &model.Devices{
				Items: []model.Device{
					*createTestDevice(1, "device1", "iPhone", "fp1"),
				},
			},
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			userUID: "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockDeviceRepo)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				NewMockUserObserver(),
			)

			// Execute
			got, err := svc.ListDevice(context.Background(), tt.userUID, nil, &tt.opts)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestUserService_RevokeDevice(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockUserRepository, *MockDeviceRepository, *MockUserDeviceRepository)
		userUID    string
		deviceUID  string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				dr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.Device, error) {
					return createTestDevice(1, "device-uid", "iPhone", "fp123"), nil
				}
				udr.RevokeFunc = func(ctx context.Context, userID int64, deviceID int64) error {
					return nil
				}
			},
			userUID:   "test-uid",
			deviceUID: "device-uid",
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return nil, domainerrors.ErrUserNotFound
				}
			},
			userUID:   "nonexistent-uid",
			deviceUID: "device-uid",
			wantErr:   domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - device not found",
			setupMocks: func(ur *MockUserRepository, dr *MockDeviceRepository, udr *MockUserDeviceRepository) {
				ur.GetByUIDFunc = func(ctx context.Context, uid string) (*model.User, error) {
					return createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil
				}
				dr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.Device, error) {
					return nil, domainerrors.ErrDeviceNotFound
				}
			},
			userUID:   "test-uid",
			deviceUID: "nonexistent-device",
			wantErr:   domainerrors.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockUserRepo := NewMockUserRepository()
			mockProfileRepo := NewMockUserProfileRepository()
			mockPinRepo := NewMockUserPinRepository()
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()
			mockHasher := NewMockHasher()
			mockUIDGen := NewMockUIDGenerator()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockDeviceRepo, mockUserDeviceRepo)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockHasher,
				mockHasher,
				mockUIDGen,
				NewMockUserObserver(),
			)

			// Execute
			err := svc.RevokeDevice(context.Background(), tt.userUID, tt.deviceUID)

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

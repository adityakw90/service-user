package service

import (
	"context"
	"errors"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	"github.com/adityakw90/service-user/internal/adapter/publisher"
	repomocks "github.com/adityakw90/service-user/test/mocks/repository"
	securitymocks "github.com/adityakw90/service-user/test/mocks/security"
	observermocks "github.com/adityakw90/service-user/test/mocks/observer"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function for UserStatus pointer
func userStatusPtr(s model.UserStatus) *model.UserStatus {
	return &s
}

// setupObserverAny allows any OnSignal calls on the observer (useful when not testing signal behavior)
func setupObserverAny(t *testing.T, observer *observermocks.MockServiceObserver[signal.UserSignal]) {
	// Allow any OnSignal call without checking parameters
	// Use Maybe() to make the expectation optional (can be called 0 or more times)
	// Note: Using EXPECT().OnSignal() pattern for better type safety
	observer.EXPECT().OnSignal(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
}

func TestUserService_Get(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*repomocks.MockUserRepository)
		uid        string
		want       *model.User
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().
					GetByUID(mock.Anything, "test-uid").
					Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "hashed_password", model.UserStatusActive), nil).
					Once()
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
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().
					GetByUID(mock.Anything, "nonexistent-uid").
					Return(nil, domainerrors.ErrUserNotFound).
					Once()
			},
			uid:     "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - deleted user",
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().
					GetByUID(mock.Anything, "deleted-uid").
					Return(createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil).
					Once()
			},
			uid:     "deleted-uid",
			wantErr: domainerrors.ErrUserDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

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
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository)
		pagination *params.PaginationParam
		filter     *params.UserListFilterParam
		want       *model.Users
		wantErr    error
	}{
		{
			name: "Happy Path - default pagination",
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().
					List(mock.Anything, mock.AnythingOfType("*params.PaginationParam"), mock.AnythingOfType("*params.UserListFilterParam")).
					Return(&model.Users{
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
					}, nil).
					Once()
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
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().
					List(mock.Anything, mock.AnythingOfType("*params.PaginationParam"), mock.AnythingOfType("*params.UserListFilterParam")).
					Return(&model.Users{
						Items: []model.User{},
						Meta:  model.Meta{Page: 2, Limit: 20},
					}, nil).
					Once()
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
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().
					List(mock.Anything, mock.AnythingOfType("*params.PaginationParam"), mock.AnythingOfType("*params.UserListFilterParam")).
					Return(&model.Users{Items: []model.User{}}, nil).
					Once()
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
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

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
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository, *repomocks.MockUserProfileRepository, *securitymocks.MockHasher, *securitymocks.MockUIDGenerator)
		input      *params.UserCreateParam
		want       *model.User
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository, h *securitymocks.MockHasher, ug *securitymocks.MockUIDGenerator) {
				ur.EXPECT().GetByEmail(mock.Anything, "newuser@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
				ur.EXPECT().GetByUsername(mock.Anything, "newuser").Return(nil, domainerrors.ErrUserNotFound).Once()
				h.EXPECT().Hash("password123").Return("hashed_password", nil).Once()
				ug.EXPECT().New().Return("new-uid").Once()
				ur.EXPECT().Create(mock.Anything, mock.MatchedBy(func(u *model.User) bool {
					return u.Email == "newuser@example.com" && u.Username == "newuser"
				})).RunAndReturn(func(ctx context.Context, user *model.User) (*model.User, error) {
					user.ID = 1
					return user, nil
				}).Once()
				pr.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.UserProfile")).RunAndReturn(func(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
					profile.UserID = 1
					return profile, nil
				}).Once()
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
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository, h *securitymocks.MockHasher, ug *securitymocks.MockUIDGenerator) {
				ur.EXPECT().GetByEmail(mock.Anything, "existing@example.com").Return(createTestUser(1, "existing-uid", "existing", "existing@example.com", "pass", model.UserStatusActive), nil).Once()
			},
			input:   createUserCreateParams("newuser", "existing@example.com", "password123"),
			wantErr: domainerrors.ErrDuplicateEmail,
		},
		{
			name: "Error - duplicate username",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository, h *securitymocks.MockHasher, ug *securitymocks.MockUIDGenerator) {
				ur.EXPECT().GetByEmail(mock.Anything, "newuser@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
				ur.EXPECT().GetByUsername(mock.Anything, "existing").Return(createTestUser(1, "existing-uid", "existing", "existing@example.com", "pass", model.UserStatusActive), nil).Once()
			},
			input:   createUserCreateParams("existing", "newuser@example.com", "password123"),
			wantErr: domainerrors.ErrDuplicateUsername,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockProfileRepo, mockPasswordHasher, mockUIDGen)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository, *securitymocks.MockHasher)
		uid        string
		param      *params.UserUpdateParam
		wantErr    error
	}{
		{
			name: "Happy Path - update username",
			setupMocks: func(ur *repomocks.MockUserRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "olduser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				ur.EXPECT().GetByUsername(mock.Anything, "newuser").Return(nil, domainerrors.ErrUserNotFound).Once()
				ur.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Once()
			},
			uid:   "test-uid",
			param: createUserUpdateParams(util.Ptr("newuser"), nil, nil, nil),
		},
		{
			name: "Happy Path - update email",
			setupMocks: func(ur *repomocks.MockUserRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "old@example.com", "pass", model.UserStatusActive), nil).Once()
				ur.EXPECT().GetByEmail(mock.Anything, "new@example.com").Return(nil, domainerrors.ErrUserNotFound).Once()
				ur.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Once()
			},
			uid:   "test-uid",
			param: createUserUpdateParams(nil, util.Ptr("new@example.com"), nil, nil),
		},
		{
			name: "Happy Path - update password",
			setupMocks: func(ur *repomocks.MockUserRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "oldpass", model.UserStatusActive), nil).Once()
				h.EXPECT().Hash("newpassword").Return("new_hashed_pass", nil).Once()
				ur.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Once()
			},
			uid:   "test-uid",
			param: createUserUpdateParams(nil, nil, util.Ptr("newpassword"), nil),
		},
		{
			name: "Happy Path - update status",
			setupMocks: func(ur *repomocks.MockUserRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				ur.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Once()
			},
			uid:   "test-uid",
			param: createUserUpdateParams(nil, nil, nil, userStatusPtr(model.UserStatusInactive)),
		},
		{
			name: "Error - username conflict",
			setupMocks: func(ur *repomocks.MockUserRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				ur.EXPECT().GetByUsername(mock.Anything, "existing").Return(createTestUser(2, "other-uid", "existing", "other@example.com", "pass", model.UserStatusActive), nil).Once()
			},
			uid:     "test-uid",
			param:   createUserUpdateParams(util.Ptr("existing"), nil, nil, nil),
			wantErr: domainerrors.ErrDuplicateUsername,
		},
		{
			name: "Error - email conflict",
			setupMocks: func(ur *repomocks.MockUserRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				ur.EXPECT().GetByEmail(mock.Anything, "existing@example.com").Return(createTestUser(2, "other-uid", "other", "existing@example.com", "pass", model.UserStatusActive), nil).Once()
			},
			uid:     "test-uid",
			param:   createUserUpdateParams(nil, util.Ptr("existing@example.com"), nil, nil),
			wantErr: domainerrors.ErrDuplicateEmail,
		},
		{
			name: "Error - deleted user",
			setupMocks: func(ur *repomocks.MockUserRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "deleted-uid").Return(createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil).Once()
			},
			uid:     "deleted-uid",
			param:   createUserUpdateParams(util.Ptr("newname"), nil, nil, nil),
			wantErr: domainerrors.ErrUserDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockPasswordHasher)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository)
		uid        string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				ur.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*model.User")).Return(nil).Once()
			},
			uid: "test-uid",
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *repomocks.MockUserRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "nonexistent-uid").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			uid:     "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

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
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository, *repomocks.MockUserProfileRepository)
		userUID    string
		want       *model.UserProfile
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				pr.EXPECT().GetByUserID(mock.Anything, int64(1)).Return(createTestProfile(1, "test-uid"), nil).Once()
			},
			userUID: "test-uid",
			want:    createTestProfile(1, "test-uid"),
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "nonexistent-uid").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			userUID: "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - profile not found",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				pr.EXPECT().GetByUserID(mock.Anything, int64(1)).Return(nil, errors.New("profile not found")).Once()
			},
			userUID: "test-uid",
			wantErr: domainerrors.ErrProfileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

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
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository, *repomocks.MockUserProfileRepository)
		userUID    string
		opts       params.UserProfileUpdateParam
		wantErr    error
	}{
		{
			name: "Happy Path - update all fields",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				pr.EXPECT().GetByUserID(mock.Anything, int64(1)).Return(createTestProfile(1, "test-uid"), nil).Once()
				pr.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.UserProfile")).Return(nil).Once()
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
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				pr.EXPECT().GetByUserID(mock.Anything, int64(1)).Return(createTestProfile(1, "test-uid"), nil).Once()
				pr.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.UserProfile")).Return(nil).Once()
			},
			userUID: "test-uid",
			opts: params.UserProfileUpdateParam{
				Bio: util.Ptr("Updated bio only"),
			},
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "nonexistent-uid").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			userUID: "nonexistent-uid",
			opts: params.UserProfileUpdateParam{
				FirstName: util.Ptr("John"),
			},
			wantErr: domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - profile not found",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserProfileRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				pr.EXPECT().GetByUserID(mock.Anything, int64(1)).Return(nil, errors.New("not found")).Once()
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
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

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
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository, *repomocks.MockUserPinRepository, *securitymocks.MockHasher)
		userUID    string
		pin        string
		wantErr    error
	}{
		{
			name: "Happy Path - new PIN",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserPinRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				pr.EXPECT().GetByUserID(mock.Anything, int64(1)).Return(nil, domainerrors.ErrUserNotFound).Once()
				h.EXPECT().Hash("1234").Return("hashed_1234", nil).Once()
				pr.EXPECT().Create(mock.Anything, mock.AnythingOfType("*model.UserPin")).RunAndReturn(func(ctx context.Context, pin *model.UserPin) (*model.UserPin, error) {
					return pin, nil
				}).Once()
			},
			userUID: "test-uid",
			pin:     "1234",
		},
		{
			name: "Happy Path - update existing PIN",
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserPinRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				pr.EXPECT().GetByUserID(mock.Anything, int64(1)).Return(createUserPin(1, "test-uid", "old_hashed_pin"), nil).Once()
				h.EXPECT().Hash("9999").Return("new_hashed_pin", nil).Once()
				pr.EXPECT().Update(mock.Anything, mock.AnythingOfType("*model.UserPin")).Return(nil).Once()
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
			setupMocks: func(ur *repomocks.MockUserRepository, pr *repomocks.MockUserPinRepository, h *securitymocks.MockHasher) {
				ur.EXPECT().GetByUID(mock.Anything, "deleted-uid").Return(createDeletedUser(1, "test-uid", "testuser", "test@example.com"), nil).Once()
			},
			userUID: "deleted-uid",
			pin:     "1234",
			wantErr: domainerrors.ErrUserDeleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockPinRepo, mockPinHasher)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository, *repomocks.MockDeviceRepository)
		userUID    string
		opts       params.UserDeviceListFilterParam
		want       *model.Devices
		wantErr    error
	}{
		{
			name: "Happy Path with defaults",
			setupMocks: func(ur *repomocks.MockUserRepository, dr *repomocks.MockDeviceRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				dr.EXPECT().ListByUserID(mock.Anything, int64(1), mock.AnythingOfType("*params.PaginationParam"), mock.AnythingOfType("*params.DeviceListFilterParam")).Return(&model.Devices{
					Items: []model.Device{
						*createTestDevice(1, "device1", "iPhone", "fp1"),
					},
				}, nil).Once()
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
			setupMocks: func(ur *repomocks.MockUserRepository, dr *repomocks.MockDeviceRepository) {
				ur.EXPECT().GetByUID(mock.Anything, "nonexistent-uid").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			userUID: "nonexistent-uid",
			wantErr: domainerrors.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

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
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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
		setupMocks func(*repomocks.MockUserRepository, *repomocks.MockDeviceRepository, *repomocks.MockUserDeviceRepository, *securitymocks.MockTokenStore)
		userUID    string
		deviceUID  string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(ur *repomocks.MockUserRepository, dr *repomocks.MockDeviceRepository, udr *repomocks.MockUserDeviceRepository, tw *securitymocks.MockTokenStore) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				dr.EXPECT().GetByUID(mock.Anything, "device-uid").Return(createTestDevice(1, "device-uid", "iPhone", "fp123"), nil).Once()
				udr.EXPECT().GetByUserIDAndDeviceID(mock.Anything, int64(1), int64(1)).Return(&model.UserDevice{
					UserID:    1,
					DeviceID:  1,
					SessionID: "session-123",
				}, nil).Once()
				tw.EXPECT().Remove(mock.Anything, "test-uid", "session-123").Return(nil).Once()
				udr.EXPECT().Revoke(mock.Anything, int64(1), int64(1)).Return(nil).Once()
			},
			userUID:   "test-uid",
			deviceUID: "device-uid",
		},
		{
			name: "Error - user not found",
			setupMocks: func(ur *repomocks.MockUserRepository, dr *repomocks.MockDeviceRepository, udr *repomocks.MockUserDeviceRepository, tw *securitymocks.MockTokenStore) {
				ur.EXPECT().GetByUID(mock.Anything, "nonexistent-uid").Return(nil, domainerrors.ErrUserNotFound).Once()
			},
			userUID:   "nonexistent-uid",
			deviceUID: "device-uid",
			wantErr:   domainerrors.ErrUserNotFound,
		},
		{
			name: "Error - device not found",
			setupMocks: func(ur *repomocks.MockUserRepository, dr *repomocks.MockDeviceRepository, udr *repomocks.MockUserDeviceRepository, tw *securitymocks.MockTokenStore) {
				ur.EXPECT().GetByUID(mock.Anything, "test-uid").Return(createTestUser(1, "test-uid", "testuser", "test@example.com", "pass", model.UserStatusActive), nil).Once()
				dr.EXPECT().GetByUID(mock.Anything, "nonexistent-device").Return(nil, domainerrors.ErrDeviceNotFound).Once()
			},
			userUID:   "test-uid",
			deviceUID: "nonexistent-device",
			wantErr:   domainerrors.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockUserRepo := repomocks.NewMockUserRepository(t)
			mockProfileRepo := repomocks.NewMockUserProfileRepository(t)
			mockPinRepo := repomocks.NewMockUserPinRepository(t)
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)
			mockPasswordHasher := securitymocks.NewMockHasher(t)
			mockPinHasher := securitymocks.NewMockHasher(t)
			mockUIDGen := securitymocks.NewMockUIDGenerator(t)
			mockObserver := observermocks.NewMockServiceObserver[signal.UserSignal](t)
			setupObserverAny(t, mockObserver)
			mockTokenWhitelist := securitymocks.NewMockTokenStore(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockUserRepo, mockDeviceRepo, mockUserDeviceRepo, mockTokenWhitelist)
			}

			// Create service
			svc := NewUserService(
				mockUserRepo,
				mockProfileRepo,
				mockPinRepo,
				mockDeviceRepo,
				mockUserDeviceRepo,
				mockPasswordHasher,
				mockPinHasher,
				mockUIDGen,
				mockTokenWhitelist,
				mockObserver,
				publisher.NewNoOpPublisher(),
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

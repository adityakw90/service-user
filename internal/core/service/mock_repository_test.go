package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

// MockUserRepository is a mock implementation of repository.UserRepository.
type MockUserRepository struct {
	CreateFunc        func(ctx context.Context, user *model.User) (*model.User, error)
	UpdateFunc        func(ctx context.Context, user *model.User) error
	DeleteFunc        func(ctx context.Context, user *model.User) error
	GetByIDFunc       func(ctx context.Context, id int64) (*model.User, error)
	GetByUIDFunc      func(ctx context.Context, uid string) (*model.User, error)
	GetByUsernameFunc func(ctx context.Context, username string) (*model.User, error)
	GetByEmailFunc    func(ctx context.Context, email string) (*model.User, error)
	GetByPhoneFunc    func(ctx context.Context, phone string) (*model.User, error)
	ListFunc          func(ctx context.Context, pagination *params.PaginationParam, filter *params.UserListFilterParam) (*model.Users, error)
	AddUserDeviceFunc func(ctx context.Context, user *model.User, device *model.Device) error

	createCalls        int
	updateCalls        int
	deleteCalls        int
	getByIDCalls       int
	getByUIDCalls      int
	getByUsernameCalls int
	getByEmailCalls    int
	getByPhoneCalls    int
	listCalls          int
	addUserDeviceCalls int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{}
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	m.createCalls++
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	m.updateCalls++
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, user *model.User) error {
	m.deleteCalls++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, user)
	}
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	m.getByIDCalls++
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUID(ctx context.Context, uid string) (*model.User, error) {
	m.getByUIDCalls++
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(ctx, uid)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	m.getByUsernameCalls++
	if m.GetByUsernameFunc != nil {
		return m.GetByUsernameFunc(ctx, username)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	m.getByEmailCalls++
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *MockUserRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	m.getByPhoneCalls++
	if m.GetByPhoneFunc != nil {
		return m.GetByPhoneFunc(ctx, phone)
	}
	return nil, nil
}

func (m *MockUserRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserListFilterParam) (*model.Users, error) {
	m.listCalls++
	if m.ListFunc != nil {
		return m.ListFunc(ctx, pagination, filter)
	}
	return &model.Users{}, nil
}

func (m *MockUserRepository) AddUserDevice(ctx context.Context, user *model.User, device *model.Device) error {
	m.addUserDeviceCalls++
	if m.AddUserDeviceFunc != nil {
		return m.AddUserDeviceFunc(ctx, user, device)
	}
	return nil
}

// MockUserProfileRepository is a mock implementation of repository.UserProfileRepository.
type MockUserProfileRepository struct {
	CreateFunc      func(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error)
	UpdateFunc      func(ctx context.Context, profile *model.UserProfile) error
	DeleteFunc      func(ctx context.Context, profile *model.UserProfile) error
	GetByIDFunc     func(ctx context.Context, id int64) (*model.UserProfile, error)
	GetByUIDFunc    func(ctx context.Context, uid string) (*model.UserProfile, error)
	GetByUserIDFunc func(ctx context.Context, userID int64) (*model.UserProfile, error)
	ListFunc        func(ctx context.Context, pagination *params.PaginationParam, filter *params.UserProfileListFilterParam) (*model.UserProfiles, error)

	createCalls      int
	updateCalls      int
	deleteCalls      int
	getByIDCalls     int
	getByUIDCalls    int
	getByUserIDCalls int
	listCalls        int
}

func NewMockUserProfileRepository() *MockUserProfileRepository {
	return &MockUserProfileRepository{}
}

func (m *MockUserProfileRepository) Create(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
	m.createCalls++
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, profile)
	}
	return nil, nil
}

func (m *MockUserProfileRepository) Update(ctx context.Context, profile *model.UserProfile) error {
	m.updateCalls++
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, profile)
	}
	return nil
}

func (m *MockUserProfileRepository) GetByID(ctx context.Context, id int64) (*model.UserProfile, error) {
	m.getByIDCalls++
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserProfileRepository) GetByUID(ctx context.Context, uid string) (*model.UserProfile, error) {
	m.getByUIDCalls++
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(ctx, uid)
	}
	return nil, nil
}

func (m *MockUserProfileRepository) GetByUserID(ctx context.Context, userID int64) (*model.UserProfile, error) {
	m.getByUserIDCalls++
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserProfileRepository) Delete(ctx context.Context, profile *model.UserProfile) error {
	m.deleteCalls++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, profile)
	}
	return nil
}

func (m *MockUserProfileRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserProfileListFilterParam) (*model.UserProfiles, error) {
	m.listCalls++
	if m.ListFunc != nil {
		return m.ListFunc(ctx, pagination, filter)
	}
	return &model.UserProfiles{}, nil
}

// MockDeviceRepository is a mock implementation of repository.DeviceRepository.
type MockDeviceRepository struct {
	CreateFunc           func(ctx context.Context, device *model.Device) (*model.Device, error)
	UpdateFunc           func(ctx context.Context, device *model.Device) error
	DeleteFunc           func(ctx context.Context, device *model.Device) error
	GetByIDFunc          func(ctx context.Context, id int64) (*model.Device, error)
	GetByUIDFunc         func(ctx context.Context, uid string) (*model.Device, error)
	GetByFingerprintFunc func(ctx context.Context, fingerprint string) (*model.Device, error)
	ListFunc             func(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error)
	ListByUserIDFunc     func(ctx context.Context, userId int64, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error)

	createCalls           int
	updateCalls           int
	deleteCalls           int
	getByIDCalls          int
	getByUIDCalls         int
	getByFingerprintCalls int
	listCalls             int
	listByUserIDCalls     int
}

func NewMockDeviceRepository() *MockDeviceRepository {
	return &MockDeviceRepository{}
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *model.Device) (*model.Device, error) {
	m.createCalls++
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, device)
	}
	return nil, nil
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *model.Device) error {
	m.updateCalls++
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, device)
	}
	return nil
}

func (m *MockDeviceRepository) Delete(ctx context.Context, device *model.Device) error {
	m.deleteCalls++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, device)
	}
	return nil
}

func (m *MockDeviceRepository) GetByID(ctx context.Context, id int64) (*model.Device, error) {
	m.getByIDCalls++
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockDeviceRepository) GetByUID(ctx context.Context, uid string) (*model.Device, error) {
	m.getByUIDCalls++
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(ctx, uid)
	}
	return nil, nil
}

func (m *MockDeviceRepository) GetByFingerprint(ctx context.Context, fingerprint string) (*model.Device, error) {
	m.getByFingerprintCalls++
	if m.GetByFingerprintFunc != nil {
		return m.GetByFingerprintFunc(ctx, fingerprint)
	}
	return nil, nil
}

func (m *MockDeviceRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	m.listCalls++
	if m.ListFunc != nil {
		return m.ListFunc(ctx, pagination, filter)
	}
	return &model.Devices{}, nil
}

func (m *MockDeviceRepository) ListByUserID(ctx context.Context, userId int64, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	m.listByUserIDCalls++
	if m.ListByUserIDFunc != nil {
		return m.ListByUserIDFunc(ctx, userId, pagination, filter)
	}
	return &model.Devices{}, nil
}

// MockUserDeviceRepository is a mock implementation of repository.UserDeviceRepository.
type MockUserDeviceRepository struct {
	CreateFunc                 func(ctx context.Context, userDevice *model.UserDevice) (*model.UserDevice, error)
	UpdateFunc                 func(ctx context.Context, userDevice *model.UserDevice) error
	DeleteFunc                 func(ctx context.Context, userDevice *model.UserDevice) error
	GetByIDFunc                func(ctx context.Context, id int64) (*model.UserDevice, error)
	GetByUIDFunc               func(ctx context.Context, uid string) (*model.UserDevice, error)
	GetByUserIDAndDeviceIDFunc func(ctx context.Context, userID int64, deviceID int64) (*model.UserDevice, error)
	ListFunc                   func(ctx context.Context, pagination *params.PaginationParam, filter *params.UserDeviceListFilterParam) (*model.UserDevices, error)
	RevokeFunc                 func(ctx context.Context, userID int64, deviceID int64) error

	createCalls                 int
	updateCalls                 int
	deleteCalls                 int
	getByIDCalls                int
	getByUIDCalls               int
	getByUserIDAndDeviceIDCalls int
	listCalls                   int
	revokeCalls                 int
}

func NewMockUserDeviceRepository() *MockUserDeviceRepository {
	return &MockUserDeviceRepository{}
}

func (m *MockUserDeviceRepository) Create(ctx context.Context, userDevice *model.UserDevice) (*model.UserDevice, error) {
	m.createCalls++
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userDevice)
	}
	return nil, nil
}

func (m *MockUserDeviceRepository) Update(ctx context.Context, userDevice *model.UserDevice) error {
	m.updateCalls++
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, userDevice)
	}
	return nil
}

func (m *MockUserDeviceRepository) Delete(ctx context.Context, userDevice *model.UserDevice) error {
	m.deleteCalls++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, userDevice)
	}
	return nil
}

func (m *MockUserDeviceRepository) GetByID(ctx context.Context, id int64) (*model.UserDevice, error) {
	m.getByIDCalls++
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserDeviceRepository) GetByUID(ctx context.Context, uid string) (*model.UserDevice, error) {
	m.getByUIDCalls++
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(ctx, uid)
	}
	return nil, nil
}

func (m *MockUserDeviceRepository) GetByUserIDAndDeviceID(ctx context.Context, userID int64, deviceID int64) (*model.UserDevice, error) {
	m.getByUserIDAndDeviceIDCalls++
	if m.GetByUserIDAndDeviceIDFunc != nil {
		return m.GetByUserIDAndDeviceIDFunc(ctx, userID, deviceID)
	}
	return nil, nil
}

func (m *MockUserDeviceRepository) Revoke(ctx context.Context, userID int64, deviceID int64) error {
	m.revokeCalls++
	if m.RevokeFunc != nil {
		return m.RevokeFunc(ctx, userID, deviceID)
	}
	return nil
}

func (m *MockUserDeviceRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserDeviceListFilterParam) (*model.UserDevices, error) {
	m.listCalls++
	if m.ListFunc != nil {
		return m.ListFunc(ctx, pagination, filter)
	}
	return &model.UserDevices{}, nil
}

// MockUserPinRepository is a mock implementation of repository.UserPinRepository.
type MockUserPinRepository struct {
	CreateFunc      func(ctx context.Context, pin *model.UserPin) (*model.UserPin, error)
	UpdateFunc      func(ctx context.Context, pin *model.UserPin) error
	DeleteFunc      func(ctx context.Context, pin *model.UserPin) error
	GetByIDFunc     func(ctx context.Context, id int64) (*model.UserPin, error)
	GetByUIDFunc    func(ctx context.Context, uid string) (*model.UserPin, error)
	GetByUserIDFunc func(ctx context.Context, userID int64) (*model.UserPin, error)
	ListFunc        func(ctx context.Context, pagination *params.PaginationParam, filter *params.UserPinListFilterParam) (*model.UserPins, error)

	createCalls      int
	updateCalls      int
	deleteCalls      int
	getByIDCalls     int
	getByUIDCalls    int
	getByUserIDCalls int
	listCalls        int
}

func NewMockUserPinRepository() *MockUserPinRepository {
	return &MockUserPinRepository{}
}

func (m *MockUserPinRepository) Create(ctx context.Context, pin *model.UserPin) (*model.UserPin, error) {
	m.createCalls++
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, pin)
	}
	return nil, nil
}

func (m *MockUserPinRepository) Update(ctx context.Context, pin *model.UserPin) error {
	m.updateCalls++
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, pin)
	}
	return nil
}

func (m *MockUserPinRepository) Delete(ctx context.Context, pin *model.UserPin) error {
	m.deleteCalls++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, pin)
	}
	return nil
}

func (m *MockUserPinRepository) GetByID(ctx context.Context, id int64) (*model.UserPin, error) {
	m.getByIDCalls++
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserPinRepository) GetByUID(ctx context.Context, uid string) (*model.UserPin, error) {
	m.getByUIDCalls++
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(ctx, uid)
	}
	return nil, nil
}

func (m *MockUserPinRepository) GetByUserID(ctx context.Context, userID int64) (*model.UserPin, error) {
	m.getByUserIDCalls++
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserPinRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserPinListFilterParam) (*model.UserPins, error) {
	m.listCalls++
	if m.ListFunc != nil {
		return m.ListFunc(ctx, pagination, filter)
	}
	return &model.UserPins{}, nil
}

// MockUserFileRepository is a mock implementation of repository.UserFileRepository.
type MockUserFileRepository struct {
	CreateFunc      func(ctx context.Context, file *model.UserFile) (*model.UserFile, error)
	UpdateFunc      func(ctx context.Context, file *model.UserFile) error
	DeleteFunc      func(ctx context.Context, file *model.UserFile) error
	GetByIDFunc     func(ctx context.Context, id int64) (*model.UserFile, error)
	GetByUIDFunc    func(ctx context.Context, uid string) (*model.UserFile, error)
	GetByUserIDFunc func(ctx context.Context, userID int64) (*model.UserFile, error)
	ListFunc        func(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error)

	createCalls      int
	updateCalls      int
	deleteCalls      int
	getByIDCalls     int
	getByUIDCalls    int
	getByUserIDCalls int
	listCalls        int
}

func NewMockUserFileRepository() *MockUserFileRepository {
	return &MockUserFileRepository{}
}

func (m *MockUserFileRepository) Create(ctx context.Context, file *model.UserFile) (*model.UserFile, error) {
	m.createCalls++
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, file)
	}
	return nil, nil
}

func (m *MockUserFileRepository) Update(ctx context.Context, file *model.UserFile) error {
	m.updateCalls++
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, file)
	}
	return nil
}

func (m *MockUserFileRepository) Delete(ctx context.Context, file *model.UserFile) error {
	m.deleteCalls++
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, file)
	}
	return nil
}

func (m *MockUserFileRepository) GetByID(ctx context.Context, id int64) (*model.UserFile, error) {
	m.getByIDCalls++
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUserFileRepository) GetByUID(ctx context.Context, uid string) (*model.UserFile, error) {
	m.getByUIDCalls++
	if m.GetByUIDFunc != nil {
		return m.GetByUIDFunc(ctx, uid)
	}
	return nil, nil
}

func (m *MockUserFileRepository) GetByUserID(ctx context.Context, userID int64) (*model.UserFile, error) {
	m.getByUserIDCalls++
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserFileRepository) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserFileListFilterParam) (*model.UserFiles, error) {
	m.listCalls++
	if m.ListFunc != nil {
		return m.ListFunc(ctx, pagination, filter)
	}
	return &model.UserFiles{}, nil
}

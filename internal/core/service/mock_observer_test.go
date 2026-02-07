package service

import (
	"context"
	"sync"

	"github.com/adityakw90/service-user/internal/core/domain/signal"
)

// MockAuthObserver is a mock observer for testing auth service.
type MockAuthObserver struct {
	signals []mockSignal[signal.AuthSignal]
	mu      sync.Mutex
}

func NewMockAuthObserver() *MockAuthObserver {
	return &MockAuthObserver{signals: []mockSignal[signal.AuthSignal]{}}
}

func (m *MockAuthObserver) OnSignal(ctx context.Context, sig signal.SignalType, data signal.AuthSignal, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = append(m.signals, mockSignal[signal.AuthSignal]{
		SignalType: sig,
		Data:       data,
		Error:      err,
	})
}

func (m *MockAuthObserver) GetSignals() []mockSignal[signal.AuthSignal] {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signals
}

func (m *MockAuthObserver) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = nil
}

// MockUserObserver is a mock observer for testing user service.
type MockUserObserver struct {
	signals []mockSignal[signal.UserSignal]
	mu      sync.Mutex
}

func NewMockUserObserver() *MockUserObserver {
	return &MockUserObserver{signals: []mockSignal[signal.UserSignal]{}}
}

func (m *MockUserObserver) OnSignal(ctx context.Context, sig signal.SignalType, data signal.UserSignal, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = append(m.signals, mockSignal[signal.UserSignal]{
		SignalType: sig,
		Data:       data,
		Error:      err,
	})
}

func (m *MockUserObserver) GetSignals() []mockSignal[signal.UserSignal] {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signals
}

func (m *MockUserObserver) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = nil
}

// MockDeviceObserver is a mock observer for testing device service.
type MockDeviceObserver struct {
	signals []mockSignal[signal.DeviceSignal]
	mu      sync.Mutex
}

func NewMockDeviceObserver() *MockDeviceObserver {
	return &MockDeviceObserver{signals: []mockSignal[signal.DeviceSignal]{}}
}

func (m *MockDeviceObserver) OnSignal(ctx context.Context, sig signal.SignalType, data signal.DeviceSignal, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = append(m.signals, mockSignal[signal.DeviceSignal]{
		SignalType: sig,
		Data:       data,
		Error:      err,
	})
}

func (m *MockDeviceObserver) GetSignals() []mockSignal[signal.DeviceSignal] {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signals
}

func (m *MockDeviceObserver) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = nil
}

// MockUserFileObserver is a mock observer for testing user file service.
type MockUserFileObserver struct {
	signals []mockSignal[signal.UserFileSignal]
	mu      sync.Mutex
}

func NewMockUserFileObserver() *MockUserFileObserver {
	return &MockUserFileObserver{signals: []mockSignal[signal.UserFileSignal]{}}
}

func (m *MockUserFileObserver) OnSignal(ctx context.Context, sig signal.SignalType, data signal.UserFileSignal, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = append(m.signals, mockSignal[signal.UserFileSignal]{
		SignalType: sig,
		Data:       data,
		Error:      err,
	})
}

func (m *MockUserFileObserver) GetSignals() []mockSignal[signal.UserFileSignal] {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signals
}

func (m *MockUserFileObserver) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = nil
}

// MockPinObserver is a mock observer for testing PIN operations.
type MockPinObserver struct {
	signals []mockSignal[signal.PinSignal]
	mu      sync.Mutex
}

func NewMockPinObserver() *MockPinObserver {
	return &MockPinObserver{signals: []mockSignal[signal.PinSignal]{}}
}

func (m *MockPinObserver) OnSignal(ctx context.Context, sig signal.SignalType, data signal.PinSignal, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = append(m.signals, mockSignal[signal.PinSignal]{
		SignalType: sig,
		Data:       data,
		Error:      err,
	})
}

func (m *MockPinObserver) GetSignals() []mockSignal[signal.PinSignal] {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.signals
}

func (m *MockPinObserver) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.signals = nil
}

// mockSignal is a generic type for capturing signals
type mockSignal[T any] struct {
	SignalType signal.SignalType
	Data       T
	Error      error
}

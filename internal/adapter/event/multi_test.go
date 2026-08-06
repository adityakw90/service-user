package event

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	mocksevent "github.com/adityakw90/service-user/mocks/event"
	"github.com/stretchr/testify/mock"
)

func TestMultiEventPublisher_Name(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()
	tracer := newMockTracer()
	publisher := NewMultiEventPublisher(logger, tracer)

	if got := publisher.Name(); got != "MultiEventPublisher" {
		t.Errorf("Name() = %v, want %v", got, "MultiEventPublisher")
	}
}

func TestMultiEventPublisher_Publish(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(t *testing.T) []*mocksevent.MockEventPublisher
		eventType     event.EventType
		eventData     any
		wantErr       bool
		validateCalls func(t *testing.T, publishers []*mocksevent.MockEventPublisher)
		validateLogs  func(t *testing.T, logger *mockLogger)
	}{
		{
			name: "Empty publishers list - returns nil",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				return []*mocksevent.MockEventPublisher{}
			},
			eventType: event.EventLogin,
			eventData: event.EventLoginData{Identifier: "test@example.com"},
			wantErr:   false,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 0 {
					t.Errorf("Expected no error logs, got %d", len(logger.errorMessages))
				}
			},
		},
		{
			name: "Single publisher success - returns nil",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub := mocksevent.NewMockEventPublisher(t)
				mockPub.On("Name").Maybe().Return("publisher-1")
				mockPub.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)
				return []*mocksevent.MockEventPublisher{mockPub}
			},
			eventType: event.EventLogin,
			eventData: event.EventLoginData{Identifier: "test@example.com"},
			wantErr:   false,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
				publishers[0].AssertCalled(t, "Publish", mock.Anything, mock.Anything)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 0 {
					t.Errorf("Expected no error logs, got %d", len(logger.errorMessages))
				}
			},
		},
		{
			name: "Single publisher failure - returns error",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub := mocksevent.NewMockEventPublisher(t)
				mockPub.On("Name").Maybe().Return("publisher-1")
				mockPub.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("publish failed"))
				return []*mocksevent.MockEventPublisher{mockPub}
			},
			eventType: event.EventLogin,
			eventData: event.EventLoginData{Identifier: "test@example.com"},
			wantErr:   true,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
				publishers[0].AssertCalled(t, "Publish", mock.Anything, mock.Anything)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 1 {
					t.Errorf("Expected 1 error log, got %d", len(logger.errorMessages))
				}
				entry := logger.errorMessages[0]
				if entry.message != "failed to publish event" {
					t.Errorf("Expected 'failed to publish event' message, got %s", entry.message)
				}
				if entry.fields["publisher.name"] != "publisher-1" {
					t.Errorf("Expected publisher name 'publisher-1', got %v", entry.fields["publisher.name"])
				}
				if entry.fields["event.type"] != event.EventLogin {
					t.Errorf("Expected event type %v, got %v", event.EventLogin, entry.fields["event.type"])
				}
			},
		},
		{
			name: "Multiple publishers all succeed - returns nil",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				mockPub3 := mocksevent.NewMockEventPublisher(t)
				mockPub3.On("Name").Maybe().Return("publisher-3")
				mockPub3.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}
			},
			eventType: event.EventUserCreated,
			eventData: struct{ UserUID string }{UserUID: "user-123"},
			wantErr:   false,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
				for i, p := range publishers {
					if !p.AssertCalled(t, "Publish", mock.Anything, mock.Anything) {
						t.Errorf("Publisher %d: Publish was not called", i)
					}
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 0 {
					t.Errorf("Expected no error logs, got %d", len(logger.errorMessages))
				}
			},
		},
		{
			name: "Multiple publishers - first fails, others succeed - returns first error",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("first error"))

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				mockPub3 := mocksevent.NewMockEventPublisher(t)
				mockPub3.On("Name").Maybe().Return("publisher-3")
				mockPub3.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}
			},
			eventType: event.EventPINVerify,
			eventData: event.EventPinVerifyData{UserUID: "user-123", Success: true},
			wantErr:   true,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
				for i, p := range publishers {
					if !p.AssertCalled(t, "Publish", mock.Anything, mock.Anything) {
						t.Errorf("Publisher %d: Publish was not called", i)
					}
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 1 {
					t.Errorf("Expected 1 error log, got %d", len(logger.errorMessages))
				}
				entry := logger.errorMessages[0]
				if entry.fields["publisher.name"] != "publisher-1" {
					t.Errorf("Expected publisher name 'publisher-1', got %v", entry.fields["publisher.name"])
				}
			},
		},
		{
			name: "Multiple publishers - middle fails, others succeed - logs error for failed",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("middle error"))

				mockPub3 := mocksevent.NewMockEventPublisher(t)
				mockPub3.On("Name").Maybe().Return("publisher-3")
				mockPub3.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}
			},
			eventType: event.EventTokenRefresh,
			eventData: event.EventTokenRefreshData{Identifier: "test@example.com"},
			wantErr:   true,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
				for i, p := range publishers {
					if !p.AssertCalled(t, "Publish", mock.Anything, mock.Anything) {
						t.Errorf("Publisher %d: Publish was not called", i)
					}
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 1 {
					t.Errorf("Expected 1 error log, got %d", len(logger.errorMessages))
				}
				entry := logger.errorMessages[0]
				if entry.fields["publisher.name"] != "publisher-2" {
					t.Errorf("Expected publisher name 'publisher-2', got %v", entry.fields["publisher.name"])
				}
			},
		},
		{
			name: "Multiple publishers all fail - returns first error, logs all errors",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("error 1"))

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("error 2"))

				mockPub3 := mocksevent.NewMockEventPublisher(t)
				mockPub3.On("Name").Maybe().Return("publisher-3")
				mockPub3.EXPECT().Publish(mock.Anything, mock.Anything).Return(errors.New("error 3"))

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}
			},
			eventType: event.EventLoginFailed,
			eventData: event.EventLoginFailedData{Identifier: "bad@example.com", FailureReason: "invalid"},
			wantErr:   true,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
				for i, p := range publishers {
					if !p.AssertCalled(t, "Publish", mock.Anything, mock.Anything) {
						t.Errorf("Publisher %d: Publish was not called", i)
					}
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 3 {
					t.Errorf("Expected 3 error logs, got %d", len(logger.errorMessages))
				}
				for i, entry := range logger.errorMessages {
					if entry.message != "failed to publish event" {
						t.Errorf("Log %d: expected 'failed to publish event', got %s", i, entry.message)
					}
					expectedName := fmt.Sprintf("publisher-%d", i+1)
					if entry.fields["publisher.name"] != expectedName {
						t.Errorf("Log %d: expected publisher name '%s', got %v", i, expectedName, entry.fields["publisher.name"])
					}
				}
			},
		},
		{
			name: "Context propagation - all publishers receive context with span",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Publish(mock.Anything, mock.Anything).Return(nil)

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2}
			},
			eventType: event.EventUserUpdated,
			eventData: struct{ UserUID string }{UserUID: "user-456"},
			wantErr:   false,
			validateCalls: func(t *testing.T, publishers []*mocksevent.MockEventPublisher) {
				for i, p := range publishers {
					if !p.AssertCalled(t, "Publish", mock.Anything, mock.Anything) {
						t.Errorf("Publisher %d: Publish was not called", i)
					}
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 0 {
					t.Errorf("Expected no error logs, got %d", len(logger.errorMessages))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := newMockLogger()
			tracer := newMockTracer()

			mockPublishers := tt.setupMocks(t)
			publisherSlice := make([]portEvent.EventPublisher, len(mockPublishers))
			for i, p := range mockPublishers {
				publisherSlice[i] = p
			}

			mp := NewMultiEventPublisher(logger, tracer, publisherSlice...)
			ctx := context.Background()

			err := mp.Publish(ctx, event.Message{Type: tt.eventType, Entity: event.Entity{ID: "entity-1", Type: "user"}, Metadata: tt.eventData})

			if (err != nil) != tt.wantErr {
				t.Errorf("Publish() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.validateCalls != nil {
				tt.validateCalls(t, mockPublishers)
			}

			if tt.validateLogs != nil {
				tt.validateLogs(t, logger)
			}
		})
	}
}

func TestMultiEventPublisher_Close(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(t *testing.T) []*mocksevent.MockEventPublisher
		wantErr      bool
		validateLogs func(t *testing.T, logger *mockLogger)
	}{
		{
			name: "Empty publishers list - returns nil",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				return []*mocksevent.MockEventPublisher{}
			},
			wantErr: false,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 0 {
					t.Errorf("Expected no error logs, got %d", len(logger.errorMessages))
				}
			},
		},
		{
			name: "All close successfully - returns nil",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Close().Return(nil)

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Close().Return(nil)

				mockPub3 := mocksevent.NewMockEventPublisher(t)
				mockPub3.On("Name").Maybe().Return("publisher-3")
				mockPub3.EXPECT().Close().Return(nil)

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}
			},
			wantErr: false,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 0 {
					t.Errorf("Expected no error logs, got %d", len(logger.errorMessages))
				}
			},
		},
		{
			name: "First close fails, others succeed - returns first error",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Close().Return(errors.New("close failed"))

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Close().Return(nil)

				mockPub3 := mocksevent.NewMockEventPublisher(t)
				mockPub3.On("Name").Maybe().Return("publisher-3")
				mockPub3.EXPECT().Close().Return(nil)

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 1 {
					t.Errorf("Expected 1 error log, got %d", len(logger.errorMessages))
				}
				entry := logger.errorMessages[0]
				if entry.message != "failed to close publisher" {
					t.Errorf("Expected 'failed to close publisher' message, got %s", entry.message)
				}
				if entry.fields["publisher.name"] != "publisher-1" {
					t.Errorf("Expected publisher name 'publisher-1', got %v", entry.fields["publisher.name"])
				}
			},
		},
		{
			name: "All close fail - returns first error, logs all errors",
			setupMocks: func(t *testing.T) []*mocksevent.MockEventPublisher {
				mockPub1 := mocksevent.NewMockEventPublisher(t)
				mockPub1.On("Name").Maybe().Return("publisher-1")
				mockPub1.EXPECT().Close().Return(errors.New("close error 1"))

				mockPub2 := mocksevent.NewMockEventPublisher(t)
				mockPub2.On("Name").Maybe().Return("publisher-2")
				mockPub2.EXPECT().Close().Return(errors.New("close error 2"))

				mockPub3 := mocksevent.NewMockEventPublisher(t)
				mockPub3.On("Name").Maybe().Return("publisher-3")
				mockPub3.EXPECT().Close().Return(errors.New("close error 3"))

				return []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				if len(logger.errorMessages) != 3 {
					t.Errorf("Expected 3 error logs, got %d", len(logger.errorMessages))
				}
				for i, entry := range logger.errorMessages {
					if entry.message != "failed to close publisher" {
						t.Errorf("Log %d: expected 'failed to close publisher', got %s", i, entry.message)
					}
					expectedName := fmt.Sprintf("publisher-%d", i+1)
					if entry.fields["publisher.name"] != expectedName {
						t.Errorf("Log %d: expected publisher name '%s', got %v", i, expectedName, entry.fields["publisher.name"])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := newMockLogger()
			tracer := newMockTracer()

			mockPublishers := tt.setupMocks(t)
			publisherSlice := make([]portEvent.EventPublisher, len(mockPublishers))
			for i, p := range mockPublishers {
				publisherSlice[i] = p
			}

			mp := NewMultiEventPublisher(logger, tracer, publisherSlice...)

			err := mp.Close()

			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}

			for i, p := range mockPublishers {
				if !p.AssertCalled(t, "Close") {
					t.Errorf("Publisher %d: Close was not called", i)
				}
			}

			if tt.validateLogs != nil {
				tt.validateLogs(t, logger)
			}
		})
	}
}

func TestMultiEventPublisher_Publish_ConcurrentCalls(t *testing.T) {
	t.Parallel()

	logger := newMockLogger()
	tracer := newMockTracer()

	mockPub1 := mocksevent.NewMockEventPublisher(t)
	mockPub1.On("Name").Return("publisher-1").Maybe()
	mockPub1.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockPub2 := mocksevent.NewMockEventPublisher(t)
	mockPub2.On("Name").Return("publisher-2").Maybe()
	mockPub2.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockPub3 := mocksevent.NewMockEventPublisher(t)
	mockPub3.On("Name").Return("publisher-3").Maybe()
	mockPub3.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockPublishers := []*mocksevent.MockEventPublisher{mockPub1, mockPub2, mockPub3}

	publisherSlice := make([]portEvent.EventPublisher, len(mockPublishers))
	for i, p := range mockPublishers {
		publisherSlice[i] = p
	}

	mp := NewMultiEventPublisher(logger, tracer, publisherSlice...)
	ctx := context.Background()

	var wg sync.WaitGroup
	numGoroutines := 10
	callsPerGoroutine := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				_ = mp.Publish(ctx, event.Message{Type: event.EventLogin, Entity: event.Entity{ID: fmt.Sprintf("user-%d-%d", id, j), Type: "user"}, Metadata: event.EventLoginData{Identifier: fmt.Sprintf("user-%d-%d", id, j)}})
			}
		}(i)
	}
	wg.Wait()

	for i, p := range mockPublishers {
		callCount := len(p.Calls)
		expectedCalls := numGoroutines * callsPerGoroutine
		if callCount != expectedCalls {
			t.Errorf("Publisher %d: expected %d calls, got %d", i, expectedCalls, callCount)
		}
	}

	if len(logger.errorMessages) != 0 {
		t.Errorf("Expected no error logs, got %d", len(logger.errorMessages))
	}
}

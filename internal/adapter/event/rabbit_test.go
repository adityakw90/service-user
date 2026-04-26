package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/stretchr/testify/assert"
)

type mockRabbitConn struct {
	publishWithConfirmCalled            bool
	publishWithConfirmArgCtx            context.Context
	publishWithConfirmArgExchange       string
	publishWithConfirmArgRoutingKey     string
	publishWithConfirmArgContentType    string
	publishWithConfirmArgHeaders        map[string]string
	publishWithConfirmArgBody           []byte
	publishWithConfirmArgDeliveryMode   *uint8
	publishWithConfirmArgPriority       *uint8
	publishWithConfirmArgTimestamp      *time.Time
	publishWithConfirmArgExpiration     *time.Duration
	publishWithConfirmArgMaxRetries     int
	publishWithConfirmArgRetryInterval  time.Duration
	publishWithConfirmArgConfirmTimeout time.Duration
	publishWithConfirmReturnErr         error
	publishWithConfirmCheckContext      bool

	closeCalled    bool
	closeReturnErr error
}

func (m *mockRabbitConn) PublishWithConfirm(
	ctx context.Context,
	exchange string,
	routingKey string,
	contentType string,
	headers map[string]string,
	body []byte,
	deliveryMode *uint8,
	priority *uint8,
	timestamp *time.Time,
	expiration *time.Duration,
	maxRetries int,
	retryInterval time.Duration,
	confirmTimeout time.Duration,
) error {
	if m.publishWithConfirmCheckContext && ctx.Err() != nil {
		return ctx.Err()
	}
	m.publishWithConfirmCalled = true
	m.publishWithConfirmArgCtx = ctx
	m.publishWithConfirmArgExchange = exchange
	m.publishWithConfirmArgRoutingKey = routingKey
	m.publishWithConfirmArgContentType = contentType
	m.publishWithConfirmArgHeaders = headers
	m.publishWithConfirmArgBody = body
	m.publishWithConfirmArgDeliveryMode = deliveryMode
	m.publishWithConfirmArgPriority = priority
	m.publishWithConfirmArgTimestamp = timestamp
	m.publishWithConfirmArgExpiration = expiration
	m.publishWithConfirmArgMaxRetries = maxRetries
	m.publishWithConfirmArgRetryInterval = retryInterval
	m.publishWithConfirmArgConfirmTimeout = confirmTimeout
	return m.publishWithConfirmReturnErr
}

func (m *mockRabbitConn) Close() error {
	m.closeCalled = true
	return m.closeReturnErr
}

func TestNewRabbitmqPublisher(t *testing.T) {
	mockConn := &mockRabbitConn{}
	config := RabbitmqPublisherConfig{
		Exchange:         "test-exchange",
		RoutingKeyPrefix: "user.service",
		ConfirmTimeout:   5 * time.Second,
		MaxRetries:       3,
		RetryInterval:    500 * time.Millisecond,
	}
	logger := newMockLogger()
	tracer := newMockTracer()

	publisher := NewRabbitmqPublisher(mockConn, config, logger, tracer)

	rabbitPub, ok := publisher.(*RabbitmqPublisher)
	assert.True(t, ok, "Publisher should be *RabbitmqPublisher type")
	assert.Equal(t, mockConn, rabbitPub.conn)
	assert.Equal(t, "test-exchange", rabbitPub.exchange)
	assert.Equal(t, "user.service.", rabbitPub.routingKeyPrefix)
	assert.Equal(t, 5*time.Second, rabbitPub.confirmTimeout)
	assert.Equal(t, 3, rabbitPub.maxRetries)
	assert.Equal(t, 500*time.Millisecond, rabbitPub.retryInterval)
	assert.Equal(t, logger, rabbitPub.logger)
	assert.Equal(t, tracer, rabbitPub.tracer)
}

func TestRabbitmqPublisher_Name(t *testing.T) {
	mockConn := &mockRabbitConn{}
	config := RabbitmqPublisherConfig{}
	logger := newMockLogger()
	tracer := newMockTracer()

	publisher := NewRabbitmqPublisher(mockConn, config, logger, tracer)

	assert.Equal(t, "RabbitmqPublisher", publisher.Name())
}

func TestRabbitmqPublisher_Publish(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		eventType   event.EventType
		eventData   any
		config      RabbitmqPublisherConfig
		setupMock   func(*mockRabbitConn)
		wantErr     bool
		errContains string
		validate    func(*testing.T, *mockRabbitConn)
	}{
		{
			name: "Happy Path - EventLogin with full context",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test-client", "user-123", "user")
			},
			eventType: event.EventLogin,
			eventData: event.EventLoginData{
				Identifier:     "test@example.com",
				IdentifierType: "email",
				UserUID:        "uid-123",
				UserName:       "Test User",
				DeviceUID:      "device-123",
				DeviceName:     "iPhone",
				IPAddress:      "192.168.1.1",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "test-exchange",
				RoutingKeyPrefix: "user.service",
				ConfirmTimeout:   5 * time.Second,
				MaxRetries:       3,
				RetryInterval:    500 * time.Millisecond,
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled, "PublishWithConfirm should be called")
				assert.Equal(t, "test-exchange", m.publishWithConfirmArgExchange)
				assert.Equal(t, "user.service.auth.login", m.publishWithConfirmArgRoutingKey)
				assert.Equal(t, "application/cloudevents+json", m.publishWithConfirmArgContentType)

				assert.NotNil(t, m.publishWithConfirmArgHeaders)
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_type"])
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_source"])
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_id"])
				assert.NotEmpty(t, m.publishWithConfirmArgHeaders["ce_specversion"])
				assert.Equal(t, "auth.login", m.publishWithConfirmArgHeaders["ce_type"])
				assert.Equal(t, Source, m.publishWithConfirmArgHeaders["ce_source"])
				assert.Equal(t, SpecVersion, m.publishWithConfirmArgHeaders["ce_specversion"])

				assert.NotNil(t, m.publishWithConfirmArgBody)
				assert.NotEmpty(t, m.publishWithConfirmArgBody)

				assert.Equal(t, uint8(2), *m.publishWithConfirmArgDeliveryMode)
				assert.Equal(t, uint8(10), *m.publishWithConfirmArgPriority)
				assert.Nil(t, m.publishWithConfirmArgTimestamp)
				assert.Nil(t, m.publishWithConfirmArgExpiration)
				assert.Equal(t, 3, m.publishWithConfirmArgMaxRetries)
				assert.Equal(t, 500*time.Millisecond, m.publishWithConfirmArgRetryInterval)
				assert.Equal(t, 5*time.Second, m.publishWithConfirmArgConfirmTimeout)
			},
		},
		{
			name: "Happy Path - EventUserCreated",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin-client", "admin-456", "admin")
			},
			eventType: event.EventUserCreated,
			eventData: event.EventUserCreatedData{
				UserUID:  "new-user-uid",
				ActorUID: "admin-456",
				Username: "newuser",
				Email:    "newuser@example.com",
				Status:   "active",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "user-events",
				RoutingKeyPrefix: "user",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "user-events", m.publishWithConfirmArgExchange)
				assert.Equal(t, "user.user.created", m.publishWithConfirmArgRoutingKey)
				assert.Equal(t, "user.created", m.publishWithConfirmArgHeaders["ce_type"])
			},
		},
		{
			name: "Happy Path - Context without client/actor uses defaults",
			setupCtx: func() context.Context {
				return context.Background()
			},
			eventType: event.EventUserDeleted,
			eventData: event.EventUserDeletedData{
				UserUID:  "deleted-user-uid",
				ActorUID: "admin-789",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "user.deleted", m.publishWithConfirmArgRoutingKey)
			},
		},
		{
			name: "Happy Path - EventUserUpdatePassword",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "user-789", "user")
			},
			eventType: event.EventUserUpdatePassword,
			eventData: event.EventUserUpdatePasswordData{
				UserUID:  "user-789",
				ActorUID: "user-789",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "security-events",
				RoutingKeyPrefix: "auth",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "auth.user.update_password", m.publishWithConfirmArgRoutingKey)
			},
		},
		{
			name: "Happy Path - EventUserCreatePin",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "mobile", "user-101", "user")
			},
			eventType: event.EventUserCreatePin,
			eventData: event.EventUserCreatePinData{
				UserUID:  "user-101",
				ActorUID: "user-101",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "user",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
			validate: func(t *testing.T, m *mockRabbitConn) {
				assert.True(t, m.publishWithConfirmCalled)
				assert.Equal(t, "user.user.create_pin", m.publishWithConfirmArgRoutingKey)
			},
		},
		{
			name: "Happy Path - EventUserUpdateProfile",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "web", "user-202", "user")
			},
			eventType: event.EventUserUpdateProfile,
			eventData: event.EventUserUpdateProfileData{
				UserUID:  "user-202",
				ActorUID: "user-202",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "user",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventUserRevokeDevice",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin", "user-303", "admin")
			},
			eventType: event.EventUserRevokeDevice,
			eventData: event.EventUserRevokeDeviceData{
				UserUID:   "user-303",
				ActorUID:  "user-303",
				DeviceUID: "device-303",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "events",
				RoutingKeyPrefix: "device",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventUserFileCreated",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "user-404", "user")
			},
			eventType: event.EventUserFileCreated,
			eventData: event.EventUserFileCreatedData{
				UserUID: "user-404",
				FileUID: "file-404",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "file-events",
				RoutingKeyPrefix: "file",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventUserFileUpdated",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "user-505", "user")
			},
			eventType: event.EventUserFileUpdated,
			eventData: event.EventUserFileUpdatedData{
				UserUID: "user-505",
				FileUID: "file-505",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "file-events",
				RoutingKeyPrefix: "file",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Happy Path - EventUserFileDeleted",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "api", "user-606", "user")
			},
			eventType: event.EventUserFileDeleted,
			eventData: event.EventUserFileDeletedData{
				UserUID: "user-606",
				FileUID: "file-606",
			},
			config: RabbitmqPublisherConfig{
				Exchange:         "file-events",
				RoutingKeyPrefix: "file",
			},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Error - RabbitMQ publish failure",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test-client", "user-123", "user")
			},
			eventType: event.EventLogin,
			eventData: event.EventLoginData{
				Identifier: "test@example.com",
			},
			config: RabbitmqPublisherConfig{},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = errors.New("connection failed")
			},
			wantErr:     true,
			errContains: "failed to publish message to RabbitMQ",
		},
		{
			name: "Error - Context cancellation",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			eventType: event.EventLogin,
			eventData: event.EventLoginData{
				Identifier: "test@example.com",
			},
			config: RabbitmqPublisherConfig{},
			setupMock: func(m *mockRabbitConn) {
				m.publishWithConfirmReturnErr = nil
				m.publishWithConfirmCheckContext = true
			},
			wantErr:     true,
			errContains: "failed to publish message to RabbitMQ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &mockRabbitConn{}
			tt.setupMock(mockConn)

			logger := newMockLogger()
			tracer := newMockTracer()

			publisher := NewRabbitmqPublisher(mockConn, tt.config, logger, tracer)

			ctx := tt.setupCtx()
			err := publisher.Publish(ctx, tt.eventType, tt.eventData)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.validate != nil {
				tt.validate(t, mockConn)
			}
		})
	}
}

func TestRabbitmqPublisher_Close(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*mockRabbitConn)
		wantErr     bool
		errContains string
	}{
		{
			name: "Happy Path - Successful close",
			setupMock: func(m *mockRabbitConn) {
				m.closeReturnErr = nil
			},
			wantErr: false,
		},
		{
			name: "Error - Close propagates RabbitMQ error",
			setupMock: func(m *mockRabbitConn) {
				m.closeReturnErr = errors.New("connection already closed")
			},
			wantErr:     true,
			errContains: "connection already closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConn := &mockRabbitConn{}
			tt.setupMock(mockConn)

			config := RabbitmqPublisherConfig{}
			logger := newMockLogger()
			tracer := newMockTracer()

			publisher := NewRabbitmqPublisher(mockConn, config, logger, tracer)

			err := publisher.Close()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.True(t, mockConn.closeCalled, "Close should be called on underlying Rabbit connection")
		})
	}
}

func TestRabbitmqPublisher_getRoutingKey(t *testing.T) {
	tests := []struct {
		name            string
		prefix          string
		eventType       string
		expectedRouting string
	}{
		{
			name:            "Standard prefix with event type",
			prefix:          "user.service.",
			eventType:       "auth.login",
			expectedRouting: "user.service.auth.login",
		},
		{
			name:            "Empty prefix",
			prefix:          "",
			eventType:       "user.created",
			expectedRouting: "user.created",
		},
		{
			name:            "Empty event type",
			prefix:          "events",
			eventType:       "",
			expectedRouting: "events",
		},
		{
			name:            "Empty both",
			prefix:          "",
			eventType:       "",
			expectedRouting: "",
		},
		{
			name:            "Single word prefix",
			prefix:          "auth.",
			eventType:       "login",
			expectedRouting: "auth.login",
		},
		{
			name:            "Multi-level event type",
			prefix:          "user.",
			eventType:       "file.upload.complete",
			expectedRouting: "user.file.upload.complete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := &RabbitmqPublisher{
				routingKeyPrefix: tt.prefix,
			}

			routingKey := publisher.getRoutingKey(tt.eventType)

			assert.Equal(t, tt.expectedRouting, routingKey)
		})
	}
}

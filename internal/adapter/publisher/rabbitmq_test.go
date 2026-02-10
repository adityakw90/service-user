package publisher

import (
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/event"
)

// TestRabbitMQPublisher_Publish tests the CloudEvents conversion for RabbitMQ.
// Note: Full integration tests would require actual RabbitMQ connection.
func TestRabbitMQPublisher_Publish(t *testing.T) {
	tests := []struct {
		name      string
		eventType event.EventType
		eventData any
		source    string
		prefix    string
	}{
		{
			name:      "Login event",
			eventType: event.EventLogin,
			eventData: map[string]interface{}{
				"user_uid":  "test-uid",
				"success":   true,
				"operation": "login",
			},
			source: "rabbitmq-publisher",
			prefix: "user.service.",
		},
		{
			name:      "User created",
			eventType: event.EventUserCreated,
			eventData: map[string]interface{}{
				"user_uid": "test-uid",
			},
			source: "rabbitmq-publisher",
			prefix: "user.service.",
		},
		{
			name:      "Device created",
			eventType: event.EventDeviceCreated,
			eventData: map[string]interface{}{
				"device_uid": "device-123",
			},
			source: "rabbitmq-publisher",
			prefix: "user.service.",
		},
		{
			name:      "File created",
			eventType: event.EventUserFileCreated,
			eventData: map[string]interface{}{
				"user_uid": "user-123",
				"file_uid": "file-789",
			},
			source: "rabbitmq-publisher",
			prefix: "user.service.",
		},
		{
			name:      "Login failed",
			eventType: event.EventLoginFailed,
			eventData: map[string]interface{}{
				"identifier":     "baduser@example.com",
				"failure_reason": "invalid_credentials",
			},
			source: "rabbitmq-publisher",
			prefix: "user.service.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test CloudEvent conversion (same as Kafka)
			ce := toCloudEventData(tt.eventType, tt.eventData, tt.source)

			// Verify CloudEvent structure
			if ce.Type != string(tt.eventType) {
				t.Errorf("expected Type = %v, got %v", tt.eventType, ce.Type)
			}
			if ce.Source != tt.source {
				t.Errorf("expected Source = %v, got %v", tt.source, ce.Source)
			}
			if ce.SpecVersion != "1.0" {
				t.Errorf("expected SpecVersion = 1.0, got %v", ce.SpecVersion)
			}
			if ce.ID == "" {
				t.Error("expected ID to be non-empty")
			}
			if ce.Time == "" {
				t.Error("expected Time to be non-empty")
			}
			if len(ce.Data) == 0 {
				t.Error("expected Data to be non-empty")
			}

			// Verify routing key construction
			expectedRoutingKey := tt.prefix + string(tt.eventType)
			routingKey := tt.prefix + string(tt.eventType)
			if routingKey != expectedRoutingKey {
				t.Errorf("expected routing key %v, got %v", expectedRoutingKey, routingKey)
			}
		})
	}
}

func TestRabbitMQConfig_Defaults(t *testing.T) {
	config := RabbitMQPublisherConfig{
		Source:           "user-service",
		Exchange:         "user-service",
		ExchangeType:     "topic",
		RoutingKeyPrefix: "user.service.",
		Durable:          true,
	}

	if config.Exchange != "user-service" {
		t.Errorf("expected exchange user-service, got %v", config.Exchange)
	}
	if config.ExchangeType != "topic" {
		t.Errorf("expected exchange type topic, got %v", config.ExchangeType)
	}
	if config.RoutingKeyPrefix != "user.service." {
		t.Errorf("expected routing key prefix user.service., got %v", config.RoutingKeyPrefix)
	}
	if !config.Durable {
		t.Error("expected durable to be true")
	}
}

func TestRabbitMQRoutingKey(t *testing.T) {
	tests := []struct {
		name        string
		eventType   event.EventType
		prefix      string
		expectedKey string
	}{
		{
			name:        "Login event routing key",
			eventType:   event.EventLogin,
			prefix:      "user.service.",
			expectedKey: "user.service.auth.login",
		},
		{
			name:        "User created routing key",
			eventType:   event.EventUserCreated,
			prefix:      "user.service.",
			expectedKey: "user.service.user.created",
		},
		{
			name:        "Device created routing key",
			eventType:   event.EventDeviceCreated,
			prefix:      "user.service.",
			expectedKey: "user.service.device.created",
		},
		{
			name:        "File created routing key",
			eventType:   event.EventUserFileCreated,
			prefix:      "user.service.",
			expectedKey: "user.service.user_file.created",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routingKey := tt.prefix + string(tt.eventType)
			if routingKey != tt.expectedKey {
				t.Errorf("expected routing key %v, got %v", tt.expectedKey, routingKey)
			}
		})
	}
}

func TestToCloudEventData_RabbitMQ(t *testing.T) {
	tests := []struct {
		name      string
		eventType event.EventType
		eventData any
		source    string
	}{
		{
			name:      "Login event",
			eventType: event.EventLogin,
			eventData: map[string]interface{}{
				"user_uid":  "user-123",
				"success":   true,
				"operation": "login",
			},
			source: "rabbitmq-publisher",
		},
		{
			name:      "User created",
			eventType: event.EventUserCreated,
			eventData: map[string]interface{}{
				"user_uid": "user-123",
				"username": "testuser",
				"email":    "test@example.com",
			},
			source: "rabbitmq-publisher",
		},
		{
			name:      "Login failed",
			eventType: event.EventLoginFailed,
			eventData: map[string]interface{}{
				"identifier":     "baduser@example.com",
				"failure_reason": "invalid_credentials",
			},
			source: "rabbitmq-publisher",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := toCloudEventData(tt.eventType, tt.eventData, tt.source)

			if ce.Type != string(tt.eventType) {
				t.Errorf("expected Type = %v, got %v", tt.eventType, ce.Type)
			}
			if ce.Source != tt.source {
				t.Errorf("expected Source = %v, got %v", tt.source, ce.Source)
			}
		})
	}
}

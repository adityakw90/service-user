package publisher

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/google/uuid"
)

// CloudEvent represents a CloudEvents message.
type CloudEvent struct {
	// Required
	Type        string         `json:"type"`
	Source      string         `json:"source"`
	SpecVersion string         `json:"specversion"`
	ID          string         `json:"id"`
	Time        string         `json:"time"`

	// Data
	Data json.RawMessage `json:"data"`
}

// EventData represents the common structure for event data.
type EventData struct {
	Type          string                 `json:"type"`
	UserUID       string                 `json:"user_uid,omitempty"`
	ActorUID      string                 `json:"actor_uid,omitempty"`
	DeviceUID     string                 `json:"device_uid,omitempty"`
	FileUID       string                 `json:"file_uid,omitempty"`
	Identifier    string                 `json:"identifier,omitempty"`
	Operation     string                 `json:"operation,omitempty"`
	Success       bool                   `json:"success"`
	FailureReason string                 `json:"failure_reason,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ToCloudEvent converts any domain event data to CloudEvents format.
// This is a legacy function that marshals the event data directly.
// Prefer using toCloudEventData for new code.
func ToCloudEvent(e any, source string) (CloudEvent, error) {
	ce := CloudEvent{
		Source:      source,
		SpecVersion: "1.0",
		ID:          generateEventID(),
		Time:        time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal event data to JSON
	data, err := json.Marshal(e)
	if err != nil {
		return CloudEvent{}, fmt.Errorf("failed to marshal event data: %w", err)
	}
	ce.Data = data

	return ce, nil
}

// toCloudEventData converts eventType and eventData to CloudEvents format.
// This is the simplified version used by the new EventPublisher interface.
func toCloudEventData(eventType event.EventType, eventData any, source string) CloudEvent {
	ce := CloudEvent{
		Type:        string(eventType),
		Source:      source,
		SpecVersion: "1.0",
		ID:          generateEventID(),
		Time:        time.Now().UTC().Format(time.RFC3339),
	}

	// Marshal event data to JSON
	data, err := json.Marshal(eventData)
	if err != nil {
		// If marshaling fails, wrap in error structure
		data, _ = json.Marshal(map[string]interface{}{
			"error": fmt.Sprintf("failed to marshal event data: %v", err),
		})
	}
	ce.Data = data

	return ce
}

// generateEventID generates a unique event ID using UUID.
func generateEventID() string {
	return uuid.New().String()
}

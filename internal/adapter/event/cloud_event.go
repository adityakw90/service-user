package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/google/uuid"
)

const (
	Source      = "service-user"
	SpecVersion = "1.0"
)

type CloudEvent struct {
	ID          string         `json:"id"` // event id
	Source      string         `json:"source"`
	SpecVersion string         `json:"specversion"`
	Type        string         `json:"type"`
	Time        time.Time      `json:"time"`
	Data        CloudEventData `json:"data"`
}

type CloudEventData struct {
	Client    string          `json:"client"`
	ActorId   string          `json:"actor_id"`
	ActorType string          `json:"actor_type"`
	MetaData  json.RawMessage `json:"metadata"`
}

func NewCloudEvent(ctx context.Context, eventType event.EventType, eventData any) CloudEvent {
	clientName := util.GetClientName(ctx)
	actorId, actorType := util.GetActor(ctx)
	metadata, err := json.Marshal(eventData)
	if err != nil {
		// If marshaling fails, wrap in error structure
		metadata, _ = json.Marshal(map[string]interface{}{
			"error": fmt.Sprintf("failed to marshal event data: %v", err),
		})
	}
	return CloudEvent{
		ID:          uuid.New().String(),
		Source:      Source,
		SpecVersion: SpecVersion,
		Type:        string(eventType),
		Time:        time.Now().UTC(),
		Data: CloudEventData{
			ActorId:   actorId,
			ActorType: actorType,
			Client:    clientName,
			MetaData:  metadata,
		},
	}
}

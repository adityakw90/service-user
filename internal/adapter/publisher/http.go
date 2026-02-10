package publisher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
)

// HTTPPublisher publishes events via HTTP to a CloudEvents endpoint.
type HTTPPublisher struct {
	client *http.Client
	config HttpPublisherConfig
}

// HttpPublisherConfig holds configuration for the HTTP event publisher.
type HttpPublisherConfig struct {
	Endpoint string
	Source   string
	Timeout  time.Duration
}

// NewHTTPPublisher creates a new HTTP event publisher.
func NewHTTPPublisher(config HttpPublisherConfig) portEvent.EventPublisher {
	return &HTTPPublisher{
		client: &http.Client{
			Timeout: config.Timeout,
		},
		config: config,
	}
}

// Publish publishes an event via HTTP.
func (p *HTTPPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	// Convert to CloudEvent format
	ce := toCloudEventData(eventType, eventData, p.config.Source)

	body, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.config.Endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("Ce-Type", ce.Type)
	req.Header.Set("Ce-Source", ce.Source)
	req.Header.Set("Ce-Specversion", ce.SpecVersion)
	req.Header.Set("Ce-ID", ce.ID)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("event publish failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Close closes the publisher connection.
func (p *HTTPPublisher) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

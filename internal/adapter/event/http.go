package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HTTPPublisher publishes events via HTTP to a CloudEvents endpoint.
type HTTPPublisher struct {
	client *http.Client
	url    string
	logger monitoring.Logger
	tracer monitoring.Tracer
}

// NewHTTPPublisher creates a new HTTP event publisher.
func NewHTTPPublisher(
	url string,
	timeout time.Duration,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portEvent.EventPublisher {
	return &HTTPPublisher{
		client: &http.Client{
			Timeout: timeout,
		},
		url:    url,
		logger: logger,
		tracer: tracer,
	}
}

func (p *HTTPPublisher) Name() string {
	return "HTTPPublisher"
}

// Publish publishes an event via HTTP.
func (p *HTTPPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	newCtx, span := p.tracer.StartSpan(ctx, "HTTPPublisher.Publish")
	defer span.End()

	// Convert to CloudEvent format
	ce := NewCloudEvent(ctx, eventType, eventData)
	span.AddEvent(string(eventType),
		trace.WithAttributes(
			attribute.String("type", ce.Type),
			attribute.String("source", ce.Source),
			attribute.String("specversion", ce.SpecVersion),
			attribute.String("id", ce.ID),
		),
	)

	body, err := json.Marshal(ce.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	req, err := http.NewRequestWithContext(newCtx, "POST", p.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/cloudevents+json")
	req.Header.Set("Ce-Type", ce.Type)
	req.Header.Set("Ce-Source", ce.Source)
	req.Header.Set("Ce-Specversion", ce.SpecVersion)
	req.Header.Set("Ce-ID", ce.ID)
	req.Header.Set("Client", ce.Data.Client)
	req.Header.Set("Actor-Id", ce.Data.ActorId)
	req.Header.Set("Actor-Type", ce.Data.ActorType)

	// inject trace header
	md := p.tracer.InjectContext(newCtx)
	for k, v := range md {
		req.Header.Set(k, v[0])
	}

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

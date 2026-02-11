package publisher

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
)

// mockPublisher is a mock implementation of EventPublisher for testing.
type mockPublisher struct {
	publishFunc func(ctx context.Context, eventType event.EventType, eventData any) error
	closeFunc   func() error
	published   []mockPublishedEvent
	mu          sync.Mutex
}

type mockPublishedEvent struct {
	eventType event.EventType
	eventData any
}

func (m *mockPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	m.mu.Lock()
	m.published = append(m.published, mockPublishedEvent{eventType: eventType, eventData: eventData})
	m.mu.Unlock()
	if m.publishFunc != nil {
		return m.publishFunc(ctx, eventType, eventData)
	}
	return nil
}

func (m *mockPublisher) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestAsyncPublisher_Publish(t *testing.T) {
	tests := []struct {
		name        string
		queueSize   int
		batchSize   int
		eventCount  int
		wantPublish bool
		wantError   bool
	}{
		{
			name:        "Single event published",
			queueSize:   10,
			batchSize:   5,
			eventCount:  1,
			wantPublish: true,
			wantError:   false,
		},
		{
			name:        "Multiple events published in batch",
			queueSize:   10,
			batchSize:   3,
			eventCount:  3,
			wantPublish: true,
			wantError:   false,
		},
		{
			name:        "Queue full blocks and eventually publishes",
			queueSize:   2,
			batchSize:   5,
			eventCount:  5,
			wantPublish: true,
			wantError:   false, // Now blocks instead of dropping
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPub := &mockPublisher{published: make([]mockPublishedEvent, 0)}

			pub := NewAsyncPublisher(mockPub, AsyncPublisherConfig{
				WorkerCount:  1,
				QueueSize:    tt.queueSize,
				BatchSize:    tt.batchSize,
				BatchTimeout: 100 * time.Millisecond,
			})
			defer pub.Close()

			ctx := context.Background()
			hadError := false

			for i := 0; i < tt.eventCount; i++ {
				err := pub.Publish(ctx, event.EventLogin, event.EventLoginData{
					Identifier:     "test@example.com",
					IdentifierType: "email",
				})
				if err != nil {
					hadError = true
				}
			}

			// Wait for batch processing
			time.Sleep(200 * time.Millisecond)

			if tt.wantError && !hadError {
				t.Errorf("Expected error when queue is full, got none")
			}

			mockPub.mu.Lock()
			publishedCount := len(mockPub.published)
			mockPub.mu.Unlock()
			if tt.wantPublish && publishedCount == 0 {
				t.Errorf("Expected events to be published, got none")
			}
		})
	}
}

func TestAsyncPublisher_BatchAccumulator(t *testing.T) {
	mockPub := &mockPublisher{published: make([]mockPublishedEvent, 0)}

	pub := NewAsyncPublisher(mockPub, AsyncPublisherConfig{
		WorkerCount:  1,
		QueueSize:    100,
		BatchSize:    5,
		BatchTimeout: 50 * time.Millisecond,
	})
	defer pub.Close()

	ctx := context.Background()

	// Publish fewer events than batch size
	for i := 0; i < 3; i++ {
		err := pub.Publish(ctx, event.EventLogin, event.EventLoginData{
			Identifier:     "test@example.com",
			IdentifierType: "email",
		})
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	// Wait for batch timeout
	time.Sleep(100 * time.Millisecond)

	// Verify events were published after timeout
	mockPub.mu.Lock()
	publishedCount := len(mockPub.published)
	mockPub.mu.Unlock()
	if publishedCount != 3 {
		t.Errorf("Expected 3 events to be published after timeout, got %d", publishedCount)
	}
}

func TestAsyncPublisher_Close(t *testing.T) {
	mockPub := &mockPublisher{published: make([]mockPublishedEvent, 0)}

	pub := NewAsyncPublisher(mockPub, AsyncPublisherConfig{
		WorkerCount:  1,
		QueueSize:    100,
		BatchSize:    10,
		BatchTimeout: 1 * time.Second,
	})

	ctx := context.Background()

	// Publish enough events to trigger a batch
	for i := 0; i < 10; i++ {
		err := pub.Publish(ctx, event.EventLogin, event.EventLoginData{
			Identifier:     "test@example.com",
			IdentifierType: "email",
		})
		if err != nil {
			t.Fatalf("Publish() error = %v", err)
		}
	}

	// Wait a bit for batch processing
	time.Sleep(100 * time.Millisecond)

	// Close should flush pending events
	err := pub.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Verify events were published (at least one batch)
	mockPub.mu.Lock()
	publishedCount := len(mockPub.published)
	mockPub.mu.Unlock()
	if publishedCount == 0 {
		t.Errorf("Expected events to be published after close, got 0")
	}
}

package publisher

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
)

// AsyncPublisher wraps an EventPublisher to provide async, non-blocking publishing
// with batching and worker pool.
type AsyncPublisher struct {
	underlying portEvent.EventPublisher
	queue      chan *eventWrapper
	stopCh     chan struct{}
	wg         sync.WaitGroup

	config  AsyncPublisherConfig
	metrics *asyncMetrics
}

// eventWrapper holds the event type and data for queuing.
type eventWrapper struct {
	eventType event.EventType
	eventData any
	queuedAt  time.Time // For latency tracking
}

// AsyncPublisherConfig holds configuration for the async publisher.
type AsyncPublisherConfig struct {
	WorkerCount  int
	QueueSize    int
	BatchSize    int
	BatchTimeout time.Duration
	// Retry settings for failed publishes
	MaxRetries    int
	RetryInterval time.Duration
}

// asyncMetrics holds atomic metrics.
type asyncMetrics struct {
	queuedEvents     atomic.Int64
	publishedEvents  atomic.Int64
	failedEvents     atomic.Int64
	droppedEvents    atomic.Int64 // Queue full drops (should be 0 with blocking behavior)
	retriedEvents    atomic.Int64 // Number of events that were retried
	currentQueueSize atomic.Int64

	// Latency tracking (in milliseconds)
	totalLatencyMs atomic.Int64
	latencyCount   atomic.Int64
}

// NewAsyncPublisher creates a new async wrapper for any EventPublisher.
func NewAsyncPublisher(
	underlying portEvent.EventPublisher,
	config AsyncPublisherConfig,
) portEvent.EventPublisher {
	if config.WorkerCount <= 0 {
		config.WorkerCount = 2
	}
	if config.QueueSize <= 0 {
		config.QueueSize = 1000
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 50
	}
	if config.BatchTimeout <= 0 {
		config.BatchTimeout = 5 * time.Second
	}
	// Retry defaults
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.RetryInterval <= 0 {
		config.RetryInterval = 2 * time.Second
	}

	p := &AsyncPublisher{
		underlying: underlying,
		queue:      make(chan *eventWrapper, config.QueueSize),
		stopCh:     make(chan struct{}),
		config:     config,
		metrics:    &asyncMetrics{},
	}

	// Start worker pool
	for i := 0; i < config.WorkerCount; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	return p
}

// Publish is non-blocking, queues event for background publishing.
// Blocks if the queue is full (instead of dropping events).
func (p *AsyncPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	wrapper := &eventWrapper{
		eventType: eventType,
		eventData: eventData,
		queuedAt:  time.Now(),
	}

	select {
	case p.queue <- wrapper:
		p.metrics.queuedEvents.Add(1)
		p.metrics.currentQueueSize.Add(1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
	// Note: No default case - we block when queue is full instead of dropping events
}

// worker processes events from the queue.
// Uses a timer that only starts when events are queued to avoid idle wake-ups.
func (p *AsyncPublisher) worker(int) {
	defer p.wg.Done()

	batch := make([]*eventWrapper, 0, p.config.BatchSize)
	// Create timer but don't start it yet - only start when we have events
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}

		// Stop timer while flushing
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}

		// Publish all events in batch with retry
		for _, ew := range batch {
			var publishErr error

			// Retry loop with exponential backoff
			for attempt := 0; attempt < p.config.MaxRetries; attempt++ {
				if attempt > 0 {
					p.metrics.retriedEvents.Add(1)
					// Exponential backoff (capped at 5x)
					backoffDuration := p.config.RetryInterval * time.Duration(min(attempt, 5))
					select {
					case <-time.After(backoffDuration):
					case <-p.stopCh:
						// Worker is stopping, exit retry loop
						break
					}
				}

				// Track latency from queuing to publishing
				latency := time.Since(ew.queuedAt)
				p.recordLatency(latency)

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				publishErr = p.underlying.Publish(ctx, ew.eventType, ew.eventData)
				cancel()

				if publishErr == nil {
					p.metrics.publishedEvents.Add(1)
					break // Success, exit retry loop
				}

				// Log retry attempt
				if attempt < p.config.MaxRetries-1 {
					// Will retry
				} else {
					// Last attempt failed
					p.metrics.failedEvents.Add(1)
				}
			}

			p.metrics.currentQueueSize.Add(-1)
		}
		batch = batch[:0]
	}

	resetTimer := func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(p.config.BatchTimeout)
	}

	for {
		select {
		case ew := <-p.queue:
			batch = append(batch, ew)
			// Start/reset timer only when batch becomes non-empty
			if len(batch) == 1 {
				// First event in batch - start the timer
				resetTimer()
			}
			if len(batch) >= p.config.BatchSize {
				flushBatch()
			}
		case <-timer.C:
			flushBatch()
		case <-p.stopCh:
			flushBatch()
			return
		}
	}
}

// recordLatency records the publish latency in milliseconds.
func (p *AsyncPublisher) recordLatency(latency time.Duration) {
	ms := latency.Milliseconds()
	p.metrics.totalLatencyMs.Add(ms)
	p.metrics.latencyCount.Add(1)
}

// Close closes the underlying publisher and stops workers.
func (p *AsyncPublisher) Close() error {
	close(p.stopCh)

	// Wait for workers to finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All workers finished, close underlying
		return p.underlying.Close()
	case <-time.After(10 * time.Second):
		// Timeout, but still try to close
		return p.underlying.Close()
	}
}

// GetMetrics returns the current metrics for observability.
func (p *AsyncPublisher) GetMetrics() AsyncMetricsSnapshot {
	return AsyncMetricsSnapshot{
		QueuedEvents:     p.metrics.queuedEvents.Load(),
		PublishedEvents:  p.metrics.publishedEvents.Load(),
		FailedEvents:     p.metrics.failedEvents.Load(),
		DroppedEvents:    p.metrics.droppedEvents.Load(),
		RetriedEvents:    p.metrics.retriedEvents.Load(),
		CurrentQueueSize: p.metrics.currentQueueSize.Load(),
		AvgLatencyMs:     p.getAvgLatencyMs(),
	}
}

// getAvgLatencyMs calculates the average latency in milliseconds.
func (p *AsyncPublisher) getAvgLatencyMs() float64 {
	count := p.metrics.latencyCount.Load()
	if count == 0 {
		return 0
	}
	total := p.metrics.totalLatencyMs.Load()
	return float64(total) / float64(count)
}

// AsyncMetricsSnapshot is a snapshot of async publisher metrics.
type AsyncMetricsSnapshot struct {
	QueuedEvents     int64
	PublishedEvents  int64
	FailedEvents     int64
	DroppedEvents    int64
	RetriedEvents    int64
	CurrentQueueSize int64
	AvgLatencyMs     float64
}

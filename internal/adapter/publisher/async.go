package publisher

import (
	"context"
	"fmt"
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
}

// AsyncPublisherConfig holds configuration for the async publisher.
type AsyncPublisherConfig struct {
	WorkerCount  int
	QueueSize    int
	BatchSize    int
	BatchTimeout time.Duration
}

// asyncMetrics holds atomic metrics.
type asyncMetrics struct {
	queuedEvents     atomic.Int64
	publishedEvents  atomic.Int64
	failedEvents     atomic.Int64
	currentQueueSize atomic.Int64
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
func (p *AsyncPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	select {
	case p.queue <- &eventWrapper{eventType: eventType, eventData: eventData}:
		p.metrics.queuedEvents.Add(1)
		p.metrics.currentQueueSize.Add(1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Queue is full, drop the event
		p.metrics.failedEvents.Add(1)
		return fmt.Errorf("event queue is full")
	}
}

// worker processes events from the queue.
// Uses a timer that only starts when events are queued to avoid idle wake-ups.
func (p *AsyncPublisher) worker(id int) {
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

		// Publish all events in batch
		for _, ew := range batch {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := p.underlying.Publish(ctx, ew.eventType, ew.eventData); err != nil {
				p.metrics.failedEvents.Add(1)
			} else {
				p.metrics.publishedEvents.Add(1)
			}
			cancel()
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

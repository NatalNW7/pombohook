package queue

import (
	"errors"
	"sync"

	"github.com/NatalNW7/pombohook/internal/tunnel"
)

// ErrQueueFull is returned when attempting to enqueue into a full queue.
var ErrQueueFull = errors.New("webhook queue is full")

// WebhookQueue is a thread-safe in-memory FIFO queue for webhook frames.
type WebhookQueue struct {
	mu    sync.Mutex
	items []tunnel.Frame
	max   int
}

// NewWebhookQueue creates a new WebhookQueue with the given max capacity.
func NewWebhookQueue(max int) *WebhookQueue {
	return &WebhookQueue{
		items: make([]tunnel.Frame, 0, max),
		max:   max,
	}
}

// Enqueue adds a frame to the queue. Returns ErrQueueFull if capacity is reached.
func (q *WebhookQueue) Enqueue(frame tunnel.Frame) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) >= q.max {
		return ErrQueueFull
	}

	q.items = append(q.items, frame)
	return nil
}

// DrainAll returns all queued frames in FIFO order and clears the queue.
func (q *WebhookQueue) DrainAll() []tunnel.Frame {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return []tunnel.Frame{}
	}

	drained := q.items
	q.items = make([]tunnel.Frame, 0, q.max)
	return drained
}

// Len returns the current number of items in the queue.
func (q *WebhookQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

package events

import (
	"context"

	"github.com/google/uuid"
)

// Subscriber represents an event subscription
type Subscriber struct {
	ID       string
	DeviceID string
	events   chan Event
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSubscriber creates a new subscriber
func NewSubscriber(ctx context.Context, deviceID string, bufferSize int) *Subscriber {
	subCtx, cancel := context.WithCancel(ctx)

	return &Subscriber{
		ID:       uuid.New().String(),
		DeviceID: deviceID,
		events:   make(chan Event, bufferSize),
		ctx:      subCtx,
		cancel:   cancel,
	}
}

// Events returns the event channel
func (s *Subscriber) Events() <-chan Event {
	return s.events
}

// Close closes the subscriber
func (s *Subscriber) Close() {
	s.cancel()
	close(s.events)
}

// Context returns the subscriber context
func (s *Subscriber) Context() context.Context {
	return s.ctx
}

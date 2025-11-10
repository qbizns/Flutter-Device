package events

import (
	"context"
	"sync"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Event represents a device event
type Event struct {
	DeviceID string
	Type     string
	Data     interface{}
}

// Bus manages event publishing and subscription
type Bus struct {
	subscribers map[string]map[string]*Subscriber // deviceID -> subscriberID -> Subscriber
	mu          sync.RWMutex
	logger      *telemetry.Logger
}

// NewBus creates a new event bus
func NewBus(logger *telemetry.Logger) *Bus {
	return &Bus{
		subscribers: make(map[string]map[string]*Subscriber),
		logger:      logger,
	}
}

// Subscribe creates a new subscription for a device
func (b *Bus) Subscribe(ctx context.Context, deviceID string) *Subscriber {
	sub := NewSubscriber(ctx, deviceID, 100)

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.subscribers[deviceID] == nil {
		b.subscribers[deviceID] = make(map[string]*Subscriber)
	}

	b.subscribers[deviceID][sub.ID] = sub

	b.logger.Info("subscriber added",
		telemetry.String("device_id", deviceID),
		telemetry.String("subscriber_id", sub.ID),
	)

	// Handle unsubscribe on context cancel
	go func() {
		<-sub.ctx.Done()
		b.Unsubscribe(deviceID, sub.ID)
	}()

	return sub
}

// Unsubscribe removes a subscription
func (b *Bus) Unsubscribe(deviceID, subscriberID string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, exists := b.subscribers[deviceID]; exists {
		if sub, exists := subs[subscriberID]; exists {
			sub.Close()
			delete(subs, subscriberID)

			b.logger.Info("subscriber removed",
				telemetry.String("device_id", deviceID),
				telemetry.String("subscriber_id", subscriberID),
			)

			// Clean up empty device map
			if len(subs) == 0 {
				delete(b.subscribers, deviceID)
			}
		}
	}
}

// Publish publishes an event to all subscribers of a device
func (b *Bus) Publish(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	subs, exists := b.subscribers[event.DeviceID]
	if !exists || len(subs) == 0 {
		return
	}

	b.logger.Debug("publishing event",
		telemetry.String("device_id", event.DeviceID),
		telemetry.String("type", event.Type),
		telemetry.Int("subscribers", len(subs)),
	)

	// Publish to all subscribers
	for _, sub := range subs {
		select {
		case sub.events <- event:
			// Event sent
		default:
			// Channel full, drop event
			b.logger.Warn("dropped event, channel full",
				telemetry.String("device_id", event.DeviceID),
				telemetry.String("subscriber_id", sub.ID),
			)
		}
	}
}

// SubscriberCount returns the number of subscribers for a device
func (b *Bus) SubscriberCount(deviceID string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, exists := b.subscribers[deviceID]; exists {
		return len(subs)
	}
	return 0
}

// TotalSubscribers returns the total number of subscribers
func (b *Bus) TotalSubscribers() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	total := 0
	for _, subs := range b.subscribers {
		total += len(subs)
	}
	return total
}

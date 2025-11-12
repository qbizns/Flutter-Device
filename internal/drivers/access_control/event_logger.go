package access_control

import (
	"context"
	"sync"
	"time"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Memory-based event logger implementation (for demonstration)
// In production, this would use a persistent storage backend
type memoryEventLogger struct {
	events []*pb.AccessEvent
	mu     sync.RWMutex
	maxEvents int
}

// NewMemoryEventLogger creates a new memory-based event logger
func NewMemoryEventLogger() EventLogger {
	return &memoryEventLogger{
		events:    make([]*pb.AccessEvent, 0, 1000),
		maxEvents: 10000, // Keep last 10,000 events
	}
}

// LogEvent logs an access event
func (l *memoryEventLogger) LogEvent(ctx context.Context, event *pb.AccessEvent) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Add event
	l.events = append(l.events, event)

	// Trim if exceeded max
	if len(l.events) > l.maxEvents {
		// Keep most recent events
		l.events = l.events[len(l.events)-l.maxEvents:]
	}

	return nil
}

// GetEvents retrieves events from log
func (l *memoryEventLogger) GetEvents(ctx context.Context, filter *EventFilter) ([]*pb.AccessEvent, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]*pb.AccessEvent, 0)

	for _, event := range l.events {
		if l.matchesFilter(event, filter) {
			filtered = append(filtered, event)
		}
	}

	// Apply pagination
	pageSize := int(filter.PageSize)
	if pageSize == 0 {
		pageSize = 100 // Default page size
	}

	start := 0
	if filter.PageToken != "" {
		// In a real implementation, decode page token to get offset
		// For now, just return from beginning
	}

	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	if start >= len(filtered) {
		return []*pb.AccessEvent{}, nil
	}

	return filtered[start:end], nil
}

// matchesFilter checks if an event matches the filter
func (l *memoryEventLogger) matchesFilter(event *pb.AccessEvent, filter *EventFilter) bool {
	// Filter by device ID
	if filter.DeviceID != "" && event.DeviceId != filter.DeviceID {
		return false
	}

	// Filter by door IDs
	if len(filter.DoorIDs) > 0 {
		match := false
		for _, doorID := range filter.DoorIDs {
			if event.DoorId == doorID {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Filter by time range
	eventTime := event.Timestamp.AsTime()
	if !filter.StartTime.IsZero() && eventTime.Before(filter.StartTime) {
		return false
	}
	if !filter.EndTime.IsZero() && eventTime.After(filter.EndTime) {
		return false
	}

	// Filter by event types
	if len(filter.EventTypes) > 0 {
		match := false
		for _, eventType := range filter.EventTypes {
			if event.Type == eventType {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	// Filter by user IDs
	if len(filter.UserIDs) > 0 {
		if event.UserInfo == nil {
			return false
		}
		match := false
		for _, userID := range filter.UserIDs {
			if event.UserInfo.UserId == userID {
				match = true
				break
			}
		}
		if !match {
			return false
		}
	}

	return true
}

// GetEventCount returns the total number of events
func (l *memoryEventLogger) GetEventCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.events)
}

// Clear clears all events
func (l *memoryEventLogger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.events = make([]*pb.AccessEvent, 0, 1000)
}

// GetRecentEvents returns the N most recent events
func (l *memoryEventLogger) GetRecentEvents(n int) []*pb.AccessEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if n > len(l.events) {
		n = len(l.events)
	}

	if n == 0 {
		return []*pb.AccessEvent{}
	}

	// Return last N events
	start := len(l.events) - n
	return l.events[start:]
}

// GetEventsInTimeRange returns events within a time range
func (l *memoryEventLogger) GetEventsInTimeRange(start, end time.Time) []*pb.AccessEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]*pb.AccessEvent, 0)

	for _, event := range l.events {
		eventTime := event.Timestamp.AsTime()
		if eventTime.After(start) && eventTime.Before(end) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventsByType returns events of a specific type
func (l *memoryEventLogger) GetEventsByType(eventType pb.AccessEventType) []*pb.AccessEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]*pb.AccessEvent, 0)

	for _, event := range l.events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventsByDoor returns events for a specific door
func (l *memoryEventLogger) GetEventsByDoor(doorID string) []*pb.AccessEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]*pb.AccessEvent, 0)

	for _, event := range l.events {
		if event.DoorId == doorID {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// GetEventsByUser returns events for a specific user
func (l *memoryEventLogger) GetEventsByUser(userID string) []*pb.AccessEvent {
	l.mu.RLock()
	defer l.mu.RUnlock()

	filtered := make([]*pb.AccessEvent, 0)

	for _, event := range l.events {
		if event.UserInfo != nil && event.UserInfo.UserId == userID {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

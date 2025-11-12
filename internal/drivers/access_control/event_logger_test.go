package access_control

import (
	"context"
	"fmt"
	"testing"
	"time"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewMemoryEventLogger(t *testing.T) {
	logger := NewMemoryEventLogger()

	assert.NotNil(t, logger)

	memLogger := logger.(*memoryEventLogger)
	assert.Empty(t, memLogger.events)
	assert.Equal(t, 10000, memLogger.maxEvents)
}

func TestEventLogger_LogEvent(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	event := &pb.AccessEvent{
		EventId:   "evt-1",
		DeviceId:  "controller-1",
		DoorId:    "door-1",
		Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
		Timestamp: timestamppb.Now(),
		Severity:  pb.EventSeverity_EVENT_SEVERITY_INFO,
	}

	err := logger.LogEvent(ctx, event)
	require.NoError(t, err)

	memLogger := logger.(*memoryEventLogger)
	assert.Len(t, memLogger.events, 1)
	assert.Equal(t, "evt-1", memLogger.events[0].EventId)
}

func TestEventLogger_GetEvents_EmptyFilter(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	// Add some events
	for i := 1; i <= 5; i++ {
		event := &pb.AccessEvent{
			EventId:   fmt.Sprintf("evt-%d", i),
			DeviceId:  "controller-1",
			DoorId:    "door-1",
			Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
			Timestamp: timestamppb.Now(),
		}
		logger.LogEvent(ctx, event)
	}

	filter := &EventFilter{}
	events, err := logger.GetEvents(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, events, 5)
}

func TestEventLogger_GetEvents_FilterByDevice(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	// Add events for different devices
	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:  "evt-1",
		DeviceId: "controller-1",
		DoorId:   "door-1",
		Type:     pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:  "evt-2",
		DeviceId: "controller-2",
		DoorId:   "door-2",
		Type:     pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	filter := &EventFilter{
		DeviceID: "controller-1",
	}

	events, err := logger.GetEvents(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "controller-1", events[0].DeviceId)
}

func TestEventLogger_GetEvents_FilterByDoors(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	// Add events for different doors
	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-1",
		DoorId:  "door-1",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-2",
		DoorId:  "door-2",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-3",
		DoorId:  "door-3",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	filter := &EventFilter{
		DoorIDs: []string{"door-1", "door-2"},
	}

	events, err := logger.GetEvents(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestEventLogger_GetEvents_FilterByEventType(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	// Add events of different types
	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-1",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-2",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_DENIED,
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-3",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_DOOR_FORCED,
	})

	filter := &EventFilter{
		EventTypes: []pb.AccessEventType{
			pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_DENIED,
			pb.AccessEventType_ACCESS_EVENT_TYPE_DOOR_FORCED,
		},
	}

	events, err := logger.GetEvents(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestEventLogger_GetEvents_FilterByTimeRange(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	now := time.Now()

	// Add events at different times
	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:   "evt-1",
		Timestamp: timestamppb.New(now.Add(-2 * time.Hour)),
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:   "evt-2",
		Timestamp: timestamppb.New(now.Add(-1 * time.Hour)),
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:   "evt-3",
		Timestamp: timestamppb.New(now),
	})

	filter := &EventFilter{
		StartTime: now.Add(-90 * time.Minute),
		EndTime:   now.Add(10 * time.Minute),
	}

	events, err := logger.GetEvents(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestEventLogger_GetEvents_FilterByUser(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	// Add events for different users
	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:  "evt-1",
		UserInfo: &pb.CredentialInfo{UserId: "user1"},
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:  "evt-2",
		UserInfo: &pb.CredentialInfo{UserId: "user2"},
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-3",
		// No user info
	})

	filter := &EventFilter{
		UserIDs: []string{"user1"},
	}

	events, err := logger.GetEvents(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "user1", events[0].UserInfo.UserId)
}

func TestEventLogger_GetEvents_Pagination(t *testing.T) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	// Add 10 events
	for i := 1; i <= 10; i++ {
		logger.LogEvent(ctx, &pb.AccessEvent{
			EventId: fmt.Sprintf("evt-%d", i),
			Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
		})
	}

	// Request first page (3 events)
	filter := &EventFilter{
		PageSize: 3,
	}

	events, err := logger.GetEvents(ctx, filter)

	require.NoError(t, err)
	assert.Len(t, events, 3)
}

func TestEventLogger_MaxEventsLimit(t *testing.T) {
	logger := NewMemoryEventLogger().(*memoryEventLogger)
	logger.maxEvents = 5 // Set lower limit for testing
	ctx := context.Background()

	// Add 10 events (exceeding limit)
	for i := 1; i <= 10; i++ {
		logger.LogEvent(ctx, &pb.AccessEvent{
			EventId: fmt.Sprintf("evt-%d", i),
		})
	}

	// Should only keep last 5 events
	assert.Len(t, logger.events, 5)
	assert.Equal(t, "evt-6", logger.events[0].EventId)
	assert.Equal(t, "evt-10", logger.events[4].EventId)
}

func TestEventLogger_GetRecentEvents(t *testing.T) {
	logger := NewMemoryEventLogger().(*memoryEventLogger)
	ctx := context.Background()

	// Add 5 events
	for i := 1; i <= 5; i++ {
		logger.LogEvent(ctx, &pb.AccessEvent{
			EventId: fmt.Sprintf("evt-%d", i),
		})
	}

	// Get last 3 events
	recent := logger.GetRecentEvents(3)

	assert.Len(t, recent, 3)
	assert.Equal(t, "evt-3", recent[0].EventId)
	assert.Equal(t, "evt-5", recent[2].EventId)
}

func TestEventLogger_GetEventsByType(t *testing.T) {
	logger := NewMemoryEventLogger().(*memoryEventLogger)
	ctx := context.Background()

	// Add mixed event types
	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-1",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-2",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_DENIED,
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-3",
		Type:    pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
	})

	events := logger.GetEventsByType(pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED)

	assert.Len(t, events, 2)
	assert.Equal(t, "evt-1", events[0].EventId)
	assert.Equal(t, "evt-3", events[1].EventId)
}

func TestEventLogger_GetEventsByDoor(t *testing.T) {
	logger := NewMemoryEventLogger().(*memoryEventLogger)
	ctx := context.Background()

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-1",
		DoorId:  "door-1",
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-2",
		DoorId:  "door-2",
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId: "evt-3",
		DoorId:  "door-1",
	})

	events := logger.GetEventsByDoor("door-1")

	assert.Len(t, events, 2)
	assert.Equal(t, "evt-1", events[0].EventId)
	assert.Equal(t, "evt-3", events[1].EventId)
}

func TestEventLogger_GetEventsByUser(t *testing.T) {
	logger := NewMemoryEventLogger().(*memoryEventLogger)
	ctx := context.Background()

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:  "evt-1",
		UserInfo: &pb.CredentialInfo{UserId: "user1"},
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:  "evt-2",
		UserInfo: &pb.CredentialInfo{UserId: "user2"},
	})

	logger.LogEvent(ctx, &pb.AccessEvent{
		EventId:  "evt-3",
		UserInfo: &pb.CredentialInfo{UserId: "user1"},
	})

	events := logger.GetEventsByUser("user1")

	assert.Len(t, events, 2)
	assert.Equal(t, "evt-1", events[0].EventId)
	assert.Equal(t, "evt-3", events[1].EventId)
}

func TestEventLogger_Clear(t *testing.T) {
	logger := NewMemoryEventLogger().(*memoryEventLogger)
	ctx := context.Background()

	// Add events
	for i := 1; i <= 5; i++ {
		logger.LogEvent(ctx, &pb.AccessEvent{
			EventId: fmt.Sprintf("evt-%d", i),
		})
	}

	assert.Len(t, logger.events, 5)

	// Clear
	logger.Clear()

	assert.Empty(t, logger.events)
}

func TestEventLogger_GetEventCount(t *testing.T) {
	logger := NewMemoryEventLogger().(*memoryEventLogger)
	ctx := context.Background()

	assert.Equal(t, 0, logger.GetEventCount())

	// Add 3 events
	for i := 1; i <= 3; i++ {
		logger.LogEvent(ctx, &pb.AccessEvent{
			EventId: fmt.Sprintf("evt-%d", i),
		})
	}

	assert.Equal(t, 3, logger.GetEventCount())
}

func BenchmarkEventLogger_LogEvent(b *testing.B) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	event := &pb.AccessEvent{
		EventId:   "evt-bench",
		DeviceId:  "controller-1",
		DoorId:    "door-1",
		Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
		Timestamp: timestamppb.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = logger.LogEvent(ctx, event)
	}
}

func BenchmarkEventLogger_GetEvents(b *testing.B) {
	logger := NewMemoryEventLogger()
	ctx := context.Background()

	// Pre-populate with 1000 events
	for i := 1; i <= 1000; i++ {
		logger.LogEvent(ctx, &pb.AccessEvent{
			EventId:  fmt.Sprintf("evt-%d", i),
			DeviceId: "controller-1",
			DoorId:   fmt.Sprintf("door-%d", i%10),
			Type:     pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED,
		})
	}

	filter := &EventFilter{
		DoorIDs:  []string{"door-5"},
		PageSize: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = logger.GetEvents(ctx, filter)
	}
}

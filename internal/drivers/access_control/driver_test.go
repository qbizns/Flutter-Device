package access_control

import (
	"context"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDriver(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller-1"
	config.Name = "Test Access Controller"
	config.Doors = []DoorConfig{
		{
			ID:       "door-1",
			Name:     "Main Door",
			LockType: LockTypeElectricStrike,
			Mode:     pb.DoorMode_DOOR_MODE_NORMAL,
		},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)

	require.NoError(t, err)
	assert.NotNil(t, driver)
	assert.Equal(t, "test-controller-1", driver.id)
	assert.Equal(t, "Test Access Controller", driver.name)
	assert.Len(t, driver.doorControllers, 1)
	assert.NotNil(t, driver.decisionEngine)
	assert.NotNil(t, driver.eventLogger)
}

func TestDriver_GetDoorStatus(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Test Door"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	req := &pb.GetDoorStatusRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
	}

	status, err := driver.GetDoorStatus(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, "door-1", status.DoorId)
	assert.Equal(t, "Test Door", status.DoorName)
	assert.Equal(t, pb.LockStatus_LOCK_STATUS_LOCKED, status.LockStatus)
	assert.Equal(t, pb.DoorPosition_DOOR_POSITION_CLOSED, status.DoorPosition)
	assert.Equal(t, pb.DoorMode_DOOR_MODE_NORMAL, status.Mode)
}

func TestDriver_UnlockDoor(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Test Door"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	// Start driver
	err = driver.Start()
	require.NoError(t, err)
	defer driver.Stop()

	ctx := context.Background()
	req := &pb.UnlockDoorRequest{
		DeviceId:        config.ID,
		DoorId:          "door-1",
		DurationSeconds: 5,
		Reason:          "Test unlock",
	}

	resp, err := driver.UnlockDoor(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, "door-1", resp.DoorId)
	assert.Equal(t, int32(5), resp.DurationSeconds)

	// Check door status is unlocked
	statusReq := &pb.GetDoorStatusRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
	}
	status, err := driver.GetDoorStatus(ctx, statusReq)
	require.NoError(t, err)
	assert.Equal(t, pb.LockStatus_LOCK_STATUS_UNLOCKED, status.LockStatus)
}

func TestDriver_LockDoor(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Test Door"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	err = driver.Start()
	require.NoError(t, err)
	defer driver.Stop()

	ctx := context.Background()

	// Unlock first
	unlockReq := &pb.UnlockDoorRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
		DurationSeconds: -1, // Indefinite
	}
	_, err = driver.UnlockDoor(ctx, unlockReq)
	require.NoError(t, err)

	// Now lock
	lockReq := &pb.LockDoorRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
		Reason:   "Test lock",
	}

	resp, err := driver.LockDoor(ctx, lockReq)
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// Check door status is locked
	statusReq := &pb.GetDoorStatusRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
	}
	status, err := driver.GetDoorStatus(ctx, statusReq)
	require.NoError(t, err)
	assert.Equal(t, pb.LockStatus_LOCK_STATUS_LOCKED, status.LockStatus)
}

func TestDriver_CheckAccess(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Test Door"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	req := &pb.CheckAccessRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
		Credential: &pb.Credential{
			Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
			Data: &pb.Credential_CardUid{
				CardUid: []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
	}

	resp, err := driver.CheckAccess(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_GRANTED, resp.Decision)
	assert.NotNil(t, resp.CredentialInfo)
}

func TestDriver_GrantAccess(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Test Door"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	err = driver.Start()
	require.NoError(t, err)
	defer driver.Stop()

	ctx := context.Background()
	req := &pb.GrantAccessRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
		Credential: &pb.Credential{
			Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
			Data: &pb.Credential_CardUid{
				CardUid: []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
	}

	resp, err := driver.GrantAccess(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Granted)
	assert.True(t, resp.DoorUnlocked)
	assert.NotEmpty(t, resp.EventId)
	assert.NotNil(t, resp.UserInfo)
}

func TestDriver_SetDoorMode(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Test Door"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	err = driver.Start()
	require.NoError(t, err)
	defer driver.Stop()

	ctx := context.Background()
	req := &pb.SetDoorModeRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
		Mode:     pb.DoorMode_DOOR_MODE_UNLOCKED,
	}

	resp, err := driver.SetDoorMode(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, pb.DoorMode_DOOR_MODE_UNLOCKED, resp.CurrentMode)

	// Check door status
	statusReq := &pb.GetDoorStatusRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
	}
	status, err := driver.GetDoorStatus(ctx, statusReq)
	require.NoError(t, err)
	assert.Equal(t, pb.DoorMode_DOOR_MODE_UNLOCKED, status.Mode)
	assert.Equal(t, pb.LockStatus_LOCK_STATUS_UNLOCKED, status.LockStatus)
}

func TestDriver_Lockdown(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Door 1"},
		{ID: "door-2", Name: "Door 2"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	err = driver.Start()
	require.NoError(t, err)
	defer driver.Stop()

	ctx := context.Background()

	// Activate lockdown
	activateReq := &pb.ActivateLockdownRequest{
		DeviceId:    config.ID,
		Level:       pb.LockdownLevel_LOCKDOWN_LEVEL_FULL,
		Reason:      "Test emergency",
		InitiatedBy: "test_user",
	}

	activateResp, err := driver.ActivateLockdown(ctx, activateReq)
	require.NoError(t, err)
	assert.True(t, activateResp.Success)
	assert.Len(t, activateResp.DoorsLocked, 2)

	// Check lockdown status
	assert.True(t, driver.lockdownState.IsActive())
	assert.Equal(t, pb.LockdownLevel_LOCKDOWN_LEVEL_FULL, driver.lockdownState.GetLevel())

	// Try to grant access during lockdown
	accessReq := &pb.CheckAccessRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
		Credential: &pb.Credential{
			Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
			Data: &pb.Credential_CardUid{
				CardUid: []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
	}

	accessResp, err := driver.CheckAccess(ctx, accessReq)
	require.NoError(t, err)
	assert.Equal(t, pb.AccessDecision_ACCESS_DECISION_DENIED, accessResp.Decision)
	assert.Contains(t, accessResp.Reason, "Lockdown")

	// Deactivate lockdown
	deactivateReq := &pb.DeactivateLockdownRequest{
		DeviceId:      config.ID,
		Reason:        "Emergency resolved",
		DeactivatedBy: "test_user",
	}

	deactivateResp, err := driver.DeactivateLockdown(ctx, deactivateReq)
	require.NoError(t, err)
	assert.True(t, deactivateResp.Success)
	assert.Len(t, deactivateResp.DoorsRestored, 2)

	// Check lockdown is deactivated
	assert.False(t, driver.lockdownState.IsActive())
}

func TestDriver_GetControllerStatus(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-controller"
	config.Name = "Test Controller"
	config.Model = "AC-1000"
	config.Manufacturer = "Generic"
	config.Doors = []DoorConfig{
		{ID: "door-1", Name: "Door 1"},
	}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	req := &pb.GetControllerStatusRequest{
		DeviceId: config.ID,
	}

	status, err := driver.GetControllerStatus(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "test-controller", status.DeviceId)
	assert.Equal(t, "Test Controller", status.Name)
	assert.Equal(t, "AC-1000", status.Model)
	assert.Equal(t, "Generic", status.Manufacturer)
	assert.Len(t, status.Doors, 1)
	assert.NotNil(t, status.Statistics)
	assert.False(t, status.LockdownActive)
}

func TestDoorState(t *testing.T) {
	state := &DoorState{
		DoorID:   "door-1",
		DoorName: "Test Door",
	}

	// Test lock status
	state.SetLockStatus(pb.LockStatus_LOCK_STATUS_UNLOCKED)
	assert.Equal(t, pb.LockStatus_LOCK_STATUS_UNLOCKED, state.LockStatus)
	assert.False(t, state.UnlockedSince.IsZero())

	// Test door position
	state.SetDoorPosition(pb.DoorPosition_DOOR_POSITION_OPEN)
	assert.Equal(t, pb.DoorPosition_DOOR_POSITION_OPEN, state.DoorPosition)

	// Test mode
	state.SetMode(pb.DoorMode_DOOR_MODE_LOCKED)
	assert.Equal(t, pb.DoorMode_DOOR_MODE_LOCKED, state.Mode)

	// Test alarm status
	state.SetAlarmStatus(pb.AlarmStatus_ALARM_STATUS_DOOR_FORCED)
	assert.Equal(t, pb.AlarmStatus_ALARM_STATUS_DOOR_FORCED, state.AlarmStatus)

	// Test GetStatus
	status := state.GetStatus()
	assert.Equal(t, "door-1", status.DoorId)
	assert.Equal(t, pb.LockStatus_LOCK_STATUS_UNLOCKED, status.LockStatus)
}

func TestLockdownState(t *testing.T) {
	state := &LockdownState{}

	// Initially not active
	assert.False(t, state.IsActive())

	// Activate lockdown
	state.Activate(
		pb.LockdownLevel_LOCKDOWN_LEVEL_FULL,
		"Test emergency",
		"admin",
		[]string{"door-1", "door-2"},
	)

	assert.True(t, state.IsActive())
	assert.Equal(t, pb.LockdownLevel_LOCKDOWN_LEVEL_FULL, state.GetLevel())
	assert.Equal(t, "Test emergency", state.Reason)
	assert.Equal(t, "admin", state.InitiatedBy)
	assert.Len(t, state.AffectedDoors, 2)

	// Deactivate lockdown
	state.Deactivate()
	assert.False(t, state.IsActive())
}

func TestControllerStatistics(t *testing.T) {
	stats := &ControllerStatistics{
		OnlineSince: time.Now(),
	}

	// Test access attempts
	stats.IncrementAccessAttempts(true)
	assert.Equal(t, int64(1), stats.TotalAccessAttempts)
	assert.Equal(t, int64(1), stats.TotalGrants)
	assert.Equal(t, int64(0), stats.TotalDenials)

	stats.IncrementAccessAttempts(false)
	assert.Equal(t, int64(2), stats.TotalAccessAttempts)
	assert.Equal(t, int64(1), stats.TotalGrants)
	assert.Equal(t, int64(1), stats.TotalDenials)

	// Test alarms
	stats.IncrementAlarms()
	assert.Equal(t, int64(1), stats.TotalAlarms)

	// Test ToProto
	proto := stats.ToProto()
	assert.Equal(t, int64(2), proto.TotalAccessAttempts)
	assert.Equal(t, int64(1), proto.TotalGrants)
}

func TestDoorStatisticsTracker(t *testing.T) {
	tracker := &DoorStatisticsTracker{
		StartedAt: time.Now(),
	}

	// Test grants
	tracker.IncrementGrants()
	assert.Equal(t, int64(1), tracker.TotalGrants)
	assert.False(t, tracker.LastGrant.IsZero())

	// Test denials
	tracker.IncrementDenials()
	assert.Equal(t, int64(1), tracker.TotalDenials)
	assert.False(t, tracker.LastDenial.IsZero())

	// Test alarms
	tracker.IncrementAlarms()
	assert.Equal(t, int64(1), tracker.TotalAlarms)
	assert.False(t, tracker.LastAlarm.IsZero())

	// Test unlocks
	tracker.IncrementUnlocks()
	assert.Equal(t, int64(1), tracker.TotalUnlocks)

	// Test ToProto
	proto := tracker.ToProto()
	assert.Equal(t, int64(1), proto.TotalGrants)
	assert.Equal(t, int64(1), proto.TotalDenials)
}

func TestAntiPassbackTracker(t *testing.T) {
	tracker := NewAntiPassbackTracker()

	// Test initial state
	_, exists := tracker.GetLocation("user1")
	assert.False(t, exists)

	// Test first entry (should allow)
	allowed := tracker.CheckTransition("user1", "zone_a", "zone_b")
	assert.True(t, allowed)

	// Update location
	tracker.UpdateLocation("user1", "zone_b")
	location, exists := tracker.GetLocation("user1")
	assert.True(t, exists)
	assert.Equal(t, "zone_b", location)

	// Test valid transition
	allowed = tracker.CheckTransition("user1", "zone_b", "zone_c")
	assert.True(t, allowed)

	// Test invalid transition (anti-passback violation)
	allowed = tracker.CheckTransition("user1", "zone_a", "zone_c")
	assert.False(t, allowed)
}

func TestWiegandParsing(t *testing.T) {
	// Test 26-bit Wiegand
	data26 := uint32(0x12345678)
	wiegand26 := ParseWiegand26(data26)
	assert.Equal(t, 26, wiegand26.Bits)
	assert.NotZero(t, wiegand26.FacilityCode)
	assert.NotZero(t, wiegand26.CardNumber)

	// Test 37-bit Wiegand
	data37 := uint64(0x123456789A)
	wiegand37 := ParseWiegand37(data37)
	assert.Equal(t, 37, wiegand37.Bits)
	assert.NotZero(t, wiegand37.FacilityCode)
	assert.NotZero(t, wiegand37.CardNumber)
}

func BenchmarkDriverGrantAccess(b *testing.B) {
	config := DefaultConfig()
	config.ID = "bench-controller"
	config.Doors = []DoorConfig{{ID: "door-1", Name: "Test Door"}}

	logger := telemetry.NewLogger("test")
	driver, _ := NewDriver(config.ID, config.Name, config, logger)
	driver.Start()
	defer driver.Stop()

	ctx := context.Background()
	req := &pb.GrantAccessRequest{
		DeviceId: config.ID,
		DoorId:   "door-1",
		Credential: &pb.Credential{
			Type: pb.CredentialType_CREDENTIAL_TYPE_CARD,
			Data: &pb.Credential_CardUid{
				CardUid: []byte{0x01, 0x02, 0x03, 0x04},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = driver.GrantAccess(ctx, req)
	}
}

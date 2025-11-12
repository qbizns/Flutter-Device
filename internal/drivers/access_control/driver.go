package access_control

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Driver implements access control functionality
type Driver struct {
	id     string
	name   string
	config Config
	logger *telemetry.Logger

	// Connection
	conn   Connection
	connMu sync.RWMutex

	// Door controllers
	doorControllers map[string]*DoorState
	doorStats       map[string]*DoorStatisticsTracker
	doorMu          sync.RWMutex

	// Access decision engine
	decisionEngine AccessDecisionEngine

	// Event logger
	eventLogger EventLogger

	// Current state
	status   devices.DeviceStatus
	statusMu sync.RWMutex

	// Lockdown state
	lockdownState *LockdownState

	// Statistics
	stats *ControllerStatistics

	// Anti-passback tracking
	antiPassback *AntiPassbackTracker

	// Event channels
	accessEventChan chan *pb.AccessEvent

	// Request queues
	accessQueue chan *AccessRequest
	unlockQueue chan *UnlockRequest

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDriver creates a new access control driver
func NewDriver(id, name string, config Config, logger *telemetry.Logger) (*Driver, error) {
	if logger == nil {
		logger = telemetry.NewLogger("access_control")
	}

	ctx, cancel := context.WithCancel(context.Background())

	driver := &Driver{
		id:              id,
		name:            name,
		config:          config,
		logger:          logger,
		status:          devices.StatusDisconnected,
		doorControllers: make(map[string]*DoorState),
		doorStats:       make(map[string]*DoorStatisticsTracker),
		lockdownState:   &LockdownState{},
		stats: &ControllerStatistics{
			OnlineSince: time.Now(),
		},
		antiPassback:    NewAntiPassbackTracker(),
		accessEventChan: make(chan *pb.AccessEvent, 100),
		accessQueue:     make(chan *AccessRequest, 10),
		unlockQueue:     make(chan *UnlockRequest, 10),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize doors
	for _, doorConfig := range config.Doors {
		driver.doorControllers[doorConfig.ID] = &DoorState{
			DoorID:       doorConfig.ID,
			DoorName:     doorConfig.Name,
			LockStatus:   pb.LockStatus_LOCK_STATUS_LOCKED,
			DoorPosition: pb.DoorPosition_DOOR_POSITION_CLOSED,
			Mode:         doorConfig.Mode,
			LockedSince:  time.Now(),
		}

		driver.doorStats[doorConfig.ID] = &DoorStatisticsTracker{
			StartedAt: time.Now(),
		}
	}

	// Initialize decision engine
	driver.decisionEngine = NewSimpleDecisionEngine(logger)

	// Initialize event logger
	driver.eventLogger = NewMemoryEventLogger()

	return driver, nil
}

// Start starts the access control driver
func (d *Driver) Start() error {
	d.statusMu.Lock()
	if d.status != devices.StatusDisconnected {
		d.statusMu.Unlock()
		return fmt.Errorf("driver already started")
	}
	d.statusMu.Unlock()

	d.logger.Info("starting access control driver",
		telemetry.String("id", d.id),
		telemetry.String("transport", d.config.Transport),
		telemetry.Int("doors", len(d.config.Doors)),
	)

	// Establish connection (if network-based)
	if d.config.Transport == "tcp" || d.config.Transport == "serial" || d.config.Transport == "rs485" {
		if err := d.connect(); err != nil {
			d.setStatus(devices.StatusError)
			return fmt.Errorf("failed to connect: %w", err)
		}
	}

	d.setStatus(devices.StatusReady)

	// Start worker goroutines
	d.wg.Add(3)
	go d.doorMonitorLoop()
	go d.accessRequestWorker()
	go d.unlockRequestWorker()

	d.logger.Info("access control driver started successfully")

	return nil
}

// Stop stops the access control driver
func (d *Driver) Stop() error {
	d.logger.Info("stopping access control driver")

	// Cancel context
	d.cancel()

	// Lock all doors before shutdown (if not in evacuation mode)
	if !d.lockdownState.IsActive() || d.lockdownState.GetLevel() != pb.LockdownLevel_LOCKDOWN_LEVEL_EVACUATION {
		for doorID := range d.doorControllers {
			d.lockDoorImmediate(doorID)
		}
	}

	// Close connection
	d.connMu.Lock()
	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}
	d.connMu.Unlock()

	// Wait for workers
	d.wg.Wait()

	d.setStatus(devices.StatusDisconnected)

	d.logger.Info("access control driver stopped")

	return nil
}

// Status returns the current device status
func (d *Driver) Status() devices.DeviceStatus {
	d.statusMu.RLock()
	defer d.statusMu.RUnlock()
	return d.status
}

// setStatus sets the device status
func (d *Driver) setStatus(status devices.DeviceStatus) {
	d.statusMu.Lock()
	defer d.statusMu.Unlock()
	d.status = status
}

// connect establishes connection to the controller
func (d *Driver) connect() error {
	var conn Connection
	var err error

	switch d.config.Transport {
	case "tcp":
		conn, err = connectTCP(d.config.Address, d.config.Port, d.logger)
	case "serial", "rs485":
		conn, err = connectSerial(d.config.SerialDev, d.config.BaudRate, d.logger)
	case "wiegand":
		// Wiegand uses GPIO, no network connection needed
		return nil
	default:
		return fmt.Errorf("unsupported transport: %s", d.config.Transport)
	}

	if err != nil {
		return err
	}

	d.connMu.Lock()
	d.conn = conn
	d.connMu.Unlock()

	return nil
}

// ==================================================
// RPC Method Implementations
// ==================================================

// UnlockDoor unlocks a door for specified duration
func (d *Driver) UnlockDoor(ctx context.Context, req *pb.UnlockDoorRequest) (*pb.UnlockDoorResponse, error) {
	d.logger.Debug("unlock door request",
		telemetry.String("device_id", req.DeviceId),
		telemetry.String("door_id", req.DoorId),
		telemetry.Int("duration", int(req.DurationSeconds)),
	)

	doorID := req.DoorId
	if doorID == "" {
		// Use first door if not specified
		if len(d.config.Doors) > 0 {
			doorID = d.config.Doors[0].ID
		} else {
			return nil, fmt.Errorf("no doors configured")
		}
	}

	duration := time.Duration(req.DurationSeconds) * time.Second
	if req.DurationSeconds == 0 {
		duration = time.Duration(d.config.DefaultUnlockDuration) * time.Second
	} else if req.DurationSeconds == -1 {
		duration = 0 // Indefinite
	}

	unlockReq := &UnlockRequest{
		DoorID:      doorID,
		Duration:    duration,
		Reason:      req.Reason,
		TriggeredBy: req.TriggeredBy,
		ResultChan:  make(chan *UnlockResult, 1),
		Timestamp:   time.Now(),
	}

	// Send to unlock queue
	select {
	case d.unlockQueue <- unlockReq:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Wait for result
	select {
	case result := <-unlockReq.ResultChan:
		if result.Error != nil {
			return nil, result.Error
		}

		return &pb.UnlockDoorResponse{
			Success:         result.Success,
			DoorId:          doorID,
			DurationSeconds: int32(result.Duration.Seconds()),
			Timestamp:       timestamppb.New(result.Timestamp),
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// LockDoor locks a door immediately
func (d *Driver) LockDoor(ctx context.Context, req *pb.LockDoorRequest) (*pb.LockDoorResponse, error) {
	d.logger.Debug("lock door request",
		telemetry.String("device_id", req.DeviceId),
		telemetry.String("door_id", req.DoorId),
	)

	doorID := req.DoorId
	if doorID == "" && len(d.config.Doors) > 0 {
		doorID = d.config.Doors[0].ID
	}

	if err := d.lockDoorImmediate(doorID); err != nil {
		return nil, err
	}

	// Emit event
	d.emitAccessEvent(&pb.AccessEvent{
		EventId:   generateEventID(),
		DeviceId:  d.id,
		DoorId:    doorID,
		Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_DOOR_LOCKED,
		Timestamp: timestamppb.Now(),
		Severity:  pb.EventSeverity_EVENT_SEVERITY_INFO,
	})

	return &pb.LockDoorResponse{
		Success:   true,
		DoorId:    doorID,
		Timestamp: timestamppb.Now(),
	}, nil
}

// GetDoorStatus returns current door status
func (d *Driver) GetDoorStatus(ctx context.Context, req *pb.GetDoorStatusRequest) (*pb.DoorStatus, error) {
	doorID := req.DoorId
	if doorID == "" && len(d.config.Doors) > 0 {
		doorID = d.config.Doors[0].ID
	}

	d.doorMu.RLock()
	doorState, exists := d.doorControllers[doorID]
	doorStats, statsExists := d.doorStats[doorID]
	d.doorMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("door not found: %s", doorID)
	}

	status := doorState.GetStatus()
	status.DeviceId = d.id

	if statsExists {
		status.Statistics = doorStats.ToProto()
	}

	return status, nil
}

// CheckAccess checks if access should be granted
func (d *Driver) CheckAccess(ctx context.Context, req *pb.CheckAccessRequest) (*pb.CheckAccessResponse, error) {
	doorID := req.DoorId
	if doorID == "" && len(d.config.Doors) > 0 {
		doorID = d.config.Doors[0].ID
	}

	// Check if lockdown is active
	if d.lockdownState.IsActive() {
		level := d.lockdownState.GetLevel()
		if level == pb.LockdownLevel_LOCKDOWN_LEVEL_FULL {
			return &pb.CheckAccessResponse{
				Decision:  pb.AccessDecision_ACCESS_DECISION_DENIED,
				Reason:    "Lockdown active",
				Timestamp: timestamppb.Now(),
			}, nil
		}
	}

	// Use decision engine
	result, err := d.decisionEngine.CheckAccess(req.Credential, doorID, req.Context)
	if err != nil {
		return nil, err
	}

	return &pb.CheckAccessResponse{
		Decision:       result.Decision,
		Reason:         result.Reason,
		CredentialInfo: result.CredentialInfo,
		RuleId:         result.RuleID,
		Timestamp:      timestamppb.New(result.Timestamp),
	}, nil
}

// GrantAccess grants access and unlocks door
func (d *Driver) GrantAccess(ctx context.Context, req *pb.GrantAccessRequest) (*pb.GrantAccessResponse, error) {
	d.logger.Debug("grant access request",
		telemetry.String("device_id", req.DeviceId),
		telemetry.String("door_id", req.DoorId),
	)

	doorID := req.DoorId
	if doorID == "" && len(d.config.Doors) > 0 {
		doorID = d.config.Doors[0].ID
	}

	accessReq := &AccessRequest{
		ID:         generateRequestID(),
		DoorID:     doorID,
		Credential: req.Credential,
		Context:    req.Context,
		ResultChan: make(chan *AccessRequestResult, 1),
		Timestamp:  time.Now(),
	}

	// Send to access queue
	select {
	case d.accessQueue <- accessReq:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Wait for result
	select {
	case result := <-accessReq.ResultChan:
		if result.Error != nil {
			return nil, result.Error
		}

		resp := &pb.GrantAccessResponse{
			Granted:      result.Granted,
			Decision:     result.Decision.Decision,
			Reason:       result.Decision.Reason,
			DoorUnlocked: result.DoorUnlocked,
			Timestamp:    timestamppb.New(result.Timestamp),
			EventId:      result.EventID,
		}

		if result.Granted {
			resp.UserInfo = result.Decision.CredentialInfo
		}

		return resp, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetControllerStatus returns controller status
func (d *Driver) GetControllerStatus(ctx context.Context, req *pb.GetControllerStatusRequest) (*pb.ControllerStatus, error) {
	d.statusMu.RLock()
	status := d.status
	d.statusMu.RUnlock()

	// Collect door statuses
	doors := make([]*pb.DoorStatus, 0, len(d.doorControllers))
	d.doorMu.RLock()
	for _, doorState := range d.doorControllers {
		doors = append(doors, doorState.GetStatus())
	}
	d.doorMu.RUnlock()

	return &pb.ControllerStatus{
		DeviceId:       d.id,
		Name:           d.name,
		Status:         convertDeviceStatus(status),
		Type:           d.config.ControllerType,
		Doors:          doors,
		Model:          d.config.Model,
		Manufacturer:   d.config.Manufacturer,
		ConnectionType: d.config.Transport,
		LockdownActive: d.lockdownState.IsActive(),
		LockdownLevel:  d.lockdownState.GetLevel(),
		Statistics:     d.stats.ToProto(),
	}, nil
}

// SetDoorMode sets the door operating mode
func (d *Driver) SetDoorMode(ctx context.Context, req *pb.SetDoorModeRequest) (*pb.SetDoorModeResponse, error) {
	doorID := req.DoorId
	if doorID == "" && len(d.config.Doors) > 0 {
		doorID = d.config.Doors[0].ID
	}

	d.doorMu.Lock()
	doorState, exists := d.doorControllers[doorID]
	if !exists {
		d.doorMu.Unlock()
		return nil, fmt.Errorf("door not found: %s", doorID)
	}
	doorState.SetMode(req.Mode)
	d.doorMu.Unlock()

	// Apply mode change
	switch req.Mode {
	case pb.DoorMode_DOOR_MODE_UNLOCKED:
		d.unlockDoorIndefinite(doorID)
	case pb.DoorMode_DOOR_MODE_LOCKED:
		d.lockDoorImmediate(doorID)
	case pb.DoorMode_DOOR_MODE_NORMAL:
		// Lock door (access control will handle unlocks)
		d.lockDoorImmediate(doorID)
	}

	d.logger.Info("door mode changed",
		telemetry.String("door_id", doorID),
		telemetry.String("mode", req.Mode.String()),
	)

	// Emit event
	d.emitAccessEvent(&pb.AccessEvent{
		EventId:   generateEventID(),
		DeviceId:  d.id,
		DoorId:    doorID,
		Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_MODE_CHANGED,
		Timestamp: timestamppb.Now(),
		Severity:  pb.EventSeverity_EVENT_SEVERITY_INFO,
	})

	return &pb.SetDoorModeResponse{
		Success:     true,
		CurrentMode: req.Mode,
	}, nil
}

// ActivateLockdown activates emergency lockdown
func (d *Driver) ActivateLockdown(ctx context.Context, req *pb.ActivateLockdownRequest) (*pb.ActivateLockdownResponse, error) {
	d.logger.Warn("lockdown activated",
		telemetry.String("level", req.Level.String()),
		telemetry.String("reason", req.Reason),
		telemetry.String("initiated_by", req.InitiatedBy),
	)

	// Determine affected doors
	affectedDoors := req.DoorIds
	if len(affectedDoors) == 0 {
		// All doors
		for doorID := range d.doorControllers {
			affectedDoors = append(affectedDoors, doorID)
		}
	}

	// Activate lockdown state
	d.lockdownState.Activate(req.Level, req.Reason, req.InitiatedBy, affectedDoors)

	// Apply lockdown based on level
	doorsLocked := make([]string, 0)

	switch req.Level {
	case pb.LockdownLevel_LOCKDOWN_LEVEL_FULL, pb.LockdownLevel_LOCKDOWN_LEVEL_PARTIAL:
		// Lock all affected doors
		for _, doorID := range affectedDoors {
			if err := d.lockDoorImmediate(doorID); err != nil {
				d.logger.Error("failed to lock door during lockdown",
					telemetry.String("door_id", doorID),
					telemetry.String("error", err.Error()),
				)
			} else {
				doorsLocked = append(doorsLocked, doorID)
			}
		}

	case pb.LockdownLevel_LOCKDOWN_LEVEL_EVACUATION:
		// Unlock all doors for evacuation
		for _, doorID := range affectedDoors {
			if err := d.unlockDoorIndefinite(doorID); err != nil {
				d.logger.Error("failed to unlock door during evacuation",
					telemetry.String("door_id", doorID),
					telemetry.String("error", err.Error()),
				)
			} else {
				doorsLocked = append(doorsLocked, doorID)
			}
		}
	}

	// Emit lockdown event
	d.emitAccessEvent(&pb.AccessEvent{
		EventId:   generateEventID(),
		DeviceId:  d.id,
		Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_LOCKDOWN_ACTIVATED,
		Timestamp: timestamppb.Now(),
		Severity:  pb.EventSeverity_EVENT_SEVERITY_EMERGENCY,
		EventData: &pb.AccessEvent_Lockdown{
			Lockdown: &pb.LockdownEvent{
				Activated:   true,
				Level:       req.Level,
				Reason:      req.Reason,
				InitiatedBy: req.InitiatedBy,
			},
		},
	})

	return &pb.ActivateLockdownResponse{
		Success:     true,
		DoorsLocked: doorsLocked,
		Timestamp:   timestamppb.Now(),
	}, nil
}

// DeactivateLockdown deactivates emergency lockdown
func (d *Driver) DeactivateLockdown(ctx context.Context, req *pb.DeactivateLockdownRequest) (*pb.DeactivateLockdownResponse, error) {
	d.logger.Info("lockdown deactivated",
		telemetry.String("reason", req.Reason),
		telemetry.String("deactivated_by", req.DeactivatedBy),
	)

	affectedDoors := d.lockdownState.AffectedDoors

	// Deactivate lockdown
	d.lockdownState.Deactivate()

	// Restore doors to normal mode
	doorsRestored := make([]string, 0)
	for _, doorID := range affectedDoors {
		d.doorMu.Lock()
		if doorState, exists := d.doorControllers[doorID]; exists {
			doorState.SetMode(pb.DoorMode_DOOR_MODE_NORMAL)
			d.lockDoorImmediate(doorID)
			doorsRestored = append(doorsRestored, doorID)
		}
		d.doorMu.Unlock()
	}

	// Emit event
	d.emitAccessEvent(&pb.AccessEvent{
		EventId:   generateEventID(),
		DeviceId:  d.id,
		Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_LOCKDOWN_DEACTIVATED,
		Timestamp: timestamppb.Now(),
		Severity:  pb.EventSeverity_EVENT_SEVERITY_WARNING,
		EventData: &pb.AccessEvent_Lockdown{
			Lockdown: &pb.LockdownEvent{
				Activated: false,
			},
		},
	})

	return &pb.DeactivateLockdownResponse{
		Success:       true,
		DoorsRestored: doorsRestored,
		Timestamp:     timestamppb.Now(),
	}, nil
}

// ==================================================
// Worker Goroutines
// ==================================================

// doorMonitorLoop monitors door status
func (d *Driver) doorMonitorLoop() {
	defer d.wg.Done()

	d.logger.Debug("door monitor loop started")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.checkDoorStatuses()
		}
	}
}

// checkDoorStatuses checks all door statuses for alarms
func (d *Driver) checkDoorStatuses() {
	// TODO: Implement door sensor checking
	// - Check for forced door (door open when locked)
	// - Check for door held open (door open too long)
	// - Check tamper switches
	// - Check REX buttons
}

// accessRequestWorker processes access requests
func (d *Driver) accessRequestWorker() {
	defer d.wg.Done()

	d.logger.Debug("access request worker started")

	for {
		select {
		case <-d.ctx.Done():
			return
		case req := <-d.accessQueue:
			d.processAccessRequest(req)
		}
	}
}

// processAccessRequest processes a single access request
func (d *Driver) processAccessRequest(req *AccessRequest) {
	// Check access using decision engine
	decision, err := d.decisionEngine.CheckAccess(req.Credential, req.DoorID, req.Context)
	if err != nil {
		req.ResultChan <- &AccessRequestResult{
			Granted:   false,
			Error:     err,
			Timestamp: time.Now(),
		}
		return
	}

	granted := decision.Decision == pb.AccessDecision_ACCESS_DECISION_GRANTED
	eventID := generateEventID()

	// Update statistics
	d.stats.IncrementAccessAttempts(granted)
	if stats, exists := d.doorStats[req.DoorID]; exists {
		if granted {
			stats.IncrementGrants()
		} else {
			stats.IncrementDenials()
		}
	}

	// Emit event
	eventType := pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_DENIED
	severity := pb.EventSeverity_EVENT_SEVERITY_WARNING

	if granted {
		eventType = pb.AccessEventType_ACCESS_EVENT_TYPE_ACCESS_GRANTED
		severity = pb.EventSeverity_EVENT_SEVERITY_INFO
	}

	event := &pb.AccessEvent{
		EventId:        eventID,
		DeviceId:       d.id,
		DoorId:         req.DoorID,
		Type:           eventType,
		Timestamp:      timestamppb.New(req.Timestamp),
		Credential:     req.Credential,
		UserInfo:       decision.CredentialInfo,
		Decision:       decision.Decision,
		Reason:         decision.Reason,
		Severity:       severity,
	}

	d.emitAccessEvent(event)
	d.eventLogger.LogEvent(d.ctx, event)

	// Unlock door if granted
	doorUnlocked := false
	if granted {
		if err := d.unlockDoorDuration(req.DoorID, time.Duration(d.config.DefaultUnlockDuration)*time.Second); err == nil {
			doorUnlocked = true

			// Update anti-passback if enabled
			if d.config.EnableAntiPassback && decision.CredentialInfo != nil {
				// Determine zone from door config
				for _, doorCfg := range d.config.Doors {
					if doorCfg.ID == req.DoorID {
						d.antiPassback.UpdateLocation(decision.CredentialInfo.UserId, doorCfg.Zone)
						break
					}
				}
			}
		}
	}

	req.ResultChan <- &AccessRequestResult{
		Granted:      granted,
		Decision:     decision,
		DoorUnlocked: doorUnlocked,
		EventID:      eventID,
		Timestamp:    time.Now(),
	}
}

// unlockRequestWorker processes unlock requests
func (d *Driver) unlockRequestWorker() {
	defer d.wg.Done()

	d.logger.Debug("unlock request worker started")

	for {
		select {
		case <-d.ctx.Done():
			return
		case req := <-d.unlockQueue:
			d.processUnlockRequest(req)
		}
	}
}

// processUnlockRequest processes a single unlock request
func (d *Driver) processUnlockRequest(req *UnlockRequest) {
	var err error

	if req.Duration == 0 {
		// Indefinite unlock
		err = d.unlockDoorIndefinite(req.DoorID)
	} else {
		// Timed unlock
		err = d.unlockDoorDuration(req.DoorID, req.Duration)
	}

	if err != nil {
		req.ResultChan <- &UnlockResult{
			Success:   false,
			Error:     err,
			Timestamp: time.Now(),
		}
		return
	}

	// Update statistics
	if stats, exists := d.doorStats[req.DoorID]; exists {
		stats.IncrementUnlocks()
	}

	// Emit event
	d.emitAccessEvent(&pb.AccessEvent{
		EventId:   generateEventID(),
		DeviceId:  d.id,
		DoorId:    req.DoorID,
		Type:      pb.AccessEventType_ACCESS_EVENT_TYPE_DOOR_UNLOCKED,
		Timestamp: timestamppb.Now(),
		Reason:    req.Reason,
		Severity:  pb.EventSeverity_EVENT_SEVERITY_INFO,
	})

	req.ResultChan <- &UnlockResult{
		Success:   true,
		Duration:  req.Duration,
		Timestamp: time.Now(),
	}
}

// ==================================================
// Door Control Helpers
// ==================================================

// unlockDoorDuration unlocks a door for specified duration
func (d *Driver) unlockDoorDuration(doorID string, duration time.Duration) error {
	d.doorMu.Lock()
	doorState, exists := d.doorControllers[doorID]
	if !exists {
		d.doorMu.Unlock()
		return fmt.Errorf("door not found: %s", doorID)
	}
	doorState.SetLockStatus(pb.LockStatus_LOCK_STATUS_UNLOCKED)
	d.doorMu.Unlock()

	d.logger.Debug("door unlocked",
		telemetry.String("door_id", doorID),
		telemetry.Int("duration_sec", int(duration.Seconds())),
	)

	// TODO: Send unlock command to hardware

	// Schedule re-lock
	if duration > 0 {
		go func() {
			time.Sleep(duration)
			d.lockDoorImmediate(doorID)
		}()
	}

	return nil
}

// unlockDoorIndefinite unlocks a door indefinitely
func (d *Driver) unlockDoorIndefinite(doorID string) error {
	return d.unlockDoorDuration(doorID, 0)
}

// lockDoorImmediate locks a door immediately
func (d *Driver) lockDoorImmediate(doorID string) error {
	d.doorMu.Lock()
	doorState, exists := d.doorControllers[doorID]
	if !exists {
		d.doorMu.Unlock()
		return fmt.Errorf("door not found: %s", doorID)
	}
	doorState.SetLockStatus(pb.LockStatus_LOCK_STATUS_LOCKED)
	d.doorMu.Unlock()

	d.logger.Debug("door locked",
		telemetry.String("door_id", doorID),
	)

	// TODO: Send lock command to hardware

	return nil
}

// ==================================================
// Event Emission
// ==================================================

// emitAccessEvent emits an access event
func (d *Driver) emitAccessEvent(event *pb.AccessEvent) {
	select {
	case d.accessEventChan <- event:
	default:
		d.logger.Warn("access event channel full, dropping event")
	}
}

// ==================================================
// Helper Functions
// ==================================================

func convertDeviceStatus(status devices.DeviceStatus) pb.DeviceStatus {
	switch status {
	case devices.StatusReady:
		return pb.DeviceStatus_DEVICE_STATUS_READY
	case devices.StatusBusy:
		return pb.DeviceStatus_DEVICE_STATUS_BUSY
	case devices.StatusError:
		return pb.DeviceStatus_DEVICE_STATUS_ERROR
	case devices.StatusDisconnected:
		return pb.DeviceStatus_DEVICE_STATUS_DISCONNECTED
	default:
		return pb.DeviceStatus_DEVICE_STATUS_UNKNOWN
	}
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

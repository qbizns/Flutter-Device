package access_control

import (
	"context"
	"sync"
	"time"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// Config holds access control configuration
type Config struct {
	// Controller identification
	ID   string
	Name string
	Kind string

	// Connection settings
	Transport string // "wiegand", "rs485", "tcp", "serial"
	Address   string // For TCP controllers
	Port      int    // For TCP controllers
	SerialDev string // For serial controllers
	BaudRate  int    // For serial controllers

	// Wiegand settings (for Wiegand input/output)
	WiegandGPIOData0 int // GPIO pin for Wiegand data 0
	WiegandGPIOData1 int // GPIO pin for Wiegand data 1
	WiegandBits      int // Wiegand format (26, 37, etc.)

	// RS-485 settings
	RS485Address int // Device address on RS-485 bus

	// Controller type
	ControllerType pb.ControllerType

	// Doors managed by this controller
	Doors []DoorConfig

	// Default settings
	DefaultUnlockDuration int32 // Seconds
	DoorHeldOpenThreshold int32 // Seconds before alarm
	RexUnlockDuration     int32 // Request to exit unlock duration

	// Access control settings
	EnableAntiPassback    bool
	EnableTwoPersonRule   bool
	RequirePinWithCard    bool
	MaxFailedAttempts     int32
	LockoutDuration       int32 // Seconds

	// Model/manufacturer info
	Model        string
	Manufacturer string
}

// DoorConfig holds configuration for a single door
type DoorConfig struct {
	// Door identification
	ID   string
	Name string

	// Lock type
	LockType LockType

	// Lock output (GPIO pin, relay number, etc.)
	LockOutput int

	// Door sensor input
	DoorSensorInput int

	// Request to exit (REX) button input
	RexButtonInput int

	// Tamper switch input
	TamperInput int

	// Initial mode
	Mode pb.DoorMode

	// Schedule ID for office mode
	ScheduleID string

	// Zone/area
	Zone string
}

// LockType specifies the type of door lock
type LockType string

const (
	LockTypeElectricStrike LockType = "electric_strike"   // Electric strike (fail-secure)
	LockTypeMagneticLock   LockType = "magnetic_lock"     // Magnetic lock (fail-safe)
	LockTypeMotorizedLock  LockType = "motorized_lock"    // Motorized deadbolt
	LockTypeRelay          LockType = "relay"             // Generic relay output
	LockTypeWiegandOutput  LockType = "wiegand_output"    // Wiegand output to downstream device
)

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Transport:              "tcp",
		ControllerType:         pb.ControllerType_CONTROLLER_TYPE_SINGLE_DOOR,
		DefaultUnlockDuration:  5,
		DoorHeldOpenThreshold:  30,
		RexUnlockDuration:      5,
		EnableAntiPassback:     false,
		EnableTwoPersonRule:    false,
		RequirePinWithCard:     false,
		MaxFailedAttempts:      3,
		LockoutDuration:        60,
		Doors: []DoorConfig{
			{
				ID:       "door-1",
				Name:     "Main Door",
				LockType: LockTypeElectricStrike,
				Mode:     pb.DoorMode_DOOR_MODE_NORMAL,
			},
		},
	}
}

// Connection interface for controller communication
type Connection interface {
	// Write sends data to the controller
	Write(data []byte) (int, error)

	// Read reads data from the controller
	Read(buf []byte) (int, error)

	// Close closes the connection
	Close() error

	// IsConnected checks if the connection is active
	IsConnected() bool

	// SetTimeout sets read/write timeout
	SetTimeout(timeout time.Duration) error
}

// DoorController interface for controlling door locks
type DoorController interface {
	// UnlockDoor unlocks a door for specified duration
	UnlockDoor(doorID string, duration time.Duration) error

	// LockDoor locks a door immediately
	LockDoor(doorID string) error

	// GetDoorStatus returns current door status
	GetDoorStatus(doorID string) (*DoorState, error)

	// SetDoorMode sets door operating mode
	SetDoorMode(doorID string, mode pb.DoorMode) error
}

// AccessDecisionEngine makes access control decisions
type AccessDecisionEngine interface {
	// CheckAccess checks if access should be granted
	CheckAccess(credential *pb.Credential, door string, context *pb.AccessContext) (*AccessDecisionResult, error)

	// AddRule adds an access rule
	AddRule(rule *pb.AccessRule) error

	// RemoveRule removes an access rule
	RemoveRule(ruleID string) error

	// GetRules returns all rules
	GetRules() []*pb.AccessRule
}

// AccessDecisionResult contains the result of an access decision
type AccessDecisionResult struct {
	Decision       pb.AccessDecision
	Reason         string
	CredentialInfo *pb.CredentialInfo
	RuleID         string
	Timestamp      time.Time
}

// DoorState tracks the current state of a door
type DoorState struct {
	DoorID        string
	DoorName      string
	LockStatus    pb.LockStatus
	DoorPosition  pb.DoorPosition
	Mode          pb.DoorMode
	LockedSince   time.Time
	UnlockedSince time.Time
	AlarmStatus   pb.AlarmStatus
	LastAccess    *pb.AccessEvent
	mu            sync.RWMutex
}

// SetLockStatus sets the lock status
func (s *DoorState) SetLockStatus(status pb.LockStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.LockStatus = status
	if status == pb.LockStatus_LOCK_STATUS_LOCKED {
		s.LockedSince = time.Now()
	} else if status == pb.LockStatus_LOCK_STATUS_UNLOCKED {
		s.UnlockedSince = time.Now()
	}
}

// SetDoorPosition sets the door position
func (s *DoorState) SetDoorPosition(position pb.DoorPosition) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DoorPosition = position
}

// SetMode sets the door mode
func (s *DoorState) SetMode(mode pb.DoorMode) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Mode = mode
}

// SetAlarmStatus sets the alarm status
func (s *DoorState) SetAlarmStatus(status pb.AlarmStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AlarmStatus = status
}

// GetStatus returns current status as protobuf
func (s *DoorState) GetStatus() *pb.DoorStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &pb.DoorStatus{
		DoorId:       s.DoorID,
		DoorName:     s.DoorName,
		LockStatus:   s.LockStatus,
		DoorPosition: s.DoorPosition,
		Mode:         s.Mode,
		AlarmStatus:  s.AlarmStatus,
		LastAccess:   s.LastAccess,
	}
}

// AccessRequest represents an access request
type AccessRequest struct {
	ID         string
	DoorID     string
	Credential *pb.Credential
	Context    *pb.AccessContext
	ResultChan chan *AccessRequestResult
	Timestamp  time.Time
}

// AccessRequestResult contains the result of an access request
type AccessRequestResult struct {
	Granted        bool
	Decision       *AccessDecisionResult
	DoorUnlocked   bool
	Error          error
	EventID        string
	Timestamp      time.Time
}

// UnlockRequest represents a door unlock request
type UnlockRequest struct {
	DoorID     string
	Duration   time.Duration
	Reason     string
	TriggeredBy string
	ResultChan chan *UnlockResult
	Timestamp  time.Time
}

// UnlockResult contains the result of an unlock request
type UnlockResult struct {
	Success   bool
	Duration  time.Duration
	Error     error
	Timestamp time.Time
}

// LockdownState tracks lockdown status
type LockdownState struct {
	Active      bool
	Level       pb.LockdownLevel
	Reason      string
	InitiatedBy string
	ActivatedAt time.Time
	AffectedDoors []string
	mu          sync.RWMutex
}

// Activate activates lockdown
func (s *LockdownState) Activate(level pb.LockdownLevel, reason string, initiatedBy string, doors []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = true
	s.Level = level
	s.Reason = reason
	s.InitiatedBy = initiatedBy
	s.ActivatedAt = time.Now()
	s.AffectedDoors = doors
}

// Deactivate deactivates lockdown
func (s *LockdownState) Deactivate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Active = false
	s.Level = pb.LockdownLevel_LOCKDOWN_LEVEL_UNKNOWN
	s.Reason = ""
	s.InitiatedBy = ""
}

// IsActive checks if lockdown is active
func (s *LockdownState) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Active
}

// GetLevel returns current lockdown level
func (s *LockdownState) GetLevel() pb.LockdownLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Level
}

// ControllerStatistics tracks controller statistics
type ControllerStatistics struct {
	TotalAccessAttempts int64
	TotalGrants         int64
	TotalDenials        int64
	TotalAlarms         int64
	OnlineSince         time.Time
	mu                  sync.RWMutex
}

// IncrementAccessAttempts increments access attempt counter
func (s *ControllerStatistics) IncrementAccessAttempts(granted bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalAccessAttempts++
	if granted {
		s.TotalGrants++
	} else {
		s.TotalDenials++
	}
}

// IncrementAlarms increments alarm counter
func (s *ControllerStatistics) IncrementAlarms() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalAlarms++
}

// ToProto converts statistics to protobuf
func (s *ControllerStatistics) ToProto() *pb.ControllerStatistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &pb.ControllerStatistics{
		TotalAccessAttempts: s.TotalAccessAttempts,
		TotalGrants:         s.TotalGrants,
		TotalDenials:        s.TotalDenials,
		TotalAlarms:         s.TotalAlarms,
	}
}

// DoorStatisticsTracker tracks statistics for a door
type DoorStatisticsTracker struct {
	TotalGrants  int64
	TotalDenials int64
	TotalAlarms  int64
	TotalUnlocks int64
	LastGrant    time.Time
	LastDenial   time.Time
	LastAlarm    time.Time
	StartedAt    time.Time
	mu           sync.RWMutex
}

// IncrementGrants increments grant counter
func (t *DoorStatisticsTracker) IncrementGrants() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.TotalGrants++
	t.LastGrant = time.Now()
}

// IncrementDenials increments denial counter
func (t *DoorStatisticsTracker) IncrementDenials() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.TotalDenials++
	t.LastDenial = time.Now()
}

// IncrementAlarms increments alarm counter
func (t *DoorStatisticsTracker) IncrementAlarms() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.TotalAlarms++
	t.LastAlarm = time.Now()
}

// IncrementUnlocks increments unlock counter
func (t *DoorStatisticsTracker) IncrementUnlocks() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.TotalUnlocks++
}

// ToProto converts statistics to protobuf
func (t *DoorStatisticsTracker) ToProto() *pb.DoorStatistics {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return &pb.DoorStatistics{
		TotalGrants:  t.TotalGrants,
		TotalDenials: t.TotalDenials,
		TotalAlarms:  t.TotalAlarms,
		TotalUnlocks: t.TotalUnlocks,
	}
}

// EventLogger logs access events to storage
type EventLogger interface {
	// LogEvent logs an access event
	LogEvent(ctx context.Context, event *pb.AccessEvent) error

	// GetEvents retrieves events from log
	GetEvents(ctx context.Context, filter *EventFilter) ([]*pb.AccessEvent, error)
}

// EventFilter filters access events
type EventFilter struct {
	DeviceID   string
	DoorIDs    []string
	StartTime  time.Time
	EndTime    time.Time
	EventTypes []pb.AccessEventType
	UserIDs    []string
	PageSize   int32
	PageToken  string
}

// AntiPassbackTracker tracks user locations for anti-passback
type AntiPassbackTracker struct {
	userLocations map[string]string // user ID -> zone
	mu            sync.RWMutex
}

// NewAntiPassbackTracker creates a new anti-passback tracker
func NewAntiPassbackTracker() *AntiPassbackTracker {
	return &AntiPassbackTracker{
		userLocations: make(map[string]string),
	}
}

// UpdateLocation updates user location
func (t *AntiPassbackTracker) UpdateLocation(userID string, zone string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.userLocations[userID] = zone
}

// GetLocation gets user current location
func (t *AntiPassbackTracker) GetLocation(userID string) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	zone, ok := t.userLocations[userID]
	return zone, ok
}

// CheckTransition checks if zone transition is allowed
func (t *AntiPassbackTracker) CheckTransition(userID string, fromZone string, toZone string) bool {
	currentZone, exists := t.GetLocation(userID)
	if !exists {
		// First entry, allow
		return true
	}

	// User must be in the expected fromZone
	return currentZone == fromZone
}

// WiegandData represents Wiegand format data
type WiegandData struct {
	FacilityCode int
	CardNumber   int
	Bits         int
	RawData      []byte
}

// ParseWiegand26 parses 26-bit Wiegand format
func ParseWiegand26(data uint32) *WiegandData {
	// 26-bit format: [parity][8-bit facility][16-bit card][parity]
	facilityCode := int((data >> 17) & 0xFF)
	cardNumber := int((data >> 1) & 0xFFFF)

	return &WiegandData{
		FacilityCode: facilityCode,
		CardNumber:   cardNumber,
		Bits:         26,
	}
}

// ParseWiegand37 parses 37-bit Wiegand format (HID Corporate 1000)
func ParseWiegand37(data uint64) *WiegandData {
	// 37-bit format: [parity][12-bit facility][20-bit card][parity bits]
	facilityCode := int((data >> 23) & 0xFFF)
	cardNumber := int((data >> 1) & 0xFFFFF)

	return &WiegandData{
		FacilityCode: facilityCode,
		CardNumber:   cardNumber,
		Bits:         37,
	}
}

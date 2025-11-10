package devices

import (
	"context"
	"time"
)

// Device is the base interface all devices must implement
type Device interface {
	// ID returns the unique device identifier
	ID() string

	// Kind returns the device type (e.g., "printer.escpos", "scale.serial")
	Kind() string

	// Name returns the human-readable device name
	Name() string

	// Start initializes the device and begins operation
	Start(ctx context.Context) error

	// Stop gracefully shuts down the device
	Stop() error

	// Health returns the current health status
	Health() DeviceHealth

	// Metadata returns device-specific metadata
	Metadata() map[string]string
}

// DeviceHealth represents device health status
type DeviceHealth struct {
	Status    HealthStatus
	Message   string
	Timestamp time.Time
	Error     error
}

// HealthStatus represents device status
type HealthStatus int

const (
	HealthUnknown HealthStatus = iota
	HealthReady
	HealthDegraded
	HealthOffline
	HealthError
)

// String returns string representation of health status
func (h HealthStatus) String() string {
	switch h {
	case HealthUnknown:
		return "unknown"
	case HealthReady:
		return "ready"
	case HealthDegraded:
		return "degraded"
	case HealthOffline:
		return "offline"
	case HealthError:
		return "error"
	default:
		return "unknown"
	}
}

// ToProtoStatus converts to proto enum value
func (h HealthStatus) ToProtoStatus() int {
	return int(h)
}

// BaseDevice provides common functionality for all devices
type BaseDevice struct {
	id       string
	kind     string
	name     string
	metadata map[string]string
	health   DeviceHealth
}

// NewBaseDevice creates a new base device
func NewBaseDevice(id, kind, name string, metadata map[string]string) *BaseDevice {
	return &BaseDevice{
		id:       id,
		kind:     kind,
		name:     name,
		metadata: metadata,
		health: DeviceHealth{
			Status:    HealthUnknown,
			Timestamp: time.Now(),
		},
	}
}

// ID returns device ID
func (b *BaseDevice) ID() string {
	return b.id
}

// Kind returns device kind
func (b *BaseDevice) Kind() string {
	return b.kind
}

// Name returns device name
func (b *BaseDevice) Name() string {
	return b.name
}

// Health returns current health
func (b *BaseDevice) Health() DeviceHealth {
	return b.health
}

// Metadata returns device metadata
func (b *BaseDevice) Metadata() map[string]string {
	return b.metadata
}

// SetHealth updates device health status
func (b *BaseDevice) SetHealth(status HealthStatus, message string, err error) {
	b.health = DeviceHealth{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Error:     err,
	}
}

// ConnectionInfo contains device connection details
type ConnectionInfo struct {
	Transport string // usb, serial, tcp
	Address   string // IP address or device path
	Port      int    // Port for TCP
	BaudRate  int    // Baud rate for serial
	VendorID  int    // USB vendor ID
	ProductID int    // USB product ID
}

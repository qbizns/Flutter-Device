package scanner

import (
	"context"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
)

// Scanner interface extends Device with scan event streaming
type Scanner interface {
	devices.Device

	// Events returns a channel of scan events
	Events(ctx context.Context) (<-chan ScanEvent, error)

	// Configure applies scanner settings
	Configure(ctx context.Context, config ScannerConfig) error
}

// ScanEvent represents a barcode/QR scan event
type ScanEvent struct {
	DeviceID  string
	Data      string
	Symbology string
	Timestamp time.Time
	EventID   string
}

// ScannerConfig contains scanner configuration
type ScannerConfig struct {
	Prefix  string
	Suffix  string
	Timeout time.Duration
}

package virtual_devices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// VirtualDisplay simulates a customer-facing display
type VirtualDisplay struct {
	id      string
	name    string
	logger  *telemetry.Logger
	mu      sync.RWMutex
	status  devices.HealthStatus
	line1   string // Upper display line
	line2   string // Lower display line
	cleared bool
}

// NewVirtualDisplay creates a new virtual display device
func NewVirtualDisplay(id, name string, logger *telemetry.Logger) *VirtualDisplay {
	return &VirtualDisplay{
		id:      id,
		name:    name,
		logger:  logger,
		status:  devices.HealthUnknown,
		cleared: true,
	}
}

// ID returns the device ID
func (d *VirtualDisplay) ID() string {
	return d.id
}

// Name returns the device name
func (d *VirtualDisplay) Name() string {
	return d.name
}

// Kind returns the device kind
func (d *VirtualDisplay) Kind() string {
	return "display.virtual"
}

// Start initializes the virtual display
func (d *VirtualDisplay) Start(ctx context.Context) error {
	d.logger.Info("starting virtual display",
		telemetry.String("device_id", d.id),
		telemetry.String("name", d.name),
	)

	d.mu.Lock()
	d.status = devices.HealthReady
	d.line1 = ""
	d.line2 = ""
	d.cleared = true
	d.mu.Unlock()

	return nil
}

// Stop shuts down the virtual display
func (d *VirtualDisplay) Stop() error {
	d.logger.Info("stopping virtual display",
		telemetry.String("device_id", d.id),
	)

	d.mu.Lock()
	d.status = devices.HealthOffline
	d.mu.Unlock()

	return nil
}

// Health returns the device health status
func (d *VirtualDisplay) Health() devices.DeviceHealth {
	d.mu.RLock()
	status := d.status
	d.mu.RUnlock()

	return devices.DeviceHealth{
		Status:    status,
		Message:   "Display operational",
		Timestamp: time.Now(),
	}
}

// Metadata returns device metadata
func (d *VirtualDisplay) Metadata() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]string{
		"type":       "customer_display",
		"lines":      "2",
		"columns":    "20",
		"line1":      d.line1,
		"line2":      d.line2,
		"cleared":    fmt.Sprintf("%v", d.cleared),
		"brightness": "100",
	}
}

// Show displays text on the screen
func (d *VirtualDisplay) Show(line1, line2 string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.line1 = line1
	d.line2 = line2
	d.cleared = false

	d.logger.Info("virtual display showing text",
		telemetry.String("device_id", d.id),
		telemetry.String("line1", line1),
		telemetry.String("line2", line2),
	)

	// Log to console for visibility during testing
	fmt.Printf("\n")
	fmt.Printf("┌──────────────────────┐\n")
	fmt.Printf("│ %-20s │\n", truncate(line1, 20))
	fmt.Printf("│ %-20s │\n", truncate(line2, 20))
	fmt.Printf("└──────────────────────┘\n")
	fmt.Printf("\n")

	return nil
}

// Clear clears the display
func (d *VirtualDisplay) Clear() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.line1 = ""
	d.line2 = ""
	d.cleared = true

	d.logger.Info("virtual display cleared",
		telemetry.String("device_id", d.id),
	)

	// Log to console
	fmt.Printf("\n")
	fmt.Printf("┌──────────────────────┐\n")
	fmt.Printf("│                      │\n")
	fmt.Printf("│                      │\n")
	fmt.Printf("└──────────────────────┘\n")
	fmt.Printf("  [Display Cleared]\n")
	fmt.Printf("\n")

	return nil
}

// GetText returns the currently displayed text
func (d *VirtualDisplay) GetText() (string, string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.line1, d.line2
}

// SetBrightness sets the display brightness (0-100)
func (d *VirtualDisplay) SetBrightness(level int) error {
	if level < 0 || level > 100 {
		return fmt.Errorf("brightness must be between 0 and 100")
	}

	d.logger.Info("virtual display brightness set",
		telemetry.String("device_id", d.id),
		telemetry.Int("level", level),
	)

	return nil
}

// truncate truncates or pads a string to the specified length
func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen]
	}
	// Pad with spaces
	for len(s) < maxLen {
		s += " "
	}
	return s
}

package virtual_devices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// VirtualDrawer simulates a cash drawer
type VirtualDrawer struct {
	id          string
	name        string
	logger      *telemetry.Logger
	mu          sync.RWMutex
	status      devices.HealthStatus
	isOpen      bool
	lastOpened  time.Time
	openCount   int
	autoCloseAt *time.Time
}

// NewVirtualDrawer creates a new virtual drawer device
func NewVirtualDrawer(id, name string, logger *telemetry.Logger) *VirtualDrawer {
	return &VirtualDrawer{
		id:     id,
		name:   name,
		logger: logger,
		status: devices.HealthUnknown,
	}
}

// ID returns the device ID
func (d *VirtualDrawer) ID() string {
	return d.id
}

// Name returns the device name
func (d *VirtualDrawer) Name() string {
	return d.name
}

// Kind returns the device kind
func (d *VirtualDrawer) Kind() string {
	return "drawer.virtual"
}

// Start initializes the virtual drawer
func (d *VirtualDrawer) Start(ctx context.Context) error {
	d.logger.Info("starting virtual drawer",
		telemetry.String("device_id", d.id),
		telemetry.String("name", d.name),
	)

	d.mu.Lock()
	d.status = devices.HealthReady
	d.isOpen = false
	d.openCount = 0
	d.mu.Unlock()

	// Start auto-close monitor
	go d.autoCloseMonitor(ctx)

	return nil
}

// Stop shuts down the virtual drawer
func (d *VirtualDrawer) Stop() error {
	d.logger.Info("stopping virtual drawer",
		telemetry.String("device_id", d.id),
	)

	d.mu.Lock()
	d.status = devices.HealthOffline
	d.mu.Unlock()

	return nil
}

// Health returns the device health status
func (d *VirtualDrawer) Health() devices.DeviceHealth {
	d.mu.RLock()
	status := d.status
	isOpen := d.isOpen
	d.mu.RUnlock()

	message := "Drawer closed"
	if isOpen {
		message = "Drawer open"
	}

	return devices.DeviceHealth{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// Metadata returns device metadata
func (d *VirtualDrawer) Metadata() map[string]string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]string{
		"type":       "cash_drawer",
		"is_open":    fmt.Sprintf("%v", d.isOpen),
		"open_count": fmt.Sprintf("%d", d.openCount),
		"last_opened": func() string {
			if d.lastOpened.IsZero() {
				return "never"
			}
			return d.lastOpened.Format(time.RFC3339)
		}(),
	}
}

// Open opens the cash drawer
func (d *VirtualDrawer) Open() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.isOpen {
		return fmt.Errorf("drawer is already open")
	}

	d.isOpen = true
	d.lastOpened = time.Now()
	d.openCount++

	// Auto-close after 5 seconds
	autoCloseTime := time.Now().Add(5 * time.Second)
	d.autoCloseAt = &autoCloseTime

	d.logger.Info("virtual drawer opened",
		telemetry.String("device_id", d.id),
		telemetry.Int("open_count", d.openCount),
	)

	// Visual feedback
	fmt.Printf("\n")
	fmt.Printf("┌────────────────────────────────────┐\n")
	fmt.Printf("│      💵 DRAWER OPENED 💵           │\n")
	fmt.Printf("│                                    │\n")
	fmt.Printf("│  Drawer #%d opened at %s      │\n", d.openCount, d.lastOpened.Format("15:04:05"))
	fmt.Printf("│  Will auto-close in 5 seconds...  │\n")
	fmt.Printf("└────────────────────────────────────┘\n")
	fmt.Printf("\n")

	return nil
}

// Close closes the cash drawer
func (d *VirtualDrawer) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.isOpen {
		return fmt.Errorf("drawer is already closed")
	}

	d.isOpen = false
	d.autoCloseAt = nil

	d.logger.Info("virtual drawer closed",
		telemetry.String("device_id", d.id),
	)

	// Visual feedback
	fmt.Printf("\n")
	fmt.Printf("┌────────────────────────────────────┐\n")
	fmt.Printf("│      🔒 DRAWER CLOSED 🔒           │\n")
	fmt.Printf("│                                    │\n")
	fmt.Printf("│  Drawer secured at %s         │\n", time.Now().Format("15:04:05"))
	fmt.Printf("└────────────────────────────────────┘\n")
	fmt.Printf("\n")

	return nil
}

// IsOpen returns whether the drawer is currently open
func (d *VirtualDrawer) IsOpen() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.isOpen
}

// GetOpenCount returns the number of times the drawer has been opened
func (d *VirtualDrawer) GetOpenCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.openCount
}

// autoCloseMonitor automatically closes the drawer after timeout
func (d *VirtualDrawer) autoCloseMonitor(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			d.mu.RLock()
			autoCloseAt := d.autoCloseAt
			isOpen := d.isOpen
			d.mu.RUnlock()

			if isOpen && autoCloseAt != nil && time.Now().After(*autoCloseAt) {
				d.logger.Debug("auto-closing drawer",
					telemetry.String("device_id", d.id),
				)

				if err := d.Close(); err != nil {
					d.logger.Error("failed to auto-close drawer",
						telemetry.String("device_id", d.id),
						telemetry.Error(err),
					)
				}
			}
		}
	}
}

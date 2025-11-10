package scanner_hid

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/scanner"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Driver implements USB HID barcode scanner support
//
// Supports standard USB HID POS scanners that follow the USB HID POS
// specification. Compatible with most major scanner manufacturers:
// - Symbol/Zebra (DS2208, LS2208, etc.)
// - Honeywell (Voyager, Xenon series)
// - Datalogic (QuickScan series)
// - Generic HID POS scanners
//
// The driver handles:
// - Device enumeration and hotplug detection
// - HID report parsing
// - Barcode data extraction
// - Multiple scanner support (concurrent devices)
// - Error recovery and reconnection
type Driver struct {
	*devices.BaseDevice
	logger  *telemetry.Logger
	eventBus *events.Bus

	// USB device information
	vendorID  uint16
	productID uint16
	serial    string

	// Connection state
	mu        sync.RWMutex
	connected bool
	device    USBDevice  // Platform-specific USB device handle

	// Reading state
	ctx       context.Context
	cancel    context.CancelFunc
	scanChan  chan scanner.ScanEvent

	// Configuration
	config Config
}

// Config contains scanner driver configuration
type Config struct {
	// VendorID is the USB vendor ID (e.g., 0x05e0 for Symbol)
	VendorID uint16 `json:"vendor_id" yaml:"vendor_id"`

	// ProductID is the USB product ID
	ProductID uint16 `json:"product_id" yaml:"product_id"`

	// Serial number (optional, for specific device selection)
	Serial string `json:"serial" yaml:"serial"`

	// BufferSize for scan event channel
	BufferSize int `json:"buffer_size" yaml:"buffer_size"`

	// ReconnectDelay after disconnection
	ReconnectDelay time.Duration `json:"reconnect_delay" yaml:"reconnect_delay"`

	// ReadTimeout for USB reads
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout"`
}

// DefaultConfig returns default scanner configuration
func DefaultConfig() Config {
	return Config{
		BufferSize:     10,
		ReconnectDelay: 5 * time.Second,
		ReadTimeout:    1 * time.Second,
	}
}

// USBDevice is a platform-specific USB device interface
// Implementations exist in hid_linux.go, hid_darwin.go, hid_windows.go
type USBDevice interface {
	// Read reads a HID report from the device
	Read([]byte) (int, error)

	// Close closes the device connection
	Close() error

	// GetInfo returns device information
	GetInfo() (vendorID, productID uint16, serial string, err error)
}

// NewDriver creates a new USB HID scanner driver
func NewDriver(id, name string, config Config, logger *telemetry.Logger, eventBus *events.Bus) *Driver {
	if config.BufferSize == 0 {
		config = DefaultConfig()
	}

	return &Driver{
		BaseDevice: devices.NewBaseDevice(id, "scanner.hid", name, map[string]string{
			"type":       "scanner",
			"connection": "usb",
			"protocol":   "hid",
		}),
		logger:   logger,
		eventBus: eventBus,
		config:   config,
		scanChan: make(chan scanner.ScanEvent, config.BufferSize),
	}
}

// Start initializes and starts the USB HID scanner
func (d *Driver) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.connected {
		return fmt.Errorf("scanner already started")
	}

	// Find and open USB device
	device, err := d.findDevice()
	if err != nil {
		d.SetHealth(devices.HealthError, "Failed to find USB device", err)
		return fmt.Errorf("failed to find USB device: %w", err)
	}

	// Store device info
	vendorID, productID, serial, err := device.GetInfo()
	if err != nil {
		device.Close()
		return fmt.Errorf("failed to get device info: %w", err)
	}

	d.device = device
	d.vendorID = vendorID
	d.productID = productID
	d.serial = serial
	d.connected = true

	// Start reading
	d.ctx, d.cancel = context.WithCancel(ctx)
	go d.readLoop()

	d.SetHealth(devices.HealthReady, "USB HID scanner ready", nil)

	d.logger.Info("USB HID scanner started",
		telemetry.String("device_id", d.ID()),
		telemetry.String("name", d.Name()),
		telemetry.String("vendor_id", fmt.Sprintf("0x%04x", vendorID)),
		telemetry.String("product_id", fmt.Sprintf("0x%04x", productID)),
		telemetry.String("serial", serial),
	)

	return nil
}

// Stop stops the USB HID scanner
func (d *Driver) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil
	}

	// Cancel reading
	if d.cancel != nil {
		d.cancel()
	}

	// Close device
	if d.device != nil {
		if err := d.device.Close(); err != nil {
			d.logger.Error("error closing USB device",
				telemetry.String("device_id", d.ID()),
				telemetry.Error(err),
			)
		}
		d.device = nil
	}

	d.connected = false
	d.SetHealth(devices.HealthOffline, "USB HID scanner stopped", nil)

	d.logger.Info("USB HID scanner stopped",
		telemetry.String("device_id", d.ID()),
	)

	return nil
}

// Scan returns the channel for receiving scan events
func (d *Driver) Scan() <-chan scanner.ScanEvent {
	return d.scanChan
}

// findDevice locates and opens the USB HID scanner device
func (d *Driver) findDevice() (USBDevice, error) {
	// This will be implemented in platform-specific files
	// hid_linux.go, hid_darwin.go, hid_windows.go
	return findUSBDevice(d.config)
}

// readLoop continuously reads from the USB device and emits scan events
func (d *Driver) readLoop() {
	buffer := make([]byte, 64) // Standard HID report size
	parser := NewParser()

	for {
		select {
		case <-d.ctx.Done():
			return

		default:
			// Read HID report
			n, err := d.device.Read(buffer)
			if err != nil {
				d.logger.Error("USB read error",
					telemetry.String("device_id", d.ID()),
					telemetry.Error(err),
				)

				// Try to reconnect
				d.handleDisconnect()
				return
			}

			// Parse HID report
			if n > 0 {
				barcode, symbology, ok := parser.Parse(buffer[:n])
				if ok {
					// Create scan event
					now := time.Now()
					event := scanner.ScanEvent{
						DeviceID:  d.ID(),
						Data:      barcode,
						Symbology: symbology,
						Timestamp: now,
					}

					// Send to channel (non-blocking)
					select {
					case d.scanChan <- event:
						d.logger.Info("barcode scanned",
							telemetry.String("device_id", d.ID()),
							telemetry.String("barcode", barcode),
							telemetry.String("symbology", symbology),
						)

						// Publish to event bus
						if d.eventBus != nil {
							d.eventBus.Publish(events.Event{
								Type:     "scanner.scan",
								DeviceID: d.ID(),
								Data: map[string]interface{}{
									"barcode":   barcode,
									"symbology": symbology,
									"timestamp": now,
								},
							})
						}

					default:
						d.logger.Warn("scan event dropped (channel full)",
							telemetry.String("device_id", d.ID()),
						)
					}
				}
			}
		}
	}
}

// handleDisconnect handles device disconnection and attempts reconnection
func (d *Driver) handleDisconnect() {
	d.mu.Lock()
	d.connected = false
	d.SetHealth(devices.HealthError, "USB device disconnected", nil)
	d.mu.Unlock()

	d.logger.Warn("USB device disconnected, will attempt reconnection",
		telemetry.String("device_id", d.ID()),
	)

	// Attempt reconnection
	go d.reconnectLoop()
}

// reconnectLoop attempts to reconnect to the USB device
func (d *Driver) reconnectLoop() {
	ticker := time.NewTicker(d.config.ReconnectDelay)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return

		case <-ticker.C:
			device, err := d.findDevice()
			if err != nil {
				d.logger.Debug("reconnect attempt failed",
					telemetry.String("device_id", d.ID()),
					telemetry.Error(err),
				)
				continue
			}

			d.mu.Lock()
			d.device = device
			d.connected = true
			d.SetHealth(devices.HealthReady, "USB device reconnected", nil)
			d.mu.Unlock()

			d.logger.Info("USB device reconnected",
				telemetry.String("device_id", d.ID()),
			)

			// Resume reading
			go d.readLoop()
			return
		}
	}
}

// Configure configures scanner settings (if supported by hardware)
func (d *Driver) Configure(ctx context.Context, config map[string]interface{}) error {
	// Most HID scanners don't support runtime configuration
	// Configuration is typically done via special programming barcodes
	d.logger.Info("configure requested (not supported by HID scanners)",
		telemetry.String("device_id", d.ID()),
	)

	return fmt.Errorf("runtime configuration not supported by USB HID scanners")
}

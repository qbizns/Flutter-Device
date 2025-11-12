package printer_usb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
	"github.com/Macber-eg/Flutter-Device/internal/drivers/printer_escpos"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Driver implements USB ESC/POS printer driver
//
// Supports standard USB ESC/POS receipt printers from major manufacturers:
// - Epson TM series (TM-T88, TM-T20, TM-U220, etc.)
// - Star Micronics (TSP100, TSP650, etc.)
// - Citizen CT-S series
// - Bixolon SRP series
// - Generic ESC/POS compatible printers
//
// The driver uses the USB Printer Class (0x07) interface and reuses
// the ESC/POS command renderer from the printer_escpos package.
type Driver struct {
	*devices.BaseDevice
	logger *telemetry.Logger

	// USB device information
	vendorID  uint16
	productID uint16
	serial    string

	// Connection state
	mu        sync.RWMutex
	connected bool
	device    USBDevice // Platform-specific USB device handle

	// Configuration
	config Config

	// ESC/POS renderer (reused from printer_escpos)
	renderer *printer_escpos.Driver
}

// Config contains USB printer driver configuration
type Config struct {
	// VendorID is the USB vendor ID (e.g., 0x04b8 for Epson)
	VendorID uint16 `json:"vendor_id" yaml:"vendor_id"`

	// ProductID is the USB product ID
	ProductID uint16 `json:"product_id" yaml:"product_id"`

	// Serial number (optional, for specific device selection)
	Serial string `json:"serial" yaml:"serial"`

	// Timeout for USB operations
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// ReconnectDelay after disconnection
	ReconnectDelay time.Duration `json:"reconnect_delay" yaml:"reconnect_delay"`

	// Metadata for device identification
	Metadata map[string]string `json:"metadata" yaml:"metadata"`
}

// DefaultConfig returns default USB printer configuration
func DefaultConfig() Config {
	return Config{
		Timeout:        5 * time.Second,
		ReconnectDelay: 5 * time.Second,
		Metadata:       make(map[string]string),
	}
}

// USBDevice is a platform-specific USB printer interface
// Implementations exist in usb_linux.go, usb_darwin.go, usb_windows.go
type USBDevice interface {
	// Write sends data to the printer
	Write([]byte) (int, error)

	// Read reads data from the printer (for status queries)
	Read([]byte) (int, error)

	// Close closes the device connection
	Close() error

	// GetInfo returns device information
	GetInfo() (vendorID, productID uint16, serial string, err error)

	// Reset resets the USB device
	Reset() error
}

// NewDriver creates a new USB ESC/POS printer driver
func NewDriver(id, name string, config Config, logger *telemetry.Logger) *Driver {
	if config.Timeout == 0 {
		config = DefaultConfig()
	}

	base := devices.NewBaseDevice(id, "printer.usb", name, config.Metadata)

	return &Driver{
		BaseDevice: base,
		logger:     logger.WithDeviceID(id),
		config:     config,
	}
}

// Start initializes and starts the USB printer
func (d *Driver) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.connected {
		return fmt.Errorf("printer already started")
	}

	d.logger.Info("starting USB printer driver")

	// Find and open USB device
	device, err := d.findDevice()
	if err != nil {
		d.SetHealth(devices.HealthOffline, "failed to find USB device", err)
		return fmt.Errorf("failed to find USB device: %w", err)
	}

	// Store device info
	vendorID, productID, serial, err := device.GetInfo()
	if err != nil {
		device.Close()
		d.SetHealth(devices.HealthOffline, "failed to get device info", err)
		return fmt.Errorf("failed to get device info: %w", err)
	}

	d.device = device
	d.vendorID = vendorID
	d.productID = productID
	d.serial = serial
	d.connected = true

	// Initialize printer with ESC/POS init command
	if err := d.initialize(); err != nil {
		d.SetHealth(devices.HealthDegraded, "failed to initialize", err)
		return fmt.Errorf("failed to initialize printer: %w", err)
	}

	d.SetHealth(devices.HealthReady, "printer ready", nil)

	d.logger.Info("USB printer started successfully",
		telemetry.String("vendor_id", fmt.Sprintf("0x%04x", vendorID)),
		telemetry.String("product_id", fmt.Sprintf("0x%04x", productID)),
		telemetry.String("serial", serial),
	)

	return nil
}

// Stop stops the USB printer
func (d *Driver) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.connected {
		return nil
	}

	d.logger.Info("stopping USB printer driver")

	// Close device
	if d.device != nil {
		if err := d.device.Close(); err != nil {
			d.logger.Error("error closing USB device", telemetry.Error(err))
		}
		d.device = nil
	}

	d.connected = false
	d.SetHealth(devices.HealthOffline, "stopped", nil)

	d.logger.Info("USB printer stopped")
	return nil
}

// Print sends a document to the printer
func (d *Driver) Print(ctx context.Context, doc *printer.PrintDocument) error {
	d.logger.Info("printing document")

	// Check connection
	d.mu.RLock()
	if !d.connected || d.device == nil {
		d.mu.RUnlock()
		d.SetHealth(devices.HealthOffline, "not connected", nil)
		return fmt.Errorf("printer not connected")
	}
	d.mu.RUnlock()

	// Render document to ESC/POS commands
	commands, err := renderDocument(doc)
	if err != nil {
		return fmt.Errorf("failed to render document: %w", err)
	}

	// Send to printer
	if err := d.send(ctx, commands); err != nil {
		d.SetHealth(devices.HealthDegraded, "print failed", err)
		return fmt.Errorf("failed to send to printer: %w", err)
	}

	d.logger.Info("document printed successfully",
		telemetry.Int("bytes_sent", len(commands)),
	)

	return nil
}

// OpenDrawer opens the cash drawer
func (d *Driver) OpenDrawer(ctx context.Context) error {
	d.logger.Info("opening cash drawer")

	// Check connection
	d.mu.RLock()
	if !d.connected || d.device == nil {
		d.mu.RUnlock()
		return fmt.Errorf("printer not connected")
	}
	d.mu.RUnlock()

	// ESC p 0 50 50 (connector 0, 250ms pulse)
	cmd := []byte{0x1B, 0x70, 0x00, 0x32, 0x32}

	if err := d.send(ctx, cmd); err != nil {
		return fmt.Errorf("failed to open drawer: %w", err)
	}

	d.logger.Info("cash drawer opened")
	return nil
}

// GetStatus queries printer status
func (d *Driver) GetStatus(ctx context.Context) (*printer.PrinterStatus, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.connected || d.device == nil {
		return &printer.PrinterStatus{
			Online:       false,
			PaperPresent: false,
			DrawerOpen:   false,
		}, nil
	}

	// Send status query command (DLE EOT n)
	// DLE EOT 1 - Printer status
	statusCmd := []byte{0x10, 0x04, 0x01}

	if _, err := d.device.Write(statusCmd); err != nil {
		d.logger.Error("failed to send status query", telemetry.Error(err))
		return &printer.PrinterStatus{
			Online:       true,
			PaperPresent: true, // Assume paper present on error
			DrawerOpen:   false,
		}, nil
	}

	// Read status response (1 byte)
	response := make([]byte, 1)
	n, err := d.device.Read(response)
	if err != nil || n == 0 {
		d.logger.Debug("status read failed, using defaults", telemetry.Error(err))
		return &printer.PrinterStatus{
			Online:       true,
			PaperPresent: true,
			DrawerOpen:   false,
		}, nil
	}

	// Parse status byte
	status := response[0]
	paperOut := (status & 0x20) != 0       // Bit 5: paper end sensor
	drawerOpen := (status & 0x04) != 0     // Bit 2: drawer kick-out connector pin 3

	return &printer.PrinterStatus{
		Online:       true,
		PaperPresent: !paperOut,
		DrawerOpen:   drawerOpen,
	}, nil
}

// findDevice locates and opens the USB printer device
func (d *Driver) findDevice() (USBDevice, error) {
	// This will be implemented in platform-specific files
	// usb_linux.go, usb_darwin.go, usb_windows.go
	return findUSBPrinter(d.config)
}

// initialize sends initialization commands to printer
func (d *Driver) initialize() error {
	d.logger.Info("initializing printer")

	// ESC @ - Initialize printer
	initCmd := []byte{0x1B, 0x40}

	_, err := d.device.Write(initCmd)
	if err != nil {
		return fmt.Errorf("failed to send init command: %w", err)
	}

	return nil
}

// send sends data to the printer
func (d *Driver) send(ctx context.Context, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.device == nil {
		return fmt.Errorf("device not open")
	}

	// Write data
	n, err := d.device.Write(data)
	if err != nil {
		d.logger.Error("write failed", telemetry.Error(err))
		// Connection lost, close it
		d.device.Close()
		d.device = nil
		d.connected = false
		d.SetHealth(devices.HealthOffline, "write failed", err)
		return err
	}

	if n != len(data) {
		return fmt.Errorf("short write: %d/%d bytes", n, len(data))
	}

	return nil
}

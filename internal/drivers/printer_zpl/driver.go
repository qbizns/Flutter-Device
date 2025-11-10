package printer_zpl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Config holds ZPL printer configuration
type Config struct {
	ID       string
	Name     string
	Address  string
	Port     int
	Timeout  time.Duration
	DPI      int // 203 or 300 DPI
	Width    int // Label width in dots
	Height   int // Label height in dots
	Metadata map[string]string
}

// Driver implements a ZPL label printer driver
type Driver struct {
	config Config
	logger *telemetry.Logger
	conn   net.Conn
	mu     sync.Mutex
	status devices.HealthStatus
}

// NewDriver creates a new ZPL printer driver
func NewDriver(cfg Config, logger *telemetry.Logger) *Driver {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.DPI == 0 {
		cfg.DPI = 203 // Default to 203 DPI
	}
	if cfg.Width == 0 {
		cfg.Width = 800 // Default 4" width at 203 DPI
	}
	if cfg.Height == 0 {
		cfg.Height = 600 // Default 3" height at 203 DPI
	}

	return &Driver{
		config: cfg,
		logger: logger,
		status: devices.HealthUnknown,
	}
}

// ID returns the device ID
func (d *Driver) ID() string {
	return d.config.ID
}

// Name returns the device name
func (d *Driver) Name() string {
	return d.config.Name
}

// Kind returns the device kind
func (d *Driver) Kind() string {
	return "printer.zpl"
}

// Start initializes the ZPL printer driver
func (d *Driver) Start(ctx context.Context) error {
	d.logger.Info("starting ZPL printer driver",
		telemetry.String("device_id", d.config.ID),
	)

	// Connect to printer
	if err := d.connect(); err != nil {
		d.setStatus(devices.HealthOffline)
		return fmt.Errorf("failed to connect: %w", err)
	}

	d.setStatus(devices.HealthReady)
	return nil
}

// Stop closes the connection
func (d *Driver) Stop() error {
	d.logger.Info("stopping ZPL printer driver",
		telemetry.String("device_id", d.config.ID),
	)

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}

	d.status = devices.HealthOffline
	return nil
}

// Health checks printer health
func (d *Driver) Health() devices.DeviceHealth {
	d.mu.Lock()
	status := d.status
	d.mu.Unlock()

	health := devices.DeviceHealth{
		Status:    status,
		Timestamp: time.Now(),
	}

	// Try to query printer status
	if err := d.checkConnection(); err != nil {
		health.Message = fmt.Sprintf("Connection error: %v", err)
		health.Status = devices.HealthOffline
	} else {
		health.Message = "Printer online"
	}

	return health
}

// Metadata returns device metadata
func (d *Driver) Metadata() map[string]string {
	meta := make(map[string]string)
	for k, v := range d.config.Metadata {
		meta[k] = v
	}
	meta["dpi"] = fmt.Sprintf("%d", d.config.DPI)
	meta["width"] = fmt.Sprintf("%d", d.config.Width)
	meta["height"] = fmt.Sprintf("%d", d.config.Height)
	meta["address"] = fmt.Sprintf("%s:%d", d.config.Address, d.config.Port)
	return meta
}

// connect establishes connection to the printer
func (d *Driver) connect() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn != nil {
		return nil // Already connected
	}

	address := fmt.Sprintf("%s:%d", d.config.Address, d.config.Port)
	d.logger.Info("connecting to ZPL printer",
		telemetry.String("device_id", d.config.ID),
		telemetry.String("address", address),
	)

	dialer := &net.Dialer{
		Timeout: d.config.Timeout,
	}

	conn, err := dialer.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	d.conn = conn
	d.logger.Info("connected to ZPL printer",
		telemetry.String("device_id", d.config.ID),
	)

	return nil
}

// checkConnection verifies the connection is alive
func (d *Driver) checkConnection() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn == nil {
		return fmt.Errorf("not connected")
	}

	// Send status query command
	d.conn.SetWriteDeadline(time.Now().Add(d.config.Timeout))
	_, err := d.conn.Write([]byte("~HQES\n")) // Host query status
	if err != nil {
		d.conn.Close()
		d.conn = nil
		return fmt.Errorf("write failed: %w", err)
	}

	return nil
}

// setStatus updates the device status
func (d *Driver) setStatus(status devices.HealthStatus) {
	d.mu.Lock()
	d.status = status
	d.mu.Unlock()
}

// Print sends a ZPL document to the printer
func (d *Driver) Print(ctx context.Context, data []byte) error {
	d.logger.Debug("printing ZPL label",
		telemetry.String("device_id", d.config.ID),
		telemetry.Int("size", len(data)),
	)

	// Ensure we're connected
	if err := d.connect(); err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	d.mu.Lock()
	conn := d.conn
	d.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(d.config.Timeout))

	// Send ZPL data
	_, err := conn.Write(data)
	if err != nil {
		d.setStatus(devices.HealthError)
		d.mu.Lock()
		if d.conn != nil {
			d.conn.Close()
			d.conn = nil
		}
		d.mu.Unlock()
		return fmt.Errorf("write failed: %w", err)
	}

	d.logger.Info("ZPL label printed successfully",
		telemetry.String("device_id", d.config.ID),
	)

	return nil
}

// PrintText prints plain text by converting to ZPL
func (d *Driver) PrintText(ctx context.Context, text string) error {
	zpl := d.generateTextLabel(text)
	return d.Print(ctx, []byte(zpl))
}

// generateTextLabel generates a simple ZPL label from text
func (d *Driver) generateTextLabel(text string) string {
	var buf bytes.Buffer

	// Start label
	buf.WriteString("^XA\n")

	// Set label home position
	buf.WriteString("^LH0,0\n")

	// Set field origin (50 dots from left, 50 from top)
	buf.WriteString("^FO50,50\n")

	// Set font (A = 9x5, height = 30)
	buf.WriteString("^A0N,30,30\n")

	// Field data
	buf.WriteString("^FD")
	buf.WriteString(text)
	buf.WriteString("^FS\n")

	// End label
	buf.WriteString("^XZ\n")

	return buf.String()
}

// PrintBarcode prints a barcode label
func (d *Driver) PrintBarcode(ctx context.Context, barcode, format string) error {
	zpl := d.generateBarcodeLabel(barcode, format)
	return d.Print(ctx, []byte(zpl))
}

// generateBarcodeLabel generates a ZPL barcode label
func (d *Driver) generateBarcodeLabel(barcode, format string) string {
	var buf bytes.Buffer

	// Start label
	buf.WriteString("^XA\n")

	// Set label home position
	buf.WriteString("^LH0,0\n")

	// Barcode field
	buf.WriteString("^FO50,50\n")

	// Barcode type (B = Code 39, default)
	switch format {
	case "CODE128":
		buf.WriteString("^BCN,100,Y,N,N\n") // Code 128
	case "EAN13":
		buf.WriteString("^BEN,100,Y,N\n") // EAN-13
	case "QR":
		buf.WriteString("^BQN,2,5\n") // QR Code
	default:
		buf.WriteString("^B3N,N,100,Y,N\n") // Code 39 (default)
	}

	// Field data
	buf.WriteString("^FD")
	buf.WriteString(barcode)
	buf.WriteString("^FS\n")

	// Add human-readable text below barcode
	buf.WriteString("^FO50,160\n")
	buf.WriteString("^A0N,25,25\n")
	buf.WriteString("^FD")
	buf.WriteString(barcode)
	buf.WriteString("^FS\n")

	// End label
	buf.WriteString("^XZ\n")

	return buf.String()
}

// GetStatus queries printer status
func (d *Driver) GetStatus(ctx context.Context) (string, error) {
	d.mu.Lock()
	conn := d.conn
	d.mu.Unlock()

	if conn == nil {
		return "", fmt.Errorf("not connected")
	}

	// Send status query
	conn.SetWriteDeadline(time.Now().Add(d.config.Timeout))
	_, err := conn.Write([]byte("~HQES\n"))
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	// Read response
	conn.SetReadDeadline(time.Now().Add(d.config.Timeout))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read failed: %w", err)
	}

	return string(buf[:n]), nil
}

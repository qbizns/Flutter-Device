package printer_escpos

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Driver implements ESC/POS printer driver
type Driver struct {
	*devices.BaseDevice
	config       Config
	conn         net.Conn
	mu           sync.Mutex
	logger       *telemetry.Logger
	arabicShaper *ArabicShaper // Arabic text shaping engine
}

// Config contains ESC/POS printer configuration
type Config struct {
	ID       string
	Name     string
	Address  string // IP:port for TCP
	Port     int
	Timeout  time.Duration
	Metadata map[string]string
}

// NewDriver creates a new ESC/POS printer driver
func NewDriver(config Config, logger *telemetry.Logger) *Driver {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	base := devices.NewBaseDevice(config.ID, "printer.escpos", config.Name, config.Metadata)

	return &Driver{
		BaseDevice:   base,
		config:       config,
		logger:       logger.WithDeviceID(config.ID),
		arabicShaper: NewArabicShaper(), // Initialize Arabic shaping engine
	}
}

// Start initializes the printer connection
func (d *Driver) Start(ctx context.Context) error {
	d.logger.Info("starting ESC/POS printer driver")

	// Connect to printer
	if err := d.connect(ctx); err != nil {
		d.SetHealth(devices.HealthOffline, "failed to connect", err)
		return err
	}

	// Initialize printer
	if err := d.initialize(); err != nil {
		d.SetHealth(devices.HealthDegraded, "failed to initialize", err)
		return err
	}

	d.SetHealth(devices.HealthReady, "printer ready", nil)
	d.logger.Info("ESC/POS printer started successfully")

	return nil
}

// Stop closes the printer connection
func (d *Driver) Stop() error {
	d.logger.Info("stopping ESC/POS printer driver")

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn != nil {
		if err := d.conn.Close(); err != nil {
			d.logger.Error("error closing connection", telemetry.Error(err))
		}
		d.conn = nil
	}

	d.SetHealth(devices.HealthOffline, "stopped", nil)
	return nil
}

// Print sends a document to the printer
func (d *Driver) Print(ctx context.Context, doc *printer.PrintDocument) error {
	d.logger.Info("printing document")

	// Render document to ESC/POS commands
	commands, err := d.renderDocument(doc)
	if err != nil {
		return fmt.Errorf("failed to render document: %w", err)
	}

	// Send to printer
	if err := d.send(ctx, commands); err != nil {
		d.SetHealth(devices.HealthDegraded, "print failed", err)
		return fmt.Errorf("failed to send to printer: %w", err)
	}

	d.logger.Info("document printed successfully")
	return nil
}

// OpenDrawer opens the cash drawer
func (d *Driver) OpenDrawer(ctx context.Context) error {
	d.logger.Info("opening cash drawer")

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
	// For TCP printers, status query is not always reliable
	// Return basic status based on connection health

	d.mu.Lock()
	online := d.conn != nil
	d.mu.Unlock()

	return &printer.PrinterStatus{
		Online:       online,
		PaperPresent: true, // Assume paper is present
		DrawerOpen:   false,
	}, nil
}

// connect establishes connection to the printer
func (d *Driver) connect(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	addr := fmt.Sprintf("%s:%d", d.config.Address, d.config.Port)
	d.logger.Info("connecting to printer", telemetry.String("address", addr))

	dialer := &net.Dialer{
		Timeout: d.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		d.logger.Error("connection failed", telemetry.Error(err))
		return fmt.Errorf("failed to connect: %w", err)
	}

	d.conn = conn
	d.logger.Info("connected to printer")

	return nil
}

// initialize sends initialization commands to printer
func (d *Driver) initialize() error {
	d.logger.Info("initializing printer")

	// ESC @ - Initialize printer
	initCmd := []byte{0x1B, 0x40}

	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn == nil {
		return fmt.Errorf("not connected")
	}

	if err := d.conn.SetWriteDeadline(time.Now().Add(d.config.Timeout)); err != nil {
		return err
	}

	_, err := d.conn.Write(initCmd)
	return err
}

// send sends data to the printer
func (d *Driver) send(ctx context.Context, data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.conn == nil {
		// Try to reconnect
		if err := d.reconnect(ctx); err != nil {
			return fmt.Errorf("not connected and reconnect failed: %w", err)
		}
	}

	// Set write deadline
	if err := d.conn.SetWriteDeadline(time.Now().Add(d.config.Timeout)); err != nil {
		return err
	}

	// Write data
	n, err := d.conn.Write(data)
	if err != nil {
		d.logger.Error("write failed", telemetry.Error(err))
		// Connection lost, close it
		d.conn.Close()
		d.conn = nil
		return err
	}

	if n != len(data) {
		return fmt.Errorf("short write: %d/%d bytes", n, len(data))
	}

	return nil
}

// reconnect attempts to reconnect to the printer
func (d *Driver) reconnect(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", d.config.Address, d.config.Port)
	d.logger.Info("reconnecting to printer", telemetry.String("address", addr))

	dialer := &net.Dialer{
		Timeout: d.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	d.conn = conn

	// Re-initialize
	initCmd := []byte{0x1B, 0x40}
	if err := d.conn.SetWriteDeadline(time.Now().Add(d.config.Timeout)); err != nil {
		return err
	}
	if _, err := d.conn.Write(initCmd); err != nil {
		return err
	}

	d.logger.Info("reconnected to printer")
	return nil
}

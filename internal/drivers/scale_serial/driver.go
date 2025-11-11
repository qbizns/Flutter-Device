package scale_serial

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.bug.st/serial"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/scale"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Driver implements a serial scale driver
type Driver struct {
	*devices.BaseDevice

	logger   *telemetry.Logger
	protocol Protocol
	config   Config

	mu     sync.RWMutex
	port   serial.Port
	opened bool
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDriver creates a new serial scale driver
func NewDriver(id, name string, config Config, logger *telemetry.Logger) (*Driver, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create protocol handler
	protocol, err := createProtocol(config.Protocol)
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol: %w", err)
	}

	d := &Driver{
		BaseDevice: devices.NewBaseDevice(id, "scale.serial", name, nil),
		logger:     logger,
		protocol:   protocol,
		config:     config,
		opened:     false,
	}

	return d, nil
}

// Start starts the scale driver
func (d *Driver) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.opened {
		return fmt.Errorf("driver already started")
	}

	d.ctx, d.cancel = context.WithCancel(ctx)

	// Open serial port
	if err := d.openPort(); err != nil {
		return fmt.Errorf("failed to open port: %w", err)
	}

	d.opened = true
	d.SetHealth(devices.HealthReady, "ready", nil)

	d.logger.Info("serial scale driver started",
		telemetry.String("device_id", d.ID()),
		telemetry.String("port", d.config.Port),
		telemetry.Int("baud_rate", d.config.BaudRate),
		telemetry.String("protocol", d.protocol.Name()),
	)

	return nil
}

// Stop stops the scale driver
func (d *Driver) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.opened {
		return nil
	}

	if d.cancel != nil {
		d.cancel()
	}

	// Close serial port
	if d.port != nil {
		if err := d.port.Close(); err != nil {
			d.logger.Error("failed to close serial port",
				telemetry.String("device_id", d.ID()),
				telemetry.Error(err),
			)
		}
		d.port = nil
	}

	d.opened = false
	d.SetHealth(devices.HealthOffline, "stopped", nil)

	d.logger.Info("serial scale driver stopped",
		telemetry.String("device_id", d.ID()),
	)

	return nil
}

// ReadWeight reads the current weight from the scale
func (d *Driver) ReadWeight(ctx context.Context) (scale.WeightReading, error) {
	d.mu.RLock()
	if !d.opened || d.port == nil {
		d.mu.RUnlock()
		return scale.WeightReading{}, fmt.Errorf("driver not started")
	}
	d.mu.RUnlock()

	// Retry logic
	var lastErr error
	for attempt := 0; attempt <= d.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-ctx.Done():
				return scale.WeightReading{}, ctx.Err()
			case <-time.After(d.config.RetryDelay):
			}

			d.logger.Debug("retrying weight read",
				telemetry.String("device_id", d.ID()),
				telemetry.Int("attempt", attempt+1),
			)
		}

		reading, err := d.readWeightOnce(ctx)
		if err == nil {
			return reading, nil
		}

		lastErr = err
	}

	return scale.WeightReading{}, fmt.Errorf("failed to read weight after %d attempts: %w",
		d.config.RetryAttempts+1, lastErr)
}

// readWeightOnce performs a single weight read
func (d *Driver) readWeightOnce(ctx context.Context) (scale.WeightReading, error) {
	// Send weight command if protocol requires it
	if d.protocol.RequiresCommand() {
		cmd, err := d.protocol.SendWeight()
		if err != nil {
			return scale.WeightReading{}, fmt.Errorf("failed to generate command: %w", err)
		}

		d.mu.Lock()
		_, err = d.port.Write(cmd)
		d.mu.Unlock()

		if err != nil {
			return scale.WeightReading{}, fmt.Errorf("failed to write command: %w", err)
		}
	}

	// Read response
	response, err := d.readResponse(ctx)
	if err != nil {
		return scale.WeightReading{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	reading, status, err := d.protocol.ParseResponse(response)
	if err != nil {
		return scale.WeightReading{}, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check status
	if status != StatusStable && status != StatusUnstable {
		return scale.WeightReading{}, fmt.Errorf("scale status error: %s", status.String())
	}

	// Convert to scale.WeightReading
	result := scale.WeightReading{
		Weight:    reading.Value,
		Unit:      reading.Unit.ToScaleUnit(),
		Stable:    reading.Stable,
		Timestamp: reading.Timestamp,
	}

	// Convert to preferred unit if configured
	if d.config.PreferredUnit != reading.Unit {
		result.Weight = ConvertWeight(reading.Value, reading.Unit, d.config.PreferredUnit)
		result.Unit = d.config.PreferredUnit.ToScaleUnit()
	}

	return result, nil
}

// Zero sets the current reading as zero point
func (d *Driver) Zero(ctx context.Context) error {
	d.mu.RLock()
	if !d.opened || d.port == nil {
		d.mu.RUnlock()
		return fmt.Errorf("driver not started")
	}
	d.mu.RUnlock()

	// Send zero command
	cmd, err := d.protocol.SendZero()
	if err != nil {
		return fmt.Errorf("failed to generate zero command: %w", err)
	}

	d.mu.Lock()
	_, err = d.port.Write(cmd)
	d.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to write zero command: %w", err)
	}

	// Read response to verify
	response, err := d.readResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to read zero response: %w", err)
	}

	d.logger.Debug("zero response",
		telemetry.String("device_id", d.ID()),
		telemetry.String("response", string(response)),
	)

	return nil
}

// Tare subtracts container weight
func (d *Driver) Tare(ctx context.Context) error {
	d.mu.RLock()
	if !d.opened || d.port == nil {
		d.mu.RUnlock()
		return fmt.Errorf("driver not started")
	}
	d.mu.RUnlock()

	// Send tare command
	cmd, err := d.protocol.SendTare()
	if err != nil {
		return fmt.Errorf("failed to generate tare command: %w", err)
	}

	d.mu.Lock()
	_, err = d.port.Write(cmd)
	d.mu.Unlock()

	if err != nil {
		return fmt.Errorf("failed to write tare command: %w", err)
	}

	// Read response to verify
	response, err := d.readResponse(ctx)
	if err != nil {
		return fmt.Errorf("failed to read tare response: %w", err)
	}

	d.logger.Debug("tare response",
		telemetry.String("device_id", d.ID()),
		telemetry.String("response", string(response)),
	)

	return nil
}

// openPort opens the serial port
func (d *Driver) openPort() error {
	mode := &serial.Mode{
		BaudRate: d.config.BaudRate,
		DataBits: d.config.DataBits,
	}

	// Set parity
	switch d.config.Parity {
	case "none":
		mode.Parity = serial.NoParity
	case "even":
		mode.Parity = serial.EvenParity
	case "odd":
		mode.Parity = serial.OddParity
	default:
		return fmt.Errorf("invalid parity: %s", d.config.Parity)
	}

	// Set stop bits
	switch d.config.StopBits {
	case 1:
		mode.StopBits = serial.OneStopBit
	case 2:
		mode.StopBits = serial.TwoStopBits
	default:
		return fmt.Errorf("invalid stop bits: %d", d.config.StopBits)
	}

	// Open port
	port, err := serial.Open(d.config.Port, mode)
	if err != nil {
		return fmt.Errorf("failed to open serial port %s: %w", d.config.Port, err)
	}

	// Set read timeout
	if err := port.SetReadTimeout(d.config.ReadTimeout); err != nil {
		port.Close()
		return fmt.Errorf("failed to set read timeout: %w", err)
	}

	d.port = port

	return nil
}

// readResponse reads a complete response from the serial port
func (d *Driver) readResponse(ctx context.Context) ([]byte, error) {
	terminator := d.protocol.LineTerminator()
	buf := make([]byte, 0, 256)
	readBuf := make([]byte, 1)

	deadline := time.Now().Add(d.config.ReadTimeout)

	for {
		// Check context
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Check timeout
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("read timeout")
		}

		// Read one byte
		d.mu.Lock()
		n, err := d.port.Read(readBuf)
		d.mu.Unlock()

		if err != nil {
			return nil, fmt.Errorf("read error: %w", err)
		}

		if n > 0 {
			buf = append(buf, readBuf[0])

			// Check for terminator
			if len(buf) >= len(terminator) {
				if bytesEqual(buf[len(buf)-len(terminator):], terminator) {
					return buf, nil
				}
			}

			// Prevent infinite buffer growth
			if len(buf) > 1024 {
				return nil, fmt.Errorf("response too large")
			}
		}
	}
}

// bytesEqual checks if two byte slices are equal
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// createProtocol creates a protocol handler based on name
func createProtocol(name string) (Protocol, error) {
	switch name {
	case "mtsics", "mt-sics", "mettler":
		return NewMTSICSProtocol(), nil
	default:
		return nil, fmt.Errorf("unknown protocol: %s", name)
	}
}

// EnumeratePorts returns a list of available serial ports
func EnumeratePorts() ([]string, error) {
	return serial.GetPortsList()
}

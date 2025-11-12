package scale_serial

import (
	"bytes"
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
	case "cas":
		return NewCASProtocol(), nil
	case "dibal":
		return NewDibalProtocol(), nil
	case "toledo", "toledo8217", "toledo-8217":
		return NewToledoProtocol(), nil
	case "generic":
		return NewGenericProtocol(), nil
	case "auto":
		// Auto-detection will be handled separately
		return nil, fmt.Errorf("auto-detection requires reading from scale first")
	default:
		return nil, fmt.Errorf("unknown protocol: %s (supported: mtsics, cas, dibal, toledo, generic, auto)", name)
	}
}

// DetectProtocol attempts to auto-detect the scale protocol from sample data
//
// This function analyzes the response format to determine which protocol the scale is using.
// It returns the detected protocol and a confidence score (0.0-1.0).
//
// Detection Strategy:
//   1. Check for MT-SICS format (command-response, specific status codes)
//   2. Check for CAS format (STX/ETX with specific status bytes)
//   3. Check for Dibal format (STX/ETX with +/- status)
//   4. Check for Toledo format (continuous, specific status codes)
//   5. Fall back to generic if nothing matches
func DetectProtocol(data []byte) (Protocol, float64, error) {
	if len(data) == 0 {
		return nil, 0.0, fmt.Errorf("no data provided for protocol detection")
	}

	// Score for each protocol (higher = more confident)
	scores := make(map[string]float64)

	// Check for MT-SICS format
	// Characteristics: CR/LF terminator, status codes (S, D, +, -, I, L), space-separated fields
	if bytes.Contains(data, []byte("\r\n")) {
		scores["mtsics"] += 0.3

		// Check for MT-SICS specific status codes
		if len(data) > 0 && (data[0] == 'S' || data[0] == 'D' || data[0] == 'I' || data[0] == 'L') {
			scores["mtsics"] += 0.3

			// Check for MT-SICS weight format (spaces between fields)
			if bytes.Count(data, []byte(" ")) >= 2 {
				scores["mtsics"] += 0.2
			}
		}

		// Also check for Toledo (similar but different)
		if len(data) > 0 && (data[0] == 'S' || data[0] == 'D' || data[0] == 'M' || data[0] == '?') {
			scores["toledo"] += 0.4

			// Toledo often has fixed-width format or specific spacing
			if bytes.Count(data, []byte("  ")) >= 1 {
				scores["toledo"] += 0.2
			}
		}
	}

	// Check for CAS format
	// Characteristics: STX at start, ETX at end, status bytes (S, U, E)
	if len(data) >= 3 && data[0] == 0x02 {
		scores["cas"] += 0.4

		// Find ETX
		if bytes.Contains(data, []byte{0x03}) {
			scores["cas"] += 0.3

			// Check for CAS status bytes
			if len(data) > 1 && (data[1] == 'S' || data[1] == 'U' || data[1] == 'E') {
				scores["cas"] += 0.2
			}
		}
	}

	// Check for Dibal format
	// Characteristics: STX at start, ETX at end, status bytes (+, -, O, U, E)
	if len(data) >= 3 && data[0] == 0x02 {
		scores["dibal"] += 0.3

		// Find ETX
		if bytes.Contains(data, []byte{0x03}) {
			scores["dibal"] += 0.2

			// Check for Dibal status bytes
			if len(data) > 1 && (data[1] == '+' || data[1] == '-' || data[1] == 'O' || data[1] == 'U' || data[1] == 'E') {
				scores["dibal"] += 0.4
			}
		}
	}

	// Find the highest scoring protocol
	var bestProtocol string
	var bestScore float64
	for protocol, score := range scores {
		if score > bestScore {
			bestScore = score
			bestProtocol = protocol
		}
	}

	// If no clear winner, use generic
	if bestScore < 0.5 {
		bestProtocol = "generic"
		bestScore = 0.3
	}

	// Create protocol instance
	proto, err := createProtocol(bestProtocol)
	if err != nil {
		return NewGenericProtocol(), 0.3, err
	}

	return proto, bestScore, nil
}

// AutoDetectProtocol attempts to auto-detect the scale protocol by reading sample data
//
// This function opens the serial port temporarily, reads some data, and attempts to
// detect the protocol. It's useful for "auto" protocol configuration.
//
// Parameters:
//   - config: Serial port configuration (port, baud rate, etc.)
//   - timeout: How long to wait for data (e.g., 3 seconds)
//   - samples: Number of samples to collect for detection (e.g., 3-5)
//
// Returns the detected protocol, confidence score, and any error.
func AutoDetectProtocol(config Config, timeout time.Duration, samples int) (Protocol, float64, error) {
	// Open serial port temporarily
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: config.DataBits,
		Parity:   parseParity(config.Parity),
		StopBits: serial.StopBits(config.StopBits),
	}

	port, err := serial.Open(config.Port, mode)
	if err != nil {
		return nil, 0.0, fmt.Errorf("failed to open port for detection: %w", err)
	}
	defer port.Close()

	// Set read timeout
	if err := port.SetReadTimeout(timeout); err != nil {
		return nil, 0.0, fmt.Errorf("failed to set read timeout: %w", err)
	}

	// Collect multiple samples
	var allData []byte
	var bestProtocol Protocol
	var bestScore float64

	for i := 0; i < samples; i++ {
		// Read some data
		buf := make([]byte, 256)
		n, err := port.Read(buf)
		if err != nil && err.Error() != "EOF" {
			// If we already have some samples, continue
			if i > 0 {
				break
			}
			return nil, 0.0, fmt.Errorf("failed to read from port: %w", err)
		}

		if n > 0 {
			sample := buf[:n]
			allData = append(allData, sample...)

			// Try to detect protocol from this sample
			protocol, score, _ := DetectProtocol(sample)
			if score > bestScore {
				bestProtocol = protocol
				bestScore = score
			}
		}

		// Small delay between samples
		time.Sleep(100 * time.Millisecond)
	}

	if bestProtocol == nil {
		// Try detection on all combined data
		protocol, score, err := DetectProtocol(allData)
		if err != nil {
			return nil, 0.0, fmt.Errorf("protocol detection failed: %w", err)
		}
		return protocol, score, nil
	}

	return bestProtocol, bestScore, nil
}

// parseParity converts a string parity value to serial.Parity
func parseParity(parity string) serial.Parity {
	switch parity {
	case "none":
		return serial.NoParity
	case "even":
		return serial.EvenParity
	case "odd":
		return serial.OddParity
	default:
		return serial.NoParity
	}
}

// EnumeratePorts returns a list of available serial ports
func EnumeratePorts() ([]string, error) {
	return serial.GetPortsList()
}

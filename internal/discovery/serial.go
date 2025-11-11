package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"go.bug.st/serial"
)

// SerialScanner scans for serial port devices (scales, printers)
type SerialScanner struct {
	logger           *telemetry.Logger
	baudRates        []int
	probeTimeout     time.Duration
	discoverScales   bool
	discoverPrinters bool
}

// SerialScannerConfig configures serial discovery
type SerialScannerConfig struct {
	BaudRates        []int  // Baud rates to try (default: 9600, 19200)
	ProbeTimeout     time.Duration
	DiscoverScales   bool
	DiscoverPrinters bool
}

// NewSerialScanner creates a new serial port scanner
func NewSerialScanner(cfg SerialScannerConfig, logger *telemetry.Logger) *SerialScanner {
	// Default baud rates (most common for scales)
	baudRates := cfg.BaudRates
	if len(baudRates) == 0 {
		baudRates = []int{9600, 19200, 4800, 38400}
	}

	// Default probe timeout
	probeTimeout := cfg.ProbeTimeout
	if probeTimeout == 0 {
		probeTimeout = 2 * time.Second
	}

	return &SerialScanner{
		logger:           logger,
		baudRates:        baudRates,
		probeTimeout:     probeTimeout,
		discoverScales:   cfg.DiscoverScales,
		discoverPrinters: cfg.DiscoverPrinters,
	}
}

// Scan scans for serial port devices
func (s *SerialScanner) Scan(ctx context.Context) ([]DiscoveredDevice, error) {
	// Enumerate all serial ports
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate serial ports: %w", err)
	}

	if len(ports) == 0 {
		s.logger.Debug("no serial ports found")
		return []DiscoveredDevice{}, nil
	}

	s.logger.Debug("enumerating serial ports",
		telemetry.Int("count", len(ports)),
	)

	var devices []DiscoveredDevice

	// Probe each port
	for _, portName := range ports {
		select {
		case <-ctx.Done():
			return devices, ctx.Err()
		default:
		}

		// Skip Bluetooth serial ports (virtual ports)
		if isBluetoothPort(portName) {
			s.logger.Debug("skipping bluetooth port", telemetry.String("port", portName))
			continue
		}

		// Probe the port
		device, err := s.probePort(ctx, portName)
		if err != nil {
			s.logger.Debug("port probe failed",
				telemetry.String("port", portName),
				telemetry.Error(err),
			)
			continue
		}

		if device != nil {
			devices = append(devices, *device)
			s.logger.Info("serial device discovered",
				telemetry.String("port", portName),
				telemetry.String("kind", device.Kind),
			)
		}
	}

	s.logger.Debug("serial scan completed",
		telemetry.Int("devices", len(devices)),
	)

	return devices, nil
}

// probePort probes a serial port to identify the device
func (s *SerialScanner) probePort(ctx context.Context, portName string) (*DiscoveredDevice, error) {
	// Only probe for scales at the moment
	if !s.discoverScales {
		return nil, nil
	}

	// Try different baud rates
	for _, baudRate := range s.baudRates {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Try to detect a scale at this baud rate
		device, err := s.probeScale(ctx, portName, baudRate)
		if err != nil {
			continue
		}

		if device != nil {
			return device, nil
		}
	}

	return nil, fmt.Errorf("no device detected")
}

// probeScale probes a serial port for scale devices
func (s *SerialScanner) probeScale(ctx context.Context, portName string, baudRate int) (*DiscoveredDevice, error) {
	// Open serial port
	mode := &serial.Mode{
		BaudRate: baudRate,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(portName, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open port: %w", err)
	}
	defer port.Close()

	// Set read timeout
	port.SetReadTimeout(s.probeTimeout)

	// Try MT-SICS protocol (Mettler Toledo)
	protocol, err := s.tryMTSICS(port)
	if err == nil && protocol != "" {
		return &DiscoveredDevice{
			ID:        fmt.Sprintf("serial-%s", sanitizePortName(portName)),
			Name:      fmt.Sprintf("Mettler Toledo Scale on %s", portName),
			Kind:      "scale.serial",
			Transport: "serial",
			Address:   portName,
			Metadata: map[string]string{
				"port":           portName,
				"baud_rate":      fmt.Sprintf("%d", baudRate),
				"protocol":       protocol,
				"discovered_via": "serial_probe",
			},
		}, nil
	}

	// Try CAS protocol
	protocol, err = s.tryCAS(port)
	if err == nil && protocol != "" {
		return &DiscoveredDevice{
			ID:        fmt.Sprintf("serial-%s", sanitizePortName(portName)),
			Name:      fmt.Sprintf("CAS Scale on %s", portName),
			Kind:      "scale.serial",
			Transport: "serial",
			Address:   portName,
			Metadata: map[string]string{
				"port":           portName,
				"baud_rate":      fmt.Sprintf("%d", baudRate),
				"protocol":       protocol,
				"discovered_via": "serial_probe",
			},
		}, nil
	}

	// Try Dibal protocol
	protocol, err = s.tryDibal(port)
	if err == nil && protocol != "" {
		return &DiscoveredDevice{
			ID:        fmt.Sprintf("serial-%s", sanitizePortName(portName)),
			Name:      fmt.Sprintf("Dibal Scale on %s", portName),
			Kind:      "scale.serial",
			Transport: "serial",
			Address:   portName,
			Metadata: map[string]string{
				"port":           portName,
				"baud_rate":      fmt.Sprintf("%d", baudRate),
				"protocol":       protocol,
				"discovered_via": "serial_probe",
			},
		}, nil
	}

	// Try Toledo protocol
	protocol, err = s.tryToledo(port)
	if err == nil && protocol != "" {
		return &DiscoveredDevice{
			ID:        fmt.Sprintf("serial-%s", sanitizePortName(portName)),
			Name:      fmt.Sprintf("Toledo Scale on %s", portName),
			Kind:      "scale.serial",
			Transport: "serial",
			Address:   portName,
			Metadata: map[string]string{
				"port":           portName,
				"baud_rate":      fmt.Sprintf("%d", baudRate),
				"protocol":       protocol,
				"discovered_via": "serial_probe",
			},
		}, nil
	}

	return nil, fmt.Errorf("no scale protocol detected")
}

// tryMTSICS tries to detect MT-SICS protocol (Mettler Toledo)
func (s *SerialScanner) tryMTSICS(port serial.Port) (string, error) {
	// Send immediate weight request: SI\r\n
	_, err := port.Write([]byte("SI\r\n"))
	if err != nil {
		return "", err
	}

	// Read response
	buf := make([]byte, 128)
	n, err := port.Read(buf)
	if err != nil {
		return "", err
	}

	response := string(buf[:n])

	// Check for MT-SICS response format
	// Expected: "S       1234.5 g\r\n" or "D       1234.5 g\r\n"
	if len(response) > 2 && (response[0] == 'S' || response[0] == 'D' || response[0] == '+' || response[0] == '-') {
		// Look for units (g, kg, lb, oz)
		if contains(response, "g") || contains(response, "kg") || contains(response, "lb") || contains(response, "oz") {
			return "mtsics", nil
		}
	}

	return "", fmt.Errorf("not MT-SICS")
}

// tryCAS tries to detect CAS protocol
func (s *SerialScanner) tryCAS(port serial.Port) (string, error) {
	// Send weight request: STX W ETX
	_, err := port.Write([]byte{0x02, 'W', 0x03})
	if err != nil {
		return "", err
	}

	// Read response
	buf := make([]byte, 128)
	n, err := port.Read(buf)
	if err != nil {
		return "", err
	}

	// Check for CAS response format
	// Expected: STX <status> <6-digit weight> <unit> ETX
	if n > 4 && buf[0] == 0x02 && buf[n-1] == 0x03 {
		// Status byte should be 'S', 'U', or 'E'
		if buf[1] == 'S' || buf[1] == 'U' || buf[1] == 'E' {
			return "cas", nil
		}
	}

	return "", fmt.Errorf("not CAS")
}

// tryDibal tries to detect Dibal protocol
func (s *SerialScanner) tryDibal(port serial.Port) (string, error) {
	// Dibal scales send continuous output, so just read
	buf := make([]byte, 128)
	n, err := port.Read(buf)
	if err != nil {
		return "", err
	}

	// Check for Dibal format
	// Expected: STX <status +/-> <weight> <unit> ETX
	if n > 4 && buf[0] == 0x02 {
		// Status should be '+' or '-'
		if buf[1] == '+' || buf[1] == '-' {
			// Look for ETX
			for i := 0; i < n; i++ {
				if buf[i] == 0x03 {
					return "dibal", nil
				}
			}
		}
	}

	return "", fmt.Errorf("not Dibal")
}

// tryToledo tries to detect Toledo 8217 protocol
func (s *SerialScanner) tryToledo(port serial.Port) (string, error) {
	// Toledo sends continuous ASCII output, just read
	buf := make([]byte, 128)
	n, err := port.Read(buf)
	if err != nil {
		return "", err
	}

	response := string(buf[:n])

	// Check for Toledo format
	// Expected: <status> <weight> <unit> CR LF
	// Status: S (stable), D (dynamic), E (error), etc.
	if n > 4 {
		status := response[0]
		if status == 'S' || status == 's' || status == 'D' || status == 'd' ||
			status == 'M' || status == 'm' || status == 'E' || status == 'e' {
			// Look for CR or LF
			if contains(response, "\r") || contains(response, "\n") {
				return "toledo", nil
			}
		}
	}

	return "", fmt.Errorf("not Toledo")
}

// Transport returns the transport name
func (s *SerialScanner) Transport() string {
	return "serial"
}

// Helper functions

func isBluetoothPort(portName string) bool {
	// Skip Bluetooth serial ports (macOS/Linux)
	return contains(portName, "Bluetooth") || contains(portName, "bluetooth")
}

func sanitizePortName(portName string) string {
	// Replace / and \ with - for use in IDs
	result := ""
	for _, ch := range portName {
		if ch == '/' || ch == '\\' || ch == ':' {
			result += "-"
		} else {
			result += string(ch)
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

package scale_serial

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ToledoProtocol implements the Toledo 8217 serial protocol
//
// Toledo (now part of Mettler Toledo) 8217 is a legacy continuous output protocol
// used in older Toledo scales. It runs in continuous mode, sending weight automatically.
//
// Protocol Characteristics:
//   - Continuous ASCII output (no commands required)
//   - Fixed-width fields
//   - Status indicators
//   - Supports multiple units
//   - CR/LF line termination
type ToledoProtocol struct{}

// NewToledoProtocol creates a new Toledo protocol handler
func NewToledoProtocol() *ToledoProtocol {
	return &ToledoProtocol{}
}

// Name returns the protocol name
func (p *ToledoProtocol) Name() string {
	return "Toledo-8217"
}

// SendWeight sends a command to request the current weight
//
// Note: Toledo 8217 typically runs in continuous mode and doesn't accept commands.
// This method is provided for compatibility but may not work on all scales.
func (p *ToledoProtocol) SendWeight() ([]byte, error) {
	// Toledo 8217 continuous mode doesn't accept commands
	// Return empty command (scale will continue sending automatically)
	return []byte{}, nil
}

// SendStableWeight sends a command to request stable weight only
func (p *ToledoProtocol) SendStableWeight() ([]byte, error) {
	return p.SendWeight()
}

// SendZero sends a command to zero the scale
func (p *ToledoProtocol) SendZero() ([]byte, error) {
	// Toledo 8217 typically doesn't support remote zero via serial
	// Some models may support "Z\r\n"
	return []byte("Z\r\n"), nil
}

// SendTare sends a command to tare the scale
func (p *ToledoProtocol) SendTare() ([]byte, error) {
	// Toledo 8217 typically doesn't support remote tare via serial
	// Some models may support "T\r\n"
	return []byte("T\r\n"), nil
}

// SendReset sends a command to reset the scale
func (p *ToledoProtocol) SendReset() ([]byte, error) {
	// No standard reset command for Toledo 8217
	return []byte{}, nil
}

// RequiresCommand returns false since Toledo 8217 runs in continuous mode
func (p *ToledoProtocol) RequiresCommand() bool {
	return false
}

// LineTerminator returns the line terminator for Toledo 8217
func (p *ToledoProtocol) LineTerminator() []byte {
	return []byte("\r\n")
}

// ParseResponse parses a response from the scale
//
// Toledo 8217 Response Format (Continuous Mode):
//   <Status><Space><Weight><Space><Unit><CR><LF>
//
// Status indicators (1 character):
//   S  = Stable (steady) weight
//   D  = Dynamic (unstable) weight
//   M  = Motion detected
//   +  = Overload (over capacity)
//   -  = Underload (under zero)
//   ?  = Error or invalid reading
//
// Weight: Variable length, typically 6-9 characters including decimal point and sign
//   - May include leading spaces
//   - Decimal point position varies by scale capacity
//   - Sign (+/-) may be present
//
// Unit: 2-4 characters (kg, g, lb, oz, t)
//
// Examples:
//   "S    123.45 kg\r\n"     # Stable, 123.45 kg
//   "D     50.2 lb\r\n"      # Unstable, 50.2 lb
//   "S   -001.23 kg\r\n"     # Stable, -1.23 kg (tare)
//   "+   9999.99 g\r\n"      # Overload
//   "?      0.00 kg\r\n"     # Error
//
// Toledo 8217 Variant Formats:
//
// Format 1 (Fixed width, 16 bytes):
//   "S  00001234 kg\r\n"     # Status + 2 spaces + 8-digit weight + space + unit
//
// Format 2 (Variable width):
//   "S 123.45 kg\r\n"        # Status + space + weight + space + unit
//
// Format 3 (With additional flags):
//   "S M 123.45 kg\r\n"      # Status + motion flag + weight + unit
func (p *ToledoProtocol) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	// Trim whitespace and line terminators
	response := strings.TrimSpace(string(data))

	// Empty response
	if len(response) == 0 {
		return nil, StatusError, fmt.Errorf("empty response")
	}

	// Minimum length check: Status + space + weight + space + unit
	// Example: "S 1 kg" = 6 characters minimum
	if len(response) < 6 {
		return nil, StatusError, fmt.Errorf("response too short: %s", response)
	}

	// Extract status byte (first character)
	statusByte := response[0]
	var status WeightStatus
	var stable bool

	switch statusByte {
	case 'S', 's':
		status = StatusStable
		stable = true
	case 'D', 'd', 'M', 'm':
		status = StatusUnstable
		stable = false
	case '+':
		return nil, StatusOverload, fmt.Errorf("scale overload")
	case '-':
		// Check if this is a negative weight or underload
		// If followed by space and digits, it's underload
		// If followed by digits, it's a negative weight
		if len(response) > 1 && response[1] == ' ' {
			// Check if rest looks like zero or very small number
			remaining := strings.TrimSpace(response[1:])
			if strings.HasPrefix(remaining, "0") || strings.HasPrefix(remaining, "-0") {
				return nil, StatusUnderload, fmt.Errorf("scale underload")
			}
		}
		// Treat as stable with negative weight
		status = StatusStable
		stable = true
	case '?':
		return nil, StatusError, fmt.Errorf("scale error")
	default:
		return nil, StatusError, fmt.Errorf("unknown status byte: %c", statusByte)
	}

	// Extract weight and unit from rest of response
	// Format: <Status><Space><Weight><Space><Unit>
	// or: <Status><Space><MotionFlag><Space><Weight><Space><Unit>
	rest := strings.TrimSpace(response[1:])

	if len(rest) == 0 {
		return nil, StatusError, fmt.Errorf("no weight data after status")
	}

	// Check for motion flag (M) after status
	if rest[0] == 'M' || rest[0] == 'm' {
		// Motion detected, but continue parsing
		rest = strings.TrimSpace(rest[1:])
		status = StatusUnstable
		stable = false
	}

	// Split remaining data into weight and unit
	parts := strings.Fields(rest)
	if len(parts) < 2 {
		return nil, StatusError, fmt.Errorf("invalid format: expected weight and unit, got: %s", rest)
	}

	weightStr := parts[0]
	unitStr := parts[1]

	// Parse weight value (may include sign and decimal point)
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse weight '%s': %w", weightStr, err)
	}

	// Parse unit
	unit, err := ParseWeightUnit(unitStr)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse unit '%s': %w", unitStr, err)
	}

	reading := &WeightReading{
		Value:     weight,
		Unit:      unit,
		Stable:    stable,
		Timestamp: time.Now(),
		Raw:       response,
	}

	return reading, status, nil
}

// ParseFixedWidth parses a fixed-width Toledo 8217 response
//
// Some Toledo models use a fixed-width format:
// Format: <Status><2 spaces><8-digit weight><space><unit><CR><LF>
// Example: "S  00001234 kg\r\n"
//
// The 8-digit weight has an implied decimal point position that depends
// on the scale's capacity and resolution.
func (p *ToledoProtocol) ParseFixedWidth(data []byte, decimalPlaces int) (*WeightReading, WeightStatus, error) {
	// Trim whitespace and line terminators
	response := strings.TrimSpace(string(data))

	// Fixed width format is typically 16 bytes: "S  00001234 kg\r\n"
	// After trimming: "S  00001234 kg" = 14 characters
	if len(response) < 12 {
		return nil, StatusError, fmt.Errorf("response too short for fixed-width format: %s", response)
	}

	// Extract status byte
	statusByte := response[0]
	var status WeightStatus
	var stable bool

	switch statusByte {
	case 'S', 's':
		status = StatusStable
		stable = true
	case 'D', 'd', 'M', 'm':
		status = StatusUnstable
		stable = false
	case '+':
		return nil, StatusOverload, fmt.Errorf("scale overload")
	case '-':
		return nil, StatusUnderload, fmt.Errorf("scale underload")
	case '?':
		return nil, StatusError, fmt.Errorf("scale error")
	default:
		return nil, StatusError, fmt.Errorf("unknown status byte: %c", statusByte)
	}

	// Skip 2 spaces and extract 8-digit weight
	if len(response) < 11 {
		return nil, StatusError, fmt.Errorf("response too short for weight extraction")
	}

	// Extract 8-digit weight field (positions 3-10)
	weightStr := strings.TrimSpace(response[3:11])

	// Parse as integer
	weightInt, err := strconv.ParseInt(weightStr, 10, 64)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse weight '%s': %w", weightStr, err)
	}

	// Apply decimal places
	divisor := 1.0
	for i := 0; i < decimalPlaces; i++ {
		divisor *= 10.0
	}
	weight := float64(weightInt) / divisor

	// Extract unit (after weight field)
	unitStr := strings.TrimSpace(response[11:])
	unit, err := ParseWeightUnit(unitStr)
	if err != nil {
		return nil, StatusError, fmt.Errorf("failed to parse unit '%s': %w", unitStr, err)
	}

	reading := &WeightReading{
		Value:     weight,
		Unit:      unit,
		Stable:    stable,
		Timestamp: time.Now(),
		Raw:       response,
	}

	return reading, status, nil
}

// ToledoProtocolConfig holds configuration for Toledo protocol variants
type ToledoProtocolConfig struct {
	FixedWidth    bool // Use fixed-width parsing
	DecimalPlaces int  // Number of decimal places for fixed-width format (0, 1, 2, or 3)
	Continuous    bool // Continuous output mode (default)
}

// NewToledoProtocolWithConfig creates a Toledo protocol with custom configuration
func NewToledoProtocolWithConfig(config ToledoProtocolConfig) *ToledoProtocolConfigurable {
	return &ToledoProtocolConfigurable{
		ToledoProtocol: &ToledoProtocol{},
		fixedWidth:     config.FixedWidth,
		decimalPlaces:  config.DecimalPlaces,
		continuous:     config.Continuous,
	}
}

// ToledoProtocolConfigurable extends ToledoProtocol with configurable options
type ToledoProtocolConfigurable struct {
	*ToledoProtocol
	fixedWidth    bool
	decimalPlaces int
	continuous    bool
}

// RequiresCommand returns the configured mode
func (p *ToledoProtocolConfigurable) RequiresCommand() bool {
	return !p.continuous
}

// ParseResponse overrides the base implementation with configurable options
func (p *ToledoProtocolConfigurable) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	if p.fixedWidth {
		return p.ParseFixedWidth(data, p.decimalPlaces)
	}
	return p.ToledoProtocol.ParseResponse(data)
}

// ValidateResponse checks if a response looks like valid Toledo format
func (p *ToledoProtocol) ValidateResponse(data []byte) bool {
	// Check minimum length
	if len(data) < 6 {
		return false
	}

	// Check for CR/LF terminator
	if !bytes.HasSuffix(data, []byte("\r\n")) {
		return false
	}

	// Check status byte
	statusByte := data[0]
	validStatus := statusByte == 'S' || statusByte == 's' ||
		statusByte == 'D' || statusByte == 'd' ||
		statusByte == 'M' || statusByte == 'm' ||
		statusByte == '+' || statusByte == '-' ||
		statusByte == '?'

	return validStatus
}

// IsStableResponse checks if a response indicates stable weight
func (p *ToledoProtocol) IsStableResponse(data []byte) bool {
	if len(data) < 1 {
		return false
	}
	return data[0] == 'S' || data[0] == 's'
}

// IsErrorResponse checks if a response indicates an error
func (p *ToledoProtocol) IsErrorResponse(data []byte) bool {
	if len(data) < 1 {
		return true
	}
	return data[0] == '?' || data[0] == '+' || (data[0] == '-' && len(data) > 1 && data[1] == ' ')
}

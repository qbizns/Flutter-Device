package scale_serial

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// MTSICSProtocol implements the Mettler Toledo Standard Interface Command Set (MT-SICS) protocol
type MTSICSProtocol struct{}

// NewMTSICSProtocol creates a new MT-SICS protocol handler
func NewMTSICSProtocol() *MTSICSProtocol {
	return &MTSICSProtocol{}
}

// Name returns the protocol name
func (p *MTSICSProtocol) Name() string {
	return "MT-SICS"
}

// SendWeight sends a command to request the current weight immediately (stable or unstable)
func (p *MTSICSProtocol) SendWeight() ([]byte, error) {
	// SI = Send weight immediately (stable or unstable)
	return []byte("SI\r\n"), nil
}

// SendStableWeight sends a command to request stable weight only
func (p *MTSICSProtocol) SendStableWeight() ([]byte, error) {
	// S = Send stable weight
	return []byte("S\r\n"), nil
}

// SendZero sends a command to zero the scale
func (p *MTSICSProtocol) SendZero() ([]byte, error) {
	// ZI = Zero immediately
	return []byte("ZI\r\n"), nil
}

// SendTare sends a command to tare the scale
func (p *MTSICSProtocol) SendTare() ([]byte, error) {
	// T = Tare
	return []byte("T\r\n"), nil
}

// SendReset sends a command to reset the scale
func (p *MTSICSProtocol) SendReset() ([]byte, error) {
	// @ = Reset
	return []byte("@\r\n"), nil
}

// RequiresCommand returns true since MT-SICS is command-based
func (p *MTSICSProtocol) RequiresCommand() bool {
	return true
}

// LineTerminator returns the line terminator for MT-SICS
func (p *MTSICSProtocol) LineTerminator() []byte {
	return []byte("\r\n")
}

// ParseResponse parses a response from the scale
//
// MT-SICS Response Format:
//   <Status><Space><Weight><Space><Unit><CR><LF>
//
// Status codes:
//   S  = Stable weight
//   D  = Dynamic (unstable) weight
//   +  = Overload
//   -  = Underload
//   I  = Invalid command
//   L  = Command not executable
//
// Examples:
//   "S       1234.5 g\r\n"
//   "D        500.2 kg\r\n"
//   "S     -00012.3 lb\r\n"
//   "+       9999.9 g\r\n"
//   "I\r\n"
func (p *MTSICSProtocol) ParseResponse(data []byte) (*WeightReading, WeightStatus, error) {
	// Trim whitespace and line terminators
	response := strings.TrimSpace(string(data))

	// Empty response
	if len(response) == 0 {
		return nil, StatusError, fmt.Errorf("empty response")
	}

	// Check for error responses
	if response == "I" {
		return nil, StatusError, fmt.Errorf("invalid command")
	}
	if response == "L" {
		return nil, StatusError, fmt.Errorf("command not executable")
	}

	// Extract status byte
	if len(response) < 1 {
		return nil, StatusError, fmt.Errorf("response too short")
	}
	statusByte := response[0]

	var status WeightStatus
	var stable bool

	switch statusByte {
	case 'S':
		status = StatusStable
		stable = true
	case 'D':
		status = StatusUnstable
		stable = false
	case '+':
		return nil, StatusOverload, fmt.Errorf("scale overload")
	case '-':
		return nil, StatusUnderload, fmt.Errorf("scale underload")
	default:
		return nil, StatusError, fmt.Errorf("unknown status byte: %c", statusByte)
	}

	// Parse weight and unit
	// Format after status: "<Space><Weight><Space><Unit>"
	// Example: "S       1234.5 g"
	parts := strings.Fields(response[1:]) // Skip status byte and split by whitespace

	if len(parts) < 2 {
		return nil, StatusError, fmt.Errorf("invalid response format: expected weight and unit")
	}

	weightStr := parts[0]
	unitStr := parts[1]

	// Parse weight value
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

// SendContinuousMode sends a command to start continuous weight output
func (p *MTSICSProtocol) SendContinuousMode() ([]byte, error) {
	// SIR = Send weight immediately and repeatedly
	return []byte("SIR\r\n"), nil
}

// SendStopContinuous sends a command to stop continuous output
func (p *MTSICSProtocol) SendStopContinuous() ([]byte, error) {
	// Any command stops continuous mode, we use SI (send weight immediately)
	return []byte("SI\r\n"), nil
}

// SendGetInfo sends a command to get scale information
func (p *MTSICSProtocol) SendGetInfo() ([]byte, error) {
	// I0 = Get scale information
	return []byte("I0\r\n"), nil
}

// SendClearTare sends a command to clear the tare weight
func (p *MTSICSProtocol) SendClearTare() ([]byte, error) {
	// TAC = Tare clear
	return []byte("TAC\r\n"), nil
}

// ParseInfoResponse parses scale information response
//
// Response format:
//   I0 A "Scale_Type" "SNR" "SW_Version"<CR><LF>
//
// Example:
//   "I0 A \"PS60\" \"1234567\" \"V1.0\"\r\n"
func (p *MTSICSProtocol) ParseInfoResponse(data []byte) (scaleType, serialNumber, version string, err error) {
	response := strings.TrimSpace(string(data))

	// Check for error
	if strings.HasPrefix(response, "I0 I") {
		return "", "", "", fmt.Errorf("invalid command")
	}
	if strings.HasPrefix(response, "I0 L") {
		return "", "", "", fmt.Errorf("command not executable")
	}

	// Expected format: I0 A "type" "serial" "version"
	if !strings.HasPrefix(response, "I0 A") {
		return "", "", "", fmt.Errorf("unexpected info response format")
	}

	// Extract quoted strings
	parts := extractQuotedStrings(response)
	if len(parts) < 3 {
		return "", "", "", fmt.Errorf("incomplete info response")
	}

	return parts[0], parts[1], parts[2], nil
}

// extractQuotedStrings extracts all quoted strings from a string
func extractQuotedStrings(s string) []string {
	var result []string
	var current bytes.Buffer
	inQuote := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if c == '"' {
			if inQuote {
				// End of quoted string
				result = append(result, current.String())
				current.Reset()
				inQuote = false
			} else {
				// Start of quoted string
				inQuote = true
			}
		} else if inQuote {
			current.WriteByte(c)
		}
	}

	return result
}

// ValidateResponse checks if a response is valid MT-SICS format
func (p *MTSICSProtocol) ValidateResponse(data []byte) bool {
	response := strings.TrimSpace(string(data))

	if len(response) == 0 {
		return false
	}

	// Check for valid status codes
	statusByte := response[0]
	validStatus := statusByte == 'S' || statusByte == 'D' ||
		          statusByte == '+' || statusByte == '-' ||
		          statusByte == 'I' || statusByte == 'L'

	return validStatus
}

// IsStableResponse checks if a response indicates stable weight
func (p *MTSICSProtocol) IsStableResponse(data []byte) bool {
	response := strings.TrimSpace(string(data))
	return len(response) > 0 && response[0] == 'S'
}

// IsErrorResponse checks if a response indicates an error
func (p *MTSICSProtocol) IsErrorResponse(data []byte) bool {
	response := strings.TrimSpace(string(data))
	if len(response) == 0 {
		return true
	}

	statusByte := response[0]
	return statusByte == '+' || statusByte == '-' ||
	       statusByte == 'I' || statusByte == 'L'
}

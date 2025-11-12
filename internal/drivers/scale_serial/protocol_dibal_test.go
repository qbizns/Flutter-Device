package scale_serial

import (
	"testing"
)

func TestDibalProtocol_Name(t *testing.T) {
	protocol := NewDibalProtocol()
	if protocol.Name() != "Dibal" {
		t.Errorf("Expected protocol name 'Dibal', got '%s'", protocol.Name())
	}
}

func TestDibalProtocol_Commands(t *testing.T) {
	protocol := NewDibalProtocol()

	tests := []struct {
		name     string
		command  func() ([]byte, error)
		expected []byte
	}{
		{
			name:     "SendWeight",
			command:  protocol.SendWeight,
			expected: []byte{0x02, 'W', 0x03},
		},
		{
			name:     "SendZero",
			command:  protocol.SendZero,
			expected: []byte{0x02, 'Z', 0x03},
		},
		{
			name:     "SendTare",
			command:  protocol.SendTare,
			expected: []byte{0x02, 'T', 0x03},
		},
		{
			name:     "SendReset",
			command:  protocol.SendReset,
			expected: []byte{0x02, 'R', 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.command()
			if err != nil {
				t.Fatalf("Failed to generate command: %v", err)
			}
			if string(cmd) != string(tt.expected) {
				t.Errorf("Expected command %v, got %v", tt.expected, cmd)
			}
		})
	}
}

func TestDibalProtocol_LineTerminator(t *testing.T) {
	protocol := NewDibalProtocol()
	terminator := protocol.LineTerminator()
	expected := []byte{0x03} // ETX

	if string(terminator) != string(expected) {
		t.Errorf("Expected terminator %v, got %v", expected, terminator)
	}
}

func TestDibalProtocol_RequiresCommand(t *testing.T) {
	protocol := NewDibalProtocol()
	// Dibal typically runs in continuous mode
	if protocol.RequiresCommand() != false {
		t.Error("Expected RequiresCommand() to return false for continuous mode")
	}
}

func TestDibalProtocol_ParseResponse(t *testing.T) {
	protocol := NewDibalProtocol()

	tests := []struct {
		name           string
		response       []byte
		expectedValue  float64
		expectedUnit   WeightUnit
		expectedStable bool
		expectedStatus WeightStatus
		expectError    bool
	}{
		{
			name:           "Stable weight in kg",
			response:       []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g', 0x03},
			expectedValue:  1.23,
			expectedUnit:   UnitKilogram,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Unstable weight in g",
			response:       []byte{0x02, '-', '0', '1', '5', '0', '0', 'g', 0x03},
			expectedValue:  15.00,
			expectedUnit:   UnitGram,
			expectedStable: false,
			expectedStatus: StatusUnstable,
			expectError:    false,
		},
		{
			name:           "Stable weight in lb",
			response:       []byte{0x02, '+', '0', '0', '0', '5', '5', 'l', 'b', 0x03},
			expectedValue:  0.55,
			expectedUnit:   UnitPound,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Overload",
			response:       []byte{0x02, 'O', '9', '9', '9', '9', '9', 'k', 'g', 0x03},
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusOverload,
			expectError:    true,
		},
		{
			name:           "Error",
			response:       []byte{0x02, 'E', '0', '0', '0', '0', '0', 0x03},
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusError,
			expectError:    true,
		},
		{
			name:           "Underload",
			response:       []byte{0x02, 'U', '0', '0', '0', '0', '0', 'k', 'g', 0x03},
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusUnderload,
			expectError:    true,
		},
		{
			name:           "Too short",
			response:       []byte{0x02, '+', 0x03},
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusError,
			expectError:    true,
		},
		{
			name:           "Missing STX",
			response:       []byte{'+', '0', '0', '1', '2', '3', 'k', 'g', 0x03},
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusError,
			expectError:    true,
		},
		{
			name:           "Missing ETX",
			response:       []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g'},
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, status, err := protocol.ParseResponse(tt.response)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if status != tt.expectedStatus {
					t.Errorf("Expected status %v, got %v", tt.expectedStatus, status)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if reading == nil {
				t.Fatal("Expected reading, got nil")
			}

			if reading.Value != tt.expectedValue {
				t.Errorf("Expected value %.2f, got %.2f", tt.expectedValue, reading.Value)
			}

			if reading.Unit != tt.expectedUnit {
				t.Errorf("Expected unit %v, got %v", tt.expectedUnit, reading.Unit)
			}

			if reading.Stable != tt.expectedStable {
				t.Errorf("Expected stable %v, got %v", tt.expectedStable, reading.Stable)
			}

			if status != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, status)
			}
		})
	}
}

func TestDibalProtocol_ParseWithChecksum(t *testing.T) {
	protocol := NewDibalProtocol()

	// Create a response with checksum
	// Format: <STX><Status><Weight><Unit><Checksum><ETX>
	// Example: <STX>+00123kg<checksum><ETX>
	content := []byte{'+', '0', '0', '1', '2', '3', 'k', 'g'}

	// Calculate checksum (XOR of content bytes)
	checksum := byte(0)
	for _, b := range content {
		checksum ^= b
	}

	// Build complete response
	response := []byte{0x02} // STX
	response = append(response, content...)
	response = append(response, checksum)
	response = append(response, 0x03) // ETX

	reading, status, err := protocol.ParseWithChecksum(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if reading == nil {
		t.Fatal("Expected reading, got nil")
	}

	if reading.Value != 1.23 {
		t.Errorf("Expected value 1.23, got %.2f", reading.Value)
	}

	if reading.Unit != UnitKilogram {
		t.Errorf("Expected unit kg, got %v", reading.Unit)
	}

	if reading.Stable != true {
		t.Error("Expected stable weight")
	}

	if status != StatusStable {
		t.Errorf("Expected status stable, got %v", status)
	}
}

func TestDibalProtocol_ParseWithChecksum_Invalid(t *testing.T) {
	protocol := NewDibalProtocol()

	// Create response with wrong checksum
	response := []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g', 0xFF, 0x03}

	_, status, err := protocol.ParseWithChecksum(response)
	if err == nil {
		t.Error("Expected checksum error, got nil")
	}

	if status != StatusError {
		t.Errorf("Expected status error, got %v", status)
	}

	if err != nil && !contains(err.Error(), "checksum") {
		t.Errorf("Expected checksum error message, got: %v", err)
	}
}

func TestDibalProtocolConfigurable_DecimalPlaces(t *testing.T) {
	tests := []struct {
		name           string
		decimalPlaces  int
		response       []byte
		expectedValue  float64
	}{
		{
			name:          "0 decimal places",
			decimalPlaces: 0,
			response:      []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g', 0x03},
			expectedValue: 123.0,
		},
		{
			name:          "1 decimal place",
			decimalPlaces: 1,
			response:      []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g', 0x03},
			expectedValue: 12.3,
		},
		{
			name:          "2 decimal places (default)",
			decimalPlaces: 2,
			response:      []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g', 0x03},
			expectedValue: 1.23,
		},
		{
			name:          "3 decimal places",
			decimalPlaces: 3,
			response:      []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g', 0x03},
			expectedValue: 0.123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DibalProtocolConfig{
				DecimalPlaces: tt.decimalPlaces,
				UseChecksum:   false,
				Continuous:    true,
			}
			protocol := NewDibalProtocolWithConfig(config)

			reading, _, err := protocol.ParseResponse(tt.response)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if reading == nil {
				t.Fatal("Expected reading, got nil")
			}

			// Allow small floating point differences
			diff := reading.Value - tt.expectedValue
			if diff < -0.001 || diff > 0.001 {
				t.Errorf("Expected value %.3f, got %.3f", tt.expectedValue, reading.Value)
			}
		})
	}
}

func TestDibalProtocolConfigurable_RequiresCommand(t *testing.T) {
	tests := []struct {
		name           string
		continuous     bool
		expectedResult bool
	}{
		{
			name:           "Continuous mode",
			continuous:     true,
			expectedResult: false,
		},
		{
			name:           "Command mode",
			continuous:     false,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := DibalProtocolConfig{
				DecimalPlaces: 2,
				UseChecksum:   false,
				Continuous:    tt.continuous,
			}
			protocol := NewDibalProtocolWithConfig(config)

			result := protocol.RequiresCommand()
			if result != tt.expectedResult {
				t.Errorf("Expected RequiresCommand() = %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

// BenchmarkDibalProtocol_ParseResponse benchmarks the parsing performance
func BenchmarkDibalProtocol_ParseResponse(b *testing.B) {
	protocol := NewDibalProtocol()
	response := []byte{0x02, '+', '0', '0', '1', '2', '3', 'k', 'g', 0x03}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = protocol.ParseResponse(response)
	}
}

// Helper function to check if error message contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

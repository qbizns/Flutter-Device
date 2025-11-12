package scale_serial

import (
	"testing"
)

func TestToledoProtocol_Name(t *testing.T) {
	protocol := NewToledoProtocol()
	if protocol.Name() != "Toledo-8217" {
		t.Errorf("Expected protocol name 'Toledo-8217', got '%s'", protocol.Name())
	}
}

func TestToledoProtocol_LineTerminator(t *testing.T) {
	protocol := NewToledoProtocol()
	terminator := protocol.LineTerminator()
	expected := []byte("\r\n")

	if string(terminator) != string(expected) {
		t.Errorf("Expected terminator %v, got %v", expected, terminator)
	}
}

func TestToledoProtocol_RequiresCommand(t *testing.T) {
	protocol := NewToledoProtocol()
	// Toledo 8217 typically runs in continuous mode
	if protocol.RequiresCommand() != false {
		t.Error("Expected RequiresCommand() to return false for continuous mode")
	}
}

func TestToledoProtocol_ParseResponse(t *testing.T) {
	protocol := NewToledoProtocol()

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
			response:       []byte("S    123.45 kg\r\n"),
			expectedValue:  123.45,
			expectedUnit:   UnitKilogram,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Dynamic weight in lb",
			response:       []byte("D     50.2 lb\r\n"),
			expectedValue:  50.2,
			expectedUnit:   UnitPound,
			expectedStable: false,
			expectedStatus: StatusUnstable,
			expectError:    false,
		},
		{
			name:           "Stable with negative (tare)",
			response:       []byte("S   -001.23 kg\r\n"),
			expectedValue:  -1.23,
			expectedUnit:   UnitKilogram,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Motion detected",
			response:       []byte("M    100.5 g\r\n"),
			expectedValue:  100.5,
			expectedUnit:   UnitGram,
			expectedStable: false,
			expectedStatus: StatusUnstable,
			expectError:    false,
		},
		{
			name:           "Status with motion flag",
			response:       []byte("S M 123.45 kg\r\n"),
			expectedValue:  123.45,
			expectedUnit:   UnitKilogram,
			expectedStable: false, // Motion flag makes it unstable
			expectedStatus: StatusUnstable,
			expectError:    false,
		},
		{
			name:           "Overload",
			response:       []byte("+   9999.99 g\r\n"),
			expectedValue:  0,
			expectedUnit:   UnitGram,
			expectedStable: false,
			expectedStatus: StatusOverload,
			expectError:    true,
		},
		{
			name:           "Error",
			response:       []byte("?      0.00 kg\r\n"),
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusError,
			expectError:    true,
		},
		{
			name:           "Compact format",
			response:       []byte("S 1.5 kg\r\n"),
			expectedValue:  1.5,
			expectedUnit:   UnitKilogram,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Weight in ounces",
			response:       []byte("S   16.00 oz\r\n"),
			expectedValue:  16.00,
			expectedUnit:   UnitOunce,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Too short",
			response:       []byte("S \r\n"),
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusError,
			expectError:    true,
		},
		{
			name:           "Empty response",
			response:       []byte(""),
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusError,
			expectError:    true,
		},
		{
			name:           "Invalid format (no unit)",
			response:       []byte("S 123.45\r\n"),
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

func TestToledoProtocol_ParseFixedWidth(t *testing.T) {
	protocol := NewToledoProtocol()

	tests := []struct {
		name           string
		response       []byte
		decimalPlaces  int
		expectedValue  float64
		expectedUnit   WeightUnit
		expectedStable bool
		expectedStatus WeightStatus
		expectError    bool
	}{
		{
			name:           "Fixed width 2 decimals",
			response:       []byte("S  00001234 kg\r\n"),
			decimalPlaces:  2,
			expectedValue:  12.34,
			expectedUnit:   UnitKilogram,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Fixed width 1 decimal",
			response:       []byte("S  00001234 kg\r\n"),
			decimalPlaces:  1,
			expectedValue:  123.4,
			expectedUnit:   UnitKilogram,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Fixed width 0 decimals",
			response:       []byte("S  00001234 kg\r\n"),
			decimalPlaces:  0,
			expectedValue:  1234.0,
			expectedUnit:   UnitKilogram,
			expectedStable: true,
			expectedStatus: StatusStable,
			expectError:    false,
		},
		{
			name:           "Fixed width 3 decimals",
			response:       []byte("D  00001234 g\r\n"),
			decimalPlaces:  3,
			expectedValue:  1.234,
			expectedUnit:   UnitGram,
			expectedStable: false,
			expectedStatus: StatusUnstable,
			expectError:    false,
		},
		{
			name:           "Fixed width overload",
			response:       []byte("+  99999999 kg\r\n"),
			decimalPlaces:  2,
			expectedValue:  0,
			expectedUnit:   UnitKilogram,
			expectedStable: false,
			expectedStatus: StatusOverload,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, status, err := protocol.ParseFixedWidth(tt.response, tt.decimalPlaces)

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
				t.Errorf("Expected value %.3f, got %.3f", tt.expectedValue, reading.Value)
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

func TestToledoProtocolConfigurable(t *testing.T) {
	tests := []struct {
		name           string
		config         ToledoProtocolConfig
		response       []byte
		expectedValue  float64
		expectedRequiresCommand bool
	}{
		{
			name: "Variable width continuous mode",
			config: ToledoProtocolConfig{
				FixedWidth:    false,
				DecimalPlaces: 0,
				Continuous:    true,
			},
			response:                []byte("S 123.45 kg\r\n"),
			expectedValue:           123.45,
			expectedRequiresCommand: false,
		},
		{
			name: "Fixed width command mode",
			config: ToledoProtocolConfig{
				FixedWidth:    true,
				DecimalPlaces: 2,
				Continuous:    false,
			},
			response:                []byte("S  00001234 kg\r\n"),
			expectedValue:           12.34,
			expectedRequiresCommand: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protocol := NewToledoProtocolWithConfig(tt.config)

			// Test RequiresCommand
			if protocol.RequiresCommand() != tt.expectedRequiresCommand {
				t.Errorf("Expected RequiresCommand() = %v, got %v",
					tt.expectedRequiresCommand, protocol.RequiresCommand())
			}

			// Test ParseResponse
			reading, _, err := protocol.ParseResponse(tt.response)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if reading == nil {
				t.Fatal("Expected reading, got nil")
			}

			if reading.Value != tt.expectedValue {
				t.Errorf("Expected value %.2f, got %.2f", tt.expectedValue, reading.Value)
			}
		})
	}
}

func TestToledoProtocol_ValidateResponse(t *testing.T) {
	protocol := NewToledoProtocol()

	tests := []struct {
		name     string
		response []byte
		expected bool
	}{
		{
			name:     "Valid stable weight",
			response: []byte("S 123.45 kg\r\n"),
			expected: true,
		},
		{
			name:     "Valid dynamic weight",
			response: []byte("D 50.2 lb\r\n"),
			expected: true,
		},
		{
			name:     "Valid motion",
			response: []byte("M 100.0 g\r\n"),
			expected: true,
		},
		{
			name:     "Valid overload",
			response: []byte("+ 9999.99 kg\r\n"),
			expected: true,
		},
		{
			name:     "Valid error",
			response: []byte("? 0.00 kg\r\n"),
			expected: true,
		},
		{
			name:     "Invalid - too short",
			response: []byte("S\r\n"),
			expected: false,
		},
		{
			name:     "Invalid - no CR/LF",
			response: []byte("S 123.45 kg"),
			expected: false,
		},
		{
			name:     "Invalid - wrong status",
			response: []byte("X 123.45 kg\r\n"),
			expected: false,
		},
		{
			name:     "Empty",
			response: []byte(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := protocol.ValidateResponse(tt.response)
			if result != tt.expected {
				t.Errorf("Expected ValidateResponse() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToledoProtocol_IsStableResponse(t *testing.T) {
	protocol := NewToledoProtocol()

	tests := []struct {
		name     string
		response []byte
		expected bool
	}{
		{
			name:     "Stable uppercase",
			response: []byte("S 123.45 kg\r\n"),
			expected: true,
		},
		{
			name:     "Stable lowercase",
			response: []byte("s 123.45 kg\r\n"),
			expected: true,
		},
		{
			name:     "Dynamic",
			response: []byte("D 50.2 lb\r\n"),
			expected: false,
		},
		{
			name:     "Motion",
			response: []byte("M 100.0 g\r\n"),
			expected: false,
		},
		{
			name:     "Empty",
			response: []byte(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := protocol.IsStableResponse(tt.response)
			if result != tt.expected {
				t.Errorf("Expected IsStableResponse() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToledoProtocol_IsErrorResponse(t *testing.T) {
	protocol := NewToledoProtocol()

	tests := []struct {
		name     string
		response []byte
		expected bool
	}{
		{
			name:     "Error",
			response: []byte("? 0.00 kg\r\n"),
			expected: true,
		},
		{
			name:     "Overload",
			response: []byte("+ 9999.99 kg\r\n"),
			expected: true,
		},
		{
			name:     "Underload",
			response: []byte("- 0.00 kg\r\n"),
			expected: true,
		},
		{
			name:     "Stable",
			response: []byte("S 123.45 kg\r\n"),
			expected: false,
		},
		{
			name:     "Dynamic",
			response: []byte("D 50.2 lb\r\n"),
			expected: false,
		},
		{
			name:     "Empty",
			response: []byte(""),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := protocol.IsErrorResponse(tt.response)
			if result != tt.expected {
				t.Errorf("Expected IsErrorResponse() = %v, got %v", tt.expected, result)
			}
		})
	}
}

// BenchmarkToledoProtocol_ParseResponse benchmarks the parsing performance
func BenchmarkToledoProtocol_ParseResponse(b *testing.B) {
	protocol := NewToledoProtocol()
	response := []byte("S    123.45 kg\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = protocol.ParseResponse(response)
	}
}

// BenchmarkToledoProtocol_ParseFixedWidth benchmarks fixed-width parsing
func BenchmarkToledoProtocol_ParseFixedWidth(b *testing.B) {
	protocol := NewToledoProtocol()
	response := []byte("S  00001234 kg\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = protocol.ParseFixedWidth(response, 2)
	}
}

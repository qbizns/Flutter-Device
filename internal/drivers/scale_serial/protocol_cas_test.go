package scale_serial

import (
	"testing"
)

func TestCASProtocol_Name(t *testing.T) {
	p := NewCASProtocol()
	if p.Name() != "CAS" {
		t.Errorf("expected 'CAS', got '%s'", p.Name())
	}
}

func TestCASProtocol_Commands(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name     string
		method   func() ([]byte, error)
		expected []byte
	}{
		{"SendWeight", p.SendWeight, []byte{0x02, 'W', 0x03}},
		{"SendStableWeight", p.SendStableWeight, []byte{0x02, 'W', 0x03}},
		{"SendZero", p.SendZero, []byte{0x02, 'Z', 0x03}},
		{"SendTare", p.SendTare, []byte{0x02, 'T', 0x03}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.method()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if len(cmd) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(cmd))
			}
			for i := range cmd {
				if cmd[i] != tt.expected[i] {
					t.Errorf("byte %d: expected 0x%02x, got 0x%02x", i, tt.expected[i], cmd[i])
				}
			}
		})
	}
}

func TestCASProtocol_RequiresCommand(t *testing.T) {
	p := NewCASProtocol()
	if !p.RequiresCommand() {
		t.Error("CAS protocol should require commands")
	}
}

func TestCASProtocol_LineTerminator(t *testing.T) {
	p := NewCASProtocol()
	term := p.LineTerminator()
	expected := []byte{0x03} // ETX

	if len(term) != 1 || term[0] != expected[0] {
		t.Errorf("expected ETX (0x03), got %v", term)
	}
}

func TestCASProtocol_ParseResponse_Stable(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name       string
		response   []byte
		wantValue  float64
		wantUnit   WeightUnit
		wantStable bool
		wantStatus WeightStatus
		wantError  bool
	}{
		{
			name:       "stable grams",
			response:   []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			wantValue:  123.4,
			wantUnit:   UnitGram,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
		{
			name:       "stable kilograms",
			response:   []byte{0x02, 'S', '0', '0', '5', '0', '0', '2', 'k', 'g', 0x03},
			wantValue:  500.2,
			wantUnit:   UnitKilogram,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
		{
			name:       "stable pounds",
			response:   []byte{0x02, 'S', '0', '0', '0', '5', '0', '0', 'l', 'b', 0x03},
			wantValue:  50.0,
			wantUnit:   UnitPound,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
		{
			name:       "stable zero",
			response:   []byte{0x02, 'S', '0', '0', '0', '0', '0', '0', 'g', 0x03},
			wantValue:  0.0,
			wantUnit:   UnitGram,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, status, err := p.ParseResponse(tt.response)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Value != tt.wantValue {
				t.Errorf("value: expected %f, got %f", tt.wantValue, reading.Value)
			}

			if reading.Unit != tt.wantUnit {
				t.Errorf("unit: expected %v, got %v", tt.wantUnit, reading.Unit)
			}

			if reading.Stable != tt.wantStable {
				t.Errorf("stable: expected %v, got %v", tt.wantStable, reading.Stable)
			}

			if status != tt.wantStatus {
				t.Errorf("status: expected %v, got %v", tt.wantStatus, status)
			}

			if len(reading.Raw) == 0 {
				t.Error("raw response should be preserved")
			}
		})
	}
}

func TestCASProtocol_ParseResponse_Unstable(t *testing.T) {
	p := NewCASProtocol()

	response := []byte{0x02, 'U', '0', '0', '5', '0', '0', '2', 'k', 'g', 0x03}
	reading, status, err := p.ParseResponse(response)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.Value != 500.2 {
		t.Errorf("value: expected 500.2, got %f", reading.Value)
	}

	if reading.Unit != UnitKilogram {
		t.Errorf("unit: expected Kilogram, got %v", reading.Unit)
	}

	if reading.Stable {
		t.Error("expected unstable weight")
	}

	if status != StatusUnstable {
		t.Errorf("status: expected Unstable, got %v", status)
	}
}

func TestCASProtocol_ParseResponse_Errors(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name       string
		response   []byte
		wantStatus WeightStatus
	}{
		{
			name:       "error status",
			response:   []byte{0x02, 'E', '0', '0', '0', '0', '0', '0', 'g', 0x03},
			wantStatus: StatusError,
		},
		{
			name:       "too short",
			response:   []byte{0x02, 'S', 0x03},
			wantStatus: StatusError,
		},
		{
			name:       "missing STX",
			response:   []byte{'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			wantStatus: StatusError,
		},
		{
			name:       "missing ETX",
			response:   []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g'},
			wantStatus: StatusError,
		},
		{
			name:       "empty response",
			response:   []byte{},
			wantStatus: StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, status, err := p.ParseResponse(tt.response)

			if err == nil {
				t.Error("expected error, got nil")
			}

			if reading != nil {
				t.Errorf("expected nil reading, got %+v", reading)
			}

			if status != tt.wantStatus {
				t.Errorf("status: expected %v, got %v", tt.wantStatus, status)
			}
		})
	}
}

func TestCASProtocol_ParseResponse_InvalidFormat(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name     string
		response []byte
	}{
		{
			name:     "invalid weight - non-numeric",
			response: []byte{0x02, 'S', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 0x03},
		},
		{
			name:     "invalid unit",
			response: []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'x', 'y', 'z', 0x03},
		},
		{
			name:     "unknown status",
			response: []byte{0x02, 'X', '0', '0', '1', '2', '3', '4', 'g', 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseResponse(tt.response)

			if err == nil {
				t.Error("expected error, got nil")
			}

			if reading != nil {
				t.Errorf("expected nil reading, got %+v", reading)
			}
		})
	}
}

func TestCASProtocol_ValidateResponse(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name     string
		response []byte
		want     bool
	}{
		{
			name:     "valid stable",
			response: []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			want:     true,
		},
		{
			name:     "valid unstable",
			response: []byte{0x02, 'U', '0', '0', '5', '0', '0', '2', 'k', 'g', 0x03},
			want:     true,
		},
		{
			name:     "valid error",
			response: []byte{0x02, 'E', '0', '0', '0', '0', '0', '0', 'g', 0x03},
			want:     true,
		},
		{
			name:     "invalid - unknown status",
			response: []byte{0x02, 'X', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			want:     false,
		},
		{
			name:     "invalid - missing STX",
			response: []byte{'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			want:     false,
		},
		{
			name:     "invalid - missing ETX",
			response: []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g'},
			want:     false,
		},
		{
			name:     "invalid - too short",
			response: []byte{0x02, 'S', 0x03},
			want:     false,
		},
		{
			name:     "invalid - empty",
			response: []byte{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ValidateResponse(tt.response)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCASProtocol_IsStableResponse(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name     string
		response []byte
		want     bool
	}{
		{
			name:     "stable",
			response: []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			want:     true,
		},
		{
			name:     "unstable",
			response: []byte{0x02, 'U', '0', '0', '5', '0', '0', '2', 'k', 'g', 0x03},
			want:     false,
		},
		{
			name:     "error",
			response: []byte{0x02, 'E', '0', '0', '0', '0', '0', '0', 0x03},
			want:     false,
		},
		{
			name:     "empty",
			response: []byte{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.IsStableResponse(tt.response)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCASProtocol_IsErrorResponse(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name     string
		response []byte
		want     bool
	}{
		{
			name:     "error",
			response: []byte{0x02, 'E', '0', '0', '0', '0', '0', '0', 0x03},
			want:     true,
		},
		{
			name:     "stable",
			response: []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			want:     false,
		},
		{
			name:     "unstable",
			response: []byte{0x02, 'U', '0', '0', '5', '0', '0', '2', 'k', 'g', 0x03},
			want:     false,
		},
		{
			name:     "empty",
			response: []byte{},
			want:     true,
		},
		{
			name:     "missing STX",
			response: []byte{'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.IsErrorResponse(tt.response)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCASProtocol_ParseWithChecksum(t *testing.T) {
	p := NewCASProtocol()

	tests := []struct {
		name       string
		response   []byte
		wantValue  float64
		wantUnit   WeightUnit
		wantStable bool
		wantError  bool
	}{
		{
			name: "valid with checksum",
			// <STX>S001234kg<ETX><Checksum>
			// Content: S001234kg
			// XOR: 'S' ^ '0' ^ '0' ^ '1' ^ '2' ^ '3' ^ '4' ^ 'k' ^ 'g'
			// = 0x53 ^ 0x30 ^ 0x30 ^ 0x31 ^ 0x32 ^ 0x33 ^ 0x34 ^ 0x6B ^ 0x67
			// = 0x5B
			response:   []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'k', 'g', 0x03, 0x5B},
			wantValue:  123.4,
			wantUnit:   UnitKilogram,
			wantStable: true,
			wantError:  false,
		},
		{
			name: "invalid checksum",
			// Same data but wrong checksum
			response:   []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'k', 'g', 0x03, 0xFF},
			wantValue:  0,
			wantUnit:   0,
			wantStable: false,
			wantError:  true,
		},
		{
			name: "no checksum - should still parse",
			// Without checksum byte, should fall back to standard parsing
			response:   []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'k', 'g', 0x03},
			wantValue:  123.4,
			wantUnit:   UnitKilogram,
			wantStable: true,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseWithChecksum(tt.response)

			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Value != tt.wantValue {
				t.Errorf("value: expected %f, got %f", tt.wantValue, reading.Value)
			}

			if reading.Unit != tt.wantUnit {
				t.Errorf("unit: expected %v, got %v", tt.wantUnit, reading.Unit)
			}

			if reading.Stable != tt.wantStable {
				t.Errorf("stable: expected %v, got %v", tt.wantStable, reading.Stable)
			}
		})
	}
}

func TestCASProtocolConfigurable_DecimalPlaces(t *testing.T) {
	tests := []struct {
		name          string
		decimalPlaces int
		response      []byte
		wantValue     float64
	}{
		{
			name:          "0 decimal places",
			decimalPlaces: 0,
			response:      []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			wantValue:     1234.0, // 001234 = 1234
		},
		{
			name:          "1 decimal place (default)",
			decimalPlaces: 1,
			response:      []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			wantValue:     123.4, // 001234 = 123.4
		},
		{
			name:          "2 decimal places",
			decimalPlaces: 2,
			response:      []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03},
			wantValue:     12.34, // 001234 = 12.34
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := CASProtocolConfig{
				DecimalPlaces: tt.decimalPlaces,
				UseChecksum:   false,
			}
			p := NewCASProtocolWithConfig(config)

			reading, _, err := p.ParseResponse(tt.response)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if reading.Value != tt.wantValue {
				t.Errorf("value: expected %f, got %f", tt.wantValue, reading.Value)
			}
		})
	}
}

func TestCASProtocolConfigurable_WithChecksum(t *testing.T) {
	config := CASProtocolConfig{
		DecimalPlaces: 1,
		UseChecksum:   true,
	}
	p := NewCASProtocolWithConfig(config)

	// Valid response with checksum
	// XOR: 'S' ^ '0' ^ '0' ^ '1' ^ '2' ^ '3' ^ '4' ^ 'k' ^ 'g' = 0x5B
	response := []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'k', 'g', 0x03, 0x5B}

	reading, status, err := p.ParseResponse(response)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reading.Value != 123.4 {
		t.Errorf("value: expected 123.4, got %f", reading.Value)
	}

	if status != StatusStable {
		t.Errorf("status: expected Stable, got %v", status)
	}
}

func BenchmarkCASProtocol_ParseResponse(b *testing.B) {
	p := NewCASProtocol()
	response := []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'g', 0x03}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseResponse(response)
	}
}

func BenchmarkCASProtocol_ParseResponse_WithChecksum(b *testing.B) {
	p := NewCASProtocol()
	response := []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'k', 'g', 0x03, 0x5B}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseWithChecksum(response)
	}
}

func BenchmarkCASProtocolConfigurable_ParseResponse(b *testing.B) {
	config := CASProtocolConfig{
		DecimalPlaces: 2,
		UseChecksum:   true,
	}
	p := NewCASProtocolWithConfig(config)
	response := []byte{0x02, 'S', '0', '0', '1', '2', '3', '4', 'k', 'g', 0x03, 0x5B}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseResponse(response)
	}
}

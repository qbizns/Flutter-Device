package scale_serial

import (
	"testing"
)

func TestMTSICSProtocol_Name(t *testing.T) {
	p := NewMTSICSProtocol()
	if p.Name() != "MT-SICS" {
		t.Errorf("expected 'MT-SICS', got '%s'", p.Name())
	}
}

func TestMTSICSProtocol_Commands(t *testing.T) {
	p := NewMTSICSProtocol()

	tests := []struct {
		name     string
		method   func() ([]byte, error)
		expected string
	}{
		{"SendWeight", p.SendWeight, "SI\r\n"},
		{"SendStableWeight", p.SendStableWeight, "S\r\n"},
		{"SendZero", p.SendZero, "ZI\r\n"},
		{"SendTare", p.SendTare, "T\r\n"},
		{"SendReset", p.SendReset, "@\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := tt.method()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(cmd) != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, string(cmd))
			}
		})
	}
}

func TestMTSICSProtocol_ParseResponse_Stable(t *testing.T) {
	p := NewMTSICSProtocol()

	tests := []struct {
		name       string
		response   string
		wantValue  float64
		wantUnit   WeightUnit
		wantStable bool
		wantStatus WeightStatus
		wantError  bool
	}{
		{
			name:       "stable grams",
			response:   "S       1234.5 g\r\n",
			wantValue:  1234.5,
			wantUnit:   UnitGram,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
		{
			name:       "stable kilograms",
			response:   "S        500.2 kg\r\n",
			wantValue:  500.2,
			wantUnit:   UnitKilogram,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
		{
			name:       "stable pounds negative",
			response:   "S     -00012.3 lb\r\n",
			wantValue:  -12.3,
			wantUnit:   UnitPound,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
		{
			name:       "stable zero",
			response:   "S          0.0 g\r\n",
			wantValue:  0.0,
			wantUnit:   UnitGram,
			wantStable: true,
			wantStatus: StatusStable,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, status, err := p.ParseResponse([]byte(tt.response))

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

			if reading.Raw != "S       1234.5 g" && reading.Raw != "S        500.2 kg" &&
			   reading.Raw != "S     -00012.3 lb" && reading.Raw != "S          0.0 g" {
				// Basic check that raw is preserved (trimmed)
				if len(reading.Raw) == 0 {
					t.Error("raw response should be preserved")
				}
			}
		})
	}
}

func TestMTSICSProtocol_ParseResponse_Unstable(t *testing.T) {
	p := NewMTSICSProtocol()

	response := "D        500.2 kg\r\n"
	reading, status, err := p.ParseResponse([]byte(response))

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

func TestMTSICSProtocol_ParseResponse_Errors(t *testing.T) {
	p := NewMTSICSProtocol()

	tests := []struct {
		name       string
		response   string
		wantStatus WeightStatus
	}{
		{
			name:       "overload",
			response:   "+       9999.9 g\r\n",
			wantStatus: StatusOverload,
		},
		{
			name:       "underload",
			response:   "-          0.0 g\r\n",
			wantStatus: StatusUnderload,
		},
		{
			name:       "invalid command",
			response:   "I\r\n",
			wantStatus: StatusError,
		},
		{
			name:       "command not executable",
			response:   "L\r\n",
			wantStatus: StatusError,
		},
		{
			name:       "empty response",
			response:   "",
			wantStatus: StatusError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, status, err := p.ParseResponse([]byte(tt.response))

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

func TestMTSICSProtocol_ParseResponse_InvalidFormat(t *testing.T) {
	p := NewMTSICSProtocol()

	tests := []struct {
		name     string
		response string
	}{
		{"no unit", "S       1234.5\r\n"},
		{"no weight", "S g\r\n"},
		{"invalid weight", "S       abc g\r\n"},
		{"invalid unit", "S       123.4 xyz\r\n"},
		{"unknown status", "X       123.4 g\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading, _, err := p.ParseResponse([]byte(tt.response))

			if err == nil {
				t.Error("expected error, got nil")
			}

			if reading != nil {
				t.Errorf("expected nil reading, got %+v", reading)
			}
		})
	}
}

func TestMTSICSProtocol_ParseInfoResponse(t *testing.T) {
	p := NewMTSICSProtocol()

	response := `I0 A "PS60" "1234567" "V1.0"`
	scaleType, serial, version, err := p.ParseInfoResponse([]byte(response))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if scaleType != "PS60" {
		t.Errorf("scaleType: expected 'PS60', got '%s'", scaleType)
	}

	if serial != "1234567" {
		t.Errorf("serial: expected '1234567', got '%s'", serial)
	}

	if version != "V1.0" {
		t.Errorf("version: expected 'V1.0', got '%s'", version)
	}
}

func TestMTSICSProtocol_ValidateResponse(t *testing.T) {
	p := NewMTSICSProtocol()

	tests := []struct {
		name     string
		response string
		want     bool
	}{
		{"stable", "S       1234.5 g\r\n", true},
		{"unstable", "D        500.2 kg\r\n", true},
		{"overload", "+       9999.9 g\r\n", true},
		{"underload", "-          0.0 g\r\n", true},
		{"invalid command", "I\r\n", true},
		{"not executable", "L\r\n", true},
		{"unknown status", "X       123.4 g\r\n", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.ValidateResponse([]byte(tt.response))
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestMTSICSProtocol_IsStableResponse(t *testing.T) {
	p := NewMTSICSProtocol()

	tests := []struct {
		name     string
		response string
		want     bool
	}{
		{"stable", "S       1234.5 g\r\n", true},
		{"unstable", "D        500.2 kg\r\n", false},
		{"overload", "+       9999.9 g\r\n", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.IsStableResponse([]byte(tt.response))
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestMTSICSProtocol_IsErrorResponse(t *testing.T) {
	p := NewMTSICSProtocol()

	tests := []struct {
		name     string
		response string
		want     bool
	}{
		{"overload", "+       9999.9 g\r\n", true},
		{"underload", "-          0.0 g\r\n", true},
		{"invalid", "I\r\n", true},
		{"not executable", "L\r\n", true},
		{"stable", "S       1234.5 g\r\n", false},
		{"unstable", "D        500.2 kg\r\n", false},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.IsErrorResponse([]byte(tt.response))
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func BenchmarkMTSICSProtocol_ParseResponse(b *testing.B) {
	p := NewMTSICSProtocol()
	response := []byte("S       1234.5 g\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseResponse(response)
	}
}

func BenchmarkMTSICSProtocol_ParseResponse_Complex(b *testing.B) {
	p := NewMTSICSProtocol()
	response := []byte("S     -00012.345 kg\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ParseResponse(response)
	}
}

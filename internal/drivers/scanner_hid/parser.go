package scanner_hid

import (
	"strings"
)

// Parser parses HID reports from barcode scanners
//
// USB HID barcode scanners typically send data as keyboard input.
// Each key press is sent as a HID report containing:
// - Modifier keys (Shift, Ctrl, etc.)
// - Key code (standard USB HID usage codes)
//
// The parser accumulates key codes until Enter is detected,
// then returns the complete barcode string.
type Parser struct {
	buffer      strings.Builder
	shiftActive bool
}

// NewParser creates a new HID report parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a HID report and returns barcode data if complete
//
// Returns:
//   - barcode: The scanned barcode string
//   - symbology: The barcode symbology (e.g., "EAN13", "CODE128")
//   - ok: true if a complete barcode was parsed
//
// HID Report Format (simplified):
//   Byte 0: Modifier keys (Shift, Ctrl, Alt, GUI)
//   Byte 1: Reserved (usually 0)
//   Byte 2-7: Key codes (up to 6 simultaneous keys)
//
// Most scanners send one character per report, followed by Enter.
func (p *Parser) Parse(report []byte) (barcode, symbology string, ok bool) {
	if len(report) < 3 {
		return "", "", false
	}

	modifier := report[0]
	keyCodes := report[2:]

	// Check for Shift modifier
	p.shiftActive = (modifier & 0x02) != 0 || (modifier & 0x20) != 0 // Left or Right Shift

	// Process each key code
	for _, keyCode := range keyCodes {
		if keyCode == 0 {
			continue // No key pressed
		}

		// Check for Enter key (marks end of barcode)
		if keyCode == 0x28 { // Enter key
			if p.buffer.Len() > 0 {
				barcode = p.buffer.String()
				symbology = detectSymbology(barcode)
				p.buffer.Reset()
				p.shiftActive = false
				return barcode, symbology, true
			}
			continue
		}

		// Convert key code to character
		if char := p.keyCodeToChar(keyCode); char != 0 {
			p.buffer.WriteByte(char)
		}
	}

	return "", "", false
}

// keyCodeToChar converts a USB HID key code to ASCII character
//
// This is a simplified mapping for common keys used in barcodes.
// Full USB HID usage table: https://www.usb.org/sites/default/files/documents/hut1_12v2.pdf
func (p *Parser) keyCodeToChar(keyCode byte) byte {
	// Numbers (with Shift for special characters)
	if keyCode >= 0x1E && keyCode <= 0x27 {
		if p.shiftActive {
			// Shifted number keys: ! @ # $ % ^ & * ( )
			shifted := []byte{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')'}
			return shifted[keyCode-0x1E]
		}
		// Number keys: 1 2 3 4 5 6 7 8 9 0
		if keyCode == 0x27 {
			return '0'
		}
		return byte('1' + (keyCode - 0x1E))
	}

	// Letters A-Z
	if keyCode >= 0x04 && keyCode <= 0x1D {
		char := byte('a' + (keyCode - 0x04))
		if p.shiftActive {
			char = byte('A' + (keyCode - 0x04))
		}
		return char
	}

	// Special characters
	switch keyCode {
	case 0x2D: // Hyphen/Underscore
		if p.shiftActive {
			return '_'
		}
		return '-'
	case 0x2E: // Equals/Plus
		if p.shiftActive {
			return '+'
		}
		return '='
	case 0x2F: // Left bracket
		if p.shiftActive {
			return '{'
		}
		return '['
	case 0x30: // Right bracket
		if p.shiftActive {
			return '}'
		}
		return ']'
	case 0x31: // Backslash/Pipe
		if p.shiftActive {
			return '|'
		}
		return '\\'
	case 0x33: // Semicolon/Colon
		if p.shiftActive {
			return ':'
		}
		return ';'
	case 0x34: // Apostrophe/Quote
		if p.shiftActive {
			return '"'
		}
		return '\''
	case 0x36: // Comma/Less than
		if p.shiftActive {
			return '<'
		}
		return ','
	case 0x37: // Period/Greater than
		if p.shiftActive {
			return '>'
		}
		return '.'
	case 0x38: // Slash/Question mark
		if p.shiftActive {
			return '?'
		}
		return '/'
	case 0x2C: // Space
		return ' '
	}

	return 0
}

// Reset resets the parser state (clears buffer)
func (p *Parser) Reset() {
	p.buffer.Reset()
	p.shiftActive = false
}

// detectSymbology attempts to detect barcode symbology from the barcode data
//
// This is a heuristic approach based on barcode format and length.
// More accurate detection would require symbology data from the scanner
// (some scanners support AIM identifiers or prefix codes).
func detectSymbology(barcode string) string {
	length := len(barcode)

	// EAN-13 (European Article Number): 13 digits
	if length == 13 && isAllDigits(barcode) {
		return "EAN13"
	}

	// EAN-8: 8 digits
	if length == 8 && isAllDigits(barcode) {
		return "EAN8"
	}

	// UPC-A (Universal Product Code): 12 digits
	if length == 12 && isAllDigits(barcode) {
		return "UPCA"
	}

	// UPC-E: 6 or 8 digits
	if (length == 6 || length == 8) && isAllDigits(barcode) {
		return "UPCE"
	}

	// ITF-14 (Interleaved 2 of 5): 14 digits
	if length == 14 && isAllDigits(barcode) {
		return "ITF14"
	}

	// Code 39: Alphanumeric, typically 8-20 characters
	if length >= 8 && length <= 20 && isAlphanumericWithDash(barcode) {
		return "CODE39"
	}

	// Code 128: Variable length, alphanumeric with special chars
	if length >= 4 && length <= 128 {
		return "CODE128"
	}

	// Default to unknown
	return "UNKNOWN"
}

// isAllDigits checks if string contains only digits
func isAllDigits(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isAlphanumericWithDash checks if string contains only alphanumeric and dash
func isAlphanumericWithDash(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= 'a' && c <= 'z') ||
			c == '-' || c == ' ') {
			return false
		}
	}
	return true
}

// HID Key Code Constants
//
// These are the most commonly used USB HID key codes for barcode scanners.
// Full specification: USB HID Usage Tables v1.12
const (
	// Modifier keys (byte 0)
	HIDModifierLeftCtrl   = 0x01
	HIDModifierLeftShift  = 0x02
	HIDModifierLeftAlt    = 0x04
	HIDModifierLeftGUI    = 0x08
	HIDModifierRightCtrl  = 0x10
	HIDModifierRightShift = 0x20
	HIDModifierRightAlt   = 0x40
	HIDModifierRightGUI   = 0x80

	// Special keys
	HIDKeyEnter     = 0x28
	HIDKeyEscape    = 0x29
	HIDKeyBackspace = 0x2A
	HIDKeyTab       = 0x2B
	HIDKeySpace     = 0x2C

	// Number row (top of keyboard)
	HIDKey1 = 0x1E
	HIDKey2 = 0x1F
	HIDKey3 = 0x20
	HIDKey4 = 0x21
	HIDKey5 = 0x22
	HIDKey6 = 0x23
	HIDKey7 = 0x24
	HIDKey8 = 0x25
	HIDKey9 = 0x26
	HIDKey0 = 0x27

	// Letters
	HIDKeyA = 0x04
	HIDKeyZ = 0x1D
)

// AIMIdentifier represents AIM (Association for Automatic Identification and Mobility) identifiers
//
// Some advanced scanners prefix barcodes with AIM identifiers to indicate symbology:
//   ]E0 = EAN-13
//   ]E4 = EAN-8
//   ]A0 = Code 39
//   ]C0 = Code 128
//   ]d2 = Data Matrix
//   ]Q3 = QR Code
//
// Reference: AIM International Standard Symbology Identifiers
type AIMIdentifier struct {
	Prefix     string
	Symbology  string
	ModifierChar string
}

// Common AIM identifiers
var aimIdentifiers = map[string]string{
	"]E0": "EAN13",
	"]E3": "EAN13",
	"]E4": "EAN8",
	"]A0": "CODE39",
	"]A": "CODE39",
	"]C0": "CODE128",
	"]C1": "CODE128",
	"]d1": "DATAMATRIX",
	"]d2": "DATAMATRIX",
	"]Q1": "QRCODE",
	"]Q3": "QRCODE",
	"]e0": "GS1DATABAR",
	"]I":  "ITF",
}

// ParseWithAIM parses barcode with AIM identifier support
//
// If the barcode starts with ], it's treated as an AIM identifier.
// The function strips the identifier and returns the actual barcode data.
func ParseWithAIM(data string) (barcode, symbology string) {
	if len(data) < 3 || data[0] != ']' {
		return data, detectSymbology(data)
	}

	// Look for AIM identifier (2-3 characters starting with ])
	for prefixLen := 2; prefixLen <= 3 && prefixLen < len(data); prefixLen++ {
		prefix := data[:prefixLen]
		if sym, ok := aimIdentifiers[prefix]; ok {
			return data[prefixLen:], sym
		}
	}

	// No matching AIM identifier, return as-is
	return data, detectSymbology(data)
}

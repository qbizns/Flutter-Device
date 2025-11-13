package printer_escpos

import (
	"bytes"
	"fmt"
	"strings"
)

// TextDirection represents text flow direction
type TextDirection int

const (
	// DirectionLTR is left-to-right (English, numbers)
	DirectionLTR TextDirection = iota
	// DirectionRTL is right-to-left (Arabic, Hebrew)
	DirectionRTL
	// DirectionMixed contains both LTR and RTL runs
	DirectionMixed
)

// TextStyle represents text styling options
type TextStyle struct {
	Bold         bool
	Underline    bool
	DoubleWidth  bool
	DoubleHeight bool
}

// Alignment represents text alignment
type Alignment int

const (
	// AlignLeft aligns text to the left
	AlignLeft Alignment = iota
	// AlignCenter centers text
	AlignCenter
	// AlignRight aligns text to the right
	AlignRight
)

// String returns string representation of direction
func (d TextDirection) String() string {
	switch d {
	case DirectionLTR:
		return "LTR"
	case DirectionRTL:
		return "RTL"
	case DirectionMixed:
		return "Mixed"
	default:
		return "Unknown"
	}
}

// TextRun represents a segment of text with consistent direction
type TextRun struct {
	Text      string
	Direction TextDirection
	Start     int  // Start position in original text
	End       int  // End position in original text
}

// ArabicShaper handles Arabic text shaping for ESC/POS printers
type ArabicShaper struct {
	codePage int          // ESC/POS Code Page (864 for Arabic)
	font     *ESCPOSFont  // Font mapping and shaping engine
}

// NewArabicShaper creates a new Arabic text shaper
func NewArabicShaper() *ArabicShaper {
	return &ArabicShaper{
		codePage: 864,  // PC864 - Arabic
		font:     NewESCPOSArabicFont(),
	}
}

// ContainsArabic checks if text contains any Arabic characters
// Arabic Unicode blocks:
// - Main Arabic: U+0600 to U+06FF
// - Arabic Supplement: U+0750 to U+077F
// - Arabic Extended-A: U+08A0 to U+08FF
// - Arabic Presentation Forms-A: U+FB50 to U+FDFF
// - Arabic Presentation Forms-B: U+FE70 to U+FEFF
func ContainsArabic(text string) bool {
	for _, r := range text {
		if isArabicCharacter(r) {
			return true
		}
	}
	return false
}

// isArabicCharacter checks if a rune is an Arabic character
func isArabicCharacter(r rune) bool {
	return (r >= 0x0600 && r <= 0x06FF) ||  // Main Arabic
		(r >= 0x0750 && r <= 0x077F) ||      // Arabic Supplement
		(r >= 0x08A0 && r <= 0x08FF) ||      // Arabic Extended-A
		(r >= 0xFB50 && r <= 0xFDFF) ||      // Arabic Presentation Forms-A
		(r >= 0xFE70 && r <= 0xFEFF)         // Arabic Presentation Forms-B
}

// DetectTextDirection analyzes text and determines its primary direction
func DetectTextDirection(text string) TextDirection {
	hasArabic := false
	hasLatin := false

	for _, r := range text {
		if isArabicCharacter(r) {
			hasArabic = true
		} else if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			hasLatin = true
		}

		// If we've found both, it's mixed
		if hasArabic && hasLatin {
			return DirectionMixed
		}
	}

	if hasArabic {
		return DirectionRTL
	}
	return DirectionLTR
}

// SplitIntoRuns splits text into directional runs using enhanced BiDi algorithm
// This now uses the full bidirectional algorithm from bidi.go
func SplitIntoRuns(text string) []TextRun {
	if text == "" {
		return nil
	}

	// Detect base direction
	baseDir := DetectTextDirection(text)

	// Use enhanced BiDi analyzer for proper run splitting
	bidiRuns := AnalyzeBiDi(text, baseDir)

	// Convert BiDiRun to TextRun
	runs := make([]TextRun, len(bidiRuns))
	for i, brun := range bidiRuns {
		runs[i] = TextRun{
			Text:      brun.Text,
			Direction: brun.Direction,
			Start:     brun.Start,
			End:       brun.End,
		}
	}

	return runs
}

// RenderArabicText generates ESC/POS commands for Arabic text
// This is a placeholder implementation that will be enhanced with HarfBuzz
func (s *ArabicShaper) RenderArabicText(text string, style TextStyle) []byte {
	var buf bytes.Buffer

	// 1. Set Code Page 864 (Arabic)
	buf.Write(s.setCodePage864())

	// 2. Detect text direction
	direction := DetectTextDirection(text)

	// 3. Handle based on direction
	switch direction {
	case DirectionRTL:
		// Pure Arabic text
		buf.Write(s.setRTL())
		buf.Write(s.renderRTLText(text, style))
		buf.Write(s.resetDirection())

	case DirectionMixed:
		// Mixed Arabic/English - split into runs
		runs := SplitIntoRuns(text)

		for _, run := range runs {
			if run.Direction == DirectionRTL {
				buf.Write(s.setRTL())
				buf.Write(s.renderRTLText(run.Text, style))
				buf.Write(s.resetDirection())
			} else {
				buf.Write(s.renderLTRText(run.Text, style))
			}
		}

	default:
		// LTR (shouldn't happen if ContainsArabic is true, but handle anyway)
		buf.Write(s.renderLTRText(text, style))
	}

	// 4. Reset to default Code Page
	buf.Write(s.resetCodePage())

	return buf.Bytes()
}

// setCodePage864 returns ESC/POS command to set Code Page 864 (Arabic)
// ESC t 0x1E
func (s *ArabicShaper) setCodePage864() []byte {
	return []byte{0x1B, 't', 0x1E}
}

// resetCodePage returns ESC/POS command to reset to default Code Page
// ESC t 0x00 (PC437)
func (s *ArabicShaper) resetCodePage() []byte {
	return []byte{0x1B, 't', 0x00}
}

// setRTL returns ESC/POS command to set right-to-left printing
// ESC { 0x01
func (s *ArabicShaper) setRTL() []byte {
	// Note: Not all ESC/POS printers support this command
	// Some printers auto-detect RTL from Code Page 864
	return []byte{0x1B, 0x7B, 0x01}
}

// resetDirection returns ESC/POS command to reset to LTR
// ESC { 0x00
func (s *ArabicShaper) resetDirection() []byte {
	return []byte{0x1B, 0x7B, 0x00}
}

// renderRTLText renders right-to-left text with proper shaping
func (s *ArabicShaper) renderRTLText(text string, style TextStyle) []byte {
	var buf bytes.Buffer

	// Apply text style (bold, underline, etc.)
	buf.Write(applyTextStyle(style))

	// Shape Arabic text using font engine
	// This handles contextual forms (initial, medial, final, isolated)
	shaped, err := s.font.ShapeText(text, DirectionRTL)
	if err != nil {
		// Fallback: output text as-is if shaping fails
		buf.WriteString(text)
	} else {
		// Write shaped text as CP864 bytes
		buf.Write(shaped)
	}

	// Reset text style
	buf.Write(resetTextStyle())

	return buf.Bytes()
}

// renderLTRText renders left-to-right text
func (s *ArabicShaper) renderLTRText(text string, style TextStyle) []byte {
	var buf bytes.Buffer

	// Apply text style
	buf.Write(applyTextStyle(style))

	// For LTR text, we still use font mapping for consistency
	// This ensures proper Code Page 864 encoding
	shaped, err := s.font.ShapeText(text, DirectionLTR)
	if err != nil {
		// Fallback: output text as-is
		buf.WriteString(text)
	} else {
		buf.Write(shaped)
	}

	buf.Write(resetTextStyle())

	return buf.Bytes()
}

// applyTextStyle returns ESC/POS commands for text styling
func applyTextStyle(style TextStyle) []byte {
	var buf bytes.Buffer

	if style.Bold {
		// ESC E 1 - Enable bold
		buf.Write([]byte{0x1B, 'E', 0x01})
	}

	if style.Underline {
		// ESC - 1 - Enable underline
		buf.Write([]byte{0x1B, '-', 0x01})
	}

	if style.DoubleWidth || style.DoubleHeight {
		// GS ! n - Select character size
		size := byte(0)
		if style.DoubleWidth {
			size |= 0x20
		}
		if style.DoubleHeight {
			size |= 0x10
		}
		buf.Write([]byte{0x1D, '!', size})
	}

	return buf.Bytes()
}

// resetTextStyle returns ESC/POS commands to reset text style
func resetTextStyle() []byte {
	var buf bytes.Buffer

	// ESC E 0 - Disable bold
	buf.Write([]byte{0x1B, 'E', 0x00})
	// ESC - 0 - Disable underline
	buf.Write([]byte{0x1B, '-', 0x00})
	// GS ! 0 - Reset character size
	buf.Write([]byte{0x1D, '!', 0x00})

	return buf.Bytes()
}

// WrapArabicText wraps Arabic text at word boundaries
// maxWidth is in characters (not pixels)
func WrapArabicText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}

	// Split by spaces (word boundaries in Arabic use spaces like English)
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine []string
	currentWidth := 0

	for _, word := range words {
		wordWidth := len([]rune(word))

		if currentWidth == 0 {
			// First word on line
			currentLine = append(currentLine, word)
			currentWidth = wordWidth
		} else if currentWidth+1+wordWidth <= maxWidth {
			// Word fits on current line (including space)
			currentLine = append(currentLine, word)
			currentWidth += 1 + wordWidth  // +1 for space
		} else {
			// Word doesn't fit - start new line
			lines = append(lines, strings.Join(currentLine, " "))
			currentLine = []string{word}
			currentWidth = wordWidth
		}
	}

	// Add final line
	if len(currentLine) > 0 {
		lines = append(lines, strings.Join(currentLine, " "))
	}

	return lines
}

// FormatArabicReceipt is a helper to format a full Arabic receipt line
// with proper alignment and wrapping
func FormatArabicReceipt(text string, alignment Alignment, maxWidth int) []string {
	// 1. Wrap text if needed
	lines := WrapArabicText(text, maxWidth)

	// 2. Apply alignment to each line
	// Note: For Arabic, "right" alignment is natural, "left" is unusual
	aligned := make([]string, len(lines))
	for i, line := range lines {
		aligned[i] = alignText(line, alignment, maxWidth)
	}

	return aligned
}

// alignText aligns a single line of text
func alignText(text string, alignment Alignment, width int) string {
	textWidth := len([]rune(text))
	if textWidth >= width {
		return text
	}

	padding := width - textWidth

	switch alignment {
	case AlignLeft:
		return text + strings.Repeat(" ", padding)
	case AlignCenter:
		leftPad := padding / 2
		rightPad := padding - leftPad
		return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
	case AlignRight:
		return strings.Repeat(" ", padding) + text
	default:
		return text
	}
}

// GetArabicExamples returns common Arabic phrases for testing
func GetArabicExamples() map[string]string {
	return map[string]string{
		"hello":         "مرحباً",
		"thank_you":     "شكراً",
		"total":         "إجمالي",
		"invoice":       "الفاتورة",
		"invoice_total": "إجمالي الفاتورة",
		"price":         "السعر",
		"sar":           "ر.س",    // Saudi Riyal
		"egp":           "ج.م",    // Egyptian Pound
		"aed":           "د.إ",    // UAE Dirham
		"kwd":           "د.ك",    // Kuwaiti Dinar
		"welcome":       "مرحباً بك",
		"receipt":       "إيصال",
		"date":          "التاريخ",
		"time":          "الوقت",
		"quantity":      "الكمية",
		"product":       "المنتج",
	}
}

// Example usage:
//
//	shaper := NewArabicShaper()
//	style := TextStyle{Bold: true}
//	commands := shaper.RenderArabicText("مرحباً", style)
//	// Send commands to printer
//
// For mixed text:
//
//	text := "Coca-Cola كولا"
//	commands := shaper.RenderArabicText(text, style)
//
// For wrapping:
//
//	longText := "محل البقالة الكبير للمواد الغذائية والمشروبات"
//	lines := WrapArabicText(longText, 40)  // 40 characters wide
//	for _, line := range lines {
//	    commands := shaper.RenderArabicText(line, style)
//	    // Print each line
//	}

// TODO List for Phase 2 (HarfBuzz Integration):
// - [ ] Load ESC/POS font metadata
// - [ ] Integrate go-text/typesetting for text shaping
// - [ ] Map shaped glyphs to Code Page 864 character codes
// - [ ] Handle glyph positioning and advances
// - [ ] Implement proper BiDi algorithm (consider using golang.org/x/text/unicode/bidi)
// - [ ] Performance optimization and caching
// - [ ] Test with real printers (Epson, Star, Chinese brands)

// PrinterCapabilities represents what an ESC/POS printer supports
type PrinterCapabilities struct {
	SupportsCodePage864 bool  // Arabic support
	SupportsRTLCommand  bool  // RTL direction command
	MaxLineWidth        int   // Characters per line (usually 42 or 48)
	SupportsAlignment   bool  // Text alignment commands
}

// DetectArabicSupport checks if printer supports Arabic
// This would query the printer in a real implementation
func DetectArabicSupport() PrinterCapabilities {
	// Default capabilities for standard ESC/POS printer
	return PrinterCapabilities{
		SupportsCodePage864: true,  // Most modern printers support CP864
		SupportsRTLCommand:  false, // Not all printers support ESC {
		MaxLineWidth:        42,    // Standard 80mm paper, 12cpi
		SupportsAlignment:   true,  // ESC a command is standard
	}
}

// Error types
var (
	ErrArabicNotSupported   = fmt.Errorf("printer does not support Arabic (Code Page 864)")
	ErrTextTooLong          = fmt.Errorf("text exceeds maximum line width")
	ErrInvalidDirection     = fmt.Errorf("invalid text direction")
)

// ValidateArabicText validates that text can be printed
func ValidateArabicText(text string, caps PrinterCapabilities) error {
	if !caps.SupportsCodePage864 && ContainsArabic(text) {
		return ErrArabicNotSupported
	}

	lineWidth := len([]rune(text))
	if lineWidth > caps.MaxLineWidth {
		return fmt.Errorf("%w: %d characters (max %d)", ErrTextTooLong, lineWidth, caps.MaxLineWidth)
	}

	return nil
}

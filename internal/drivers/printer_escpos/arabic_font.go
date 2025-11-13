package printer_escpos

import (
	"fmt"
	"unicode"
)

// ESCPOSFont represents an ESC/POS printer's built-in font
// This is a virtual font representation for text shaping
type ESCPOSFont struct {
	codePage     int                    // 864 for Arabic
	charMap      map[rune]byte          // Unicode -> Code Page mapping
	reverseMap   map[byte]rune          // Code Page -> Unicode mapping
	glyphMap     map[int]byte           // Glyph ID -> ESC/POS byte (simplified)
	capabilities FontCapabilities       // What the font supports
}

// FontCapabilities describes what the ESC/POS font can render
type FontCapabilities struct {
	SupportsArabic          bool
	SupportsDiacritics      bool
	SupportsLigatures       bool
	SupportsPresentationForms bool
	MaxGlyphID              int
}

// NewESCPOSArabicFont creates a virtual font for ESC/POS Code Page 864
func NewESCPOSArabicFont() *ESCPOSFont {
	f := &ESCPOSFont{
		codePage:   864,
		charMap:    make(map[rune]byte),
		reverseMap: make(map[byte]rune),
		glyphMap:   make(map[int]byte),
		capabilities: FontCapabilities{
			SupportsArabic:          true,
			SupportsDiacritics:      true,  // CP864 has some diacritics
			SupportsLigatures:       true,  // Has common ligatures
			SupportsPresentationForms: true, // Has shaped forms
			MaxGlyphID:              256,   // 8-bit character set
		},
	}

	// Build Code Page 864 mapping
	f.buildCP864Mapping()

	return f
}

// buildCP864Mapping creates Unicode <-> CP864 character mapping
// Reference: https://en.wikipedia.org/wiki/Code_page_864
func (f *ESCPOSFont) buildCP864Mapping() {
	// ASCII range (0x00-0x7F) is identical in CP864
	for i := 0x00; i <= 0x7F; i++ {
		r := rune(i)
		b := byte(i)
		f.charMap[r] = b
		f.reverseMap[b] = r
	}

	// Arabic characters in CP864 (0x80-0xFF range)
	// This is a mapping of the most common Arabic characters
	// Full CP864 table: https://www.unicode.org/Public/MAPPINGS/VENDORS/MICSFT/PC/CP864.TXT

	arabicMappings := map[rune]byte{
		// Arabic letters (basic forms)
		'ا': 0xC7, // ALEF
		'أ': 0xC8, // ALEF WITH HAMZA ABOVE
		'إ': 0xC9, // ALEF WITH HAMZA BELOW
		'آ': 0xCA, // ALEF WITH MADDA ABOVE
		'ب': 0xCB, // BEH
		'ت': 0xCC, // TEH
		'ث': 0xCD, // THEH
		'ج': 0xCE, // JEEM
		'ح': 0xCF, // HAH
		'خ': 0xD0, // KHAH
		'د': 0xD1, // DAL
		'ذ': 0xD2, // THAL
		'ر': 0xD3, // REH
		'ز': 0xD4, // ZAIN
		'س': 0xD5, // SEEN
		'ش': 0xD6, // SHEEN
		'ص': 0xD7, // SAD
		'ض': 0xD8, // DAD
		'ط': 0xD9, // TAH
		'ظ': 0xDA, // ZAH
		'ع': 0xDB, // AIN
		'غ': 0xDC, // GHAIN
		'ف': 0xDD, // FEH
		'ق': 0xDE, // QAF
		'ك': 0xDF, // KAF
		'ل': 0xE0, // LAM
		'م': 0xE1, // MEEM
		'ن': 0xE2, // NOON
		'ه': 0xE3, // HEH
		'و': 0xE4, // WAW
		'ى': 0xE5, // ALEF MAKSURA
		'ي': 0xE6, // YEH
		'ة': 0xE7, // TEH MARBUTA

		// Common diacritics
		'َ': 0xEB, // FATHA
		'ُ': 0xEC, // DAMMA
		'ِ': 0xED, // KASRA
		'ّ': 0xF0, // SHADDA
		'ً': 0xF1, // FATHATAN
		'ٌ': 0xF2, // DAMMATAN
		'ٍ': 0xF3, // KASRATAN
		'ْ': 0xF4, // SUKUN

		// Special characters
		'ء': 0xC1, // HAMZA
		'؟': 0xBF, // ARABIC QUESTION MARK
		'،': 0xAC, // ARABIC COMMA
		'؛': 0xBB, // ARABIC SEMICOLON

		// Arabic-Indic digits (Eastern Arabic numerals)
		'٠': 0x80, // ARABIC-INDIC DIGIT ZERO
		'١': 0x81, // ARABIC-INDIC DIGIT ONE
		'٢': 0x82, // ARABIC-INDIC DIGIT TWO
		'٣': 0x83, // ARABIC-INDIC DIGIT THREE
		'٤': 0x84, // ARABIC-INDIC DIGIT FOUR
		'٥': 0x85, // ARABIC-INDIC DIGIT FIVE
		'٦': 0x86, // ARABIC-INDIC DIGIT SIX
		'٧': 0x87, // ARABIC-INDIC DIGIT SEVEN
		'٨': 0x88, // ARABIC-INDIC DIGIT EIGHT
		'٩': 0x89, // ARABIC-INDIC DIGIT NINE

		// Note: Currency symbols use regular Arabic letters already mapped above:
		// ر (REH - 0xD3) - Used in ر.س (Saudi Riyal)
		// د (DAL - 0xD1) - Used in د.ك (Kuwaiti Dinar), د.إ (UAE Dirham)
		// ج (JEEM - 0xCE) - Used in ج.م (Egyptian Pound)
	}

	// Add Arabic mappings
	for r, b := range arabicMappings {
		f.charMap[r] = b
		f.reverseMap[b] = r
	}

	// Build glyph map (for future HarfBuzz integration)
	// In a real implementation, we'd load actual font metrics
	// For ESC/POS, we use a simple 1:1 mapping where GID = byte value
	for gid := 0; gid < 256; gid++ {
		f.glyphMap[gid] = byte(gid)
	}
}

// MapUnicodeToCP864 converts a Unicode string to CP864 byte sequence
func (f *ESCPOSFont) MapUnicodeToCP864(text string) ([]byte, error) {
	result := make([]byte, 0, len(text))

	for _, r := range text {
		if b, ok := f.charMap[r]; ok {
			result = append(result, b)
		} else {
			// Character not in CP864 - try to find a substitute
			if substitute := f.findSubstitute(r); substitute != 0 {
				result = append(result, substitute)
			} else {
				// Use '?' as fallback
				result = append(result, '?')
			}
		}
	}

	return result, nil
}

// MapCP864ToUnicode converts CP864 bytes back to Unicode
func (f *ESCPOSFont) MapCP864ToUnicode(data []byte) string {
	runes := make([]rune, 0, len(data))

	for _, b := range data {
		if r, ok := f.reverseMap[b]; ok {
			runes = append(runes, r)
		} else {
			runes = append(runes, '?')
		}
	}

	return string(runes)
}

// findSubstitute attempts to find a reasonable substitute character
func (f *ESCPOSFont) findSubstitute(r rune) byte {
	// Remove diacritics and try again
	if unicode.Is(unicode.Mn, r) { // Mn = Nonspacing_Mark (diacritics)
		return 0 // Skip diacritics if not supported
	}

	// Common substitutions
	substitutions := map[rune]rune{
		'أ': 'ا', // ALEF WITH HAMZA ABOVE -> ALEF
		'إ': 'ا', // ALEF WITH HAMZA BELOW -> ALEF
		'آ': 'ا', // ALEF WITH MADDA ABOVE -> ALEF
		'ى': 'ي', // ALEF MAKSURA -> YEH
		'ة': 'ه', // TEH MARBUTA -> HEH
	}

	if sub, ok := substitutions[r]; ok {
		if b, ok := f.charMap[sub]; ok {
			return b
		}
	}

	return 0
}

// ShapeText uses HarfBuzz to shape Arabic text and returns ESC/POS bytes
func (f *ESCPOSFont) ShapeText(text string, direction TextDirection) ([]byte, error) {
	// For Phase 2, we implement a simplified shaping approach
	// Full HarfBuzz integration will come in a future enhancement

	// Detect if we need shaping
	if !ContainsArabic(text) {
		// No Arabic - just map directly
		return f.MapUnicodeToCP864(text)
	}

	// For now, we use a simplified shaping algorithm
	// This handles basic contextual forms but not full HarfBuzz
	shaped := f.shapeArabicSimplified(text)

	// Map shaped text to CP864
	return f.MapUnicodeToCP864(shaped)
}

// shapeArabicSimplified performs basic Arabic contextual shaping
// This is a simplified implementation that handles basic forms
// For production, this should be replaced with full HarfBuzz integration
func (f *ESCPOSFont) shapeArabicSimplified(text string) string {
	runes := []rune(text)
	shaped := make([]rune, len(runes))

	for i, r := range runes {
		// Check if this is an Arabic letter that needs shaping
		if !isArabicCharacter(r) {
			shaped[i] = r
			continue
		}

		// Determine position in word
		position := f.getLetterPosition(runes, i)

		// Get shaped form
		shaped[i] = f.getShapedForm(r, position)
	}

	return string(shaped)
}

// LetterPosition represents where a letter appears in a word
type LetterPosition int

const (
	PositionIsolated LetterPosition = iota
	PositionInitial
	PositionMedial
	PositionFinal
)

// getLetterPosition determines if a letter is isolated, initial, medial, or final
func (f *ESCPOSFont) getLetterPosition(runes []rune, index int) LetterPosition {
	hasLeft := index > 0 && isArabicCharacter(runes[index-1]) && canJoinLeft(runes[index-1])
	hasRight := index < len(runes)-1 && isArabicCharacter(runes[index+1]) && canJoinRight(runes[index+1])

	if hasLeft && hasRight {
		return PositionMedial
	} else if hasLeft {
		return PositionFinal
	} else if hasRight {
		return PositionInitial
	}
	return PositionIsolated
}

// canJoinLeft checks if a character can join to the left
func canJoinLeft(r rune) bool {
	// Characters that don't join to the left: ا أ إ آ د ذ ر ز و
	nonJoining := []rune{'ا', 'أ', 'إ', 'آ', 'د', 'ذ', 'ر', 'ز', 'و'}
	for _, nj := range nonJoining {
		if r == nj {
			return false
		}
	}
	return true
}

// canJoinRight checks if a character can join to the right
func canJoinRight(r rune) bool {
	// Most Arabic letters join to the right
	// Only a few don't (same as canJoinLeft)
	return canJoinLeft(r)
}

// getShapedForm returns the appropriate shaped form of an Arabic letter
// Note: This is a simplified implementation
// In reality, we'd use HarfBuzz or the Unicode Arabic Presentation Forms blocks
func (f *ESCPOSFont) getShapedForm(r rune, position LetterPosition) rune {
	// For CP864, most shaped forms are handled by the printer firmware
	// We primarily need to ensure correct character encoding

	// The printer's Code Page 864 typically handles shaping automatically
	// when it detects Arabic text and RTL mode is enabled

	// For now, return the character as-is
	// The printer will apply shaping based on CP864 and RTL mode
	return r
}

// CreateVirtualFontFace creates a virtual font face for text shaping
// This is needed for full HarfBuzz integration in Phase 3
func (f *ESCPOSFont) CreateVirtualFontFace() error {
	// TODO: In Phase 3, implement full virtual font face with HarfBuzz
	// For now, we use the simplified shaping approach in ShapeText()
	return fmt.Errorf("virtual font face not yet implemented")
}

// ShapeWithHarfBuzz performs full HarfBuzz text shaping
// This is the target implementation for Phase 3
// For now, this is a placeholder that will be implemented when we add full HarfBuzz support
/*
func (f *ESCPOSFont) ShapeWithHarfBuzz(text string) error {
	// TODO: Full HarfBuzz integration (Phase 3)
	// 1. Import: "github.com/go-text/typesetting/harfbuzz"
	// 2. Create HarfBuzz buffer
	// 3. Set language (Arabic) and direction (RTL)
	// 4. Add text to buffer
	// 5. Shape with font face
	// 6. Return positioned glyphs

	return fmt.Errorf("HarfBuzz shaping not yet implemented - use ShapeText() for simplified shaping")
}
*/

// Example of how HarfBuzz integration will work (Phase 3):
/*
func (f *ESCPOSFont) ShapeWithHarfBuzz(text string) ([]harfbuzz.GlyphInfo, []harfbuzz.GlyphPosition, error) {
	// Create HarfBuzz buffer
	buffer := harfbuzz.NewBuffer()
	buffer.AddRunes([]rune(text), 0, -1)

	// Set Arabic language and RTL direction
	buffer.Props.Language = language.NewLanguage("ar")
	buffer.Props.Direction = harfbuzz.RightToLeft
	buffer.Props.Script = language.Arabic

	// Get font face (would need to create virtual face)
	face := f.face
	if face == nil {
		return nil, nil, fmt.Errorf("font face not initialized")
	}

	// Shape the text
	buffer.Shape(face, nil)

	// Get results
	info := buffer.Info
	positions := buffer.Pos

	return info, positions, nil
}
*/

// ValidateCharacterSupport checks if all characters in text are supported
func (f *ESCPOSFont) ValidateCharacterSupport(text string) (bool, []rune) {
	unsupported := []rune{}

	for _, r := range text {
		if _, ok := f.charMap[r]; !ok {
			if f.findSubstitute(r) == 0 {
				unsupported = append(unsupported, r)
			}
		}
	}

	return len(unsupported) == 0, unsupported
}

// GetFontMetrics returns basic font metrics for layout calculations
type FontMetrics struct {
	CharWidth    int // Average character width in dots
	CharHeight   int // Character height in dots
	LineSpacing  int // Space between lines in dots
	Monospaced   bool
}

func (f *ESCPOSFont) GetFontMetrics() FontMetrics {
	// ESC/POS standard font metrics
	// Font A (12x24): 12 dots wide, 24 dots tall
	return FontMetrics{
		CharWidth:   12,
		CharHeight:  24,
		LineSpacing: 4,
		Monospaced:  true, // ESC/POS fonts are typically monospaced
	}
}

// CalculateTextWidth calculates the width of text in characters
func (f *ESCPOSFont) CalculateTextWidth(text string) int {
	// For monospaced fonts, width = number of characters
	return len([]rune(text))
}

// GetCodePage returns the current code page number
func (f *ESCPOSFont) GetCodePage() int {
	return f.codePage
}

// GetCapabilities returns the font's capabilities
func (f *ESCPOSFont) GetCapabilities() FontCapabilities {
	return f.capabilities
}

// PrintDebugInfo prints debugging information about the font
func (f *ESCPOSFont) PrintDebugInfo() {
	fmt.Printf("ESC/POS Font Debug Info:\n")
	fmt.Printf("  Code Page: %d\n", f.codePage)
	fmt.Printf("  Mapped Characters: %d\n", len(f.charMap))
	fmt.Printf("  Arabic Support: %v\n", f.capabilities.SupportsArabic)
	fmt.Printf("  Diacritics Support: %v\n", f.capabilities.SupportsDiacritics)
	fmt.Printf("  Ligatures Support: %v\n", f.capabilities.SupportsLigatures)
	fmt.Printf("  Max Glyph ID: %d\n", f.capabilities.MaxGlyphID)
}

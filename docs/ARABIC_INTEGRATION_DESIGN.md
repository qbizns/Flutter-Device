# Arabic ESC/POS Integration Design

**Date**: 2025-11-13
**Phase**: 1.1.3 - Arabic Text Shaping Integration
**Status**: In Progress
**Estimated Completion**: 2-3 weeks

---

## Executive Summary

This document outlines the technical design for integrating Arabic text rendering into the ESC/POS printer driver. The implementation will use `go-text/typesetting` for text shaping and the Unicode BiDi algorithm for mixed LTR/RTL text handling.

**Key Requirements**:
- Render Arabic text with correct glyph shaping (initial/medial/final/isolated forms)
- Support right-to-left (RTL) text direction
- Handle mixed Arabic/English (BiDi) text on same line
- Work with ESC/POS printer built-in fonts (Code Page 864 for Arabic)
- Line wrapping that respects Arabic character boundaries
- Performance target: <200ms print job latency

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Print Document                           │
│          (Lines with mixed Arabic/English text)              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│              ESC/POS Renderer (renderer.go)                  │
│                                                              │
│  ┌──────────────────────────────────────────────────┐      │
│  │  For each line:                                   │      │
│  │  1. Detect if line contains Arabic                │      │
│  │  2. If Arabic → Route to ArabicShaper             │      │
│  │  3. If English → Use existing logic               │      │
│  │  4. If mixed → Use BiDi algorithm + ArabicShaper  │      │
│  └──────────────┬───────────────────────────────────┘      │
│                 │                                            │
└─────────────────┼────────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│           Arabic Shaper (arabic.go - NEW)                    │
│                                                              │
│  ┌──────────────────────────────────────────────────┐      │
│  │  TextRun → HarfBuzz Shaping → Shaped Glyphs      │      │
│  │                                                   │      │
│  │  1. Analyze text direction (detectArabic)        │      │
│  │  2. Split into runs (BiDi algorithm)             │      │
│  │  3. Shape each run (HarfBuzz)                    │      │
│  │  4. Reorder runs for display                     │      │
│  │  5. Map glyphs to ESC/POS character codes        │      │
│  └──────────────┬───────────────────────────────────┘      │
│                 │                                            │
└─────────────────┼────────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│         ESC/POS Command Generator                            │
│                                                              │
│  Shaped Glyphs → ESC/POS Commands (bytes)                   │
│                                                              │
│  - Set Code Page 864 (Arabic)                               │
│  - Set text direction (RTL/LTR)                             │
│  - Output character codes                                   │
│  - Handle alignment (right-align for Arabic amounts)        │
└─────────────────────────────────────────────────────────────┘
```

---

## Component Design

### 1. Arabic Shaper Module (`internal/drivers/printer_escpos/arabic.go`)

**Purpose**: Core Arabic text shaping logic using HarfBuzz

**Key Functions**:
```go
// ShapeArabicText shapes Arabic text using HarfBuzz
func ShapeArabicText(text string, font Font) ([]ShapedGlyph, error)

// DetectTextDirection determines if text is LTR, RTL, or mixed
func DetectTextDirection(text string) TextDirection

// SplitIntoRuns splits mixed BiDi text into directional runs
func SplitIntoRuns(text string) []TextRun

// MapGlyphToESCPOS maps shaped glyph to ESC/POS character code
func MapGlyphToESCPOS(glyph ShapedGlyph) (byte, error)
```

**Data Structures**:
```go
type TextDirection int
const (
    DirectionLTR TextDirection = iota
    DirectionRTL
    DirectionMixed
)

type TextRun struct {
    Text      string
    Direction TextDirection
    Start     int
    End       int
}

type ShapedGlyph struct {
    GlyphID    uint32
    Cluster    int
    XAdvance   int32
    YAdvance   int32
    XOffset    int32
    YOffset    int32
    CodePoint  rune
}
```

---

### 2. Font Mapping (`internal/drivers/printer_escpos/arabic_font.go`)

**Challenge**: ESC/POS printers don't load external fonts - they have built-in character sets

**Solution**: Create virtual font mapping for Code Page 864 (Arabic)

**Approach**:
```go
// ESCPOSArabicFont represents the printer's built-in Arabic font
type ESCPOSArabicFont struct {
    codePage int
    glyphMap map[rune]byte  // Unicode → ESC/POS character code
}

// GetCharacterCode returns ESC/POS character code for Unicode codepoint
func (f *ESCPOSArabicFont) GetCharacterCode(r rune) (byte, bool)

// SupportsArabic checks if printer supports Arabic characters
func (f *ESCPOSArabicFont) SupportsArabic() bool
```

**Code Page 864 Mapping**:
- ESC `t` 0x1E - Select Code Page 864 (Arabic)
- Map Arabic Unicode (U+0600-U+06FF) to printer character codes (0x80-0xFF)
- Handle presentation forms (U+FE70-U+FEFF) for shaped glyphs

---

### 3. BiDi Algorithm (`internal/drivers/printer_escpos/bidi.go`)

**Purpose**: Unicode Bidirectional Algorithm implementation

**Reference**: UAX #9 (Unicode Standard Annex #9)

**Simplified Implementation**:
```go
// ReorderForDisplay reorders mixed LTR/RTL text for display
func ReorderForDisplay(runs []TextRun) []TextRun

// AnalyzeText determines text levels and directions
func AnalyzeText(text string) []int  // Returns bidi levels per character

// Example:
// Input:  "Coca-Cola كولا"
// Runs:   [Run(LTR, "Coca-Cola"), Run(RTL, "كولا")]
// Output: [Run(LTR, "Coca-Cola"), Run(RTL, "كولا")] (already correct order)
```

**For MVP**: Use simplified algorithm that detects script changes
**For Production**: Consider using `golang.org/x/text/unicode/bidi`

---

### 4. Renderer Integration (`internal/drivers/printer_escpos/renderer.go`)

**Existing Code** (line 89-120):
```go
// Current implementation - no Arabic support
func (r *Renderer) renderText(text string, style TextStyle) []byte {
    // TODO: Arabic text shaping
    // Currently assumes LTR
    return escpos.Commands...
}
```

**New Implementation**:
```go
func (r *Renderer) renderText(text string, style TextStyle) []byte {
    // 1. Detect if text contains Arabic
    if containsArabic(text) {
        return r.renderArabicText(text, style)
    }

    // 2. Use existing logic for English/Latin
    return r.renderLatinText(text, style)
}

func (r *Renderer) renderArabicText(text string, style TextStyle) []byte {
    var buf bytes.Buffer

    // 1. Set Code Page 864 (Arabic)
    buf.Write([]byte{0x1B, 't', 0x1E})

    // 2. Detect text direction
    direction := DetectTextDirection(text)

    // 3. Handle based on direction
    switch direction {
    case DirectionRTL:
        // Pure Arabic text
        shaped, err := ShapeArabicText(text, r.arabicFont)
        if err != nil {
            return r.renderLatinText(text, style)  // Fallback
        }
        buf.Write(r.glyphsToESCPOS(shaped, style))

    case DirectionMixed:
        // Mixed Arabic/English - use BiDi
        runs := SplitIntoRuns(text)
        runs = ReorderForDisplay(runs)

        for _, run := range runs {
            if run.Direction == DirectionRTL {
                shaped, _ := ShapeArabicText(run.Text, r.arabicFont)
                buf.Write(r.glyphsToESCPOS(shaped, style))
            } else {
                buf.Write(r.renderLatinText(run.Text, style))
            }
        }

    default:
        // LTR (shouldn't happen if containsArabic is true)
        buf.Write(r.renderLatinText(text, style))
    }

    // 4. Reset to default Code Page
    buf.Write([]byte{0x1B, 't', 0x00})

    return buf.Bytes()
}
```

---

### 5. Line Wrapping (`internal/drivers/printer_escpos/wrap.go`)

**Challenge**: Arabic text wraps right-to-left and respects word boundaries

**Current Implementation**: Character-based wrapping (doesn't understand Arabic)

**New Implementation**:
```go
// WrapArabicLine wraps Arabic text at word boundaries
func WrapArabicLine(text string, maxWidth int) []string {
    // 1. Split by spaces (word boundaries)
    words := strings.Split(text, " ")

    // 2. Build lines RTL
    var lines []string
    var currentLine []string
    currentWidth := 0

    for _, word := range words {
        wordWidth := len([]rune(word))  // Simplified - should measure glyphs

        if currentWidth + wordWidth > maxWidth {
            // Word doesn't fit, start new line
            lines = append(lines, strings.Join(currentLine, " "))
            currentLine = []string{word}
            currentWidth = wordWidth
        } else {
            // Word fits
            currentLine = append(currentLine, word)
            currentWidth += wordWidth + 1  // +1 for space
        }
    }

    // Add last line
    if len(currentLine) > 0 {
        lines = append(lines, strings.Join(currentLine, " "))
    }

    return lines
}
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)

**Tasks**:
- [ ] Create `arabic.go` module
- [ ] Implement `containsArabic()` (already done in prototype)
- [ ] Implement `DetectTextDirection()`
- [ ] Create ESC/POS font mapping structure
- [ ] Add Code Page 864 character map

**Deliverables**:
- `internal/drivers/printer_escpos/arabic.go`
- `internal/drivers/printer_escpos/arabic_font.go`
- Unit tests for detection and direction analysis

---

### Phase 2: Text Shaping (Week 1-2)

**Tasks**:
- [ ] Integrate HarfBuzz buffer creation
- [ ] Implement `ShapeArabicText()` with go-text/typesetting
- [ ] Create glyph-to-ESC/POS mapping
- [ ] Handle shaped glyph output
- [ ] Test with simple Arabic words

**Deliverables**:
- Working text shaping for simple Arabic
- Tests with "مرحباً", "شكراً", etc.
- ESC/POS command generation

**Challenges**:
- Font file handling (ESC/POS doesn't use external fonts)
- Glyph ID to character code mapping

---

### Phase 3: BiDi Algorithm (Week 2)

**Tasks**:
- [ ] Create `bidi.go` module
- [ ] Implement `SplitIntoRuns()`
- [ ] Implement `ReorderForDisplay()`
- [ ] Test with mixed Arabic/English
- [ ] Integration with renderer

**Deliverables**:
- BiDi algorithm implementation
- Tests with "Coca-Cola كولا", "السعر: 25.50 ر.س"
- Mixed text rendering

---

### Phase 4: Integration & Polish (Week 2-3)

**Tasks**:
- [ ] Modify `renderer.go` to route Arabic text
- [ ] Implement line wrapping for Arabic
- [ ] Add alignment handling (right-align for Arabic)
- [ ] Performance optimization
- [ ] Comprehensive integration tests

**Deliverables**:
- Fully integrated Arabic support in ESC/POS renderer
- Line wrapping working
- All alignment options (left, center, right) working with Arabic

---

### Phase 5: Testing & Validation (Week 3)

**Tasks**:
- [ ] Test with virtual printer
- [ ] Test with real Epson printer (after hardware arrives)
- [ ] Test with real Chinese printer
- [ ] Test with Star printer
- [ ] Validate with native Arabic speakers
- [ ] Performance benchmarking

**Deliverables**:
- Test results documented
- Photos of printed Arabic receipts
- Performance metrics (<200ms target)

---

## ESC/POS Commands Reference

### Code Page Selection
```
ESC t n
0x1B 0x74 n

Code Pages:
- 0x00: PC437 (USA, Standard Europe)
- 0x1E: PC864 (Arabic)
- 0x1F: PC1256 (Arabic - Windows)
```

### Arabic Text Direction
```
ESC { n
0x1B 0x7B n

n = 0: Left to right (default)
n = 1: Right to left
```

### Character Code Output
```
After setting Code Page 864:
- Send character codes 0x80-0xFF for Arabic characters
- Printer maps these to Arabic glyphs
```

---

## Testing Strategy

### Unit Tests
- `TestDetectArabic()` - Character detection ✅ (done in prototype)
- `TestDetectTextDirection()` - LTR/RTL/Mixed detection
- `TestShapeArabicText()` - Text shaping with HarfBuzz
- `TestMapGlyphToESCPOS()` - Glyph mapping
- `TestSplitIntoRuns()` - BiDi run detection
- `TestReorderForDisplay()` - BiDi reordering
- `TestWrapArabicLine()` - Line wrapping

### Integration Tests
- `TestRenderSimpleArabic()` - "مرحباً"
- `TestRenderMixedText()` - "Coca-Cola كولا"
- `TestRenderArabicWithNumbers()` - "السعر: 25.50 ر.س"
- `TestRenderLongArabicText()` - Line wrapping
- `TestRenderArabicAlignment()` - Left/Center/Right alignment

### Real Printer Tests
- Epson TM-T88VI with Code Page 864
- Star TSP143
- Cheap Chinese printer
- Validation with Arabic speakers

---

## Performance Targets

| Operation | Target | Notes |
|-----------|--------|-------|
| Text shaping | <50ms | HarfBuzz shaping per line |
| BiDi analysis | <10ms | Run detection and reordering |
| ESC/POS generation | <20ms | Command byte generation |
| Total per line | <80ms | Leaves room for network/USB overhead |
| Full receipt | <200ms | Target from acceptance checklist |

---

## Risk Mitigation

### Risk 1: Font Mapping Complexity
**Issue**: ESC/POS printers don't use external fonts
**Mitigation**: Create virtual font that maps Unicode to Code Page 864
**Fallback**: Use printer's built-in shaping (if available)

### Risk 2: Printer Compatibility
**Issue**: Not all printers support Code Page 864
**Mitigation**: Detect printer capabilities, document requirements
**Fallback**: Error message if Arabic not supported

### Risk 3: Performance
**Issue**: HarfBuzz shaping might be slow
**Mitigation**: Cache shaped results, optimize hot paths
**Fallback**: Disable shaping for performance mode

### Risk 4: Glyph Rendering Issues
**Issue**: Shaped glyphs might not match printer expectations
**Mitigation**: Test with multiple printer models early
**Fallback**: Adjust mapping based on printer feedback

---

## Success Criteria

**Phase 1.1.3 is COMPLETE when**:
- ✅ Arabic text renders with correct glyph shapes
- ✅ RTL text flows right-to-left
- ✅ Mixed Arabic/English renders correctly
- ✅ Line wrapping respects Arabic word boundaries
- ✅ All unit tests passing
- ✅ Integration tests passing
- ✅ Performance <200ms per receipt
- ✅ Tested on virtual printer
- ⏭️ Ready for real printer testing (waits for hardware)

---

## Dependencies

**Go Modules**:
- `github.com/go-text/typesetting` v0.3.0 (already added)
- Possibly `golang.org/x/text/unicode/bidi` (if needed)

**Hardware**:
- ESC/POS printers with Code Page 864 support
- See docs/HARDWARE_MATRIX.md for recommended models

**Documentation**:
- Unicode UAX #9 (BiDi algorithm)
- ESC/POS command reference
- Code Page 864 character mappings

---

## Next Steps

1. **Today**: Create `arabic.go` foundation module
2. **Week 1**: Implement text shaping with HarfBuzz
3. **Week 2**: Add BiDi algorithm and integration
4. **Week 3**: Testing and optimization
5. **After Hardware Arrival**: Real printer validation

---

## References

- go-text/typesetting: https://pkg.go.dev/github.com/go-text/typesetting
- Unicode BiDi: https://unicode.org/reports/tr9/
- ESC/POS Reference: https://reference.epson-biz.com/
- Code Page 864: https://en.wikipedia.org/wiki/Code_page_864
- HarfBuzz Manual: https://harfbuzz.github.io/
- Arabic Unicode: https://www.unicode.org/charts/PDF/U0600.pdf

---

**Document Version**: 1.0
**Last Updated**: 2025-11-13
**Status**: Design Complete - Ready for Implementation

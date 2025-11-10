package printer_zpl

import (
	"bytes"
	"fmt"
	"strings"
)

// LabelBuilder helps construct ZPL labels
type LabelBuilder struct {
	buf       bytes.Buffer
	dpi       int
	width     int
	height    int
	hasStart  bool
	fieldOpen bool
}

// NewLabelBuilder creates a new ZPL label builder
func NewLabelBuilder(dpi, width, height int) *LabelBuilder {
	return &LabelBuilder{
		dpi:    dpi,
		width:  width,
		height: height,
	}
}

// Start begins a new label
func (b *LabelBuilder) Start() *LabelBuilder {
	if b.hasStart {
		return b
	}
	b.buf.WriteString("^XA\n")
	b.buf.WriteString("^LH0,0\n") // Label home
	b.hasStart = true
	return b
}

// End finalizes the label
func (b *LabelBuilder) End() *LabelBuilder {
	if b.fieldOpen {
		b.buf.WriteString("^FS\n")
		b.fieldOpen = false
	}
	b.buf.WriteString("^XZ\n")
	return b
}

// SetOrigin sets the field origin position
func (b *LabelBuilder) SetOrigin(x, y int) *LabelBuilder {
	if b.fieldOpen {
		b.buf.WriteString("^FS\n")
		b.fieldOpen = false
	}
	b.buf.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))
	return b
}

// Text adds text at current position
func (b *LabelBuilder) Text(text string, fontSize int) *LabelBuilder {
	if fontSize == 0 {
		fontSize = 30
	}
	b.buf.WriteString(fmt.Sprintf("^A0N,%d,%d\n", fontSize, fontSize))
	b.buf.WriteString("^FD")
	b.buf.WriteString(text)
	b.buf.WriteString("^FS\n")
	return b
}

// TextBold adds bold text
func (b *LabelBuilder) TextBold(text string, fontSize int) *LabelBuilder {
	if fontSize == 0 {
		fontSize = 30
	}
	b.buf.WriteString(fmt.Sprintf("^A0B,%d,%d\n", fontSize, fontSize))
	b.buf.WriteString("^FD")
	b.buf.WriteString(text)
	b.buf.WriteString("^FS\n")
	return b
}

// Barcode adds a barcode
func (b *LabelBuilder) Barcode(data, barcodeType string, height int) *LabelBuilder {
	if height == 0 {
		height = 100
	}

	switch strings.ToUpper(barcodeType) {
	case "CODE128":
		b.buf.WriteString(fmt.Sprintf("^BCN,%d,Y,N,N\n", height))
	case "CODE39":
		b.buf.WriteString(fmt.Sprintf("^B3N,N,%d,Y,N\n", height))
	case "EAN13":
		b.buf.WriteString(fmt.Sprintf("^BEN,%d,Y,N\n", height))
	case "UPCA":
		b.buf.WriteString(fmt.Sprintf("^BUN,%d,Y,N,N\n", height))
	case "QR":
		b.buf.WriteString("^BQN,2,5\n")
	default:
		// Default to Code 128
		b.buf.WriteString(fmt.Sprintf("^BCN,%d,Y,N,N\n", height))
	}

	b.buf.WriteString("^FD")
	b.buf.WriteString(data)
	b.buf.WriteString("^FS\n")
	return b
}

// QRCode adds a QR code
func (b *LabelBuilder) QRCode(data string, size int) *LabelBuilder {
	if size == 0 {
		size = 5
	}
	b.buf.WriteString(fmt.Sprintf("^BQN,2,%d\n", size))
	b.buf.WriteString("^FD")
	b.buf.WriteString(data)
	b.buf.WriteString("^FS\n")
	return b
}

// Line draws a horizontal or vertical line
func (b *LabelBuilder) Line(length, thickness int, vertical bool) *LabelBuilder {
	if vertical {
		b.buf.WriteString(fmt.Sprintf("^GB%d,%d,%d,B,0\n", thickness, length, thickness))
	} else {
		b.buf.WriteString(fmt.Sprintf("^GB%d,%d,%d,B,0\n", length, thickness, thickness))
	}
	return b
}

// Box draws a rectangle
func (b *LabelBuilder) Box(width, height, thickness int) *LabelBuilder {
	b.buf.WriteString(fmt.Sprintf("^GB%d,%d,%d,B,0\n", width, height, thickness))
	return b
}

// Image adds a graphic field (for logos, etc.)
func (b *LabelBuilder) Image(data string) *LabelBuilder {
	b.buf.WriteString("^GFA,")
	b.buf.WriteString(data)
	b.buf.WriteString("\n")
	return b
}

// SetQuantity sets the number of labels to print
func (b *LabelBuilder) SetQuantity(qty int) *LabelBuilder {
	b.buf.WriteString(fmt.Sprintf("^PQ%d,0,0,Y\n", qty))
	return b
}

// SetDarkness adjusts print darkness (0-30)
func (b *LabelBuilder) SetDarkness(level int) *LabelBuilder {
	if level < 0 {
		level = 0
	}
	if level > 30 {
		level = 30
	}
	b.buf.WriteString(fmt.Sprintf("~SD%d\n", level))
	return b
}

// SetSpeed sets print speed (2-6, slower = better quality)
func (b *LabelBuilder) SetSpeed(speed int) *LabelBuilder {
	if speed < 2 {
		speed = 2
	}
	if speed > 6 {
		speed = 6
	}
	b.buf.WriteString(fmt.Sprintf("^PR%d\n", speed))
	return b
}

// Reverse inverts black and white
func (b *LabelBuilder) Reverse() *LabelBuilder {
	b.buf.WriteString("^LRY\n")
	return b
}

// Build returns the ZPL string
func (b *LabelBuilder) Build() string {
	if !b.hasStart {
		b.Start()
	}
	// Make a copy before ending
	result := b.buf.String()
	if !strings.HasSuffix(result, "^XZ\n") {
		b.End()
		result = b.buf.String()
	}
	return result
}

// Reset clears the builder for reuse
func (b *LabelBuilder) Reset() *LabelBuilder {
	b.buf.Reset()
	b.hasStart = false
	b.fieldOpen = false
	return b
}

// Common label templates

// BuildShippingLabel creates a shipping label
func BuildShippingLabel(dpi int, tracking, recipient, sender, address string) string {
	builder := NewLabelBuilder(dpi, 800, 1200)

	return builder.Start().
		SetOrigin(50, 50).
		TextBold("SHIPPING LABEL", 40).
		SetOrigin(50, 120).
		Text("Tracking: "+tracking, 30).
		SetOrigin(50, 180).
		Line(700, 3, false).
		SetOrigin(50, 200).
		TextBold("TO:", 35).
		SetOrigin(50, 250).
		Text(recipient, 30).
		SetOrigin(50, 290).
		Text(address, 25).
		SetOrigin(50, 400).
		Line(700, 3, false).
		SetOrigin(50, 420).
		TextBold("FROM:", 35).
		SetOrigin(50, 470).
		Text(sender, 30).
		SetOrigin(200, 600).
		Barcode(tracking, "CODE128", 120).
		Build()
}

// BuildProductLabel creates a product label
func BuildProductLabel(dpi int, name, sku, price, barcode string) string {
	builder := NewLabelBuilder(dpi, 800, 600)

	return builder.Start().
		SetOrigin(50, 30).
		TextBold(name, 35).
		SetOrigin(50, 80).
		Text("SKU: "+sku, 25).
		SetOrigin(50, 120).
		TextBold("$"+price, 40).
		SetOrigin(150, 200).
		Barcode(barcode, "CODE128", 100).
		Build()
}

// BuildNameBadge creates a name badge
func BuildNameBadge(dpi int, name, company, title string) string {
	builder := NewLabelBuilder(dpi, 800, 600)

	return builder.Start().
		SetOrigin(50, 50).
		Box(700, 500, 5).
		SetOrigin(200, 150).
		TextBold(name, 50).
		SetOrigin(200, 220).
		Text(title, 30).
		SetOrigin(200, 270).
		Text(company, 25).
		Build()
}

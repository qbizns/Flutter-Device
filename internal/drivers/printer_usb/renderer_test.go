package printer_usb

import (
	"bytes"
	"testing"

	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
)

func TestRenderDocument_Initialize(t *testing.T) {
	doc := &printer.PrintDocument{
		Sections: []printer.Section{},
		Options:  printer.PrintOptions{},
	}

	result, err := renderDocument(doc)
	if err != nil {
		t.Fatalf("renderDocument failed: %v", err)
	}

	// Check for initialization sequence
	// ESC @ - Initialize printer
	initSeq := []byte{0x1B, 0x40}
	if !bytes.Contains(result, initSeq) {
		t.Error("Missing initialization sequence (ESC @)")
	}

	// Check for charset selection
	// ESC t 0 - Set charset to CP437
	charsetSeq := []byte{0x1B, 0x74, 0x00}
	if !bytes.Contains(result, charsetSeq) {
		t.Error("Missing charset selection (ESC t 0)")
	}
}

func TestRenderDocument_Cut(t *testing.T) {
	doc := &printer.PrintDocument{
		Sections: []printer.Section{},
		Options: printer.PrintOptions{
			Cut: true,
		},
	}

	result, err := renderDocument(doc)
	if err != nil {
		t.Fatalf("renderDocument failed: %v", err)
	}

	// Check for cut command
	// GS V 0 - Full cut
	cutSeq := []byte{0x1D, 0x56, 0x00}
	if !bytes.Contains(result, cutSeq) {
		t.Error("Missing cut command (GS V 0)")
	}
}

func TestRenderDocument_DrawerPulse(t *testing.T) {
	doc := &printer.PrintDocument{
		Sections: []printer.Section{},
		Options: printer.PrintOptions{
			DrawerPulse: true,
		},
	}

	result, err := renderDocument(doc)
	if err != nil {
		t.Fatalf("renderDocument failed: %v", err)
	}

	// Check for drawer pulse command
	// ESC p 0 50 50
	drawerSeq := []byte{0x1B, 0x70, 0x00, 0x32, 0x32}
	if !bytes.Contains(result, drawerSeq) {
		t.Error("Missing drawer pulse command (ESC p 0 50 50)")
	}
}

func TestRenderLine_Alignment(t *testing.T) {
	tests := []struct {
		name      string
		alignment printer.Alignment
		expected  []byte
	}{
		{"left", printer.AlignLeft, []byte{0x1B, 0x61, 0x00}},
		{"center", printer.AlignCenter, []byte{0x1B, 0x61, 0x01}},
		{"right", printer.AlignRight, []byte{0x1B, 0x61, 0x02}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			line := printer.Line{
				Alignment: tt.alignment,
				Runs: []printer.Run{
					{Text: "Test", Style: printer.TextStyle{}},
				},
			}

			renderLine(buf, line)

			if !bytes.Contains(buf.Bytes(), tt.expected) {
				t.Errorf("Missing alignment command for %s: %v", tt.name, tt.expected)
			}
		})
	}
}

func TestRenderRun_Bold(t *testing.T) {
	buf := &bytes.Buffer{}
	run := printer.Run{
		Text: "Bold Text",
		Style: printer.TextStyle{
			Bold: true,
		},
	}

	renderRun(buf, run)

	// Check for bold on command
	boldOn := []byte{0x1B, 0x45, 0x01}
	if !bytes.Contains(buf.Bytes(), boldOn) {
		t.Error("Missing bold on command (ESC E 1)")
	}

	// Check that text is present
	if !bytes.Contains(buf.Bytes(), []byte("Bold Text")) {
		t.Error("Missing text content")
	}

	// Check for bold off command (reset)
	boldOff := []byte{0x1B, 0x45, 0x00}
	if !bytes.Contains(buf.Bytes(), boldOff) {
		t.Error("Missing bold off command (ESC E 0)")
	}
}

func TestRenderRun_Underline(t *testing.T) {
	buf := &bytes.Buffer{}
	run := printer.Run{
		Text: "Underlined",
		Style: printer.TextStyle{
			Underline: true,
		},
	}

	renderRun(buf, run)

	// Check for underline on command
	underlineOn := []byte{0x1B, 0x2D, 0x01}
	if !bytes.Contains(buf.Bytes(), underlineOn) {
		t.Error("Missing underline on command (ESC - 1)")
	}
}

func TestRenderRun_DoubleSize(t *testing.T) {
	buf := &bytes.Buffer{}
	run := printer.Run{
		Text: "Large",
		Style: printer.TextStyle{
			DoubleWidth:  true,
			DoubleHeight: true,
		},
	}

	renderRun(buf, run)

	// Check for size command (GS ! n where n = 0x30 for both width and height)
	// DoubleWidth: 0x20, DoubleHeight: 0x10, Combined: 0x30
	sizeCmd := []byte{0x1D, 0x21, 0x30}
	if !bytes.Contains(buf.Bytes(), sizeCmd) {
		t.Error("Missing double size command (GS ! 30)")
	}
}

func TestRenderBarcode_Code128(t *testing.T) {
	buf := &bytes.Buffer{}
	barcode := printer.Barcode{
		Data:      "123456789",
		Symbology: printer.BarcodeCode128,
		Height:    100,
		Width:     3,
	}

	renderBarcode(buf, barcode)

	result := buf.Bytes()

	// Check for height command (GS h n)
	heightCmd := []byte{0x1D, 0x68, 100}
	if !bytes.Contains(result, heightCmd) {
		t.Error("Missing barcode height command")
	}

	// Check for width command (GS w n)
	widthCmd := []byte{0x1D, 0x77, 3}
	if !bytes.Contains(result, widthCmd) {
		t.Error("Missing barcode width command")
	}

	// Check for Code128 symbology (73)
	if !bytes.Contains(result, []byte{73}) {
		t.Error("Missing Code128 symbology code")
	}

	// Check for barcode data
	if !bytes.Contains(result, []byte("123456789")) {
		t.Error("Missing barcode data")
	}
}

func TestRenderBarcode_EAN13(t *testing.T) {
	buf := &bytes.Buffer{}
	barcode := printer.Barcode{
		Data:      "5901234123457",
		Symbology: printer.BarcodeEAN13,
	}

	renderBarcode(buf, barcode)

	result := buf.Bytes()

	// Check for EAN13 symbology (67)
	if !bytes.Contains(result, []byte{67}) {
		t.Error("Missing EAN13 symbology code")
	}

	// Check for default height (162)
	heightCmd := []byte{0x1D, 0x68, 162}
	if !bytes.Contains(result, heightCmd) {
		t.Error("Missing default height")
	}
}

func TestRenderQRCode(t *testing.T) {
	buf := &bytes.Buffer{}
	qrcode := printer.QRCode{
		Data: "https://example.com",
		Size: 6,
	}

	renderQRCode(buf, qrcode)

	result := buf.Bytes()

	// Check for QR code commands (GS ( k sequences)
	if !bytes.Contains(result, []byte{0x1D, 0x28, 0x6B}) {
		t.Error("Missing QR code command prefix")
	}

	// Check that data is included
	if !bytes.Contains(result, []byte("https://example.com")) {
		t.Error("Missing QR code data")
	}
}

func TestRenderDocument_CompleteReceipt(t *testing.T) {
	// Build a complete receipt with multiple elements
	doc := &printer.PrintDocument{
		Sections: []printer.Section{
			{
				Elements: []printer.Element{
					{
						Type: printer.ElementTypeLine,
						Line: &printer.Line{
							Alignment: printer.AlignCenter,
							Runs: []printer.Run{
								{
									Text: "RECEIPT",
									Style: printer.TextStyle{
										Bold:         true,
										DoubleWidth:  true,
										DoubleHeight: true,
									},
								},
							},
						},
					},
					{
						Type: printer.ElementTypeLine,
						Line: &printer.Line{
							Alignment: printer.AlignLeft,
							Runs: []printer.Run{
								{Text: "Item 1", Style: printer.TextStyle{}},
								{Text: "  $10.00", Style: printer.TextStyle{}},
							},
						},
					},
					{
						Type: printer.ElementTypeBarcode,
						Barcode: &printer.Barcode{
							Data:      "123456789012",
							Symbology: printer.BarcodeEAN13,
						},
					},
					{
						Type: printer.ElementTypeQRCode,
						QRCode: &printer.QRCode{
							Data: "https://receipt.example.com/12345",
							Size: 6,
						},
					},
				},
			},
		},
		Options: printer.PrintOptions{
			Cut:         true,
			DrawerPulse: false,
		},
	}

	result, err := renderDocument(doc)
	if err != nil {
		t.Fatalf("renderDocument failed: %v", err)
	}

	// Verify initialization
	if !bytes.Contains(result, []byte{0x1B, 0x40}) {
		t.Error("Missing initialization")
	}

	// Verify text content
	if !bytes.Contains(result, []byte("RECEIPT")) {
		t.Error("Missing header text")
	}

	if !bytes.Contains(result, []byte("Item 1")) {
		t.Error("Missing item text")
	}

	// Verify barcode
	if !bytes.Contains(result, []byte("123456789012")) {
		t.Error("Missing barcode data")
	}

	// Verify QR code
	if !bytes.Contains(result, []byte("https://receipt.example.com/12345")) {
		t.Error("Missing QR code data")
	}

	// Verify cut command
	if !bytes.Contains(result, []byte{0x1D, 0x56, 0x00}) {
		t.Error("Missing cut command")
	}

	// Verify no drawer pulse (since DrawerPulse is false)
	drawerSeq := []byte{0x1B, 0x70, 0x00, 0x32, 0x32}
	if bytes.Contains(result, drawerSeq) {
		t.Error("Unexpected drawer pulse command")
	}
}

func TestRenderDocument_EmptyDocument(t *testing.T) {
	doc := &printer.PrintDocument{
		Sections: []printer.Section{},
		Options:  printer.PrintOptions{},
	}

	result, err := renderDocument(doc)
	if err != nil {
		t.Fatalf("renderDocument failed: %v", err)
	}

	// Should at least have initialization
	if len(result) < 10 {
		t.Error("Result too short for empty document")
	}
}

func BenchmarkRenderDocument_SimpleReceipt(b *testing.B) {
	doc := &printer.PrintDocument{
		Sections: []printer.Section{
			{
				Elements: []printer.Element{
					{
						Type: printer.ElementTypeLine,
						Line: &printer.Line{
							Alignment: printer.AlignCenter,
							Runs: []printer.Run{
								{Text: "RECEIPT", Style: printer.TextStyle{Bold: true}},
							},
						},
					},
					{
						Type: printer.ElementTypeLine,
						Line: &printer.Line{
							Alignment: printer.AlignLeft,
							Runs: []printer.Run{
								{Text: "Item 1  $10.00", Style: printer.TextStyle{}},
							},
						},
					},
				},
			},
		},
		Options: printer.PrintOptions{Cut: true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = renderDocument(doc)
	}
}

func BenchmarkRenderBarcode(b *testing.B) {
	buf := &bytes.Buffer{}
	barcode := printer.Barcode{
		Data:      "5901234123457",
		Symbology: printer.BarcodeEAN13,
		Height:    162,
		Width:     3,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		renderBarcode(buf, barcode)
	}
}

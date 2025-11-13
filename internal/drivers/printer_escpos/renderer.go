package printer_escpos

import (
	"bytes"

	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
)

// renderDocument converts a PrintDocument to ESC/POS commands
func (d *Driver) renderDocument(doc *printer.PrintDocument) ([]byte, error) {
	buf := &bytes.Buffer{}

	// Initialize printer
	buf.Write([]byte{0x1B, 0x40}) // ESC @

	// Set charset to UTF-8 (if supported) or CP437
	// ESC t 0 = CP437 (most compatible)
	buf.Write([]byte{0x1B, 0x74, 0x00})

	// Render all sections
	for _, section := range doc.Sections {
		if err := d.renderSection(buf, section); err != nil {
			return nil, err
		}
	}

	// Handle options
	if doc.Options.Cut {
		// GS V 0 - Full cut
		buf.Write([]byte{0x1D, 0x56, 0x00})
	}

	if doc.Options.DrawerPulse {
		// ESC p 0 50 50
		buf.Write([]byte{0x1B, 0x70, 0x00, 0x32, 0x32})
	}

	return buf.Bytes(), nil
}

// renderSection renders a document section
func (d *Driver) renderSection(buf *bytes.Buffer, section printer.Section) error {
	for _, element := range section.Elements {
		switch element.Type {
		case printer.ElementTypeLine:
			if element.Line != nil {
				d.renderLine(buf, *element.Line)
			}
		case printer.ElementTypeBarcode:
			if element.Barcode != nil {
				d.renderBarcode(buf, *element.Barcode)
			}
		case printer.ElementTypeQRCode:
			if element.QRCode != nil {
				d.renderQRCode(buf, *element.QRCode)
			}
		case printer.ElementTypeImage:
			if element.Image != nil {
				// Image rendering is complex, skip for now
				d.logger.Warn("image printing not yet implemented")
			}
		}
	}

	return nil
}

// renderLine renders a text line
func (d *Driver) renderLine(buf *bytes.Buffer, line printer.Line) {
	// Set alignment
	switch line.Alignment {
	case printer.AlignLeft:
		buf.Write([]byte{0x1B, 0x61, 0x00}) // ESC a 0
	case printer.AlignCenter:
		buf.Write([]byte{0x1B, 0x61, 0x01}) // ESC a 1
	case printer.AlignRight:
		buf.Write([]byte{0x1B, 0x61, 0x02}) // ESC a 2
	}

	// Render runs
	for _, run := range line.Runs {
		d.renderRun(buf, run)
	}

	// Line feed
	buf.WriteByte('\n')

	// Reset alignment to left
	buf.Write([]byte{0x1B, 0x61, 0x00})
}

// renderRun renders styled text
func (d *Driver) renderRun(buf *bytes.Buffer, run printer.Run) {
	// Check if text contains Arabic - if so, use specialized rendering
	if ContainsArabic(run.Text) {
		// Use Arabic shaper for proper text rendering
		// The shaper handles code page, direction, styling, and shaping
		textStyle := TextStyle{
			Bold:         run.Style.Bold,
			Underline:    run.Style.Underline,
			DoubleWidth:  run.Style.DoubleWidth,
			DoubleHeight: run.Style.DoubleHeight,
		}

		// Render Arabic text with proper shaping
		// This handles all ESC/POS commands internally
		arabicBytes := d.arabicShaper.RenderArabicText(run.Text, textStyle)
		buf.Write(arabicBytes)
		return
	}

	// Regular LTR text rendering
	// Set bold
	if run.Style.Bold {
		buf.Write([]byte{0x1B, 0x45, 0x01}) // ESC E 1
	} else {
		buf.Write([]byte{0x1B, 0x45, 0x00}) // ESC E 0
	}

	// Set underline
	if run.Style.Underline {
		buf.Write([]byte{0x1B, 0x2D, 0x01}) // ESC - 1
	} else {
		buf.Write([]byte{0x1B, 0x2D, 0x00}) // ESC - 0
	}

	// Set inverse
	if run.Style.Inverse {
		buf.Write([]byte{0x1D, 0x42, 0x01}) // GS B 1
	} else {
		buf.Write([]byte{0x1D, 0x42, 0x00}) // GS B 0
	}

	// Set font
	if run.Style.Font == printer.FontB {
		buf.Write([]byte{0x1B, 0x4D, 0x01}) // ESC M 1
	} else {
		buf.Write([]byte{0x1B, 0x4D, 0x00}) // ESC M 0
	}

	// Set size
	var sizeCmd byte = 0x00
	if run.Style.DoubleWidth {
		sizeCmd |= 0x20
	}
	if run.Style.DoubleHeight {
		sizeCmd |= 0x10
	}
	buf.Write([]byte{0x1D, 0x21, sizeCmd}) // GS ! n

	// Write text
	buf.WriteString(run.Text)

	// Reset styles
	buf.Write([]byte{0x1B, 0x45, 0x00}) // Bold off
	buf.Write([]byte{0x1B, 0x2D, 0x00}) // Underline off
	buf.Write([]byte{0x1D, 0x42, 0x00}) // Inverse off
	buf.Write([]byte{0x1D, 0x21, 0x00}) // Size normal
}

// renderBarcode renders a 1D barcode
func (d *Driver) renderBarcode(buf *bytes.Buffer, barcode printer.Barcode) {
	// Set barcode height
	// GS h n (default 162)
	height := byte(barcode.Height)
	if height == 0 {
		height = 162
	}
	buf.Write([]byte{0x1D, 0x68, height})

	// Set barcode width
	// GS w n (default 3)
	width := byte(barcode.Width)
	if width == 0 {
		width = 3
	}
	buf.Write([]byte{0x1D, 0x77, width})

	// Print barcode
	// GS k m d1...dk NUL
	buf.WriteByte(0x1D) // GS
	buf.WriteByte(0x6B) // k

	// Map symbology to ESC/POS code
	var symbology byte
	switch barcode.Symbology {
	case printer.BarcodeUPCA:
		symbology = 65
	case printer.BarcodeUPCE:
		symbology = 66
	case printer.BarcodeEAN13:
		symbology = 67
	case printer.BarcodeEAN8:
		symbology = 68
	case printer.BarcodeCode39:
		symbology = 69
	case printer.BarcodeCode128:
		symbology = 73
	case printer.BarcodeITF:
		symbology = 70
	case printer.BarcodeCodabar:
		symbology = 71
	default:
		symbology = 73 // Default to Code128
	}

	buf.WriteByte(symbology)
	buf.WriteByte(byte(len(barcode.Data)))
	buf.WriteString(barcode.Data)

	// Line feed after barcode
	buf.WriteByte('\n')
}

// renderQRCode renders a QR code
func (d *Driver) renderQRCode(buf *bytes.Buffer, qrcode printer.QRCode) {
	// QR code model
	// GS ( k pL pH cn fn n1 n2
	buf.Write([]byte{0x1D, 0x28, 0x6B, 0x04, 0x00, 0x31, 0x41, 0x32, 0x00})

	// QR code size
	size := byte(qrcode.Size)
	if size == 0 {
		size = 6
	}
	buf.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x43, size})

	// Store data
	dataLen := len(qrcode.Data) + 3
	pL := byte(dataLen % 256)
	pH := byte(dataLen / 256)
	buf.Write([]byte{0x1D, 0x28, 0x6B, pL, pH, 0x31, 0x50, 0x30})
	buf.WriteString(qrcode.Data)

	// Print QR code
	buf.Write([]byte{0x1D, 0x28, 0x6B, 0x03, 0x00, 0x31, 0x51, 0x30})

	// Line feed after QR code
	buf.WriteByte('\n')
}

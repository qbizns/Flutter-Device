package virtual_devices

import (
	"context"
	"fmt"
	"strings"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// VirtualPrinter is a virtual printer for testing
type VirtualPrinter struct {
	*devices.BaseDevice
	logger       *telemetry.Logger
	printHistory []string
}

// NewVirtualPrinter creates a new virtual printer
func NewVirtualPrinter(id, name string, logger *telemetry.Logger) *VirtualPrinter {
	base := devices.NewBaseDevice(id, "printer.virtual", name, map[string]string{
		"virtual": "true",
	})

	return &VirtualPrinter{
		BaseDevice:   base,
		logger:       logger.WithDeviceID(id),
		printHistory: []string{},
	}
}

// Start initializes the virtual printer
func (vp *VirtualPrinter) Start(ctx context.Context) error {
	vp.logger.Info("starting virtual printer")
	vp.SetHealth(devices.HealthReady, "virtual printer ready", nil)
	return nil
}

// Stop stops the virtual printer
func (vp *VirtualPrinter) Stop() error {
	vp.logger.Info("stopping virtual printer")
	vp.SetHealth(devices.HealthOffline, "stopped", nil)
	return nil
}

// Print "prints" a document (saves to history)
func (vp *VirtualPrinter) Print(ctx context.Context, doc *printer.PrintDocument) error {
	vp.logger.Info("virtual print")

	// Convert document to text representation
	text := vp.documentToText(doc)

	// Add to history
	vp.printHistory = append(vp.printHistory, text)

	// Log the print
	vp.logger.Info("document printed", telemetry.Int("length", len(text)))

	// Output to console for visibility
	fmt.Println("\n===== VIRTUAL PRINTER OUTPUT =====")
	fmt.Println(text)
	fmt.Println("===== END VIRTUAL PRINTER OUTPUT =====\n")

	return nil
}

// OpenDrawer simulates opening cash drawer
func (vp *VirtualPrinter) OpenDrawer(ctx context.Context) error {
	vp.logger.Info("virtual drawer open")
	fmt.Println("\n===== VIRTUAL DRAWER OPENED =====\n")
	return nil
}

// GetStatus returns printer status
func (vp *VirtualPrinter) GetStatus(ctx context.Context) (*printer.PrinterStatus, error) {
	return &printer.PrinterStatus{
		Online:       true,
		PaperPresent: true,
		DrawerOpen:   false,
	}, nil
}

// GetPrintHistory returns all printed documents
func (vp *VirtualPrinter) GetPrintHistory() []string {
	return vp.printHistory
}

// ClearHistory clears print history
func (vp *VirtualPrinter) ClearHistory() {
	vp.printHistory = []string{}
}

// documentToText converts a print document to text representation
func (vp *VirtualPrinter) documentToText(doc *printer.PrintDocument) string {
	var sb strings.Builder

	for _, section := range doc.Sections {
		for _, element := range section.Elements {
			switch element.Type {
			case printer.ElementTypeLine:
				if element.Line != nil {
					sb.WriteString(vp.lineToText(*element.Line))
					sb.WriteString("\n")
				}
			case printer.ElementTypeBarcode:
				if element.Barcode != nil {
					sb.WriteString(vp.barcodeToText(*element.Barcode))
					sb.WriteString("\n")
				}
			case printer.ElementTypeQRCode:
				if element.QRCode != nil {
					sb.WriteString(vp.qrcodeToText(*element.QRCode))
					sb.WriteString("\n")
				}
			case printer.ElementTypeImage:
				sb.WriteString("[IMAGE]\n")
			}
		}
	}

	if doc.Options.Cut {
		sb.WriteString("--- CUT ---\n")
	}

	if doc.Options.DrawerPulse {
		sb.WriteString("[DRAWER PULSE]\n")
	}

	return sb.String()
}

// lineToText converts a line to text
func (vp *VirtualPrinter) lineToText(line printer.Line) string {
	var sb strings.Builder

	// Build text
	var text string
	for _, run := range line.Runs {
		text += run.Text
	}

	// Apply alignment (assuming 48 character width)
	width := 48
	switch line.Alignment {
	case printer.AlignLeft:
		sb.WriteString(text)
	case printer.AlignCenter:
		padding := (width - len(text)) / 2
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(text)
	case printer.AlignRight:
		padding := width - len(text)
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(text)
	}

	return sb.String()
}

// barcodeToText converts barcode to text
func (vp *VirtualPrinter) barcodeToText(barcode printer.Barcode) string {
	return fmt.Sprintf("[BARCODE: %s | Data: %s]", vp.symbologyName(barcode.Symbology), barcode.Data)
}

// qrcodeToText converts QR code to text
func (vp *VirtualPrinter) qrcodeToText(qrcode printer.QRCode) string {
	return fmt.Sprintf("[QR CODE: %s]", qrcode.Data)
}

// symbologyName returns the name of a barcode symbology
func (vp *VirtualPrinter) symbologyName(symbology printer.BarcodeSymbology) string {
	switch symbology {
	case printer.BarcodeUPCA:
		return "UPC-A"
	case printer.BarcodeUPCE:
		return "UPC-E"
	case printer.BarcodeEAN13:
		return "EAN13"
	case printer.BarcodeEAN8:
		return "EAN8"
	case printer.BarcodeCode39:
		return "Code39"
	case printer.BarcodeCode128:
		return "Code128"
	case printer.BarcodeITF:
		return "ITF"
	case printer.BarcodeCodabar:
		return "Codabar"
	default:
		return "Unknown"
	}
}

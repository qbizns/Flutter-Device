package printer

import (
	"context"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
)

// Printer interface extends Device with print-specific methods
type Printer interface {
	devices.Device

	// Print sends a document to the printer
	Print(ctx context.Context, doc *PrintDocument) error

	// OpenDrawer opens the cash drawer (if connected)
	OpenDrawer(ctx context.Context) error

	// GetStatus queries printer status
	GetStatus(ctx context.Context) (*PrinterStatus, error)
}

// PrintDocument represents a high-level print job
type PrintDocument struct {
	Sections []Section
	Options  PrintOptions
}

// Section represents a logical section of the document
type Section struct {
	Elements []Element
}

// Element can be a line, barcode, QR code, or image
type Element struct {
	Type ElementType
	Line *Line
	Barcode *Barcode
	QRCode *QRCode
	Image *Image
}

// ElementType defines the type of element
type ElementType int

const (
	ElementTypeLine ElementType = iota
	ElementTypeBarcode
	ElementTypeQRCode
	ElementTypeImage
)

// Line represents a single line of output
type Line struct {
	Runs      []Run
	Alignment Alignment
}

// Run represents styled text
type Run struct {
	Text  string
	Style TextStyle
}

// TextStyle defines text formatting
type TextStyle struct {
	Bold         bool
	Underline    bool
	Inverse      bool
	DoubleWidth  bool
	DoubleHeight bool
	Font         Font
}

// Alignment defines text alignment
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
)

// Font defines font selection
type Font int

const (
	FontA Font = iota // Usually 12x24
	FontB             // Usually 9x17
)

// Barcode represents a 1D barcode
type Barcode struct {
	Data      string
	Symbology BarcodeSymbology
	Height    int
	Width     int
}

// BarcodeSymbology defines barcode types
type BarcodeSymbology int

const (
	BarcodeUPCA BarcodeSymbology = iota
	BarcodeUPCE
	BarcodeEAN13
	BarcodeEAN8
	BarcodeCode39
	BarcodeCode128
	BarcodeITF
	BarcodeCodabar
)

// QRCode represents a QR code
type QRCode struct {
	Data string
	Size int
}

// Image represents a bitmap image
type Image struct {
	Data   []byte // PNG or JPEG
	Width  int
	Height int
}

// PrintOptions contains print job options
type PrintOptions struct {
	Cut         bool
	DrawerPulse bool
	Copies      int
}

// PrinterStatus represents current printer state
type PrinterStatus struct {
	Online       bool
	PaperPresent bool
	DrawerOpen   bool
	Error        string
}

// NewPrintDocument creates a new empty print document
func NewPrintDocument() *PrintDocument {
	return &PrintDocument{
		Sections: []Section{},
		Options: PrintOptions{
			Cut:    true,
			Copies: 1,
		},
	}
}

// AddSection adds a section to the document
func (d *PrintDocument) AddSection(section Section) {
	d.Sections = append(d.Sections, section)
}

// AddTextLine adds a text line to the current section
func (d *PrintDocument) AddTextLine(text string, align Alignment, style TextStyle) {
	line := Line{
		Runs: []Run{
			{Text: text, Style: style},
		},
		Alignment: align,
	}
	element := Element{
		Type: ElementTypeLine,
		Line: &line,
	}

	if len(d.Sections) == 0 {
		d.Sections = append(d.Sections, Section{})
	}
	lastSection := &d.Sections[len(d.Sections)-1]
	lastSection.Elements = append(lastSection.Elements, element)
}

// AddBarcode adds a barcode
func (d *PrintDocument) AddBarcode(data string, symbology BarcodeSymbology, height int) {
	barcode := &Barcode{
		Data:      data,
		Symbology: symbology,
		Height:    height,
		Width:     2,
	}
	element := Element{
		Type:    ElementTypeBarcode,
		Barcode: barcode,
	}

	if len(d.Sections) == 0 {
		d.Sections = append(d.Sections, Section{})
	}
	lastSection := &d.Sections[len(d.Sections)-1]
	lastSection.Elements = append(lastSection.Elements, element)
}

// AddQRCode adds a QR code
func (d *PrintDocument) AddQRCode(data string, size int) {
	qrcode := &QRCode{
		Data: data,
		Size: size,
	}
	element := Element{
		Type:   ElementTypeQRCode,
		QRCode: qrcode,
	}

	if len(d.Sections) == 0 {
		d.Sections = append(d.Sections, Section{})
	}
	lastSection := &d.Sections[len(d.Sections)-1]
	lastSection.Elements = append(lastSection.Elements, element)
}

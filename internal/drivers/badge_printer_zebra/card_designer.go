package badge_printer_zebra

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// CardDesigner creates ZPL commands for badge printing
type CardDesigner struct {
	dpi    int // 300 or 600
	logger *telemetry.Logger
}

// NewCardDesigner creates a new card designer
func NewCardDesigner(dpi int, logger *telemetry.Logger) *CardDesigner {
	if dpi != 300 && dpi != 600 {
		dpi = 300 // Default
	}

	return &CardDesigner{
		dpi:    dpi,
		logger: logger,
	}
}

// RenderToZPL converts a badge design to ZPL commands
func (cd *CardDesigner) RenderToZPL(design *pb.BadgeDesign, options *pb.PrintOptions) (string, error) {
	var zpl strings.Builder

	// Start ZPL
	zpl.WriteString(ZPLStart + "\n")

	// Set print origin
	zpl.WriteString("^LH0,0\n")

	// Set field orientation to portrait/landscape
	if design.Orientation == pb.Orientation_ORIENTATION_LANDSCAPE {
		zpl.WriteString("^FWR\n") // Rotate 90 degrees
	} else {
		zpl.WriteString("^FWN\n") // Normal (portrait)
	}

	// Resolve design (template or custom)
	elements := design.FrontElements
	if design.GetTemplateId() != "" {
		// Apply template with variables
		template, err := GetTemplateByID(design.GetTemplateId())
		if err != nil {
			return "", fmt.Errorf("template not found: %w", err)
		}
		elements = cd.applyVariables(template.Design.FrontElements, design.Variables)
	}

	// Render front side
	if err := cd.renderSide(&zpl, elements, "front"); err != nil {
		return "", fmt.Errorf("front side render error: %w", err)
	}

	// If dual-sided and back elements exist
	if options.GetSide() == pb.PrintOptions_PRINT_SIDE_BOTH && len(design.BackElements) > 0 {
		zpl.WriteString("^POI\n") // Print orientation inverted (back side)
		if err := cd.renderSide(&zpl, design.BackElements, "back"); err != nil {
			return "", fmt.Errorf("back side render error: %w", err)
		}
	}

	// Print quantity
	copies := options.GetCopies()
	if copies > 0 {
		zpl.WriteString(fmt.Sprintf("^PQ%d\n", copies))
	}

	// End ZPL
	zpl.WriteString(ZPLEnd + "\n")

	return zpl.String(), nil
}

// renderSide renders one side of the card
func (cd *CardDesigner) renderSide(zpl *strings.Builder, elements []*pb.BadgeElement, side string) error {
	// Sort elements by z-index (lower first, so higher are on top)
	sortedElements := cd.sortByZIndex(elements)

	for _, elem := range sortedElements {
		switch elem.Type {
		case pb.BadgeElement_ELEMENT_TYPE_TEXT:
			cd.renderText(zpl, elem)
		case pb.BadgeElement_ELEMENT_TYPE_IMAGE, pb.BadgeElement_ELEMENT_TYPE_PHOTO, pb.BadgeElement_ELEMENT_TYPE_LOGO:
			cd.renderImage(zpl, elem)
		case pb.BadgeElement_ELEMENT_TYPE_BARCODE:
			cd.renderBarcode(zpl, elem)
		case pb.BadgeElement_ELEMENT_TYPE_QR_CODE:
			cd.renderQRCode(zpl, elem)
		case pb.BadgeElement_ELEMENT_TYPE_RECTANGLE:
			cd.renderRectangle(zpl, elem)
		case pb.BadgeElement_ELEMENT_TYPE_LINE:
			cd.renderLine(zpl, elem)
		case pb.BadgeElement_ELEMENT_TYPE_CIRCLE:
			cd.renderCircle(zpl, elem)
		default:
			cd.logger.Warn("unsupported element type",
				telemetry.String("type", elem.Type.String()),
			)
		}
	}
	return nil
}

// renderText renders a text element
func (cd *CardDesigner) renderText(zpl *strings.Builder, elem *pb.BadgeElement) {
	x := elem.Position.X
	y := elem.Position.Y

	// Field origin
	zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

	// Font selection
	fontSize := elem.Style.GetFontSize()
	if fontSize == 0 {
		fontSize = 24
	}

	// Scale font size for DPI
	if cd.dpi == 600 {
		fontSize = fontSize * 2
	}

	// ZPL font command
	// ^A0N,height,width (scalable font)
	fontWidth := fontSize
	if elem.Style.GetBold() {
		fontWidth = int32(float32(fontSize) * 1.2) // Slightly wider for bold
	}

	zpl.WriteString(fmt.Sprintf("^A0N,%d,%d\n", fontSize, fontWidth))

	// Alignment (for centered text, we need to adjust position)
	// This is simplified - full implementation would calculate text width
	if elem.Style.GetAlignment() == pb.ElementStyle_TEXT_ALIGN_CENTER {
		zpl.WriteString("^FB" + fmt.Sprintf("%d,1,0,C,0\n", elem.Size.Width))
	} else if elem.Style.GetAlignment() == pb.ElementStyle_TEXT_ALIGN_RIGHT {
		zpl.WriteString("^FB" + fmt.Sprintf("%d,1,0,R,0\n", elem.Size.Width))
	}

	// Field data
	content := elem.Content
	if content == "" {
		content = "N/A"
	}

	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", content))
}

// renderImage renders an image element
func (cd *CardDesigner) renderImage(zpl *strings.Builder, elem *pb.BadgeElement) {
	x := elem.Position.X
	y := elem.Position.Y

	zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

	// For images, content should be base64 encoded image data
	// In production, convert image to ZPL ^GF (graphic field) command
	// This is a placeholder - full implementation requires image processing

	if elem.Content != "" {
		// Decode base64 image
		imageData, err := base64.StdEncoding.DecodeString(elem.Content)
		if err == nil {
			// Convert to ZPL graphic format
			// This is complex - would use Zebra SDK or image conversion library
			cd.logger.Debug("image rendering",
				telemetry.Int("bytes", len(imageData)),
			)

			// Placeholder for actual image rendering
			// ^GFA,<bytes>,<bytes>,<rowBytes>,<hexData>
			zpl.WriteString("^GFA,0,0,0,^FS\n")
		}
	}
}

// renderBarcode renders a barcode
func (cd *CardDesigner) renderBarcode(zpl *strings.Builder, elem *pb.BadgeElement) {
	x := elem.Position.X
	y := elem.Position.Y

	zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

	// Determine barcode type
	barcodeType := elem.Style.GetBarcodeType()
	showText := elem.Style.GetShowBarcodeText()

	height := elem.Size.Height
	if height == 0 {
		height = 100
	}

	// Scale for DPI
	if cd.dpi == 600 {
		height = height * 2
	}

	switch barcodeType {
	case pb.ElementStyle_BARCODE_TYPE_CODE128:
		// ^BCN,height,printInterpretationLine
		showTextFlag := "N"
		if showText {
			showTextFlag = "Y"
		}
		zpl.WriteString(fmt.Sprintf("^BCN,%d,%s,N,N\n", height, showTextFlag))

	case pb.ElementStyle_BARCODE_TYPE_CODE39:
		showTextFlag := "N"
		if showText {
			showTextFlag = "Y"
		}
		zpl.WriteString(fmt.Sprintf("^B3N,N,%d,%s,N\n", height, showTextFlag))

	case pb.ElementStyle_BARCODE_TYPE_EAN13:
		showTextFlag := "N"
		if showText {
			showTextFlag = "Y"
		}
		zpl.WriteString(fmt.Sprintf("^BEN,%d,%s,N\n", height, showTextFlag))

	default:
		// Default to Code 128
		zpl.WriteString(fmt.Sprintf("^BCN,%d,Y,N,N\n", height))
	}

	zpl.WriteString(fmt.Sprintf("^FD%s^FS\n", elem.Content))
}

// renderQRCode renders a QR code
func (cd *CardDesigner) renderQRCode(zpl *strings.Builder, elem *pb.BadgeElement) {
	x := elem.Position.X
	y := elem.Position.Y

	zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

	// QR code magnification (module size)
	magnification := 3
	if cd.dpi == 600 {
		magnification = 6
	}

	// ^BQN,model,magnification,errorCorrection,maskValue
	// Model 2 (enhanced), magnification, error correction M (15%)
	zpl.WriteString(fmt.Sprintf("^BQN,2,%d,M,7\n", magnification))

	// QR code data - mode A for automatic
	zpl.WriteString(fmt.Sprintf("^FDMA,%s^FS\n", elem.Content))
}

// renderRectangle renders a rectangle
func (cd *CardDesigner) renderRectangle(zpl *strings.Builder, elem *pb.BadgeElement) {
	x := elem.Position.X
	y := elem.Position.Y
	w := elem.Size.Width
	h := elem.Size.Height

	// Scale for DPI
	if cd.dpi == 600 {
		w = w * 2
		h = h * 2
	}

	zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

	thickness := int32(1)
	if elem.Style != nil && elem.Style.BorderWidth > 0 {
		thickness = elem.Style.BorderWidth
		if cd.dpi == 600 {
			thickness = thickness * 2
		}
	}

	// ^GB: Graphic Box (width, height, thickness)
	// If thickness equals width or height, it's filled
	if elem.Style != nil && elem.Style.Background != "" {
		// Filled rectangle
		zpl.WriteString(fmt.Sprintf("^GB%d,%d,%d^FS\n", w, h, w))
	} else {
		// Outline rectangle
		zpl.WriteString(fmt.Sprintf("^GB%d,%d,%d^FS\n", w, h, thickness))
	}
}

// renderLine renders a line
func (cd *CardDesigner) renderLine(zpl *strings.Builder, elem *pb.BadgeElement) {
	x := elem.Position.X
	y := elem.Position.Y
	w := elem.Size.Width
	h := elem.Size.Height

	// Scale for DPI
	if cd.dpi == 600 {
		w = w * 2
		h = h * 2
	}

	thickness := int32(2)
	if elem.Style != nil && elem.Style.BorderWidth > 0 {
		thickness = elem.Style.BorderWidth
		if cd.dpi == 600 {
			thickness = thickness * 2
		}
	}

	zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

	// Determine if horizontal or vertical
	if h > w {
		// Vertical line
		zpl.WriteString(fmt.Sprintf("^GB%d,%d,%d^FS\n", thickness, h, thickness))
	} else {
		// Horizontal line
		zpl.WriteString(fmt.Sprintf("^GB%d,%d,%d^FS\n", w, thickness, thickness))
	}
}

// renderCircle renders a circle
func (cd *CardDesigner) renderCircle(zpl *strings.Builder, elem *pb.BadgeElement) {
	x := elem.Position.X
	y := elem.Position.Y
	diameter := elem.Size.Width // Use width as diameter

	// Scale for DPI
	if cd.dpi == 600 {
		diameter = diameter * 2
	}

	zpl.WriteString(fmt.Sprintf("^FO%d,%d\n", x, y))

	thickness := int32(1)
	if elem.Style != nil && elem.Style.BorderWidth > 0 {
		thickness = elem.Style.BorderWidth
		if cd.dpi == 600 {
			thickness = thickness * 2
		}
	}

	// ^GC: Graphic Circle (diameter, thickness)
	zpl.WriteString(fmt.Sprintf("^GC%d,%d^FS\n", diameter, thickness))
}

// applyVariables replaces variable placeholders with actual values
func (cd *CardDesigner) applyVariables(elements []*pb.BadgeElement, variables map[string]string) []*pb.BadgeElement {
	result := make([]*pb.BadgeElement, len(elements))

	for i, elem := range elements {
		// Clone element
		newElem := &pb.BadgeElement{
			Type:     elem.Type,
			Position: elem.Position,
			Size:     elem.Size,
			Style:    elem.Style,
			ZIndex:   elem.ZIndex,
			Content:  elem.Content,
		}

		// Replace variables in content
		content := elem.Content
		for key, value := range variables {
			placeholder := "{{" + key + "}}"
			content = strings.ReplaceAll(content, placeholder, value)
		}
		newElem.Content = content

		result[i] = newElem
	}

	return result
}

// sortByZIndex sorts elements by z-index (lower first)
func (cd *CardDesigner) sortByZIndex(elements []*pb.BadgeElement) []*pb.BadgeElement {
	sorted := make([]*pb.BadgeElement, len(elements))
	copy(sorted, elements)

	// Simple bubble sort (fine for small arrays)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].ZIndex > sorted[j].ZIndex {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// PreviewBadge generates a preview of the badge design
// In a full implementation, this would render to an image
func (cd *CardDesigner) PreviewBadge(design *pb.BadgeDesign) ([]byte, error) {
	// This would generate a PNG/JPEG preview
	// For now, return empty
	return []byte{}, fmt.Errorf("preview not implemented")
}

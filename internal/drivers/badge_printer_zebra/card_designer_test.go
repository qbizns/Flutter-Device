package badge_printer_zebra

import (
	"strings"
	"testing"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

func TestNewCardDesigner(t *testing.T) {
	tests := []struct {
		dpi      int
		expected int
	}{
		{300, 300},
		{600, 600},
		{150, 300}, // Invalid, should default to 300
		{0, 300},   // Invalid, should default to 300
	}

	for _, tt := range tests {
		designer := NewCardDesigner(tt.dpi, nil)
		if designer.dpi != tt.expected {
			t.Errorf("DPI: expected %d, got %d", tt.expected, designer.dpi)
		}
	}
}

func TestCardDesigner_RenderText(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	elem := &pb.BadgeElement{
		Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
		Position: &pb.Position{X: 100, Y: 200},
		Size:     &pb.Size{Width: 400, Height: 50},
		Content:  "Hello World",
		Style: &pb.ElementStyle{
			FontSize:  30,
			Bold:      true,
			Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
		},
	}

	var zpl strings.Builder
	designer.renderText(&zpl, elem)

	result := zpl.String()

	// Check for field origin
	if !strings.Contains(result, "^FO100,200") {
		t.Error("ZPL missing field origin")
	}

	// Check for font command
	if !strings.Contains(result, "^A0N") {
		t.Error("ZPL missing font command")
	}

	// Check for content
	if !strings.Contains(result, "Hello World") {
		t.Error("ZPL missing content")
	}

	// Check for field separator
	if !strings.Contains(result, "^FS") {
		t.Error("ZPL missing field separator")
	}
}

func TestCardDesigner_RenderBarcode(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	elem := &pb.BadgeElement{
		Type:     pb.BadgeElement_ELEMENT_TYPE_BARCODE,
		Position: &pb.Position{X: 100, Y: 500},
		Size:     &pb.Size{Width: 800, Height: 100},
		Content:  "1234567890",
		Style: &pb.ElementStyle{
			BarcodeType:     pb.ElementStyle_BARCODE_TYPE_CODE128,
			ShowBarcodeText: true,
		},
	}

	var zpl strings.Builder
	designer.renderBarcode(&zpl, elem)

	result := zpl.String()

	// Check for barcode command
	if !strings.Contains(result, "^BC") { // Code 128
		t.Error("ZPL missing barcode command")
	}

	// Check for content
	if !strings.Contains(result, "1234567890") {
		t.Error("ZPL missing barcode content")
	}
}

func TestCardDesigner_RenderQRCode(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	elem := &pb.BadgeElement{
		Type:     pb.BadgeElement_ELEMENT_TYPE_QR_CODE,
		Position: &pb.Position{X: 400, Y: 300},
		Size:     &pb.Size{Width: 200, Height: 200},
		Content:  "https://example.com/badge/12345",
	}

	var zpl strings.Builder
	designer.renderQRCode(&zpl, elem)

	result := zpl.String()

	// Check for QR code command
	if !strings.Contains(result, "^BQN") {
		t.Error("ZPL missing QR code command")
	}

	// Check for content (prefixed with MA for automatic mode)
	if !strings.Contains(result, "MA,https://example.com") {
		t.Error("ZPL missing QR code content")
	}
}

func TestCardDesigner_RenderRectangle(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	elem := &pb.BadgeElement{
		Type:     pb.BadgeElement_ELEMENT_TYPE_RECTANGLE,
		Position: &pb.Position{X: 10, Y: 10},
		Size:     &pb.Size{Width: 1000, Height: 640},
		Style: &pb.ElementStyle{
			BorderWidth: 5,
		},
	}

	var zpl strings.Builder
	designer.renderRectangle(&zpl, elem)

	result := zpl.String()

	// Check for graphic box command
	if !strings.Contains(result, "^GB") {
		t.Error("ZPL missing graphic box command")
	}
}

func TestCardDesigner_RenderToZPL_Simple(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	design := &pb.BadgeDesign{
		Orientation: pb.Orientation_ORIENTATION_PORTRAIT,
		FrontElements: []*pb.BadgeElement{
			{
				Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
				Position: &pb.Position{X: 100, Y: 100},
				Size:     &pb.Size{Width: 800, Height: 60},
				Content:  "Test Badge",
				Style: &pb.ElementStyle{
					FontSize:  40,
					Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
				},
			},
		},
	}

	options := &pb.PrintOptions{
		Side:    pb.PrintOptions_PRINT_SIDE_FRONT_ONLY,
		Quality: pb.PrintOptions_PRINT_QUALITY_NORMAL,
		Copies:  1,
	}

	zpl, err := designer.RenderToZPL(design, options)
	if err != nil {
		t.Fatalf("RenderToZPL failed: %v", err)
	}

	// Check ZPL structure
	if !strings.HasPrefix(zpl, "^XA") {
		t.Error("ZPL should start with ^XA")
	}

	if !strings.HasSuffix(strings.TrimSpace(zpl), "^XZ") {
		t.Error("ZPL should end with ^XZ")
	}

	// Check for content
	if !strings.Contains(zpl, "Test Badge") {
		t.Error("ZPL missing badge content")
	}
}

func TestCardDesigner_RenderToZPL_WithTemplate(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	design := &pb.BadgeDesign{
		DesignSource: &pb.BadgeDesign_TemplateId{
			TemplateId: "conference_attendee",
		},
		Variables: map[string]string{
			"name":     "Alice Smith",
			"company":  "Tech Corp",
			"badge_id": "ATT-001",
		},
	}

	options := &pb.PrintOptions{
		Side:   pb.PrintOptions_PRINT_SIDE_FRONT_ONLY,
		Copies: 1,
	}

	zpl, err := designer.RenderToZPL(design, options)
	if err != nil {
		t.Fatalf("RenderToZPL failed: %v", err)
	}

	// Check that variables were substituted
	if !strings.Contains(zpl, "Alice Smith") {
		t.Error("ZPL missing substituted name")
	}

	if !strings.Contains(zpl, "Tech Corp") {
		t.Error("ZPL missing substituted company")
	}

	if !strings.Contains(zpl, "ATT-001") {
		t.Error("ZPL missing substituted badge ID")
	}
}

func TestCardDesigner_ApplyVariables(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	elements := []*pb.BadgeElement{
		{
			Type:    pb.BadgeElement_ELEMENT_TYPE_TEXT,
			Content: "Hello {{name}}",
		},
		{
			Type:    pb.BadgeElement_ELEMENT_TYPE_TEXT,
			Content: "Welcome to {{event}}",
		},
	}

	variables := map[string]string{
		"name":  "John Doe",
		"event": "Tech Conference 2025",
	}

	result := designer.applyVariables(elements, variables)

	if result[0].Content != "Hello John Doe" {
		t.Errorf("Expected 'Hello John Doe', got '%s'", result[0].Content)
	}

	if result[1].Content != "Welcome to Tech Conference 2025" {
		t.Errorf("Expected 'Welcome to Tech Conference 2025', got '%s'", result[1].Content)
	}
}

func TestCardDesigner_SortByZIndex(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	elements := []*pb.BadgeElement{
		{Type: pb.BadgeElement_ELEMENT_TYPE_TEXT, ZIndex: 5},
		{Type: pb.BadgeElement_ELEMENT_TYPE_TEXT, ZIndex: 1},
		{Type: pb.BadgeElement_ELEMENT_TYPE_TEXT, ZIndex: 3},
		{Type: pb.BadgeElement_ELEMENT_TYPE_TEXT, ZIndex: 2},
	}

	sorted := designer.sortByZIndex(elements)

	expectedOrder := []int32{1, 2, 3, 5}
	for i, elem := range sorted {
		if elem.ZIndex != expectedOrder[i] {
			t.Errorf("Element %d: expected z-index %d, got %d", i, expectedOrder[i], elem.ZIndex)
		}
	}
}

func TestCardDesigner_DualSided(t *testing.T) {
	designer := NewCardDesigner(300, nil)

	design := &pb.BadgeDesign{
		FrontElements: []*pb.BadgeElement{
			{
				Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
				Position: &pb.Position{X: 100, Y: 100},
				Content:  "Front",
			},
		},
		BackElements: []*pb.BadgeElement{
			{
				Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
				Position: &pb.Position{X: 100, Y: 100},
				Content:  "Back",
			},
		},
	}

	options := &pb.PrintOptions{
		Side: pb.PrintOptions_PRINT_SIDE_BOTH,
	}

	zpl, err := designer.RenderToZPL(design, options)
	if err != nil {
		t.Fatalf("RenderToZPL failed: %v", err)
	}

	// Check for print orientation inversion command
	if !strings.Contains(zpl, "^POI") {
		t.Error("ZPL missing print orientation inversion for back side")
	}

	// Check both sides are present
	if !strings.Contains(zpl, "Front") {
		t.Error("ZPL missing front side content")
	}

	if !strings.Contains(zpl, "Back") {
		t.Error("ZPL missing back side content")
	}
}

func TestCardDesigner_DPIScaling(t *testing.T) {
	tests := []struct {
		dpi           int
		originalSize  int32
		expectContain string
	}{
		{300, 100, "^BCN,100"},
		{600, 100, "^BCN,200"}, // Should be doubled for 600 DPI
	}

	for _, tt := range tests {
		designer := NewCardDesigner(tt.dpi, nil)

		elem := &pb.BadgeElement{
			Type:     pb.BadgeElement_ELEMENT_TYPE_BARCODE,
			Position: &pb.Position{X: 100, Y: 100},
			Size:     &pb.Size{Width: 400, Height: tt.originalSize},
			Content:  "TEST",
			Style: &pb.ElementStyle{
				BarcodeType: pb.ElementStyle_BARCODE_TYPE_CODE128,
			},
		}

		var zpl strings.Builder
		designer.renderBarcode(&zpl, elem)

		result := zpl.String()
		if !strings.Contains(result, tt.expectContain) {
			t.Errorf("DPI %d: expected ZPL to contain '%s', got: %s", tt.dpi, tt.expectContain, result)
		}
	}
}

func BenchmarkCardDesigner_RenderToZPL(b *testing.B) {
	designer := NewCardDesigner(300, nil)

	design := &pb.BadgeDesign{
		DesignSource: &pb.BadgeDesign_TemplateId{
			TemplateId: "conference_attendee",
		},
		Variables: map[string]string{
			"name":     "John Doe",
			"company":  "Example Corp",
			"badge_id": "12345",
		},
	}

	options := &pb.PrintOptions{
		Side: pb.PrintOptions_PRINT_SIDE_FRONT_ONLY,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = designer.RenderToZPL(design, options)
	}
}

package badge_printer_zebra

import (
	"fmt"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

// GetPredefinedTemplates returns all predefined badge templates
func GetPredefinedTemplates() []*pb.BadgeTemplate {
	return []*pb.BadgeTemplate{
		ConferenceAttendeeTemplate(),
		VIPBadgeTemplate(),
		StaffBadgeTemplate(),
		VisitorBadgeTemplate(),
	}
}

// GetTemplateByID returns a specific template by ID
func GetTemplateByID(id string) (*pb.BadgeTemplate, error) {
	templates := map[string]*pb.BadgeTemplate{
		"conference_attendee": ConferenceAttendeeTemplate(),
		"vip_badge":           VIPBadgeTemplate(),
		"staff_badge":         StaffBadgeTemplate(),
		"visitor_badge":       VisitorBadgeTemplate(),
	}

	template, ok := templates[id]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	return template, nil
}

// ConferenceAttendeeTemplate returns a standard conference attendee badge
func ConferenceAttendeeTemplate() *pb.BadgeTemplate {
	return &pb.BadgeTemplate{
		Id:          "conference_attendee",
		Name:        "Conference Attendee",
		Description: "Standard attendee badge with name, company, and QR code",
		RequiredVariables: []string{
			"name",
			"company",
			"badge_id",
		},
		OptionalVariables: []string{
			"logo",
			"qr_data",
			"event_name",
		},
		Design: &pb.BadgeDesign{
			CardSize:        &pb.CardSize{Type: pb.CardSize_SIZE_TYPE_CR80},
			Orientation:     pb.Orientation_ORIENTATION_PORTRAIT,
			BackgroundColor: "#FFFFFF",
			FrontElements: []*pb.BadgeElement{
				// Top logo
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_LOGO,
					Position: &pb.Position{X: 308, Y: 30}, // Centered
					Size:     &pb.Size{Width: 400, Height: 80},
					Content:  "{{logo}}",
					ZIndex:   1,
				},
				// Event name (if provided)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 120},
					Size:     &pb.Size{Width: 916, Height: 30},
					Content:  "{{event_name}}",
					ZIndex:   2,
					Style: &pb.ElementStyle{
						FontSize:  18,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#666666",
					},
				},
				// Attendee name (large, centered)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 180},
					Size:     &pb.Size{Width: 916, Height: 80},
					Content:  "{{name}}",
					ZIndex:   3,
					Style: &pb.ElementStyle{
						FontSize:  40,
						Bold:      true,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#000000",
					},
				},
				// Company name
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 280},
					Size:     &pb.Size{Width: 916, Height: 50},
					Content:  "{{company}}",
					ZIndex:   4,
					Style: &pb.ElementStyle{
						FontSize:  28,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#333333",
					},
				},
				// QR Code (centered)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_QR_CODE,
					Position: &pb.Position{X: 408, Y: 360},
					Size:     &pb.Size{Width: 200, Height: 200},
					Content:  "{{qr_data}}",
					ZIndex:   5,
				},
				// Badge ID barcode (bottom)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_BARCODE,
					Position: &pb.Position{X: 100, Y: 580},
					Size:     &pb.Size{Width: 816, Height: 50},
					Content:  "{{badge_id}}",
					ZIndex:   6,
					Style: &pb.ElementStyle{
						BarcodeType:     pb.ElementStyle_BARCODE_TYPE_CODE128,
						ShowBarcodeText: true,
					},
				},
			},
		},
	}
}

// VIPBadgeTemplate returns a VIP badge with gold border
func VIPBadgeTemplate() *pb.BadgeTemplate {
	return &pb.BadgeTemplate{
		Id:          "vip_badge",
		Name:        "VIP Badge",
		Description: "VIP badge with gold border and special designation",
		RequiredVariables: []string{
			"name",
		},
		OptionalVariables: []string{
			"title",
			"access_level",
			"qr_data",
		},
		Design: &pb.BadgeDesign{
			CardSize:        &pb.CardSize{Type: pb.CardSize_SIZE_TYPE_CR80},
			Orientation:     pb.Orientation_ORIENTATION_PORTRAIT,
			BackgroundColor: "#FFFFFF",
			FrontElements: []*pb.BadgeElement{
				// Gold border rectangle
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_RECTANGLE,
					Position: &pb.Position{X: 10, Y: 10},
					Size:     &pb.Size{Width: 996, Height: 628},
					ZIndex:   1,
					Style: &pb.ElementStyle{
						BorderWidth: 8,
						BorderColor: "#FFD700", // Gold
					},
				},
				// Second inner border
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_RECTANGLE,
					Position: &pb.Position{X: 30, Y: 30},
					Size:     &pb.Size{Width: 956, Height: 588},
					ZIndex:   2,
					Style: &pb.ElementStyle{
						BorderWidth: 2,
						BorderColor: "#FFD700",
					},
				},
				// "VIP" text at top (large, gold)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 80},
					Size:     &pb.Size{Width: 916, Height: 100},
					Content:  "VIP",
					ZIndex:   3,
					Style: &pb.ElementStyle{
						FontSize:  60,
						Bold:      true,
						Color:     "#FFD700",
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
					},
				},
				// Name (large, centered)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 220},
					Size:     &pb.Size{Width: 916, Height: 80},
					Content:  "{{name}}",
					ZIndex:   4,
					Style: &pb.ElementStyle{
						FontSize:  36,
						Bold:      true,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#000000",
					},
				},
				// Title/Position
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 320},
					Size:     &pb.Size{Width: 916, Height: 50},
					Content:  "{{title}}",
					ZIndex:   5,
					Style: &pb.ElementStyle{
						FontSize:  26,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#666666",
					},
				},
				// Access level
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 400},
					Size:     &pb.Size{Width: 916, Height: 40},
					Content:  "Access: {{access_level}}",
					ZIndex:   6,
					Style: &pb.ElementStyle{
						FontSize:  22,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#333333",
					},
				},
				// QR code for validation
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_QR_CODE,
					Position: &pb.Position{X: 408, Y: 470},
					Size:     &pb.Size{Width: 150, Height: 150},
					Content:  "{{qr_data}}",
					ZIndex:   7,
				},
			},
		},
	}
}

// StaffBadgeTemplate returns a staff badge with photo
func StaffBadgeTemplate() *pb.BadgeTemplate {
	return &pb.BadgeTemplate{
		Id:          "staff_badge",
		Name:        "Staff Badge",
		Description: "Staff identification badge with photo and department",
		RequiredVariables: []string{
			"name",
			"department",
			"employee_id",
		},
		OptionalVariables: []string{
			"photo",
			"position",
			"phone",
		},
		Design: &pb.BadgeDesign{
			CardSize:        &pb.CardSize{Type: pb.CardSize_SIZE_TYPE_CR80},
			Orientation:     pb.Orientation_ORIENTATION_PORTRAIT,
			BackgroundColor: "#FFFFFF",
			FrontElements: []*pb.BadgeElement{
				// Blue header banner
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_RECTANGLE,
					Position: &pb.Position{X: 0, Y: 0},
					Size:     &pb.Size{Width: 1016, Height: 100},
					ZIndex:   1,
					Style: &pb.ElementStyle{
						Background: "#0066CC", // Blue
					},
				},
				// "STAFF" text on banner
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 30},
					Size:     &pb.Size{Width: 916, Height: 40},
					Content:  "STAFF",
					ZIndex:   2,
					Style: &pb.ElementStyle{
						FontSize:  36,
						Bold:      true,
						Color:     "#FFFFFF",
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
					},
				},
				// Photo placeholder/image
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_PHOTO,
					Position: &pb.Position{X: 100, Y: 130},
					Size:     &pb.Size{Width: 250, Height: 300},
					Content:  "{{photo}}",
					ZIndex:   3,
					Style: &pb.ElementStyle{
						HasBorder:   true,
						BorderWidth: 3,
						BorderColor: "#0066CC",
					},
				},
				// Name
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 380, Y: 150},
					Size:     &pb.Size{Width: 600, Height: 60},
					Content:  "{{name}}",
					ZIndex:   4,
					Style: &pb.ElementStyle{
						FontSize: 32,
						Bold:     true,
						Color:    "#000000",
					},
				},
				// Department
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 380, Y: 230},
					Size:     &pb.Size{Width: 600, Height: 40},
					Content:  "{{department}}",
					ZIndex:   5,
					Style: &pb.ElementStyle{
						FontSize: 24,
						Color:    "#333333",
					},
				},
				// Position
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 380, Y: 290},
					Size:     &pb.Size{Width: 600, Height: 35},
					Content:  "{{position}}",
					ZIndex:   6,
					Style: &pb.ElementStyle{
						FontSize: 20,
						Color:    "#666666",
					},
				},
				// Phone
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 380, Y: 340},
					Size:     &pb.Size{Width: 600, Height: 30},
					Content:  "Ext: {{phone}}",
					ZIndex:   7,
					Style: &pb.ElementStyle{
						FontSize: 18,
						Color:    "#666666",
					},
				},
				// Employee ID barcode (bottom)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_BARCODE,
					Position: &pb.Position{X: 100, Y: 500},
					Size:     &pb.Size{Width: 816, Height: 80},
					Content:  "{{employee_id}}",
					ZIndex:   8,
					Style: &pb.ElementStyle{
						BarcodeType:     pb.ElementStyle_BARCODE_TYPE_CODE128,
						ShowBarcodeText: true,
					},
				},
			},
		},
	}
}

// VisitorBadgeTemplate returns a visitor badge with expiration
func VisitorBadgeTemplate() *pb.BadgeTemplate {
	return &pb.BadgeTemplate{
		Id:          "visitor_badge",
		Name:        "Visitor Badge",
		Description: "Temporary visitor badge with date and host information",
		RequiredVariables: []string{
			"name",
			"company",
			"host",
			"date",
		},
		OptionalVariables: []string{
			"expires",
			"visitor_id",
		},
		Design: &pb.BadgeDesign{
			CardSize:        &pb.CardSize{Type: pb.CardSize_SIZE_TYPE_CR80},
			Orientation:     pb.Orientation_ORIENTATION_PORTRAIT,
			BackgroundColor: "#FFFFFF",
			FrontElements: []*pb.BadgeElement{
				// Red border for visitor identification
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_RECTANGLE,
					Position: &pb.Position{X: 10, Y: 10},
					Size:     &pb.Size{Width: 996, Height: 628},
					ZIndex:   1,
					Style: &pb.ElementStyle{
						BorderWidth: 5,
						BorderColor: "#FF0000", // Red
					},
				},
				// "VISITOR" banner (red background)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_RECTANGLE,
					Position: &pb.Position{X: 0, Y: 40},
					Size:     &pb.Size{Width: 1016, Height: 80},
					ZIndex:   2,
					Style: &pb.ElementStyle{
						Background: "#FF0000",
					},
				},
				// "VISITOR" text
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 55},
					Size:     &pb.Size{Width: 916, Height: 50},
					Content:  "VISITOR",
					ZIndex:   3,
					Style: &pb.ElementStyle{
						FontSize:  44,
						Bold:      true,
						Color:     "#FFFFFF",
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
					},
				},
				// Visitor name
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 160},
					Size:     &pb.Size{Width: 916, Height: 70},
					Content:  "{{name}}",
					ZIndex:   4,
					Style: &pb.ElementStyle{
						FontSize:  34,
						Bold:      true,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#000000",
					},
				},
				// Company
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 250},
					Size:     &pb.Size{Width: 916, Height: 45},
					Content:  "{{company}}",
					ZIndex:   5,
					Style: &pb.ElementStyle{
						FontSize:  26,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#333333",
					},
				},
				// Horizontal divider line
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_LINE,
					Position: &pb.Position{X: 100, Y: 320},
					Size:     &pb.Size{Width: 816, Height: 2},
					ZIndex:   6,
					Style: &pb.ElementStyle{
						BorderWidth: 2,
						BorderColor: "#CCCCCC",
					},
				},
				// Host label
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 350},
					Size:     &pb.Size{Width: 916, Height: 35},
					Content:  "Visiting:",
					ZIndex:   7,
					Style: &pb.ElementStyle{
						FontSize:  20,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#666666",
					},
				},
				// Host name
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 395},
					Size:     &pb.Size{Width: 916, Height: 40},
					Content:  "{{host}}",
					ZIndex:   8,
					Style: &pb.ElementStyle{
						FontSize:  24,
						Bold:      true,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#000000",
					},
				},
				// Date
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 460},
					Size:     &pb.Size{Width: 916, Height: 35},
					Content:  "Date: {{date}}",
					ZIndex:   9,
					Style: &pb.ElementStyle{
						FontSize:  22,
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
						Color:     "#666666",
					},
				},
				// Expiry notice (red, bold)
				{
					Type:     pb.BadgeElement_ELEMENT_TYPE_TEXT,
					Position: &pb.Position{X: 50, Y: 540},
					Size:     &pb.Size{Width: 916, Height: 60},
					Content:  "EXPIRES {{expires}}",
					ZIndex:   10,
					Style: &pb.ElementStyle{
						FontSize:  28,
						Bold:      true,
						Color:     "#FF0000",
						Alignment: pb.ElementStyle_TEXT_ALIGN_CENTER,
					},
				},
			},
		},
	}
}

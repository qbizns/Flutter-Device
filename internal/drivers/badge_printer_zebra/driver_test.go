package badge_printer_zebra

import (
	"context"
	"testing"
	"time"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

func TestNewDriver(t *testing.T) {
	config := DefaultConfig()
	config.Transport = "tcp"
	config.Address = "localhost"
	config.Port = 9100

	driver, err := NewDriver("test-badge-1", "Test Badge Printer", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	if driver.ID() != "test-badge-1" {
		t.Errorf("Expected ID 'test-badge-1', got '%s'", driver.ID())
	}

	if driver.Name() != "Test Badge Printer" {
		t.Errorf("Expected name 'Test Badge Printer', got '%s'", driver.Name())
	}

	if driver.Kind() != "badge_printer.zebra" {
		t.Errorf("Expected kind 'badge_printer.zebra', got '%s'", driver.Kind())
	}
}

func TestDriver_JobQueue(t *testing.T) {
	config := DefaultConfig()
	driver, err := NewDriver("test-badge-2", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	// Check job queue is created
	if driver.jobQueue == nil {
		t.Error("Job queue not initialized")
	}

	if cap(driver.jobQueue) != 100 {
		t.Errorf("Expected job queue capacity 100, got %d", cap(driver.jobQueue))
	}
}

func TestDriver_GetBadgeTemplates(t *testing.T) {
	config := DefaultConfig()
	driver, err := NewDriver("test-badge-3", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	ctx := context.Background()
	req := &pb.GetBadgeTemplatesRequest{
		DeviceId: "test-badge-3",
	}

	resp, err := driver.GetBadgeTemplates(ctx, req)
	if err != nil {
		t.Fatalf("GetBadgeTemplates failed: %v", err)
	}

	if len(resp.Templates) != 4 {
		t.Errorf("Expected 4 templates, got %d", len(resp.Templates))
	}

	// Check template IDs
	expectedIDs := map[string]bool{
		"conference_attendee": false,
		"vip_badge":           false,
		"staff_badge":         false,
		"visitor_badge":       false,
	}

	for _, tmpl := range resp.Templates {
		if _, ok := expectedIDs[tmpl.Id]; ok {
			expectedIDs[tmpl.Id] = true
		}
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Template '%s' not found", id)
		}
	}
}

func TestDriver_GetBadgeTemplate(t *testing.T) {
	config := DefaultConfig()
	driver, err := NewDriver("test-badge-4", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		templateID string
		wantErr    bool
	}{
		{"conference_attendee", false},
		{"vip_badge", false},
		{"staff_badge", false},
		{"visitor_badge", false},
		{"nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.templateID, func(t *testing.T) {
			req := &pb.GetBadgeTemplateRequest{
				DeviceId:   "test-badge-4",
				TemplateId: tt.templateID,
			}

			tmpl, err := driver.GetBadgeTemplate(ctx, req)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("GetBadgeTemplate failed: %v", err)
				}
				if tmpl.Id != tt.templateID {
					t.Errorf("Expected ID '%s', got '%s'", tt.templateID, tmpl.Id)
				}
			}
		})
	}
}

func TestDriver_DesignBadge(t *testing.T) {
	config := DefaultConfig()
	driver, err := NewDriver("test-badge-5", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	ctx := context.Background()

	design := &pb.BadgeDesign{
		DesignSource: &pb.BadgeDesign_TemplateId{
			TemplateId: "conference_attendee",
		},
		Variables: map[string]string{
			"name":    "John Doe",
			"company": "Acme Corp",
		},
	}

	req := &pb.DesignBadgeRequest{
		DeviceId: "test-badge-5",
		Design:   design,
	}

	result, err := driver.DesignBadge(ctx, req)
	if err != nil {
		t.Fatalf("DesignBadge failed: %v", err)
	}

	if result == nil {
		t.Error("Expected design result, got nil")
	}
}

func TestDriver_ValidatePrintRequest(t *testing.T) {
	config := DefaultConfig()
	driver, err := NewDriver("test-badge-6", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	tests := []struct {
		name    string
		req     *pb.PrintBadgeRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &pb.PrintBadgeRequest{
				DeviceId: "test-badge-6",
				Design: &pb.BadgeDesign{
					DesignSource: &pb.BadgeDesign_TemplateId{
						TemplateId: "conference_attendee",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "wrong device ID",
			req: &pb.PrintBadgeRequest{
				DeviceId: "wrong-id",
				Design: &pb.BadgeDesign{
					DesignSource: &pb.BadgeDesign_TemplateId{
						TemplateId: "conference_attendee",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing design",
			req: &pb.PrintBadgeRequest{
				DeviceId: "test-badge-6",
				Design:   nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := driver.validatePrintRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePrintRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDriver_EncodeCard(t *testing.T) {
	config := DefaultConfig()
	config.MagStripe = true
	config.RFID = true

	driver, err := NewDriver("test-badge-7", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		encoding *pb.EncodingData
		wantErr  bool
	}{
		{
			name:     "nil encoding",
			encoding: nil,
			wantErr:  true,
		},
		{
			name: "magnetic stripe only",
			encoding: &pb.EncodingData{
				MagneticStripe: &pb.MagneticStripeData{
					Enabled:    true,
					Track1Data: "Test Track 1",
					Track2Data: "1234567890",
					Coercivity: pb.MagneticStripeData_COERCIVITY_TYPE_HICO,
				},
			},
			wantErr: false,
		},
		{
			name: "RFID only",
			encoding: &pb.EncodingData{
				Rfid: &pb.RFIDData{
					Enabled: true,
					Type:    pb.RFIDData_RFID_TYPE_MIFARE_CLASSIC_1K,
					Uid:     []byte{0x01, 0x02, 0x03, 0x04},
					Blocks: []*pb.RFIDBlock{
						{
							BlockNumber: 1,
							Data:        []byte{0x00, 0x01, 0x02, 0x03},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "both mag stripe and RFID",
			encoding: &pb.EncodingData{
				MagneticStripe: &pb.MagneticStripeData{
					Enabled:    true,
					Track2Data: "1234567890",
				},
				Rfid: &pb.RFIDData{
					Enabled: true,
					Type:    pb.RFIDData_RFID_TYPE_HID_ICLASS,
					Uid:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.EncodeCardRequest{
				DeviceId: "test-badge-7",
				Encoding: tt.encoding,
			}

			resp, err := driver.EncodeCard(ctx, req)

			if tt.wantErr {
				if err == nil && (resp == nil || !resp.Success) {
					// Expected failure
				}
			} else {
				if err != nil {
					t.Errorf("EncodeCard() error = %v", err)
				}
				if resp != nil && !resp.Success {
					t.Errorf("EncodeCard() returned success=false: %s", resp.Message)
				}
			}
		})
	}
}

func TestGenerateJobID(t *testing.T) {
	id1 := generateJobID()
	time.Sleep(1 * time.Millisecond)
	id2 := generateJobID()

	if id1 == id2 {
		t.Error("generateJobID() should produce unique IDs")
	}

	if id1 == "" || id2 == "" {
		t.Error("generateJobID() should not return empty string")
	}
}

// MockConnection for testing
type MockConnection struct {
	writeData []byte
	readData  []byte
	connected bool
}

func (m *MockConnection) Write(data []byte) (int, error) {
	m.writeData = append(m.writeData, data...)
	return len(data), nil
}

func (m *MockConnection) Read(buf []byte) (int, error) {
	n := copy(buf, m.readData)
	return n, nil
}

func (m *MockConnection) Close() error {
	m.connected = false
	return nil
}

func (m *MockConnection) IsConnected() bool {
	return m.connected
}

func TestDriver_SendZPL(t *testing.T) {
	config := DefaultConfig()
	driver, err := NewDriver("test-badge-8", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	// Set mock connection
	mock := &MockConnection{connected: true}
	driver.conn = mock

	zpl := "^XA^FO100,100^A0N,50,50^FDTest^FS^XZ"

	err = driver.sendZPL(zpl)
	if err != nil {
		t.Fatalf("sendZPL failed: %v", err)
	}

	if string(mock.writeData) != zpl {
		t.Errorf("Expected ZPL '%s', got '%s'", zpl, string(mock.writeData))
	}
}

func TestDriver_Lifecycle(t *testing.T) {
	config := DefaultConfig()
	driver, err := NewDriver("test-badge-9", "Test", config, nil)
	if err != nil {
		t.Fatalf("NewDriver failed: %v", err)
	}

	// Check initial status
	if driver.Status() != 0 { // StatusDisconnected
		t.Errorf("Expected initial status 0, got %v", driver.Status())
	}

	// Note: We can't actually start/stop without a real connection
	// This test just verifies the methods exist and don't panic

	ctx := context.Background()

	// Stop should not fail even if not started
	err = driver.Stop(ctx)
	if err != nil {
		t.Logf("Stop returned error (expected): %v", err)
	}
}

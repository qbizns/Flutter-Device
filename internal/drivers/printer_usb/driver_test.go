package printer_usb

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// mockUSBDevice implements USBDevice for testing
type mockUSBDevice struct {
	mu            sync.Mutex
	writeBuffer   []byte
	readBuffer    []byte
	readPos       int
	closed        bool
	vendorID      uint16
	productID     uint16
	serial        string
	writeErr      error
	readErr       error
	resetErr      error
	writeDelay    time.Duration
	failNextWrite bool
}

func newMockUSBDevice() *mockUSBDevice {
	return &mockUSBDevice{
		vendorID:  0x04b8, // Epson
		productID: 0x0e15, // TM-T88V
		serial:    "TEST123",
	}
}

func (m *mockUSBDevice) Write(data []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, errors.New("device closed")
	}

	if m.failNextWrite {
		m.failNextWrite = false
		return 0, errors.New("write failed")
	}

	if m.writeErr != nil {
		return 0, m.writeErr
	}

	if m.writeDelay > 0 {
		time.Sleep(m.writeDelay)
	}

	m.writeBuffer = append(m.writeBuffer, data...)
	return len(data), nil
}

func (m *mockUSBDevice) Read(buffer []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return 0, errors.New("device closed")
	}

	if m.readErr != nil {
		return 0, m.readErr
	}

	if m.readPos >= len(m.readBuffer) {
		return 0, errors.New("no data available")
	}

	n := copy(buffer, m.readBuffer[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockUSBDevice) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	return nil
}

func (m *mockUSBDevice) GetInfo() (vendorID, productID uint16, serial string, err error) {
	return m.vendorID, m.productID, m.serial, nil
}

func (m *mockUSBDevice) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return errors.New("device closed")
	}

	if m.resetErr != nil {
		return m.resetErr
	}

	m.writeBuffer = nil
	m.readBuffer = nil
	m.readPos = 0
	return nil
}

func (m *mockUSBDevice) getWriteBuffer() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]byte{}, m.writeBuffer...)
}

func (m *mockUSBDevice) setReadBuffer(data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.readBuffer = data
	m.readPos = 0
}

func TestNewDriver(t *testing.T) {
	config := Config{
		VendorID:  0x04b8,
		ProductID: 0x0e15,
		Timeout:   5 * time.Second,
	}

	logger := telemetry.NewLogger()
	driver := NewDriver("test-printer", "Test Printer", config, logger)

	if driver == nil {
		t.Fatal("NewDriver returned nil")
	}

	if driver.ID() != "test-printer" {
		t.Errorf("driver.ID() = %s, want test-printer", driver.ID())
	}

	if driver.Name() != "Test Printer" {
		t.Errorf("driver.Name() = %s, want Test Printer", driver.Name())
	}

	if driver.Kind() != "printer.usb" {
		t.Errorf("driver.Kind() = %s, want printer.usb", driver.Kind())
	}
}

func TestDriver_StartStop(t *testing.T) {
	config := DefaultConfig()
	config.VendorID = 0x04b8
	config.ProductID = 0x0e15

	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	// Create mock device
	mock := newMockUSBDevice()
	driver.device = mock
	driver.connected = true
	driver.vendorID = mock.vendorID
	driver.productID = mock.productID
	driver.serial = mock.serial

	// Test Stop
	err := driver.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if !mock.closed {
		t.Error("Device not closed after Stop")
	}

	if driver.device != nil {
		t.Error("Device reference not cleared after Stop")
	}
}

func TestDriver_Print(t *testing.T) {
	config := DefaultConfig()
	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	// Setup mock device
	mock := newMockUSBDevice()
	driver.device = mock
	driver.connected = true

	// Create simple print document
	doc := &printer.PrintDocument{
		Sections: []printer.Section{
			{
				Elements: []printer.Element{
					{
						Type: printer.ElementTypeLine,
						Line: &printer.Line{
							Alignment: printer.AlignCenter,
							Runs: []printer.Run{
								{Text: "Test Receipt", Style: printer.TextStyle{}},
							},
						},
					},
				},
			},
		},
		Options: printer.PrintOptions{Cut: true},
	}

	ctx := context.Background()
	err := driver.Print(ctx, doc)
	if err != nil {
		t.Fatalf("Print failed: %v", err)
	}

	// Verify data was written
	written := mock.getWriteBuffer()
	if len(written) == 0 {
		t.Error("No data written to device")
	}

	// Check for initialization sequence
	initSeq := []byte{0x1B, 0x40}
	if len(written) < len(initSeq) || written[0] != initSeq[0] || written[1] != initSeq[1] {
		t.Error("Missing initialization sequence in written data")
	}

	// Check for text content
	hasText := false
	for i := 0; i < len(written)-11; i++ {
		if string(written[i:i+12]) == "Test Receipt" {
			hasText = true
			break
		}
	}
	if !hasText {
		t.Error("Text content not found in written data")
	}

	// Check for cut command
	cutSeq := []byte{0x1D, 0x56, 0x00}
	hasCut := false
	for i := 0; i < len(written)-2; i++ {
		if written[i] == cutSeq[0] && written[i+1] == cutSeq[1] && written[i+2] == cutSeq[2] {
			hasCut = true
			break
		}
	}
	if !hasCut {
		t.Error("Cut command not found in written data")
	}
}

func TestDriver_Print_NotConnected(t *testing.T) {
	config := DefaultConfig()
	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	// Don't connect device
	driver.connected = false

	doc := &printer.PrintDocument{
		Sections: []printer.Section{},
		Options:  printer.PrintOptions{},
	}

	ctx := context.Background()
	err := driver.Print(ctx, doc)
	if err == nil {
		t.Error("Print should fail when not connected")
	}

	if err.Error() != "printer not connected" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDriver_OpenDrawer(t *testing.T) {
	config := DefaultConfig()
	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	// Setup mock device
	mock := newMockUSBDevice()
	driver.device = mock
	driver.connected = true

	ctx := context.Background()
	err := driver.OpenDrawer(ctx)
	if err != nil {
		t.Fatalf("OpenDrawer failed: %v", err)
	}

	// Verify drawer pulse command was sent
	written := mock.getWriteBuffer()
	drawerCmd := []byte{0x1B, 0x70, 0x00, 0x32, 0x32}

	if len(written) != len(drawerCmd) {
		t.Errorf("Written data length = %d, want %d", len(written), len(drawerCmd))
	}

	for i := 0; i < len(drawerCmd); i++ {
		if written[i] != drawerCmd[i] {
			t.Errorf("Byte %d = 0x%02x, want 0x%02x", i, written[i], drawerCmd[i])
		}
	}
}

func TestDriver_GetStatus(t *testing.T) {
	config := DefaultConfig()
	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	// Setup mock device
	mock := newMockUSBDevice()
	driver.device = mock
	driver.connected = true

	// Set mock status response
	// Status byte: 0x00 = online, paper present, drawer closed
	mock.setReadBuffer([]byte{0x00})

	ctx := context.Background()
	status, err := driver.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if !status.Online {
		t.Error("Status should be online")
	}

	if !status.PaperPresent {
		t.Error("Paper should be present")
	}

	if status.DrawerOpen {
		t.Error("Drawer should be closed")
	}

	// Verify status query was sent
	written := mock.getWriteBuffer()
	statusCmd := []byte{0x10, 0x04, 0x01}
	if len(written) < len(statusCmd) {
		t.Error("Status command not sent")
	}
}

func TestDriver_GetStatus_NotConnected(t *testing.T) {
	config := DefaultConfig()
	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	// Don't connect device
	driver.connected = false

	ctx := context.Background()
	status, err := driver.GetStatus(ctx)
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	// Should return offline status
	if status.Online {
		t.Error("Status should be offline when not connected")
	}
}

func TestDriver_WriteError(t *testing.T) {
	config := DefaultConfig()
	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	// Setup mock device with write error
	mock := newMockUSBDevice()
	mock.writeErr = errors.New("USB write failed")
	driver.device = mock
	driver.connected = true

	doc := &printer.PrintDocument{
		Sections: []printer.Section{},
		Options:  printer.PrintOptions{},
	}

	ctx := context.Background()
	err := driver.Print(ctx, doc)
	if err == nil {
		t.Error("Print should fail when write fails")
	}

	// Device should be closed and disconnected on write error
	if driver.device != nil {
		t.Error("Device should be nil after write error")
	}

	if driver.connected {
		t.Error("Should be disconnected after write error")
	}
}

func TestConfig_Defaults(t *testing.T) {
	config := DefaultConfig()

	if config.Timeout != 5*time.Second {
		t.Errorf("Default timeout = %v, want 5s", config.Timeout)
	}

	if config.ReconnectDelay != 5*time.Second {
		t.Errorf("Default reconnect delay = %v, want 5s", config.ReconnectDelay)
	}

	if config.Metadata == nil {
		t.Error("Default metadata should not be nil")
	}
}

func BenchmarkDriver_Print(b *testing.B) {
	config := DefaultConfig()
	logger := telemetry.NewLogger()
	driver := NewDriver("test", "Test", config, logger)

	mock := newMockUSBDevice()
	driver.device = mock
	driver.connected = true

	doc := &printer.PrintDocument{
		Sections: []printer.Section{
			{
				Elements: []printer.Element{
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

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = driver.Print(ctx, doc)
	}
}

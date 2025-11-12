package rfid_reader

import (
	"context"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock connection for testing
type mockConnection struct {
	connected bool
	writeData []byte
	readData  []byte
	readPos   int
}

func newMockConnection() *mockConnection {
	return &mockConnection{
		connected: true,
		readData:  make([]byte, 1024),
	}
}

func (m *mockConnection) Write(data []byte) (int, error) {
	m.writeData = append(m.writeData, data...)
	return len(data), nil
}

func (m *mockConnection) Read(buf []byte) (int, error) {
	n := copy(buf, m.readData[m.readPos:])
	m.readPos += n
	return n, nil
}

func (m *mockConnection) Close() error {
	m.connected = false
	return nil
}

func (m *mockConnection) IsConnected() bool {
	return m.connected
}

func (m *mockConnection) SetTimeout(timeout time.Duration) error {
	return nil
}

func TestNewDriver(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-reader-1"
	config.Name = "Test RFID Reader"
	config.SupportedProtocols = []string{"mifare", "ntag"}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)

	require.NoError(t, err)
	assert.NotNil(t, driver)
	assert.Equal(t, "test-reader-1", driver.id)
	assert.Equal(t, "Test RFID Reader", driver.name)
	assert.Len(t, driver.handlers, 5) // 2 Mifare + 3 NTAG types
}

func TestDriver_Lifecycle(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-reader-lifecycle"
	config.Transport = "usb"
	config.VendorID = 0x1234
	config.ProductID = 0x5678

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	// Test initial status
	assert.Equal(t, pb.ReaderMode_READER_MODE_CONTINUOUS, driver.mode)

	// Note: Start() will fail because no actual USB device is connected
	// In a real test environment, you would use a mock USB connection
	// For now, we test the structure

	// Test Stop (should not panic even if not started)
	err = driver.Stop()
	assert.NoError(t, err)
}

func TestDriver_GetReaderStatus(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-reader-status"
	config.SupportedProtocols = []string{"mifare", "prox"}
	config.Model = "ACR122U"
	config.Manufacturer = "ACS"

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	req := &pb.GetReaderStatusRequest{
		DeviceId: config.ID,
	}

	status, err := driver.GetReaderStatus(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.Equal(t, config.ID, status.DeviceId)
	assert.Equal(t, config.Model, status.Model)
	assert.Equal(t, config.Manufacturer, status.Manufacturer)
	assert.True(t, status.Capabilities.CanRead)
	assert.False(t, status.CardPresent)
	assert.NotEmpty(t, status.SupportedTypes)
}

func TestDriver_SetReaderMode(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-reader-mode"

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test changing mode
	req := &pb.SetReaderModeRequest{
		DeviceId: config.ID,
		Mode:     pb.ReaderMode_READER_MODE_SINGLE_READ,
	}

	resp, err := driver.SetReaderMode(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, pb.ReaderMode_READER_MODE_SINGLE_READ, resp.CurrentMode)
	assert.Equal(t, pb.ReaderMode_READER_MODE_SINGLE_READ, driver.mode)
}

func TestDriver_GetSupportedCardTypes(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-reader-types"
	config.SupportedProtocols = []string{"mifare", "ntag", "prox"}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	req := &pb.GetSupportedCardTypesRequest{
		DeviceId: config.ID,
	}

	resp, err := driver.GetSupportedCardTypes(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.CardTypes)
	assert.Contains(t, resp.Protocols, "mifare")
	assert.Contains(t, resp.Protocols, "ntag")
	assert.Contains(t, resp.Protocols, "prox")
}

func TestReaderStatistics(t *testing.T) {
	stats := &ReaderStatistics{
		StartedAt: time.Now(),
	}

	// Test read increments
	stats.IncrementReads(true, 50*time.Millisecond)
	assert.Equal(t, int64(1), stats.TotalReads)
	assert.Equal(t, int64(1), stats.SuccessfulReads)
	assert.Equal(t, int64(0), stats.FailedReads)

	stats.IncrementReads(false, 30*time.Millisecond)
	assert.Equal(t, int64(2), stats.TotalReads)
	assert.Equal(t, int64(1), stats.SuccessfulReads)
	assert.Equal(t, int64(1), stats.FailedReads)

	// Test average read time
	avgTime := stats.GetAvgReadTime()
	assert.Equal(t, int32(40), avgTime) // (50+30)/2 = 40ms

	// Test write increments
	stats.IncrementWrites()
	assert.Equal(t, int64(1), stats.TotalWrites)

	// Test auth increments
	stats.IncrementAuths(true)
	stats.IncrementAuths(false)
	assert.Equal(t, int64(1), stats.SuccessfulAuths)
	assert.Equal(t, int64(1), stats.FailedAuths)

	// Test proto conversion
	proto := stats.ToProto()
	assert.NotNil(t, proto)
	assert.Equal(t, int64(2), proto.TotalReads)
	assert.Equal(t, int64(1), proto.TotalWrites)
	assert.Equal(t, int32(40), proto.AvgReadTimeMs)
}

func TestCardPresenceTracker(t *testing.T) {
	tracker := &CardPresenceTracker{}

	// Initially no card present
	assert.False(t, tracker.IsCardPresent())

	// Set card present
	uid := []byte{0x01, 0x02, 0x03, 0x04}
	tracker.SetCardPresent(uid, pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K)
	assert.True(t, tracker.IsCardPresent())

	// Get current card
	currentUID, cardType := tracker.GetCurrentCard()
	assert.Equal(t, uid, currentUID)
	assert.Equal(t, pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K, cardType)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Remove card
	removedUID, duration := tracker.SetCardRemoved()
	assert.Equal(t, uid, removedUID)
	assert.Greater(t, duration, time.Duration(0))
	assert.False(t, tracker.IsCardPresent())
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, "usb", config.Transport)
	assert.True(t, config.CanRead)
	assert.True(t, config.HasBuzzer)
	assert.True(t, config.HasLED)
	assert.Equal(t, pb.ReaderMode_READER_MODE_CONTINUOUS, config.Mode)
	assert.Equal(t, int32(500), config.ReadIntervalMs)
	assert.NotNil(t, config.DefaultMifareKeyA)
	assert.Len(t, config.DefaultMifareKeyA, 6)
}

func TestProtocolHandlerRegistration(t *testing.T) {
	config := DefaultConfig()
	config.ID = "test-handlers"
	config.SupportedProtocols = []string{"prox", "mifare", "iclass", "ntag", "desfire", "felica"}

	logger := telemetry.NewLogger("test")
	driver, err := NewDriver(config.ID, config.Name, config, logger)
	require.NoError(t, err)

	// Check that handlers are registered
	assert.Contains(t, driver.handlers, pb.CardType_CARD_TYPE_HID_PROX)
	assert.Contains(t, driver.handlers, pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K)
	assert.Contains(t, driver.handlers, pb.CardType_CARD_TYPE_HID_ICLASS)
	assert.Contains(t, driver.handlers, pb.CardType_CARD_TYPE_NTAG_213)
	assert.Contains(t, driver.handlers, pb.CardType_CARD_TYPE_MIFARE_DESFIRE)
	assert.Contains(t, driver.handlers, pb.CardType_CARD_TYPE_FELICA)
}

func BenchmarkReaderStatistics_IncrementReads(b *testing.B) {
	stats := &ReaderStatistics{
		StartedAt: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.IncrementReads(true, 50*time.Millisecond)
	}
}

func BenchmarkCardPresenceTracker_SetCardPresent(b *testing.B) {
	tracker := &CardPresenceTracker{}
	uid := []byte{0x01, 0x02, 0x03, 0x04}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.SetCardPresent(uid, pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K)
	}
}

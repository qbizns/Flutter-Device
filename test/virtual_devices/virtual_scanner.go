package virtual_devices

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/scanner"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// VirtualScanner is a virtual barcode scanner for testing
type VirtualScanner struct {
	*devices.BaseDevice
	logger    *telemetry.Logger
	eventBus  *events.Bus
	events    chan scanner.ScanEvent
	config    scanner.ScannerConfig
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
}

// NewVirtualScanner creates a new virtual scanner
func NewVirtualScanner(id, name string, logger *telemetry.Logger, eventBus *events.Bus) *VirtualScanner {
	return &VirtualScanner{
		BaseDevice: devices.NewBaseDevice(id, "scanner.virtual", name, map[string]string{
			"virtual": "true",
			"type":    "scanner",
		}),
		logger:   logger,
		eventBus: eventBus,
		events:   make(chan scanner.ScanEvent, 10),
		config: scanner.ScannerConfig{
			Timeout: 30 * time.Second,
		},
	}
}

// Start initializes the virtual scanner
func (s *VirtualScanner) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("virtual scanner already running")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.running = true

	s.SetHealth(devices.HealthReady, "virtual scanner ready", nil)

	s.logger.Info("virtual scanner started",
		telemetry.String("device_id", s.ID()),
		telemetry.String("name", s.Name()),
	)

	// Start event generator for testing
	go s.eventGenerator()

	return nil
}

// Stop stops the virtual scanner
func (s *VirtualScanner) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	if s.cancel != nil {
		s.cancel()
	}

	s.running = false
	close(s.events)

	s.SetHealth(devices.HealthOffline, "virtual scanner stopped", nil)

	s.logger.Info("virtual scanner stopped",
		telemetry.String("device_id", s.ID()),
	)

	return nil
}

// Events returns a channel of scan events
func (s *VirtualScanner) Events(ctx context.Context) (<-chan scanner.ScanEvent, error) {
	return s.events, nil
}

// Configure applies scanner settings
func (s *VirtualScanner) Configure(ctx context.Context, config scanner.ScannerConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.config = config

	s.logger.Info("virtual scanner configured",
		telemetry.String("device_id", s.ID()),
		telemetry.String("prefix", config.Prefix),
		telemetry.String("suffix", config.Suffix),
	)

	return nil
}

// Scan simulates a barcode scan
func (s *VirtualScanner) Scan(barcode, symbology string) error {
	s.mu.RLock()
	running := s.running
	s.mu.RUnlock()

	if !running {
		return fmt.Errorf("scanner not running")
	}

	event := scanner.ScanEvent{
		DeviceID:  s.ID(),
		Data:      s.config.Prefix + barcode + s.config.Suffix,
		Symbology: symbology,
		Timestamp: time.Now(),
		EventID:   uuid.New().String(),
	}

	// Send to local event channel
	select {
	case s.events <- event:
	default:
		s.logger.Warn("scan event dropped, channel full",
			telemetry.String("device_id", s.ID()),
		)
	}

	// Publish to event bus
	if s.eventBus != nil {
		s.eventBus.Publish(events.Event{
			DeviceID: s.ID(),
			Type:     symbology,
			Data: map[string]interface{}{
				"barcode":   event.Data,
				"symbology": symbology,
				"timestamp": event.Timestamp,
			},
		})
	}

	s.logger.Info("barcode scanned",
		telemetry.String("device_id", s.ID()),
		telemetry.String("barcode", barcode),
		telemetry.String("symbology", symbology),
	)

	fmt.Printf("\n")
	fmt.Printf("===== VIRTUAL SCANNER EVENT =====\n")
	fmt.Printf("Device:    %s\n", s.Name())
	fmt.Printf("Barcode:   %s\n", event.Data)
	fmt.Printf("Symbology: %s\n", symbology)
	fmt.Printf("Timestamp: %s\n", event.Timestamp.Format(time.RFC3339))
	fmt.Printf("=================================\n")
	fmt.Printf("\n")

	return nil
}

// eventGenerator generates test scan events periodically
func (s *VirtualScanner) eventGenerator() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	testBarcodes := []struct {
		barcode   string
		symbology string
	}{
		{"123456789012", "EAN13"},
		{"1234567890", "CODE128"},
		{"ABCD1234", "CODE39"},
		{"978014300723", "EAN13"},
		{"00012345678905", "ITF14"},
	}

	index := 0

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			// Automatically generate a scan event every 15 seconds for testing
			bc := testBarcodes[index]
			index = (index + 1) % len(testBarcodes)

			s.Scan(bc.barcode, bc.symbology)
		}
	}
}

// ScanBatch simulates scanning multiple barcodes
func (s *VirtualScanner) ScanBatch(barcodes []string, symbology string, delay time.Duration) error {
	for _, barcode := range barcodes {
		if err := s.Scan(barcode, symbology); err != nil {
			return err
		}
		if delay > 0 {
			time.Sleep(delay)
		}
	}
	return nil
}

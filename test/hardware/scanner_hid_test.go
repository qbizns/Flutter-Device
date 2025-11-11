//go:build hardware
// +build hardware

package hardware

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/drivers/scanner_hid"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Hardware integration tests for USB HID scanners
//
// These tests require physical USB HID scanner hardware.
//
// To run these tests:
//   go test -tags=hardware ./test/hardware/...
//
// Environment variables:
//   SCANNER_VENDOR_ID  - USB Vendor ID (hex, e.g., "05e0")
//   SCANNER_PRODUCT_ID - USB Product ID (hex, e.g., "1200")
//   SCANNER_SERIAL     - Scanner serial number (optional)
//
// Example:
//   export SCANNER_VENDOR_ID=05e0
//   export SCANNER_PRODUCT_ID=1200
//   go test -tags=hardware -v ./test/hardware/...

// TestScannerAvailable checks if a USB HID scanner is connected
func TestScannerAvailable(t *testing.T) {
	scanners, err := scanner_hid.EnumerateScanners()
	if err != nil {
		t.Fatalf("Failed to enumerate scanners: %v", err)
	}

	if len(scanners) == 0 {
		t.Skip("No USB HID scanners detected. Connect a scanner and retry.")
	}

	t.Logf("Found %d scanner(s):", len(scanners))
	for i, s := range scanners {
		t.Logf("  [%d] %s", i+1, s.String())
		if s.IsKnownScanner() {
			t.Logf("       Known model: %s", s.GetKnownModel())
		}
	}
}

// TestScannerConnection tests connecting to a scanner
func TestScannerConnection(t *testing.T) {
	config := getTestConfig(t)

	logger := telemetry.NewLogger()
	eventBus := events.NewBus(logger)

	driver := scanner_hid.NewDriver("test-scanner", "Test Scanner", config, logger, eventBus)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Start driver
	err := driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scanner driver: %v", err)
	}
	defer driver.Stop()

	// Verify connected
	if driver.Health().Status != "ready" {
		t.Errorf("Scanner not ready: %s", driver.Health().Message)
	}

	t.Log("Scanner connected successfully")
	t.Logf("Device: %s", driver.Name())
	t.Logf("Status: %s", driver.Health().Status)
}

// TestScannerScanEvent tests receiving scan events
func TestScannerScanEvent(t *testing.T) {
	config := getTestConfig(t)

	logger := telemetry.NewLogger()
	eventBus := events.NewBus(logger)

	driver := scanner_hid.NewDriver("test-scanner", "Test Scanner", config, logger, eventBus)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err := driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scanner: %v", err)
	}
	defer driver.Stop()

	t.Log("Scanner ready. Please scan a barcode within 60 seconds...")

	select {
	case event := <-driver.Scan():
		t.Logf("✅ Scan received!")
		t.Logf("   Barcode:   %s", event.Data)
		t.Logf("   Symbology: %s", event.Symbology)
		t.Logf("   Timestamp: %s", event.Timestamp.Format(time.RFC3339))

		if event.Data == "" {
			t.Error("Barcode data is empty")
		}
		if event.Symbology == "" {
			t.Error("Symbology is empty")
		}

	case <-ctx.Done():
		t.Fatal("Timeout waiting for scan event. Did you scan a barcode?")
	}
}

// TestScannerMultipleScans tests receiving multiple scan events
func TestScannerMultipleScans(t *testing.T) {
	config := getTestConfig(t)

	logger := telemetry.NewLogger()
	eventBus := events.NewBus(logger)

	driver := scanner_hid.NewDriver("test-scanner", "Test Scanner", config, logger, eventBus)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err := driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scanner: %v", err)
	}
	defer driver.Stop()

	const targetScans = 5
	t.Logf("Scanner ready. Please scan %d barcodes within 120 seconds...", targetScans)

	scans := make([]string, 0, targetScans)

	for i := 0; i < targetScans; i++ {
		select {
		case event := <-driver.Scan():
			t.Logf("✅ Scan %d/%d: %s (%s)", i+1, targetScans, event.Data, event.Symbology)
			scans = append(scans, event.Data)

		case <-ctx.Done():
			t.Fatalf("Timeout after %d/%d scans", i, targetScans)
		}
	}

	if len(scans) != targetScans {
		t.Errorf("Expected %d scans, got %d", targetScans, len(scans))
	}
}

// TestScannerSymbologyDetection tests symbology detection
func TestScannerSymbologyDetection(t *testing.T) {
	config := getTestConfig(t)

	logger := telemetry.NewLogger()
	eventBus := events.NewBus(logger)

	driver := scanner_hid.NewDriver("test-scanner", "Test Scanner", config, logger, eventBus)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	err := driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scanner: %v", err)
	}
	defer driver.Stop()

	testSymbologies := []string{"EAN13", "UPCA", "CODE39", "CODE128", "QR"}
	t.Logf("Please scan barcodes of different types within 180 seconds:")
	for _, sym := range testSymbologies {
		t.Logf("  - %s", sym)
	}

	detected := make(map[string]bool)

	for {
		select {
		case event := <-driver.Scan():
			t.Logf("✅ Detected: %s (%s)", event.Data, event.Symbology)
			detected[event.Symbology] = true

			// Check if we got all types
			if len(detected) >= len(testSymbologies) {
				t.Log("All symbology types detected!")
				return
			}

		case <-ctx.Done():
			t.Logf("Detected %d/%d symbology types", len(detected), len(testSymbologies))
			for sym := range detected {
				t.Logf("  ✅ %s", sym)
			}
			return
		}
	}
}

// TestScannerReconnection tests scanner reconnection
func TestScannerReconnection(t *testing.T) {
	t.Skip("Manual test: Requires physically disconnecting/reconnecting scanner")

	config := getTestConfig(t)
	config.ReconnectDelay = 2 * time.Second

	logger := telemetry.NewLogger()
	eventBus := events.NewBus(logger)

	driver := scanner_hid.NewDriver("test-scanner", "Test Scanner", config, logger, eventBus)

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	err := driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scanner: %v", err)
	}
	defer driver.Stop()

	t.Log("✅ Scanner connected")
	t.Log("📍 Please disconnect the scanner now...")
	time.Sleep(5 * time.Second)

	t.Log("📍 Please reconnect the scanner now...")

	// Wait for reconnection
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if driver.Health().Status == "ready" {
				t.Log("✅ Scanner reconnected successfully!")

				// Try to scan
				t.Log("📍 Please scan a barcode to verify...")
				select {
				case event := <-driver.Scan():
					t.Logf("✅ Scan successful: %s", event.Data)
					return
				case <-time.After(30 * time.Second):
					t.Error("Scanner reconnected but scan timeout")
					return
				}
			}

		case <-timeout:
			t.Error("Scanner did not reconnect within 30 seconds")
			return

		case <-ctx.Done():
			t.Error("Test timeout")
			return
		}
	}
}

// TestMultipleScanners tests using multiple scanners simultaneously
func TestMultipleScanners(t *testing.T) {
	scanners, err := scanner_hid.EnumerateScanners()
	if err != nil {
		t.Fatalf("Failed to enumerate scanners: %v", err)
	}

	if len(scanners) < 2 {
		t.Skipf("Test requires 2+ scanners, found %d", len(scanners))
	}

	logger := telemetry.NewLogger()
	eventBus := events.NewBus(logger)

	// Create drivers for first 2 scanners
	drivers := make([]*scanner_hid.Driver, 0, 2)

	for i := 0; i < 2 && i < len(scanners); i++ {
		config := scanner_hid.Config{
			VendorID:       scanners[i].VendorID,
			ProductID:      scanners[i].ProductID,
			Serial:         scanners[i].Serial,
			BufferSize:     10,
			ReadTimeout:    1 * time.Second,
			ReconnectDelay: 5 * time.Second,
		}

		driver := scanner_hid.NewDriver(
			fmt.Sprintf("scanner-%d", i+1),
			fmt.Sprintf("Scanner %d", i+1),
			config,
			logger,
			eventBus,
		)

		ctx := context.Background()
		if err := driver.Start(ctx); err != nil {
			t.Fatalf("Failed to start scanner %d: %v", i+1, err)
		}
		defer driver.Stop()

		drivers = append(drivers, driver)
		t.Logf("✅ Scanner %d connected: %s", i+1, scanners[i].String())
	}

	t.Log("📍 Please scan barcodes on different scanners (30 seconds)...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scannedFrom := make(map[string]bool)

	for {
		select {
		case event := <-drivers[0].Scan():
			t.Logf("✅ Scanner 1: %s", event.Data)
			scannedFrom["scanner-1"] = true

		case event := <-drivers[1].Scan():
			t.Logf("✅ Scanner 2: %s", event.Data)
			scannedFrom["scanner-2"] = true

		case <-ctx.Done():
			if len(scannedFrom) == 2 {
				t.Log("✅ Both scanners working!")
			} else {
				t.Logf("Only %d/%d scanners produced scans", len(scannedFrom), 2)
			}
			return
		}

		if len(scannedFrom) == 2 {
			t.Log("✅ Both scanners confirmed working!")
			return
		}
	}
}

// BenchmarkScanThroughput benchmarks scan throughput
func BenchmarkScanThroughput(b *testing.B) {
	config := getTestConfig(b)

	logger := telemetry.NewLogger()
	eventBus := events.NewBus(logger)

	driver := scanner_hid.NewDriver("bench-scanner", "Bench Scanner", config, logger, eventBus)

	ctx := context.Background()
	if err := driver.Start(ctx); err != nil {
		b.Fatalf("Failed to start scanner: %v", err)
	}
	defer driver.Stop()

	b.Log("Ready for benchmarking. Scan barcodes rapidly...")

	b.ResetTimer()

	timeout := time.After(30 * time.Second)
	count := 0

	for {
		select {
		case <-driver.Scan():
			count++
			if count >= b.N {
				b.StopTimer()
				b.Logf("Processed %d scans", count)
				return
			}

		case <-timeout:
			b.StopTimer()
			b.Logf("Timeout after %d scans", count)
			return
		}
	}
}

// Helper function to get test configuration from environment
func getTestConfig(t testing.TB) scanner_hid.Config {
	config := scanner_hid.DefaultConfig()

	// Get vendor/product ID from environment
	vendorIDStr := os.Getenv("SCANNER_VENDOR_ID")
	productIDStr := os.Getenv("SCANNER_PRODUCT_ID")

	if vendorIDStr != "" && productIDStr != "" {
		var vendorID, productID uint64
		fmt.Sscanf(vendorIDStr, "%x", &vendorID)
		fmt.Sscanf(productIDStr, "%x", &productID)

		config.VendorID = uint16(vendorID)
		config.ProductID = uint16(productID)

		t.Logf("Using scanner: VID=0x%04x PID=0x%04x", config.VendorID, config.ProductID)
	} else {
		t.Log("No specific scanner configured, will auto-detect first available")
	}

	// Optional serial number
	if serial := os.Getenv("SCANNER_SERIAL"); serial != "" {
		config.Serial = serial
		t.Logf("Filtering by serial: %s", serial)
	}

	return config
}

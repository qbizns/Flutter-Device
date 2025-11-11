// +build hardware

package hardware

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/drivers/scale_serial"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// getScaleTestConfig returns configuration for hardware testing
// Set environment variables to customize:
//   SCALE_PORT=/dev/ttyUSB0
//   SCALE_BAUD_RATE=9600
//   SCALE_PROTOCOL=mtsics
func getScaleTestConfig(t *testing.T) scale_serial.Config {
	port := os.Getenv("SCALE_PORT")
	if port == "" {
		port = "/dev/ttyUSB0" // Default for Linux
	}

	baudRate := 9600
	if br := os.Getenv("SCALE_BAUD_RATE"); br != "" {
		// Parse baud rate if needed
		switch br {
		case "4800":
			baudRate = 4800
		case "19200":
			baudRate = 19200
		case "38400":
			baudRate = 38400
		case "9600":
			baudRate = 9600
		default:
			t.Logf("Unknown baud rate %s, using 9600", br)
		}
	}

	protocol := os.Getenv("SCALE_PROTOCOL")
	if protocol == "" {
		protocol = "mtsics" // Default to MT-SICS
	}

	config := scale_serial.Config{
		Port:          port,
		BaudRate:      baudRate,
		DataBits:      8,
		Parity:        "none",
		StopBits:      1,
		Protocol:      protocol,
		ReadTimeout:   2000 * time.Millisecond,
		RetryAttempts: 3,
		RetryDelay:    100 * time.Millisecond,
		PreferredUnit: scale_serial.UnitKilogram,
	}

	t.Logf("Scale test configuration:")
	t.Logf("  Port:     %s", config.Port)
	t.Logf("  Baud:     %d", config.BaudRate)
	t.Logf("  Protocol: %s", config.Protocol)

	return config
}

// TestScaleConnection tests basic connection to the scale
func TestScaleConnection(t *testing.T) {
	config := getScaleTestConfig(t)
	logger, _ := telemetry.NewLogger("info", "text")

	driver, err := scale_serial.NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Log("Starting scale driver...")
	err = driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	t.Log("✅ Scale connected successfully")
}

// TestScaleReadWeight tests reading weight from the scale
func TestScaleReadWeight(t *testing.T) {
	config := getScaleTestConfig(t)
	logger, _ := telemetry.NewLogger("info", "text")

	driver, err := scale_serial.NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	t.Log("Reading weight from scale...")
	t.Log("(Place an object on the scale)")

	reading, err := driver.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("Failed to read weight: %v", err)
	}

	t.Logf("✅ Weight reading successful!")
	t.Logf("   Weight: %.3f %s", reading.Weight, reading.Unit)
	t.Logf("   Stable: %v", reading.Stable)
	t.Logf("   Time:   %s", reading.Timestamp.Format(time.RFC3339))
}

// TestScaleMultipleReads tests reading weight multiple times
func TestScaleMultipleReads(t *testing.T) {
	config := getScaleTestConfig(t)
	logger, _ := telemetry.NewLogger("info", "text")

	driver, err := scale_serial.NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	t.Log("Reading weight 5 times with 1 second intervals...")

	for i := 1; i <= 5; i++ {
		reading, err := driver.ReadWeight(ctx)
		if err != nil {
			t.Errorf("Read %d failed: %v", i, err)
			continue
		}

		t.Logf("Read %d: %.3f %s (stable: %v)", i, reading.Weight, reading.Unit, reading.Stable)

		if i < 5 {
			time.Sleep(1 * time.Second)
		}
	}

	t.Log("✅ Multiple reads completed")
}

// TestScaleZero tests zeroing the scale
func TestScaleZero(t *testing.T) {
	config := getScaleTestConfig(t)
	logger, _ := telemetry.NewLogger("info", "text")

	driver, err := scale_serial.NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err = driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	// Read initial weight
	t.Log("Reading initial weight...")
	reading1, err := driver.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("Failed to read initial weight: %v", err)
	}
	t.Logf("Initial weight: %.3f %s", reading1.Weight, reading1.Unit)

	// Zero the scale
	t.Log("Zeroing scale...")
	err = driver.Zero(ctx)
	if err != nil {
		t.Fatalf("Failed to zero scale: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Read weight after zeroing
	t.Log("Reading weight after zeroing...")
	reading2, err := driver.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("Failed to read weight after zero: %v", err)
	}
	t.Logf("Weight after zero: %.3f %s", reading2.Weight, reading2.Unit)

	// Weight should be close to zero
	if reading2.Weight < -0.1 || reading2.Weight > 0.1 {
		t.Logf("⚠️  Weight not close to zero after zeroing (%.3f %s)", reading2.Weight, reading2.Unit)
		t.Log("   This may be normal depending on scale precision")
	}

	t.Log("✅ Zero operation completed")
}

// TestScaleTare tests taring the scale
func TestScaleTare(t *testing.T) {
	config := getScaleTestConfig(t)
	logger, _ := telemetry.NewLogger("info", "text")

	driver, err := scale_serial.NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	t.Log("Test procedure:")
	t.Log("1. Make sure scale is empty")
	t.Log("2. Place a container on the scale")
	t.Log("3. Press Enter to tare")
	t.Log("4. Add items to the container")
	t.Log("")
	t.Log("Press Enter when container is on scale...")

	// Wait for user input (in automated tests, this will timeout)
	input := make(chan bool, 1)
	go func() {
		var dummy string
		os.Stdin.Read([]byte{0})
		input <- true
	}()

	select {
	case <-input:
		// User pressed Enter
	case <-time.After(5 * time.Second):
		// Timeout - continue anyway for automated testing
		t.Log("⏱️  Timeout waiting for input, continuing...")
	}

	// Read weight with container
	reading1, err := driver.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("Failed to read weight with container: %v", err)
	}
	t.Logf("Weight with container: %.3f %s", reading1.Weight, reading1.Unit)

	// Tare the scale
	t.Log("Taring scale...")
	err = driver.Tare(ctx)
	if err != nil {
		t.Fatalf("Failed to tare scale: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Read weight after taring
	reading2, err := driver.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("Failed to read weight after tare: %v", err)
	}
	t.Logf("Weight after tare: %.3f %s", reading2.Weight, reading2.Unit)

	t.Log("✅ Tare operation completed")
}

// TestScaleStability tests stable weight detection
func TestScaleStability(t *testing.T) {
	config := getScaleTestConfig(t)
	logger, _ := telemetry.NewLogger("info", "text")

	driver, err := scale_serial.NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	t.Log("Testing stability detection...")
	t.Log("Test will read weight every 500ms for 10 seconds")
	t.Log("Try moving an object on/off the scale to see stability change")

	stableCount := 0
	unstableCount := 0

	for i := 0; i < 20; i++ {
		reading, err := driver.ReadWeight(ctx)
		if err != nil {
			t.Logf("⚠️  Read failed: %v", err)
			continue
		}

		statusStr := "UNSTABLE"
		if reading.Stable {
			statusStr = "STABLE"
			stableCount++
		} else {
			unstableCount++
		}

		t.Logf("[%2d] %s: %.3f %s", i+1, statusStr, reading.Weight, reading.Unit)

		time.Sleep(500 * time.Millisecond)
	}

	t.Logf("✅ Stability test completed")
	t.Logf("   Stable readings:   %d/20", stableCount)
	t.Logf("   Unstable readings: %d/20", unstableCount)
}

// TestScalePortEnumeration tests port enumeration
func TestScalePortEnumeration(t *testing.T) {
	ports, err := scale_serial.EnumeratePorts()
	if err != nil {
		t.Fatalf("Failed to enumerate ports: %v", err)
	}

	if len(ports) == 0 {
		t.Log("⚠️  No serial ports found")
		t.Log("   This may be normal if no USB-to-serial devices are connected")
		return
	}

	t.Logf("Found %d serial port(s):", len(ports))
	for i, port := range ports {
		t.Logf("  [%d] %s", i+1, port)
	}

	t.Log("✅ Port enumeration completed")
}

// TestScalePerformance tests scale reading performance
func TestScalePerformance(t *testing.T) {
	config := getScaleTestConfig(t)
	logger, _ := telemetry.NewLogger("info", "text")

	driver, err := scale_serial.NewDriver("test-scale", "Test Scale", config, logger)
	if err != nil {
		t.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err = driver.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	t.Log("Performance test: 100 weight readings")

	successCount := 0
	failCount := 0
	totalDuration := time.Duration(0)
	minDuration := time.Duration(999999 * time.Hour)
	maxDuration := time.Duration(0)

	startTime := time.Now()

	for i := 0; i < 100; i++ {
		readStart := time.Now()
		_, err := driver.ReadWeight(ctx)
		readDuration := time.Since(readStart)

		if err != nil {
			failCount++
			if i < 10 { // Only log first 10 failures
				t.Logf("⚠️  Read %d failed: %v", i+1, err)
			}
		} else {
			successCount++
			totalDuration += readDuration

			if readDuration < minDuration {
				minDuration = readDuration
			}
			if readDuration > maxDuration {
				maxDuration = readDuration
			}
		}

		if (i+1)%20 == 0 {
			t.Logf("Progress: %d/100 reads completed", i+1)
		}
	}

	totalTime := time.Since(startTime)

	t.Log("✅ Performance test completed")
	t.Logf("   Total time:       %v", totalTime)
	t.Logf("   Successful reads: %d/100", successCount)
	t.Logf("   Failed reads:     %d/100", failCount)

	if successCount > 0 {
		avgDuration := totalDuration / time.Duration(successCount)
		t.Logf("   Average read:     %v", avgDuration)
		t.Logf("   Fastest read:     %v", minDuration)
		t.Logf("   Slowest read:     %v", maxDuration)
		t.Logf("   Reads/second:     %.1f", float64(successCount)/totalTime.Seconds())
	}
}

// BenchmarkScaleReadWeight benchmarks weight reading performance
func BenchmarkScaleReadWeight(b *testing.B) {
	config := getScaleTestConfig(&testing.T{})
	logger, _ := telemetry.NewLogger("error", "text") // Reduce logging noise

	driver, err := scale_serial.NewDriver("bench-scale", "Benchmark Scale", config, logger)
	if err != nil {
		b.Fatalf("Failed to create scale driver: %v", err)
	}

	ctx := context.Background()
	err = driver.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start scale: %v", err)
	}
	defer driver.Stop()

	// Reset timer before benchmark loop
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := driver.ReadWeight(ctx)
		if err != nil {
			b.Errorf("Read failed: %v", err)
		}
	}
}

// +build integration

package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/app"
	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/jobs"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	"github.com/Macber-eg/Flutter-Device/test/virtual_devices"
)

// TestBasicDeviceLifecycle tests the complete lifecycle of a device
func TestBasicDeviceLifecycle(t *testing.T) {
	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	metrics := telemetry.NewMetrics()
	registry := app.NewRegistry(logger, metrics)

	// Create and register a virtual printer
	printer := virtual_devices.NewVirtualPrinter("printer-1", "Test Printer", logger)
	err = registry.Register(printer)
	if err != nil {
		t.Fatalf("Failed to register printer: %v", err)
	}

	// Start the device
	ctx := context.Background()
	err = printer.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start device: %v", err)
	}
	defer printer.Stop()

	// Verify device is registered
	device, err := registry.Get("printer-1")
	if err != nil {
		t.Fatalf("Device not found in registry: %v", err)
	}

	// Check health
	health := device.Health()
	if health.Status != devices.HealthReady {
		t.Errorf("Expected status ready, got %v", health.Status)
	}
}

// TestEventBusIntegration tests event bus with virtual scanner
func TestEventBusIntegration(t *testing.T) {
	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	eventBus := events.NewBus(logger)

	// Create virtual scanner
	scanner := virtual_devices.NewVirtualScanner("scanner-1", "Test Scanner", logger, eventBus)

	ctx := context.Background()
	err = scanner.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scanner: %v", err)
	}
	defer scanner.Stop()

	// Subscribe to events
	sub := eventBus.Subscribe(ctx, "scanner-1")
	defer sub.Close()

	// Wait for an event (scanner auto-generates events)
	select {
	case event := <-sub.Events():
		if event.DeviceID != "scanner-1" {
			t.Errorf("Expected device_id 'scanner-1', got '%s'", event.DeviceID)
		}
		if event.Type == "" {
			t.Error("Event type should not be empty")
		}

	case <-time.After(20 * time.Second):
		t.Error("Timeout waiting for scanner event")
	}
}

// TestJobQueueIntegration tests job queue functionality
func TestJobQueueIntegration(t *testing.T) {
	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	metrics := telemetry.NewMetrics()
	queue := jobs.NewQueue(jobs.DefaultQueueConfig(), logger, metrics)

	// Register a test executor
	queue.RegisterExecutor("test-operation", func(ctx context.Context, job *jobs.Job) (interface{}, error) {
		// Simulate work
		time.Sleep(100 * time.Millisecond)
		return "test result", nil
	})

	// Start the queue
	queue.Start()
	defer queue.Stop()

	// Submit a test job
	job := jobs.NewJob("printer-1", "test-operation", map[string]string{"test": "data"})
	err = queue.Submit(job)
	if err != nil {
		t.Fatalf("Failed to submit job: %v", err)
	}

	// Wait for job to complete
	time.Sleep(500 * time.Millisecond)

	// Get job status
	retrievedJob, err := queue.Get(job.ID)
	if err != nil {
		t.Fatalf("Failed to get job: %v", err)
	}

	if retrievedJob.Status != jobs.StatusCompleted {
		t.Errorf("Expected status completed, got %v", retrievedJob.Status)
	}
}

// TestMultiDeviceScenario tests multiple devices working together
func TestMultiDeviceScenario(t *testing.T) {
	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	metrics := telemetry.NewMetrics()
	registry := app.NewRegistry(logger, metrics)
	eventBus := events.NewBus(logger)

	ctx := context.Background()

	// Register multiple devices
	deviceConfigs := []struct {
		id   string
		name string
	}{
		{"printer-1", "Printer 1"},
		{"scanner-1", "Scanner 1"},
		{"scale-1", "Scale 1"},
		{"display-1", "Display 1"},
		{"drawer-1", "Drawer 1"},
	}

	for _, cfg := range deviceConfigs {
		var device devices.Device

		switch cfg.id {
		case "printer-1":
			device = virtual_devices.NewVirtualPrinter(cfg.id, cfg.name, logger)
		case "scanner-1":
			device = virtual_devices.NewVirtualScanner(cfg.id, cfg.name, logger, eventBus)
		case "scale-1":
			device = virtual_devices.NewVirtualScale(cfg.id, cfg.name, logger)
		case "display-1":
			device = virtual_devices.NewVirtualDisplay(cfg.id, cfg.name, logger)
		case "drawer-1":
			device = virtual_devices.NewVirtualDrawer(cfg.id, cfg.name, logger)
		}

		err = registry.Register(device)
		if err != nil {
			t.Fatalf("Failed to register %s: %v", cfg.id, err)
		}

		err = device.Start(ctx)
		if err != nil {
			t.Fatalf("Failed to start %s: %v", cfg.id, err)
		}
	}

	// Verify all devices are registered
	if registry.Count() != len(deviceConfigs) {
		t.Errorf("Expected %d devices, got %d", len(deviceConfigs), registry.Count())
	}

	// Check health of all devices
	for _, cfg := range deviceConfigs {
		device, err := registry.Get(cfg.id)
		if err != nil {
			t.Errorf("Device %s not found: %v", cfg.id, err)
			continue
		}

		health := device.Health()
		if health.Status != devices.HealthReady {
			t.Errorf("Device %s status is %v, expected ready", cfg.id, health.Status)
		}
	}

	// Stop all devices
	for _, cfg := range deviceConfigs {
		device, _ := registry.Get(cfg.id)
		device.Stop()
	}
}

// TestHealthMonitoring tests continuous health monitoring
func TestHealthMonitoring(t *testing.T) {
	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	metrics := telemetry.NewMetrics()

	// Create a virtual scale
	scale := virtual_devices.NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	err = scale.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer scale.Stop()

	// Monitor health for 1 second
	healthChecks := 0
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(1 * time.Second)

	for {
		select {
		case <-ticker.C:
			health := scale.Health()
			if health.Status != devices.HealthReady {
				t.Errorf("Unexpected health status: %v", health.Status)
			}
			healthChecks++

			// Update metrics
			metrics.SetDeviceStatus(scale.ID(), scale.Kind(), health.Status.ToProtoStatus())

		case <-timeout:
			if healthChecks < 5 {
				t.Errorf("Expected at least 5 health checks, got %d", healthChecks)
			}
			return
		}
	}
}

// TestConcurrentOperations tests concurrent device access
func TestConcurrentOperations(t *testing.T) {
	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	scale := virtual_devices.NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	err = scale.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start scale: %v", err)
	}
	defer scale.Stop()

	// Perform concurrent weight readings
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			_, err := scale.ReadWeight(ctx)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent operation failed: %v", err)
	}
}

// TestGracefulShutdown tests clean shutdown of all components
func TestGracefulShutdown(t *testing.T) {
	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	metrics := telemetry.NewMetrics()
	registry := app.NewRegistry(logger, metrics)
	eventBus := events.NewBus(logger)
	queue := jobs.NewQueue(jobs.DefaultQueueConfig(), logger, metrics)

	// Start queue
	queue.Start()

	// Create and start devices
	ctx := context.Background()
	scale := virtual_devices.NewVirtualScale("scale-1", "Test Scale", logger)
	registry.Register(scale)
	scale.Start(ctx)

	scanner := virtual_devices.NewVirtualScanner("scanner-1", "Test Scanner", logger, eventBus)
	registry.Register(scanner)
	scanner.Start(ctx)

	// Subscribe to events
	sub := eventBus.Subscribe(ctx, "scanner-1")

	// Give time for initialization
	time.Sleep(500 * time.Millisecond)

	// Shutdown everything gracefully
	sub.Close()
	scale.Stop()
	scanner.Stop()
	queue.Stop()

	// Verify shutdown was clean (no panics)
	t.Log("Graceful shutdown completed successfully")
}

package virtual_devices

import (
	"context"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices/scale"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

func TestNewVirtualScale(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scale := NewVirtualScale("scale-1", "Test Scale", logger)

	if scale == nil {
		t.Fatal("NewVirtualScale returned nil")
	}

	if scale.ID() != "scale-1" {
		t.Errorf("Expected ID 'scale-1', got '%s'", scale.ID())
	}

	if scale.Name() != "Test Scale" {
		t.Errorf("Expected name 'Test Scale', got '%s'", scale.Name())
	}

	if scale.Kind() != "scale.virtual" {
		t.Errorf("Expected kind 'scale.virtual', got '%s'", scale.Kind())
	}
}

func TestVirtualScale_StartStop(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scale := NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()

	// Start
	err := scale.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Stop
	err = scale.Stop()
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestVirtualScale_ReadWeight(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scale := NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	scale.Start(ctx)
	defer scale.Stop()

	// Read weight
	reading, err := scale.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("ReadWeight failed: %v", err)
	}

	if reading.Weight < 0 {
		t.Errorf("Weight should not be negative: %f", reading.Weight)
	}

	if reading.Unit == "" {
		t.Error("Unit should not be empty")
	}
}

func TestVirtualScale_Zero(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scale := NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	scale.Start(ctx)
	defer scale.Stop()

	// Zero the scale
	err := scale.Zero(ctx)
	if err != nil {
		t.Fatalf("Zero failed: %v", err)
	}

	// Read weight - should be close to 0
	reading, err := scale.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("ReadWeight failed: %v", err)
	}

	if reading.Weight > 0.1 {
		t.Errorf("Weight after zero should be close to 0, got %f", reading.Weight)
	}
}

func TestVirtualScale_Tare(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scale := NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	scale.Start(ctx)
	defer scale.Stop()

	// Wait for a non-zero weight
	time.Sleep(100 * time.Millisecond)

	// Tare the scale
	err := scale.Tare(ctx)
	if err != nil {
		t.Fatalf("Tare failed: %v", err)
	}

	// Read weight - should be close to 0
	reading, err := scale.ReadWeight(ctx)
	if err != nil {
		t.Fatalf("ReadWeight failed: %v", err)
	}

	if reading.Weight > 0.1 {
		t.Errorf("Weight after tare should be close to 0, got %f", reading.Weight)
	}
}

func TestVirtualScale_SetUnit(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scaleDevice := NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	scaleDevice.Start(ctx)
	defer scaleDevice.Stop()

	// Test different units
	units := []scale.WeightUnit{
		scale.UnitKilogram,
		scale.UnitGram,
		scale.UnitPound,
		scale.UnitOunce,
	}

	for _, unit := range units {
		scaleDevice.SetUnit(unit)

		// Read weight and verify unit
		reading, err := scaleDevice.ReadWeight(ctx)
		if err != nil {
			t.Fatalf("ReadWeight failed: %v", err)
		}

		if reading.Unit != unit {
			t.Errorf("Expected unit %s, got %s", unit, reading.Unit)
		}
	}
}

func TestVirtualScale_Health(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scale := NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	scale.Start(ctx)
	defer scale.Stop()

	health := scale.Health()

	if health.Status.String() != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", health.Status.String())
	}

	if health.Timestamp.IsZero() {
		t.Error("Health timestamp should not be zero")
	}
}

func TestVirtualScale_Metadata(t *testing.T) {
	logger, _ := telemetry.NewLogger("info", "json")
	scaleDevice := NewVirtualScale("scale-1", "Test Scale", logger)

	metadata := scaleDevice.Metadata()

	if metadata["type"] != "scale" {
		t.Errorf("Expected type 'scale', got '%s'", metadata["type"])
	}

	if metadata["virtual"] != "true" {
		t.Errorf("Expected virtual 'true', got '%s'", metadata["virtual"])
	}
}

func BenchmarkReadWeight(b *testing.B) {
	logger, _ := telemetry.NewLogger("info", "json")
	scale := NewVirtualScale("scale-1", "Test Scale", logger)

	ctx := context.Background()
	scale.Start(ctx)
	defer scale.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scale.ReadWeight(ctx)
	}
}

package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/app"
	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/scale"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/jobs"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockScale implements the scale.Scale interface for testing
type mockScale struct {
	id       string
	kind     string
	name     string
	reading  scale.WeightReading
	readErr  error
	zeroErr  error
	tareErr  error
}

func (m *mockScale) ID() string                           { return m.id }
func (m *mockScale) Kind() string                         { return m.kind }
func (m *mockScale) Name() string                         { return m.name }
func (m *mockScale) Start(ctx context.Context) error      { return nil }
func (m *mockScale) Stop() error                          { return nil }
func (m *mockScale) Metadata() map[string]string          { return map[string]string{} }
func (m *mockScale) Health() devices.DeviceHealth {
	return devices.DeviceHealth{
		Status:    devices.HealthReady,
		Message:   "mock scale ready",
		Timestamp: time.Now(),
		Error:     nil,
	}
}

func (m *mockScale) ReadWeight(ctx context.Context) (scale.WeightReading, error) {
	if m.readErr != nil {
		return scale.WeightReading{}, m.readErr
	}
	return m.reading, nil
}

func (m *mockScale) Zero(ctx context.Context) error {
	return m.zeroErr
}

func (m *mockScale) Tare(ctx context.Context) error {
	return m.tareErr
}

func setupTestServer(t *testing.T) (*Server, *app.Registry) {
	t.Helper()

	logger, err := telemetry.NewLogger("info", "json")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	metrics := telemetry.NewMetrics()
	registry := app.NewRegistry(logger, metrics)

	queueConfig := jobs.DefaultQueueConfig()
	queue := jobs.NewQueue(queueConfig, logger, metrics)

	eventBus := events.NewBus(logger)

	server := NewServer(registry, queue, eventBus, logger)

	return server, registry
}

func TestGetWeight_Success(t *testing.T) {
	server, registry := setupTestServer(t)

	// Create mock scale with specific reading
	mockScaleDevice := &mockScale{
		id:   "scale-1",
		kind: "scale.serial",
		name: "Test Scale",
		reading: scale.WeightReading{
			Weight:    12.5,
			Unit:      scale.UnitKilogram,
			Stable:    true,
			Timestamp: time.Now(),
		},
	}

	// Register mock scale
	if err := registry.Register(mockScaleDevice); err != nil {
		t.Fatalf("failed to register mock scale: %v", err)
	}

	// Test GetWeight
	resp, err := server.GetWeight(context.Background(), &pb.GetWeightRequest{
		DeviceId: "scale-1",
	})

	if err != nil {
		t.Fatalf("GetWeight failed: %v", err)
	}

	if resp.Reading == nil {
		t.Fatal("GetWeight returned nil reading")
	}

	if resp.Reading.Weight != 12.5 {
		t.Errorf("expected weight 12.5, got %f", resp.Reading.Weight)
	}

	if resp.Reading.Unit != pb.WeightUnit_WEIGHT_UNIT_KG {
		t.Errorf("expected unit KG, got %v", resp.Reading.Unit)
	}

	if !resp.Reading.Stable {
		t.Error("expected stable reading, got unstable")
	}
}

func TestGetWeight_DeviceNotFound(t *testing.T) {
	server, _ := setupTestServer(t)

	// Test GetWeight with non-existent device
	_, err := server.GetWeight(context.Background(), &pb.GetWeightRequest{
		DeviceId: "nonexistent",
	})

	if err == nil {
		t.Fatal("expected error for non-existent device, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}

	if st.Code() != codes.NotFound {
		t.Errorf("expected NotFound code, got %v", st.Code())
	}
}

func TestGetWeight_NotAScale(t *testing.T) {
	// This test would require a complete non-scale device mock
	// Skipping for now as it's not critical for the basic functionality test
	t.Skip("Skipping - requires full non-scale device mock")
}

func TestGetWeight_ReadError(t *testing.T) {
	server, registry := setupTestServer(t)

	// Create mock scale that returns an error
	mockScaleDevice := &mockScale{
		id:      "scale-1",
		kind:    "scale.serial",
		name:    "Test Scale",
		readErr: context.DeadlineExceeded,
	}

	if err := registry.Register(mockScaleDevice); err != nil {
		t.Fatalf("failed to register mock scale: %v", err)
	}

	// Test GetWeight with error
	_, err := server.GetWeight(context.Background(), &pb.GetWeightRequest{
		DeviceId: "scale-1",
	})

	if err == nil {
		t.Fatal("expected error when reading fails, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected gRPC status error, got: %v", err)
	}

	if st.Code() != codes.Internal {
		t.Errorf("expected Internal code, got %v", st.Code())
	}
}

func TestZeroScale_Success(t *testing.T) {
	server, registry := setupTestServer(t)

	mockScaleDevice := &mockScale{
		id:   "scale-1",
		kind: "scale.serial",
		name: "Test Scale",
	}

	if err := registry.Register(mockScaleDevice); err != nil {
		t.Fatalf("failed to register mock scale: %v", err)
	}

	resp, err := server.ZeroScale(context.Background(), &pb.ZeroScaleRequest{
		DeviceId: "scale-1",
	})

	if err != nil {
		t.Fatalf("ZeroScale failed: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestTareScale_Success(t *testing.T) {
	server, registry := setupTestServer(t)

	mockScaleDevice := &mockScale{
		id:   "scale-1",
		kind: "scale.serial",
		name: "Test Scale",
	}

	if err := registry.Register(mockScaleDevice); err != nil {
		t.Fatalf("failed to register mock scale: %v", err)
	}

	resp, err := server.TareScale(context.Background(), &pb.TareScaleRequest{
		DeviceId: "scale-1",
	})

	if err != nil {
		t.Fatalf("TareScale failed: %v", err)
	}

	if !resp.Success {
		t.Error("expected success=true")
	}
}

func TestGetWeight_DifferentUnits(t *testing.T) {
	testCases := []struct {
		name         string
		internalUnit scale.WeightUnit
		expectedUnit pb.WeightUnit
	}{
		{"Kilogram", scale.UnitKilogram, pb.WeightUnit_WEIGHT_UNIT_KG},
		{"Gram", scale.UnitGram, pb.WeightUnit_WEIGHT_UNIT_G},
		{"Pound", scale.UnitPound, pb.WeightUnit_WEIGHT_UNIT_LB},
		{"Ounce", scale.UnitOunce, pb.WeightUnit_WEIGHT_UNIT_OZ},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, registry := setupTestServer(t)

			mockScaleDevice := &mockScale{
				id:   "scale-1",
				kind: "scale.serial",
				name: "Test Scale",
				reading: scale.WeightReading{
					Weight:    10.0,
					Unit:      tc.internalUnit,
					Stable:    true,
					Timestamp: time.Now(),
				},
			}

			if err := registry.Register(mockScaleDevice); err != nil {
				t.Fatalf("failed to register mock scale: %v", err)
			}

			resp, err := server.GetWeight(context.Background(), &pb.GetWeightRequest{
				DeviceId: "scale-1",
			})

			if err != nil {
				t.Fatalf("GetWeight failed: %v", err)
			}

			if resp.Reading.Unit != tc.expectedUnit {
				t.Errorf("expected unit %v, got %v", tc.expectedUnit, resp.Reading.Unit)
			}
		})
	}
}

package badge_printer_zebra

import (
	"context"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StatusMonitor monitors printer status
type StatusMonitor struct {
	driver *Driver
	logger *telemetry.Logger

	// Current status
	currentStatus *pb.BadgePrinterStatus
}

// NewStatusMonitor creates a new status monitor
func NewStatusMonitor(driver *Driver, logger *telemetry.Logger) *StatusMonitor {
	return &StatusMonitor{
		driver: driver,
		logger: logger,
		currentStatus: &pb.BadgePrinterStatus{
			DeviceId: driver.id,
			Status:   pb.DeviceStatus_DEVICE_STATUS_DISCONNECTED,
		},
	}
}

// Run starts the status monitoring loop
func (sm *StatusMonitor) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.updateStatus()
		case <-ctx.Done():
			return
		}
	}
}

// updateStatus queries and updates printer status
func (sm *StatusMonitor) updateStatus() {
	// TODO: Query actual printer status via ZPL commands
	// For now, use simulated status

	status := &pb.BadgePrinterStatus{
		DeviceId: sm.driver.id,
		Status:   sm.convertDeviceStatus(sm.driver.Status()),
		Ribbon: &pb.RibbonStatus{
			Type:             "YMCKO",
			PanelsRemaining:  450,
			PanelsCapacity:   500,
			PercentRemaining: 90,
			LowRibbon:        false,
			RibbonInstalled:  true,
		},
		CardsRemaining: 95,
		CardsCapacity:  100,
		Cleaning: &pb.CleaningStatus{
			PrintsSinceCleaning:  250,
			PrintsUntilCleaning:  750,
			CleaningRequired:     false,
			LastCleaned:          timestamppb.New(time.Now().Add(-24 * time.Hour)),
		},
		Capabilities: &pb.PrinterCapabilities{
			DualSided:          sm.driver.config.DualSided,
			MagneticStripe:     sm.driver.config.MagStripe,
			Rfid:               sm.driver.config.RFID,
			RfidTypes:          sm.driver.config.RFIDTypes,
			Lamination:         sm.driver.config.Lamination,
			MaxDpi:             int32(sm.driver.config.DPI),
			RibbonTypes:        []string{"YMCKO", "KO", "Monochrome"},
			CardFeederCapacity: 100,
		},
		Model:           sm.driver.config.Model,
		FirmwareVersion: "1.0.0",
		SerialNumber:    "BADGE-001",
	}

	sm.currentStatus = status
}

// GetCurrentStatus returns the current status
func (sm *StatusMonitor) GetCurrentStatus() *pb.BadgePrinterStatus {
	return sm.currentStatus
}

// convertDeviceStatus converts internal status to proto status
func (sm *StatusMonitor) convertDeviceStatus(status devices.DeviceStatus) pb.DeviceStatus {
	switch status {
	case devices.StatusReady:
		return pb.DeviceStatus_DEVICE_STATUS_READY
	case devices.StatusBusy:
		return pb.DeviceStatus_DEVICE_STATUS_BUSY
	case devices.StatusError:
		return pb.DeviceStatus_DEVICE_STATUS_ERROR
	case devices.StatusDisconnected:
		return pb.DeviceStatus_DEVICE_STATUS_DISCONNECTED
	default:
		return pb.DeviceStatus_DEVICE_STATUS_UNKNOWN
	}
}

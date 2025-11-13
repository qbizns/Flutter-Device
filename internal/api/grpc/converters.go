package grpc

import (
	"time"

	"github.com/google/uuid"
	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
	"github.com/Macber-eg/Flutter-Device/internal/jobs"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// deviceToProto converts internal device to proto
func deviceToProto(dev interface{}) *pb.Device {
	// Type assertion to get basic device info
	// This is a workaround since we can't directly assert to devices.Device without generated code
	type deviceInfo interface {
		ID() string
		Kind() string
		Name() string
		Health() devices.DeviceHealth
		Metadata() map[string]string
	}

	d, ok := dev.(deviceInfo)
	if !ok {
		return &pb.Device{}
	}

	health := d.Health()

	return &pb.Device{
		Id:   d.ID(),
		Name: d.Name(),
		Kind: d.Kind(),
		Health: &pb.DeviceHealth{
			Status:    healthStatusToProto(health.Status),
			Message:   health.Message,
			Timestamp: timestamppb.New(health.Timestamp),
		},
		Metadata: d.Metadata(),
		LastSeen: timestamppb.Now(),
	}
}

// healthStatusToProto converts internal health status to proto
func healthStatusToProto(status devices.HealthStatus) pb.DeviceStatus {
	switch status {
	case devices.HealthUnknown:
		return pb.DeviceStatus_DEVICE_STATUS_UNKNOWN
	case devices.HealthReady:
		return pb.DeviceStatus_DEVICE_STATUS_READY
	case devices.HealthDegraded:
		return pb.DeviceStatus_DEVICE_STATUS_DEGRADED
	case devices.HealthOffline:
		return pb.DeviceStatus_DEVICE_STATUS_OFFLINE
	case devices.HealthError:
		return pb.DeviceStatus_DEVICE_STATUS_ERROR
	default:
		return pb.DeviceStatus_DEVICE_STATUS_UNSPECIFIED
	}
}

// protoToDocument converts proto document to internal
func protoToDocument(pbDoc *pb.PrintDocument) *printer.PrintDocument {
	doc := printer.NewPrintDocument()

	if pbDoc == nil {
		return doc
	}

	for _, pbSection := range pbDoc.Sections {
		section := printer.Section{
			Elements: make([]printer.Element, 0),
		}

		for _, pbElem := range pbSection.Elements {
			switch elem := pbElem.Element.(type) {
			case *pb.Element_Line:
				section.Elements = append(section.Elements, printer.Element{
					Type: printer.ElementTypeLine,
					Line: protoToLine(elem.Line),
				})
			case *pb.Element_Barcode:
				section.Elements = append(section.Elements, printer.Element{
					Type:    printer.ElementTypeBarcode,
					Barcode: protoToBarcode(elem.Barcode),
				})
			case *pb.Element_QrCode:
				section.Elements = append(section.Elements, printer.Element{
					Type:   printer.ElementTypeQRCode,
					QRCode: protoToQRCode(elem.QrCode),
				})
			case *pb.Element_Image:
				section.Elements = append(section.Elements, printer.Element{
					Type:  printer.ElementTypeImage,
					Image: protoToImage(elem.Image),
				})
			}
		}

		doc.AddSection(section)
	}

	if pbDoc.Options != nil {
		doc.Options = printer.PrintOptions{
			Cut:         pbDoc.Options.Cut,
			DrawerPulse: pbDoc.Options.DrawerPulse,
			Copies:      int(pbDoc.Options.Copies),
		}
	}

	return doc
}

// protoToLine converts proto line to internal
func protoToLine(pbLine *pb.Line) *printer.Line {
	line := &printer.Line{
		Runs:      make([]printer.Run, 0),
		Alignment: protoToAlignment(pbLine.Alignment),
	}

	for _, pbRun := range pbLine.Runs {
		line.Runs = append(line.Runs, printer.Run{
			Text:  pbRun.Text,
			Style: protoToTextStyle(pbRun.Style),
		})
	}

	return line
}

// protoToTextStyle converts proto style to internal
func protoToTextStyle(pbStyle *pb.TextStyle) printer.TextStyle {
	if pbStyle == nil {
		return printer.TextStyle{}
	}

	return printer.TextStyle{
		Bold:         pbStyle.Bold,
		Underline:    pbStyle.Underline,
		Inverse:      pbStyle.Inverse,
		DoubleWidth:  pbStyle.DoubleWidth,
		DoubleHeight: pbStyle.DoubleHeight,
		Font:         protoToFont(pbStyle.Font),
	}
}

// protoToAlignment converts proto alignment to internal
func protoToAlignment(pbAlign pb.Alignment) printer.Alignment {
	switch pbAlign {
	case pb.Alignment_ALIGNMENT_LEFT:
		return printer.AlignLeft
	case pb.Alignment_ALIGNMENT_CENTER:
		return printer.AlignCenter
	case pb.Alignment_ALIGNMENT_RIGHT:
		return printer.AlignRight
	default:
		return printer.AlignLeft
	}
}

// protoToFont converts proto font to internal
func protoToFont(pbFont pb.Font) printer.Font {
	switch pbFont {
	case pb.Font_FONT_A:
		return printer.FontA
	case pb.Font_FONT_B:
		return printer.FontB
	default:
		return printer.FontA
	}
}

// protoToBarcode converts proto barcode to internal
func protoToBarcode(pbBarcode *pb.Barcode) *printer.Barcode {
	if pbBarcode == nil {
		return nil
	}

	return &printer.Barcode{
		Data:      pbBarcode.Data,
		Symbology: protoToSymbology(pbBarcode.Symbology),
		Height:    int(pbBarcode.Height),
		Width:     int(pbBarcode.Width),
	}
}

// protoToSymbology converts proto symbology to internal
func protoToSymbology(pbSym pb.BarcodeSymbology) printer.BarcodeSymbology {
	switch pbSym {
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_UPCA:
		return printer.BarcodeUPCA
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_UPCE:
		return printer.BarcodeUPCE
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_EAN13:
		return printer.BarcodeEAN13
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_EAN8:
		return printer.BarcodeEAN8
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_CODE39:
		return printer.BarcodeCode39
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_CODE128:
		return printer.BarcodeCode128
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_ITF:
		return printer.BarcodeITF
	case pb.BarcodeSymbology_BARCODE_SYMBOLOGY_CODABAR:
		return printer.BarcodeCodabar
	default:
		return printer.BarcodeCode128
	}
}

// protoToQRCode converts proto QR code to internal
func protoToQRCode(pbQR *pb.QRCode) *printer.QRCode {
	if pbQR == nil {
		return nil
	}

	return &printer.QRCode{
		Data: pbQR.Data,
		Size: int(pbQR.Size),
	}
}

// protoToImage converts proto image to internal
func protoToImage(pbImage *pb.Image) *printer.Image {
	if pbImage == nil {
		return nil
	}

	return &printer.Image{
		Data:   pbImage.Data,
		Width:  int(pbImage.Width),
		Height: int(pbImage.Height),
	}
}

// jobToProto converts internal job to proto
func jobToProto(job *jobs.Job) *pb.Job {
	status := jobStatusToProto(job.Status)

	pbJob := &pb.Job{
		JobId:     job.ID,
		DeviceId:  job.DeviceID,
		Type:      job.Type,
		Status:    status,
		CreatedAt: timestamppb.New(job.CreatedAt),
	}

	if job.StartedAt != nil {
		pbJob.StartedAt = timestamppb.New(*job.StartedAt)
	}

	if job.CompletedAt != nil {
		pbJob.CompletedAt = timestamppb.New(*job.CompletedAt)
	}

	if job.Error != nil {
		pbJob.Error = job.Error.Error()
	}

	pbJob.IdempotencyKey = job.IdempotencyKey

	return pbJob
}

// jobStatusToProto converts internal job status to proto
func jobStatusToProto(status jobs.Status) pb.JobStatus {
	switch status {
	case jobs.StatusPending:
		return pb.JobStatus_JOB_STATUS_PENDING
	case jobs.StatusInProgress:
		return pb.JobStatus_JOB_STATUS_IN_PROGRESS
	case jobs.StatusCompleted:
		return pb.JobStatus_JOB_STATUS_COMPLETED
	case jobs.StatusFailed:
		return pb.JobStatus_JOB_STATUS_FAILED
	case jobs.StatusCancelled:
		return pb.JobStatus_JOB_STATUS_CANCELLED
	default:
		return pb.JobStatus_JOB_STATUS_UNSPECIFIED
	}
}

// Helper functions

// timestampNow returns current timestamp
func timestampNow() *timestamppb.Timestamp {
	return timestamppb.New(time.Now())
}

// generateID generates a unique ID
func generateID() string {
	return uuid.New().String()
}

// convertToInterfaceSlice converts typed slice to interface slice
func convertToInterfaceSlice(devices []devices.Device) []interface{} {
	result := make([]interface{}, len(devices))
	for i, d := range devices {
		result[i] = d
	}
	return result
}

// weightUnitToProto converts internal weight unit to proto
func weightUnitToProto(unit string) pb.WeightUnit {
	switch unit {
	case "kg":
		return pb.WeightUnit_WEIGHT_UNIT_KG
	case "g":
		return pb.WeightUnit_WEIGHT_UNIT_G
	case "lb":
		return pb.WeightUnit_WEIGHT_UNIT_LB
	case "oz":
		return pb.WeightUnit_WEIGHT_UNIT_OZ
	default:
		return pb.WeightUnit_WEIGHT_UNIT_UNSPECIFIED
	}
}

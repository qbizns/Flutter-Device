package grpc

import (
	"context"
	"fmt"

	"github.com/Macber-eg/Flutter-Device/internal/app"
	"github.com/Macber-eg/Flutter-Device/internal/devices/printer"
	"github.com/Macber-eg/Flutter-Device/internal/devices/scale"
	"github.com/Macber-eg/Flutter-Device/internal/events"
	"github.com/Macber-eg/Flutter-Device/internal/jobs"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the DeviceBridge gRPC service
type Server struct {
	pb.UnimplementedDeviceBridgeServer
	registry *app.Registry
	queue    *jobs.Queue
	events   *events.Bus
	logger   *telemetry.Logger
}

// NewServer creates a new gRPC server
func NewServer(registry *app.Registry, queue *jobs.Queue, eventBus *events.Bus, logger *telemetry.Logger) *Server {
	return &Server{
		registry: registry,
		queue:    queue,
		events:   eventBus,
		logger:   logger.WithComponent("grpc_server"),
	}
}

// Ping checks if service is alive
func (s *Server) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{
		Message: "Device Bridge is running",
		Version: "2.0.0-dev",
	}, nil
}

// ListDevices returns all registered devices
func (s *Server) ListDevices(ctx context.Context, req *pb.ListDevicesRequest) (*pb.ListDevicesResponse, error) {
	s.logger.Info("listing devices")

	var devices []interface{}

	if req.KindFilter != "" {
		devices = convertToInterfaceSlice(s.registry.ListByKind(req.KindFilter))
	} else {
		devices = convertToInterfaceSlice(s.registry.List())
	}

	pbDevices := make([]*pb.Device, 0, len(devices))
	for _, dev := range devices {
		pbDev := deviceToProto(dev)
		pbDevices = append(pbDevices, pbDev)
	}

	return &pb.ListDevicesResponse{
		Devices: pbDevices,
	}, nil
}

// GetDevice returns device details
func (s *Server) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	s.logger.Info("getting device", telemetry.String("device_id", req.DeviceId))

	device, err := s.registry.Get(req.DeviceId)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	return &pb.GetDeviceResponse{
		Device: deviceToProto(device),
	}, nil
}

// Print sends a document to a printer
func (s *Server) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
	s.logger.Info("print request", telemetry.String("device_id", req.DeviceId))

	// Get device
	device, err := s.registry.Get(req.DeviceId)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Check if it's a printer
	printerDev, ok := device.(printer.Printer)
	if !ok {
		return nil, fmt.Errorf("device %s is not a printer", req.DeviceId)
	}

	// Convert proto document to internal
	doc := protoToDocument(req.Document)

	// Create print job
	job := jobs.NewJobWithIdempotency(req.DeviceId, "print", doc, req.IdempotencyKey)

	// Register executor for this job (if not already registered)
	s.queue.RegisterExecutor("print", func(ctx context.Context, job *jobs.Job) (interface{}, error) {
		doc := job.Payload.(*printer.PrintDocument)
		return nil, printerDev.Print(ctx, doc)
	})

	// Submit job
	if err := s.queue.Submit(job); err != nil {
		return nil, fmt.Errorf("failed to submit job: %w", err)
	}

	return &pb.PrintResponse{
		JobId: job.ID,
	}, nil
}

// OpenDrawer opens a cash drawer
func (s *Server) OpenDrawer(ctx context.Context, req *pb.OpenDrawerRequest) (*pb.OpenDrawerResponse, error) {
	s.logger.Info("open drawer request", telemetry.String("device_id", req.DeviceId))

	// Get device
	device, err := s.registry.Get(req.DeviceId)
	if err != nil {
		return nil, fmt.Errorf("device not found: %w", err)
	}

	// Check if it's a printer (drawers are usually connected via printer)
	printerDev, ok := device.(printer.Printer)
	if !ok {
		return nil, fmt.Errorf("device %s does not support drawer", req.DeviceId)
	}

	// Open drawer
	if err := printerDev.OpenDrawer(ctx); err != nil {
		return nil, fmt.Errorf("failed to open drawer: %w", err)
	}

	return &pb.OpenDrawerResponse{
		Success:  true,
		OpenedAt: timestampNow(),
	}, nil
}

// SubscribeScanner streams barcode scan events
func (s *Server) SubscribeScanner(req *pb.SubscribeScannerRequest, stream pb.DeviceBridge_SubscribeScannerServer) error {
	s.logger.Info("scanner subscription", telemetry.String("device_id", req.DeviceId))

	// Subscribe to events
	sub := s.events.Subscribe(stream.Context(), req.DeviceId)
	defer s.events.Unsubscribe(req.DeviceId, sub.ID)

	// Stream events
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event := <-sub.Events():
			// Convert to proto scan event
			scanEvent := &pb.ScanEvent{
				DeviceId:  event.DeviceID,
				Data:      fmt.Sprintf("%v", event.Data),
				Symbology: event.Type,
				Timestamp: timestampNow(),
				EventId:   fmt.Sprintf("evt_%d", event.Data),
			}

			if err := stream.Send(scanEvent); err != nil {
				s.logger.Error("failed to send scan event", telemetry.Error(err))
				return err
			}
		}
	}
}

// GetWeight reads current weight from scale
func (s *Server) GetWeight(ctx context.Context, req *pb.GetWeightRequest) (*pb.GetWeightResponse, error) {
	s.logger.Info("get weight request", telemetry.String("device_id", req.DeviceId))

	// Get device
	device, err := s.registry.Get(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "device not found: %v", err)
	}

	// Type assert to scale.Scale interface
	scaleDevice, ok := device.(scale.Scale)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "device %s is not a scale (kind: %s)", req.DeviceId, device.Kind())
	}

	// Read weight from scale driver
	reading, err := scaleDevice.ReadWeight(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read weight from scale: %v", err)
	}

	// Convert to protobuf response
	return &pb.GetWeightResponse{
		Reading: &pb.WeightReading{
			Weight:    reading.Weight,
			Unit:      weightUnitToProto(string(reading.Unit)),
			Stable:    reading.Stable,
			Timestamp: timestampNow(),
		},
	}, nil
}

// ZeroScale zeros the scale
func (s *Server) ZeroScale(ctx context.Context, req *pb.ZeroScaleRequest) (*pb.ZeroScaleResponse, error) {
	s.logger.Info("zero scale request", telemetry.String("device_id", req.DeviceId))

	// Get device
	device, err := s.registry.Get(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "device not found: %v", err)
	}

	// Type assert to scale.Scale interface
	scaleDevice, ok := device.(scale.Scale)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "device %s is not a scale (kind: %s)", req.DeviceId, device.Kind())
	}

	// Zero the scale
	if err := scaleDevice.Zero(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to zero scale: %v", err)
	}

	return &pb.ZeroScaleResponse{
		Success: true,
	}, nil
}

// TareScale tares the scale
func (s *Server) TareScale(ctx context.Context, req *pb.TareScaleRequest) (*pb.TareScaleResponse, error) {
	s.logger.Info("tare scale request", telemetry.String("device_id", req.DeviceId))

	// Get device
	device, err := s.registry.Get(req.DeviceId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "device not found: %v", err)
	}

	// Type assert to scale.Scale interface
	scaleDevice, ok := device.(scale.Scale)
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "device %s is not a scale (kind: %s)", req.DeviceId, device.Kind())
	}

	// Tare the scale
	if err := scaleDevice.Tare(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to tare scale: %v", err)
	}

	return &pb.TareScaleResponse{
		Success: true,
	}, nil
}

// ShowDisplay shows text on customer display
func (s *Server) ShowDisplay(ctx context.Context, req *pb.ShowDisplayRequest) (*pb.ShowDisplayResponse, error) {
	s.logger.Info("show display request", telemetry.String("device_id", req.DeviceId))

	return &pb.ShowDisplayResponse{
		Success: true,
	}, nil
}

// ClearDisplay clears the display
func (s *Server) ClearDisplay(ctx context.Context, req *pb.ClearDisplayRequest) (*pb.ClearDisplayResponse, error) {
	s.logger.Info("clear display request", telemetry.String("device_id", req.DeviceId))

	return &pb.ClearDisplayResponse{
		Success: true,
	}, nil
}

// StartPayment initiates a payment transaction
func (s *Server) StartPayment(ctx context.Context, req *pb.StartPaymentRequest) (*pb.StartPaymentResponse, error) {
	s.logger.Info("start payment request", telemetry.String("device_id", req.DeviceId))

	return &pb.StartPaymentResponse{
		PaymentId:   "pmt_" + generateID(),
		Status:      pb.PaymentState_PAYMENT_STATE_INITIATED,
		InitiatedAt: timestampNow(),
	}, nil
}

// CancelPayment cancels a payment
func (s *Server) CancelPayment(ctx context.Context, req *pb.CancelPaymentRequest) (*pb.CancelPaymentResponse, error) {
	s.logger.Info("cancel payment request", telemetry.String("device_id", req.DeviceId))

	return &pb.CancelPaymentResponse{
		Success: true,
	}, nil
}

// GetPaymentStatus queries payment status
func (s *Server) GetPaymentStatus(ctx context.Context, req *pb.GetPaymentStatusRequest) (*pb.GetPaymentStatusResponse, error) {
	s.logger.Info("get payment status request", telemetry.String("device_id", req.DeviceId))

	return &pb.GetPaymentStatusResponse{
		Status: &pb.PaymentStatus{
			PaymentId:   req.PaymentId,
			Status:      pb.PaymentState_PAYMENT_STATE_INITIATED,
			Amount:      0.0,
			Currency:    "USD",
			InitiatedAt: timestampNow(),
		},
	}, nil
}

// SubscribePayment streams payment events
func (s *Server) SubscribePayment(req *pb.SubscribePaymentRequest, stream pb.DeviceBridge_SubscribePaymentServer) error {
	s.logger.Info("payment subscription", telemetry.String("device_id", req.DeviceId))

	// Subscribe to events
	sub := s.events.Subscribe(stream.Context(), req.DeviceId)
	defer s.events.Unsubscribe(req.DeviceId, sub.ID)

	// Stream events
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case event := <-sub.Events():
			// Convert to proto payment event
			paymentEvent := &pb.PaymentEvent{
				PaymentId: req.PaymentId,
				Event:     event.Type,
				Status:    pb.PaymentState_PAYMENT_STATE_IN_PROGRESS,
				Message:   fmt.Sprintf("%v", event.Data),
				Timestamp: timestampNow(),
			}

			if err := stream.Send(paymentEvent); err != nil {
				s.logger.Error("failed to send payment event", telemetry.Error(err))
				return err
			}
		}
	}
}

// GetJob returns job status
func (s *Server) GetJob(ctx context.Context, req *pb.GetJobRequest) (*pb.GetJobResponse, error) {
	s.logger.Info("get job request", telemetry.String("job_id", req.JobId))

	job, err := s.queue.Get(req.JobId)
	if err != nil {
		return nil, fmt.Errorf("job not found: %w", err)
	}

	return &pb.GetJobResponse{
		Job: jobToProto(job),
	}, nil
}

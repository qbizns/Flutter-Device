package badge_printer_zebra

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/devices"
	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Driver implements badge printer for Zebra card printers
type Driver struct {
	id     string
	name   string
	config Config
	logger *telemetry.Logger

	// Connection
	conn   Connection
	connMu sync.RWMutex

	// State
	status   devices.DeviceStatus
	statusMu sync.RWMutex

	// Job queue
	jobQueue   chan *PrintJob
	jobResults map[string]*JobResult
	jobsMu     sync.RWMutex

	// Events
	eventChan chan *pb.BadgePrinterEvent

	// Components
	designer  *CardDesigner
	encoder   *Encoder
	statusMon *StatusMonitor

	// Printer info
	info   *PrinterInfo
	infoMu sync.RWMutex

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDriver creates a new Zebra badge printer driver
func NewDriver(id, name string, config Config, logger *telemetry.Logger) (*Driver, error) {
	if logger == nil {
		var err error
		logger, err = telemetry.NewLogger("info", "text")
		if err != nil {
			return nil, fmt.Errorf("failed to create logger: %w", err)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	d := &Driver{
		id:         id,
		name:       name,
		config:     config,
		logger:     logger,
		status:     devices.StatusDisconnected,
		jobQueue:   make(chan *PrintJob, 100),
		jobResults: make(map[string]*JobResult),
		eventChan:  make(chan *pb.BadgePrinterEvent, 50),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Initialize components
	d.designer = NewCardDesigner(config.DPI, logger)
	d.encoder = NewEncoder(logger)
	d.statusMon = NewStatusMonitor(d, logger)

	return d, nil
}

// ID returns the device ID
func (d *Driver) ID() string {
	return d.id
}

// Name returns the device name
func (d *Driver) Name() string {
	return d.name
}

// Kind returns the device kind
func (d *Driver) Kind() string {
	return "badge_printer.zebra"
}

// Status returns the current device status
func (d *Driver) Status() devices.DeviceStatus {
	d.statusMu.RLock()
	defer d.statusMu.RUnlock()
	return d.status
}

// Start initializes the printer connection and workers
func (d *Driver) Start(ctx context.Context) error {
	d.logger.Info("starting zebra badge printer driver",
		telemetry.String("device_id", d.id),
		telemetry.String("model", d.config.Model),
	)

	// Connect to printer
	if err := d.connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	// Query printer info
	if err := d.queryPrinterInfo(); err != nil {
		d.logger.Warn("failed to query printer info",
			telemetry.String("device_id", d.id),
			telemetry.Error(err),
		)
	}

	// Start job processor
	d.wg.Add(1)
	go d.processJobs()

	// Start status monitor
	d.wg.Add(1)
	go d.statusMon.Run(d.ctx)

	d.setStatus(devices.StatusReady)

	d.logger.Info("zebra badge printer started",
		telemetry.String("device_id", d.id),
	)

	return nil
}

// Stop cleanly shuts down the driver
func (d *Driver) Stop(ctx context.Context) error {
	d.logger.Info("stopping zebra badge printer driver",
		telemetry.String("device_id", d.id),
	)

	d.cancel()
	d.wg.Wait()

	if err := d.disconnect(); err != nil {
		d.logger.Error("error disconnecting printer",
			telemetry.String("device_id", d.id),
			telemetry.Error(err),
		)
	}

	close(d.jobQueue)
	close(d.eventChan)

	d.setStatus(devices.StatusDisconnected)

	d.logger.Info("zebra badge printer stopped",
		telemetry.String("device_id", d.id),
	)

	return nil
}

// PrintBadge prints a badge with the given design and encoding
func (d *Driver) PrintBadge(ctx context.Context, req *pb.PrintBadgeRequest) (*pb.PrintBadgeResponse, error) {
	// Validate request
	if err := d.validatePrintRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create job
	job := &PrintJob{
		ID:         generateJobID(),
		Design:     req.Design,
		Encoding:   req.Encoding,
		Options:    req.Options,
		Timestamp:  time.Now(),
		ResultChan: make(chan *JobResult, 1),
	}

	d.logger.Debug("queuing badge print job",
		telemetry.String("job_id", job.ID),
		telemetry.String("device_id", d.id),
	)

	// Queue job
	select {
	case d.jobQueue <- job:
		d.logger.Debug("job queued",
			telemetry.String("job_id", job.ID),
			telemetry.String("device_id", d.id),
		)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-d.ctx.Done():
		return nil, fmt.Errorf("driver stopped")
	}

	// Wait for result
	select {
	case result := <-job.ResultChan:
		if !result.Success {
			return &pb.PrintBadgeResponse{
				JobId:  result.JobID,
				Status: pb.JobStatus_JOB_STATUS_FAILED,
				Message: result.Error.Error(),
				StartedAt: timestamppb.New(result.StartTime),
			}, nil
		}

		return &pb.PrintBadgeResponse{
			JobId:        result.JobID,
			Status:       pb.JobStatus_JOB_STATUS_COMPLETED,
			Message:      "Badge printed successfully",
			CardsPrinted: int32(result.CardsPrinted),
			StartedAt:    timestamppb.New(result.StartTime),
			CompletedAt:  timestamppb.New(result.EndTime),
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	case <-d.ctx.Done():
		return nil, fmt.Errorf("driver stopped")
	}
}

// processJobs processes print jobs from the queue
func (d *Driver) processJobs() {
	defer d.wg.Done()

	for {
		select {
		case job := <-d.jobQueue:
			result := d.executeJob(job)
			job.ResultChan <- result
			close(job.ResultChan)

		case <-d.ctx.Done():
			return
		}
	}
}

// executeJob executes a single print job
func (d *Driver) executeJob(job *PrintJob) *JobResult {
	start := time.Now()

	d.logger.Info("executing print job",
		telemetry.String("job_id", job.ID),
		telemetry.String("device_id", d.id),
	)

	result := &JobResult{
		JobID:     job.ID,
		StartTime: start,
	}

	// Emit job started event
	d.emitEvent(&pb.BadgePrinterEvent{
		DeviceId:  d.id,
		Type:      pb.BadgePrinterEvent_EVENT_TYPE_JOB_STARTED,
		Timestamp: timestamppb.New(job.Timestamp),
		JobId:     job.ID,
	})

	// 1. Render badge design to ZPL
	zpl, err := d.designer.RenderToZPL(job.Design, job.Options)
	if err != nil {
		result.Error = fmt.Errorf("render failed: %w", err)
		result.EndTime = time.Now()
		d.emitJobFailedEvent(job.ID, err)
		return result
	}

	d.logger.Debug("badge design rendered",
		telemetry.String("job_id", job.ID),
		telemetry.Int("zpl_length", len(zpl)),
	)

	// 2. Encode magnetic stripe or RFID (if requested)
	if job.Encoding != nil {
		if err := d.encodeCard(job.Encoding); err != nil {
			result.Error = fmt.Errorf("encoding failed: %w", err)
			result.EndTime = time.Now()
			d.emitJobFailedEvent(job.ID, err)
			return result
		}

		d.emitEvent(&pb.BadgePrinterEvent{
			DeviceId:  d.id,
			Type:      pb.BadgePrinterEvent_EVENT_TYPE_CARD_ENCODED,
			Timestamp: timestamppb.New(time.Now()),
			JobId:     job.ID,
		})
	}

	// 3. Send ZPL to printer
	copies := int(job.Options.GetCopies())
	if copies <= 0 {
		copies = 1
	}

	for i := 0; i < copies; i++ {
		if err := d.sendZPL(zpl); err != nil {
			result.Error = fmt.Errorf("print failed on copy %d: %w", i+1, err)
			result.EndTime = time.Now()
			d.emitJobFailedEvent(job.ID, err)
			return result
		}
		result.CardsPrinted++

		// Emit card printed event
		d.emitEvent(&pb.BadgePrinterEvent{
			DeviceId:  d.id,
			Type:      pb.BadgePrinterEvent_EVENT_TYPE_CARD_PRINTED,
			Timestamp: timestamppb.New(time.Now()),
			JobId:     job.ID,
			Details:   fmt.Sprintf("Card %d of %d", i+1, copies),
		})

		d.logger.Debug("card printed",
			telemetry.String("job_id", job.ID),
			telemetry.Int("copy", i+1),
			telemetry.Int("total", copies),
		)
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	d.logger.Info("job completed",
		telemetry.String("job_id", job.ID),
		telemetry.Int("cards_printed", result.CardsPrinted),
		telemetry.Duration("duration", result.Duration),
	)

	// Emit job completed event
	d.emitEvent(&pb.BadgePrinterEvent{
		DeviceId:  d.id,
		Type:      pb.BadgePrinterEvent_EVENT_TYPE_JOB_COMPLETED,
		Timestamp: timestamppb.New(time.Now()),
		JobId:     job.ID,
		Details:   fmt.Sprintf("%d cards printed in %v", result.CardsPrinted, result.Duration),
	})

	// Store result
	d.jobsMu.Lock()
	d.jobResults[job.ID] = result
	d.jobsMu.Unlock()

	return result
}

// GetStatus returns the current printer status
func (d *Driver) GetPrinterStatus(ctx context.Context, req *pb.GetPrinterStatusRequest) (*pb.BadgePrinterStatus, error) {
	status := d.statusMon.GetCurrentStatus()
	return status, nil
}

// Events returns a channel of printer events
func (d *Driver) Events(ctx context.Context) (<-chan *pb.BadgePrinterEvent, error) {
	return d.eventChan, nil
}

// GetBadgeTemplates returns available badge templates
func (d *Driver) GetBadgeTemplates(ctx context.Context, req *pb.GetBadgeTemplatesRequest) (*pb.GetBadgeTemplatesResponse, error) {
	templates := GetPredefinedTemplates()
	return &pb.GetBadgeTemplatesResponse{
		Templates: templates,
	}, nil
}

// GetBadgeTemplate returns a specific template
func (d *Driver) GetBadgeTemplate(ctx context.Context, req *pb.GetBadgeTemplateRequest) (*pb.BadgeTemplate, error) {
	template, err := GetTemplateByID(req.TemplateId)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}
	return template, nil
}

// DesignBadge previews a badge design
func (d *Driver) DesignBadge(ctx context.Context, req *pb.DesignBadgeRequest) (*pb.BadgeDesign, error) {
	// Validate and return the design
	// In a full implementation, this would generate a preview image
	return req.Design, nil
}

// EncodeCard encodes a card without printing
func (d *Driver) EncodeCard(ctx context.Context, req *pb.EncodeCardRequest) (*pb.EncodeCardResponse, error) {
	if req.Encoding == nil {
		return nil, fmt.Errorf("encoding data is required")
	}

	if err := d.encodeCard(req.Encoding); err != nil {
		return &pb.EncodeCardResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Build card info
	cardInfo := &pb.EncodedCardInfo{}
	if req.Encoding.Rfid != nil && req.Encoding.Rfid.Enabled {
		cardInfo.Uid = req.Encoding.Rfid.Uid
		cardInfo.BlocksWritten = int32(len(req.Encoding.Rfid.Blocks))
	}
	if req.Encoding.MagneticStripe != nil && req.Encoding.MagneticStripe.Enabled {
		tracks := []int32{}
		if req.Encoding.MagneticStripe.Track1Data != "" {
			tracks = append(tracks, 1)
		}
		if req.Encoding.MagneticStripe.Track2Data != "" {
			tracks = append(tracks, 2)
		}
		if req.Encoding.MagneticStripe.Track3Data != "" {
			tracks = append(tracks, 3)
		}
		cardInfo.TracksWritten = tracks
	}

	return &pb.EncodeCardResponse{
		Success:  true,
		Message:  "Card encoded successfully",
		CardInfo: cardInfo,
	}, nil
}

// PrintBadgeBatch prints multiple badges in a batch
func (d *Driver) PrintBadgeBatch(ctx context.Context, req *pb.PrintBadgeBatchRequest) (*pb.PrintBadgeBatchResponse, error) {
	batchID := generateJobID()
	results := make([]*pb.BatchBadgeResult, 0, len(req.Badges))

	successful := 0
	failed := 0

	for _, badge := range req.Badges {
		// Use badge-specific options or default
		options := badge.Options
		if options == nil {
			options = req.DefaultOptions
		}

		// Print badge
		printReq := &pb.PrintBadgeRequest{
			DeviceId: req.DeviceId,
			Design:   badge.Design,
			Encoding: badge.Encoding,
			Options:  options,
		}

		resp, err := d.PrintBadge(ctx, printReq)

		result := &pb.BatchBadgeResult{
			BadgeId: badge.BadgeId,
		}

		if err != nil {
			result.Success = false
			result.Message = err.Error()
			failed++
		} else if resp.Status == pb.JobStatus_JOB_STATUS_FAILED {
			result.Success = false
			result.Message = resp.Message
			failed++
		} else {
			result.Success = true
			result.JobId = resp.JobId
			result.Message = "Success"
			successful++
		}

		results = append(results, result)
	}

	return &pb.PrintBadgeBatchResponse{
		BatchId:     batchID,
		TotalBadges: int32(len(req.Badges)),
		Successful:  int32(successful),
		Failed:      int32(failed),
		Results:     results,
	}, nil
}

// Helper methods

func (d *Driver) connect() error {
	var conn Connection
	var err error

	switch d.config.Transport {
	case "usb":
		conn, err = connectUSB(d.config.VendorID, d.config.ProductID, d.logger)
	case "tcp":
		conn, err = connectTCP(d.config.Address, d.config.Port, d.logger)
	default:
		return fmt.Errorf("unsupported transport: %s", d.config.Transport)
	}

	if err != nil {
		return err
	}

	d.connMu.Lock()
	d.conn = conn
	d.connMu.Unlock()

	d.logger.Info("connected to badge printer",
		telemetry.String("device_id", d.id),
		telemetry.String("transport", d.config.Transport),
	)

	return nil
}

func (d *Driver) disconnect() error {
	d.connMu.Lock()
	defer d.connMu.Unlock()

	if d.conn == nil {
		return nil
	}

	err := d.conn.Close()
	d.conn = nil
	return err
}

func (d *Driver) sendZPL(zpl string) error {
	d.connMu.RLock()
	defer d.connMu.RUnlock()

	if d.conn == nil {
		return fmt.Errorf("not connected")
	}

	_, err := d.conn.Write([]byte(zpl))
	if err != nil {
		return fmt.Errorf("failed to send ZPL: %w", err)
	}

	return nil
}

func (d *Driver) encodeCard(encoding *pb.EncodingData) error {
	if encoding == nil {
		return nil
	}

	d.connMu.RLock()
	conn := d.conn
	d.connMu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	return d.encoder.Encode(conn, encoding)
}

func (d *Driver) validatePrintRequest(req *pb.PrintBadgeRequest) error {
	if req.DeviceId != d.id {
		return fmt.Errorf("device ID mismatch")
	}
	if req.Design == nil {
		return fmt.Errorf("design is required")
	}
	return nil
}

func (d *Driver) queryPrinterInfo() error {
	// In a full implementation, query printer for model, serial, firmware
	// For now, use config values
	d.infoMu.Lock()
	defer d.infoMu.Unlock()

	d.info = &PrinterInfo{
		Model:           d.config.Model,
		SerialNumber:    "Unknown",
		FirmwareVersion: "Unknown",
		Capabilities: &PrinterCapabilities{
			DualSided:          d.config.DualSided,
			MagneticStripe:     d.config.MagStripe,
			RFID:               d.config.RFID,
			RFIDTypes:          d.config.RFIDTypes,
			Lamination:         d.config.Lamination,
			MaxDPI:             d.config.DPI,
			RibbonTypes:        []string{"YMCKO", "KO", "Monochrome"},
			CardFeederCapacity: 100,
		},
	}

	return nil
}

func (d *Driver) setStatus(status devices.DeviceStatus) {
	d.statusMu.Lock()
	d.status = status
	d.statusMu.Unlock()
}

func (d *Driver) emitEvent(event *pb.BadgePrinterEvent) {
	select {
	case d.eventChan <- event:
	default:
		d.logger.Warn("event channel full, dropping event",
			telemetry.String("device_id", d.id),
			telemetry.String("event_type", event.Type.String()),
		)
	}
}

func (d *Driver) emitJobFailedEvent(jobID string, err error) {
	d.emitEvent(&pb.BadgePrinterEvent{
		DeviceId:  d.id,
		Type:      pb.BadgePrinterEvent_EVENT_TYPE_JOB_FAILED,
		Timestamp: timestamppb.New(time.Now()),
		JobId:     jobID,
		Details:   err.Error(),
	})
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job-%d", time.Now().UnixNano())
}

package rfid_reader

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

// Driver implements RFID/NFC reader functionality
type Driver struct {
	id     string
	name   string
	config Config
	logger *telemetry.Logger

	// Connection
	conn   Connection
	connMu sync.RWMutex

	// Protocol handlers
	handlers map[pb.CardType]CardProtocolHandler

	// Current state
	status          devices.DeviceStatus
	mode            pb.ReaderMode
	statusMu        sync.RWMutex

	// Card presence tracking
	presenceTracker *CardPresenceTracker

	// Statistics
	stats *ReaderStatistics

	// Event channels
	cardReadChan   chan *pb.CardReadEvent
	readerEventChan chan *pb.ReaderEvent

	// Request queues
	readQueue  chan *ReadRequest
	writeQueue chan *WriteRequest
	authQueue  chan *AuthRequest

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewDriver creates a new RFID reader driver
func NewDriver(id, name string, config Config, logger *telemetry.Logger) (*Driver, error) {
	if logger == nil {
		logger = telemetry.NewLogger("rfid_reader")
	}

	ctx, cancel := context.WithCancel(context.Background())

	driver := &Driver{
		id:              id,
		name:            name,
		config:          config,
		logger:          logger,
		status:          devices.StatusDisconnected,
		mode:            config.Mode,
		handlers:        make(map[pb.CardType]CardProtocolHandler),
		presenceTracker: &CardPresenceTracker{},
		stats: &ReaderStatistics{
			StartedAt: time.Now(),
		},
		cardReadChan:    make(chan *pb.CardReadEvent, 100),
		readerEventChan: make(chan *pb.ReaderEvent, 100),
		readQueue:       make(chan *ReadRequest, 10),
		writeQueue:      make(chan *WriteRequest, 10),
		authQueue:       make(chan *AuthRequest, 10),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Register protocol handlers
	driver.registerProtocolHandlers()

	return driver, nil
}

// registerProtocolHandlers registers all supported protocol handlers
func (d *Driver) registerProtocolHandlers() {
	// Register handlers based on supported protocols
	for _, protocol := range d.config.SupportedProtocols {
		switch protocol {
		case "prox":
			d.handlers[pb.CardType_CARD_TYPE_HID_PROX] = NewProxHandler(d.logger)
		case "mifare":
			d.handlers[pb.CardType_CARD_TYPE_MIFARE_CLASSIC_1K] = NewMifareHandler(d.logger)
			d.handlers[pb.CardType_CARD_TYPE_MIFARE_CLASSIC_4K] = NewMifareHandler(d.logger)
		case "iclass":
			d.handlers[pb.CardType_CARD_TYPE_HID_ICLASS] = NewIClassHandler(d.logger)
		case "ntag":
			d.handlers[pb.CardType_CARD_TYPE_NTAG_213] = NewNTAGHandler(d.logger)
			d.handlers[pb.CardType_CARD_TYPE_NTAG_215] = NewNTAGHandler(d.logger)
			d.handlers[pb.CardType_CARD_TYPE_NTAG_216] = NewNTAGHandler(d.logger)
		case "desfire":
			d.handlers[pb.CardType_CARD_TYPE_MIFARE_DESFIRE] = NewDESFireHandler(d.logger)
		case "felica":
			d.handlers[pb.CardType_CARD_TYPE_FELICA] = NewFeliCaHandler(d.logger)
		}
	}

	d.logger.Info("protocol handlers registered",
		telemetry.Int("count", len(d.handlers)),
	)
}

// Start starts the RFID reader driver
func (d *Driver) Start() error {
	d.statusMu.Lock()
	if d.status != devices.StatusDisconnected {
		d.statusMu.Unlock()
		return fmt.Errorf("driver already started")
	}
	d.statusMu.Unlock()

	d.logger.Info("starting RFID reader driver",
		telemetry.String("id", d.id),
		telemetry.String("transport", d.config.Transport),
	)

	// Establish connection
	if err := d.connect(); err != nil {
		d.setStatus(devices.StatusError)
		return fmt.Errorf("failed to connect: %w", err)
	}

	d.setStatus(devices.StatusReady)

	// Start worker goroutines
	d.wg.Add(4)
	go d.cardDetectionLoop()
	go d.readWorker()
	go d.writeWorker()
	go d.authWorker()

	d.logger.Info("RFID reader driver started successfully")

	return nil
}

// Stop stops the RFID reader driver
func (d *Driver) Stop() error {
	d.logger.Info("stopping RFID reader driver")

	// Cancel context
	d.cancel()

	// Close connection
	d.connMu.Lock()
	if d.conn != nil {
		d.conn.Close()
		d.conn = nil
	}
	d.connMu.Unlock()

	// Wait for workers
	d.wg.Wait()

	d.setStatus(devices.StatusDisconnected)

	d.logger.Info("RFID reader driver stopped")

	return nil
}

// Status returns the current device status
func (d *Driver) Status() devices.DeviceStatus {
	d.statusMu.RLock()
	defer d.statusMu.RUnlock()
	return d.status
}

// setStatus sets the device status
func (d *Driver) setStatus(status devices.DeviceStatus) {
	d.statusMu.Lock()
	defer d.statusMu.Unlock()
	d.status = status
}

// connect establishes connection to the reader
func (d *Driver) connect() error {
	var conn Connection
	var err error

	switch d.config.Transport {
	case "usb":
		conn, err = connectUSB(d.config.VendorID, d.config.ProductID, d.logger)
	case "tcp":
		conn, err = connectTCP(d.config.Address, d.config.Port, d.logger)
	case "serial":
		conn, err = connectSerial(d.config.SerialDev, d.config.BaudRate, d.logger)
	default:
		return fmt.Errorf("unsupported transport: %s", d.config.Transport)
	}

	if err != nil {
		return err
	}

	d.connMu.Lock()
	d.conn = conn
	d.connMu.Unlock()

	return nil
}

// ==================================================
// RPC Method Implementations
// ==================================================

// ReadCard reads a card when presented to the reader
func (d *Driver) ReadCard(ctx context.Context, req *pb.ReadCardRequest) (*pb.ReadCardResponse, error) {
	d.logger.Debug("read card request",
		telemetry.String("device_id", req.DeviceId),
		telemetry.Int("timeout", int(req.TimeoutSeconds)),
	)

	timeout := time.Duration(req.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = time.Duration(d.config.ReadTimeoutSec) * time.Second
	}

	readReq := &ReadRequest{
		ID:            generateRequestID(),
		Timeout:       timeout,
		ExpectedTypes: req.ExpectedTypes,
		Options:       req.Options,
		ResultChan:    make(chan *ReadResult, 1),
		Timestamp:     time.Now(),
	}

	// Send to read queue
	select {
	case d.readQueue <- readReq:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting to queue read request")
	}

	// Wait for result
	select {
	case result := <-readReq.ResultChan:
		if result.Error != nil {
			return nil, result.Error
		}

		return &pb.ReadCardResponse{
			Card:           result.Card,
			DeviceId:       d.id,
			Timestamp:      timestamppb.New(result.Timestamp),
			SignalStrength: result.SignalStrength,
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for card read")
	}
}

// WriteCard writes data to a card
func (d *Driver) WriteCard(ctx context.Context, req *pb.WriteCardRequest) (*pb.WriteCardResponse, error) {
	if !d.config.CanWrite {
		return nil, fmt.Errorf("reader does not support writing")
	}

	d.logger.Debug("write card request",
		telemetry.String("device_id", req.DeviceId),
		telemetry.Int("blocks", len(req.Blocks)),
	)

	timeout := time.Duration(d.config.WriteTimeoutSec) * time.Second
	if req.Options != nil && req.Options.TimeoutSeconds > 0 {
		timeout = time.Duration(req.Options.TimeoutSeconds) * time.Second
	}

	writeReq := &WriteRequest{
		ID:          generateRequestID(),
		CardUID:     req.CardUid,
		Blocks:      req.Blocks,
		NDEFRecords: req.NdefRecords,
		Credentials: req.Credentials,
		Options:     req.Options,
		ResultChan:  make(chan *WriteResult, 1),
		Timestamp:   time.Now(),
	}

	// Send to write queue
	select {
	case d.writeQueue <- writeReq:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting to queue write request")
	}

	// Wait for result
	select {
	case result := <-writeReq.ResultChan:
		if result.Error != nil {
			return nil, result.Error
		}

		return &pb.WriteCardResponse{
			Success:       result.Success,
			BlocksWritten: result.BlocksWritten,
			Timestamp:     timestamppb.New(result.Timestamp),
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout waiting for card write")
	}
}

// AuthenticateCard authenticates a card
func (d *Driver) AuthenticateCard(ctx context.Context, req *pb.AuthenticateCardRequest) (*pb.AuthenticateCardResponse, error) {
	if !d.config.CanAuthenticate {
		return nil, fmt.Errorf("reader does not support authentication")
	}

	d.logger.Debug("authenticate card request",
		telemetry.String("device_id", req.DeviceId),
		telemetry.String("method", req.Method.String()),
	)

	authReq := &AuthRequest{
		ID:          generateRequestID(),
		CardUID:     req.CardUid,
		Method:      req.Method,
		Credentials: req.Credentials,
		Challenge:   req.Challenge,
		ResultChan:  make(chan *AuthResult, 1),
		Timestamp:   time.Now(),
	}

	// Send to auth queue
	select {
	case d.authQueue <- authReq:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Wait for result
	select {
	case result := <-authReq.ResultChan:
		if result.Error != nil {
			return nil, result.Error
		}

		return &pb.AuthenticateCardResponse{
			Authenticated:  result.Authenticated,
			Reason:         result.Reason,
			CredentialInfo: result.CredentialInfo,
			Response:       result.Response,
			Timestamp:      timestamppb.New(result.Timestamp),
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetReaderStatus returns current reader status
func (d *Driver) GetReaderStatus(ctx context.Context, req *pb.GetReaderStatusRequest) (*pb.ReaderStatus, error) {
	d.statusMu.RLock()
	status := d.status
	mode := d.mode
	d.statusMu.RUnlock()

	cardUID, cardType := d.presenceTracker.GetCurrentCard()

	supportedTypes := make([]pb.CardType, 0, len(d.handlers))
	for cardType := range d.handlers {
		supportedTypes = append(supportedTypes, cardType)
	}

	return &pb.ReaderStatus{
		DeviceId:       d.id,
		Status:         convertDeviceStatus(status),
		Mode:           mode,
		SupportedTypes: supportedTypes,
		SupportedProtocols: d.config.SupportedProtocols,
		Capabilities: &pb.ReaderCapabilities{
			CanRead:         d.config.CanRead,
			CanWrite:        d.config.CanWrite,
			CanAuthenticate: d.config.CanAuthenticate,
			HasBuzzer:       d.config.HasBuzzer,
			HasLed:          d.config.HasLED,
			SupportsNdef:    d.config.SupportsNDEF,
		},
		CardPresent:     d.presenceTracker.IsCardPresent(),
		CurrentCardUid:  cardUID,
		Model:           d.config.Model,
		Manufacturer:    d.config.Manufacturer,
		ConnectionType:  d.config.Transport,
		Statistics:      d.stats.ToProto(),
	}, nil
}

// SetReaderMode sets the reader operation mode
func (d *Driver) SetReaderMode(ctx context.Context, req *pb.SetReaderModeRequest) (*pb.SetReaderModeResponse, error) {
	d.statusMu.Lock()
	oldMode := d.mode
	d.mode = req.Mode
	d.statusMu.Unlock()

	d.logger.Info("reader mode changed",
		telemetry.String("old_mode", oldMode.String()),
		telemetry.String("new_mode", req.Mode.String()),
	)

	// Emit mode changed event
	d.emitReaderEvent(&pb.ReaderEvent{
		EventId:  generateEventID(),
		DeviceId: d.id,
		Type:     pb.ReaderEventType_READER_EVENT_TYPE_MODE_CHANGED,
		Timestamp: timestamppb.Now(),
		Data: &pb.ReaderEvent_ModeChanged{
			ModeChanged: &pb.ReaderModeChangedEvent{
				OldMode: oldMode,
				NewMode: req.Mode,
			},
		},
	})

	return &pb.SetReaderModeResponse{
		Success:     true,
		CurrentMode: req.Mode,
	}, nil
}

// GetSupportedCardTypes returns supported card types
func (d *Driver) GetSupportedCardTypes(ctx context.Context, req *pb.GetSupportedCardTypesRequest) (*pb.GetSupportedCardTypesResponse, error) {
	cardTypes := make([]*pb.CardTypeInfo, 0, len(d.handlers))

	for cardType, handler := range d.handlers {
		cardTypes = append(cardTypes, &pb.CardTypeInfo{
			Type:     cardType,
			Name:     cardType.String(),
			// Technology will be determined by card type
		})
	}

	return &pb.GetSupportedCardTypesResponse{
		CardTypes: cardTypes,
		Protocols: d.config.SupportedProtocols,
	}, nil
}

// ==================================================
// Worker Goroutines
// ==================================================

// cardDetectionLoop continuously detects cards
func (d *Driver) cardDetectionLoop() {
	defer d.wg.Done()

	d.logger.Debug("card detection loop started")

	ticker := time.NewTicker(time.Duration(d.config.ReadIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.detectCard()
		}
	}
}

// detectCard attempts to detect a card
func (d *Driver) detectCard() {
	// This is a placeholder - actual implementation depends on reader hardware
	// Real readers will have specific commands to poll for cards

	// TODO: Implement reader-specific card detection
	// For now, this is a stub that would be implemented based on the specific reader
}

// readWorker processes read requests
func (d *Driver) readWorker() {
	defer d.wg.Done()

	d.logger.Debug("read worker started")

	for {
		select {
		case <-d.ctx.Done():
			return
		case req := <-d.readQueue:
			d.processReadRequest(req)
		}
	}
}

// processReadRequest processes a single read request
func (d *Driver) processReadRequest(req *ReadRequest) {
	startTime := time.Now()

	// TODO: Implement actual card reading logic
	// This would involve:
	// 1. Detecting card presence
	// 2. Reading card UID
	// 3. Identifying card type
	// 4. Using appropriate protocol handler
	// 5. Reading card data based on options

	// Placeholder implementation
	result := &ReadResult{
		Error:     fmt.Errorf("card reading not yet implemented"),
		Timestamp: time.Now(),
	}

	d.stats.IncrementReads(false, time.Since(startTime))

	select {
	case req.ResultChan <- result:
	case <-d.ctx.Done():
	}
}

// writeWorker processes write requests
func (d *Driver) writeWorker() {
	defer d.wg.Done()

	d.logger.Debug("write worker started")

	for {
		select {
		case <-d.ctx.Done():
			return
		case req := <-d.writeQueue:
			d.processWriteRequest(req)
		}
	}
}

// processWriteRequest processes a single write request
func (d *Driver) processWriteRequest(req *WriteRequest) {
	// TODO: Implement actual card writing logic

	d.stats.IncrementWrites()

	result := &WriteResult{
		Success:   false,
		Error:     fmt.Errorf("card writing not yet implemented"),
		Timestamp: time.Now(),
	}

	select {
	case req.ResultChan <- result:
	case <-d.ctx.Done():
	}
}

// authWorker processes authentication requests
func (d *Driver) authWorker() {
	defer d.wg.Done()

	d.logger.Debug("auth worker started")

	for {
		select {
		case <-d.ctx.Done():
			return
		case req := <-d.authQueue:
			d.processAuthRequest(req)
		}
	}
}

// processAuthRequest processes a single authentication request
func (d *Driver) processAuthRequest(req *AuthRequest) {
	// TODO: Implement actual authentication logic

	authenticated := false
	d.stats.IncrementAuths(authenticated)

	result := &AuthResult{
		Authenticated: false,
		Reason:        "authentication not yet implemented",
		Timestamp:     time.Now(),
	}

	select {
	case req.ResultChan <- result:
	case <-d.ctx.Done():
	}
}

// ==================================================
// Event Emission
// ==================================================

// emitCardReadEvent emits a card read event
func (d *Driver) emitCardReadEvent(event *pb.CardReadEvent) {
	select {
	case d.cardReadChan <- event:
	default:
		d.logger.Warn("card read event channel full, dropping event")
	}
}

// emitReaderEvent emits a reader event
func (d *Driver) emitReaderEvent(event *pb.ReaderEvent) {
	select {
	case d.readerEventChan <- event:
	default:
		d.logger.Warn("reader event channel full, dropping event")
	}
}

// ==================================================
// Helper Functions
// ==================================================

func convertDeviceStatus(status devices.DeviceStatus) pb.DeviceStatus {
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

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

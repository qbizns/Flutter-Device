package payment_tcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// AuditEvent represents an audit log event for a payment transaction
type AuditEvent struct {
	Timestamp      time.Time         `json:"timestamp"`
	EventID        string            `json:"event_id"`
	EventType      string            `json:"event_type"` // transaction, connection, settlement
	Action         string            `json:"action"`     // start, complete, fail, cancel
	DeviceID       string            `json:"device_id"`
	TerminalID     string            `json:"terminal_id"`
	MerchantID     string            `json:"merchant_id"`

	// Transaction details (masked for PCI compliance)
	TransactionType string            `json:"transaction_type,omitempty"`
	TransactionID   string            `json:"transaction_id,omitempty"`
	Amount          int64             `json:"amount,omitempty"`
	Currency        string            `json:"currency,omitempty"`
	Reference       string            `json:"reference,omitempty"`

	// Response details
	ResponseCode    string            `json:"response_code,omitempty"`
	ResponseMessage string            `json:"response_message,omitempty"`
	Success         bool              `json:"success"`

	// Card info (masked)
	CardNumberMasked string           `json:"card_number_masked,omitempty"` // ****1234
	CardType         string            `json:"card_type,omitempty"`

	// Security
	UserID          string            `json:"user_id,omitempty"`
	IPAddress       string            `json:"ip_address,omitempty"`

	// Additional context
	ErrorMessage    string            `json:"error_message,omitempty"`
	Duration        time.Duration     `json:"duration_ms,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// AuditLogger handles audit logging for payment transactions
type AuditLogger struct {
	enabled  bool
	writer   AuditWriter
	mu       sync.Mutex
	buffer   []AuditEvent
	bufferSize int
}

// AuditWriter is an interface for writing audit events
type AuditWriter interface {
	Write(event AuditEvent) error
	Flush() error
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(writer AuditWriter, bufferSize int) *AuditLogger {
	if bufferSize <= 0 {
		bufferSize = 100 // Default buffer size
	}

	return &AuditLogger{
		enabled:    true,
		writer:     writer,
		buffer:     make([]AuditEvent, 0, bufferSize),
		bufferSize: bufferSize,
	}
}

// LogTransaction logs a transaction audit event
func (a *AuditLogger) LogTransaction(
	deviceID, terminalID, merchantID string,
	req TransactionRequest,
	resp *TransactionResponse,
	err error,
	duration time.Duration,
) {
	if !a.enabled || a.writer == nil {
		return
	}

	event := AuditEvent{
		Timestamp:       time.Now(),
		EventID:         generateEventID(),
		EventType:       "transaction",
		DeviceID:        deviceID,
		TerminalID:      terminalID,
		MerchantID:      merchantID,
		TransactionType: string(req.Type),
		Amount:          req.Amount,
		Currency:        req.Currency,
		Reference:       req.Reference,
		Duration:        duration,
	}

	if resp != nil {
		event.Action = "complete"
		event.TransactionID = resp.TransactionID
		event.ResponseCode = resp.ResponseCode
		event.ResponseMessage = resp.ResponseMessage
		event.Success = resp.Success
		event.CardNumberMasked = resp.CardNumberMasked
		event.CardType = string(resp.CardType)
	} else if err != nil {
		event.Action = "fail"
		event.Success = false
		event.ErrorMessage = maskSensitiveInfo(err.Error())
	} else {
		event.Action = "start"
	}

	a.log(event)
}

// LogSettlement logs a settlement audit event
func (a *AuditLogger) LogSettlement(
	deviceID, terminalID, merchantID string,
	req SettlementRequest,
	resp *SettlementResponse,
	err error,
	duration time.Duration,
) {
	if !a.enabled || a.writer == nil {
		return
	}

	event := AuditEvent{
		Timestamp:   time.Now(),
		EventID:     generateEventID(),
		EventType:   "settlement",
		DeviceID:    deviceID,
		TerminalID:  terminalID,
		MerchantID:  merchantID,
		Duration:    duration,
	}

	if resp != nil {
		event.Action = "complete"
		event.Success = resp.Success
		event.ResponseCode = resp.ResponseCode
		event.ResponseMessage = resp.ResponseMessage
		event.Metadata = map[string]string{
			"batch_number":    resp.BatchNumber,
			"total_count":     fmt.Sprintf("%d", resp.TotalCount),
			"approved_count":  fmt.Sprintf("%d", resp.ApprovedCount),
			"total_amount":    fmt.Sprintf("%d", resp.TotalAmount),
		}
	} else if err != nil {
		event.Action = "fail"
		event.Success = false
		event.ErrorMessage = maskSensitiveInfo(err.Error())
	} else {
		event.Action = "start"
	}

	a.log(event)
}

// LogConnection logs a connection event
func (a *AuditLogger) LogConnection(
	deviceID, terminalID, merchantID string,
	action string, // connect, disconnect, reconnect
	success bool,
	err error,
) {
	if !a.enabled || a.writer == nil {
		return
	}

	event := AuditEvent{
		Timestamp:  time.Now(),
		EventID:    generateEventID(),
		EventType:  "connection",
		Action:     action,
		DeviceID:   deviceID,
		TerminalID: terminalID,
		MerchantID: merchantID,
		Success:    success,
	}

	if err != nil {
		event.ErrorMessage = maskSensitiveInfo(err.Error())
	}

	a.log(event)
}

// log writes an audit event
func (a *AuditLogger) log(event AuditEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Add to buffer
	a.buffer = append(a.buffer, event)

	// Flush if buffer is full
	if len(a.buffer) >= a.bufferSize {
		a.flushLocked()
	}
}

// Flush writes all buffered events
func (a *AuditLogger) Flush() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.flushLocked()
}

// flushLocked flushes the buffer (must be called with lock held)
func (a *AuditLogger) flushLocked() error {
	if len(a.buffer) == 0 {
		return nil
	}

	// Write all buffered events
	for _, event := range a.buffer {
		if err := a.writer.Write(event); err != nil {
			// Continue writing other events even if one fails
			// In production, you might want to handle this differently
		}
	}

	// Clear buffer
	a.buffer = a.buffer[:0]

	// Flush writer
	if err := a.writer.Flush(); err != nil {
		return fmt.Errorf("error flushing writer: %w", err)
	}

	return nil
}

// Enable enables audit logging
func (a *AuditLogger) Enable() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = true
}

// Disable disables audit logging
func (a *AuditLogger) Disable() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.enabled = false
}

// Close flushes and closes the audit logger
func (a *AuditLogger) Close() error {
	return a.Flush()
}

// FileAuditWriter writes audit events to a file
type FileAuditWriter struct {
	// In a real implementation, this would have file handles and buffering
	filePath string
	mu       sync.Mutex
}

// NewFileAuditWriter creates a new file audit writer
func NewFileAuditWriter(filePath string) (*FileAuditWriter, error) {
	// In a real implementation, open the file and set up buffering
	return &FileAuditWriter{
		filePath: filePath,
	}, nil
}

// Write writes an audit event to the file
func (w *FileAuditWriter) Write(event AuditEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Marshal to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling audit event: %w", err)
	}

	// In a real implementation, write to file
	// For now, we just validate the JSON is valid
	_ = jsonData

	return nil
}

// Flush flushes any buffered data
func (w *FileAuditWriter) Flush() error {
	// In a real implementation, flush file buffers
	return nil
}

// Close closes the file
func (w *FileAuditWriter) Close() error {
	// In a real implementation, close the file
	return nil
}

// Helper functions

// generateEventID generates a unique event ID
func generateEventID() string {
	// Simple implementation using timestamp and random suffix
	// In production, use UUID or similar
	return fmt.Sprintf("%d-%06d", time.Now().Unix(), time.Now().Nanosecond()%1000000)
}

// maskSensitiveInfo masks sensitive information in error messages
func maskSensitiveInfo(msg string) string {
	// In a real implementation, use regex to find and mask:
	// - Card numbers
	// - PIN data
	// - CVV codes
	// - Personal information

	// For now, return as-is since our errors shouldn't contain sensitive data
	return msg
}

// AuditQuery represents a query for audit events
type AuditQuery struct {
	StartTime      time.Time
	EndTime        time.Time
	DeviceID       string
	TerminalID     string
	EventType      string
	TransactionType string
	Success        *bool // nil = all, true = success only, false = failures only
	Limit          int
}

// AuditReader interface for reading audit events
type AuditReader interface {
	Query(query AuditQuery) ([]AuditEvent, error)
}

// MemoryAuditReader stores audit events in memory (for testing)
type MemoryAuditReader struct {
	events []AuditEvent
	mu     sync.RWMutex
}

// NewMemoryAuditReader creates a new in-memory audit reader
func NewMemoryAuditReader() *MemoryAuditReader {
	return &MemoryAuditReader{
		events: make([]AuditEvent, 0),
	}
}

// Write implements AuditWriter
func (r *MemoryAuditReader) Write(event AuditEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, event)
	return nil
}

// Flush implements AuditWriter
func (r *MemoryAuditReader) Flush() error {
	return nil
}

// Query implements AuditReader
func (r *MemoryAuditReader) Query(query AuditQuery) ([]AuditEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []AuditEvent

	for _, event := range r.events {
		// Apply filters
		if !query.StartTime.IsZero() && event.Timestamp.Before(query.StartTime) {
			continue
		}
		if !query.EndTime.IsZero() && event.Timestamp.After(query.EndTime) {
			continue
		}
		if query.DeviceID != "" && event.DeviceID != query.DeviceID {
			continue
		}
		if query.TerminalID != "" && event.TerminalID != query.TerminalID {
			continue
		}
		if query.EventType != "" && event.EventType != query.EventType {
			continue
		}
		if query.TransactionType != "" && event.TransactionType != query.TransactionType {
			continue
		}
		if query.Success != nil && event.Success != *query.Success {
			continue
		}

		results = append(results, event)

		// Apply limit
		if query.Limit > 0 && len(results) >= query.Limit {
			break
		}
	}

	return results, nil
}

// GetAllEvents returns all events (for testing)
func (r *MemoryAuditReader) GetAllEvents() []AuditEvent {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	events := make([]AuditEvent, len(r.events))
	copy(events, r.events)
	return events
}

// Clear clears all events (for testing)
func (r *MemoryAuditReader) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = r.events[:0]
}

package telemetry

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger with convenience methods
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new structured logger
func NewLogger(level string, format string) (*Logger, error) {
	// Parse level
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Create config
	var cfg zap.Config
	if format == "json" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	// Build logger
	logger, err := cfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{Logger: logger}, nil
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{Logger: l.With(fields...)}
}

// WithDeviceID returns a logger with device_id field
func (l *Logger) WithDeviceID(deviceID string) *Logger {
	return l.WithFields(zap.String("device_id", deviceID))
}

// WithJobID returns a logger with job_id field
func (l *Logger) WithJobID(jobID string) *Logger {
	return l.WithFields(zap.String("job_id", jobID))
}

// WithRequestID returns a logger with request_id field
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithFields(zap.String("request_id", requestID))
}

// WithPaymentID returns a logger with payment_id field
func (l *Logger) WithPaymentID(paymentID string) *Logger {
	return l.WithFields(zap.String("payment_id", paymentID))
}

// WithComponent returns a logger with component field
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithFields(zap.String("component", component))
}

// LogDeviceOperation logs a device operation
func (l *Logger) LogDeviceOperation(deviceID, operation string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("device_id", deviceID),
		zap.String("operation", operation),
	}, fields...)
	l.Info("device operation", allFields...)
}

// LogJobStart logs job start
func (l *Logger) LogJobStart(jobID, deviceID, jobType string) {
	l.Info("job started",
		zap.String("job_id", jobID),
		zap.String("device_id", deviceID),
		zap.String("job_type", jobType),
	)
}

// LogJobComplete logs job completion
func (l *Logger) LogJobComplete(jobID, deviceID, jobType string, durationMs int64) {
	l.Info("job completed",
		zap.String("job_id", jobID),
		zap.String("device_id", deviceID),
		zap.String("job_type", jobType),
		zap.Int64("duration_ms", durationMs),
	)
}

// LogJobError logs job error
func (l *Logger) LogJobError(jobID, deviceID, jobType string, err error) {
	l.Error("job failed",
		zap.String("job_id", jobID),
		zap.String("device_id", deviceID),
		zap.String("job_type", jobType),
		zap.Error(err),
	)
}

// LogPaymentEvent logs payment event
func (l *Logger) LogPaymentEvent(paymentID, deviceID, event string, fields ...zap.Field) {
	allFields := append([]zap.Field{
		zap.String("payment_id", paymentID),
		zap.String("device_id", deviceID),
		zap.String("event", event),
	}, fields...)
	l.Info("payment event", allFields...)
}

// LogScanEvent logs scan event (without sensitive data)
func (l *Logger) LogScanEvent(deviceID, symbology string, dataLength int) {
	l.Info("scan event",
		zap.String("device_id", deviceID),
		zap.String("symbology", symbology),
		zap.Int("data_length", dataLength),
	)
}

// LogDeviceConnection logs device connection attempt
func (l *Logger) LogDeviceConnection(deviceID, transport, address string, success bool, err error) {
	fields := []zap.Field{
		zap.String("device_id", deviceID),
		zap.String("transport", transport),
		zap.String("address", address),
		zap.Bool("success", success),
	}
	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("device connection failed", fields...)
	} else {
		l.Info("device connected", fields...)
	}
}

// LogDeviceDisconnection logs device disconnection
func (l *Logger) LogDeviceDisconnection(deviceID string, reason string) {
	l.Info("device disconnected",
		zap.String("device_id", deviceID),
		zap.String("reason", reason),
	)
}

// Sync flushes buffered logs
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

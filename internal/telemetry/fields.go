package telemetry

import (
	"time"

	"go.uber.org/zap"
)

// Helper functions for common zap fields

// String creates a string field
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

// Int creates an int field
func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

// Int64 creates an int64 field
func Int64(key string, value int64) zap.Field {
	return zap.Int64(key, value)
}

// Float64 creates a float64 field
func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

// Bool creates a bool field
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

// Duration creates a duration field
func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

// Time creates a time field
func Time(key string, value time.Time) zap.Field {
	return zap.Time(key, value)
}

// Error creates an error field
func Error(err error) zap.Field {
	return zap.Error(err)
}

// Any creates a field with any value
func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

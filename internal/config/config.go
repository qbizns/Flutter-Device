package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Devices   []DeviceConfig  `mapstructure:"devices"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Telemetry TelemetryConfig `mapstructure:"telemetry"`
	Security  SecurityConfig  `mapstructure:"security"`
}

// ServerConfig contains server settings
type ServerConfig struct {
	GRPCPort    int    `mapstructure:"grpc_port"`
	HTTPPort    int    `mapstructure:"http_port"`
	MetricsPort int    `mapstructure:"metrics_port"`
	Host        string `mapstructure:"host"`
}

// DeviceConfig represents a single device configuration
type DeviceConfig struct {
	ID          string            `mapstructure:"id"`
	Name        string            `mapstructure:"name"`
	Kind        string            `mapstructure:"kind"`
	Transport   string            `mapstructure:"transport"`
	Address     string            `mapstructure:"address"`
	Port        int               `mapstructure:"port"`
	BaudRate    int               `mapstructure:"baud_rate"`
	VendorID    int               `mapstructure:"vendor_id"`
	ProductID   int               `mapstructure:"product_id"`
	Protocol    string            `mapstructure:"protocol"`
	Enabled     bool              `mapstructure:"enabled"`
	Tags        []string          `mapstructure:"tags"`
	Metadata    map[string]string `mapstructure:"metadata"`
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json or text
}

// TelemetryConfig contains telemetry settings
type TelemetryConfig struct {
	Metrics MetricsConfig `mapstructure:"metrics"`
	Tracing TracingConfig `mapstructure:"tracing"`
}

// MetricsConfig contains metrics settings
type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// TracingConfig contains tracing settings
type TracingConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Exporter string `mapstructure:"exporter"`
	Endpoint string `mapstructure:"endpoint"`
}

// SecurityConfig contains security settings
type SecurityConfig struct {
	Mode string     `mapstructure:"mode"` // development or production
	TLS  TLSConfig  `mapstructure:"tls"`
	ACL  ACLConfig  `mapstructure:"acl"`
}

// TLSConfig contains TLS settings
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
	CAFile   string `mapstructure:"ca_file"`
}

// ACLConfig contains ACL settings
type ACLConfig struct {
	Enabled bool       `mapstructure:"enabled"`
	Rules   []ACLRule  `mapstructure:"rules"`
}

// ACLRule defines an access control rule
type ACLRule struct {
	ClientID          string   `mapstructure:"client_id"`
	AllowedDevices    []string `mapstructure:"allowed_devices"`
	AllowedOperations []string `mapstructure:"allowed_operations"`
}

// Load loads configuration from file
func Load(configFile string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Configure viper
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.grpc_port", 50051)
	v.SetDefault("server.http_port", 8080)
	v.SetDefault("server.metrics_port", 9090)
	v.SetDefault("server.host", "localhost")

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Telemetry defaults
	v.SetDefault("telemetry.metrics.enabled", true)
	v.SetDefault("telemetry.tracing.enabled", false)

	// Security defaults
	v.SetDefault("security.mode", "development")
	v.SetDefault("security.tls.enabled", false)
	v.SetDefault("security.acl.enabled", false)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.GRPCPort < 1 || c.Server.GRPCPort > 65535 {
		return fmt.Errorf("invalid grpc_port: %d", c.Server.GRPCPort)
	}
	if c.Server.HTTPPort < 1 || c.Server.HTTPPort > 65535 {
		return fmt.Errorf("invalid http_port: %d", c.Server.HTTPPort)
	}

	// Validate devices
	deviceIDs := make(map[string]bool)
	for _, dev := range c.Devices {
		if dev.ID == "" {
			return fmt.Errorf("device missing id")
		}
		if deviceIDs[dev.ID] {
			return fmt.Errorf("duplicate device id: %s", dev.ID)
		}
		deviceIDs[dev.ID] = true

		if dev.Kind == "" {
			return fmt.Errorf("device %s missing kind", dev.ID)
		}
		if dev.Transport == "" {
			return fmt.Errorf("device %s missing transport", dev.ID)
		}
	}

	// Validate logging
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "fatal": true}
	if !validLevels[c.Logging.Level] {
		return fmt.Errorf("invalid logging level: %s", c.Logging.Level)
	}

	validFormats := map[string]bool{"json": true, "text": true}
	if !validFormats[c.Logging.Format] {
		return fmt.Errorf("invalid logging format: %s", c.Logging.Format)
	}

	// Validate security
	if c.Security.Mode != "development" && c.Security.Mode != "production" {
		return fmt.Errorf("invalid security mode: %s (must be 'development' or 'production')", c.Security.Mode)
	}

	if c.Security.TLS.Enabled {
		if c.Security.TLS.CertFile == "" || c.Security.TLS.KeyFile == "" {
			return fmt.Errorf("TLS enabled but cert_file or key_file not specified")
		}
	}

	return nil
}

// GetDeviceConfig returns device configuration by ID
func (c *Config) GetDeviceConfig(id string) (*DeviceConfig, bool) {
	for _, dev := range c.Devices {
		if dev.ID == id {
			return &dev, true
		}
	}
	return nil, false
}

// GetDevicesByKind returns all devices of a specific kind
func (c *Config) GetDevicesByKind(kind string) []DeviceConfig {
	var devices []DeviceConfig
	for _, dev := range c.Devices {
		if dev.Kind == kind && dev.Enabled {
			devices = append(devices, dev)
		}
	}
	return devices
}

// GetDevicesByTag returns all devices with a specific tag
func (c *Config) GetDevicesByTag(tag string) []DeviceConfig {
	var devices []DeviceConfig
	for _, dev := range c.Devices {
		if !dev.Enabled {
			continue
		}
		for _, t := range dev.Tags {
			if t == tag {
				devices = append(devices, dev)
				break
			}
		}
	}
	return devices
}

// JobConfig contains job execution settings
type JobConfig struct {
	DefaultTimeout time.Duration
	MaxRetries     int
	RetryBackoff   time.Duration
}

// DefaultJobConfig returns default job configuration
func DefaultJobConfig() JobConfig {
	return JobConfig{
		DefaultTimeout: 30 * time.Second,
		MaxRetries:     3,
		RetryBackoff:   1 * time.Second,
	}
}

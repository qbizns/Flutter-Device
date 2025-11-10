package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Macber-eg/Flutter-Device/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration commands",
	Long:  "Commands for working with configuration files",
}

var configValidateCmd = &cobra.Command{
	Use:   "validate <config-file>",
	Short: "Validate configuration file",
	Long:  "Validate a Device Bridge configuration file for correctness",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigValidate,
}

var configShowCmd = &cobra.Command{
	Use:   "show <config-file>",
	Short: "Show configuration",
	Long:  "Display the parsed configuration from a file",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	configFile := args[0]

	fmt.Printf("Validating configuration: %s\n", configFile)

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Printf("❌ Validation failed: %v\n", err)
		return err
	}

	fmt.Println("✅ Configuration is valid")
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Devices:  %d\n", len(cfg.Devices))
	fmt.Printf("  Security: %s mode\n", cfg.Security.Mode)
	fmt.Printf("  TLS:      %v\n", cfg.Security.TLS.Enabled)
	fmt.Printf("  ACL:      %v\n", cfg.Security.ACL.Enabled)

	// Count enabled devices
	enabled := 0
	for _, dev := range cfg.Devices {
		if dev.Enabled {
			enabled++
		}
	}
	fmt.Printf("  Enabled:  %d devices\n", enabled)

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	configFile := args[0]

	fmt.Printf("Loading configuration: %s\n\n", configFile)

	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Print server config
	fmt.Println("Server:")
	fmt.Printf("  gRPC Port:    %d\n", cfg.Server.GRPCPort)
	fmt.Printf("  HTTP Port:    %d\n", cfg.Server.HTTPPort)
	fmt.Printf("  Metrics Port: %d\n", cfg.Server.MetricsPort)
	fmt.Printf("  Host:         %s\n", cfg.Server.Host)

	// Print logging config
	fmt.Println("\nLogging:")
	fmt.Printf("  Level:  %s\n", cfg.Logging.Level)
	fmt.Printf("  Format: %s\n", cfg.Logging.Format)

	// Print security config
	fmt.Println("\nSecurity:")
	fmt.Printf("  Mode: %s\n", cfg.Security.Mode)
	fmt.Printf("  TLS:  %v\n", cfg.Security.TLS.Enabled)
	if cfg.Security.TLS.Enabled {
		fmt.Printf("    Cert: %s\n", cfg.Security.TLS.CertFile)
		fmt.Printf("    Key:  %s\n", cfg.Security.TLS.KeyFile)
		if cfg.Security.TLS.CAFile != "" {
			fmt.Printf("    CA:   %s (mTLS enabled)\n", cfg.Security.TLS.CAFile)
		}
	}
	fmt.Printf("  ACL:  %v\n", cfg.Security.ACL.Enabled)
	if cfg.Security.ACL.Enabled {
		fmt.Printf("    Rules: %d\n", len(cfg.Security.ACL.Rules))
	}

	// Print devices
	fmt.Println("\nDevices:")
	enabled := 0
	disabled := 0
	byKind := make(map[string]int)

	for _, dev := range cfg.Devices {
		if dev.Enabled {
			enabled++
		} else {
			disabled++
		}
		byKind[dev.Kind]++
	}

	fmt.Printf("  Total:    %d\n", len(cfg.Devices))
	fmt.Printf("  Enabled:  %d\n", enabled)
	fmt.Printf("  Disabled: %d\n", disabled)

	if len(byKind) > 0 {
		fmt.Println("\n  By Kind:")
		for kind, count := range byKind {
			fmt.Printf("    %-20s %d\n", kind, count)
		}
	}

	// Print telemetry config
	fmt.Println("\nTelemetry:")
	fmt.Printf("  Metrics: %v\n", cfg.Telemetry.Metrics.Enabled)
	fmt.Printf("  Tracing: %v\n", cfg.Telemetry.Tracing.Enabled)
	if cfg.Telemetry.Tracing.Enabled {
		fmt.Printf("    Exporter: %s\n", cfg.Telemetry.Tracing.Exporter)
		fmt.Printf("    Endpoint: %s\n", cfg.Telemetry.Tracing.Endpoint)
	}

	return nil
}

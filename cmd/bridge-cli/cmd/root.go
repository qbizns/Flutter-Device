package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	serverURL string
	apiKey    string
	verbose   bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "bridge-cli",
	Short: "Device Bridge CLI - Administration tool for Device Bridge v2",
	Long: `Device Bridge CLI is a command-line administration tool for Device Bridge v2.

It allows you to interact with the Device Bridge server to:
- Manage devices (list, get, discover)
- Submit and monitor jobs
- Test device functionality
- Validate configuration
- View server health and metrics`,
	Version: "2.0.0-dev",
}

// Execute adds all child commands to the root command and sets flags appropriately
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Persistent flags (available to all subcommands)
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "localhost:50051", "Device Bridge gRPC server address")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "API key for authentication")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Optional config file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (optional, for reading API key)")

	// Set version template
	rootCmd.SetVersionTemplate("Device Bridge CLI v{{.Version}}\n")
}

// getAPIKey returns the API key from flag or environment
func getAPIKey() string {
	if apiKey != "" {
		return apiKey
	}

	// Try environment variable
	if key := os.Getenv("DEVICE_BRIDGE_API_KEY"); key != "" {
		return key
	}

	return ""
}

// verbosePrintf prints only if verbose mode is enabled
func verbosePrintf(format string, args ...interface{}) {
	if verbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

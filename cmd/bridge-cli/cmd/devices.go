package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "Manage devices",
	Long:  "Commands for managing and interacting with devices",
}

var devicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered devices",
	Long:  "List all devices registered with the Device Bridge server",
	RunE:  runDevicesList,
}

var devicesGetCmd = &cobra.Command{
	Use:   "get <device-id>",
	Short: "Get device details",
	Long:  "Get detailed information about a specific device",
	Args:  cobra.ExactArgs(1),
	RunE:  runDevicesGet,
}

func init() {
	rootCmd.AddCommand(devicesCmd)
	devicesCmd.AddCommand(devicesListCmd)
	devicesCmd.AddCommand(devicesGetCmd)
}

// createClient creates a gRPC client connection
func createClient() (pb.DeviceBridgeClient, *grpc.ClientConn, error) {
	verbosePrintf("Connecting to server: %s\n", serverURL)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create gRPC connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Connect to server
	conn, err := grpc.DialContext(ctx, serverURL, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	client := pb.NewDeviceBridgeClient(conn)

	return client, conn, nil
}

// createContext creates a context with API key if provided
func createContext() context.Context {
	ctx := context.Background()

	// Add API key to metadata if provided
	if key := getAPIKey(); key != "" {
		md := metadata.New(map[string]string{
			"x-api-key": key,
		})
		ctx = metadata.NewOutgoingContext(ctx, md)
		verbosePrintf("Using API key: %s...\n", key[:16])
	}

	return ctx
}

func runDevicesList(cmd *cobra.Command, args []string) error {
	client, conn, err := createClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := createContext()

	verbosePrintf("Fetching device list...\n")

	// Call ListDevices
	resp, err := client.ListDevices(ctx, &pb.ListDevicesRequest{})
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	// Print results in table format
	if len(resp.Devices) == 0 {
		fmt.Println("No devices registered")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DEVICE ID\tNAME\tKIND\tHEALTH\tTAGS")
	fmt.Fprintln(w, "---------\t----\t----\t------\t----")

	for _, device := range resp.Devices {
		health := "unknown"
		if device.Health != nil {
			health = device.Health.Status.String()
		}

		tags := ""
		if len(device.Tags) > 0 {
			for i, tag := range device.Tags {
				if i > 0 {
					tags += ","
				}
				tags += tag
			}
		} else {
			tags = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			device.Id,
			device.Name,
			device.Kind,
			health,
			tags,
		)
	}

	w.Flush()

	fmt.Printf("\nTotal: %d devices\n", len(resp.Devices))

	return nil
}

func runDevicesGet(cmd *cobra.Command, args []string) error {
	deviceID := args[0]

	client, conn, err := createClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := createContext()

	verbosePrintf("Fetching device: %s\n", deviceID)

	// Call GetDevice
	resp, err := client.GetDevice(ctx, &pb.GetDeviceRequest{
		DeviceId: deviceID,
	})
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	device := resp.Device

	// Print device details
	fmt.Printf("Device: %s\n", device.Id)
	fmt.Printf("  Name:     %s\n", device.Name)
	fmt.Printf("  Kind:     %s\n", device.Kind)
	fmt.Printf("  Tags:     ")
	if len(device.Tags) > 0 {
		for i, tag := range device.Tags {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(tag)
		}
		fmt.Println()
	} else {
		fmt.Println("-")
	}

	// Print health
	if device.Health != nil {
		fmt.Printf("\nHealth:\n")
		fmt.Printf("  Status:   %s\n", device.Health.Status.String())
		if device.Health.Message != "" {
			fmt.Printf("  Message:  %s\n", device.Health.Message)
		}
		if device.Health.Timestamp != nil {
			fmt.Printf("  Checked:  %s\n", device.Health.Timestamp.AsTime().Format(time.RFC3339))
		}
	}

	// Print metadata
	if len(device.Metadata) > 0 {
		fmt.Printf("\nMetadata:\n")
		for key, value := range device.Metadata {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	return nil
}

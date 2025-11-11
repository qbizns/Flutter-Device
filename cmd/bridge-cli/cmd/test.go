package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test device functionality",
	Long:  "Commands for testing device functionality",
}

var testWeightCmd = &cobra.Command{
	Use:   "weight <device-id>",
	Short: "Read weight from scale",
	Long:  "Read the current weight from a scale device",
	Args:  cobra.ExactArgs(1),
	RunE:  runTestWeight,
}

var testScanCmd = &cobra.Command{
	Use:   "scan <device-id>",
	Short: "Monitor barcode scan events",
	Long:  "Subscribe to and monitor barcode scan events from a scanner device",
	Args:  cobra.ExactArgs(1),
	RunE:  runTestScan,
}

var (
	scanCount int
	scanTimeout int
)

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.AddCommand(testWeightCmd)
	testCmd.AddCommand(testScanCmd)

	// Flags for scan command
	testScanCmd.Flags().IntVarP(&scanCount, "count", "n", 0, "Exit after receiving N scans (0 = unlimited)")
	testScanCmd.Flags().IntVarP(&scanTimeout, "timeout", "t", 0, "Timeout in seconds (0 = no timeout)")
}

func runTestWeight(cmd *cobra.Command, args []string) error {
	deviceID := args[0]

	client, conn, err := createClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := createContext()

	verbosePrintf("Reading weight from device: %s\n", deviceID)

	// Read weight
	resp, err := client.GetWeight(ctx, &pb.GetWeightRequest{
		DeviceId: deviceID,
	})
	if err != nil {
		return fmt.Errorf("failed to read weight: %w", err)
	}

	reading := resp.Reading

	// Print weight reading
	fmt.Printf("\n")
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║                   WEIGHT READING                         ║\n")
	fmt.Printf("╠══════════════════════════════════════════════════════════╣\n")
	fmt.Printf("║  Device:  %-46s ║\n", deviceID)
	fmt.Printf("║  Weight:  %.3f %-40s ║\n", reading.Weight, reading.Unit.String())
	fmt.Printf("║  Stable:  %-46v ║\n", reading.Stable)
	if reading.Timestamp != nil {
		fmt.Printf("║  Time:    %-46s ║\n", reading.Timestamp.AsTime().Format("15:04:05"))
	}
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")

	return nil
}

func runTestScan(cmd *cobra.Command, args []string) error {
	deviceID := args[0]

	client, conn, err := createClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := createContext()

	verbosePrintf("Subscribing to scan events from device: %s\n", deviceID)

	// Subscribe to scan events
	stream, err := client.SubscribeScanEvents(ctx, &pb.SubscribeScanEventsRequest{
		DeviceId: deviceID,
	})
	if err != nil {
		return fmt.Errorf("failed to subscribe to scan events: %w", err)
	}

	// Print header
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                      BARCODE SCANNER - MONITORING                         ║")
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Device:   %-62s ║\n", deviceID)
	if scanCount > 0 {
		fmt.Printf("║  Count:    %-62d ║\n", scanCount)
	} else {
		fmt.Printf("║  Count:    %-62s ║\n", "Unlimited (Ctrl+C to exit)")
	}
	if scanTimeout > 0 {
		fmt.Printf("║  Timeout:  %-62s ║\n", fmt.Sprintf("%d seconds", scanTimeout))
	}
	fmt.Println("╠═══════════════════════════════════════════════════════════════════════════╣")
	fmt.Println("║  Waiting for barcode scans... Please scan a barcode.                      ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Setup timeout if specified
	var timeoutChan <-chan time.Time
	if scanTimeout > 0 {
		timeoutChan = time.After(time.Duration(scanTimeout) * time.Second)
	}

	count := 0
	startTime := time.Now()

	for {
		// Check timeout
		if timeoutChan != nil {
			select {
			case <-timeoutChan:
				fmt.Println()
				fmt.Printf("⏱️  Timeout reached after %d seconds\n", scanTimeout)
				fmt.Printf("📊 Total scans: %d\n", count)
				return nil
			default:
			}
		}

		// Check count limit
		if scanCount > 0 && count >= scanCount {
			fmt.Println()
			elapsed := time.Since(startTime)
			fmt.Printf("✅ Received %d scans in %.2f seconds\n", count, elapsed.Seconds())
			if count > 0 {
				fmt.Printf("📊 Average: %.2f scans/second\n", float64(count)/elapsed.Seconds())
			}
			return nil
		}

		// Receive scan event
		event, err := stream.Recv()
		if err == io.EOF {
			fmt.Println()
			fmt.Println("📡 Stream closed by server")
			break
		}
		if err != nil {
			fmt.Println()
			return fmt.Errorf("error receiving scan event: %w", err)
		}

		count++

		// Print scan event
		timestamp := time.Now()
		if event.Timestamp != nil {
			timestamp = event.Timestamp.AsTime()
		}

		fmt.Printf("┌─ Scan #%-4d ────────────────────────────────────────────────────────────┐\n", count)
		fmt.Printf("│  Time:      %s                                                 │\n", timestamp.Format("15:04:05.000"))
		fmt.Printf("│  Barcode:   %-62s │\n", event.Data)
		fmt.Printf("│  Symbology: %-62s │\n", event.Symbology)
		fmt.Printf("└───────────────────────────────────────────────────────────────────────────┘\n")
		fmt.Println()
	}

	if count == 0 {
		fmt.Println("No scans received")
	} else {
		elapsed := time.Since(startTime)
		fmt.Printf("Total: %d scans in %.2f seconds\n", count, elapsed.Seconds())
	}

	return nil
}

package cmd

import (
	"fmt"

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

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.AddCommand(testWeightCmd)
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

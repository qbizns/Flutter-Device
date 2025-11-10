package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check server health",
	Long:  "Check the health status of the Device Bridge server",
	RunE:  runHealth,
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the server",
	Long:  "Send a ping request to check server connectivity",
	RunE:  runPing,
}

func init() {
	rootCmd.AddCommand(healthCmd)
	rootCmd.AddCommand(pingCmd)
}

func runHealth(cmd *cobra.Command, args []string) error {
	client, conn, err := createClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := createContext()

	verbosePrintf("Checking server health...\n")

	start := time.Now()

	// Call ListDevices to check health
	resp, err := client.ListDevices(ctx, &pb.ListDevicesRequest{})
	if err != nil {
		fmt.Printf("❌ Server unhealthy: %v\n", err)
		return err
	}

	elapsed := time.Since(start)

	fmt.Println("✅ Server is healthy")
	fmt.Printf("   Devices: %d registered\n", len(resp.Devices))
	fmt.Printf("   Latency: %s\n", elapsed.Round(time.Millisecond))

	return nil
}

func runPing(cmd *cobra.Command, args []string) error {
	client, conn, err := createClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := createContext()

	verbosePrintf("Pinging server...\n")

	start := time.Now()

	// Call Ping
	resp, err := client.Ping(ctx, &pb.PingRequest{})
	if err != nil {
		fmt.Printf("❌ Ping failed: %v\n", err)
		return err
	}

	elapsed := time.Since(start)

	fmt.Println("✅ Pong!")
	fmt.Printf("   Message: %s\n", resp.Message)
	fmt.Printf("   Latency: %s\n", elapsed.Round(time.Millisecond))

	return nil
}

package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/Macber-eg/Flutter-Device/proto/devicebridge/v1"
)

var jobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage jobs",
	Long:  "Commands for managing and monitoring jobs",
}

var jobsGetCmd = &cobra.Command{
	Use:   "get <job-id>",
	Short: "Get job details",
	Long:  "Get detailed information about a specific job",
	Args:  cobra.ExactArgs(1),
	RunE:  runJobsGet,
}

func init() {
	rootCmd.AddCommand(jobsCmd)
	jobsCmd.AddCommand(jobsGetCmd)
}

func runJobsGet(cmd *cobra.Command, args []string) error {
	jobID := args[0]

	client, conn, err := createClient()
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := createContext()

	verbosePrintf("Fetching job: %s\n", jobID)

	// Call GetJob
	resp, err := client.GetJob(ctx, &pb.GetJobRequest{
		JobId: jobID,
	})
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}

	job := resp.Job

	// Print job details
	fmt.Printf("Job: %s\n", job.JobId)
	fmt.Printf("  Device:   %s\n", job.DeviceId)
	fmt.Printf("  Type:     %s\n", job.Type)
	fmt.Printf("  Status:   %s\n", job.Status.String())

	if job.CreatedAt != nil {
		fmt.Printf("  Created:  %s\n", job.CreatedAt.AsTime().Format(time.RFC3339))
	}
	if job.StartedAt != nil {
		fmt.Printf("  Started:  %s\n", job.StartedAt.AsTime().Format(time.RFC3339))
	}
	if job.CompletedAt != nil {
		fmt.Printf("  Completed: %s\n", job.CompletedAt.AsTime().Format(time.RFC3339))
	}

	if job.Error != "" {
		fmt.Printf("  Error:    %s\n", job.Error)
	}

	// Calculate duration if job is completed
	if job.StartedAt != nil && job.CompletedAt != nil {
		duration := job.CompletedAt.AsTime().Sub(job.StartedAt.AsTime())
		fmt.Printf("  Duration: %s\n", duration.Round(time.Millisecond))
	}

	return nil
}

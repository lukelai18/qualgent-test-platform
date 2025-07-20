package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "qualgent-test-platform/api/proto"
)
var (
	serverAddr string
	orgID      string
	appVersionID string
	testPath   string
	priority   int32
	target     string
	jobID      string
	jsonOutput bool
	webAppURL  string
	testType   string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "qgjob",
		Short: "A CLI for interacting with the QualGent job server",
		Long:  `qgjob is a command-line tool for submitting and monitoring test jobs on the QualGent test platform.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverAddr, "server", "localhost:8080", "RPC server address")

	// Submit command
	submitCmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a new test job",
		Long:  `Submit a new test job to the QualGent test platform.`,
		RunE:  submitJob,
	}
	submitCmd.Flags().StringVar(&orgID, "org-id", "", "Organization ID (required)")
	submitCmd.Flags().StringVar(&appVersionID, "app-version-id", "", "Application version ID (required)")
	submitCmd.Flags().StringVar(&testPath, "test", "", "Test file path (required)")
	submitCmd.Flags().Int32Var(&priority, "priority", 0, "Job priority (0-10)")
	submitCmd.Flags().StringVar(&target, "target", "emulator", "Execution target (emulator|device|browserstack|web)")
	submitCmd.Flags().StringVar(&webAppURL, "web-app-url", "", "URL of the web application (required for web target)")
	submitCmd.Flags().StringVar(&testType, "test-type", "", "Type of test (PLAYWRIGHT|ESPRESSO)")
	submitCmd.MarkFlagRequired("org-id")
	submitCmd.MarkFlagRequired("test")

	// Conditional required flags
	submitCmd.MarkFlagsMutuallyExclusive("app-version-id", "web-app-url")
	// For web target, web-app-url and test-type are required together
	// For other targets, app-version-id is required

	// Status command
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get the status of a job",
		Long:  `Get the current status of a test job by its ID.`,
		RunE:  getJobStatus,
	}
	statusCmd.Flags().StringVar(&jobID, "job-id", "", "Job ID (required)")
	statusCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	statusCmd.MarkFlagRequired("job-id")

	rootCmd.AddCommand(submitCmd, statusCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func submitJob(cmd *cobra.Command, args []string) error {
	// Validate target
	if target != "emulator" && target != "device" && target != "browserstack" && target != "web" {
		return fmt.Errorf("invalid target: %s. Must be one of: emulator, device, browserstack, web", target)
	}

	// Validate conditional required flags
	if target == "web" && webAppURL == "" {
		return fmt.Errorf("web-app-url is required for web target")
	}
	if target != "web" && appVersionID == "" {
		return fmt.Errorf("app-version-id is required for non-web targets")
	}

	// Validate test type
	if testType != "" && testType != "PLAYWRIGHT" && testType != "ESPRESSO" {
		return fmt.Errorf("invalid test-type: %s. Must be PLAYWRIGHT or ESPRESSO", testType)
	}

	// Validate priority
	if priority < 0 || priority > 10 {
		return fmt.Errorf("invalid priority: %d. Must be between 0 and 10", priority)
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	client := pb.NewJobServiceClient(conn)

	// Create request
	req := &pb.SubmitJobRequest{
		OrgId:          orgID,
		AppVersionId:   appVersionID,
		TestPath:       testPath,
		Priority:       priority,
		Target:         parseTarget(target),
		IdempotencyKey: uuid.New().String(),
		WebAppUrl:      webAppURL,
		TestType:       parseTestType(testType),
	}

	// Submit job
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.SubmitJob(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to submit job: %w", err)
	}

	// Output result
	if jsonOutput {
		output := map[string]interface{}{
			"job_id": resp.JobId,
			"status": resp.Status.String(),
		}
		jsonBytes, _ := json.Marshal(output)
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Printf("Job submitted successfully!\n")
		fmt.Printf("Job ID: %s\n", resp.JobId)
		fmt.Printf("Status: %s\n", resp.Status.String())
	}

	return nil
}

func getJobStatus(cmd *cobra.Command, args []string) error {
	// Connect to gRPC server
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer conn.Close()

	client := pb.NewJobServiceClient(conn)

	// Create request
	req := &pb.GetJobStatusRequest{
		JobId: jobID,
	}

	// Get job status
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.GetJobStatus(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to get job status: %w", err)
	}

	// Output result
	if jsonOutput {
		output := map[string]interface{}{
			"job_id":     resp.JobId,
			"status":     resp.Status.String(),
			"created_at": resp.CreatedAt.AsTime().Format(time.RFC3339),
			"completed_at": func() string { if resp.CompletedAt != nil { return resp.CompletedAt.AsTime().Format(time.RFC3339) } else { return "" } }(),
			"logs_url":   resp.LogsUrl,
			"session_id": resp.SessionId,
			"video_url":  resp.VideoUrl,
			"error_message": resp.ErrorMessage,
			"test_duration": resp.TestDuration,
		}
		jsonBytes, _ := json.Marshal(output)
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Printf("Job Status:\n")
		fmt.Printf("Job ID: %s\n", resp.JobId)
		fmt.Printf("Status: %s\n", resp.Status.String())
		fmt.Printf("Created: %s\n", resp.CreatedAt.AsTime().Format(time.RFC3339))
		if resp.CompletedAt != nil {
			fmt.Printf("Completed: %s\n", resp.CompletedAt.AsTime().Format(time.RFC3339))
		}
		if resp.SessionId != "" {
			fmt.Printf("Session ID: %s\n", resp.SessionId)
		}
		if resp.LogsUrl != "" {
			fmt.Printf("Logs URL: %s\n", resp.LogsUrl)
		}
		if resp.VideoUrl != "" {
			fmt.Printf("Video URL: %s\n", resp.VideoUrl)
		}
		if resp.ErrorMessage != "" {
			fmt.Printf("Error: %s\n", resp.ErrorMessage)
		}
		if resp.TestDuration > 0 {
			fmt.Printf("Duration: %d seconds\n", resp.TestDuration)
		}
	}

	return nil
}

func parseTarget(target string) pb.Target {
	switch target {
	case "emulator":
		return pb.Target_EMULATOR
	case "device":
		return pb.Target_DEVICE
	case "browserstack":
		return pb.Target_BROWSERSTACK
	case "web":
		return pb.Target_WEB
	default:
		return pb.Target_TARGET_UNSPECIFIED
	}
}

func parseTestType(testType string) pb.TestType {
	switch testType {
	case "PLAYWRIGHT":
		return pb.TestType_PLAYWRIGHT
	case "ESPRESSO":
		return pb.TestType_ESPRESSO
	default:
		return pb.TestType_TEST_TYPE_UNSPECIFIED
	}
}
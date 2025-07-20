package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "qualgent-test-platform/api/proto"
)

type AppWrightAgent struct {
	client       pb.JobServiceClient
	agentID      string
	hostname     string
	browserStack *BrowserStackClient
}

type BrowserStackClient struct {
	username string
	accessKey string
	baseURL   string
	httpClient *http.Client
}

type AppWrightTestConfig struct {
	AppVersionID string `json:"app_version_id"`
	TestPath     string `json:"test_path"`
	Target       string `json:"target"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

type TestResult struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	LogsURL   string `json:"logs_url"`
	VideoURL  string `json:"video_url"`
}

func NewAppWrightAgent(serverAddr, hostname string) (*AppWrightAgent, error) {
	// Connect to gRPC server
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	client := pb.NewJobServiceClient(conn)

	// Initialize BrowserStack client
	browserStack := &BrowserStackClient{
		username:   os.Getenv("BROWSERSTACK_USERNAME"),
		accessKey:  os.Getenv("BROWSERSTACK_ACCESS_KEY"),
		baseURL:    "https://api-cloud.browserstack.com/app-automate/v2",
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	if browserStack.username == "" || browserStack.accessKey == "" {
		return nil, fmt.Errorf("BROWSERSTACK_USERNAME and BROWSERSTACK_ACCESS_KEY environment variables are required")
	}

	agent := &AppWrightAgent{
		client:       client,
		agentID:      uuid.New().String(),
		hostname:     hostname,
		browserStack: browserStack,
	}

	// Register agent with server
	if err := agent.register(); err != nil {
		return nil, fmt.Errorf("failed to register agent: %w", err)
	}

	return agent, nil
}

func (a *AppWrightAgent) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &pb.RegisterAgentRequest{
		Hostname:         a.hostname,
		TargetCapability: "browserstack",
	}

	resp, err := a.client.RegisterAgent(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	log.Printf("Registered agent with ID: %s", resp.AgentId)
	return nil
}

func (a *AppWrightAgent) Start(ctx context.Context) error {
	log.Printf("AppWright Agent started on %s", a.hostname)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := a.processJobs(ctx); err != nil {
				log.Printf("Failed to process jobs: %v", err)
			}
		}
	}
}

func (a *AppWrightAgent) processJobs(ctx context.Context) error {
	// Fetch a job from the server
	req := &pb.FetchJobRequest{
		TargetCapability: "browserstack",
	}
	job, err := a.client.FetchJob(ctx, req)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			log.Println("No jobs available for target: browserstack")
			return nil
		}
		return fmt.Errorf("failed to fetch job: %w", err)
	}

	log.Printf("Processing job %s", job.JobId)

	// Update job status to RUNNING
	updateReq := &pb.UpdateJobStatusRequest{
		JobId:  job.JobId,
		Status: pb.Status_RUNNING,
		AgentId: a.agentID,
	}
	if _, err := a.client.UpdateJobStatus(ctx, updateReq); err != nil {
		log.Printf("Failed to update job %s to RUNNING: %v", job.JobId, err)
		return nil // Don't proceed with a job we can't update
	}

	// Execute the test
	result, err := a.executeAppWrightTest(ctx, &pb.SubmitJobRequest{
		AppVersionId: job.AppVersionId,
		TestPath:     job.TestPath,
	})

	// Update job status based on the result
	finalStatus := pb.Status_COMPLETED
	if err != nil {
		log.Printf("Test failed for job %s: %v", job.JobId, err)
		finalStatus = pb.Status_FAILED
	} else {
		log.Printf("Test completed for job %s with status: %s", job.JobId, result.Status)
	}

	updateReq.Status = finalStatus
	if _, err := a.client.UpdateJobStatus(ctx, updateReq); err != nil {
		log.Printf("Failed to update final status for job %s: %v", job.JobId, err)
	}

	return nil
}

func (a *AppWrightAgent) updateHeartbeat(ctx context.Context) error {
	// For heartbeat, we don't need to update job status
	// Just keep the agent alive in the system
	return nil
}

func (a *AppWrightAgent) executeAppWrightTest(ctx context.Context, job *pb.SubmitJobRequest) (*TestResult, error) {
	// Create AppWright test configuration with minimal payload
	payload := map[string]interface{}{
		"app": job.AppVersionId, // This should be the app URL or app ID
		"devices": []string{"Google Pixel 3"},
	}

	// Submit test to BrowserStack App Automate
	sessionID, err := a.browserStack.StartSession(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to start BrowserStack session: %w", err)
	}

	// Monitor test execution
	result, err := a.browserStack.WaitForCompletion(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for test completion: %w", err)
	}

	return result, nil
}

// BrowserStack API methods
func (bs *BrowserStackClient) StartSession(payload map[string]interface{}) (string, error) {
	url := fmt.Sprintf("%s/builds", bs.baseURL)
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(bs.username, bs.accessKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := bs.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to start session, status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	sessionID, ok := result["session_id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid session_id in response")
	}

	return sessionID, nil
}

func (bs *BrowserStackClient) WaitForCompletion(sessionID string) (*TestResult, error) {
	url := fmt.Sprintf("%s/sessions/%s", bs.baseURL, sessionID)
	
	// Poll for completion
	for i := 0; i < 60; i++ { // Max 10 minutes
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.SetBasicAuth(bs.username, bs.accessKey)

		resp, err := bs.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to get session status, status: %d", resp.StatusCode)
		}

		var session map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		status, ok := session["status"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid status in response")
		}

		if status == "completed" || status == "failed" {
			return &TestResult{
				SessionID: sessionID,
				Status:    status,
				LogsURL:   fmt.Sprintf("https://app-automate.browserstack.com/dashboard/v2/builds/%s/sessions/%s", sessionID, sessionID),
				VideoURL:  fmt.Sprintf("https://app-automate.browserstack.com/dashboard/v2/builds/%s/sessions/%s/video", sessionID, sessionID),
			}, nil
		}

		time.Sleep(10 * time.Second)
	}

	return nil, fmt.Errorf("test execution timeout")
} 
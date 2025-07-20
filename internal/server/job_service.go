package server

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"qualgent-test-platform/internal/store"
	pb "qualgent-test-platform/api/proto"
)

type JobService struct {
	pb.UnimplementedJobServiceServer
	postgresStore *store.PostgresStore
	redisStore    *store.RedisStore
}

func NewJobService(postgresStore *store.PostgresStore, redisStore *store.RedisStore) *JobService {
	return &JobService{
		postgresStore: postgresStore,
		redisStore:    redisStore,
	}
}

func (s *JobService) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
	// Validate request
	if req.OrgId == "" || req.TestPath == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id and test_path are required")
	}

	if req.Target == pb.Target_WEB && req.WebAppUrl == "" {
		return nil, status.Error(codes.InvalidArgument, "web_app_url is required for web target")
	}

	if req.Target != pb.Target_WEB && req.AppVersionId == "" {
		return nil, status.Error(codes.InvalidArgument, "app_version_id is required for non-web targets")
	}

	// Check idempotency if provided
	if req.IdempotencyKey != "" {
		processed, err := s.redisStore.CheckIdempotency(ctx, req.IdempotencyKey)
		if err != nil {
			log.Printf("Failed to check idempotency: %v", err)
		} else if processed {
			return nil, status.Error(codes.AlreadyExists, "job with this idempotency key already exists")
		}
	}

	// Create job
	job := &store.Job{
		OrgID:          req.OrgId,
		AppVersionID:   req.AppVersionId,
		TestPath:       req.TestPath,
		Priority:       req.Priority,
		Target:         targetToString(req.Target),
		Status:         "PENDING",
		IdempotencyKey: &req.IdempotencyKey,
		WebAppURL:      &req.WebAppUrl,
		TestType:       testTypeToString(req.TestType),
	}

	if err := s.postgresStore.CreateJob(ctx, job); err != nil {
		log.Printf("Failed to create job: %v", err)
		return nil, status.Error(codes.Internal, "failed to create job")
	}

	// Set idempotency key if provided
	if req.IdempotencyKey != "" {
		if err := s.redisStore.SetIdempotency(ctx, req.IdempotencyKey, 24*time.Hour); err != nil {
			log.Printf("Failed to set idempotency key: %v", err)
		}
	}

	// Push to ingestion queue
	if err := s.redisStore.PushToIngestionQueue(ctx, job.ID); err != nil {
		log.Printf("Failed to push to ingestion queue: %v", err)
		// Don't fail the request, just log the error
	}

	log.Printf("Created job %s for org %s, app version %s", job.ID, req.OrgId, req.AppVersionId)

	return &pb.SubmitJobResponse{
		JobId:  job.ID.String(),
		Status: stringToStatus(job.Status),
	}, nil
}

func (s *JobService) GetJobStatus(ctx context.Context, req *pb.GetJobStatusRequest) (*pb.GetJobStatusResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}

	jobID, err := uuid.Parse(req.JobId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid job_id format")
	}

	// Try to get from cache first
	statusStr, err := s.redisStore.GetJobStatus(ctx, jobID)
	if err == nil {
		// Get job details from database for timestamp
		job, err := s.postgresStore.GetJob(ctx, jobID)
		if err == nil {
			return &pb.GetJobStatusResponse{
				JobId:      req.JobId,
				Status:     stringToStatus(statusStr),
				CreatedAt:  timestamppb.New(job.CreatedAt),
			}, nil
		}
	}

	// Get from database
	job, err := s.postgresStore.GetJob(ctx, jobID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "job not found")
	}

	// Cache the status
	if err := s.redisStore.SetJobStatus(ctx, jobID, job.Status, 5*time.Minute); err != nil {
		log.Printf("Failed to cache job status: %v", err)
	}

	response := &pb.GetJobStatusResponse{
		JobId:      job.ID.String(),
		Status:     stringToStatus(job.Status),
		CreatedAt:  timestamppb.New(job.CreatedAt),
	}
	
	if job.CompletedAt != nil {
		response.CompletedAt = timestamppb.New(*job.CompletedAt)
	}
	if job.SessionID != nil {
		response.SessionId = *job.SessionID
	}
	if job.LogsURL != nil {
		response.LogsUrl = *job.LogsURL
	}
	if job.VideoURL != nil {
		response.VideoUrl = *job.VideoURL
	}
	if job.ErrorMessage != nil {
		response.ErrorMessage = *job.ErrorMessage
	}
	if job.TestDuration != nil {
		response.TestDuration = *job.TestDuration
	}
	
	// Debug logging
	log.Printf("Job %s: status=%s, session_id=%s, logs_url=%s, video_url=%s, test_duration=%d", 
		job.ID, job.Status, 
		func() string { if job.SessionID != nil { return *job.SessionID } else { return "nil" } }(),
		func() string { if job.LogsURL != nil { return *job.LogsURL } else { return "nil" } }(),
		func() string { if job.VideoURL != nil { return *job.VideoURL } else { return "nil" } }(),
		func() int32 { if job.TestDuration != nil { return *job.TestDuration } else { return 0 } }(),
	)
	
	return response, nil
}

func (s *JobService) RegisterAgent(ctx context.Context, req *pb.RegisterAgentRequest) (*pb.RegisterAgentResponse, error) {
	if req.Hostname == "" || req.TargetCapability == "" {
		return nil, status.Error(codes.InvalidArgument, "hostname and target_capability are required")
	}

	agent := &store.Agent{
		Hostname:         req.Hostname,
		TargetCapability: req.TargetCapability,
		Status:         "IDLE",
	}

	if err := s.postgresStore.CreateAgent(ctx, agent); err != nil {
		log.Printf("Failed to create agent: %v", err)
		return nil, status.Error(codes.Internal, "failed to register agent")
	}

	// Set initial heartbeat
	if err := s.redisStore.UpdateAgentHeartbeat(ctx, agent.ID, 2*time.Minute); err != nil {
		log.Printf("Failed to set initial heartbeat: %v", err)
	}

	log.Printf("Registered agent %s with capability %s", agent.ID, req.TargetCapability)

	return &pb.RegisterAgentResponse{
		AgentId: agent.ID.String(),
	}, nil
}

func (s *JobService) UpdateJobStatus(ctx context.Context, req *pb.UpdateJobStatusRequest) (*pb.UpdateJobStatusResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}

	jobID, err := uuid.Parse(req.JobId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid job_id format")
	}

	// Update job status
	statusStr := statusToString(req.Status)
	if err := s.postgresStore.UpdateJobStatus(ctx, jobID, statusStr); err != nil {
		log.Printf("Failed to update job status: %v", err)
		return nil, status.Error(codes.Internal, "failed to update job status")
	}

	// Update cache
	if err := s.redisStore.SetJobStatus(ctx, jobID, statusStr, 5*time.Minute); err != nil {
		log.Printf("Failed to update job status cache: %v", err)
	}

	// Update agent heartbeat if provided
	if req.AgentId != "" {
		agentID, err := uuid.Parse(req.AgentId)
		if err == nil {
			if err := s.redisStore.UpdateAgentHeartbeat(ctx, agentID, 2*time.Minute); err != nil {
				log.Printf("Failed to update agent heartbeat: %v", err)
			}
		}
	}

	log.Printf("Updated job %s status to %s", req.JobId, statusStr)

	return &pb.UpdateJobStatusResponse{
		Success: true,
	}, nil
}

func (s *JobService) FetchJob(ctx context.Context, req *pb.FetchJobRequest) (*pb.FetchJobResponse, error) {
	if req.TargetCapability == "" {
		return nil, status.Error(codes.InvalidArgument, "target_capability is required")
	}

	// For simplicity, this example fetches the first available job.
	// A real implementation would have more sophisticated logic.
	job, err := s.postgresStore.GetNextJob(ctx, req.TargetCapability)
	if err != nil {
		log.Printf("Failed to get next job: %v", err)
		return nil, status.Error(codes.Internal, "failed to get next job")
	}

	if job == nil {
		return nil, status.Error(codes.NotFound, "no jobs available")
	}

	return &pb.FetchJobResponse{
		JobId:          job.ID.String(),
		OrgId:          job.OrgID,
		AppVersionId:   job.AppVersionID,
		TestPath:       job.TestPath,
		Priority:       job.Priority,
		Target:         stringToTarget(job.Target),
		WebAppUrl:      *job.WebAppURL,
		TestType:       stringToTestType(*job.TestType),
	}, nil
}

// Helper functions for converting between protobuf and string representations
func targetToString(target pb.Target) string {
	switch target {
	case pb.Target_EMULATOR:
		return "emulator"
	case pb.Target_DEVICE:
		return "device"
	case pb.Target_BROWSERSTACK:
		return "browserstack"
	case pb.Target_WEB:
		return "web"
	default:
		return "unspecified"
	}
}

func stringToStatus(status string) pb.Status {
	switch status {
	case "PENDING":
		return pb.Status_PENDING
	case "SCHEDULED":
		return pb.Status_SCHEDULED
	case "ASSIGNED":
		return pb.Status_ASSIGNED
	case "RUNNING":
		return pb.Status_RUNNING
	case "COMPLETED":
		return pb.Status_COMPLETED
	case "FAILED":
		return pb.Status_FAILED
	case "RETRYING":
		return pb.Status_RETRYING
	default:
		return pb.Status_STATUS_UNSPECIFIED
	}
}

func statusToString(status pb.Status) string {
	switch status {
	case pb.Status_PENDING:
		return "PENDING"
	case pb.Status_SCHEDULED:
		return "SCHEDULED"
	case pb.Status_ASSIGNED:
		return "ASSIGNED"
	case pb.Status_RUNNING:
		return "RUNNING"
	case pb.Status_COMPLETED:
		return "COMPLETED"
	case pb.Status_FAILED:
		return "FAILED"
	case pb.Status_RETRYING:
		return "RETRYING"
	default:
		return "UNSPECIFIED"
	}
}

func stringToTarget(target string) pb.Target {
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

func testTypeToString(testType pb.TestType) string {
	switch testType {
	case pb.TestType_PLAYWRIGHT:
		return "PLAYWRIGHT"
	case pb.TestType_ESPRESSO:
		return "ESPRESSO"
	default:
		return "UNSPECIFIED"
	}
}

func stringToTestType(testType string) pb.TestType {
	switch testType {
	case "PLAYWRIGHT":
		return pb.TestType_PLAYWRIGHT
	case "ESPRESSO":
		return pb.TestType_ESPRESSO
	default:
		return pb.TestType_TEST_TYPE_UNSPECIFIED
	}
}
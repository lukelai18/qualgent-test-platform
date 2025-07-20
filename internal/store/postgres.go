package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

type Job struct {
	ID             uuid.UUID  `json:"id"`
	OrgID          string     `json:"org_id"`
	AppVersionID   string     `json:"app_version_id"`
	TestPath       string     `json:"test_path"`
	Priority       int32      `json:"priority"`
	Target         string     `json:"target"`
	Status         string     `json:"status"`
	JobGroupID     *uuid.UUID `json:"job_group_id,omitempty"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty"`
	SessionID      *string    `json:"session_id,omitempty"`
	LogsURL        *string    `json:"logs_url,omitempty"`
	VideoURL       *string    `json:"video_url,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	TestDuration   *int32     `json:"test_duration,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	WebAppURL      *string    `json:"web_app_url,omitempty"` // New field
	TestType       *string    `json:"test_type,omitempty"`   // New field
}

type JobGroup struct {
	ID           uuid.UUID  `json:"id"`
	AppVersionID string     `json:"app_version_id"`
	Target       string     `json:"target"`
	Status       string     `json:"status"`
	AgentID      *uuid.UUID `json:"agent_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type Agent struct {
	ID               uuid.UUID `json:"id"`
	Hostname         string    `json:"hostname"`
	TargetCapability string    `json:"target_capability"`
	Status           string    `json:"status"`
	LastHeartbeatAt  time.Time `json:"last_heartbeat_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

// Job operations
func (s *PostgresStore) CreateJob(ctx context.Context, job *Job) error {
	query := `
		INSERT INTO jobs (org_id, app_version_id, test_path, priority, target, status, idempotency_key, web_app_url, test_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	var id uuid.UUID
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(ctx, query,
		job.OrgID, job.AppVersionID, job.TestPath, job.Priority, job.Target, job.Status, job.IdempotencyKey, job.WebAppURL, job.TestType,
	).Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	job.ID = id
	job.CreatedAt = createdAt
	job.UpdatedAt = updatedAt
	return nil
}

func (s *PostgresStore) GetJob(ctx context.Context, id uuid.UUID) (*Job, error) {
	query := `
		SELECT id, org_id, app_version_id, test_path, priority, target, status, job_group_id, idempotency_key,
		       session_id, logs_url, video_url, error_message, test_duration, created_at, updated_at, completed_at,
		       web_app_url, test_type
		FROM jobs WHERE id = $1
	`

	job := &Job{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.OrgID, &job.AppVersionID, &job.TestPath, &job.Priority, &job.Target, &job.Status,
		&job.JobGroupID, &job.IdempotencyKey, &job.SessionID, &job.LogsURL, &job.VideoURL,
		&job.ErrorMessage, &job.TestDuration, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt,
		&job.WebAppURL, &job.TestType,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

func (s *PostgresStore) UpdateJobStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE jobs SET status = $1 WHERE id = $2`
	_, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateJobResult(ctx context.Context, id uuid.UUID, result *JobResult) error {
	query := `
		UPDATE jobs
		SET status = $1, session_id = $2, logs_url = $3, video_url = $4,
		    error_message = $5, test_duration = $6, completed_at = NOW()
		WHERE id = $7
	`
	_, err := s.db.ExecContext(ctx, query,
		result.Status, result.SessionID, result.LogsURL, result.VideoURL,
		result.ErrorMessage, result.TestDuration, id)
	if err != nil {
		return fmt.Errorf("failed to update job result: %w", err)
	}
	return nil
}

type JobResult struct {
	Status       string  `json:"status"`
	SessionID    *string `json:"session_id,omitempty"`
	LogsURL      *string `json:"logs_url,omitempty"`
	VideoURL     *string `json:"video_url,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
	TestDuration *int32  `json:"test_duration,omitempty"`
}

func (s *PostgresStore) GetPendingJobs(ctx context.Context, limit int) ([]*Job, error) {
	query := `
		SELECT id, org_id, app_version_id, test_path, priority, target, status, job_group_id, idempotency_key, created_at, updated_at, web_app_url, test_type
		FROM jobs
		WHERE status = 'PENDING'
		ORDER BY priority DESC, created_at ASC
		LIMIT $1
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		job := &Job{}
		err := rows.Scan(
			&job.ID, &job.OrgID, &job.AppVersionID, &job.TestPath, &job.Priority, &job.Target, &job.Status,
			&job.JobGroupID, &job.IdempotencyKey, &job.CreatedAt, &job.UpdatedAt, &job.WebAppURL, &job.TestType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (s *PostgresStore) GetNextJob(ctx context.Context, targetCapability string) (*Job, error) {
	query := `
		SELECT id, org_id, app_version_id, test_path, priority, target, status, job_group_id, idempotency_key, created_at, updated_at, web_app_url, test_type
		FROM jobs
		WHERE status = 'SCHEDULED' AND target = $1
		ORDER BY priority DESC, created_at ASC
		LIMIT 1
	`

	job := &Job{}
	err := s.db.QueryRowContext(ctx, query, targetCapability).Scan(
		&job.ID, &job.OrgID, &job.AppVersionID, &job.TestPath, &job.Priority, &job.Target, &job.Status,
		&job.JobGroupID, &job.IdempotencyKey, &job.CreatedAt, &job.UpdatedAt, &job.WebAppURL, &job.TestType,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No job available
		}
		return nil, fmt.Errorf("failed to get next job: %w", err)
	}

	return job, nil
}

// JobGroup operations
func (s *PostgresStore) CreateJobGroup(ctx context.Context, group *JobGroup) error {
	query := `
		INSERT INTO job_groups (app_version_id, target, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	var id uuid.UUID
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(ctx, query, group.AppVersionID, group.Target, group.Status).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return fmt.Errorf("failed to create job group: %w", err)
	}

	group.ID = id
	group.CreatedAt = createdAt
	group.UpdatedAt = updatedAt
	return nil
}

func (s *PostgresStore) UpdateJobsToGroup(ctx context.Context, jobIDs []uuid.UUID, groupID uuid.UUID) error {
	query := `UPDATE jobs SET job_group_id = $1, status = 'SCHEDULED' WHERE id = ANY($2)`
	_, err := s.db.ExecContext(ctx, query, groupID, pq.Array(jobIDs))
	if err != nil {
		return fmt.Errorf("failed to update jobs to group: %w", err)
	}
	return nil
}

// Agent operations
func (s *PostgresStore) CreateAgent(ctx context.Context, agent *Agent) error {
	query := `
		INSERT INTO agents (hostname, target_capability, status)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	var id uuid.UUID
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(ctx, query, agent.Hostname, agent.TargetCapability, agent.Status).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	agent.ID = id
	agent.CreatedAt = createdAt
	agent.UpdatedAt = updatedAt
	return nil
}

func (s *PostgresStore) UpdateAgentHeartbeat(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE agents SET last_heartbeat_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to update agent heartbeat: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetAvailableAgents(ctx context.Context, targetCapability string) ([]*Agent, error) {
	query := `
		SELECT id, hostname, target_capability, status, last_heartbeat_at, created_at, updated_at
		FROM agents
		WHERE target_capability = $1 AND status = 'IDLE' AND last_heartbeat_at > NOW() - INTERVAL '5 minutes'
	`

	rows, err := s.db.QueryContext(ctx, query, targetCapability)
	if err != nil {
		return nil, fmt.Errorf("failed to get available agents: %w", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		agent := &Agent{}
		err := rows.Scan(
			&agent.ID, &agent.Hostname, &agent.TargetCapability, &agent.Status,
			&agent.LastHeartbeatAt, &agent.CreatedAt, &agent.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		agents = append(agents, agent)
	}

	return agents, nil
}
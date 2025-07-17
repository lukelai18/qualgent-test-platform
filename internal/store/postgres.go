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
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
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
		INSERT INTO jobs (org_id, app_version_id, test_path, priority, target, status, idempotency_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	
	var id uuid.UUID
	var createdAt, updatedAt time.Time
	
	err := s.db.QueryRowContext(ctx, query,
		job.OrgID, job.AppVersionID, job.TestPath, job.Priority, job.Target, job.Status, job.IdempotencyKey,
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
		SELECT id, org_id, app_version_id, test_path, priority, target, status, job_group_id, idempotency_key, created_at, updated_at
		FROM jobs WHERE id = $1
	`
	
	job := &Job{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.OrgID, &job.AppVersionID, &job.TestPath, &job.Priority, &job.Target, &job.Status,
		&job.JobGroupID, &job.IdempotencyKey, &job.CreatedAt, &job.UpdatedAt,
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

func (s *PostgresStore) GetPendingJobs(ctx context.Context, limit int) ([]*Job, error) {
	query := `
		SELECT id, org_id, app_version_id, test_path, priority, target, status, job_group_id, idempotency_key, created_at, updated_at
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
			&job.JobGroupID, &job.IdempotencyKey, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}
	
	return jobs, nil
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
	// Convert UUID slice to string slice for PostgreSQL compatibility
	jobIDStrings := make([]string, len(jobIDs))
	for i, id := range jobIDs {
		jobIDStrings[i] = id.String()
	}
	
	query := `UPDATE jobs SET job_group_id = $1, status = 'SCHEDULED' WHERE id = ANY($2)`
	_, err := s.db.ExecContext(ctx, query, groupID, pq.Array(jobIDStrings))
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
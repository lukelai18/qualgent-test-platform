package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"qualgent-test-platform/internal/store"
)

type Scheduler struct {
	postgresStore *store.PostgresStore
	redisStore    *store.RedisStore
	lockKey       string
	instanceID    string
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

type JobGroup struct {
	AppVersionID string
	Target       string
	Jobs         []*store.Job
}

func NewScheduler(postgresStore *store.PostgresStore, redisStore *store.RedisStore, instanceID string) *Scheduler {
	return &Scheduler{
		postgresStore: postgresStore,
		redisStore:    redisStore,
		lockKey:       "scheduler:lock",
		instanceID:    instanceID,
		stopChan:      make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.run(ctx)
	log.Printf("Scheduler started with instance ID: %s", s.instanceID)
}

func (s *Scheduler) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.processBatch(ctx)
		}
	}
}

func (s *Scheduler) processBatch(ctx context.Context) {
	// Try to acquire distributed lock
	acquired, err := s.redisStore.AcquireLock(ctx, s.lockKey, 60*time.Second)
	if err != nil {
		log.Printf("Failed to acquire lock: %v", err)
		return
	}
	if !acquired {
		log.Println("Another scheduler instance is running, skipping this cycle")
		return
	}

	defer func() {
		if err := s.redisStore.ReleaseLock(ctx, s.lockKey); err != nil {
			log.Printf("Failed to release lock: %v", err)
		}
	}()

	// Process jobs in batches
	if err := s.processJobs(ctx); err != nil {
		log.Printf("Failed to process jobs: %v", err)
	}
}

func (s *Scheduler) processJobs(ctx context.Context) error {
	// Get pending jobs from database
	jobs, err := s.postgresStore.GetPendingJobs(ctx, 10)
	if err != nil {
		return fmt.Errorf("failed to get pending jobs: %w", err)
	}

	if len(jobs) == 0 {
		return nil
	}

	log.Printf("Processing %d pending jobs", len(jobs))

	// Group jobs by app_version_id and target
	jobGroups := s.groupJobs(jobs)

	// Create job groups and dispatch them
	for _, group := range jobGroups {
		if err := s.createAndDispatchGroup(ctx, group); err != nil {
			log.Printf("Failed to create and dispatch group: %v", err)
			continue
		}
	}

	return nil
}

func (s *Scheduler) groupJobs(jobs []*store.Job) []*JobGroup {
	groupMap := make(map[string]*JobGroup)

	for _, job := range jobs {
		key := fmt.Sprintf("%s:%s", job.AppVersionID, job.Target)
		
		if group, exists := groupMap[key]; exists {
			group.Jobs = append(group.Jobs, job)
		} else {
			groupMap[key] = &JobGroup{
				AppVersionID: job.AppVersionID,
				Target:       job.Target,
				Jobs:       []*store.Job{job},
			}
		}
	}

	// Convert map to slice
	var groups []*JobGroup
	for _, group := range groupMap {
		groups = append(groups, group)
	}

	return groups
}

func (s *Scheduler) createAndDispatchGroup(ctx context.Context, group *JobGroup) error {
	// Create job group in database
	jobGroup := &store.JobGroup{
		AppVersionID: group.AppVersionID,
		Target:       group.Target,
		Status:      "SCHEDULED",
	}

	if err := s.postgresStore.CreateJobGroup(ctx, jobGroup); err != nil {
		return fmt.Errorf("failed to create job group: %w", err)
	}

	// Update jobs to point to the group
	var jobIDs []uuid.UUID
	for _, job := range group.Jobs {
		jobIDs = append(jobIDs, job.ID)
	}

	if err := s.postgresStore.UpdateJobsToGroup(ctx, jobIDs, jobGroup.ID); err != nil {
		return fmt.Errorf("failed to update jobs to group: %w", err)
	}

	// Push to dispatch queue
	if err := s.redisStore.PushToDispatchQueue(ctx, group.Target, jobGroup.ID); err != nil {
		return fmt.Errorf("failed to push to dispatch queue: %w", err)
	}

	log.Printf("Created job group %s with %d jobs for target %s",		jobGroup.ID, len(group.Jobs), group.Target)

	return nil
}

// GetJobGroup retrieves a job group with all its jobs
func (s *Scheduler) GetJobGroup(ctx context.Context, groupID uuid.UUID) (*store.JobGroup, []*store.Job, error) {
	// This would need to be implemented in the store layer
	// For now, we'll return a placeholder
	return nil, nil, fmt.Errorf("not implemented")
} 
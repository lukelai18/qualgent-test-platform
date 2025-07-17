package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr string) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisStore{client: client}, nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}

// Queue operations
func (s *RedisStore) PushToIngestionQueue(ctx context.Context, jobID uuid.UUID) error {
	return s.client.LPush(ctx, "ingestion_queue", jobID.String()).Err()
}

func (s *RedisStore) PopFromIngestionQueue(ctx context.Context, timeout time.Duration) (uuid.UUID, error) {
	result, err := s.client.BRPop(ctx, timeout, "ingestion_queue").Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to pop from ingestion queue: %w", err)
	}

	if len(result) < 2 {
		return uuid.Nil, fmt.Errorf("unexpected result from BRPop")
	}

	jobID, err := uuid.Parse(result[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse job ID: %w", err)
	}

	return jobID, nil
}

func (s *RedisStore) PushToDispatchQueue(ctx context.Context, target string, groupID uuid.UUID) error {
	queueName := fmt.Sprintf("dispatch_queue:%s", target)
	return s.client.LPush(ctx, queueName, groupID.String()).Err()
}

func (s *RedisStore) PopFromDispatchQueue(ctx context.Context, target string, timeout time.Duration) (uuid.UUID, error) {
	queueName := fmt.Sprintf("dispatch_queue:%s", target)
	result, err := s.client.BRPop(ctx, timeout, queueName).Result()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to pop from dispatch queue: %w", err)
	}

	if len(result) < 2 {
		return uuid.Nil, fmt.Errorf("unexpected result from BRPop")
	}

	groupID, err := uuid.Parse(result[1])
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse group ID: %w", err)
	}

	return groupID, nil
}

// Distributed lock operations
func (s *RedisStore) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error) {
	result := s.client.SetNX(ctx, lockKey, "locked", ttl)
	return result.Val(), result.Err()
}

func (s *RedisStore) ReleaseLock(ctx context.Context, lockKey string) error {
	return s.client.Del(ctx, lockKey).Err()
}

// Heartbeat operations
func (s *RedisStore) UpdateAgentHeartbeat(ctx context.Context, agentID uuid.UUID, ttl time.Duration) error {
	key := fmt.Sprintf("agent:heartbeat:%s", agentID.String())
	return s.client.Set(ctx, key, "alive", ttl).Err()
}

func (s *RedisStore) IsAgentAlive(ctx context.Context, agentID uuid.UUID) (bool, error) {
	key := fmt.Sprintf("agent:heartbeat:%s", agentID.String())
	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check agent heartbeat: %w", err)
	}
	return exists > 0, nil
}

// Idempotency operations
func (s *RedisStore) CheckIdempotency(ctx context.Context, key string) (bool, error) {
	exists, err := s.client.Exists(ctx, fmt.Sprintf("idempotency:%s", key)).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check idempotency: %w", err)
	}
	return exists > 0, nil
}

func (s *RedisStore) SetIdempotency(ctx context.Context, key string, ttl time.Duration) error {
	return s.client.Set(ctx, fmt.Sprintf("idempotency:%s", key), "processed", ttl).Err()
}

// Cache operations
func (s *RedisStore) SetJobStatus(ctx context.Context, jobID uuid.UUID, status string, ttl time.Duration) error {
	key := fmt.Sprintf("job:status:%s", jobID.String())
	return s.client.Set(ctx, key, status, ttl).Err()
}

func (s *RedisStore) GetJobStatus(ctx context.Context, jobID uuid.UUID) (string, error) {
	key := fmt.Sprintf("job:status:%s", jobID.String())
	result, err := s.client.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("failed to get job status from cache: %w", err)
	}
	return result, nil
}

// Queue statistics
func (s *RedisStore) GetQueueLength(ctx context.Context, queueName string) (int64, error) {
	return s.client.LLen(ctx, queueName).Result()
}

func (s *RedisStore) GetIngestionQueueLength(ctx context.Context) (int64, error) {
	return s.GetQueueLength(ctx, "ingestion_queue")
}

func (s *RedisStore) GetDispatchQueueLength(ctx context.Context, target string) (int64, error) {
	queueName := fmt.Sprintf("dispatch_queue:%s", target)
	return s.GetQueueLength(ctx, queueName)
} 
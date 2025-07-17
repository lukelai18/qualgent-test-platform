package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"github.com/google/uuid"

	pb "qualgent-test-platform/api/proto"
	"qualgent-test-platform/internal/scheduler"
	"qualgent-test-platform/internal/server"
	"qualgent-test-platform/internal/store"
)

func main() {
	// Get configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "user")
	dbPassword := getEnv("DB_PASSWORD", "password")
	dbName := getEnv("DB_NAME", "qg_jobs")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	grpcPort := getEnv("GRPC_PORT", "8080")
	// Create database connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	// Initialize stores
	postgresStore, err := store.NewPostgresStore(connStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer postgresStore.Close()

	redisStore, err := store.NewRedisStore(redisAddr)
			if err != nil {
		log.Printf("Failed to connect to Redis: %v", err)
		os.Exit(1)
	}
	defer redisStore.Close()

	// Initialize scheduler
	instanceID := uuid.New().String()
	sched := scheduler.NewScheduler(postgresStore, redisStore, instanceID)

	// Initialize gRPC service
	jobService := server.NewJobService(postgresStore, redisStore)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterJobServiceServer(grpcServer, jobService)

	// Start scheduler
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sched.Start(ctx)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Printf("Job server listening on port %s", grpcPort)
	log.Printf("Scheduler instance ID: %s", instanceID)

	// Start server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	sched.Stop()
	grpcServer.GracefulStop()

	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

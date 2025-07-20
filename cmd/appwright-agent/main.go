package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"qualgent-test-platform/internal/agent"
)

func main() {
	var (
		serverAddr = flag.String("server", "localhost:8080", "gRPC server address")
		hostname   = flag.String("hostname", "", "Agent hostname (defaults to system hostname)")
	)
	flag.Parse()

	// Get hostname if not provided
	if *hostname == "" {
		hostnameFromEnv, err := os.Hostname()
		if err != nil {
			log.Fatalf("Failed to get hostname: %v", err)
		}
		*hostname = hostnameFromEnv
	}

	// Check required environment variables
	if os.Getenv("BROWSERSTACK_USERNAME") == "" || os.Getenv("BROWSERSTACK_ACCESS_KEY") == "" {
		log.Fatal("BROWSERSTACK_USERNAME and BROWSERSTACK_ACCESS_KEY environment variables are required")
	}

	// Create AppWright agent
	agent, err := agent.NewAppWrightAgent(*serverAddr, *hostname)
	if err != nil {
		log.Fatalf("Failed to create AppWright agent: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Start the agent
	log.Printf("Starting AppWright agent on %s", *hostname)
	if err := agent.Start(ctx); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	log.Println("AppWright agent stopped")
} 
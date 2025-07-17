#!/bin/bash

set -e

# ——————————————————————————————————————————————————
# QualGent Test Platform Integration Test
# ——————————————————————————————————————————————————

echo "🧪 Starting QualGent Test Platform Integration Test"
echo "=================================================="

# --- Configuration ---

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# --- Helper Functions ---

# Print a formatted status message.
# Arguments:
#   $1: The message to print.
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

# Print a formatted warning message.
# Arguments:
#   $1: The message to print.
print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Print a formatted error message.
# Arguments:
#   $1: The message to print.
print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if a command exists.
# Arguments:
#   $1: The command to check.
#   $2: An optional warning message if the command is not found.
check_command() {
    if ! command -v "$1" &> /dev/null; then
        if [ -n "$2" ]; then
            print_warning "$2"
            return 1
        else
            print_error "Command not found: $1"
            exit 1
        fi
    fi
    return 0
}

# --- Test Setup ---

# Check if required tools are installed.
check_dependencies() {
    echo "
🔎 Checking dependencies..."
    check_command "docker"
    check_command "docker-compose"
    check_command "go"
    check_command "protoc" "protoc not found, protobuf generation will be skipped." || true
    check_command "psql" "psql not found, skipping direct database initialization." || true
    print_status "Dependencies check passed"
}

# Start services using Docker Compose.
start_services() {
    echo "
🐳 Starting services with Docker Compose..."
    docker-compose up -d postgres redis
    
    echo "Waiting for services to be ready..."
    sleep 10
    
    if ! docker-compose ps | grep -q "Up"; then
        print_error "Services failed to start"
        docker-compose logs
        exit 1
    fi
    
    print_status "Services started successfully"
}

# Initialize the database schema.
init_database() {
    echo "
🗄️  Initializing database..."
    
    if ! command -v psql &> /dev/null; then
        print_warning "psql not found, skipping database initialization."
        return
    fi
    
    echo "Waiting for PostgreSQL to be ready..."
    for i in {1..30}; do
        if docker-compose exec -T postgres pg_isready -U user -d qg_jobs &> /dev/null; then
            print_status "PostgreSQL is ready."
            break
        fi
        echo " - waiting... ($i/30)"
        sleep 1
    done

    echo "Running database schema migration..."
    docker-compose exec -T postgres psql -U user -d qg_jobs -f /app/internal/store/schema.sql || \
        print_warning "Database schema migration failed (this might be expected if tables already exist)"
    
    print_status "Database initialized"
}

# Build the application binaries.
build_app() {
    echo "
🏗️  Building application..."
    
    if command -v protoc &> /dev/null; then
        echo "Generating protobuf code..."
        protoc --go_out=. \
            --go_opt=paths=source_relative \
            --go-grpc_out=. \
            --go-grpc_opt=paths=source_relative \
            api/proto/job_service.proto || print_warning "Protobuf generation failed"
    fi
    
    echo "Building job-server..."
    go build -o bin/job-server ./cmd/job-server
    
    echo "Building qgjob CLI..."
    go build -o bin/qgjob ./cmd/qgjob
    
    print_status "Application built successfully"
}

# Start the job-server in the background.
start_job_server() {
    echo "
🚀 Starting job-server..."
    sleep 5 # Give services a moment to fully initialize
    ./bin/job-server &
    JOB_SERVER_PID=$!
    
    echo "Waiting for job-server to start..."
    for i in {1..30}; do
        if netstat -tlnp 2>/dev/null | grep -q ":8080"; then
            print_status "Job-server started (PID: $JOB_SERVER_PID)"
            return
        fi
        sleep 1
    done
    
    print_error "Job-server failed to start"
    exit 1
}

# --- Test Execution ---

# Test basic CLI functionality.
test_cli() {
    echo "
🔬 Testing CLI functionality..."
    ./bin/qgjob --help &> /dev/null || { print_error "CLI --help command failed"; return 1; }
    ./bin/qgjob submit --help &> /dev/null || { print_error "CLI submit --help command failed"; return 1; }
    ./bin/qgjob status --help &> /dev/null || { print_error "CLI status --help command failed"; return 1; }
    print_status "CLI functionality tests passed"
}

# Test the job submission and status query flow.
test_job_submission() {
    echo "
🔬 Testing job submission and status..."
    
    echo "Submitting test job..."
    JOB_OUTPUT=$(./bin/qgjob submit \
        --org-id="test-org" \
        --app-version-id="test-version-123" \
        --test="tests/example.spec.js" \
        --target="emulator" \
        --priority=5 \
        --server="localhost:8080" 2>&1)
    
    if [[ $? -ne 0 ]]; then
        print_error "Job submission failed: $JOB_OUTPUT"
        return 1
    fi
    
    JOB_ID=$(echo "$JOB_OUTPUT" | grep "Job ID:" | awk '{print $3}')
    
    if [ -z "$JOB_ID" ]; then
        print_error "Failed to extract job ID from output: $JOB_OUTPUT"
        return 1
    fi
    
    echo "Submitted job with ID: $JOB_ID"
    
    echo "Testing job status query..."
    STATUS_OUTPUT=$(./bin/qgjob status --job-id="$JOB_ID" --server="localhost:8080" 2>&1)
    
    if [[ $? -ne 0 ]]; then
        print_error "Job status query failed: $STATUS_OUTPUT"
        return 1
    fi
    
    echo "Job status: $STATUS_OUTPUT"
    print_status "Job submission and status query tests passed"
}

# Test the JSON output functionality.
test_json_output() {
    echo "
🔬 Testing JSON output..."
    check_command "jq" "jq not found, skipping JSON validation."

    echo "Submitting another job for JSON testing..."
    JSON_OUTPUT=$(./bin/qgjob submit \
        --org-id="test-org-json" \
        --app-version-id="test-version-json" \
        --test="tests/json-test.spec.js" \
        --target="emulator" \
        --priority=3 \
        --server="localhost:8080" --json 2>&1)

    if [[ $? -ne 0 ]]; then
        print_error "JSON test job submission failed: $JSON_OUTPUT"
        return 1
    fi

    JSON_JOB_ID=$(echo "$JSON_OUTPUT" | jq -r '.job_id')

    if [ -z "$JSON_JOB_ID" ]; then
        print_error "Failed to extract job ID from JSON output: $JSON_OUTPUT"
        return 1
    fi

    echo "Testing JSON status output..."
    JSON_STATUS=$(./bin/qgjob status --job-id="$JSON_JOB_ID" --json --server="localhost:8080" 2>&1)

    if [[ $? -ne 0 ]]; then
        print_error "JSON status query failed: $JSON_STATUS"
        return 1
    fi

    if ! echo "$JSON_STATUS" | jq . &> /dev/null; then
        print_error "Invalid JSON output: $JSON_STATUS"
        return 1
    fi
    
    echo "JSON output: $JSON_STATUS"
    print_status "JSON output test passed"
}

# --- Teardown ---

# Clean up all created resources.
cleanup() {
    echo "
🧹 Cleaning up..."
    
    if [ -n "$JOB_SERVER_PID" ]; then
        echo "Stopping job-server (PID: $JOB_SERVER_PID)..."
        kill "$JOB_SERVER_PID" 2>/dev/null || true
    fi
    
    echo "Stopping services..."
    docker-compose down -v
    echo "Removing postgres_data volume..."
    docker volume rm qualgent-test-platform_postgres_data || true
    
    echo "Cleaning build artifacts..."
    rm -rf bin
    rm -f job-server.log
    
    print_status "Cleanup completed"
}

# --- Main Execution ---

# Main function to run the entire test suite.
main() {
    trap cleanup EXIT
    cleanup
    
    check_dependencies
    start_services
    init_database
    build_app
    start_job_server
    
    echo "
🏁 Starting tests..."
    test_cli
    test_job_submission
    test_json_output
    
    echo "
=================================================="
    echo "🎉 All tests completed successfully!"
    echo "=================================================="
}

# Run main function
main "$@"
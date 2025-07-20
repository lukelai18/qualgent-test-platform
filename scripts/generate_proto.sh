#!/bin/bash

set -e

echo "Generating protobuf code..."

# Create the proto output directory
mkdir -p api/proto/gen

# Generate Go code from proto files
protoc --go_out=. \
       --go_opt=paths=source_relative \
       --go-grpc_out=. \
       --go-grpc_opt=paths=source_relative \
       api/proto/job_service.proto

echo "Protobuf code generated successfully!" 
#!/bin/bash

set -e # Exit immediately if a command exits with a non-zero status.

JOB_ID=$1
TIMEOUT_SECONDS=600 # 10 minutes
POLL_INTERVAL_SECONDS=15 # Poll every 15 seconds

if [ -z "$JOB_ID" ]; then
  echo "Error: Job ID not provided."
  exit 1
fi

echo "Polling status for job: $JOB_ID"
end_time=$(( $(date +%s) + TIMEOUT_SECONDS ))

while [ $(date +%s) -lt $end_time ]; do
  # Use the --json flag to get machine-readable output
  status_json=$(./qgjob status --job-id "$JOB_ID" --json)
  status=$(echo "$status_json" | jq -r '.status')

  echo "Current job status: $status"

  if [ "$status" == "COMPLETED" ]; then
    echo "Job completed successfully."
    exit 0
  elif [ "$status" == "FAILED" ]; then
    echo "::error::Job failed."
    # You can further enhance this to fetch and print logs
    exit 1
  fi

  sleep $POLL_INTERVAL_SECONDS
done

echo "::error::Polling timed out after $TIMEOUT_SECONDS seconds."
exit 1

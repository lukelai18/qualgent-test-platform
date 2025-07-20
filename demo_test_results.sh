#!/bin/bash

set -e

echo "🧪 Demo: Test Results Viewing"
echo "=============================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Submit a test job
echo -e "\n${YELLOW}📤 Submitting a test job...${NC}"
JOB_OUTPUT=$(./qgjob submit \
  --org-id=demo-results \
  --app-version-id=bs://demo-app-results \
  --test=tests/login.spec.js \
  --target=browserstack \
  --priority=5)

JOB_ID=$(echo "$JOB_OUTPUT" | grep "Job ID:" | awk '{print $3}')
echo -e "${GREEN}✅ Job submitted: $JOB_ID${NC}"

# Wait for processing
echo -e "\n${YELLOW}⏳ Waiting for job processing...${NC}"
sleep 3

# Show basic status
echo -e "\n${BLUE}📊 Basic Job Status:${NC}"
./qgjob status --job-id="$JOB_ID"

# Show detailed JSON output
echo -e "\n${BLUE}📋 Detailed JSON Output:${NC}"
./qgjob status --job-id="$JOB_ID" --json

# Simulate test completion (in a real scenario, this would come from the agent)
echo -e "\n${YELLOW}🔧 Simulating test completion...${NC}"
docker-compose exec postgres psql -U user -d qg_jobs -c "
UPDATE jobs SET 
  status = 'COMPLETED',
  session_id = 'demo-session-123',
  logs_url = 'https://app-automate.browserstack.com/dashboard/v2/builds/demo-session-123/sessions/demo-session-123',
  video_url = 'https://app-automate.browserstack.com/dashboard/v2/builds/demo-session-123/sessions/demo-session-123/video',
  test_duration = 45,
  completed_at = NOW()
WHERE id = '$JOB_ID';
"

# Show completed test results
echo -e "\n${BLUE}✅ Completed Test Results:${NC}"
./qgjob status --job-id="$JOB_ID"

echo -e "\n${BLUE}📋 Completed Test Results (JSON):${NC}"
./qgjob status --job-id="$JOB_ID" --json

# Show how to access different result types
echo -e "\n${GREEN}🎯 How to View Test Results:${NC}"
echo -e "${YELLOW}1. Basic Status:${NC} ./qgjob status --job-id=<job-id>"
echo -e "${YELLOW}2. JSON Output:${NC} ./qgjob status --job-id=<job-id> --json"
echo -e "${YELLOW}3. Database Query:${NC} docker-compose exec postgres psql -U user -d qg_jobs -c \"SELECT * FROM jobs WHERE id = '<job-id>';\""
echo -e "${YELLOW}4. Logs URL:${NC} Open the logs_url in browser to view test logs"
echo -e "${YELLOW}5. Video URL:${NC} Open the video_url in browser to view test recording"

echo -e "\n${GREEN}🎉 Demo completed!${NC}" 
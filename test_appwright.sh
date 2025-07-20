#!/bin/bash

set -e

echo "🧪 Testing QualGent AppWright Integration"
echo "=========================================="

# Colors for output
RED='
[0;31m'
GREEN='
[0;32m'
YELLOW='
[1;33m'
NC='
[0m' # No Color

# Check if services are running
echo -e "${YELLOW}📋 Checking service status...${NC}"
if ! docker-compose ps | grep -q "Up"; then
    echo -e "${RED}❌ Services are not running. Please start them with: docker-compose up -d${NC}"
    exit 1
end

echo -e "${GREEN}✅ All services are running${NC}"

# Wait for the job-server to be ready
echo -e "\n${YELLOW}⏳ Waiting for job-server to initialize...${NC}"
while ! docker-compose logs job-server | grep -q "Job server listening on port"; do
  sleep 1
done
echo -e "${GREEN}✅ Job-server is ready!${NC}"

# Submit a test job
echo -e "\n${YELLOW}📤 Submitting AppWright test job...${NC}"
OUTPUT=$(./qgjob submit --org-id=qualgent-demo --app-version-id=bs://demo-app-789 --test=tests/login.spec.js --target=browserstack --priority=5)

echo -e "${GREEN}✅ Job submitted successfully!${NC}"
echo "$OUTPUT"

# Extract job ID from output
JOB_ID=$(echo "$OUTPUT" | grep "Job ID:" | awk '{print $3}')
echo "Extracted Job ID: $JOB_ID"

# Wait a moment for processing
echo -e "\n${YELLOW}⏳ Waiting for job processing...${NC}"
sleep 3

# Check job status
echo -e "\n${YELLOW}📊 Checking job status...${NC}"
./qgjob status --job-id="$JOB_ID"

# Check JSON output
echo -e "\n${YELLOW}📋 Job status (JSON format):${NC}"
./qgjob status --job-id="$JOB_ID" --json

# Show service logs
echo -e "\n${YELLOW}📋 Recent job-server logs:${NC}"
docker-compose logs job-server --tail=5

echo -e "\n${YELLOW}📋 Recent appwright-agent logs:${NC}"
docker-compose logs appwright-agent --tail=5

echo -e "\n${GREEN}🎉 AppWright integration test completed!${NC}"
echo -e "${YELLOW}💡 Note: This is a demo with test credentials.${NC}"
echo -e "${YELLOW}💡 For real BrowserStack integration, set BROWSERSTACK_USERNAME and BROWSERSTACK_ACCESS_KEY${NC}"

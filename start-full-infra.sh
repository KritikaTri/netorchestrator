#!/bin/bash

# Start Full Infrastructure Mode
# Starts only PostgreSQL and Backend API (no Redis, NATS, etc.)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_CYAN='\033[0;36m'
COLOR_NC='\033[0m'

echo -e "${COLOR_CYAN}=========================================="
echo "NetOrchestrator - Full Infrastructure Mode"
echo "==========================================${COLOR_NC}"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${COLOR_RED}✗ .env file not found${COLOR_NC}"
    echo "Please copy ENV_EXAMPLE to .env and configure it"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' .env | xargs)

# Start PostgreSQL only
echo -e "${COLOR_CYAN}1. Starting PostgreSQL...${COLOR_NC}"
docker compose -f docker-compose.monitoring.yml up -d postgres

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
sleep 5

# Check PostgreSQL health
for i in {1..30}; do
    if PGPASSWORD="${DATABASE_PASSWORD}" psql -h "${DATABASE_HOST}" -p "${DATABASE_PORT}" -U "${DATABASE_USER}" -d "${DATABASE_DBNAME}" -c "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${COLOR_GREEN}✓ PostgreSQL is ready${COLOR_NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${COLOR_RED}✗ PostgreSQL failed to start${COLOR_NC}"
        exit 1
    fi
    sleep 1
done

echo ""
echo -e "${COLOR_CYAN}2. Building API Gateway...${COLOR_NC}"
go build -o bin/api-gateway ./cmd/api-gateway
if [ $? -ne 0 ]; then
    echo -e "${COLOR_RED}✗ Failed to build API Gateway${COLOR_NC}"
    exit 1
fi
echo -e "${COLOR_GREEN}✓ API Gateway built successfully${COLOR_NC}"
echo ""

echo -e "${COLOR_CYAN}3. Starting API Gateway (migrations will run automatically)...${COLOR_NC}"
echo -e "${COLOR_YELLOW}Note: This will start in the foreground. Press Ctrl+C to stop.${COLOR_NC}"
echo ""

# Start API Gateway (will run migrations automatically)
export $(grep -v '^#' .env | xargs)
./bin/api-gateway


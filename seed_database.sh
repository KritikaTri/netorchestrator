#!/bin/bash

# Database Seeding Script
# Resets and repopulates the database with sample data
# Usage: ./seed_database.sh [--reset]

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
echo "NetOrchestrator - Database Seeding"
echo "==========================================${COLOR_NC}"
echo ""

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${COLOR_RED}✗ .env file not found${COLOR_NC}"
    echo "Please copy ENV_EXAMPLE to .env and configure it"
    exit 1
fi

# Check if config.yaml exists
if [ ! -f config.yaml ]; then
    echo -e "${COLOR_RED}✗ config.yaml not found${COLOR_NC}"
    exit 1
fi

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo -e "${COLOR_RED}✗ Go is not installed${COLOR_NC}"
    exit 1
fi

# Build the seed tool
echo -e "${COLOR_CYAN}1. Building seed tool...${COLOR_NC}"
go build -o bin/seed ./cmd/seed
if [ $? -ne 0 ]; then
    echo -e "${COLOR_RED}✗ Failed to build seed tool${COLOR_NC}"
    exit 1
fi
echo -e "${COLOR_GREEN}✓ Seed tool built successfully${COLOR_NC}"
echo ""

# Check for reset flag
RESET_FLAG=""
if [ "$1" == "--reset" ]; then
    RESET_FLAG="--reset"
    echo -e "${COLOR_YELLOW}⚠ Reset flag set - will delete existing seed data${COLOR_NC}"
    echo ""
fi

# Run the seed tool
echo -e "${COLOR_CYAN}2. Running database migrations and seeding...${COLOR_NC}"
./bin/seed $RESET_FLAG

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${COLOR_GREEN}✅ Database seeding completed successfully!${COLOR_NC}"
else
    echo ""
    echo -e "${COLOR_RED}✗ Database seeding failed${COLOR_NC}"
    exit 1
fi


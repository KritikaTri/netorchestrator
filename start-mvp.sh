#!/bin/bash

# NetOrchestrator MVP Startup Script

set -e

echo "🚀 Starting NetOrchestrator MVP..."
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

echo "${YELLOW}Step 1: Starting infrastructure services...${NC}"
docker-compose -f docker-compose-simple.yml up -d postgres redis

echo ""
echo "${YELLOW}Step 2: Waiting for services to be healthy...${NC}"
sleep 5

# Wait for postgres
echo "Waiting for PostgreSQL..."
until docker exec netorch-postgres pg_isready -U netorchestrator > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo " ✅"

# Wait for Redis
echo "Waiting for Redis..."
until docker exec netorch-redis redis-cli ping > /dev/null 2>&1; do
    echo -n "."
    sleep 1
done
echo " ✅"

echo ""
echo "${YELLOW}Step 3: Building API Gateway...${NC}"
go build -o bin/api-gateway ./cmd/api-gateway/
echo " ✅"

echo ""
echo "${GREEN}════════════════════════════════════════════════════════════${NC}"
echo "${GREEN}✅ Infrastructure is ready!${NC}"
echo "${GREEN}════════════════════════════════════════════════════════════${NC}"
echo ""
echo "📊 Services running:"
echo "   - PostgreSQL: localhost:5432"
echo "   - Redis: localhost:6379"
echo "   - pgAdmin: http://localhost:5050 (admin@netorchestrator.com / admin)"
echo ""
echo "🚀 To start the API Gateway, run:"
echo "   ./bin/api-gateway"
echo ""
echo "📖 API Documentation:"
echo "   - Swagger UI: http://localhost:8080/swagger/index.html"
echo "   - Endpoints: See API_ENDPOINTS.md"
echo ""
echo "🔐 Test credentials:"
echo "   - Username: admin"
echo "   - Password: admin123"
echo ""
echo "📝 Quick test:"
echo "   curl http://localhost:8080/health"
echo ""


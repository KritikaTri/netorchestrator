#!/bin/bash

# NetOrchestrator Observability Stack Startup Script

set -e

echo "🚀 Starting NetOrchestrator with Observability Stack..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Start Prometheus + Grafana
echo -e "${BLUE}📊 Starting Prometheus + Grafana...${NC}"
docker-compose -f docker-compose.observability.yml up -d

# Wait for services to be ready
echo -e "${YELLOW}⏳ Waiting for services to be ready...${NC}"
sleep 5

# Check if Prometheus is up
echo -e "${BLUE}🔍 Checking Prometheus...${NC}"
if curl -s http://localhost:9090/-/healthy > /dev/null; then
    echo -e "${GREEN}✓ Prometheus is healthy (http://localhost:9090)${NC}"
else
    echo -e "${YELLOW}⚠ Prometheus not responding yet...${NC}"
fi

# Check if Grafana is up
echo -e "${BLUE}🔍 Checking Grafana...${NC}"
if curl -s http://localhost:3000/api/health > /dev/null; then
    echo -e "${GREEN}✓ Grafana is healthy (http://localhost:3000)${NC}"
    echo -e "${YELLOW}  Username: admin${NC}"
    echo -e "${YELLOW}  Password: netorch-admin${NC}"
else
    echo -e "${YELLOW}⚠ Grafana not responding yet...${NC}"
fi

echo ""
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Observability Stack Started Successfully! 🎉${NC}"
echo -e "${GREEN}════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BLUE}Access Points:${NC}"
echo -e "  • Grafana Dashboard: ${GREEN}http://localhost:3000${NC}"
echo -e "  • Prometheus: ${GREEN}http://localhost:9090${NC}"
echo -e "  • API /metrics: ${GREEN}http://localhost:8080/metrics${NC} (when API is running)"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo -e "  1. Start the NetOrchestrator API:"
echo -e "     ${YELLOW}./api-gateway${NC}"
echo -e "  2. Open Grafana at http://localhost:3000"
echo -e "  3. Navigate to 'NetOrchestrator - Infrastructure & Performance' dashboard"
echo -e "  4. Create networks, nodes, and policies to see metrics appear"
echo ""
echo -e "${BLUE}Useful Commands:${NC}"
echo -e "  • View logs: ${YELLOW}docker-compose -f docker-compose.observability.yml logs -f${NC}"
echo -e "  • Stop stack: ${YELLOW}docker-compose -f docker-compose.observability.yml down${NC}"
echo -e "  • Check status: ${YELLOW}docker-compose -f docker-compose.observability.yml ps${NC}"
echo ""

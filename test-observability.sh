#!/bin/bash

# NetOrchestrator Observability Testing Script
# This script demonstrates how to manually test the logging and monitoring system

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   NetOrchestrator Observability System Testing Guide      ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Function to check service health
check_service() {
    local service=$1
    local url=$2
    echo -e "${YELLOW}Checking $service...${NC}"
    if curl -s -f "$url" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ $service is running${NC}"
        return 0
    else
        echo -e "${RED}✗ $service is NOT running${NC}"
        return 1
    fi
}

echo -e "${BLUE}═══ Step 1: Check All Services ═══${NC}"
echo ""

# Check API Gateway
check_service "API Gateway" "http://localhost:8080/health"
API_UP=$?

# Check Prometheus
check_service "Prometheus" "http://localhost:9091/-/healthy"
PROM_UP=$?

# Check Grafana
check_service "Grafana" "http://localhost:3002/api/health"
GRAFANA_UP=$?

echo ""
echo -e "${BLUE}═══ Step 2: Test /metrics Endpoint ═══${NC}"
echo ""

if [ $API_UP -eq 0 ]; then
    echo -e "${YELLOW}Fetching metrics from API...${NC}"
    METRICS=$(curl -s http://localhost:8080/metrics)
    
    # Count netorch metrics
    METRIC_COUNT=$(echo "$METRICS" | grep "^netorch_" | grep -v "^#" | wc -l | tr -d ' ')
    echo -e "${GREEN}Found $METRIC_COUNT NetOrchestrator metrics${NC}"
    
    echo ""
    echo -e "${YELLOW}Sample metrics:${NC}"
    echo "$METRICS" | grep "netorch_networks_total" | head -3
    echo "$METRICS" | grep "netorch_http_requests_total" | head -3
    
    echo ""
    echo -e "${YELLOW}Available metric types:${NC}"
    echo "$METRICS" | grep "^# HELP netorch_" | sed 's/# HELP /  • /' | head -10
else
    echo -e "${RED}Cannot test /metrics - API is not running${NC}"
fi

echo ""
echo -e "${BLUE}═══ Step 3: Test Prometheus Scraping ═══${NC}"
echo ""

if [ $PROM_UP -eq 0 ]; then
    echo -e "${YELLOW}Checking Prometheus targets...${NC}"
    TARGETS=$(curl -s http://localhost:9091/api/v1/targets | jq -r '.data.activeTargets[] | "\(.labels.job): \(.health)"')
    echo "$TARGETS"
    
    echo ""
    echo -e "${YELLOW}Querying NetOrchestrator metrics from Prometheus...${NC}"
    
    # Query networks
    echo -e "${BLUE}Networks:${NC}"
    curl -s 'http://localhost:9091/api/v1/query?query=netorch_networks_total' | \
        jq -r '.data.result[] | "  \(.metric.status): \(.value[1])"'
    
    # Query nodes
    echo -e "${BLUE}Nodes:${NC}"
    curl -s 'http://localhost:9091/api/v1/query?query=netorch_nodes_total' | \
        jq -r '.data.result[] | "  \(.metric.node_type) (\(.metric.status)): \(.value[1])"'
    
    # Query HTTP requests
    echo -e "${BLUE}HTTP Requests:${NC}"
    curl -s 'http://localhost:9091/api/v1/query?query=netorch_http_requests_total' | \
        jq -r '.data.result[] | "  \(.metric.method) \(.metric.endpoint) [\(.metric.status_code)]: \(.value[1])"'
    
    # Query containers
    echo -e "${BLUE}Containers:${NC}"
    CONTAINERS=$(curl -s 'http://localhost:9091/api/v1/query?query=netorch_containers_active' | jq -r '.data.result')
    if [ "$CONTAINERS" == "[]" ]; then
        echo "  (No active containers yet - provision nodes to see data)"
    else
        echo "$CONTAINERS" | jq -r '.[] | "  \(.metric.node_type): \(.value[1])"'
    fi
else
    echo -e "${RED}Cannot test Prometheus - service is not running${NC}"
fi

echo ""
echo -e "${BLUE}═══ Step 4: Generate Test Traffic ═══${NC}"
echo ""

if [ $API_UP -eq 0 ]; then
    echo -e "${YELLOW}Generating API requests to populate metrics...${NC}"
    
    # Make several health checks
    for i in {1..5}; do
        curl -s http://localhost:8080/health > /dev/null
        echo -e "${GREEN}✓${NC} Health check $i"
    done
    
    # Try to list networks (may require auth)
    echo ""
    echo -e "${YELLOW}Attempting to list networks...${NC}"
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/v1/networks)
    echo "Response code: $RESPONSE"
    
    # Wait for metrics to update
    echo ""
    echo -e "${YELLOW}Waiting 2 seconds for metrics to update...${NC}"
    sleep 2
    
    # Check updated metrics
    echo ""
    echo -e "${YELLOW}Updated HTTP request metrics:${NC}"
    curl -s http://localhost:8080/metrics | grep "netorch_http_requests_total" | grep -v "^#"
else
    echo -e "${RED}Cannot generate traffic - API is not running${NC}"
fi

echo ""
echo -e "${BLUE}═══ Step 5: Check Application Logs ═══${NC}"
echo ""

if [ -f "api-gateway.log" ]; then
    echo -e "${YELLOW}Recent API Gateway logs:${NC}"
    echo "─────────────────────────────────────────────────────────────"
    tail -20 api-gateway.log
    echo "─────────────────────────────────────────────────────────────"
    
    echo ""
    echo -e "${YELLOW}Log statistics:${NC}"
    echo "Total lines: $(wc -l < api-gateway.log)"
    echo "INFO messages: $(grep -c '"level":"info"' api-gateway.log 2>/dev/null || echo 0)"
    echo "ERROR messages: $(grep -c '"level":"error"' api-gateway.log 2>/dev/null || echo 0)"
    echo "FATAL messages: $(grep -c '"level":"fatal"' api-gateway.log 2>/dev/null || echo 0)"
else
    echo -e "${YELLOW}No api-gateway.log file found${NC}"
fi

echo ""
echo -e "${BLUE}═══ Step 6: Access Points Summary ═══${NC}"
echo ""

echo -e "${GREEN}Direct Access URLs:${NC}"
echo "  • API Health:        http://localhost:8080/health"
echo "  • API Metrics:       http://localhost:8080/metrics"
echo "  • Prometheus UI:     http://localhost:9091"
echo "  • Prometheus Targets: http://localhost:9091/targets"
echo "  • Grafana:           http://localhost:3002 (admin/netorch-admin)"
echo "  • Frontend:          http://localhost:3001"

echo ""
echo -e "${GREEN}Useful Prometheus Queries:${NC}"
echo "  • All networks:           netorch_networks_total"
echo "  • Active containers:      netorch_containers_active"
echo "  • HTTP request rate:      rate(netorch_http_requests_total[5m])"
echo "  • Request latency (p95):  histogram_quantile(0.95, rate(netorch_http_request_duration_seconds_bucket[5m]))"
echo "  • Provisioning ops:       netorch_provision_operations_total"
echo "  • Policy enforcement:     netorch_policy_enforcement_total"

echo ""
echo -e "${GREEN}Grafana Dashboard:${NC}"
echo "  1. Open http://localhost:3002"
echo "  2. Login: admin / netorch-admin"
echo "  3. Navigate to: Dashboards → NetOrchestrator - Infrastructure & Performance"
echo "  4. Dashboard auto-refreshes every 10 seconds"

echo ""
echo -e "${BLUE}═══ System Status Summary ═══${NC}"
echo ""

TOTAL_SERVICES=3
UP_SERVICES=$((3 - API_UP - PROM_UP - GRAFANA_UP))

if [ $UP_SERVICES -eq $TOTAL_SERVICES ]; then
    echo -e "${GREEN}✓ All services are operational ($UP_SERVICES/$TOTAL_SERVICES)${NC}"
elif [ $UP_SERVICES -gt 0 ]; then
    echo -e "${YELLOW}⚠ Some services are down ($UP_SERVICES/$TOTAL_SERVICES operational)${NC}"
else
    echo -e "${RED}✗ All services are down${NC}"
fi

echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Testing complete! Check the output above for any issues.${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"

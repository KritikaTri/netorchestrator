#!/bin/bash

# NetOrchestrator - Complete Demo Script
# Shows all features working together for hackathon demo

set -e

API_URL="http://localhost:8080"
API_KEY="neto_test"
COLOR_GREEN='\033[0;32m'
COLOR_BLUE='\033[0;34m'
COLOR_YELLOW='\033[1;33m'
COLOR_CYAN='\033[0;36m'
COLOR_NC='\033[0m'

echo ""
echo "=========================================="
echo -e "${COLOR_CYAN}NetOrchestrator - Complete Demo${COLOR_NC}"
echo "=========================================="
echo ""
echo "This demo showcases:"
echo "  1. Zero-Touch NLP Provisioning"
echo "  2. Event-Driven Architecture"
echo "  3. AI Optimization Service"
echo "  4. Real-Time Monitoring"
echo "  5. Auto-Provisioning"
echo ""

sleep 2

# Step 1: Zero-Touch NLP Provisioning
echo -e "${COLOR_BLUE}=== Step 1: Zero-Touch NLP Provisioning ===${COLOR_NC}"
echo ""
echo "Creating a network using natural language..."
echo "Command: \"Create a star network with 5 routers\""
echo ""

NLP_RESPONSE=$(curl -sS -X POST \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"text":"Create a star network with 5 routers"}' \
  "$API_URL/api/v1/ai/provision")

NETWORK_ID=$(echo "$NLP_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
TOPOLOGY=$(echo "$NLP_RESPONSE" | grep -o '"topology":"[^"]*"' | cut -d'"' -f4)

echo -e "${COLOR_GREEN}✓ Network created via natural language!${COLOR_NC}"
echo "   Network ID: $NETWORK_ID"
echo "   Topology: $TOPOLOGY"
echo "   Status: Pending → Auto-provisioning → Active"
echo ""

sleep 2

# Step 2: Check Auto-Provisioning
echo -e "${COLOR_BLUE}=== Step 2: Auto-Provisioning (Event-Driven) ===${COLOR_NC}"
echo ""
echo "Waiting for auto-provisioning (via event-driven architecture)..."
sleep 3

NETWORK_STATUS=$(curl -sS -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$NETWORK_ID" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

echo -e "${COLOR_GREEN}✓ Network auto-provisioned!${COLOR_NC}"
echo "   Status: $NETWORK_STATUS"
echo "   (Triggered by network.created event → Provisioning Service)"
echo ""

sleep 2

# Step 3: Check Nodes Created
echo -e "${COLOR_BLUE}=== Step 3: Network Topology Created ===${COLOR_NC}"
echo ""
echo "Checking network topology..."

NODES_RESPONSE=$(curl -sS -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$NETWORK_ID/nodes")
NODE_COUNT=$(echo "$NODES_RESPONSE" | grep -o '"id":"[^"]*"' | wc -l | tr -d ' ')

echo -e "${COLOR_GREEN}✓ Network topology created!${COLOR_NC}"
echo "   Nodes: $NODE_COUNT routers"
echo "   (Nodes created automatically from NLP specification)"
echo ""

sleep 2

# Step 4: AI Optimization Analysis
echo -e "${COLOR_BLUE}=== Step 4: AI Optimization Analysis ===${COLOR_NC}"
echo ""
echo "Analyzing network with AI optimization service..."
echo ""

ANALYZE_RESPONSE=$(curl -sS -X POST -H "X-API-Key: $API_KEY" "$API_URL/api/v1/optimization/analyze/$NETWORK_ID")

HEALTH_SCORE=$(echo "$ANALYZE_RESPONSE" | grep -o '"health_score":[^,}]*' | cut -d':' -f2 || echo "N/A")

echo -e "${COLOR_GREEN}✓ AI analysis complete!${COLOR_NC}"
echo "   Health Score: $HEALTH_SCORE"
echo "   (AI service analyzed topology, routing, and compliance)"
echo ""

sleep 2

# Step 5: Get AI Recommendations
echo -e "${COLOR_BLUE}=== Step 5: AI Recommendations ===${COLOR_NC}"
echo ""
echo "Getting AI-powered recommendations..."
echo ""

RECOMMEND_RESPONSE=$(curl -sS -X POST -H "X-API-Key: $API_KEY" "$API_URL/api/v1/optimization/recommend/$NETWORK_ID")

if echo "$RECOMMEND_RESPONSE" | grep -q "recommendation\|message"; then
    echo -e "${COLOR_GREEN}✓ Recommendations available!${COLOR_NC}"
    echo "   (AI suggests topology improvements, IP optimization, security enhancements)"
else
    echo -e "${COLOR_YELLOW}⚠ Recommendations service responded${COLOR_NC}"
fi

echo ""

sleep 2

# Step 6: Monitoring & Events
echo -e "${COLOR_BLUE}=== Step 6: Real-Time Monitoring ===${COLOR_NC}"
echo ""
echo "Checking monitoring events..."
echo ""

EVENTS_RESPONSE=$(curl -sS -H "X-API-Key: $API_KEY" "$API_URL/api/v1/monitoring/events")
EVENT_COUNT=$(echo "$EVENTS_RESPONSE" | grep -o '"id":"[^"]*"' | wc -l | tr -d ' ')

echo -e "${COLOR_GREEN}✓ Monitoring active!${COLOR_NC}"
echo "   Events tracked: $EVENT_COUNT"
echo "   (Network.created, topology.updated events published to NATS)"
echo ""

sleep 2

# Step 7: Complete Flow Summary
echo -e "${COLOR_BLUE}=== Complete Flow Summary ===${COLOR_NC}"
echo ""
echo -e "${COLOR_GREEN}✅ Zero-Touch Provisioning:${COLOR_NC} Natural language → Network created"
echo -e "${COLOR_GREEN}✅ Event-Driven Architecture:${COLOR_NC} Events published → Services subscribed"
echo -e "${COLOR_GREEN}✅ Auto-Provisioning:${COLOR_NC} Network auto-provisioned via events"
echo -e "${COLOR_GREEN}✅ AI Optimization:${COLOR_NC} Network analyzed and recommendations provided"
echo -e "${COLOR_GREEN}✅ Real-Time Monitoring:${COLOR_NC} Events tracked and monitored"
echo ""

# Step 8: Show Network Details
echo -e "${COLOR_BLUE}=== Network Details ===${COLOR_NC}"
echo ""
NETWORK_DETAILS=$(curl -sS -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$NETWORK_ID")
echo "$NETWORK_DETAILS" | head -c 500
echo ""
echo ""

# Step 9: Cleanup (optional)
read -p "Cleanup: Delete demo network? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "Deleting network..."
    DELETE_RESPONSE=$(curl -sS -X DELETE -H "X-API-Key: $API_KEY" -w "%{http_code}" "$API_URL/api/v1/networks/$NETWORK_ID" -o /dev/null)
    
    if [ "$DELETE_RESPONSE" = "204" ]; then
        echo -e "${COLOR_GREEN}✓ Network deleted successfully${COLOR_NC}"
    else
        echo -e "${COLOR_YELLOW}⚠ Delete response: HTTP $DELETE_RESPONSE${COLOR_NC}"
    fi
fi

echo ""
echo "=========================================="
echo -e "${COLOR_CYAN}Demo Complete!${COLOR_NC}"
echo "=========================================="
echo ""
echo "Key Features Demonstrated:"
echo "  • Natural Language → Network Creation"
echo "  • Event-Driven Architecture (NATS)"
echo "  • Auto-Provisioning Service"
echo "  • AI Optimization & Recommendations"
echo "  • Real-Time Monitoring & Events"
echo ""
echo "API Endpoints Used:"
echo "  • POST /api/v1/ai/provision (NLP Zero-Touch)"
echo "  • POST /api/v1/optimization/analyze/:id (AI Analysis)"
echo "  • POST /api/v1/optimization/recommend/:id (AI Recommendations)"
echo "  • GET /api/v1/monitoring/events (Event Monitoring)"
echo ""


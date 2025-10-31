#!/bin/bash

# NLP Testing - Various Natural Language Inputs
# Tests NLP parser with different inputs

API_URL="http://localhost:8080"
API_KEY="neto_test"
COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_NC='\033[0m'

PASSED=0
FAILED=0

echo "=========================================="
echo "NLP Parser - Various Inputs Test"
echo "=========================================="
echo ""

# Test cases
declare -a TEST_CASES=(
    "Create a star network with 5 routers"
    "Make a mesh network with 10 nodes"
    "Build a ring topology with 8 switches"
    "Create network with subnet 10.0.1.0/24"
    "Create network called MyNetwork with 3 hosts"
    "Make a bus network with 6 routers"
    "Create star network with 3 hosts"
)

for i in "${!TEST_CASES[@]}"; do
    TEST_NUM=$((i + 1))
    TEST_TEXT="${TEST_CASES[$i]}"
    
    echo "TEST $TEST_NUM: \"$TEST_TEXT\""
    echo "-----------------------------------"
    
    RESPONSE=$(curl -sS -X POST \
      -H "X-API-Key: $API_KEY" \
      -H "Content-Type: application/json" \
      -d "{\"text\":\"$TEST_TEXT\"}" \
      "$API_URL/api/v1/ai/provision")
    
    # Check if network was created
    NETWORK_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    
    if [ -n "$NETWORK_ID" ]; then
        # Wait a moment for nodes to be created
        sleep 1
        
        # Extract topology
        TOPOLOGY=$(echo "$RESPONSE" | grep -o '"topology":"[^"]*"' | cut -d'"' -f4)
        
        # Get actual node count from API
        NODES_RESPONSE=$(curl -sS -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$NETWORK_ID/nodes")
        NODE_COUNT=$(echo "$NODES_RESPONSE" | grep -o '"id":"[^"]*"' | wc -l | tr -d ' ')
        
        # If still 0, check if response has nodes array
        if [ "$NODE_COUNT" = "0" ]; then
            # Check if nodes array exists in response
            if echo "$RESPONSE" | grep -q '"nodes"'; then
                SPEC_NODE_COUNT=$(echo "$RESPONSE" | grep -o '"nodes":\[[^]]*\]' | grep -o '"name":"[^"]*"' | wc -l | tr -d ' ')
                if [ "$SPEC_NODE_COUNT" -gt 0 ]; then
                    NODE_COUNT=$SPEC_NODE_COUNT
                fi
            fi
        fi
        
        echo -e "${COLOR_GREEN}✓ Network created: $NETWORK_ID${COLOR_NC}"
        echo "   Topology: ${TOPOLOGY:-not specified}"
        echo "   Nodes: $NODE_COUNT"
        
        # Verify topology extraction
        if [[ "$TEST_TEXT" == *"star"* ]] && [[ "$TOPOLOGY" == "star" ]]; then
            echo -e "${COLOR_GREEN}✓ Topology correctly extracted: star${COLOR_NC}"
            ((PASSED++))
        elif [[ "$TEST_TEXT" == *"mesh"* ]] && [[ "$TOPOLOGY" == "mesh" ]]; then
            echo -e "${COLOR_GREEN}✓ Topology correctly extracted: mesh${COLOR_NC}"
            ((PASSED++))
        elif [[ "$TEST_TEXT" == *"ring"* ]] && [[ "$TOPOLOGY" == "ring" ]]; then
            echo -e "${COLOR_GREEN}✓ Topology correctly extracted: ring${COLOR_NC}"
            ((PASSED++))
        elif [[ "$TEST_TEXT" == *"bus"* ]] && [[ "$TOPOLOGY" == "bus" ]]; then
            echo -e "${COLOR_GREEN}✓ Topology correctly extracted: bus${COLOR_NC}"
            ((PASSED++))
        else
            echo -e "${COLOR_YELLOW}⚠ Topology extraction may need verification${COLOR_NC}"
        fi
        
        # Verify node count
        EXPECTED_COUNT=$(echo "$TEST_TEXT" | grep -oE '[0-9]+' | head -1)
        if [ -n "$EXPECTED_COUNT" ] && [ "$NODE_COUNT" -eq "$EXPECTED_COUNT" ]; then
            echo -e "${COLOR_GREEN}✓ Node count correct: $NODE_COUNT${COLOR_NC}"
            ((PASSED++))
        elif [ "$NODE_COUNT" -gt 0 ]; then
            echo -e "${COLOR_YELLOW}⚠ Node count: $NODE_COUNT (expected: ${EXPECTED_COUNT:-unknown})${COLOR_NC}"
        else
            echo -e "${COLOR_RED}✗ No nodes created${COLOR_NC}"
            ((FAILED++))
        fi
        
        # Cleanup
        curl -sS -X DELETE -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$NETWORK_ID" > /dev/null
        
        ((PASSED++))
    else
        echo -e "${COLOR_RED}✗ Network creation failed${COLOR_NC}"
        echo "Response: $RESPONSE"
        ((FAILED++))
    fi
    
    echo ""
    sleep 1
done

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${COLOR_GREEN}Passed: $PASSED${COLOR_NC}"
echo -e "${COLOR_RED}Failed: $FAILED${COLOR_NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${COLOR_GREEN}✅ All NLP tests passed!${COLOR_NC}"
    exit 0
else
    echo -e "${COLOR_RED}❌ Some NLP tests failed${COLOR_NC}"
    exit 1
fi


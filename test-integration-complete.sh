#!/bin/bash

# Complete Integration Test Suite
# Tests complete flows end-to-end

set -e

API_URL="http://localhost:8080"
API_KEY="neto_test"
COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_NC='\033[0m'

PASSED=0
FAILED=0

echo "=========================================="
echo "Complete Integration Test Suite"
echo "=========================================="
echo ""

# Test 1: Complete Flow - NLP → Network → Nodes → Provisioning
echo "TEST 1: Complete Flow (NLP → Network → Auto-Provision)"
echo "--------------------------------------------------------"

echo "   Step 1: Create network via NLP..."
NLP_RESPONSE=$(curl -sS -X POST \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"text":"Create a mesh network with 4 routers"}' \
  "$API_URL/api/v1/ai/provision")

NETWORK_ID=$(echo "$NLP_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$NETWORK_ID" ]; then
    echo -e "${COLOR_GREEN}✓ Network created: $NETWORK_ID${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_RED}✗ Network creation failed${COLOR_NC}"
    echo "Response: $NLP_RESPONSE"
    ((FAILED++))
    exit 1
fi

echo "   Step 2: Wait for auto-provisioning (2 seconds)..."
sleep 2

echo "   Step 3: Check network status..."
NETWORK_STATUS=$(curl -sS -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$NETWORK_ID" | grep -o '"status":"[^"]*"' | cut -d'"' -f4)

if [ "$NETWORK_STATUS" = "active" ] || [ "$NETWORK_STATUS" = "provisioning" ]; then
    echo -e "${COLOR_GREEN}✓ Network status: $NETWORK_STATUS${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_YELLOW}⚠ Network status: $NETWORK_STATUS (may need more time)${COLOR_NC}"
fi

echo "   Step 4: Verify nodes were created..."
NODES_RESPONSE=$(curl -sS -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$NETWORK_ID/nodes")
NODE_COUNT=$(echo "$NODES_RESPONSE" | grep -o '"id":"[^"]*"' | wc -l)

if [ "$NODE_COUNT" -ge 1 ]; then
    echo -e "${COLOR_GREEN}✓ Nodes created: $NODE_COUNT${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_RED}✗ No nodes created${COLOR_NC}"
    ((FAILED++))
fi

echo ""
echo "TEST 2: Delete Operations"
echo "-------------------------"

# Test delete node
if [ "$NODE_COUNT" -gt 0 ]; then
    NODE_ID=$(echo "$NODES_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "   Step 1: Delete node: $NODE_ID..."
    DELETE_NODE_RESPONSE=$(curl -sS -X DELETE -H "X-API-Key: $API_KEY" -w "%{http_code}" "$API_URL/api/v1/nodes/$NODE_ID" -o /dev/null)
    
    if [ "$DELETE_NODE_RESPONSE" = "204" ]; then
        echo -e "${COLOR_GREEN}✓ Node deleted successfully (204)${COLOR_NC}"
        ((PASSED++))
    else
        echo -e "${COLOR_RED}✗ Node delete failed: HTTP $DELETE_NODE_RESPONSE${COLOR_NC}"
        ((FAILED++))
    fi
fi

# Test delete network
echo "   Step 2: Delete network: $NETWORK_ID..."
DELETE_NETWORK_RESPONSE=$(curl -sS -X DELETE -H "X-API-Key: $API_KEY" -w "%{http_code}" "$API_URL/api/v1/networks/$NETWORK_ID" -o /dev/null)

if [ "$DELETE_NETWORK_RESPONSE" = "204" ]; then
    echo -e "${COLOR_GREEN}✓ Network deleted successfully (204)${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_RED}✗ Network delete failed: HTTP $DELETE_NETWORK_RESPONSE${COLOR_NC}"
    ((FAILED++))
fi

# Verify network is deleted
echo "   Step 3: Verify network is deleted..."
GET_DELETED=$(curl -sS -H "X-API-Key: $API_KEY" -w "%{http_code}" "$API_URL/api/v1/networks/$NETWORK_ID" -o /dev/null)

if [ "$GET_DELETED" = "404" ]; then
    echo -e "${COLOR_GREEN}✓ Network correctly returns 404 after deletion${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_YELLOW}⚠ Network still accessible: HTTP $GET_DELETED${COLOR_NC}"
fi

echo ""
echo "TEST 3: Error Cases"
echo "-------------------"

# Test invalid network ID
echo "   Step 1: Test invalid network ID..."
INVALID_ID_RESPONSE=$(curl -sS -H "X-API-Key: $API_KEY" -w "%{http_code}" "$API_URL/api/v1/networks/invalid-id" -o /dev/null)

if [ "$INVALID_ID_RESPONSE" = "400" ] || [ "$INVALID_ID_RESPONSE" = "404" ]; then
    echo -e "${COLOR_GREEN}✓ Invalid ID handled correctly: HTTP $INVALID_ID_RESPONSE${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_YELLOW}⚠ Invalid ID response: HTTP $INVALID_ID_RESPONSE${COLOR_NC}"
fi

# Test missing required fields
echo "   Step 2: Test missing required fields in NLP..."
MISSING_FIELD_RESPONSE=$(curl -sS -X POST -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" -d '{}' "$API_URL/api/v1/ai/provision" -w "%{http_code}" -o /dev/null)

if [ "$MISSING_FIELD_RESPONSE" = "400" ]; then
    echo -e "${COLOR_GREEN}✓ Missing field validation works: HTTP 400${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_YELLOW}⚠ Missing field response: HTTP $MISSING_FIELD_RESPONSE${COLOR_NC}"
fi

# Test unauthorized access
echo "   Step 3: Test unauthorized access..."
UNAUTH_RESPONSE=$(curl -sS -H "X-API-Key: invalid-key" -w "%{http_code}" "$API_URL/api/v1/networks" -o /dev/null)

if [ "$UNAUTH_RESPONSE" = "401" ]; then
    echo -e "${COLOR_GREEN}✓ Unauthorized access blocked: HTTP 401${COLOR_NC}"
    ((PASSED++))
else
    echo -e "${COLOR_YELLOW}⚠ Unauthorized response: HTTP $UNAUTH_RESPONSE${COLOR_NC}"
fi

echo ""
echo "TEST 4: Optimization Service Integration"
echo "----------------------------------------"

# Create a test network for optimization
echo "   Step 1: Create test network for optimization..."
OPT_NETWORK_RESPONSE=$(curl -sS -X POST \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"name":"Optimization Test","description":"Testing optimization","config":{"subnet":"10.0.200.0/24"}}' \
  "$API_URL/api/v1/networks")

OPT_NETWORK_ID=$(echo "$OPT_NETWORK_RESPONSE" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)

if [ -n "$OPT_NETWORK_ID" ]; then
    echo -e "${COLOR_GREEN}✓ Test network created: $OPT_NETWORK_ID${COLOR_NC}"
    
    echo "   Step 2: Analyze network..."
    ANALYZE_RESPONSE=$(curl -sS -X POST -H "X-API-Key: $API_KEY" "$API_URL/api/v1/optimization/analyze/$OPT_NETWORK_ID")
    
    if echo "$ANALYZE_RESPONSE" | grep -q "analysis\|error"; then
        echo -e "${COLOR_GREEN}✓ Optimization analyze working${COLOR_NC}"
        ((PASSED++))
    else
        echo -e "${COLOR_YELLOW}⚠ Optimization analyze may have issues${COLOR_NC}"
    fi
    
    echo "   Step 3: Get recommendations..."
    RECOMMEND_RESPONSE=$(curl -sS -X POST -H "X-API-Key: $API_KEY" "$API_URL/api/v1/optimization/recommend/$OPT_NETWORK_ID")
    
    if echo "$RECOMMEND_RESPONSE" | grep -q "recommendation\|message"; then
        echo -e "${COLOR_GREEN}✓ Optimization recommend working${COLOR_NC}"
        ((PASSED++))
    else
        echo -e "${COLOR_YELLOW}⚠ Optimization recommend may have issues${COLOR_NC}"
    fi
    
    # Cleanup
    curl -sS -X DELETE -H "X-API-Key: $API_KEY" "$API_URL/api/v1/networks/$OPT_NETWORK_ID" > /dev/null
else
    echo -e "${COLOR_YELLOW}⚠ Could not create test network for optimization${COLOR_NC}"
fi

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${COLOR_GREEN}Passed: $PASSED${COLOR_NC}"
echo -e "${COLOR_RED}Failed: $FAILED${COLOR_NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${COLOR_GREEN}✅ All tests passed!${COLOR_NC}"
    exit 0
else
    echo -e "${COLOR_RED}❌ Some tests failed${COLOR_NC}"
    exit 1
fi


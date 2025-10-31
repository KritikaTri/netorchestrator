#!/bin/bash

# NetOrchestrator MVP API Test Script

set -e

API_URL="http://localhost:8080"
TOKEN=""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧪 Testing NetOrchestrator API..."
echo "=================================="
echo ""

# Test 1: Health Check
echo "${YELLOW}Test 1: Health Check${NC}"
response=$(curl -s -w "\n%{http_code}" ${API_URL}/health)
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$status_code" = "200" ]; then
    echo "${GREEN}✅ Health check passed${NC}"
    echo "Response: $body"
else
    echo "${RED}❌ Health check failed (HTTP $status_code)${NC}"
    echo "Response: $body"
    exit 1
fi

echo ""

# Test 2: Login
echo "${YELLOW}Test 2: User Login${NC}"
response=$(curl -s -w "\n%{http_code}" -X POST ${API_URL}/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$status_code" = "200" ]; then
    echo "${GREEN}✅ Login successful${NC}"
    TOKEN=$(echo "$body" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
    echo "Token received: ${TOKEN:0:50}..."
else
    echo "${RED}❌ Login failed (HTTP $status_code)${NC}"
    echo "Response: $body"
fi

echo ""

# Test 3: Get Profile
if [ -n "$TOKEN" ]; then
    echo "${YELLOW}Test 3: Get User Profile${NC}"
    response=$(curl -s -w "\n%{http_code}" ${API_URL}/api/v1/auth/profile \
      -H "Authorization: Bearer $TOKEN")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$status_code" = "200" ]; then
        echo "${GREEN}✅ Profile retrieved${NC}"
        echo "Response: $body"
    else
        echo "${RED}❌ Profile request failed (HTTP $status_code)${NC}"
        echo "Response: $body"
    fi
fi

echo ""

# Test 4: List Networks
if [ -n "$TOKEN" ]; then
    echo "${YELLOW}Test 4: List Networks${NC}"
    response=$(curl -s -w "\n%{http_code}" ${API_URL}/api/v1/networks \
      -H "Authorization: Bearer $TOKEN")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$status_code" = "200" ]; then
        echo "${GREEN}✅ Networks listed${NC}"
        echo "Response: $body"
    else
        echo "${RED}❌ List networks failed (HTTP $status_code)${NC}"
        echo "Response: $body"
    fi
fi

echo ""

# Test 5: Create Network
if [ -n "$TOKEN" ]; then
    echo "${YELLOW}Test 5: Create Network${NC}"
    response=$(curl -s -w "\n%{http_code}" -X POST ${API_URL}/api/v1/networks \
      -H "Authorization: Bearer $TOKEN" \
      -H "Content-Type: application/json" \
      -d '{"name":"test-network","subnet":"10.0.0.0/24","description":"Test network"}')
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$status_code" = "200" ] || [ "$status_code" = "201" ]; then
        echo "${GREEN}✅ Network created${NC}"
        echo "Response: $body"
    else
        echo "${YELLOW}⚠️  Create network returned HTTP $status_code${NC}"
        echo "Response: $body"
    fi
fi

echo ""

# Test 6: Get Intelligence Models
if [ -n "$TOKEN" ]; then
    echo "${YELLOW}Test 6: Get AI/ML Models${NC}"
    response=$(curl -s -w "\n%{http_code}" ${API_URL}/api/v1/intelligence/models \
      -H "Authorization: Bearer $TOKEN")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$status_code" = "200" ]; then
        echo "${GREEN}✅ Models listed${NC}"
        echo "Response: $body"
    else
        echo "${YELLOW}⚠️  Get models returned HTTP $status_code${NC}"
        echo "Response: $body"
    fi
fi

echo ""

# Test 7: Logout
if [ -n "$TOKEN" ]; then
    echo "${YELLOW}Test 7: Logout${NC}"
    response=$(curl -s -w "\n%{http_code}" -X POST ${API_URL}/api/v1/auth/logout \
      -H "Authorization: Bearer $TOKEN")
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$status_code" = "200" ]; then
        echo "${GREEN}✅ Logout successful${NC}"
        echo "Response: $body"
    else
        echo "${RED}❌ Logout failed (HTTP $status_code)${NC}"
        echo "Response: $body"
    fi
fi

echo ""
echo "=================================="
echo "${GREEN}✅ API tests completed!${NC}"
echo ""


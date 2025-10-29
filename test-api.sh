#!/bin/bash

# Test script for NetOrchestrator API Gateway

echo "Testing NetOrchestrator API Gateway..."

# Test health endpoint (should work even without database)
echo "Testing health endpoint..."
curl -s http://localhost:8080/health || echo "Health endpoint not available"

echo ""
echo "Testing metrics endpoint..."
curl -s http://localhost:8080/metrics || echo "Metrics endpoint not available"

echo ""
echo "Testing API endpoints (these will fail without database but should return proper error codes)..."
echo "Testing networks endpoint..."
curl -s -w "HTTP Status: %{http_code}\n" http://localhost:8080/api/v1/networks || echo "Networks endpoint not available"

echo ""
echo "API Gateway test completed!"

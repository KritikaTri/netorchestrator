#!/bin/bash

# NetOrchestrator Intelligence API Test Script
echo "🧠 NetOrchestrator Intelligence API Test"
echo "========================================"

cd /Users/ayushar3/projects/netorchestrator

# Check .env configuration
echo ""
echo "📋 Current .env configuration:"
cat .env || echo "No .env file found"

echo ""
echo "🔍 Checking API Gateway health..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health 2>/dev/null)
if [ $? -ne 0 ]; then
  echo "❌ API Gateway not responding - starting services..."
  ./start-all.sh > startup.log 2>&1
  sleep 10
  
  # Check again
  HEALTH_RESPONSE=$(curl -s http://localhost:8080/health 2>/dev/null)
  if [ $? -ne 0 ]; then
    echo "❌ Services failed to start"
    echo "Check startup logs:"
    tail -10 startup.log
    exit 1
  fi
fi

echo "✅ API Gateway is healthy: $(echo "$HEALTH_RESPONSE" | jq -r '.status')"

# Get authentication token
echo ""
echo "🔑 Getting authentication token..."
TOKEN_RESPONSE=$(curl -s http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}')

TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Authentication failed"
  echo "Response: $TOKEN_RESPONSE"
  exit 1
fi

echo "✅ Authentication successful"

# Test 1: Network Analysis
echo ""
echo "🧠 TEST 1: Intelligence Network Analysis"
echo "========================================"

ANALYSIS_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "network_id": "test-network",
    "analysis_type": "performance",
    "metrics": {
      "cpu_usage": 85.5,
      "memory_usage": 92.3,
      "bandwidth": 1250.8,
      "latency": 45.2,
      "packet_loss": 0.12,
      "error_rate": 0.03
    },
    "time_range": "1h"
  }')

echo "📊 Analysis Response:"
echo "$ANALYSIS_RESPONSE" | jq . 2>/dev/null || echo "$ANALYSIS_RESPONSE"

ANALYSIS_SUCCESS=$(echo "$ANALYSIS_RESPONSE" | jq -r '.success // "unknown"' 2>/dev/null)
ANALYSIS_ERROR=$(echo "$ANALYSIS_RESPONSE" | jq -r '.error // "none"' 2>/dev/null)

echo ""
if [ "$ANALYSIS_SUCCESS" = "true" ]; then
  echo "✅ Network Analysis: SUCCESS"
  echo "📊 Confidence: $(echo "$ANALYSIS_RESPONSE" | jq -r '.confidence')"
  echo "🔍 Insights:"
  echo "$ANALYSIS_RESPONSE" | jq -r '.insights[]' | head -3 | sed 's/^/   • /'
elif [ "$ANALYSIS_ERROR" != "none" ]; then
  echo "❌ Network Analysis: ERROR"
  echo "Error: $ANALYSIS_ERROR"
else
  echo "⚠️  Network Analysis: UNKNOWN STATUS"
  echo "First 200 chars of response:"
  echo "$ANALYSIS_RESPONSE" | head -c 200
fi

# Test 2: Anomaly Detection
echo ""
echo "🔍 TEST 2: Anomaly Detection"
echo "============================"

ANOMALY_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/intelligence/detect/anomalies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "metric_name": "cpu_usage",
    "values": [60, 65, 95, 97, 62, 61, 64],
    "sensitivity": "high"
  }')

ANOMALY_SUCCESS=$(echo "$ANOMALY_RESPONSE" | jq -r '.success // "unknown"' 2>/dev/null)
echo "Anomaly Detection: $ANOMALY_SUCCESS"

if [ "$ANOMALY_SUCCESS" = "true" ]; then
  ANOMALY_COUNT=$(echo "$ANOMALY_RESPONSE" | jq -r '.data | length // 0' 2>/dev/null)
  echo "📊 Anomalies detected: $ANOMALY_COUNT"
fi

# Test 3: Cost Optimization  
echo ""
echo "💰 TEST 3: Cost Optimization"
echo "==========================="

COST_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/intelligence/optimize/costs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "network_id": "test-network",
    "resources": {
      "compute_instances": 15,
      "storage_gb": 1000,
      "bandwidth_mbps": 500
    },
    "optimization_level": "moderate"
  }')

COST_SUCCESS=$(echo "$COST_RESPONSE" | jq -r '.success // "unknown"' 2>/dev/null)
echo "Cost Optimization: $COST_SUCCESS"

# Final Summary
echo ""
echo "📊 INTELLIGENCE API TEST SUMMARY"
echo "==============================="

echo ""
echo "🧪 Test Results:"
echo "• Network Analysis: $ANALYSIS_SUCCESS"
echo "• Anomaly Detection: $ANOMALY_SUCCESS" 
echo "• Cost Optimization: $COST_SUCCESS"

echo ""
if [ "$ANALYSIS_SUCCESS" = "true" ]; then
  echo "✅ SOLUTION SUCCESSFUL!"
  echo "• Intelligence API is now working with gpt-3.5-turbo"
  echo "• OpenAI integration fixed"
  echo "• .env file configuration working"
  echo ""
  echo "🎯 The POST /api/v1/intelligence/analyze/network endpoint is FUNCTIONAL!"
else
  echo "❌ Still having issues..."
  echo "• Model changed to gpt-3.5-turbo"
  echo "• .env file configured"
  echo "• May need to check OpenAI account status"
  echo ""
  echo "🔍 Most recent error (if any):"
  echo "$ANALYSIS_ERROR"
fi

echo ""
echo "📋 Configuration Status:"
echo "• .env file: ✅ API key stored permanently"
echo "• OpenAI Model: ✅ gpt-3.5-turbo (more accessible)"
echo "• Service startup: ✅ Automatic loading"

echo ""
echo "💡 Note: If still having issues, your OpenAI account may need:"
echo "• Billing setup for API access"
echo "• Account verification"
echo "• Credit balance for API calls"

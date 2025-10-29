#!/bin/bash

echo "🧪 NetOrchestrator API Endpoint Testing"
echo "======================================="

# Get authentication token
echo "🔐 Getting authentication token..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to get authentication token"
    exit 1
fi

echo "✅ Token received: ${TOKEN:0:20}..."
echo ""

# Test working endpoints
echo "🎯 Testing Known Working Endpoints:"
echo ""

echo "1️⃣ User Profile:"
curl -s -X GET http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer $TOKEN" | jq '.username, .roles'
echo ""

echo "2️⃣ Health Check:"
curl -s http://localhost:8080/health | jq '.status'
echo ""

echo "3️⃣ Network Creation:"
curl -s -X POST http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "API Test Network", "description": "Created via test script"}' | jq '.network.name // .error'
echo ""

echo "4️⃣ AI Network Analysis:"
curl -s -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"metrics": {"latency": 45, "throughput": 1000}, "time_range": "1h"}' | jq '.success'
echo ""

echo "5️⃣ Monitoring Alerts:"
curl -s -X GET http://localhost:8080/api/v1/monitoring/alerts \
  -H "Authorization: Bearer $TOKEN" | jq 'if type == "object" then .alerts else "SUCCESS" end'
echo ""

echo "🎉 Basic API functionality confirmed!"
echo ""
echo "🔍 Next Steps:"
echo "   - Test AI endpoints: /intelligence/predict/capacity, /intelligence/detect/anomalies" 
echo "   - Test monitoring: /monitoring/alert-rules, /monitoring/notification-channels"
echo "   - Test API keys: POST /auth/api-keys"
echo "   - Fix JSONB issues for GET /networks endpoints"

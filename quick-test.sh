#!/bin/bash

echo "🚀 NetOrchestrator Quick Test"
echo "============================"

# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.access_token')

echo "✅ Authentication: $(if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then echo "SUCCESS"; else echo "FAILED"; fi)"

# Test health
HEALTH=$(curl -s http://localhost:8080/health | jq -r '.status')
echo "✅ Health Check: $HEALTH"

# Test network creation
NETWORK_RESULT=$(curl -s -X POST http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Quick Test Network", "description": "Testing"}' | jq -r '.network.id // "FAILED"')

echo "✅ Network Creation: $(if [ "$NETWORK_RESULT" != "FAILED" ] && [ "$NETWORK_RESULT" != "null" ]; then echo "SUCCESS ($NETWORK_RESULT)"; else echo "FAILED"; fi)"

# Test AI analysis
AI_RESULT=$(curl -s -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"metrics": {"latency": 45, "throughput": 1000}, "time_range": "1h"}' | jq -r '.success')

echo "✅ AI Intelligence: $(if [ "$AI_RESULT" = "true" ]; then echo "SUCCESS"; else echo "FAILED"; fi)"

echo ""
echo "🎉 NetOrchestrator is ready for development!"
echo ""
echo "🔗 Available endpoints:"
echo "   • Health: http://localhost:8080/health"
echo "   • API Docs: http://localhost:8080/swagger/index.html"
echo "   • Metrics: http://localhost:8080/metrics"
echo ""
echo "🔐 Demo credentials:"
echo "   • Admin: admin / admin123"
echo "   • User: user / user123"

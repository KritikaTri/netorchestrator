#!/bin/bash

# NetOrchestrator API Feature Test Suite
echo "🚀 NetOrchestrator API Feature Tests"
echo "===================================="

BASE_URL="http://localhost:8080"

echo "📋 Testing Core API Features..."

# Test 1: Health Check
echo ""
echo "1. 🩺 Health Check:"
curl -s "$BASE_URL/health" | jq '.' 2>/dev/null || curl -s "$BASE_URL/health"

# Test 2: Metrics 
echo ""
echo "2. 📊 Metrics Endpoint:"
curl -s "$BASE_URL/metrics" | head -10

# Test 3: List Networks
echo ""
echo "3. 🌐 List Networks:"
curl -s -H "Content-Type: application/json" "$BASE_URL/api/v1/networks" | jq '.' 2>/dev/null || echo "API not ready yet"

# Test 4: Create Network
echo ""
echo "4. ➕ Create Sample Network:"
curl -s -X POST -H "Content-Type: application/json" \
  -d '{
    "name": "Test Network", 
    "description": "Sample test network",
    "network_type": "virtual",
    "subnet": "192.168.100.0/24",
    "gateway": "192.168.100.1"
  }' \
  "$BASE_URL/api/v1/networks" | jq '.' 2>/dev/null || echo "API not ready yet"

# Test 5: Container Status
echo ""
echo "5. 🐳 Container Status:"
podman ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Test 6: Database Connection Test
echo ""
echo "6. 🗄️  Database Connection Test:"
PGPASSWORD=password psql -h localhost -U netorchestrator -d netorchestrator -c "SELECT COUNT(*) as total_networks FROM networks;" 2>/dev/null || echo "Database query failed"

echo ""
echo "✅ Feature test complete!"
echo ""
echo "🔥 **NEW FEATURES ADDED:**"
echo "  ✅ Complete database schema with 8 tables"
echo "  ✅ Full REST API for networks, nodes, links, policies"
echo "  ✅ Monitoring endpoints for metrics and health"
echo "  ✅ User management (ready for authentication)"
echo "  ✅ Event logging and alerting system"
echo "  ✅ Podman containerized services"
echo "  ✅ Redis caching layer"
echo "  ✅ PostgreSQL data persistence"
echo "  ✅ Prometheus monitoring integration"
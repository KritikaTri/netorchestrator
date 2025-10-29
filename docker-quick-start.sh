#!/bin/bash

echo "🐳 NetOrchestrator Docker Quick Start"
echo "===================================="
echo ""

set -e

echo "📋 Starting with Docker..."
echo ""

# Clean up any existing containers (optional)
echo "🧹 Cleaning up existing containers..."
docker stop postgres redis influxdb 2>/dev/null || true
docker rm postgres redis influxdb 2>/dev/null || true

echo "🚀 Starting fresh Docker containers..."

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
docker run -d --name postgres \
  -e POSTGRES_DB=netorchestrator \
  -e POSTGRES_USER=netorchestrator \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 postgres:15

# Start Redis
echo "🔴 Starting Redis..."
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Start InfluxDB
echo "📊 Starting InfluxDB..."
docker run -d --name influxdb \
  -e DOCKER_INFLUXDB_INIT_MODE=setup \
  -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
  -e DOCKER_INFLUXDB_INIT_PASSWORD=password \
  -e DOCKER_INFLUXDB_INIT_ORG=netorchestrator \
  -e DOCKER_INFLUXDB_INIT_BUCKET=network_metrics \
  -p 8086:8086 influxdb:2.7

echo ""
echo "⏳ Waiting for containers to be ready..."
sleep 10

# Verify containers are running
echo "🔍 Checking container status..."
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "📊 Initializing database..."
docker exec -i postgres psql -U netorchestrator -d netorchestrator < schema.sql > /dev/null

# Add GORM compatibility
echo "🔧 Adding GORM compatibility..."
docker exec postgres psql -U netorchestrator -d netorchestrator -c "
ALTER TABLE networks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE networks ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}';
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}';
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS position JSONB DEFAULT '{}';
ALTER TABLE links ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE links ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}';
ALTER TABLE policies ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
ALTER TABLE policies ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;
" > /dev/null

# Add sample data
echo "📊 Adding sample networks..."
docker exec postgres psql -U netorchestrator -d netorchestrator -c "
INSERT INTO networks (name, description, status, network_type, subnet, gateway, user_id, metadata)
SELECT 'Production Network', 'Main production environment', 'active', 'enterprise', '10.0.0.0/16', '10.0.0.1', u.id, '{\"environment\": \"production\"}'
FROM users u WHERE u.username = 'admin'
ON CONFLICT DO NOTHING;

INSERT INTO networks (name, description, status, network_type, subnet, gateway, user_id, metadata)  
SELECT 'Development Network', 'Development and testing environment', 'active', 'development', '172.16.0.0/24', '172.16.0.1', u.id, '{\"environment\": \"development\"}'
FROM users u WHERE u.username = 'admin'
ON CONFLICT DO NOTHING;
" > /dev/null

echo ""
echo "🔨 Building and starting API Gateway..."
go build -o bin/api-gateway ./cmd/api-gateway/

# Stop existing API Gateway
pkill -f api-gateway 2>/dev/null || true
sleep 2

# Start API Gateway
nohup ./bin/api-gateway > api-gateway.log 2>&1 &

echo "⏳ Waiting for API Gateway startup..."
sleep 5

echo ""
echo "✅ Testing platform..."

# Test health
HEALTH=$(curl -s http://localhost:8080/health | jq -r '.status // "FAILED"')
echo "✅ Health Check: $HEALTH"

# Test auth
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.access_token // "FAILED"')
echo "✅ Authentication: $(if [ "$TOKEN" != "FAILED" ] && [ "$TOKEN" != "null" ]; then echo "SUCCESS"; else echo "FAILED"; fi)"

# Test AI
AI_TEST=$(curl -s -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"metrics": {"latency": 45, "throughput": 1000}, "time_range": "1h"}' | jq -r '.success // "false"')
echo "✅ AI Intelligence: $(if [ "$AI_TEST" = "true" ]; then echo "SUCCESS"; else echo "FAILED"; fi)"

# Count networks
NETWORKS=$(docker exec postgres psql -U netorchestrator -d netorchestrator -t -c "SELECT COUNT(*) FROM networks;" 2>/dev/null | tr -d ' ')
echo "✅ Sample Data: $NETWORKS networks created"

echo ""
echo "🎉 NETORCHESTRATOR RUNNING WITH DOCKER!"
echo "======================================"
echo ""
echo "🌐 Platform: http://localhost:8080"
echo "📚 API Docs: http://localhost:8080/swagger/index.html"
echo "💊 Health: http://localhost:8080/health"
echo ""
echo "🔐 Demo Credentials:"
echo "   • admin / admin123 (full access)"  
echo "   • user / user123 (standard access)"
echo ""
echo "🗄️ Database: postgresql://netorchestrator:password@localhost:5432/netorchestrator"
echo ""
echo "🧪 Run tests:"
echo "   ./quick-test.sh"
echo "   ./test-ai-complete.sh"

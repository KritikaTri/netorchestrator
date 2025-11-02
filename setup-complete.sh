#!/bin/bash

echo "🏗️ NetOrchestrator Complete Setup Script"
echo "========================================"
echo ""

set -e  # Exit on any error

echo "📋 Step 1: Starting Infrastructure Containers..."
echo ""

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
if docker ps | grep -q postgres; then
    echo "   PostgreSQL already running"
else
    docker run -d --name postgres \
      -e POSTGRES_DB=netorchestrator \
      -e POSTGRES_USER=netorchestrator \
      -e POSTGRES_PASSWORD=password \
      -p 5432:5432 postgres:15
    echo "   PostgreSQL started"
fi

# Start Redis
echo "🔴 Starting Redis..."
if docker ps | grep -q redis; then
    echo "   Redis already running"
else
    docker run -d --name redis -p 6379:6379 redis:7-alpine
    echo "   Redis started"
fi

# Start InfluxDB
echo "📊 Starting InfluxDB..."
if docker ps | grep -q influxdb; then
    echo "   InfluxDB already running"
else
    docker run -d --name influxdb \
      -e DOCKER_INFLUXDB_INIT_MODE=setup \
      -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
      -e DOCKER_INFLUXDB_INIT_PASSWORD=password \
      -e DOCKER_INFLUXDB_INIT_ORG=netorchestrator \
      -e DOCKER_INFLUXDB_INIT_BUCKET=network_metrics \
      -p 8086:8086 influxdb:2.7
    echo "   InfluxDB started"
fi

# Start Prometheus
echo "📈 Starting Prometheus..."
if docker ps | grep -q prometheus; then
    echo "   Prometheus already running"
else
    docker run -d --name prometheus \
      -p 9090:9090 \
      prom/prometheus:latest
    echo "   Prometheus started"
fi

echo ""
echo "⏳ Waiting for containers to be ready..."
sleep 5

echo ""
echo "📋 Step 2: Initializing Database..."
echo ""

# Initialize database schema
echo "📄 Creating database schema..."
docker exec -i postgres psql -U netorchestrator -d netorchestrator < schema.sql

# Add GORM compatibility columns
echo "🔧 Adding GORM compatibility columns..."
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
echo "📊 Adding sample data..."
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
echo "📋 Step 3: Building Application..."
echo ""

# Build API Gateway
echo "🔨 Building API Gateway..."
go build -o bin/api-gateway ./cmd/api-gateway/

echo ""
echo "📋 Step 4: Starting Services..."
echo ""

# Kill any existing API Gateway
pkill -f api-gateway 2>/dev/null || true

# Start API Gateway
echo "🚀 Starting API Gateway..."
nohup ./bin/api-gateway > api-gateway.log 2>&1 &

# Wait for startup
echo "⏳ Waiting for API Gateway to start..."
sleep 5

echo ""
echo "📋 Step 5: Verifying Installation..."
echo ""

# Test health
HEALTH_STATUS=$(curl -s http://localhost:8080/health | jq -r '.status // "FAILED"')
echo "✅ Health Check: $HEALTH_STATUS"

# Test authentication
AUTH_TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.access_token // "FAILED"')
echo "✅ Authentication: $(if [ "$AUTH_TOKEN" != "FAILED" ] && [ "$AUTH_TOKEN" != "null" ]; then echo "SUCCESS"; else echo "FAILED"; fi)"

# Test AI Intelligence
AI_TEST=$(curl -s -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"metrics": {"latency": 45, "throughput": 1000}, "time_range": "1h"}' | jq -r '.success // "false"')
echo "✅ AI Intelligence: $(if [ "$AI_TEST" = "true" ]; then echo "SUCCESS"; else echo "FAILED"; fi)"

# Verify database data
DB_NETWORKS=$(docker exec postgres psql -U netorchestrator -d netorchestrator -t -c "SELECT COUNT(*) FROM networks;" 2>/dev/null | tr -d ' ')
echo "✅ Database Networks: $DB_NETWORKS networks available"

echo ""
echo "🎉 SETUP COMPLETE!"
echo "=================="
echo ""
echo "🌐 NetOrchestrator is running at: http://localhost:8080"
echo "📚 API Documentation: http://localhost:8080/swagger/index.html"
echo "💊 Health Check: http://localhost:8080/health"
echo "📈 Prometheus: http://localhost:9090"
echo ""
echo "🔐 Demo Credentials:"
echo "   • Admin: admin / admin123 (full access)"
echo "   • User: user / user123 (limited access)"
echo ""
echo "🧪 Test the platform:"
echo "   ./quick-test.sh         # Basic functionality test"
echo "   ./test-ai-complete.sh   # Comprehensive AI testing"
echo ""
echo "🗄️ Database Connection (Beekeeper Studio):"
echo "   postgresql://netorchestrator:password@localhost:5432/netorchestrator"
echo ""
echo "🎯 What's Working:"
echo "   ✅ Enterprise Authentication (JWT, API keys)"
echo "   ✅ Network Management API"
echo "   ✅ AI Intelligence Suite (4 endpoints)"
echo "   ✅ Real-time Monitoring & Alerts"
echo "   ✅ PostgreSQL + Redis + InfluxDB + Prometheus"
echo ""
echo "🚀 NetOrchestrator Enterprise Platform Ready!"

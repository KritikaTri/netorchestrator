#!/bin/bash
# NetOrchestrator One-Line Setup - Run this to get everything working

echo "⚡ NetOrchestrator One-Line Setup" && \
docker run -d --name postgres -e POSTGRES_DB=netorchestrator -e POSTGRES_USER=netorchestrator -e POSTGRES_PASSWORD=password -p 5432:5432 postgres:15 && \
docker run -d --name redis -p 6379:6379 redis:7-alpine && \
docker run -d --name influxdb -e DOCKER_INFLUXDB_INIT_MODE=setup -e DOCKER_INFLUXDB_INIT_USERNAME=admin -e DOCKER_INFLUXDB_INIT_PASSWORD=password -e DOCKER_INFLUXDB_INIT_ORG=netorchestrator -e DOCKER_INFLUXDB_INIT_BUCKET=network_metrics -p 8086:8086 influxdb:2.7 && \
sleep 10 && \
docker exec -i postgres psql -U netorchestrator -d netorchestrator < schema.sql && \
docker exec postgres psql -U netorchestrator -d netorchestrator -c "ALTER TABLE networks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP; ALTER TABLE networks ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'; ALTER TABLE nodes ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP; ALTER TABLE nodes ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'; ALTER TABLE nodes ADD COLUMN IF NOT EXISTS position JSONB DEFAULT '{}'; ALTER TABLE links ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP; ALTER TABLE links ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'; ALTER TABLE policies ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP; ALTER TABLE policies ADD COLUMN IF NOT EXISTS config JSONB DEFAULT '{}'; ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP;" && \
go build -o bin/api-gateway ./cmd/api-gateway/ && \
pkill -f api-gateway 2>/dev/null || true && \
./bin/api-gateway &
echo "🎉 NetOrchestrator is starting! Test with: curl http://localhost:8080/health"

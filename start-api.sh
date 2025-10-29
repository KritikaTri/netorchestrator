#!/bin/bash

# Quick API startup script
echo "🚀 Starting NetOrchestrator API Gateway..."

# Check prerequisites
echo "📋 Checking prerequisites..."
if ! command -v docker &> /dev/null; then
    echo "❌ Podman not found"
    exit 1
fi

if ! docker ps | grep -q postgres; then
    echo "❌ PostgreSQL container not running"
    exit 1
fi

echo "✅ Prerequisites met"

# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=netorchestrator
export DB_PASSWORD=password
export DB_NAME=netorchestrator
export REDIS_HOST=localhost
export REDIS_PORT=6379

# Build and run
echo "🔨 Building API..."
go build -o bin/api-gateway ./cmd/api-gateway

echo "🚀 Starting API Gateway..."
./bin/api-gateway

echo "🎉 API Gateway started on http://localhost:8080"
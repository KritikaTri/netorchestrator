#!/bin/bash

# Test script for NetOrchestrator API Gateway with Podman
# This script works with both Docker and Podman

echo "🧪 Testing NetOrchestrator API Gateway..."

# Check if services are running
echo "📊 Checking service status..."
if command -v podman &> /dev/null; then
    echo "Using Podman..."
    COMPOSE_CMD="podman-compose"
    CONTAINER_CMD="podman"
elif command -v docker &> /dev/null; then
    echo "Using Docker..."
    COMPOSE_CMD="docker-compose"
    CONTAINER_CMD="docker"
else
    echo "❌ Neither Podman nor Docker found!"
    exit 1
fi

# Check if containers are running
echo "🔍 Checking running containers..."
$CONTAINER_CMD ps

echo ""
echo "🩺 Testing health endpoint..."
curl -s http://localhost:8080/health || echo "Health endpoint not available"

echo ""
echo "📈 Testing metrics endpoint..."
curl -s http://localhost:8080/metrics || echo "Metrics endpoint not available"

echo ""
echo "🌐 Testing API endpoints (these will fail without database but should return proper error codes)..."
echo "Testing networks endpoint..."
curl -s -w "HTTP Status: %{http_code}\n" http://localhost:8080/api/v1/networks || echo "Networks endpoint not available"

echo ""
echo "✅ API Gateway test completed!"
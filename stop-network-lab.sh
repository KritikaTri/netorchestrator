#!/bin/bash

# NetOrchestrator Network Lab Cleanup Script

echo "🛑 Stopping NetOrchestrator Network Lab..."

# Stop and remove containers
echo "📦 Stopping containers..."
podman stop frr-router alpine-node ubuntu-router vyos-node 2>/dev/null || true
podman rm frr-router alpine-node ubuntu-router vyos-node 2>/dev/null || true

# Remove network
echo "🌐 Cleaning up network..."
podman network rm netorchestrator-lab 2>/dev/null || true

echo "✅ Network Lab stopped and cleaned up!"
echo "💡 Run ./start-network-lab.sh to restart"

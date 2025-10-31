#!/bin/bash

echo "🧹 Stopping and Cleaning NetOrchestrator"
echo "======================================="
echo ""

echo "⚠️  This will completely remove all containers and data!"
echo "      Press Ctrl+C within 5 seconds to cancel..."
sleep 5

echo ""
echo "🛑 Stopping all services..."

# Stop API Gateway
echo "📋 Stopping API Gateway..."
pkill -f api-gateway 2>/dev/null || true

# Stop and remove containers
echo "🐳 Stopping and removing Docker containers..."
docker stop postgres redis influxdb 2>/dev/null || true
docker rm -f postgres redis influxdb 2>/dev/null || true

# Clean up Docker resources
echo "🧹 Cleaning up Docker resources..."
docker container prune -f 2>/dev/null || true
docker volume prune -f 2>/dev/null || true

# Remove generated files
echo "🗑️ Cleaning up generated files..."
rm -f api-gateway.log
rm -rf bin/

echo ""
echo "📊 Final cleanup status:"
CONTAINERS=$(docker ps -a --filter "name=postgres" --filter "name=redis" --filter "name=influxdb" --format "{{.Names}}" | wc -l | tr -d ' ')
PROCESSES=$(pgrep -f api-gateway | wc -l | tr -d ' ')

echo "   • NetOrchestrator containers: $CONTAINERS"
echo "   • API Gateway processes: $PROCESSES"

if [ "$CONTAINERS" -eq "0" ] && [ "$PROCESSES" -eq "0" ]; then
    echo ""
    echo "✅ COMPLETE CLEANUP SUCCESSFUL!"
    echo ""
    echo "🚀 To start fresh:"
    echo "   ./setup-complete.sh        # Complete setup" 
    echo "   ./docker-quick-start.sh    # Docker setup"
    echo "   ./one-line-setup.sh        # Quick setup"
else
    echo ""
    echo "⚠️  Manual cleanup may be needed:"
    echo "   docker ps -a              # Check remaining containers"
    echo "   pgrep -f api-gateway      # Check remaining processes"
fi

echo ""
echo "🧹 NetOrchestrator cleanup complete!"

#!/bin/bash

echo "🛑 Stopping NetOrchestrator Platform"
echo "==================================="
echo ""

echo "📋 Stopping API Gateway..."
# Stop API Gateway process
if pgrep -f api-gateway > /dev/null; then
    pkill -f api-gateway
    echo "✅ API Gateway stopped"
else
    echo "   API Gateway not running"
fi

echo ""
echo "📋 Stopping ALL Docker containers..."

# Get list of all running containers
RUNNING_CONTAINERS=$(docker ps -q)

if [ -n "$RUNNING_CONTAINERS" ]; then
    echo "🐳 Found running containers, stopping them..."
    
    # Stop containers gracefully with timeout
    echo "   Stopping containers gracefully (30s timeout)..."
    docker stop $RUNNING_CONTAINERS --time=30
    
    # Force kill any remaining containers
    REMAINING_CONTAINERS=$(docker ps -q)
    if [ -n "$REMAINING_CONTAINERS" ]; then
        echo "   Force killing remaining containers..."
        docker kill $REMAINING_CONTAINERS
    fi
    
    # Remove stopped containers (optional - commented out for safety)
    # echo "   Cleaning up stopped containers..."
    # docker container prune -f
    
    echo "✅ All containers stopped"
else
    echo "   No containers running"
fi

# Detailed status of known infrastructure containers
echo ""
echo "📊 Infrastructure container status:"
for container in postgres redis influxdb prometheus grafana policy-router policy-host; do
    if docker ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
        STATUS=$(docker ps -a --filter "name=^${container}$" --format "{{.Status}}")
        echo "   • $container: $STATUS"
    fi
done

echo ""
echo "📊 API Gateway status:"
if pgrep -f api-gateway > /dev/null; then
    echo "   ❌ API Gateway still running"
else
    echo "   ✅ API Gateway stopped"
fi

echo ""
echo "📊 Final status check..."
TOTAL_RUNNING_CONTAINERS=$(docker ps -q | wc -l | tr -d ' ')
API_RUNNING=$(pgrep -f api-gateway | wc -l | tr -d ' ')

echo "📊 Summary:"
echo "   • Total running containers: $TOTAL_RUNNING_CONTAINERS"
echo "   • API Gateway processes: $API_RUNNING"

if [ "$TOTAL_RUNNING_CONTAINERS" -eq "0" ] && [ "$API_RUNNING" -eq "0" ]; then
    echo ""
    echo "✅ ALL SERVICES STOPPED SUCCESSFULLY!"
    echo ""
    echo "🔄 To start again:"
    echo "   ./setup-complete.sh    # Complete setup"
    echo "   ./docker-quick-start.sh    # Clean Docker setup"
    echo "   ./one-line-setup.sh    # Quick setup"
else
    echo ""
    echo "⚠️  Some services may still be running"
    echo "   Check with: docker ps"
    echo "   Check with: pgrep -f api-gateway"
fi

echo ""
echo "🛑 NetOrchestrator shutdown complete!"

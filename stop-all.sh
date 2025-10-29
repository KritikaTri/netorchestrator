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
echo "📋 Stopping Docker containers..."

# Stop containers gracefully
if docker ps | grep -q postgres; then
    echo "🐘 Stopping PostgreSQL..."
    docker stop postgres
    echo "✅ PostgreSQL stopped"
else
    echo "   PostgreSQL not running"
fi

if docker ps | grep -q redis; then
    echo "🔴 Stopping Redis..."
    docker stop redis
    echo "✅ Redis stopped"
else
    echo "   Redis not running"
fi

if docker ps | grep -q influxdb; then
    echo "📊 Stopping InfluxDB..."
    docker stop influxdb
    echo "✅ InfluxDB stopped"
else
    echo "   InfluxDB not running"
fi

echo ""
echo "📊 Final status check..."
RUNNING_CONTAINERS=$(docker ps --filter "name=postgres" --filter "name=redis" --filter "name=influxdb" --format "{{.Names}}" | wc -l | tr -d ' ')
API_RUNNING=$(pgrep -f api-gateway | wc -l | tr -d ' ')

echo "📊 Summary:"
echo "   • Running containers: $RUNNING_CONTAINERS"
echo "   • API Gateway processes: $API_RUNNING"

if [ "$RUNNING_CONTAINERS" -eq "0" ] && [ "$API_RUNNING" -eq "0" ]; then
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

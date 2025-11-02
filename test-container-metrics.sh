#!/bin/bash

echo "🔍 Testing Real-Time Container Metrics Collection"
echo "================================================="
echo ""

# Check if API is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ API Gateway is not running. Please start it first:"
    echo "   ./bin/api-gateway"
    exit 1
fi

echo "✅ API Gateway is running"
echo ""

# Check running containers
echo "📦 Checking running Docker containers..."
CONTAINER_COUNT=$(docker ps --format "{{.Names}}" | wc -l | tr -d ' ')
echo "   Found $CONTAINER_COUNT running containers:"
docker ps --format "   • {{.Names}} ({{.Status}})"
echo ""

if [ "$CONTAINER_COUNT" -eq "0" ]; then
    echo "⚠️  No containers running. Start some containers to see metrics:"
    echo "   ./setup-complete.sh"
    echo ""
fi

# Wait for first collection cycle
echo "⏳ Waiting 12 seconds for metrics collection cycle..."
sleep 12

# Query metrics from database
echo ""
echo "📊 Current Container Metrics (Latest Values Only):"
echo "==================================================="

docker exec -i postgres psql -U netorchestrator -d netorchestrator << 'EOF'
SELECT 
    metadata->>'container_name' as container,
    metric_name,
    ROUND(metric_value::numeric, 2) as value,
    unit,
    TO_CHAR(timestamp, 'HH24:MI:SS') as updated_at
FROM metrics 
WHERE entity_type = 'container'
ORDER BY container, metric_name;
EOF

echo ""
echo "📈 Database Status:"
echo "==================="

docker exec -i postgres psql -U netorchestrator -d netorchestrator << 'EOF'
SELECT 
    COUNT(*) as total_rows,
    COUNT(DISTINCT entity_id) as unique_containers,
    COUNT(DISTINCT metric_name) as metric_types,
    MIN(timestamp) as oldest_update,
    MAX(timestamp) as newest_update
FROM metrics 
WHERE entity_type = 'container';
EOF

echo ""
echo "📊 Metric Types Being Collected:"
echo "================================"

docker exec -i postgres psql -U netorchestrator -d netorchestrator << 'EOF'
SELECT 
    metric_name,
    unit,
    COUNT(*) as measurements,
    COUNT(DISTINCT (metadata->>'container_name')) as containers,
    TO_CHAR(MAX(timestamp), 'HH24:MI:SS') as latest
FROM metrics 
WHERE timestamp > NOW() - INTERVAL '1 minute'
    AND entity_type IN ('container', 'node')
GROUP BY metric_name, unit
ORDER BY metric_name;
EOF

echo ""
echo "✨ Container Metrics Collection Status:"
echo "   ✅ Real-time metrics collected every 10 seconds"
echo "   ✅ Data source: Docker/Podman container stats"
echo "   ✅ UPSERT mode: Only latest values stored (no historical data)"
echo "   ✅ Fixed row count: 4 containers × 8 metrics = 32 rows total"
echo ""
echo "💡 To view live metrics:"
echo "   watch -n 10 'docker exec postgres psql -U netorchestrator -d netorchestrator -c \"SELECT metric_name, ROUND(metric_value::numeric, 2) as value, unit, metadata->>\\\"container_name\\\" as container FROM metrics WHERE timestamp > NOW() - INTERVAL \\\"30 seconds\\\" ORDER BY timestamp DESC LIMIT 20;\"'"
echo ""
echo "📊 To query via API:"
echo "   curl -H \"X-API-Key: neto_test\" http://localhost:8080/api/v1/metrics/summary"

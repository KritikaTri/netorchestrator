#!/bin/bash

echo "🚀 Starting Stopped NetOrchestrator Containers"
echo "=============================================="
echo ""

# Function to check if container is running
check_container() {
    local container_name=$1
    if docker ps --format "{{.Names}}" | grep -q "^${container_name}$"; then
        return 0  # Container is running
    else
        return 1  # Container is not running
    fi
}

# Get all stopped containers
stopped_containers=$(docker ps -a --filter "status=exited" --format "{{.Names}}")

if [ -z "$stopped_containers" ]; then
    echo "✅ No stopped containers found. All containers are already running or don't exist."
    echo ""
else
    echo "🔄 Found stopped containers. Starting them..."
    echo ""
    started_count=0
    failed_count=0
    
    for container in $stopped_containers; do
        echo "   🔄 Starting $container..."
        docker start "$container" > /dev/null 2>&1
        sleep 1
        if check_container "$container"; then
            echo "   ✅ $container started successfully"
            ((started_count++))
        else
            echo "   ❌ Failed to start $container"
            ((failed_count++))
        fi
    done
    
    echo ""
    echo "📊 Results: $started_count started, $failed_count failed"
fi

echo ""
echo "📊 Final Container Status:"
echo "========================="

# Get all containers (running and stopped)
all_containers=$(docker ps -a --format "{{.Names}}")
running_count=0
total_count=0

if [ -z "$all_containers" ]; then
    echo "   No containers found"
else
    for container in $all_containers; do
        ((total_count++))
        if check_container "$container"; then
            status=$(docker ps --filter "name=^${container}$" --format "{{.Status}}" | head -1)
            echo "   ✅ $container: Running ($status)"
            ((running_count++))
        else
            status=$(docker ps -a --filter "name=^${container}$" --format "{{.Status}}" | head -1)
            echo "   ❌ $container: Stopped ($status)"
        fi
    done
    
    echo ""
    echo "📊 Summary: $running_count/$total_count containers running"
fi


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


echo ""
echo "✅ Done!"

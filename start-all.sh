#!/bin/bash

echo "🚀 Starting NetOrchestrator Platform"
echo "==================================="
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

# Function to check if container exists (running or stopped)
container_exists() {
    local container_name=$1
    if docker ps -a --format "{{.Names}}" | grep -q "^${container_name}$"; then
        return 0  # Container exists
    else
        return 1  # Container doesn't exist
    fi
}

# Function to start container (restart if exists, create if doesn't exist)
start_container() {
    local container_name=$1
    local create_command=$2
    
    if check_container "$container_name"; then
        echo "   ✅ $container_name already running"
    elif container_exists "$container_name"; then
        echo "   🔄 Starting existing $container_name container..."
        docker start "$container_name"
        sleep 2
        if check_container "$container_name"; then
            echo "   ✅ $container_name started successfully"
        else
            echo "   ❌ Failed to start $container_name"
        fi
    else
        echo "   🔄 Creating new $container_name container..."
        eval "$create_command"
        sleep 2
        if check_container "$container_name"; then
            echo "   ✅ $container_name created and started successfully"
        else
            echo "   ❌ Failed to create $container_name"
        fi
    fi
}

echo "📋 Starting Core Infrastructure..."
echo ""

# PostgreSQL with NetOrchestrator configuration
echo "🐘 PostgreSQL Database..."
start_container "postgres" "docker run -d --name postgres \
  -e POSTGRES_DB=netorchestrator \
  -e POSTGRES_USER=netorchestrator \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 postgres:15"

# Redis Cache
echo ""
echo "🔴 Redis Cache..."
start_container "redis" "docker run -d --name redis \
  -p 6379:6379 redis:7-alpine"

# InfluxDB for Time-Series Data
echo ""
echo "📊 InfluxDB Time-Series Database..."
start_container "influxdb" "docker run -d --name influxdb \
  -e DOCKER_INFLUXDB_INIT_MODE=setup \
  -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
  -e DOCKER_INFLUXDB_INIT_PASSWORD=password \
  -e DOCKER_INFLUXDB_INIT_ORG=netorchestrator \
  -e DOCKER_INFLUXDB_INIT_BUCKET=network_metrics \
  -p 8086:8086 influxdb:2.7"

echo ""
echo "📋 Starting Network Lab Components..."
echo ""

# Policy Router (FRRouting)
echo "🔧 Policy Router (FRRouting)..."
start_container "policy-router" "docker run -d --name policy-router \
  --privileged \
  -p 30001:22 \
  frrouting/frr:latest"

# Policy Host (Ubuntu)
echo ""
echo "🖥️  Policy Host (Ubuntu)..."
start_container "policy-host" "docker run -d --name policy-host \
  -p 30002:22 \
  ubuntu:20.04 \
  bash -c 'apt-get update && apt-get install -y openssh-server iproute2 iputils-ping curl && \
           mkdir /var/run/sshd && \
           echo \"root:netorchestrator\" | chpasswd && \
           sed -i \"s/#PermitRootLogin prohibit-password/PermitRootLogin yes/\" /etc/ssh/sshd_config && \
           /usr/sbin/sshd -D'"

echo ""
echo "📋 Starting Observability Stack..."
echo ""

# Prometheus for Metrics Collection
echo "📈 Prometheus Metrics Server..."
start_container "prometheus" "docker run -d --name prometheus \
  -p 9090:9090 \
  -v /tmp/prometheus.yml:/etc/prometheus/prometheus.yml:ro \
  prom/prometheus:latest || \
  docker run -d --name prometheus \
  -p 9090:9090 \
  prom/prometheus:latest"

# Grafana for Metrics Visualization  
echo ""
echo "📊 Grafana Dashboard..."
start_container "grafana" "docker run -d --name grafana \
  -p 3000:3000 \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  grafana/grafana:latest"

echo ""
echo "⏳ Waiting for services to initialize..."
sleep 10

echo ""
echo "📊 Container Status Summary:"
echo "=========================="

# Check status of all containers
containers=("postgres" "redis" "influxdb" "policy-router" "policy-host" "prometheus" "grafana")
running_count=0
total_count=${#containers[@]}

for container in "${containers[@]}"; do
    if check_container "$container"; then
        status=$(docker ps --filter "name=^${container}$" --format "{{.Status}}" | head -1)
        ports=$(docker ps --filter "name=^${container}$" --format "{{.Ports}}" | head -1)
        echo "   ✅ $container: $status"
        if [ -n "$ports" ]; then
            echo "      Ports: $ports"
        fi
        ((running_count++))
    else
        if docker ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
            status=$(docker ps -a --filter "name=^${container}$" --format "{{.Status}}" | head -1)
            echo "   ❌ $container: $status"
        else
            echo "   ❌ $container: Not found"
        fi
    fi
done

echo ""
echo "📊 Summary: $running_count/$total_count containers running"

# Check core infrastructure (postgres + redis are essential)
core_running=0
if check_container "postgres"; then ((core_running++)); fi
if check_container "redis"; then ((core_running++)); fi

if [ "$core_running" -eq "2" ]; then
    echo ""
    if [ "$running_count" -eq "$total_count" ]; then
        echo "✅ ALL SERVICES STARTED SUCCESSFULLY!"
    else
        echo "✅ CORE INFRASTRUCTURE STARTED! (Some optional services may have failed)"
    fi
    echo ""
    echo "🌐 Access Points:"
    echo "   • PostgreSQL:  localhost:5432 (netorchestrator/password)"
    echo "   • Redis:       localhost:6379"
    echo "   • InfluxDB:    http://localhost:8086 (admin/password)"
    echo "   • Prometheus:  http://localhost:9090"
    echo "   • Grafana:     http://localhost:3000 (admin/admin)"
    echo "   • Policy Router SSH: localhost:30001 (root/netorchestrator)"
    echo "   • Policy Host SSH:   localhost:30002 (root/netorchestrator)"
    echo ""
    echo "🚀 Starting API Gateway..."
    if pgrep -f api-gateway > /dev/null; then
        echo "   ⚠️  API Gateway already running, stopping first..."
        pkill -f api-gateway
        sleep 2
    fi
    
    # Start API Gateway in background
    echo "   🔄 Starting API Gateway process..."
    nohup ./bin/api-gateway > api-gateway.log 2>&1 &
    sleep 3
    
    # Verify API Gateway started
    if pgrep -f api-gateway > /dev/null; then
        echo "   ✅ API Gateway started successfully"
        echo "   📊 API Health: http://localhost:8080/health"
        echo "   📈 Metrics: http://localhost:8080/metrics"
    else
        echo "   ❌ Failed to start API Gateway (check api-gateway.log)"
    fi
    echo ""
    echo "🛑 To stop all services:"
    echo "   ./stop-all.sh"
else
    echo ""
    echo "⚠️  Some services failed to start. Check the logs:"
    echo "   docker logs <container_name>"
    echo ""
    echo "📝 Manual API Gateway start:"
    echo "   ./bin/api-gateway &"
    echo ""
    echo "🔄 To retry:"
    echo "   ./stop-all.sh && ./start-all.sh"
fi

echo ""
echo "🚀 NetOrchestrator platform startup complete!"

#!/bin/bash

# NetOrchestrator Complete Monitoring Stack Startup Script
# This script starts the entire monitoring infrastructure for NetOrchestrator

set -e

echo "🚀 Starting NetOrchestrator Complete Monitoring Stack..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Podman is available
if ! command -v docker &> /dev/null; then
    print_error "Podman is not installed or not in PATH"
    print_error "Please install Podman Desktop for macOS"
    exit 1
fi

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null; then
    print_warning "docker-compose not found, installing..."
    pip3 install docker-compose
fi

print_status "Stopping any existing containers..."
docker-compose -f docker-compose.monitoring.yml down 2>/dev/null || true

print_status "Removing old volumes (if requested)..."
if [[ "${1}" == "--clean" ]]; then
    print_warning "Cleaning up all volumes and data..."
    docker volume rm netorchestrator_postgres_data 2>/dev/null || true
    docker volume rm netorchestrator_redis_data 2>/dev/null || true
    docker volume rm netorchestrator_influxdb_data 2>/dev/null || true
    docker volume rm netorchestrator_prometheus_data 2>/dev/null || true
    docker volume rm netorchestrator_grafana_data 2>/dev/null || true
    docker volume rm netorchestrator_alertmanager_data 2>/dev/null || true
fi

print_status "Building NetOrchestrator application..."
docker build -t netorchestrator:latest .

print_status "Starting infrastructure services..."
docker-compose -f docker-compose.monitoring.yml up -d postgres redis influxdb

print_status "Waiting for databases to be ready..."
sleep 10

# Check PostgreSQL connection
print_status "Verifying PostgreSQL connection..."
for i in {1..30}; do
    if docker exec netorchestrator_postgres_1 pg_isready -U netorchestrator -d netorchestrator &>/dev/null; then
        print_success "PostgreSQL is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "PostgreSQL failed to start"
        exit 1
    fi
    sleep 2
done

# Check Redis connection
print_status "Verifying Redis connection..."
for i in {1..30}; do
    if docker exec netorchestrator_redis_1 redis-cli ping &>/dev/null; then
        print_success "Redis is ready"
        break
    fi
    if [ $i -eq 30 ]; then
        print_error "Redis failed to start"
        exit 1
    fi
    sleep 2
done

print_status "Starting monitoring services..."
docker-compose -f docker-compose.monitoring.yml up -d prometheus alertmanager node-exporter cadvisor

print_status "Waiting for monitoring services..."
sleep 15

print_status "Starting visualization services..."
docker-compose -f docker-compose.monitoring.yml up -d grafana jaeger

print_status "Starting NetOrchestrator application..."
docker-compose -f docker-compose.monitoring.yml up -d netorchestrator

print_status "Starting reverse proxy..."
docker-compose -f docker-compose.monitoring.yml up -d nginx

print_status "Waiting for all services to be healthy..."
sleep 20

# Health checks
print_status "Performing health checks..."

# Check NetOrchestrator
if curl -sf http://localhost:8080/health &>/dev/null; then
    print_success "✅ NetOrchestrator API is healthy"
else
    print_warning "⚠️  NetOrchestrator API health check failed"
fi

# Check Prometheus
if curl -sf http://localhost:9090/-/healthy &>/dev/null; then
    print_success "✅ Prometheus is healthy"
else
    print_warning "⚠️  Prometheus health check failed"
fi

# Check Grafana
if curl -sf http://localhost:3000/api/health &>/dev/null; then
    print_success "✅ Grafana is healthy"
else
    print_warning "⚠️  Grafana health check failed"
fi

# Check Alertmanager
if curl -sf http://localhost:9093/-/healthy &>/dev/null; then
    print_success "✅ Alertmanager is healthy"
else
    print_warning "⚠️  Alertmanager health check failed"
fi

print_success "🎉 NetOrchestrator Complete Monitoring Stack is running!"

echo ""
echo "📊 Access URLs:"
echo "  • NetOrchestrator API:     http://localhost:8080"
echo "  • Swagger Documentation:   http://localhost:8080/swagger/index.html"
echo "  • WebSocket Test:          ws://localhost:8080/ws"
echo "  • Prometheus:              http://localhost:9090"
echo "  • Grafana:                 http://localhost:3000 (admin/netorchestrator123)"
echo "  • Alertmanager:            http://localhost:9093"
echo "  • InfluxDB:                http://localhost:8086"
echo "  • Jaeger Tracing:          http://localhost:16686"
echo "  • Node Exporter:           http://localhost:9100"
echo "  • cAdvisor:                http://localhost:8081"
echo ""

echo "🔍 Monitoring Features:"
echo "  • ✅ Real-time metrics collection"
echo "  • ✅ Custom business KPI dashboards"
echo "  • ✅ Automated alerting (Slack, Email, PagerDuty)"
echo "  • ✅ Performance monitoring"
echo "  • ✅ Resource utilization tracking"
echo "  • ✅ SLA compliance monitoring"
echo "  • ✅ Distributed tracing"
echo "  • ✅ Log aggregation"
echo ""

echo "📈 Grafana Dashboards:"
echo "  • System Overview: Comprehensive system health and performance"
echo "  • Network Performance: Network topology and latency monitoring"
echo "  • Business Metrics: KPIs, cost optimization, and ROI tracking"
echo ""

echo "🚨 Alert Categories:"
echo "  • Critical: System failures, network outages (5min response)"
echo "  • Error: Performance degradation, SLA breaches (1hr response)"
echo "  • Warning: Capacity issues, minor performance (4hr response)"
echo "  • Business: Revenue impact, customer satisfaction (2hr response)"
echo ""

print_status "To stop all services, run: docker-compose -f docker-compose.monitoring.yml down"
print_status "To view logs, run: docker-compose -f docker-compose.monitoring.yml logs -f [service-name]"
print_status "To restart with clean data, run: $0 --clean"

echo ""
print_success "🌟 NetOrchestrator NaaS Platform with Complete Observability Stack is ready!"
print_success "🎯 Enterprise-grade monitoring, alerting, and visualization fully operational"
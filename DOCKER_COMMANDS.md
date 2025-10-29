# 🐳 NetOrchestrator Docker Commands Reference

## Essential Docker commands for NetOrchestrator

---

## 🚀 **Quick Setup (One Command)**

```bash
# Complete setup with Docker
./setup-complete.sh
```

## 🐳 **Manual Docker Container Management**

### Start Infrastructure Containers
```bash
# PostgreSQL Database
docker run -d --name postgres \
  -e POSTGRES_DB=netorchestrator \
  -e POSTGRES_USER=netorchestrator \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 postgres:15

# Redis Cache
docker run -d --name redis -p 6379:6379 redis:7-alpine

# InfluxDB Metrics
docker run -d --name influxdb \
  -e DOCKER_INFLUXDB_INIT_MODE=setup \
  -e DOCKER_INFLUXDB_INIT_USERNAME=admin \
  -e DOCKER_INFLUXDB_INIT_PASSWORD=password \
  -e DOCKER_INFLUXDB_INIT_ORG=netorchestrator \
  -e DOCKER_INFLUXDB_INIT_BUCKET=network_metrics \
  -p 8086:8086 influxdb:2.7
```

### Container Status & Management
```bash
# Check running containers
docker ps

# Check all containers (including stopped)
docker ps -a

# View container logs
docker logs postgres
docker logs redis
docker logs influxdb

# Stop containers
docker stop postgres redis influxdb

# Start stopped containers
docker start postgres redis influxdb

# Remove containers completely
docker rm -f postgres redis influxdb

# Clean up everything (nuclear option)
docker container prune -f
docker volume prune -f
```

## 🗄️ **Database Operations**

### Connect to Database
```bash
# PostgreSQL command line
docker exec -it postgres psql -U netorchestrator -d netorchestrator

# Run SQL files
docker exec -i postgres psql -U netorchestrator -d netorchestrator < schema.sql

# Run single SQL commands
docker exec postgres psql -U netorchestrator -d netorchestrator -c "SELECT COUNT(*) FROM networks;"

# Redis command line
docker exec -it redis redis-cli
```

### Database Backup & Restore
```bash
# Backup database
docker exec postgres pg_dump -U netorchestrator netorchestrator > backup.sql

# Restore database
docker exec -i postgres psql -U netorchestrator -d netorchestrator < backup.sql

# Reset database (clean slate)
docker exec postgres psql -U netorchestrator -d netorchestrator -c "
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO netorchestrator;
"
```

## 🛠️ **NetOrchestrator Application**

### Build & Run
```bash
# Build API Gateway
go build -o bin/api-gateway ./cmd/api-gateway/

# Start API Gateway
./bin/api-gateway

# Start in background  
nohup ./bin/api-gateway > api-gateway.log 2>&1 &

# Stop API Gateway
pkill -f api-gateway

# View API logs
tail -f api-gateway.log
```

## 🧪 **Testing Commands**

### Quick Tests
```bash
# Basic functionality test
./quick-test.sh

# Comprehensive AI test
./test-ai-complete.sh

# Health check
curl http://localhost:8080/health

# Authentication test
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}'
```

### Complete API Testing
```bash
# Get authentication token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.access_token')

# Create network
curl -X POST http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Docker Test Network", "description": "Testing with Docker"}'

# AI Network Analysis
curl -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"metrics": {"latency": 45.2, "throughput": 1200}, "time_range": "1h"}'

# AI Capacity Prediction  
curl -X POST http://localhost:8080/api/v1/intelligence/predict/capacity \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "historical_data": [
      {"cpu_usage": 65, "memory_usage": 45, "timestamp": "2023-10-01"},
      {"cpu_usage": 70, "memory_usage": 50, "timestamp": "2023-10-02"},
      {"cpu_usage": 75, "memory_usage": 55, "timestamp": "2023-10-03"}
    ],
    "forecast_days": 30,
    "confidence": 0.95
  }'
```

## 🔄 **Complete Restart Workflow**

### Clean Restart (Fresh Setup)
```bash
# 1. Stop everything
docker stop postgres redis influxdb
pkill -f api-gateway

# 2. Clean up containers
docker rm -f postgres redis influxdb

# 3. Fresh setup
./setup-complete.sh

# 4. Verify
curl http://localhost:8080/health
```

### Quick Restart (Keep Data)
```bash  
# 1. Restart containers
docker restart postgres redis influxdb

# 2. Restart API Gateway
pkill -f api-gateway
./bin/api-gateway &

# 3. Test
./quick-test.sh
```

## 🎯 **Docker Compose Alternative**

If you prefer Docker Compose, you can also use:

```bash
# Using the provided docker-compose.yml
docker-compose up -d

# Or for monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d
```

## 🛠️ **Development Workflow**

### Daily Development
```bash
# Start infrastructure (once per day)
docker start postgres redis influxdb

# Rebuild and restart API Gateway (after code changes)
go build -o bin/api-gateway ./cmd/api-gateway/
pkill -f api-gateway
./bin/api-gateway &

# Test your changes
./quick-test.sh
```

### Debugging
```bash
# Check what's using port 8080
lsof -i :8080

# View API Gateway logs
tail -f api-gateway.log

# Check database connectivity
docker exec postgres psql -U netorchestrator -d netorchestrator -c "SELECT version();"

# Check Redis connectivity  
docker exec redis redis-cli ping
```

---

## 📊 **Container Resource Usage**

```bash
# Monitor Docker resource usage
docker stats

# Check disk usage
docker system df

# Clean up unused resources
docker system prune -f
```

---

## 🎉 **Ready to Go!**

**Your NetOrchestrator platform is now fully configured for Docker!**

All scripts use standard Docker commands that work with:
- ✅ **Docker Desktop** (Mac/Windows)
- ✅ **Docker Engine** (Linux) 
- ✅ **Docker with your Podman alias**
- ✅ **Any Docker-compatible runtime**

**Start developing with confidence!** 🚀

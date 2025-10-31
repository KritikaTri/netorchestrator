# NetOrchestrator API Endpoints - Quick Reference

**Base URL:** `http://localhost:8080`

## 🔧 System Endpoints

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics (Prometheus)
```bash
curl http://localhost:8080/metrics
```

### Swagger Documentation
Open in browser: `http://localhost:8080/swagger/index.html`

## 🔐 Authentication Endpoints

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### Get Profile
```bash
curl http://localhost:8080/api/v1/auth/profile \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Create API Key
```bash
curl -X POST http://localhost:8080/api/v1/auth/api-keys \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My API Key",
    "scopes": ["read", "write"]
  }'
```

## 🌐 Network Management

### List Networks
```bash
curl http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Create Network
```bash
curl -X POST http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-network",
    "subnet": "10.0.0.0/24",
    "description": "My test network"
  }'
```

### Get Network
```bash
curl http://localhost:8080/api/v1/networks/NETWORK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Update Network
```bash
curl -X PUT http://localhost:8080/api/v1/networks/NETWORK_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "updated-network",
    "description": "Updated description"
  }'
```

### Delete Network
```bash
curl -X DELETE http://localhost:8080/api/v1/networks/NETWORK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Start Network
```bash
curl -X POST http://localhost:8080/api/v1/networks/NETWORK_ID/start \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Stop Network
```bash
curl -X POST http://localhost:8080/api/v1/networks/NETWORK_ID/stop \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Restart Network
```bash
curl -X POST http://localhost:8080/api/v1/networks/NETWORK_ID/restart \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🖥️ Node Management

### List All Nodes
```bash
curl http://localhost:8080/api/v1/nodes \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### List Network Nodes
```bash
curl http://localhost:8080/api/v1/networks/NETWORK_ID/nodes \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Create Node
```bash
curl -X POST http://localhost:8080/api/v1/networks/NETWORK_ID/nodes \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "node1",
    "type": "router",
    "ip_address": "10.0.0.1"
  }'
```

### Get Node
```bash
curl http://localhost:8080/api/v1/nodes/NODE_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Update Node
```bash
curl -X PUT http://localhost:8080/api/v1/nodes/NODE_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "updated-node",
    "ip_address": "10.0.0.2"
  }'
```

### Delete Node
```bash
curl -X DELETE http://localhost:8080/api/v1/nodes/NODE_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Start/Stop/Restart Node
```bash
# Start
curl -X POST http://localhost:8080/api/v1/nodes/NODE_ID/start \
  -H "Authorization: Bearer YOUR_TOKEN"

# Stop
curl -X POST http://localhost:8080/api/v1/nodes/NODE_ID/stop \
  -H "Authorization: Bearer YOUR_TOKEN"

# Restart
curl -X POST http://localhost:8080/api/v1/nodes/NODE_ID/restart \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🔗 Link Management

### List Network Links
```bash
curl http://localhost:8080/api/v1/networks/NETWORK_ID/links \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Create Link
```bash
curl -X POST http://localhost:8080/api/v1/networks/NETWORK_ID/links \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "source_node_id": "NODE_ID_1",
    "target_node_id": "NODE_ID_2",
    "bandwidth": "1000"
  }'
```

### Get Link
```bash
curl http://localhost:8080/api/v1/links/LINK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Update Link
```bash
curl -X PUT http://localhost:8080/api/v1/links/LINK_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "bandwidth": "2000"
  }'
```

### Delete Link
```bash
curl -X DELETE http://localhost:8080/api/v1/links/LINK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📊 Monitoring

### Get Network Metrics
```bash
curl http://localhost:8080/api/v1/monitoring/networks/NETWORK_ID/metrics \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Network Health
```bash
curl http://localhost:8080/api/v1/monitoring/networks/NETWORK_ID/health \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Node Metrics
```bash
curl http://localhost:8080/api/v1/monitoring/nodes/NODE_ID/metrics \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Node Health
```bash
curl http://localhost:8080/api/v1/monitoring/nodes/NODE_ID/health \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### List Alerts
```bash
curl http://localhost:8080/api/v1/monitoring/alerts \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 🔒 Policy Management

### List Policies
```bash
curl http://localhost:8080/api/v1/policies \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### List Network Policies
```bash
curl http://localhost:8080/api/v1/networks/NETWORK_ID/policies \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Create Policy
```bash
curl -X POST http://localhost:8080/api/v1/networks/NETWORK_ID/policies \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "allow-http",
    "type": "firewall",
    "rules": {
      "action": "allow",
      "protocol": "tcp",
      "port": 80
    }
  }'
```

## 🧠 AI Intelligence Endpoints

### Analyze Network
```bash
curl -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "network_id": "NETWORK_ID",
    "metrics": {
      "latency": 45.2,
      "throughput": 1200,
      "error_rate": 0.02
    }
  }'
```

### Predict Capacity
```bash
curl -X POST http://localhost:8080/api/v1/intelligence/predict/capacity \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "resource_id": "RESOURCE_ID",
    "forecast_days": 30
  }'
```

### Optimize Costs
```bash
curl -X POST http://localhost:8080/api/v1/intelligence/optimize/costs \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "resources": [...],
    "constraints": ["maintain_performance"]
  }'
```

### Detect Anomalies
```bash
curl -X POST http://localhost:8080/api/v1/intelligence/detect/anomalies \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "metrics": {...},
    "sensitivity": "medium"
  }'
```

### Get ML Models
```bash
curl http://localhost:8080/api/v1/intelligence/models \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get AI Chains
```bash
curl http://localhost:8080/api/v1/intelligence/chains \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Insights
```bash
curl http://localhost:8080/api/v1/intelligence/insights \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📝 Notes

- Replace `YOUR_TOKEN` with actual JWT token from login response
- Replace `NETWORK_ID`, `NODE_ID`, `LINK_ID` with actual UUIDs
- All timestamps are in UTC
- Default port is `8080` (configurable in `config.yaml`)

## 🚀 Quick Start

1. Start the API gateway:
```bash
./bin/api-gateway
```

2. Check health:
```bash
curl http://localhost:8080/health
```

3. Login (once authentication is implemented):
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

4. Create a network:
```bash
curl -X POST http://localhost:8080/api/v1/networks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"test-network","subnet":"10.0.0.0/24"}'
```


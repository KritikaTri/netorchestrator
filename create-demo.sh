#!/bin/bash

# Demo network creation script
echo "🌐 Creating Demo Network Topology..."

API_BASE="http://localhost:8080/api/v1"

# Create a sample network
echo "📡 Creating demo network..."
NETWORK_RESPONSE=$(curl -s -X POST "$API_BASE/networks" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Demo Lab Network",
    "description": "Sample network topology with web servers and database",
    "network_type": "virtual",
    "subnet": "192.168.10.0/24",
    "gateway": "192.168.10.1"
  }')

echo "Network Response: $NETWORK_RESPONSE"

# Extract network ID (would need jq in real scenario)
NETWORK_ID="demo-network-id"

echo "🖥️  Creating demo nodes..."

# Create web server node
curl -s -X POST "$API_BASE/networks/$NETWORK_ID/nodes" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "web-server-1",
    "node_type": "container",
    "ip_address": "192.168.10.10",
    "image": "nginx:alpine",
    "cpu_cores": 2,
    "memory_mb": 512,
    "ports": {"80": "8080"},
    "environment": {"ENV": "demo"}
  }'

# Create database node  
curl -s -X POST "$API_BASE/networks/$NETWORK_ID/nodes" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "database-1",
    "node_type": "container", 
    "ip_address": "192.168.10.20",
    "image": "postgres:13",
    "cpu_cores": 2,
    "memory_mb": 1024,
    "environment": {
      "POSTGRES_DB": "demo",
      "POSTGRES_USER": "demo",
      "POSTGRES_PASSWORD": "demo123"
    }
  }'

echo "✅ Demo network topology created!"
echo "🌐 Access at: http://localhost:8080/api/v1/networks/$NETWORK_ID/topology"
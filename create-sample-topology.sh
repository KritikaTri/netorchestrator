#!/bin/bash

echo "🚀 Creating Sample Network Topology for Testing"
echo "=============================================="

# Get JWT token
echo "🔑 Getting authentication token..."
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin123"}' | jq -r '.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to get authentication token"
    exit 1
fi

echo "✅ Authentication successful"

# Use Production Network ID
NETWORK_ID="7f839bc5-4ef7-46b2-9727-94a7929dfc2b"
echo "🌐 Using Production Network: $NETWORK_ID"

# Create nodes
echo ""
echo "📍 Creating network nodes..."

# Node 1: Router (Central hub)
echo "  Creating router node..."
ROUTER_ID=$(curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/nodes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Central Router",
    "type": "router", 
    "ip_address": "10.0.1.1",
    "status": "active",
    "config": {
      "cpu": 2,
      "memory": 4096,
      "storage": 50
    },
    "position": {
      "x": 400,
      "y": 300
    }
  }' | jq -r '.node.id')

# Node 2: Switch
echo "  Creating switch node..."
SWITCH_ID=$(curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/nodes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Main Switch",
    "type": "switch",
    "ip_address": "10.0.1.2", 
    "status": "active",
    "config": {
      "cpu": 1,
      "memory": 2048,
      "storage": 20
    },
    "position": {
      "x": 200,
      "y": 200
    }
  }' | jq -r '.node.id')

# Node 3: Host 1
echo "  Creating host node 1..."
HOST1_ID=$(curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/nodes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web Server",
    "type": "host",
    "ip_address": "10.0.1.10",
    "status": "active", 
    "config": {
      "cpu": 4,
      "memory": 8192,
      "storage": 100
    },
    "position": {
      "x": 100,
      "y": 400
    }
  }' | jq -r '.node.id')

# Node 4: Host 2  
echo "  Creating host node 2..."
HOST2_ID=$(curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/nodes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Database Server", 
    "type": "host",
    "ip_address": "10.0.1.11",
    "status": "active",
    "config": {
      "cpu": 8,
      "memory": 16384,
      "storage": 500
    },
    "position": {
      "x": 600,
      "y": 400  
    }
  }' | jq -r '.node.id')

# Node 5: Firewall
echo "  Creating firewall node..."
FIREWALL_ID=$(curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/nodes" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Security Firewall",
    "type": "firewall",
    "ip_address": "10.0.1.254",
    "status": "active",
    "config": {
      "cpu": 2,
      "memory": 4096,
      "storage": 50
    },
    "position": {
      "x": 400,
      "y": 100
    }
  }' | jq -r '.node.id')

echo "✅ Created 5 nodes"

# Create links between nodes
echo ""
echo "🔗 Creating network links..."

# Link 1: Router <-> Switch
echo "  Creating router-switch link..."
curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/links" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_node_id\": \"$ROUTER_ID\",
    \"target_node_id\": \"$SWITCH_ID\", 
    \"link_type\": \"ethernet\",
    \"bandwidth_mbps\": 1000,
    \"latency_ms\": 1,
    \"status\": \"active\"
  }" > /dev/null

# Link 2: Switch <-> Web Server
echo "  Creating switch-webserver link..."
curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/links" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_node_id\": \"$SWITCH_ID\",
    \"target_node_id\": \"$HOST1_ID\",
    \"link_type\": \"ethernet\", 
    \"bandwidth_mbps\": 1000,
    \"latency_ms\": 2,
    \"status\": \"active\"
  }" > /dev/null

# Link 3: Router <-> Database Server
echo "  Creating router-database link..."
curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/links" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_node_id\": \"$ROUTER_ID\",
    \"target_node_id\": \"$HOST2_ID\",
    \"link_type\": \"ethernet\",
    \"bandwidth_mbps\": 1000, 
    \"latency_ms\": 1,
    \"status\": \"active\"
  }" > /dev/null

# Link 4: Firewall <-> Router
echo "  Creating firewall-router link..."
curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/links" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_node_id\": \"$FIREWALL_ID\",
    \"target_node_id\": \"$ROUTER_ID\",
    \"link_type\": \"ethernet\",
    \"bandwidth_mbps\": 1000,
    \"latency_ms\": 1, 
    \"status\": \"active\"
  }" > /dev/null

# Link 5: Web Server <-> Database Server (cross-link)
echo "  Creating webserver-database link..."
curl -s -X POST "http://localhost:8080/api/v1/networks/$NETWORK_ID/links" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"source_node_id\": \"$HOST1_ID\",
    \"target_node_id\": \"$HOST2_ID\",
    \"link_type\": \"ethernet\",
    \"bandwidth_mbps\": 100,
    \"latency_ms\": 5,
    \"status\": \"active\"
  }" > /dev/null

echo "✅ Created 5 links"

# Verify the topology
echo ""
echo "🔍 Verifying network topology..."
topology_result=$(curl -s -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/networks/$NETWORK_ID/enhanced-topology")

nodes_count=$(echo "$topology_result" | jq '.topology.nodes | length')
links_count=$(echo "$topology_result" | jq '.topology.links | length') 

echo "📊 Topology Summary:"
echo "  - Nodes: $nodes_count"
echo "  - Links: $links_count"
echo ""

if [ "$nodes_count" -gt 0 ] && [ "$links_count" -gt 0 ]; then
    echo "✅ Sample topology created successfully!"
    echo ""
    echo "🎯 Now refresh your Network Topology page to see the visualization with:"
    echo "   - 5 interconnected nodes (router, switch, hosts, firewall)"
    echo "   - 5 links connecting them in a realistic network topology"
else
    echo "❌ Topology creation may have failed"
    echo "Nodes: $nodes_count, Links: $links_count"
fi

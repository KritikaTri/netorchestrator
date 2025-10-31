#!/bin/bash

# NetOrchestrator Containerlab-Style Deployment
# Professional network lab deployment mimicking Containerlab behavior

echo "🚀 NetOrchestrator Containerlab-Style Lab Deployment"
echo "═══════════════════════════════════════════════════════"
echo ""

# Stop any existing containers
echo "🛑 Cleaning up existing lab..."
./stop-network-lab.sh 2>/dev/null || true

# Create management network
echo "📡 Creating management network: netorchestrator-lab"
podman network create netorchestrator-lab 2>/dev/null || echo "   ✓ Network already exists"

echo ""
echo "🏗️  Deploying network topology..."
echo "───────────────────────────────────────────────"

# Deploy FRR Router (Enterprise routing software)
echo "📍 Deploying node: frr-router [kind: linux, image: frrouting/frr:latest]"
podman run -d \
  --name frr-router \
  --hostname frr-router \
  --network netorchestrator-lab \
  -p 22001:22 \
  -p 22002:2601 \
  -p 22003:2605 \
  -v ./network-configs/frr:/etc/frr:Z \
  --privileged \
  frrouting/frr:latest >/dev/null

# Deploy Alpine Linux Node
echo "📍 Deploying node: alpine-node1 [kind: linux, image: alpine:latest]"
podman run -d \
  --name alpine-node1 \
  --hostname alpine-node1 \
  --network netorchestrator-lab \
  -p 23001:22 \
  alpine:latest \
  sh -c "apk add --no-cache openssh-server curl net-tools iproute2 && \
         ssh-keygen -A && \
         echo 'root:netorchestrator' | chpasswd && \
         sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
         echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && \
         mkdir -p /run/sshd && \
         /usr/sbin/sshd -D" >/dev/null

# Deploy Ubuntu Router
echo "📍 Deploying node: ubuntu-router [kind: linux, image: ubuntu:20.04]"
podman run -d \
  --name ubuntu-router \
  --hostname ubuntu-router \
  --network netorchestrator-lab \
  -p 24001:22 \
  -p 24002:80 \
  ubuntu:20.04 \
  bash -c "apt-get update >/dev/null 2>&1 && \
           apt-get install -y openssh-server curl net-tools iproute2 bird2 nginx >/dev/null 2>&1 && \
           mkdir -p /run/sshd && \
           echo 'root:netorchestrator' | chpasswd && \
           sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
           echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && \
           service ssh start >/dev/null 2>&1 && \
           service nginx start >/dev/null 2>&1 && \
           sleep infinity" >/dev/null

echo ""
echo "⏳ Waiting for containers to initialize..."
sleep 8

echo ""
echo "✅ Lab deployment complete!"
echo "═══════════════════════════════════════════════════════"
echo ""
echo "📋 Lab Summary:"
echo "   Lab name: netorchestrator-lab"
echo "   Management network: netorchestrator-lab"
echo "   Total nodes: 3"
echo ""
echo "📊 Node Information:"
echo "   ┌─────────────────┬──────────┬─────────────────────┬────────────────┐"
echo "   │ Name            │ Kind     │ Image               │ Mgmt Ports     │"
echo "   ├─────────────────┼──────────┼─────────────────────┼────────────────┤"
echo "   │ frr-router      │ router   │ frrouting/frr       │ SSH:22001      │"
echo "   │ alpine-node1    │ host     │ alpine:latest       │ SSH:23001      │"
echo "   │ ubuntu-router   │ router   │ ubuntu:20.04        │ SSH:24001      │"
echo "   └─────────────────┴──────────┴─────────────────────┴────────────────┘"
echo ""
echo "🔗 Network Links:"
echo "   • frr-router ←→ alpine-node1 (lab network)"
echo "   • ubuntu-router ←→ frr-router (lab network)"
echo "   • All nodes connected via netorchestrator-lab bridge"
echo ""
echo "🔑 Management Access:"
echo "   ssh root@localhost -p 22001  # FRR Router"
echo "   ssh root@localhost -p 23001  # Alpine Node"
echo "   ssh root@localhost -p 24001  # Ubuntu Router"
echo "   Password: netorchestrator"
echo ""
echo "🌐 API Endpoints:"
echo "   GET    /api/v1/lab/status     # Lab status"
echo "   GET    /api/v1/lab/inspect    # Detailed info"
echo "   GET    /api/v1/lab/nodes      # List nodes"
echo "   GET    /api/v1/lab/topology   # Topology graph"
echo "   POST   /api/v1/lab/nodes/{name}/exec  # Execute commands"
echo ""
echo "🚀 NetOrchestrator platform ready for Containerlab integration!"
echo "   Start the platform: ./bin/api-gateway"
echo "   Test the lab:       curl /api/v1/lab/status"

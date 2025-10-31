#!/bin/bash

# NetOrchestrator Network Lab Startup Script
# Uses Podman to create realistic network containers for demo

echo "🚀 Starting NetOrchestrator Network Lab..."

# Create network
echo "📡 Creating lab network..."
podman network create netorchestrator-lab 2>/dev/null || echo "Network already exists"

# Start FRRouting container (Open Source Router)
echo "🌐 Starting FRR Router..."
podman run -d \
  --name frr-router \
  --hostname frr-router \
  --network netorchestrator-lab \
  -p 23001:22 \
  -p 23002:2601 \
  -p 23003:2605 \
  -v ./network-configs/frr:/etc/frr:Z \
  --privileged \
  frrouting/frr:latest

# Start Alpine Linux nodes (Simple network devices)
echo "🏔️  Starting Alpine Node..."
podman run -d \
  --name alpine-node \
  --hostname alpine-node \
  --network netorchestrator-lab \
  -p 24001:22 \
  alpine:latest \
  sh -c "apk add --no-cache openssh-server curl net-tools iproute2 && \
         ssh-keygen -A && \
         echo 'root:netorchestrator' | chpasswd && \
         sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
         echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && \
         mkdir -p /run/sshd && \
         /usr/sbin/sshd -D"

# Start Ubuntu-based router (More realistic enterprise node)
echo "🖥️  Starting Ubuntu Router..."
podman run -d \
  --name ubuntu-router \
  --hostname ubuntu-router \
  --network netorchestrator-lab \
  -p 25001:22 \
  -p 25002:80 \
  ubuntu:20.04 \
  bash -c "apt-get update && \
           apt-get install -y openssh-server curl net-tools iproute2 iptables bird2 nginx && \
           mkdir -p /run/sshd && \
           echo 'root:netorchestrator' | chpasswd && \
           sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
           echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && \
           service ssh start && \
           service nginx start && \
           sleep infinity"

# Start Vyos-like container (Network OS simulation)
echo "🔧 Starting VyOS-like Node..."
podman run -d \
  --name vyos-node \
  --hostname vyos-node \
  --network netorchestrator-lab \
  -p 26001:22 \
  -p 26002:443 \
  --privileged \
  alpine:latest \
  sh -c "apk add --no-cache openssh-server curl net-tools iproute2 iptables quagga lighttpd && \
         ssh-keygen -A && \
         echo 'root:vyos' | chpasswd && \
         echo 'vyos:vyos' | chpasswd && \
         adduser -D vyos && \
         echo 'vyos:vyos' | chpasswd && \
         sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config && \
         echo 'PasswordAuthentication yes' >> /etc/ssh/sshd_config && \
         mkdir -p /run/sshd && \
         /usr/sbin/sshd -D"

echo ""
echo "✅ Network Lab Started Successfully!"
echo ""
echo "📊 Available Network Devices:"
echo "  🔸 FRR Router:    SSH=localhost:23001, Telnet=localhost:23002"
echo "  🔸 Alpine Node:   SSH=localhost:24001"  
echo "  🔸 Ubuntu Router: SSH=localhost:25001, HTTP=localhost:25002"
echo "  🔸 VyOS Node:     SSH=localhost:26001, HTTPS=localhost:26002"
echo ""
echo "🔑 Default Credentials:"
echo "  📋 Username: root"
echo "  📋 Password: netorchestrator (vyos for VyOS node)"
echo ""
echo "🧪 Test Connectivity:"
echo "  ssh root@localhost -p 23001  # FRR Router"
echo "  ssh root@localhost -p 24001  # Alpine Node"
echo "  ssh root@localhost -p 25001  # Ubuntu Router"
echo "  ssh root@localhost -p 26001  # VyOS Node"
echo ""
echo "🚀 Ready for NetOrchestrator integration!"

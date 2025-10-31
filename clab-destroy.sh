#!/bin/bash

# NetOrchestrator Containerlab-Style Lab Destruction
# Professional cleanup of network lab

echo "🔥 NetOrchestrator Lab Destruction"
echo "════════════════════════════════════"
echo ""

# Get list of running lab containers
CONTAINERS=(frr-router alpine-node1 ubuntu-router)

echo "🛑 Stopping and removing containers..."
for container in "${CONTAINERS[@]}"; do
    if podman ps -a --format "{{.Names}}" | grep -q "^${container}$"; then
        echo "   📦 Removing: $container"
        podman rm -f "$container" >/dev/null 2>&1
    else
        echo "   ⚠️  Not found: $container"
    fi
done

echo ""
echo "🌐 Removing management network..."
if podman network ls --format "{{.Name}}" | grep -q "^netorchestrator-lab$"; then
    podman network rm netorchestrator-lab >/dev/null 2>&1
    echo "   ✓ Network netorchestrator-lab removed"
else
    echo "   ⚠️  Network not found"
fi

echo ""
echo "🧹 Cleanup Summary:"
echo "   • All lab containers stopped and removed"
echo "   • Management network destroyed"
echo "   • Lab topology reset"
echo ""
echo "✅ Lab destruction complete!"
echo "💡 Run ./clab-deploy.sh to redeploy the lab"

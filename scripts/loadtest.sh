#!/bin/bash
# Reproducible local load test: 3 topology-svc replicas behind the gateway.
# Usage: scripts/loadtest.sh        (needs only Go; binaries go to ./.loadtest-bin)
# Numbers depend on your hardware: on a small box the load generator, gateway
# and replicas all share the same cores, so this shows behaviour (latency,
# failover), not horizontal scale-out. Run replicas on separate hosts for that.
set -u
BIN=./.loadtest-bin
mkdir -p "$BIN"
go build -o "$BIN"/ ./cmd/topology-svc ./cmd/gateway ./cmd/loadgen || exit 1

export GATEWAY_JWT_SECRET=$(head -c 24 /dev/urandom | base64)
pids=()
cleanup(){ kill "${pids[@]}" 2>/dev/null; wait 2>/dev/null; }
trap cleanup EXIT
start_topo(){ ADDR=:$1 TOPOLOGY_CACHE_ENTRIES=0 "$BIN"/topology-svc 2>/dev/null & pids+=($!); }
for p in 18081 18082 18083; do start_topo $p; done
GATEWAY_TOPOLOGY_UPSTREAMS=http://localhost:18081,http://localhost:18082,http://localhost:18083 \
GATEWAY_RATE_RPS=1000000 GATEWAY_RATE_BURST=1000000 ADDR=:18080 "$BIN"/gateway 2>/dev/null & pids+=($!)
sleep 1

echo "== A: 1 replica direct, no cache, 100 nodes, 5 clients"
"$BIN"/loadgen -url http://localhost:18081 -nodes 100 -variants 50 -c 5 -d 6s
echo; echo "== B: via gateway -> 3 replicas, no cache, 100 nodes, 5 clients"
"$BIN"/loadgen -url http://localhost:18080 -nodes 100 -variants 50 -c 5 -d 6s
echo; echo "== C: 1 replica direct, no cache, 1000 nodes"
"$BIN"/loadgen -url http://localhost:18081 -nodes 1000 -variants 20 -c 5 -d 6s
echo; echo "== D: 10000 nodes, preflight, direct"
"$BIN"/loadgen -url http://localhost:18081 -endpoint /v1/analyze/preflight -nodes 10000 -variants 3 -c 4 -d 6s
echo; echo "== E: failover: kill replica 2 mid-run (via gateway, 100 nodes)"
( sleep 3; kill "${pids[1]}" ) &
"$BIN"/loadgen -url http://localhost:18080 -nodes 100 -variants 50 -c 5 -d 8s

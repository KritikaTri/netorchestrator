#!/bin/bash
# End-to-end demo of the change workflow through the gateway:
#   client -> gateway (auth, rate limit) -> change-svc -> topology-svc (preflight)
# Needs only Go and curl. Binaries go to ./.loadtest-bin, audit log to a temp file.
set -eu
BIN=./.loadtest-bin; mkdir -p "$BIN"
go build -o "$BIN"/ ./cmd/topology-svc ./cmd/gateway ./cmd/change-svc ./cmd/devtoken
LOG=$(mktemp)
export GATEWAY_JWT_SECRET=$(head -c 24 /dev/urandom | base64)
pids=(); trap 'kill "${pids[@]}" 2>/dev/null; wait 2>/dev/null' EXIT
ADDR=:19081 "$BIN"/topology-svc 2>/dev/null & pids+=($!)
ADDR=:19082 CHANGE_LOG_PATH=$LOG CHANGE_TOPOLOGY_URL=http://localhost:19081 "$BIN"/change-svc 2>/dev/null & pids+=($!)
ADDR=:19080 GATEWAY_TOPOLOGY_UPSTREAMS=http://localhost:19081 GATEWAY_CHANGE_UPSTREAMS=http://localhost:19082 "$BIN"/gateway 2>/dev/null & pids+=($!)
sleep 1
tok(){ "$BIN"/devtoken -sub "$1"; }
api(){ local who=$1 method=$2 path=$3; shift 3; curl -s -X "$method" "http://localhost:19080$path" -H "Authorization: Bearer $(tok "$who")" -H 'Content-Type: application/json' "$@"; }

TOPO='"nodes":[{"id":"core","name":"core-sw"},{"id":"a1"},{"id":"a2"},{"id":"a3"}],"links":[{"id":"l1","source":"core","target":"a1"},{"id":"l2","source":"core","target":"a2"},{"id":"l3","source":"core","target":"a3"}]'
echo "1) alice proposes taking down core-sw (preflight runs inside change-svc via topology-svc)"
RESP=$(api alice POST /v1/changes -d "{\"title\":\"replace core-sw\",\"change\":{\"kind\":\"remove_node\",\"id\":\"core\"},$TOPO}")
echo "$RESP" | head -c 400; echo
ID=$(echo "$RESP" | sed -n 's/.*"id":"\(chg_[a-f0-9]*\)".*/\1/p' | head -1)
echo; echo "2) alice tries to approve her own change"; api alice POST /v1/changes/$ID/approve
echo; echo "3) bob approves (critical risk needs two reviewers)"; api bob POST /v1/changes/$ID/approve | head -c 200; echo
echo; echo "4) apply before quorum is refused"; api bob POST /v1/changes/$ID/apply
echo; echo "5) carol approves, then apply"; api carol POST /v1/changes/$ID/approve >/dev/null; api carol POST /v1/changes/$ID/apply | head -c 200; echo
echo; echo "6) audit chain verification"; api auditor GET /v1/audit/verify
echo; echo "7) unauthenticated request is rejected at the gateway"; curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:19080/v1/changes

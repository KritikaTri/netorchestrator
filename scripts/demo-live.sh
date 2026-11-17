#!/usr/bin/env bash
# Runs the approval + audit story against a DEPLOYED stack (see deploy/gce/).
# From Cloud Shell:  ./scripts/demo-live.sh https://<ip>.sslip.io
# Change routes are limited to 1 rps (burst 5) on the public demo, hence the sleeps.
set -euo pipefail
H="${1:?usage: $0 https://<host>}"
HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
tok(){ "$HERE/../deploy/gce/deploy.sh" token "$1" 2h | tail -1; }
A=$(tok alice); B=$(tok bob); C=$(tok carol); U=$(tok auditor)
api(){ local t=$1 m=$2 p=$3; shift 3; sleep 1.2; curl -s -X "$m" "$H$p" -H "Authorization: Bearer $t" -H 'Content-Type: application/json' "$@"; echo; }

TOPO='"nodes":[{"id":"core","name":"core-sw"},{"id":"a1"},{"id":"a2"},{"id":"a3"}],"links":[{"id":"l1","source":"core","target":"a1"},{"id":"l2","source":"core","target":"a2"},{"id":"l3","source":"core","target":"a3"}]'

echo "0) no token is rejected at the gateway"; curl -s -o /dev/null -w "HTTP %{http_code}\n" "$H/v1/changes"
echo "1) alice proposes removing core-sw (preflight runs inside change-svc)"
RESP=$(api "$A" POST /v1/changes -d "{\"title\":\"replace core-sw\",\"change\":{\"kind\":\"remove_node\",\"id\":\"core\"},$TOPO}")
echo "$RESP" | head -c 300; echo
ID=$(echo "$RESP" | sed -n 's/.*"id":"\(chg_[a-f0-9]*\)".*/\1/p' | head -1)
echo "2) alice approves her own change (refused)"; api "$A" POST "/v1/changes/$ID/approve"
echo "3) bob approves (critical risk needs two reviewers)"; api "$B" POST "/v1/changes/$ID/approve" | head -c 160; echo
echo "4) apply before quorum (refused)"; api "$B" POST "/v1/changes/$ID/apply"
echo "5) carol approves, then apply"; api "$C" POST "/v1/changes/$ID/approve" >/dev/null; api "$C" POST "/v1/changes/$ID/apply" | head -c 160; echo
echo "6) audit hash chain verifies"; api "$U" GET /v1/audit/verify

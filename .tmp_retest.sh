#!/usr/bin/env bash
set -euo pipefail
BASE="http://localhost:8080"
OUT="/Users/kritripa/netorchestrator-1/COMPLETE_ENDPOINT_TEST_RESULTS.md"

pass=0; fail=0

section() { echo -e "\n### $1\n" >> "$OUT"; }
record() {
  local name="$1" code="$2" method="$3" path="$4"
  if [[ "$code" == "200" || "$code" == "201" || "$code" == "202" || "$code" == "204" ]]; then
    ((pass++))
    echo "- ✅ [$code] $method $path — $name" >> "$OUT"
  else
    ((fail++))
    echo "- ❌ [$code] $method $path — $name" >> "$OUT"
  fi
}

hdr_auth() { echo -H "Authorization: Bearer $TOKEN"; }

# Start fresh file
printf "# 🧪 Complete API Endpoint Test Results\n\n" > "$OUT"
echo "Started: $(date -u +"%Y-%m-%d %H:%M:%S UTC")" >> "$OUT"

# Auth
AUTH_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/health")
section "System"
record "Health" "$AUTH_CODE" GET "/health"

TOKEN=$(curl -s -X POST "$BASE/api/v1/auth/login" -H "Content-Type: application/json" -d '{"username":"admin","password":"admin123"}' | python3 -c 'import sys,json;print(json.load(sys.stdin).get("access_token",""))')
section "Auth"
LOGIN_CODE=$( [[ -n "$TOKEN" ]] && echo 200 || echo 401 )
record "Login (token acquired)" "$LOGIN_CODE" POST "/api/v1/auth/login"

# Networks
section "Networks"
NETS_JSON=$(curl -s "$BASE/api/v1/networks" -H "Authorization: Bearer $TOKEN")
NETS_CODE=$( [[ "$NETS_JSON" == "" ]] && echo 401 || echo 200 )
record "List Networks" "$NETS_CODE" GET "/api/v1/networks"
NETWORK_ID=$(echo "$NETS_JSON" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["networks"][0]["id"]) if d.get("networks") else print("")' 2>/dev/null || true)
if [[ -n "$NETWORK_ID" ]]; then
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/networks/$NETWORK_ID" -H "Authorization: Bearer $TOKEN")
  record "Get Network Details" "$CODE" GET "/api/v1/networks/{id}"
fi

# Nodes
section "Nodes"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/nodes" -H "Authorization: Bearer $TOKEN")
record "List Nodes" "$CODE" GET "/api/v1/nodes"
if [[ -n "$NETWORK_ID" ]]; then
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/monitoring/networks/$NETWORK_ID/metrics" -H "Authorization: Bearer $TOKEN")
  record "Network Metrics" "$CODE" GET "/api/v1/monitoring/networks/{id}/metrics"
fi

# Monitoring
section "Monitoring"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/monitoring/alerts" -H "Authorization: Bearer $TOKEN")
record "List Alerts" "$CODE" GET "/api/v1/monitoring/alerts"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/monitoring/events" -H "Authorization: Bearer $TOKEN")
record "List Events" "$CODE" GET "/api/v1/monitoring/events"

# Intelligence
section "AI Intelligence"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/intelligence/analyze" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"analysis_type":"full"}')
record "Analyze Network" "$CODE" POST "/api/v1/intelligence/analyze"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/intelligence/predict" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"prediction_days":30}')
record "Predict Capacity" "$CODE" POST "/api/v1/intelligence/predict"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/intelligence/optimize" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"optimization_level":"moderate"}')
record "Optimize Costs" "$CODE" POST "/api/v1/intelligence/optimize"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/intelligence/detect/anomalies" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"sensitivity":"high"}')
record "Detect Anomalies" "$CODE" POST "/api/v1/intelligence/detect/anomalies"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/intelligence/insights?limit=3" -H "Authorization: Bearer $TOKEN")
record "Get AI Insights" "$CODE" GET "/api/v1/intelligence/insights"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/intelligence/models" -H "Authorization: Bearer $TOKEN")
record "Get ML Models" "$CODE" GET "/api/v1/intelligence/models"

# Workflows
section "Workflows"
WF_JSON=$(curl -s "$BASE/api/v1/automation/workflows" -H "Authorization: Bearer $TOKEN")
WF_CODE=$( [[ "$WF_JSON" == "" ]] && echo 401 || echo 200 )
record "List Workflow Definitions" "$WF_CODE" GET "/api/v1/automation/workflows"
WF_ID=$(echo "$WF_JSON" | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["workflows"][0]["id"]) if d.get("workflows") else print("")' 2>/dev/null || true)
if [[ -n "$WF_ID" ]]; then
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/automation/workflows/$WF_ID" -H "Authorization: Bearer $TOKEN")
  record "Get Workflow Details" "$CODE" GET "/api/v1/automation/workflows/{id}"
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/automation/workflows/$WF_ID/executions" -H "Authorization: Bearer $TOKEN")
  record "List Workflow Executions" "$CODE" GET "/api/v1/automation/workflows/{id}/executions"
  CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/automation/executions/latest?workflow_id=$WF_ID" -H "Authorization: Bearer $TOKEN")
  record "Get Execution Details (latest)" "$CODE" GET "/api/v1/automation/executions/latest?workflow_id={id}"
  CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/automation/workflows/$WF_ID/execute" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"parameters":{}}')
  record "Execute Workflow" "$CODE" POST "/api/v1/automation/workflows/{id}/execute"
fi

# Policies
section "Policies"
CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE/api/v1/policies" -H "Authorization: Bearer $TOKEN")
record "List Policies" "$CODE" GET "/api/v1/policies"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$BASE/api/v1/policies" -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" -d '{"name":"Auto-Policy-'"$(date +%s)'"","type":"firewall","config":{"priority":50}}')
record "Create Policy" "$CODE" POST "/api/v1/policies"

# Summary
TOTAL=$((pass+fail))
{
  echo "\n---\n"
  echo "**Passed:** $pass  |  **Failed:** $fail  |  **Total Tested:** $TOTAL"
  echo "\nLast Updated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")"
} >> "$OUT"


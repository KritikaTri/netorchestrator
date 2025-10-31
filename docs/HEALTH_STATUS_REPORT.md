# Docker Compose Health Status Report

**Date:** 2025-10-31  
**Compose File:** `docker-compose.monitoring.yml`

---

## Container Health Summary

### ✅ Healthy Containers

| Container | Service | Status | Health | Notes |
|-----------|---------|--------|--------|-------|
| `postgres-1` | postgres | Up 3 hours | ✅ healthy | Database running correctly |
| `redis-1` | redis | Up 2 hours | ✅ healthy | Cache service operational |
| `nats-1` | nats | Up 2 hours | ✅ healthy | Message bus active |
| `cadvisor-1` | cadvisor | Up 3 minutes | ✅ healthy | Container metrics |
| `alertmanager-1` | alertmanager | Up 8 seconds | ✅ healthy | **FIXED** - Config corrected |

### ⚠️ Containers Without Health Checks (Running)

| Container | Service | Status | Notes |
|-----------|---------|--------|-------|
| `influxdb-1` | influxdb | Up 3 minutes | Running (no healthcheck defined) |
| `jaeger-1` | jaeger | Up 3 minutes | Running (no healthcheck defined) |
| `node-exporter-1` | node-exporter | Up 3 minutes | Running (no healthcheck defined) |

### ❌ Not Started Containers

| Container | Service | Status | Issue |
|-----------|---------|--------|-------|
| `prometheus` | prometheus | Not started | Depends on `netorchestrator` which can't start |
| `grafana` | grafana | Not started | Depends on `prometheus` |
| `netorchestrator` | netorchestrator | Failed to start | Port 8080 already in use |

---

## Unhealthy Container Analysis

### 1. Alertmanager (FIXED ✅)

**Previous Status:** Exiting repeatedly with config errors  
**Current Status:** ✅ Healthy (Up 8 seconds)

#### Root Cause
The Alertmanager configuration file (`deployments/alertmanager/alertmanager.yml`) was using invalid field names:
- Used `subject:` directly in `email_configs` (not supported in Alertmanager 0.26.0)
- Used `body:` instead of `html:` or `text:`

#### Exact Fix Applied
**File:** `deployments/alertmanager/alertmanager.yml`

**Changed:**
```yaml
# BEFORE (Invalid)
email_configs:
  - to: 'ops-team@company.com'
    subject: '🚨 CRITICAL: {{ .GroupLabels.alertname }}'
    body: |
      ...

# AFTER (Valid)
email_configs:
  - to: 'ops-team@company.com'
    headers:
      Subject: '🚨 CRITICAL: {{ .GroupLabels.alertname }}'
    html: |
      ...
```

**Applied to:**
- `critical-alerts` receiver (line 44-45)
- `error-alerts` receiver (line 76-77)
- `warning-alerts` receiver (line 101-102)
- `business-alerts` receiver (line 128-129)

#### Verification
```bash
docker compose -f docker-compose.monitoring.yml logs alertmanager --tail 5
# Should show: "Listening on" and "TLS is disabled" (no config errors)
```

---

### 2. NetOrchestrator (Not Started)

**Status:** ❌ Failed to start

#### Root Cause
Port 8080 is already bound by another process (likely a local API gateway instance).

**Error Message:**
```
Error response from daemon: something went wrong with the request: "listen tcp :8080: bind: address already in use"
```

#### Exact Fix Required

**Option 1: Stop the conflicting process (Recommended)**
```bash
# Find process using port 8080
lsof -i :8080
# or
netstat -an | grep 8080

# Kill the process (replace PID with actual process ID)
kill -9 <PID>
```

**Option 2: Change port mapping**
**File:** `docker-compose.monitoring.yml`

**Change:**
```yaml
# BEFORE
netorchestrator:
  ports:
    - "8080:8080"

# AFTER
netorchestrator:
  ports:
    - "8081:8080"  # Use different host port
```

**Also update:**
- `config.yaml`: Change `server.port` if needed
- `.env`: Update `SERVER_PORT=8081` if changed
- Any service dependencies that connect to `http://netorchestrator:8080`

---

### 3. Prometheus (Not Started)

**Status:** ❌ Not started

#### Root Cause
Prometheus depends on `netorchestrator` service:
```yaml
depends_on:
  - netorchestrator
```
Since `netorchestrator` failed to start, Prometheus cannot start.

#### Exact Fix Required
Once `netorchestrator` is fixed (port 8080 issue resolved), Prometheus will start automatically:
```bash
docker compose -f docker-compose.monitoring.yml up -d prometheus
```

---

### 4. Grafana (Not Started)

**Status:** ❌ Not started

#### Root Cause
Grafana depends on `prometheus` service:
```yaml
depends_on:
  - prometheus
```
Since Prometheus depends on `netorchestrator`, and that failed, Grafana cannot start.

#### Exact Fix Required
After fixing `netorchestrator` and `prometheus`, Grafana will start automatically:
```bash
docker compose -f docker-compose.monitoring.yml up -d grafana
```

---

## Additional Issues

### Dockerfile Build Issue (Fixed)

**Issue:** Docker build was failing due to architecture mismatch (trying to build `linux/amd64` on `linux/arm64`).

**Fix Applied:**
**File:** `Dockerfile`

**Changed:**
```dockerfile
# BEFORE
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/api-gateway ./cmd/api-gateway

# AFTER
RUN CGO_ENABLED=0 go build -o /app/bin/api-gateway ./cmd/api-gateway
```

This now builds for the native architecture (ARM64) which matches the container runtime.

---

## Summary of Fixes

1. ✅ **Alertmanager** - Fixed config YAML (subject/body → headers.Subject/html)
2. ✅ **Dockerfile** - Fixed architecture mismatch (removed GOARCH=amd64)
3. ⚠️ **NetOrchestrator** - Port 8080 conflict (requires manual intervention)
4. ⚠️ **Prometheus** - Waiting on netorchestrator fix
5. ⚠️ **Grafana** - Waiting on prometheus fix

---

## Recommended Next Steps

1. **Free port 8080** or change port mapping for netorchestrator
2. **Restart netorchestrator**: `docker compose -f docker-compose.monitoring.yml up -d netorchestrator`
3. **Start Prometheus**: Will auto-start once netorchestrator is up
4. **Start Grafana**: Will auto-start once prometheus is up
5. **Verify all services**: `docker compose -f docker-compose.monitoring.yml ps`

---

## Commands to Check Status

```bash
# Full status
docker compose -f docker-compose.monitoring.yml ps

# Health checks only
docker compose -f docker-compose.monitoring.yml ps --format "table {{.Service}}\t{{.Health}}"

# Container logs
docker compose -f docker-compose.monitoring.yml logs <service-name>

# Port usage
lsof -i :8080
```

---

**Report Generated:** 2025-10-31 05:17 IST


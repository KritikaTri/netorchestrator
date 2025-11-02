# NetOrchestrator Observability Setup

Production-grade observability with Prometheus metrics and Grafana dashboards.

## Overview

This observability stack provides comprehensive monitoring for NetOrchestrator including:

- **Infrastructure Metrics**: Networks, nodes, containers, links, policies
- **Provisioning Performance**: Duration histograms, operation counters
- **Policy Enforcement**: iptables/tc rules, enforcement rates
- **Container Operations**: Lifecycle tracking, Podman errors
- **API Performance**: HTTP request rates, duration quantiles
- **Database Performance**: Query durations and counts

## Architecture

```
┌─────────────────┐
│ NetOrchestrator │
│   API Gateway   │──────┐
│   (Port 8080)   │      │
└─────────────────┘      │
                         │ /metrics scrape (15s)
                         ▼
                  ┌──────────────┐
                  │  Prometheus  │
                  │ (Port 9090)  │
                  └──────┬───────┘
                         │
                         │ Query metrics
                         ▼
                  ┌──────────────┐
                  │   Grafana    │
                  │ (Port 3000)  │
                  └──────────────┘
```

## Metrics Design

### Infrastructure Metrics (Gauges)
| Metric | Labels | Description |
|--------|--------|-------------|
| `netorch_networks_total` | `status` | Total networks by status (active/provisioning/failed) |
| `netorch_nodes_total` | `node_type`, `status` | Total nodes by type and status |
| `netorch_containers_active` | `node_type` | Active containers by node type |
| `netorch_links_total` | `status` | Total links by status |
| `netorch_policies_total` | `policy_type`, `status` | Total policies by type and status |

### Provisioning Metrics
| Metric | Labels | Description |
|--------|--------|-------------|
| `netorch_provision_duration_seconds` | `operation` | Histogram of provisioning durations (network/node/link) |
| `netorch_provision_operations_total` | `operation`, `result` | Counter of provisioning operations (success/failure) |

### Policy Enforcement Metrics
| Metric | Labels | Description |
|--------|--------|-------------|
| `netorch_policy_enforcement_total` | `policy_type`, `result` | Counter of policy enforcements |
| `netorch_iptables_rules_active` | `network` | Active iptables rules per network |
| `netorch_tc_rules_active` | `network` | Active tc (traffic control) rules per network |

### Container Operations Metrics
| Metric | Labels | Description |
|--------|--------|-------------|
| `netorch_container_operations_total` | `operation`, `result` | Container lifecycle operations (create/start/stop/delete) |
| `netorch_podman_errors_total` | `operation` | Podman command failures by operation type |

### API Metrics
| Metric | Labels | Description |
|--------|--------|-------------|
| `netorch_http_requests_total` | `method`, `path`, `status_code` | HTTP request counter |
| `netorch_http_request_duration_seconds` | `method`, `path` | HTTP request duration histogram |

### Database Metrics
| Metric | Labels | Description |
|--------|--------|-------------|
| `netorch_db_queries_total` | `operation`, `result` | Database query counter |
| `netorch_db_query_duration_seconds` | `operation` | Database query duration histogram |

## Quick Start

### 1. Start Observability Stack

```bash
# Start Prometheus + Grafana
docker-compose -f docker-compose.observability.yml up -d

# Verify services
docker ps | grep netorch
```

### 2. Start NetOrchestrator API

```bash
# Build and start API (exposes /metrics endpoint)
go build -o api-gateway cmd/api-gateway/main.go
./api-gateway
```

### 3. Access Dashboards

- **Grafana**: http://localhost:3000
  - Username: `admin`
  - Password: `netorch-admin`
  - Dashboard: "NetOrchestrator - Infrastructure & Performance"

- **Prometheus**: http://localhost:9090
  - Query metrics directly
  - Check scrape targets: http://localhost:9090/targets

- **NetOrchestrator /metrics**: http://localhost:8080/metrics
  - Raw Prometheus metrics endpoint

## Dashboard Panels

The main Grafana dashboard includes:

1. **Active Containers by Node Type** - Time series showing router, switch, server, firewall containers
2. **Infrastructure Overview** - Bar gauge of networks and nodes by status
3. **Provisioning Duration Heatmap** - Histogram heatmap of operation durations
4. **Policy Enforcement Rate by Type** - Success/failure rates per policy type
5. **Active Policy Rules per Network** - Stacked iptables and tc rules
6. **HTTP Request Duration (p95 & p50)** - API latency quantiles
7. **HTTP Request Rate by Status** - Request throughput by status code
8. **Links & Policies Overview** - Gauge showing current counts
9. **Error Rates** - Podman and provisioning failures

## Metrics Collection

### Automatic Collection (via Middleware)

HTTP requests are automatically instrumented via Prometheus middleware:

```go
router.Use(observability.PrometheusMiddleware(metrics))
```

### Periodic Collection (via Collector)

Infrastructure metrics are updated every 15 seconds by polling the database:

```go
metricsCollector := observability.NewMetricsCollector(db, metrics, logger)
metricsCollector.Start(ctx, 15*time.Second)
```

### Inline Instrumentation

Provisioning operations record timing and errors:

```go
startTime := time.Now()
// ... perform operation ...
metrics.RecordProvisioningOperation("network", startTime, err)
```

## Configuration

### Prometheus Configuration

Located at: `deployments/prometheus/prometheus.yml`

```yaml
scrape_configs:
  - job_name: 'netorchestrator-api'
    static_configs:
      - targets: ['host.docker.internal:8080']
    scrape_interval: 15s
```

### Grafana Datasource

Located at: `deployments/grafana/provisioning/datasources/prometheus.yml`

Automatically configured to use Prometheus at `http://prometheus:9090`.

## Histogram Buckets

Carefully tuned for each operation type:

- **Provisioning**: `[0.1, 0.5, 1, 2, 5, 10, 30]` seconds
  - Covers container creation, network setup, policy application
- **HTTP Requests**: `[0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1]` seconds
  - Typical API response times
- **Database Queries**: `[0.001, 0.005, 0.01, 0.025, 0.05, 0.1]` seconds
  - Fast queries with proper indexing

## Example Queries

### Top 10 Slowest Endpoints (p95)

```promql
topk(10, histogram_quantile(0.95, 
  rate(netorch_http_request_duration_seconds_bucket[5m])
))
```

### Provisioning Success Rate

```promql
rate(netorch_provision_operations_total{result="success"}[5m]) / 
rate(netorch_provision_operations_total[5m])
```

### Container Error Rate

```promql
rate(netorch_podman_errors_total[5m])
```

### Active Infrastructure

```promql
sum(netorch_containers_active) by (node_type)
sum(netorch_networks_total) by (status)
```

## Troubleshooting

### Metrics Not Appearing

1. Check API is running: `curl http://localhost:8080/health`
2. Check /metrics endpoint: `curl http://localhost:8080/metrics`
3. Check Prometheus targets: http://localhost:9090/targets
4. Verify scrape config uses `host.docker.internal` (Docker Desktop)

### Collector Not Updating

Check logs for database connection issues:

```bash
# API logs
tail -f api-gateway.log

# Check for "Metrics collector started" message
```

### Dashboard Not Loading

1. Verify Grafana is running: `docker ps | grep grafana`
2. Check datasource: Grafana → Configuration → Data Sources → Prometheus
3. Test query in Explore tab: `netorch_networks_total`

## Production Recommendations

1. **Enable Alerting**: Add Alertmanager for threshold-based alerts
2. **Add Recording Rules**: Pre-compute expensive queries
3. **Increase Retention**: Default is 15 days, increase for long-term trends
4. **Secure Grafana**: Change default password, enable TLS
5. **Add Authentication**: Protect /metrics endpoint with API key middleware
6. **Tune Scrape Interval**: Balance resolution vs storage (15s is good default)
7. **Add Remote Storage**: Send metrics to Thanos/Cortex for long-term retention

## Files

- `docker-compose.observability.yml` - Docker Compose for Prometheus + Grafana
- `deployments/prometheus/prometheus.yml` - Prometheus scrape configuration
- `deployments/grafana/provisioning/datasources/prometheus.yml` - Grafana datasource
- `deployments/grafana/provisioning/dashboards/netorchestrator.yml` - Dashboard provisioning
- `deployments/grafana/dashboards/netorchestrator-main.json` - Main dashboard JSON
- `internal/observability/metrics.go` - Metrics definitions and helpers
- `internal/observability/collector.go` - Periodic metrics collector
- `internal/observability/middleware.go` - HTTP request middleware

## Next Steps

1. Create network via API: `POST /api/v1/networks`
2. Watch metrics appear in Grafana dashboard
3. Provision nodes and links
4. Observe provisioning duration histogram
5. Apply policies and track enforcement
6. Monitor iptables/tc rules per network
7. Analyze API performance and error rates

Happy monitoring! 🚀

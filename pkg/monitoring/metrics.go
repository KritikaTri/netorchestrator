package monitoring

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsCollector contains all Prometheus metrics for NetOrchestrator
type MetricsCollector struct {
	// HTTP Metrics
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Network Metrics
	NetworksTotal     prometheus.Gauge
	NetworksActive    prometheus.Gauge
	NodesTotal        prometheus.Gauge
	NodesActive       prometheus.Gauge
	LinksTotal        prometheus.Gauge
	LinksActive       prometheus.Gauge
	PoliciesTotal     prometheus.Gauge
	PoliciesActive    prometheus.Gauge

	// System Metrics
	WebSocketConnections prometheus.Gauge
	DatabaseConnections  prometheus.Gauge
	CacheHitRate        prometheus.Gauge
	EventsProcessed     *prometheus.CounterVec
	AlertsTriggered     *prometheus.CounterVec

	// Performance Metrics
	NetworkLatency      *prometheus.HistogramVec
	NetworkThroughput   *prometheus.GaugeVec
	NodeCPUUsage        *prometheus.GaugeVec
	NodeMemoryUsage     *prometheus.GaugeVec
	NodeDiskUsage       *prometheus.GaugeVec

	// Business Metrics
	NetworkProvisioningTime *prometheus.HistogramVec
	PolicyViolations        *prometheus.CounterVec
	SLACompliance          *prometheus.GaugeVec
}

// NewMetricsCollector creates a new metrics collector with all Prometheus metrics
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		// HTTP Metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorchestrator_http_requests_total",
				Help: "Total number of HTTP requests processed",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorchestrator_http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
		),

		// Network Metrics
		NetworksTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_networks_total",
				Help: "Total number of networks",
			},
		),
		NetworksActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_networks_active",
				Help: "Number of active networks",
			},
		),
		NodesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_nodes_total",
				Help: "Total number of nodes",
			},
		),
		NodesActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_nodes_active",
				Help: "Number of active nodes",
			},
		),
		LinksTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_links_total",
				Help: "Total number of links",
			},
		),
		LinksActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_links_active",
				Help: "Number of active links",
			},
		),
		PoliciesTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_policies_total",
				Help: "Total number of policies",
			},
		),
		PoliciesActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_policies_active",
				Help: "Number of active policies",
			},
		),

		// System Metrics
		WebSocketConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_websocket_connections",
				Help: "Number of active WebSocket connections",
			},
		),
		DatabaseConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_database_connections",
				Help: "Number of active database connections",
			},
		),
		CacheHitRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "netorchestrator_cache_hit_rate",
				Help: "Cache hit rate percentage",
			},
		),
		EventsProcessed: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorchestrator_events_processed_total",
				Help: "Total number of events processed",
			},
			[]string{"event_type", "status"},
		),
		AlertsTriggered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorchestrator_alerts_triggered_total",
				Help: "Total number of alerts triggered",
			},
			[]string{"severity", "source"},
		),

		// Performance Metrics
		NetworkLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorchestrator_network_latency_seconds",
				Help:    "Network latency in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
			},
			[]string{"network_id", "source", "destination"},
		),
		NetworkThroughput: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorchestrator_network_throughput_mbps",
				Help: "Network throughput in Mbps",
			},
			[]string{"network_id", "direction"},
		),
		NodeCPUUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorchestrator_node_cpu_usage_percent",
				Help: "Node CPU usage percentage",
			},
			[]string{"node_id", "network_id"},
		),
		NodeMemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorchestrator_node_memory_usage_percent",
				Help: "Node memory usage percentage",
			},
			[]string{"node_id", "network_id"},
		),
		NodeDiskUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorchestrator_node_disk_usage_percent",
				Help: "Node disk usage percentage",
			},
			[]string{"node_id", "network_id"},
		),

		// Business Metrics
		NetworkProvisioningTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorchestrator_network_provisioning_duration_seconds",
				Help:    "Time taken to provision a network in seconds",
				Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1200, 3600},
			},
			[]string{"network_type", "status"},
		),
		PolicyViolations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorchestrator_policy_violations_total",
				Help: "Total number of policy violations",
			},
			[]string{"policy_type", "severity", "network_id"},
		),
		SLACompliance: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorchestrator_sla_compliance_percent",
				Help: "SLA compliance percentage",
			},
			[]string{"service", "sla_type"},
		),
	}
}

// RecordHTTPRequest records an HTTP request
func (m *MetricsCollector) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// UpdateNetworkCounts updates network-related counts
func (m *MetricsCollector) UpdateNetworkCounts(total, active int) {
	m.NetworksTotal.Set(float64(total))
	m.NetworksActive.Set(float64(active))
}

// UpdateNodeCounts updates node-related counts
func (m *MetricsCollector) UpdateNodeCounts(total, active int) {
	m.NodesTotal.Set(float64(total))
	m.NodesActive.Set(float64(active))
}

// UpdateLinkCounts updates link-related counts
func (m *MetricsCollector) UpdateLinkCounts(total, active int) {
	m.LinksTotal.Set(float64(total))
	m.LinksActive.Set(float64(active))
}

// UpdatePolicyCounts updates policy-related counts
func (m *MetricsCollector) UpdatePolicyCounts(total, active int) {
	m.PoliciesTotal.Set(float64(total))
	m.PoliciesActive.Set(float64(active))
}

// RecordEvent records an event being processed
func (m *MetricsCollector) RecordEvent(eventType, status string) {
	m.EventsProcessed.WithLabelValues(eventType, status).Inc()
}

// RecordAlert records an alert being triggered
func (m *MetricsCollector) RecordAlert(severity, source string) {
	m.AlertsTriggered.WithLabelValues(severity, source).Inc()
}

// UpdateNodePerformance updates node performance metrics
func (m *MetricsCollector) UpdateNodePerformance(nodeID, networkID string, cpuUsage, memoryUsage, diskUsage float64) {
	m.NodeCPUUsage.WithLabelValues(nodeID, networkID).Set(cpuUsage)
	m.NodeMemoryUsage.WithLabelValues(nodeID, networkID).Set(memoryUsage)
	m.NodeDiskUsage.WithLabelValues(nodeID, networkID).Set(diskUsage)
}

// RecordNetworkLatency records network latency
func (m *MetricsCollector) RecordNetworkLatency(networkID, source, destination string, latency time.Duration) {
	m.NetworkLatency.WithLabelValues(networkID, source, destination).Observe(latency.Seconds())
}

// UpdateNetworkThroughput updates network throughput
func (m *MetricsCollector) UpdateNetworkThroughput(networkID, direction string, throughput float64) {
	m.NetworkThroughput.WithLabelValues(networkID, direction).Set(throughput)
}

// RecordNetworkProvisioning records network provisioning time
func (m *MetricsCollector) RecordNetworkProvisioning(networkType, status string, duration time.Duration) {
	m.NetworkProvisioningTime.WithLabelValues(networkType, status).Observe(duration.Seconds())
}

// RecordPolicyViolation records a policy violation
func (m *MetricsCollector) RecordPolicyViolation(policyType, severity, networkID string) {
	m.PolicyViolations.WithLabelValues(policyType, severity, networkID).Inc()
}

// UpdateSLACompliance updates SLA compliance
func (m *MetricsCollector) UpdateSLACompliance(service, slaType string, compliance float64) {
	m.SLACompliance.WithLabelValues(service, slaType).Set(compliance)
}

// StartMetricsCollection starts periodic metrics collection
func (m *MetricsCollector) StartMetricsCollection() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			// Simulate some metrics updates (in real implementation, gather actual data)
			m.UpdateSLACompliance("network_availability", "uptime", 99.95)
			m.UpdateSLACompliance("response_time", "api", 99.8)
			m.CacheHitRate.Set(87.5)
		}
	}()
}
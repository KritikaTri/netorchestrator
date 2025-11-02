package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for NetOrchestrator
type Metrics struct {
	// Infrastructure Metrics
	NetworksTotal     *prometheus.GaugeVec
	NodesTotal        *prometheus.GaugeVec
	ContainersActive  *prometheus.GaugeVec
	LinksTotal        *prometheus.GaugeVec
	PoliciesTotal     *prometheus.GaugeVec

	// Provisioning Metrics
	ProvisionDuration  *prometheus.HistogramVec
	ProvisionOpsTotal  *prometheus.CounterVec

	// Enforcement Metrics
	PolicyEnforcementTotal    *prometheus.CounterVec
	PolicyEnforcementDuration *prometheus.HistogramVec
	IptablesRulesActive       *prometheus.GaugeVec
	TcRulesActive             *prometheus.GaugeVec

	// Container Operations
	ContainerOpsTotal *prometheus.CounterVec
	PodmanErrorsTotal *prometheus.CounterVec

	// NEW: Network-Scoped Metrics (with network_id labels)
	NetworkContainersActive *prometheus.GaugeVec
	NetworkProvisionLatency *prometheus.HistogramVec
	NetworkPolicyFailures   *prometheus.CounterVec
	NetworkLinkStatus       *prometheus.GaugeVec

	// API Layer
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec

	// Database Operations
	DBQueriesTotal    *prometheus.CounterVec
	DBQueryDuration   *prometheus.HistogramVec
}

var (
	// Global metrics instance
	globalMetrics *Metrics
)

// InitMetrics initializes all Prometheus metrics
func InitMetrics() *Metrics {
	m := &Metrics{
		// Infrastructure Metrics
		NetworksTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_networks_total",
				Help: "Total number of networks by status",
			},
			[]string{"status"},
		),
		NodesTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_nodes_total",
				Help: "Total number of nodes by type and status",
			},
			[]string{"type", "status"},
		),
		ContainersActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_containers_active",
				Help: "Number of active containers by node type",
			},
			[]string{"node_type"},
		),
		LinksTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_links_total",
				Help: "Total number of links by status",
			},
			[]string{"status"},
		),
		PoliciesTotal: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_policies_total",
				Help: "Total number of policies by type and status",
			},
			[]string{"type", "status"},
		),

		// Provisioning Metrics
		ProvisionDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorch_provision_duration_seconds",
				Help:    "Duration of provisioning operations",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"operation", "status"},
		),
		ProvisionOpsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorch_provision_operations_total",
				Help: "Total number of provisioning operations",
			},
			[]string{"operation", "status"},
		),

		// Enforcement Metrics
		PolicyEnforcementTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorch_policy_enforcement_total",
				Help: "Total number of policy enforcement operations",
			},
			[]string{"policy_type", "status"},
		),
		PolicyEnforcementDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorch_policy_enforcement_duration_seconds",
				Help:    "Duration of policy enforcement operations",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"policy_type"},
		),
		IptablesRulesActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_iptables_rules_active",
				Help: "Number of active iptables rules per network",
			},
			[]string{"network_id"},
		),
		TcRulesActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_tc_rules_active",
				Help: "Number of active traffic control rules per network",
			},
			[]string{"network_id"},
		),

		// Container Operations
		ContainerOpsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorch_container_operations_total",
				Help: "Total number of container operations",
			},
			[]string{"operation", "status"},
		),
		PodmanErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorch_podman_errors_total",
				Help: "Total number of Podman errors by type",
			},
			[]string{"error_type"},
		),

		// NEW: Network-Scoped Metrics
		NetworkContainersActive: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_network_containers_active",
				Help: "Number of active containers per network",
			},
			[]string{"network_id", "network_name", "node_type"},
		),
		NetworkProvisionLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorch_network_provision_latency_seconds",
				Help:    "Provisioning latency per network and node type",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60},
			},
			[]string{"network_id", "network_name", "node_type", "operation"},
		),
		NetworkPolicyFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorch_network_policy_failures_total",
				Help: "Total policy enforcement failures per network",
			},
			[]string{"network_id", "network_name", "policy_type"},
		),
		NetworkLinkStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "netorch_network_link_status",
				Help: "Link status per network (1=up, 0=down)",
			},
			[]string{"network_id", "network_name", "node_a", "node_b", "link_id"},
		),

		// API Layer
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorch_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorch_http_request_duration_seconds",
				Help:    "Duration of HTTP requests",
				Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
			},
			[]string{"method", "endpoint"},
		),

		// Database Operations
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "netorch_db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "netorch_db_query_duration_seconds",
				Help:    "Duration of database queries",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
			},
			[]string{"operation"},
		),
	}

	globalMetrics = m
	return m
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	if globalMetrics == nil {
		return InitMetrics()
	}
	return globalMetrics
}

// RecordProvisioningOperation records a provisioning operation with timing
func (m *Metrics) RecordProvisioningOperation(operation string, startTime time.Time, err error) {
	duration := time.Since(startTime).Seconds()
	status := "success"
	if err != nil {
		status = "failure"
	}

	m.ProvisionDuration.WithLabelValues(operation, status).Observe(duration)
	m.ProvisionOpsTotal.WithLabelValues(operation, status).Inc()
}

// RecordPolicyEnforcement records a policy enforcement operation
func (m *Metrics) RecordPolicyEnforcement(policyType string, startTime time.Time, err error) {
	duration := time.Since(startTime).Seconds()
	status := "success"
	if err != nil {
		status = "failure"
	}

	m.PolicyEnforcementDuration.WithLabelValues(policyType).Observe(duration)
	m.PolicyEnforcementTotal.WithLabelValues(policyType, status).Inc()
}

// RecordContainerOperation records a container lifecycle operation
func (m *Metrics) RecordContainerOperation(operation string, err error) {
	status := "success"
	if err != nil {
		status = "failure"
	}
	m.ContainerOpsTotal.WithLabelValues(operation, status).Inc()
}

// RecordPodmanError records a Podman-specific error
func (m *Metrics) RecordPodmanError(errorType string) {
	m.PodmanErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration time.Duration) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordDBQuery records a database query
func (m *Metrics) RecordDBQuery(operation string, startTime time.Time, err error) {
	duration := time.Since(startTime).Seconds()
	status := "success"
	if err != nil {
		status = "failure"
	}

	m.DBQueryDuration.WithLabelValues(operation).Observe(duration)
	m.DBQueriesTotal.WithLabelValues(operation, status).Inc()
}

// UpdateInfrastructureMetrics updates infrastructure gauge metrics from database
func (m *Metrics) UpdateInfrastructureMetrics(
	networksByStatus map[string]int,
	nodesByTypeAndStatus map[string]map[string]int,
	containersByType map[string]int,
	linksByStatus map[string]int,
	policiesByTypeAndStatus map[string]map[string]int,
) {
	// Update networks
	for status, count := range networksByStatus {
		m.NetworksTotal.WithLabelValues(status).Set(float64(count))
	}

	// Update nodes
	for nodeType, statusMap := range nodesByTypeAndStatus {
		for status, count := range statusMap {
			m.NodesTotal.WithLabelValues(nodeType, status).Set(float64(count))
		}
	}

	// Update containers
	for nodeType, count := range containersByType {
		m.ContainersActive.WithLabelValues(nodeType).Set(float64(count))
	}

	// Update links
	for status, count := range linksByStatus {
		m.LinksTotal.WithLabelValues(status).Set(float64(count))
	}

	// Update policies
	for policyType, statusMap := range policiesByTypeAndStatus {
		for status, count := range statusMap {
			m.PoliciesTotal.WithLabelValues(policyType, status).Set(float64(count))
		}
	}
}

// UpdateIptablesRuleCount updates the iptables rule count for a network
func (m *Metrics) UpdateIptablesRuleCount(network string, delta float64) {
	m.IptablesRulesActive.WithLabelValues(network).Add(delta)
}

// UpdateTcRuleCount updates the tc rule count for a network
func (m *Metrics) UpdateTcRuleCount(network string, delta float64) {
	m.TcRulesActive.WithLabelValues(network).Add(delta)
}

// NEW: Network-Scoped Metric Helpers

// SetNetworkContainerCount sets the active container count for a specific network
func (m *Metrics) SetNetworkContainerCount(networkID, networkName, nodeType string, count int) {
	m.NetworkContainersActive.WithLabelValues(networkID, networkName, nodeType).Set(float64(count))
}

// IncrementNetworkContainerCount increments the active container count for a network
func (m *Metrics) IncrementNetworkContainerCount(networkID, networkName, nodeType string) {
	m.NetworkContainersActive.WithLabelValues(networkID, networkName, nodeType).Inc()
}

// DecrementNetworkContainerCount decrements the active container count for a network
func (m *Metrics) DecrementNetworkContainerCount(networkID, networkName, nodeType string) {
	m.NetworkContainersActive.WithLabelValues(networkID, networkName, nodeType).Dec()
}

// RecordNetworkProvisionLatency records provisioning latency for a specific network
func (m *Metrics) RecordNetworkProvisionLatency(networkID, networkName, nodeType, operation string, startTime time.Time) {
	duration := time.Since(startTime).Seconds()
	m.NetworkProvisionLatency.WithLabelValues(networkID, networkName, nodeType, operation).Observe(duration)
}

// RecordNetworkPolicyFailure records a policy enforcement failure for a network
func (m *Metrics) RecordNetworkPolicyFailure(networkID, networkName, policyType string) {
	m.NetworkPolicyFailures.WithLabelValues(networkID, networkName, policyType).Inc()
}

// SetNetworkLinkStatus sets the link status (1=up, 0=down) for a specific link
func (m *Metrics) SetNetworkLinkStatus(networkID, networkName, nodeA, nodeB, linkID string, isUp bool) {
	value := 0.0
	if isUp {
		value = 1.0
	}
	m.NetworkLinkStatus.WithLabelValues(networkID, networkName, nodeA, nodeB, linkID).Set(value)
}

// DeleteNetworkLinkStatus removes a link metric (when link is deleted)
func (m *Metrics) DeleteNetworkLinkStatus(networkID, networkName, nodeA, nodeB, linkID string) {
	m.NetworkLinkStatus.DeleteLabelValues(networkID, networkName, nodeA, nodeB, linkID)
}

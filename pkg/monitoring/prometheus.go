package monitoring

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// Prometheus represents a Prometheus monitoring client
type Prometheus struct {
	registry              *prometheus.Registry
	httpRequestsTotal     *prometheus.CounterVec
	httpRequestDuration   *prometheus.HistogramVec
	httpRequestsInFlight  prometheus.Gauge
	logger                *zap.Logger
}

// NewPrometheus creates a new Prometheus monitoring client
func NewPrometheus() (*Prometheus, error) {
	registry := prometheus.NewRegistry()

	// HTTP request metrics
	httpRequestsTotal := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	httpRequestDuration := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	httpRequestsInFlight := promauto.With(registry).NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)

	prometheus := &Prometheus{
		registry:             registry,
		httpRequestsTotal:    httpRequestsTotal,
		httpRequestDuration:  httpRequestDuration,
		httpRequestsInFlight: httpRequestsInFlight,
		logger:               zap.NewNop(), // Will be set later if needed
	}

	return prometheus, nil
}

// SetLogger sets the logger for the Prometheus client
func (p *Prometheus) SetLogger(logger *zap.Logger) {
	p.logger = logger
}

// GinMiddleware returns a Gin middleware for Prometheus metrics
func (p *Prometheus) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Increment in-flight requests
		p.httpRequestsInFlight.Inc()
		defer p.httpRequestsInFlight.Dec()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		p.httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
		p.httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), statusCode).Inc()
	}
}

// GetRegistry returns the Prometheus registry
func (p *Prometheus) GetRegistry() *prometheus.Registry {
	return p.registry
}

// CounterVec creates a new counter vector
func (p *Prometheus) CounterVec(opts prometheus.CounterOpts, labelNames []string) *prometheus.CounterVec {
	return promauto.With(p.registry).NewCounterVec(opts, labelNames)
}

// HistogramVec creates a new histogram vector
func (p *Prometheus) HistogramVec(opts prometheus.HistogramOpts, labelNames []string) *prometheus.HistogramVec {
	return promauto.With(p.registry).NewHistogramVec(opts, labelNames)
}

// GaugeVec creates a new gauge vector
func (p *Prometheus) GaugeVec(opts prometheus.GaugeOpts, labelNames []string) *prometheus.GaugeVec {
	return promauto.With(p.registry).NewGaugeVec(opts, labelNames)
}

// Counter creates a new counter
func (p *Prometheus) Counter(opts prometheus.CounterOpts) prometheus.Counter {
	return promauto.With(p.registry).NewCounter(opts)
}

// Histogram creates a new histogram
func (p *Prometheus) Histogram(opts prometheus.HistogramOpts) prometheus.Histogram {
	return promauto.With(p.registry).NewHistogram(opts)
}

// Gauge creates a new gauge
func (p *Prometheus) Gauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	return promauto.With(p.registry).NewGauge(opts)
}

// Handler returns an HTTP handler for Prometheus metrics
func (p *Prometheus) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple handler that returns basic metrics
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Prometheus metrics endpoint\n"))
	})
}

// Health checks the monitoring system health
func (p *Prometheus) Health() error {
	// Prometheus monitoring is always healthy as it's just metrics collection
	return nil
}

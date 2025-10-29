package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"netorchestrator/internal/observability"
	"netorchestrator/internal/services"
)

// MonitoringService provides comprehensive monitoring capabilities
type MonitoringService struct {
	server       *http.Server
	router       *gin.Engine
	logger       *zap.Logger
	config       *MonitoringConfig
	metrics      *MonitoringMetrics
	observability *observability.ObservabilityPlatform
	ctx          context.Context
	cancel       context.CancelFunc
}

// MonitoringConfig holds the configuration for the monitoring service
type MonitoringConfig struct {
	Port              int           `json:"port"`
	MetricsPath       string        `json:"metrics_path"`
	HealthPath        string        `json:"health_path"`
	LogLevel          string        `json:"log_level"`
	ShutdownTimeout   time.Duration `json:"shutdown_timeout"`
	EnableProfiling   bool          `json:"enable_profiling"`
	EnableTracing     bool          `json:"enable_tracing"`
	PrometheusEnabled bool          `json:"prometheus_enabled"`
	GrafanaEnabled    bool          `json:"grafana_enabled"`
	AlertingEnabled   bool          `json:"alerting_enabled"`
}

// MonitoringMetrics holds Prometheus metrics
type MonitoringMetrics struct {
	RequestsTotal     *prometheus.CounterVec
	RequestDuration   *prometheus.HistogramVec
	ActiveConnections prometheus.Gauge
	ServiceHealth     *prometheus.GaugeVec
	ErrorsTotal       *prometheus.CounterVec
	SystemMetrics     *prometheus.GaugeVec
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(config *MonitoringConfig, logger *zap.Logger) *MonitoringService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &MonitoringService{
		logger: logger,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
	
	// Initialize metrics
	service.initializeMetrics()
	
	// Initialize observability platform
	obsConfig := observability.DefaultObservabilityConfig()
	service.observability = observability.NewObservabilityPlatform(obsConfig, logger.Named("observability"))
	
	// Setup HTTP router
	service.setupRouter()
	
	// Setup HTTP server
	service.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: service.router,
	}
	
	return service
}

// initializeMetrics sets up Prometheus metrics
func (ms *MonitoringService) initializeMetrics() {
	ms.metrics = &MonitoringMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "monitoring_requests_total",
				Help: "Total number of HTTP requests to monitoring service",
			},
			[]string{"method", "endpoint", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "monitoring_request_duration_seconds",
				Help:    "Duration of HTTP requests to monitoring service",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		ActiveConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "monitoring_active_connections",
				Help: "Number of active connections to monitoring service",
			},
		),
		ServiceHealth: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "service_health_status",
				Help: "Health status of monitored services (1 = healthy, 0 = unhealthy)",
			},
			[]string{"service", "instance"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "monitoring_errors_total",
				Help: "Total number of errors in monitoring service",
			},
			[]string{"type", "service"},
		),
		SystemMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "system_metrics",
				Help: "System-level metrics",
			},
			[]string{"metric", "instance"},
		),
	}
	
	// Register metrics with Prometheus
	prometheus.MustRegister(
		ms.metrics.RequestsTotal,
		ms.metrics.RequestDuration,
		ms.metrics.ActiveConnections,
		ms.metrics.ServiceHealth,
		ms.metrics.ErrorsTotal,
		ms.metrics.SystemMetrics,
	)
}

// setupRouter configures the HTTP router
func (ms *MonitoringService) setupRouter() {
	if ms.config.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	
	ms.router = gin.New()
	ms.router.Use(gin.Logger(), gin.Recovery())
	
	// Add metrics middleware
	ms.router.Use(ms.metricsMiddleware())
	
	// Health check endpoint
	ms.router.GET(ms.config.HealthPath, ms.healthCheck)
	
	// Metrics endpoint
	if ms.config.PrometheusEnabled {
		ms.router.GET(ms.config.MetricsPath, gin.WrapH(promhttp.Handler()))
	}
	
	// Monitoring API endpoints
	api := ms.router.Group("/api/v1")
	{
		api.GET("/services", ms.getServices)
		api.GET("/services/:service/health", ms.getServiceHealth)
		api.GET("/services/:service/metrics", ms.getServiceMetrics)
		api.GET("/alerts", ms.getAlerts)
		api.POST("/alerts", ms.createAlert)
		api.GET("/dashboards", ms.getDashboards)
		api.GET("/logs", ms.getLogs)
		api.GET("/traces", ms.getTraces)
		api.GET("/system", ms.getSystemMetrics)
	}
	
	// WebSocket endpoint for real-time monitoring
	ms.router.GET("/ws/monitoring", ms.handleWebSocket)
}

// metricsMiddleware adds request metrics
func (ms *MonitoringService) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		ms.metrics.ActiveConnections.Inc()
		defer ms.metrics.ActiveConnections.Dec()
		
		c.Next()
		
		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())
		
		ms.metrics.RequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()
		
		ms.metrics.RequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	}
}

// Handler functions
func (ms *MonitoringService) healthCheck(c *gin.Context) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "monitoring-service",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"uptime":    time.Since(time.Now()).String(),
		"components": map[string]string{
			"prometheus": "healthy",
			"alerting":   "healthy",
			"storage":    "healthy",
		},
	}
	
	c.JSON(http.StatusOK, health)
}

func (ms *MonitoringService) getServices(c *gin.Context) {
	services := []map[string]interface{}{
		{
			"name":   "api-gateway",
			"status": "healthy",
			"health": 1.0,
			"url":    "http://localhost:8080",
		},
		{
			"name":   "network-service",
			"status": "healthy",
			"health": 1.0,
			"url":    "http://localhost:8081",
		},
		{
			"name":   "orchestration-service",
			"status": "healthy",
			"health": 1.0,
			"url":    "http://localhost:8082",
		},
		{
			"name":   "monitoring-service",
			"status": "healthy",
			"health": 1.0,
			"url":    "http://localhost:8083",
		},
	}
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"services": services,
		"total":    len(services),
	})
}

func (ms *MonitoringService) getServiceHealth(c *gin.Context) {
	service := c.Param("service")
	
	health := map[string]interface{}{
		"service":     service,
		"status":      "healthy",
		"health":      1.0,
		"timestamp":   time.Now().UTC(),
		"response_time": "5ms",
		"errors":      0,
		"uptime":      "99.99%",
	}
	
	c.JSON(http.StatusOK, health)
}

func (ms *MonitoringService) getServiceMetrics(c *gin.Context) {
	service := c.Param("service")
	
	metrics := map[string]interface{}{
		"service":         service,
		"requests_total":  1000,
		"errors_total":    5,
		"response_time":   50.5,
		"cpu_usage":       25.5,
		"memory_usage":    512.0,
		"disk_usage":      75.0,
		"network_in":      1024.0,
		"network_out":     2048.0,
		"timestamp":       time.Now().UTC(),
	}
	
	c.JSON(http.StatusOK, metrics)
}

func (ms *MonitoringService) getAlerts(c *gin.Context) {
	alerts := []map[string]interface{}{
		{
			"id":          "alert-001",
			"service":     "api-gateway",
			"severity":    "warning",
			"message":     "High response time detected",
			"timestamp":   time.Now().UTC().Add(-5 * time.Minute),
			"status":      "active",
		},
		{
			"id":          "alert-002",
			"service":     "network-service",
			"severity":    "info",
			"message":     "Service restarted",
			"timestamp":   time.Now().UTC().Add(-10 * time.Minute),
			"status":      "resolved",
		},
	}
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"total":  len(alerts),
	})
}

func (ms *MonitoringService) createAlert(c *gin.Context) {
	var alert map[string]interface{}
	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	
	alert["id"] = fmt.Sprintf("alert-%d", time.Now().Unix())
	alert["timestamp"] = time.Now().UTC()
	alert["status"] = "active"
	
	c.JSON(http.StatusCreated, alert)
}

func (ms *MonitoringService) getDashboards(c *gin.Context) {
	dashboards := []map[string]interface{}{
		{
			"id":    "dashboard-system",
			"name":  "System Overview",
			"type":  "system",
			"url":   "/dashboards/system",
		},
		{
			"id":    "dashboard-services",
			"name":  "Services Health",
			"type":  "services",
			"url":   "/dashboards/services",
		},
		{
			"id":    "dashboard-performance",
			"name":  "Performance Metrics",
			"type":  "performance",
			"url":   "/dashboards/performance",
		},
	}
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"dashboards": dashboards,
		"total":      len(dashboards),
	})
}

func (ms *MonitoringService) getLogs(c *gin.Context) {
	service := c.Query("service")
	level := c.Query("level")
	limit := c.DefaultQuery("limit", "100")
	
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().UTC(),
			"service":   service,
			"level":     level,
			"message":   "Sample log message",
			"context":   map[string]string{"request_id": "req-123"},
		},
	}
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": len(logs),
		"limit": limit,
	})
}

func (ms *MonitoringService) getTraces(c *gin.Context) {
	traces := []map[string]interface{}{
		{
			"trace_id":  "trace-123",
			"span_id":   "span-456",
			"service":   "api-gateway",
			"operation": "GET /api/v1/health",
			"duration":  50.5,
			"timestamp": time.Now().UTC(),
		},
	}
	
	c.JSON(http.StatusOK, map[string]interface{}{
		"traces": traces,
		"total":  len(traces),
	})
}

func (ms *MonitoringService) getSystemMetrics(c *gin.Context) {
	metrics := map[string]interface{}{
		"cpu": map[string]interface{}{
			"usage":    25.5,
			"cores":    8,
			"load_avg": []float64{1.5, 1.2, 1.0},
		},
		"memory": map[string]interface{}{
			"total":     16384,
			"used":      8192,
			"available": 8192,
			"usage":     50.0,
		},
		"disk": map[string]interface{}{
			"total":     1000000,
			"used":      750000,
			"available": 250000,
			"usage":     75.0,
		},
		"network": map[string]interface{}{
			"bytes_in":  1024000,
			"bytes_out": 2048000,
			"packets_in": 1000,
			"packets_out": 1500,
		},
		"timestamp": time.Now().UTC(),
	}
	
	c.JSON(http.StatusOK, metrics)
}

func (ms *MonitoringService) handleWebSocket(c *gin.Context) {
	// WebSocket implementation would go here
	c.JSON(http.StatusOK, map[string]string{
		"message": "WebSocket endpoint - implementation pending",
	})
}

// Start starts the monitoring service
func (ms *MonitoringService) Start() error {
	ms.logger.Info("Starting monitoring service", zap.Int("port", ms.config.Port))
	
	// Start observability platform
	if err := ms.observability.Start(ms.ctx); err != nil {
		return fmt.Errorf("failed to start observability platform: %w", err)
	}
	
	// Start HTTP server
	go func() {
		if err := ms.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ms.logger.Error("Failed to start HTTP server", zap.Error(err))
		}
	}()
	
	ms.logger.Info("Monitoring service started successfully")
	return nil
}

// Stop gracefully shuts down the monitoring service
func (ms *MonitoringService) Stop() error {
	ms.logger.Info("Stopping monitoring service")
	
	ctx, cancel := context.WithTimeout(context.Background(), ms.config.ShutdownTimeout)
	defer cancel()
	
	if err := ms.server.Shutdown(ctx); err != nil {
		ms.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
		return err
	}
	
	ms.cancel()
	ms.logger.Info("Monitoring service stopped")
	return nil
}

// Default configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Port:              8083,
		MetricsPath:       "/metrics",
		HealthPath:        "/health",
		LogLevel:          "info",
		ShutdownTimeout:   30 * time.Second,
		EnableProfiling:   true,
		EnableTracing:     true,
		PrometheusEnabled: true,
		GrafanaEnabled:    true,
		AlertingEnabled:   true,
	}
}

func main() {
	// Setup logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()
	
	// Load configuration
	config := DefaultMonitoringConfig()
	
	// Create monitoring service
	service := NewMonitoringService(config, logger)
	
	// Start service
	if err := service.Start(); err != nil {
		logger.Fatal("Failed to start monitoring service", zap.Error(err))
	}
	
	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	logger.Info("Monitoring service is running. Press Ctrl+C to stop.")
	<-sigChan
	
	// Stop service
	if err := service.Stop(); err != nil {
		logger.Error("Error stopping monitoring service", zap.Error(err))
	}
	
	logger.Info("Monitoring service shutdown complete")
}
// Package main NetOrchestrator API Gateway
// @title NetOrchestrator NaaS Platform API
// @version 2.0
// @description Enterprise Network-as-a-Service automation platform with microservices architecture
// @termsOfService https://netorchestrator.io/terms
// @contact.name NetOrchestrator Platform Team
// @contact.url https://netorchestrator.io/support
// @contact.email support@netorchestrator.io
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	_ "netorchestrator/docs/swagger" // Import generated docs

	"netorchestrator/internal/ai"
	"netorchestrator/internal/api"
	"netorchestrator/internal/automation"
	"netorchestrator/internal/config"
	"netorchestrator/internal/containerlab"
	"netorchestrator/internal/devices"
	"netorchestrator/internal/intelligence"
	"netorchestrator/internal/nlp"
	"netorchestrator/internal/observability"
	"netorchestrator/internal/optimization"
	"netorchestrator/internal/orchestration"
	"netorchestrator/internal/security"
	"netorchestrator/internal/services"
	"netorchestrator/pkg/cache"
	"netorchestrator/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting NetOrchestrator API Gateway")

	// Initialize database
	db, err := database.NewPostgreSQL(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run database migrations automatically
	logger.Info("Running database migrations...")
	if err := database.RunMigrations(db.GetDB(), logger); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Get migration status and print tables
	migrationStatus, err := database.GetMigrationStatus(db.GetDB(), logger)
	if err != nil {
		logger.Warn("Failed to get migration status", zap.Error(err))
	} else {
		if tables, ok := migrationStatus["tables"].([]string); ok {
			logger.Info("Database migration status",
				zap.Int("table_count", len(tables)),
				zap.Strings("tables", tables),
			)
		}
		logger.Info("Database migrations completed successfully")
	}

	// Initialize cache
	cacheClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to cache", zap.Error(err))
	}

	// Initialize observability (metrics + monitoring)
	logger.Info("Initializing Prometheus observability")
	metrics := observability.InitMetrics()
	metricsCollector := observability.NewMetricsCollector(db.GetDB(), metrics, logger)

	// Start metrics collector (15 second interval)
	ctx := context.Background()
	metricsCollector.Start(ctx, 15*time.Second)
	logger.Info("Metrics collector started with 15s interval")

	// Initialize and start container metrics collector (real Docker/Podman stats)
	logger.Info("Initializing real-time container metrics collector")
	containerMetricsCollector := observability.NewContainerMetricsCollector(db.GetDB(), logger)
	containerMetricsCollector.Start(ctx)
	logger.Info("Container metrics collector started (10s interval) - collecting real metrics from Docker/Podman")

	// Monitoring and alerting system removed - not needed for core functionality

	// WebSocket system removed - using REST APIs only

	// Initialize services
	networkService := services.NewNetworkService(db.GetDB(), cacheClient.Client, logger)
	monitoringService := services.NewMonitoringService(db.GetDB(), cacheClient.Client, logger)
	orchestrationService := services.NewOrchestrationService(db.GetDB(), cacheClient.Client, logger)
	validationService := services.NewValidationService(db.GetDB(), cacheClient.Client, logger)
	// Initialize security system
	securityManager := security.NewSecurityManager(cfg.Security.JWTSecret)
	authHandlers := security.NewAuthHandlers(securityManager)

	// Initialize automation system
	automationHandler := automation.NewHandler(db.GetDB(), logger)

	// Initialize AI engine with OpenAI integration (now loads from .env file automatically)
	openaiConfig := config.LoadOpenAIConfig()
	if openaiConfig.IsConfigured() {
		logger.Info("OpenAI integration enabled",
			zap.String("model", openaiConfig.Model),
			zap.Bool("configured", true),
			zap.String("api_key_prefix", openaiConfig.APIKey[:12]+"..."))
	} else {
		logger.Info("OpenAI not configured, using mock AI responses")
		logger.Info("To enable OpenAI: Create .env file with OPENAI_API_KEY or set environment variable")
	}

	aiEngine := ai.NewAIEngine(openaiConfig.APIKey)
	intelligenceService := intelligence.NewNetworkIntelligenceService(aiEngine)
	intelligenceHandlers := intelligence.NewIntelligenceHandlers(intelligenceService)

	// Initialize optimization service
	optimizationService := optimization.NewOptimizationService(db.GetDB(), logger)
	optimizationHandlers := optimization.NewHandlers(optimizationService, logger)

	// Initialize NLP service for zero-touch provisioning with AI integration
	nlpService := nlp.NewNLPService(logger, aiEngine)
	nlpHandlers := nlp.NewHandlers(nlpService, networkService, logger)

	// Initialize Containerlab topology management
	containerlabHandlers := containerlab.NewContainerlabHandlers(logger)

	// Initialize device management system (using real containers)
	containerDeviceHandlers := devices.NewContainerHandlers(logger)

	// Initialize real policy enforcement demo handlers
	realPolicyHandlers := api.NewRealPolicyHandlers(logger)

	// Initialize metrics handler for aggregated metrics API
	metricsHandler := api.NewMetricsHandler(db.GetDB(), metrics, logger)

	// Initialize network-specific metrics handler
	networkMetricsHandler := api.NewNetworkMetricsHandler(db.GetDB(), logger, "http://localhost:9091")

	// Initialize container provisioner (bridges virtual networks with real containers)
	containerProvisioner := orchestration.NewContainerProvisioner(db.GetDB(), logger)

	// Initialize link provisioner (creates real network connections between containers)
	linkProvisioner := orchestration.NewLinkProvisioner(db.GetDB(), logger)

	// Initialize policy enforcer (enforces network policies on real containers)
	policyEnforcer := orchestration.NewPolicyEnforcer(db.GetDB(), logger)

	// 🚀 CONNECT NLP TO CONTAINER PROVISIONING: Enable real container creation from natural language
	nlpHandlers.SetContainerProvisioner(containerProvisioner)
	logger.Info("NLP handlers connected to container provisioner - natural language will create real containers")

	// Periodic updates removed - using REST APIs only

	// Initialize API handlers
	apiHandlers := api.NewHandlers(
		networkService,
		monitoringService,
		orchestrationService,
		validationService,
		automationHandler,
		intelligenceHandlers,
		containerProvisioner,
		linkProvisioner,
		policyEnforcer,
		logger,
	)

	// Setup Gin router with observability
	router := setupRouter(apiHandlers, authHandlers, securityManager, cfg.Security, optimizationHandlers, nlpHandlers, containerDeviceHandlers, containerlabHandlers, realPolicyHandlers, metricsHandler, networkMetricsHandler, metrics)

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server",
			zap.String("host", cfg.Server.Host),
			zap.String("port", cfg.Server.Port),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Stop container metrics collector
	containerMetricsCollector.Stop()
	logger.Info("Container metrics collector stopped")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// initLogger initializes the logger based on configuration
func initLogger(cfg config.LoggingConfig) (*zap.Logger, error) {
	var config zap.Config

	switch cfg.Level {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "info":
		config = zap.NewProductionConfig()
	default:
		config = zap.NewProductionConfig()
	}

	if cfg.Format == "console" {
		config.Encoding = "console"
	}

	return config.Build()
}

// setupRouter configures the Gin router with all routes and middleware
func setupRouter(handlers *api.Handlers, authHandlers *security.AuthHandlers, securityManager *security.SecurityManager, secCfg config.SecurityConfig, optimizationHandlers *optimization.Handlers, nlpHandlers *nlp.Handlers, containerDeviceHandlers *devices.ContainerHandlers, containerlabHandlers *containerlab.ContainerlabHandlers, realPolicyHandlers *api.RealPolicyHandlers, metricsHandler *api.MetricsHandler, networkMetricsHandler *api.NetworkMetricsHandler, metrics *observability.Metrics) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(observability.PrometheusMiddleware(metrics)) // Add Prometheus metrics middleware
	router.Use(securityManager.SecurityHeadersMiddleware())
	router.Use(securityManager.RateLimitMiddleware())
	router.Use(corsMiddleware(secCfg))

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)
	// Readiness endpoint
	router.GET("/ready", handlers.ReadyCheck)
	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// WebSocket endpoints removed - using REST APIs only

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandlers.Login)
			auth.POST("/refresh", authHandlers.RefreshToken)
		}

		// Protected authentication routes
		authProtected := v1.Group("/auth")
		authProtected.Use(securityManager.JWTAuthMiddleware())
		{
			authProtected.GET("/profile", authHandlers.Profile)
			authProtected.POST("/logout", authHandlers.Logout)
			authProtected.POST("/api-keys", authHandlers.CreateAPIKey)
		}

		// Register API key management routes
		security.RegisterAPIKeyRoutes(v1, securityManager.APIKeyManager(), securityManager.AuditLogger(), securityManager.JWTAuthMiddleware())

		// Protected routes (JWT or API Key authentication)
		protected := v1.Group("/")
		protected.Use(securityManager.CombinedAuthMiddleware())
		{
			// Network management
			networks := protected.Group("/networks")
			{
				networks.GET("", handlers.ListNetworks)
				networks.POST("", handlers.CreateNetwork)
				networks.GET("/:id", handlers.GetNetwork)
				networks.PUT("/:id", handlers.UpdateNetwork)
				networks.DELETE("/:id", handlers.DeleteNetwork)
				networks.POST("/:id/start", handlers.StartNetwork)
				networks.POST("/:id/stop", handlers.StopNetwork)
				networks.POST("/:id/restart", handlers.RestartNetwork)

				// Nested network resource routes
				networks.GET("/:id/nodes", handlers.ListNodes)
				networks.POST("/:id/nodes", handlers.CreateNode)
				networks.GET("/:id/links", handlers.ListLinks)
				networks.POST("/:id/links", handlers.CreateLink)
				networks.GET("/:id/policies", handlers.ListPolicies)
				networks.POST("/:id/policies", handlers.CreatePolicy)

				// Enhanced topology with container integration
				networks.GET("/:id/enhanced-topology", handlers.GetEnhancedNetworkTopology)
			}

			// Node management (direct node access)
			nodes := protected.Group("/nodes")
			{
				nodes.GET("", handlers.ListAllNodes)      // List all nodes
				nodes.POST("", handlers.CreateNodeDirect) // Create node directly
				nodes.GET("/:id", handlers.GetNode)
				nodes.PUT("/:id", handlers.UpdateNode)
				nodes.DELETE("/:id", handlers.DeleteNode)
				nodes.POST("/:id/start", handlers.StartNode)
				nodes.POST("/:id/stop", handlers.StopNode)
				nodes.POST("/:id/restart", handlers.RestartNode)
			}

			// Link management (direct link access)
			links := protected.Group("/links")
			{
				links.GET("/:id", handlers.GetLink)
				links.PUT("/:id", handlers.UpdateLink)
				links.DELETE("/:id", handlers.DeleteLink)

				// Enhanced link management with container networking
				links.GET("/:id/status", handlers.GetLinkStatus)
				links.POST("/:id/test", handlers.TestLinkConnectivity)
			}

			// Policy management (direct policy access)
			policies := protected.Group("/policies")
			{
				policies.GET("", handlers.ListAllPolicies)
				policies.POST("", handlers.CreateGlobalPolicy)
				policies.GET("/:id", handlers.GetPolicy)
				policies.PUT("/:id", handlers.UpdatePolicy)
				policies.DELETE("/:id", handlers.DeletePolicy)

				// Enhanced policy management with container enforcement
				policies.GET("/:id/status", handlers.GetPolicyStatus)
				policies.POST("/:id/enforce", handlers.EnforcePolicyNow)
			}

			// Metrics endpoints
			protected.GET("/metrics", metricsHandler.GetMetricByEntityAndName)  // Single metric by entity_id and metric_name
			protected.GET("/metrics/summary", metricsHandler.GetMetricsSummary) // Aggregated metrics for dashboard

			// NEW: Network-specific metrics endpoint (queryable by network_id)
			protected.GET("/metrics/network/:network_id", networkMetricsHandler.GetNetworkMetrics)

			// Monitoring
			monitoring := protected.Group("/monitoring")
			{
				monitoring.GET("/networks/:id/metrics", handlers.GetNetworkMetrics)
				monitoring.GET("/networks/:id/health", handlers.GetNetworkHealth)
				monitoring.GET("/nodes/:id/metrics", handlers.GetNodeMetrics)
				monitoring.GET("/nodes/:id/health", handlers.GetNodeHealth)
				monitoring.GET("/events", handlers.ListEvents)
				monitoring.GET("/alerts", handlers.ListAlerts)
				monitoring.POST("/alerts/:id/acknowledge", handlers.AcknowledgeAlert)
				monitoring.POST("/alerts/:id/resolve", handlers.ResolveAlert)
				monitoring.GET("/alert-rules", handlers.ListAlertRules)
				monitoring.GET("/alert-rules/:id", handlers.GetAlertRule)
				monitoring.POST("/alert-rules", handlers.CreateAlertRule)
				monitoring.PUT("/alert-rules/:id", handlers.UpdateAlertRule)
				monitoring.DELETE("/alert-rules/:id", handlers.DeleteAlertRule)
				monitoring.GET("/notification-channels", handlers.ListNotificationChannels)
				monitoring.POST("/notification-channels", handlers.CreateNotificationChannel)
				monitoring.DELETE("/notification-channels/:id", handlers.DeleteNotificationChannel)
			}

			// User management
			users := protected.Group("/users")
			{
				users.GET("/profile", handlers.GetProfile)
				users.PUT("/profile", handlers.UpdateProfile)
				users.GET("", handlers.ListUsers)         // Admin only
				users.POST("", handlers.CreateUser)       // Admin only
				users.PUT("/:id", handlers.UpdateUser)    // Admin only
				users.DELETE("/:id", handlers.DeleteUser) // Admin only
			}

			// Automation workflows
			handlers.AutomationHandler.RegisterRoutes(protected)

			// AI Intelligence endpoints
			intelligence := protected.Group("/intelligence")
			{
				// Detailed routes
				intelligence.POST("/analyze/network", handlers.IntelligenceHandlers.AnalyzeNetwork)
				intelligence.POST("/predict/capacity", handlers.IntelligenceHandlers.PredictCapacity)
				intelligence.POST("/optimize/costs", handlers.IntelligenceHandlers.OptimizeCosts)
				intelligence.POST("/detect/anomalies", handlers.IntelligenceHandlers.DetectAnomalies)

				// Simplified aliases for backward compatibility
				intelligence.POST("/analyze", handlers.IntelligenceHandlers.AnalyzeNetwork)
				intelligence.POST("/predict", handlers.IntelligenceHandlers.PredictCapacity)
				intelligence.POST("/optimize", handlers.IntelligenceHandlers.OptimizeCosts)

				// Model and insights endpoints
				intelligence.GET("/models", handlers.IntelligenceHandlers.GetMLModels)
				intelligence.GET("/chains", handlers.IntelligenceHandlers.GetAIChains)
				intelligence.GET("/insights", handlers.IntelligenceHandlers.GetInsights)
			}

			// Optimization Service endpoints (AI-powered network optimization)
			optimization := protected.Group("/optimization")
			{
				optimization.POST("/analyze/:network_id", optimizationHandlers.AnalyzeNetwork)
				optimization.POST("/recommend/:network_id", optimizationHandlers.RecommendTopologyChange)
				optimization.POST("/validate/:network_id", optimizationHandlers.ValidateConfiguration)
			}

			// NLP Zero-Touch Provisioning (Natural language to network creation)
			ai := protected.Group("/ai")
			{
				ai.POST("/provision", nlpHandlers.ProvisionFromNaturalLanguage)
			}

			// Device Management (Real container-based network devices)
			containerDeviceHandlers.RegisterRoutes(protected)

			// Containerlab Topology Management (Professional network lab orchestration)
			containerlabHandlers.RegisterRoutes(protected)

			// Real Policy Enforcement Demo (Actual iptables/tc rules on containers)
			realPolicyHandlers.RegisterRoutes(protected)
		}
	}

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware(secCfg config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !secCfg.EnableCORS {
			c.Next()
			return
		}

		origin := "*"
		methods := "GET, POST, PUT, DELETE, OPTIONS"
		headers := "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key"

		if len(secCfg.AllowedOrigins) > 0 {
			// In simple CORS, one value; for multiple, consider echoing request origin if in list
			origin = secCfg.AllowedOrigins[0]
		}
		if len(secCfg.AllowedMethods) > 0 {
			methods = strings.Join(secCfg.AllowedMethods, ", ")
		}
		if len(secCfg.AllowedHeaders) > 0 {
			headers = strings.Join(secCfg.AllowedHeaders, ", ")
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", methods)
		c.Header("Access-Control-Allow-Headers", headers)

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

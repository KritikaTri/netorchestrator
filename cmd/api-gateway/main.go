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
	"netorchestrator/internal/events"
	"netorchestrator/internal/intelligence"
	"netorchestrator/internal/security"
	"netorchestrator/internal/services"
	"netorchestrator/internal/websocket"
	"netorchestrator/pkg/alerting"
	"netorchestrator/pkg/cache"
	"netorchestrator/pkg/database"
	"netorchestrator/pkg/monitoring"
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

	// Initialize cache
	cacheClient, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to cache", zap.Error(err))
	}

	// Initialize monitoring
	monitor, err := monitoring.NewPrometheus()
	if err != nil {
		logger.Fatal("Failed to initialize monitoring", zap.Error(err))
	}

	// Initialize alerting
	alertManager := alerting.NewAlertManager(logger)
	alertManager.Initialize()
	alertManager.StartAlertEvaluation()

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(logger)
	go wsHub.Run()

	// Initialize services
	networkService := services.NewNetworkService(db.GetDB(), cacheClient.Client, logger)
	monitoringService := services.NewMonitoringService(db.GetDB(), cacheClient.Client, logger)
	orchestrationService := services.NewOrchestrationService(db.GetDB(), cacheClient.Client, logger)
	validationService := services.NewValidationService(db.GetDB(), cacheClient.Client, logger)
	eventService := services.NewEventService(wsHub, logger)

	// Initialize security system
	securityManager := security.NewSecurityManager(cfg.Security.JWTSecret)
	authHandlers := security.NewAuthHandlers(securityManager)

	// Initialize automation system
	automationHandler := automation.NewHandler(db.GetDB(), logger)

	// Initialize AI engine and intelligence services
	aiEngine := ai.NewAIEngine()
	eventStore := events.NewEventStore()
	intelligenceService := intelligence.NewNetworkIntelligenceService(aiEngine, eventStore)
	intelligenceHandlers := intelligence.NewIntelligenceHandlers(intelligenceService)

	// Start periodic WebSocket updates
	eventService.StartPeriodicUpdates()

	// Initialize API handlers
	apiHandlers := api.NewHandlers(
		networkService,
		monitoringService,
		orchestrationService,
		validationService,
		automationHandler,
		intelligenceHandlers,
		logger,
	)

	// Setup Gin router
	router := setupRouter(apiHandlers, authHandlers, securityManager, monitor, wsHub)

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
func setupRouter(handlers *api.Handlers, authHandlers *security.AuthHandlers, securityManager *security.SecurityManager, prometheus *monitoring.Prometheus, wsHub *websocket.Hub) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(securityManager.SecurityHeadersMiddleware())
	router.Use(securityManager.RateLimitMiddleware())
	router.Use(prometheus.GinMiddleware())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// WebSocket endpoint
	router.GET("/ws", wsHub.HandleWebSocket)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
		protected.Use(func(c *gin.Context) {
			// Try JWT first, then API key
			authHeader := c.GetHeader("Authorization")
			apiKey := c.GetHeader("X-API-Key")
			
			if authHeader != "" {
				securityManager.JWTAuthMiddleware()(c)
			} else if apiKey != "" {
				securityManager.APIKeyAuthMiddleware()(c)
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}
		})
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
			}

			// Node management (direct node access)
			nodes := protected.Group("/nodes")
			{
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
			}

			// Policy management (direct policy access)
			policies := protected.Group("/policies")
			{
				policies.GET("/:id", handlers.GetPolicy)
				policies.PUT("/:id", handlers.UpdatePolicy)
				policies.DELETE("/:id", handlers.DeletePolicy)
			}

			// Monitoring
			monitoring := protected.Group("/monitoring")
			{
				monitoring.GET("/networks/:id/metrics", handlers.GetNetworkMetrics)
				monitoring.GET("/networks/:id/health", handlers.GetNetworkHealth)
				monitoring.GET("/nodes/:id/metrics", handlers.GetNodeMetrics)
				monitoring.GET("/nodes/:id/health", handlers.GetNodeHealth)
				monitoring.GET("/alerts", handlers.ListAlerts)
				monitoring.POST("/alerts/:id/acknowledge", handlers.AcknowledgeAlert)
				monitoring.POST("/alerts/:id/resolve", handlers.ResolveAlert)
				monitoring.GET("/alert-rules", handlers.ListAlertRules)
				monitoring.POST("/alert-rules", handlers.CreateAlertRule)
				monitoring.PUT("/alert-rules/:id", handlers.UpdateAlertRule)
				monitoring.DELETE("/alert-rules/:id", handlers.DeleteAlertRule)
				monitoring.GET("/notification-channels", handlers.ListNotificationChannels)
				monitoring.POST("/notification-channels", handlers.CreateNotificationChannel)
				monitoring.PUT("/notification-channels/:id", handlers.UpdateNotificationChannel)
				monitoring.DELETE("/notification-channels/:id", handlers.DeleteNotificationChannel)
			}

			// User management
			users := protected.Group("/users")
			{
				users.GET("/profile", handlers.GetProfile)
				users.PUT("/profile", handlers.UpdateProfile)
				users.GET("", handlers.ListUsers) // Admin only
				users.POST("", handlers.CreateUser) // Admin only
				users.PUT("/:id", handlers.UpdateUser) // Admin only
				users.DELETE("/:id", handlers.DeleteUser) // Admin only
			}

			// Automation workflows
			handlers.AutomationHandler.RegisterRoutes(protected)

			// AI Intelligence endpoints
			intelligence := protected.Group("/intelligence")
			{
				intelligence.POST("/analyze/network", handlers.IntelligenceHandlers.AnalyzeNetwork)
				intelligence.POST("/predict/capacity", handlers.IntelligenceHandlers.PredictCapacity)
				intelligence.POST("/optimize/costs", handlers.IntelligenceHandlers.OptimizeCosts)
				intelligence.POST("/detect/anomalies", handlers.IntelligenceHandlers.DetectAnomalies)
				intelligence.GET("/models", handlers.IntelligenceHandlers.GetMLModels)
				intelligence.GET("/chains", handlers.IntelligenceHandlers.GetAIChains)
				intelligence.GET("/insights", handlers.IntelligenceHandlers.GetInsights)
			}
		}
	}

	return router
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

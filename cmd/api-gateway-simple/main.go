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
	"go.uber.org/zap"

	"netorchestrator/internal/api"
	"netorchestrator/internal/config"
	"netorchestrator/internal/services"
	"netorchestrator/pkg/monitoring"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config-simple.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting NetOrchestrator API Gateway (Simple Mode - No External Dependencies)")

	// Initialize monitoring (no external dependencies)
	monitor, err := monitoring.NewPrometheus()
	if err != nil {
		logger.Fatal("Failed to initialize monitoring", zap.Error(err))
	}

	// Initialize services with nil dependencies (in-memory mode)
	networkService := services.NewNetworkService(nil, nil, logger)
	monitoringService := services.NewMonitoringService(nil, nil, logger)
	orchestrationService := services.NewOrchestrationService(nil, nil, logger)
	validationService := services.NewValidationService(nil, nil, logger)

	// Initialize API handlers (with nil for automation and intelligence for simple version)
	apiHandlers := api.NewHandlers(
		networkService,
		monitoringService,
		orchestrationService,
		validationService,
		nil, // automation handler - not needed for simple version
		nil, // intelligence handlers - not needed for simple version
		logger,
	)

	// Setup Gin router
	router := setupRouter(apiHandlers, monitor)

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
func setupRouter(handlers *api.Handlers, monitor *monitoring.Prometheus) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(monitor.GinMiddleware())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		auth := v1.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
			auth.POST("/refresh", handlers.RefreshToken)
			auth.POST("/logout", handlers.Logout)
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(handlers.AuthMiddleware())
		{
			// Network management
			protected.GET("/networks", handlers.ListNetworks)
			protected.POST("/networks", handlers.CreateNetwork)
			protected.GET("/networks/:id", handlers.GetNetwork)
			protected.PUT("/networks/:id", handlers.UpdateNetwork)
			protected.DELETE("/networks/:id", handlers.DeleteNetwork)
			protected.POST("/networks/:id/start", handlers.StartNetwork)
			protected.POST("/networks/:id/stop", handlers.StopNetwork)
			protected.POST("/networks/:id/restart", handlers.RestartNetwork)

			// Node management
			protected.GET("/nodes", handlers.ListNodes)
			protected.POST("/nodes", handlers.CreateNode)
			protected.GET("/nodes/:id", handlers.GetNode)
			protected.PUT("/nodes/:id", handlers.UpdateNode)
			protected.DELETE("/nodes/:id", handlers.DeleteNode)
			protected.POST("/nodes/:id/start", handlers.StartNode)
			protected.POST("/nodes/:id/stop", handlers.StopNode)
			protected.POST("/nodes/:id/restart", handlers.RestartNode)

			// Link management
			protected.GET("/links", handlers.ListLinks)
			protected.POST("/links", handlers.CreateLink)
			protected.GET("/links/:id", handlers.GetLink)
			protected.PUT("/links/:id", handlers.UpdateLink)
			protected.DELETE("/links/:id", handlers.DeleteLink)

			// Policy management
			protected.GET("/policies", handlers.ListPolicies)
			protected.POST("/policies", handlers.CreatePolicy)
			protected.GET("/policies/:id", handlers.GetPolicy)
			protected.PUT("/policies/:id", handlers.UpdatePolicy)
			protected.DELETE("/policies/:id", handlers.DeletePolicy)

			// Monitoring
			monitoring := protected.Group("/monitoring")
			{
				monitoring.GET("/networks/:id/metrics", handlers.GetNetworkMetrics)
				monitoring.GET("/networks/:id/health", handlers.GetNetworkHealth)
				monitoring.GET("/nodes/:id/metrics", handlers.GetNodeMetrics)
				monitoring.GET("/nodes/:id/health", handlers.GetNodeHealth)
				monitoring.GET("/alerts", handlers.ListAlerts)
				monitoring.POST("/alerts/:id/acknowledge", handlers.AcknowledgeAlert)
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

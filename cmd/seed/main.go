package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"

	"netorchestrator/internal/config"
	"netorchestrator/pkg/database"
)

func main() {
	reset := flag.Bool("reset", false, "Reset database before seeding")
	flag.Parse()

	// Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting database seed tool")

	// Initialize database
	db, err := database.NewPostgreSQL(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	logger.Info("Running migrations...")
	if err := database.RunMigrations(db.GetDB(), logger); err != nil {
		log.Fatal("Failed to run migrations:", err)
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
			fmt.Println("\n📊 Database Tables:")
			for _, table := range tables {
				fmt.Printf("  - %s\n", table)
			}
		}
	}

	// Reset database if requested
	if *reset {
		logger.Info("Resetting database...")
		if err := database.ResetDatabase(db.GetDB(), logger); err != nil {
			logger.Error("Failed to reset database", zap.Error(err))
			os.Exit(1)
		}
		fmt.Println("✅ Database reset completed")
	}

	// Seed database
	logger.Info("Seeding database...")
	if err := database.SeedDatabase(db.GetDB(), logger); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	// Get actual row counts after seeding
	networks, nodes, links, err := database.GetRowCounts(db.GetDB(), logger)
	if err != nil {
		logger.Error("Failed to get row counts", zap.Error(err))
		log.Fatal("Failed to get row counts:", err)
	}

	// Print summary with actual counts
	fmt.Println("\n✅ Database seeding completed successfully!")
	fmt.Println("\n📊 Database Row Counts:")
	fmt.Printf("  - Networks: %d\n", networks)
	fmt.Printf("  - Nodes: %d\n", nodes)
	fmt.Printf("  - Links: %d\n", links)
	fmt.Println("")
}


package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"netorchestrator/internal/automation"
)

// Demo script to showcase the NetOrchestrator automation system
func main() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Initialize database (use your actual database config)
	dsn := "host=localhost user=netorchestrator password=yourpassword dbname=netorchestrator port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize automation engine
	engine := automation.NewAutomationEngine(db, logger)

	// Register all executors
	engine.RegisterExecutor("terraform", automation.NewTerraformExecutor(logger))
	engine.RegisterExecutor("ansible", automation.NewAnsibleExecutor(logger))
	engine.RegisterExecutor("api_call", automation.NewAPICallExecutor(logger))
	engine.RegisterExecutor("validation", automation.NewValidationExecutor(logger))
	engine.RegisterExecutor("script", automation.NewScriptExecutor(logger))

	fmt.Println("🚀 NetOrchestrator Automation Engine Demo")
	fmt.Println("==========================================")

	// Create demo workflows
	createDemoWorkflows(engine, logger)

	// Execute a demo workflow
	executeDemoWorkflow(engine, logger)

	fmt.Println("\n✅ Demo completed successfully!")
	fmt.Println("Check the database for workflow execution results")
}

func createDemoWorkflows(engine *automation.AutomationEngine, logger *zap.Logger) {
	fmt.Println("\n📝 Creating demo workflows...")

	// 1. Network Health Check Workflow
	healthCheckWorkflow := &automation.WorkflowDefinition{
		Name:        "Network Health Check",
		Description: "Automated network health monitoring and validation",
		Tasks: []automation.TaskDefinition{
			{
				ID:   "api_health_check",
				Name: "API Health Check",
				Type: "api_call",
				Action: `{
					"url": "https://httpbin.org/status/200",
					"method": "GET",
					"expected_code": 200,
					"timeout": "10s"
				}`,
			},
			{
				ID:           "validate_response",
				Name:         "Validate Health Response",
				Type:         "validation",
				Dependencies: []string{"api_health_check"},
				Action: `{
					"type": "condition",
					"conditions": [
						{"field": "api_response", "operator": "exists", "value": null}
					]
				}`,
			},
		},
		Variables: map[string]interface{}{
			"check_interval": "5m",
			"alert_threshold": 3,
		},
		Schedule: "", // Manual execution for demo
		Enabled:  true,
		Version:  1,
	}

	if err := engine.CreateWorkflow(healthCheckWorkflow); err != nil {
		logger.Error("Failed to create health check workflow", zap.Error(err))
		return
	}
	fmt.Printf("✅ Created workflow: %s (ID: %d)\n", healthCheckWorkflow.Name, healthCheckWorkflow.ID)

	// 2. Infrastructure Provisioning Workflow
	infraWorkflow := &automation.WorkflowDefinition{
		Name:        "Infrastructure Provisioning",
		Description: "Automated infrastructure provisioning with Terraform and configuration with Ansible",
		Tasks: []automation.TaskDefinition{
			{
				ID:   "validate_terraform_config",
				Name: "Validate Terraform Configuration",
				Type: "terraform",
				Action: `{
					"command": "validate",
					"working_dir": "/terraform/demo",
					"variables": {
						"region": "us-west-2",
						"instance_type": "t3.micro"
					}
				}`,
			},
			{
				ID:           "plan_infrastructure",
				Name:         "Plan Infrastructure Changes",
				Type:         "terraform",
				Dependencies: []string{"validate_terraform_config"},
				Action: `{
					"command": "plan",
					"working_dir": "/terraform/demo",
					"auto_approve": false
				}`,
			},
			{
				ID:           "deploy_infrastructure",
				Name:         "Deploy Infrastructure",
				Type:         "terraform",
				Dependencies: []string{"plan_infrastructure"},
				Action: `{
					"command": "apply",
					"working_dir": "/terraform/demo",
					"auto_approve": false
				}`,
			},
			{
				ID:           "configure_services",
				Name:         "Configure Services with Ansible",
				Type:         "ansible",
				Dependencies: []string{"deploy_infrastructure"},
				Action: `{
					"playbook": "/ansible/configure-demo.yml",
					"inventory": "/ansible/inventory/demo",
					"variables": {
						"app_env": "demo",
						"enable_monitoring": true
					}
				}`,
			},
			{
				ID:           "validate_deployment",
				Name:         "Validate Deployment",
				Type:         "api_call",
				Dependencies: []string{"configure_services"},
				Action: `{
					"url": "{{infrastructure_endpoint}}/health",
					"method": "GET",
					"expected_code": 200,
					"retry": 3,
					"retry_delay": "30s"
				}`,
			},
		},
		Variables: map[string]interface{}{
			"environment": "demo",
			"region":      "us-west-2",
		},
		Schedule: "", // Manual execution for demo
		Enabled:  true,
		Version:  1,
	}

	if err := engine.CreateWorkflow(infraWorkflow); err != nil {
		logger.Error("Failed to create infrastructure workflow", zap.Error(err))
		return
	}
	fmt.Printf("✅ Created workflow: %s (ID: %d)\n", infraWorkflow.Name, infraWorkflow.ID)

	// 3. Automated Backup Workflow
	backupWorkflow := &automation.WorkflowDefinition{
		Name:        "Automated Database Backup",
		Description: "Scheduled database backup with validation and cleanup",
		Tasks: []automation.TaskDefinition{
			{
				ID:   "create_backup",
				Name: "Create Database Backup",
				Type: "script",
				Action: `{
					"script": "#!/bin/bash\necho 'Creating database backup...'\necho 'Backup completed successfully'",
					"interpreter": "bash",
					"timeout": "10m"
				}`,
			},
			{
				ID:           "validate_backup",
				Name:         "Validate Backup",
				Type:         "validation",
				Dependencies: []string{"create_backup"},
				Action: `{
					"type": "file_exists",
					"target": "/backups/database_$(date +%Y%m%d).sql"
				}`,
			},
			{
				ID:           "cleanup_old_backups",
				Name:         "Cleanup Old Backups",
				Type:         "script",
				Dependencies: []string{"validate_backup"},
				Action: `{
					"script": "#!/bin/bash\necho 'Cleaning up backups older than 30 days...'\nfind /backups -name '*.sql' -mtime +30 -delete\necho 'Cleanup completed'",
					"interpreter": "bash",
					"timeout": "5m"
				}`,
			},
		},
		Variables: map[string]interface{}{
			"retention_days": 30,
			"backup_path":    "/backups",
		},
		Schedule: "0 2 * * *", // Daily at 2 AM
		Enabled:  false,       // Disabled for demo
		Version:  1,
	}

	if err := engine.CreateWorkflow(backupWorkflow); err != nil {
		logger.Error("Failed to create backup workflow", zap.Error(err))
		return
	}
	fmt.Printf("✅ Created workflow: %s (ID: %d)\n", backupWorkflow.Name, backupWorkflow.ID)
}

func executeDemoWorkflow(engine *automation.AutomationEngine, logger *zap.Logger) {
	fmt.Println("\n🔄 Executing demo workflow...")

	// Find the health check workflow
	var workflow automation.WorkflowDefinition
	if err := engine.DB().Where("name = ?", "Network Health Check").First(&workflow).Error; err != nil {
		logger.Error("Failed to find health check workflow", zap.Error(err))
		return
	}

	// Execute the workflow
	execution, err := engine.ExecuteWorkflow(workflow.ID, map[string]interface{}{
		"demo_mode": true,
		"timeout":   "30s",
	})
	if err != nil {
		logger.Error("Failed to execute workflow", zap.Error(err))
		return
	}

	fmt.Printf("🚀 Started workflow execution (ID: %d)\n", execution.ID)
	fmt.Printf("📊 Workflow: %s\n", workflow.Name)
	fmt.Printf("⏰ Started at: %s\n", execution.StartedAt.Format(time.RFC3339))

	// Monitor execution progress
	fmt.Println("\n📈 Monitoring execution progress...")
	
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		time.Sleep(1 * time.Second)

		// Refresh execution status
		if err := engine.DB().Preload("TaskExecutions").First(execution, execution.ID).Error; err != nil {
			logger.Error("Failed to refresh execution status", zap.Error(err))
			break
		}

		fmt.Printf("⏳ Status: %s | Tasks: %d | Progress: ", execution.Status, len(execution.TaskExecutions))

		for _, task := range execution.TaskExecutions {
			switch task.Status {
			case "completed":
				fmt.Print("✅ ")
			case "failed":
				fmt.Print("❌ ")
			case "running":
				fmt.Print("🔄 ")
			default:
				fmt.Print("⏸️ ")
			}
		}
		fmt.Println()

		if execution.Status == "completed" || execution.Status == "failed" {
			break
		}
	}

	// Show final results
	fmt.Println("\n📋 Execution Summary:")
	fmt.Printf("Status: %s\n", execution.Status)
	if execution.CompletedAt != nil {
		duration := execution.CompletedAt.Sub(execution.StartedAt)
		fmt.Printf("Duration: %s\n", duration)
	}

	if len(execution.TaskExecutions) > 0 {
		fmt.Println("\n📄 Task Results:")
		for _, task := range execution.TaskExecutions {
			fmt.Printf("  • %s: %s", task.Name, task.Status)
			if task.ErrorMessage != "" {
				fmt.Printf(" (Error: %s)", task.ErrorMessage)
			}
			fmt.Println()

			if len(task.Logs) > 0 {
				fmt.Printf("    Logs: %s\n", task.Logs[len(task.Logs)-1])
			}

			if len(task.Output) > 0 {
				outputJson, _ := json.MarshalIndent(task.Output, "    ", "  ")
				fmt.Printf("    Output: %s\n", string(outputJson))
			}
		}
	}

	if execution.ErrorMessage != "" {
		fmt.Printf("\n❌ Execution Error: %s\n", execution.ErrorMessage)
	}
}
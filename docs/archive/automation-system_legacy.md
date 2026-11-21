# NetOrchestrator Automation System

## Overview

The NetOrchestrator Automation System is a comprehensive workflow automation engine designed for enterprise network and infrastructure management. It provides a powerful, scalable platform for orchestrating complex automation tasks with support for multiple execution engines including Terraform, Ansible, API calls, and custom scripts.

## Key Features

### 🚀 **Workflow Orchestration**
- **Multi-step Workflows**: Define complex workflows with task dependencies and parallel execution
- **Conditional Logic**: Support for conditional task execution based on previous results
- **Error Handling**: Comprehensive error handling with retry mechanisms and rollback capabilities
- **Variable Management**: Dynamic variable substitution and context passing between tasks

### 🔧 **Task Executors**
- **Terraform Integration**: Infrastructure-as-Code automation with plan, apply, destroy operations
- **Ansible Integration**: Configuration management with playbook execution and ad-hoc commands
- **API Calls**: HTTP/REST API integration with authentication and retry logic
- **Script Execution**: Shell script execution with multiple interpreter support
- **Validation Engine**: Automated validation of conditions, API responses, and file existence
- **Network Provisioning**: Native network resource management and provisioning

### ⏰ **Scheduling & Automation**
- **Cron-based Scheduling**: Schedule workflows using standard cron expressions
- **Event-driven Execution**: Trigger workflows based on system events and conditions
- **Automatic Retry**: Configurable retry logic with exponential backoff
- **Parallel Execution**: Execute independent tasks simultaneously for optimal performance

### 📊 **Monitoring & Observability**
- **Real-time Status**: Live workflow and task execution monitoring
- **Comprehensive Logging**: Detailed execution logs for debugging and auditing
- **Metrics Collection**: Performance metrics and execution statistics
- **WebSocket Updates**: Real-time status updates through WebSocket connections

### 🛡️ **Enterprise Features**
- **Role-based Access**: Fine-grained permissions and access control
- **Audit Trail**: Complete audit trail of all workflow executions
- **High Availability**: Distributed execution with failover capabilities
- **Scalability**: Horizontal scaling support for high-volume automation

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Automation Engine                         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Workflow   │  │   Task      │  │  Execution  │         │
│  │ Definition  │  │ Executor    │  │   Engine    │         │
│  │             │  │   Manager   │  │             │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│                    Task Executors                           │
├─────────────────────────────────────────────────────────────┤
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│ │Terraform │ │ Ansible  │ │API Calls │ │ Scripts  │   ...  │
│ │Executor  │ │Executor  │ │Executor  │ │Executor  │        │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘        │
├─────────────────────────────────────────────────────────────┤
│                   Infrastructure                            │
├─────────────────────────────────────────────────────────────┤
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐        │
│ │PostgreSQL│ │  Redis   │ │WebSocket │ │Scheduler │        │
│ │Database  │ │  Cache   │ │   Hub    │ │ Service  │        │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## API Endpoints

### Workflow Management
```
POST   /api/v1/automation/workflows              # Create workflow
GET    /api/v1/automation/workflows              # List workflows
GET    /api/v1/automation/workflows/{id}         # Get workflow
PUT    /api/v1/automation/workflows/{id}         # Update workflow
DELETE /api/v1/automation/workflows/{id}         # Delete workflow
```

### Workflow Execution
```
POST   /api/v1/automation/workflows/{id}/execute # Execute workflow
GET    /api/v1/automation/workflows/{id}/executions # List executions
GET    /api/v1/automation/executions/{id}        # Get execution details
POST   /api/v1/automation/executions/{id}/cancel # Cancel execution
```

### Task Management
```
GET    /api/v1/automation/executions/{id}/tasks  # Get execution tasks
GET    /api/v1/automation/tasks/{id}             # Get task details
```

### Templates & Discovery
```
GET    /api/v1/automation/templates              # List workflow templates
GET    /api/v1/automation/templates/{name}       # Get specific template
GET    /api/v1/automation/executors              # List available executors
```

## Workflow Definition

### Basic Structure
```json
{
  "name": "Infrastructure Deployment",
  "description": "Deploy and configure infrastructure",
  "tasks": [
    {
      "id": "validate_config",
      "name": "Validate Configuration",
      "type": "validation",
      "action": "{...}"
    },
    {
      "id": "deploy_infra",
      "name": "Deploy Infrastructure",
      "type": "terraform",
      "dependencies": ["validate_config"],
      "action": "{...}"
    }
  ],
  "variables": {
    "environment": "production",
    "region": "us-west-2"
  },
  "schedule": "0 2 * * 1",
  "enabled": true
}
```

### Task Types

#### Terraform Tasks
```json
{
  "type": "terraform",
  "action": {
    "command": "apply",
    "working_dir": "/terraform/infrastructure",
    "variables": {
      "instance_count": 3,
      "region": "us-west-2"
    },
    "auto_approve": false,
    "timeout": "30m"
  }
}
```

#### Ansible Tasks
```json
{
  "type": "ansible",
  "action": {
    "playbook": "/ansible/deploy-app.yml",
    "inventory": "/ansible/inventory/production",
    "variables": {
      "app_version": "v2.1.0",
      "enable_ssl": true
    },
    "timeout": "15m"
  }
}
```

#### API Call Tasks
```json
{
  "type": "api_call",
  "action": {
    "url": "https://api.example.com/deploy",
    "method": "POST",
    "headers": {
      "Content-Type": "application/json"
    },
    "body": {
      "environment": "{{environment}}",
      "version": "{{app_version}}"
    },
    "auth": {
      "type": "bearer",
      "token": "{{api_token}}"
    },
    "expected_code": 200,
    "retry": 3
  }
}
```

#### Validation Tasks
```json
{
  "type": "validation",
  "action": {
    "type": "api_response",
    "target": "https://app.example.com/health",
    "conditions": [
      {
        "field": "status_code",
        "operator": "eq",
        "value": 200
      }
    ]
  }
}
```

## Quick Start

### 1. Database Setup
```sql
-- Create automation engine tables
-- (Automatic migration will handle this)
```

### 2. Basic Usage
```go
package main

import (
    "netorchestrator/internal/automation"
)

func main() {
    // Initialize engine
    engine := automation.NewAutomationEngine(db, logger)
    
    // Register executors
    engine.RegisterExecutor("terraform", automation.NewTerraformExecutor(logger))
    engine.RegisterExecutor("ansible", automation.NewAnsibleExecutor(logger))
    
    // Create workflow
    workflow := &automation.WorkflowDefinition{
        Name: "Deploy Application",
        Tasks: []automation.TaskDefinition{
            {
                ID: "deploy",
                Name: "Deploy Infrastructure",
                Type: "terraform",
                Action: `{"command": "apply", "working_dir": "/terraform"}`,
            },
        },
    }
    
    engine.CreateWorkflow(workflow)
    
    // Execute workflow
    execution, err := engine.ExecuteWorkflow(workflow.ID, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Execution started: %d\n", execution.ID)
}
```

### 3. REST API Usage
```bash
# Create a workflow
curl -X POST http://localhost:8080/api/v1/automation/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Health Check",
    "description": "API health monitoring",
    "tasks": [
      {
        "id": "check_api",
        "name": "Check API Health",
        "type": "api_call",
        "action": "{\"url\": \"https://api.example.com/health\", \"method\": \"GET\"}"
      }
    ]
  }'

# Execute workflow
curl -X POST http://localhost:8080/api/v1/automation/workflows/1/execute \
  -H "Content-Type: application/json" \
  -d '{"variables": {"timeout": "30s"}}'

# Check execution status
curl http://localhost:8080/api/v1/automation/executions/1
```

## Workflow Templates

### Network Provisioning
```json
{
  "name": "network_provisioning",
  "description": "Automated network provisioning workflow",
  "tasks": [
    {
      "id": "validate_network_config",
      "name": "Validate Network Configuration",
      "type": "validation",
      "action": {
        "type": "condition",
        "conditions": [
          {"field": "network_cidr", "operator": "exists"},
          {"field": "subnet_count", "operator": "gt", "value": 0}
        ]
      }
    },
    {
      "id": "provision_vpc",
      "name": "Provision VPC",
      "type": "terraform",
      "dependencies": ["validate_network_config"],
      "action": {
        "command": "apply",
        "working_dir": "/terraform/network",
        "variables": {
          "cidr_block": "{{network_cidr}}",
          "availability_zones": "{{az_list}}"
        }
      }
    },
    {
      "id": "configure_routing",
      "name": "Configure Network Routing",
      "type": "ansible",
      "dependencies": ["provision_vpc"],
      "action": {
        "playbook": "/ansible/configure-routing.yml",
        "inventory": "dynamic"
      }
    },
    {
      "id": "test_connectivity",
      "name": "Test Network Connectivity",
      "type": "script",
      "dependencies": ["configure_routing"],
      "action": {
        "script": "/scripts/network-connectivity-test.sh",
        "timeout": "5m"
      }
    }
  ]
}
```

### Service Deployment
```json
{
  "name": "service_deployment",
  "description": "Deploy microservices with infrastructure",
  "tasks": [
    {
      "id": "deploy_infrastructure",
      "name": "Deploy Infrastructure",
      "type": "terraform",
      "action": {
        "command": "apply",
        "working_dir": "/terraform/services"
      }
    },
    {
      "id": "configure_load_balancer",
      "name": "Configure Load Balancer",
      "type": "ansible",
      "dependencies": ["deploy_infrastructure"],
      "action": {
        "playbook": "/ansible/configure-lb.yml"
      }
    },
    {
      "id": "deploy_application",
      "name": "Deploy Application",
      "type": "api_call",
      "dependencies": ["configure_load_balancer"],
      "action": {
        "url": "{{deployment_api}}/deploy",
        "method": "POST",
        "body": {
          "service": "{{service_name}}",
          "version": "{{version}}"
        }
      }
    },
    {
      "id": "health_check",
      "name": "Application Health Check",
      "type": "validation",
      "dependencies": ["deploy_application"],
      "action": {
        "type": "api_response",
        "target": "{{service_url}}/health",
        "conditions": [
          {"field": "status_code", "operator": "eq", "value": 200}
        ]
      }
    }
  ]
}
```

## Advanced Features

### Conditional Execution
```json
{
  "id": "conditional_task",
  "name": "Deploy to Production",
  "type": "terraform",
  "condition": "{{environment}} == 'production' && {{tests_passed}} == true",
  "action": {...}
}
```

### Error Handling
```json
{
  "id": "deploy_with_rollback",
  "name": "Deploy with Rollback",
  "type": "terraform",
  "retry": {
    "max_attempts": 3,
    "delay": "30s",
    "backoff": "exponential"
  },
  "on_failure": {
    "action": "rollback",
    "notify": ["ops-team@company.com"]
  },
  "action": {...}
}
```

### Parallel Execution
```json
{
  "tasks": [
    {
      "id": "deploy_frontend",
      "name": "Deploy Frontend",
      "type": "api_call",
      "parallel_group": "deployment"
    },
    {
      "id": "deploy_backend",
      "name": "Deploy Backend",
      "type": "api_call",
      "parallel_group": "deployment"
    },
    {
      "id": "update_dns",
      "name": "Update DNS",
      "type": "script",
      "dependencies": ["deploy_frontend", "deploy_backend"]
    }
  ]
}
```

## Monitoring & Troubleshooting

### Execution Monitoring
- Real-time execution status via WebSocket
- Detailed task logs and execution traces
- Performance metrics and timing information
- Error details and stack traces

### Common Issues
1. **Task Timeout**: Increase timeout values for long-running operations
2. **Dependency Cycles**: Ensure no circular dependencies in task definitions
3. **Variable Resolution**: Check variable scoping and substitution
4. **Executor Errors**: Verify executor configuration and permissions

### Best Practices
1. **Workflow Design**: Keep workflows focused and modular
2. **Error Handling**: Implement comprehensive error handling and rollback procedures
3. **Testing**: Test workflows in development environments first
4. **Security**: Use secure credential management and access controls
5. **Monitoring**: Implement comprehensive monitoring and alerting

## Contributing

The automation system is designed to be extensible. To add new task executors:

1. Implement the `TaskExecutor` interface
2. Register the executor with the engine
3. Create appropriate task action schemas
4. Add comprehensive tests and documentation

## License

Apache 2.0 - See LICENSE file for details
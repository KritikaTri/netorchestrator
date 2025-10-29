# NetOrchestrator API Documentation

## Overview

NetOrchestrator is an enterprise-grade Network-as-a-Service (NaaS) automation platform that provides comprehensive network orchestration, monitoring, and management capabilities through a modern microservices architecture.

## API Features

### 🏗️ **Network Management**
- **Create, Update, Delete Networks**: Full CRUD operations for network lifecycle management
- **Network Discovery**: Automatic topology detection and mapping
- **Multi-tenancy Support**: Isolated network environments per organization
- **CIDR Management**: Intelligent IP address allocation and subnet management

### 📊 **Monitoring & Analytics**
- **Real-time Metrics**: Performance monitoring with sub-second granularity
- **Historical Analytics**: Time-series data storage and analysis
- **Custom Dashboards**: Configurable monitoring interfaces
- **Alerting System**: Proactive notification and escalation

### 🛡️ **Security & Policies**
- **Policy Engine**: Rule-based network access control
- **Compliance Monitoring**: Automated security posture assessment
- **Audit Logging**: Comprehensive activity tracking
- **Role-based Access**: Granular permission management

### 🔧 **Infrastructure Automation**
- **Event-driven Architecture**: Asynchronous processing and workflows
- **Service Discovery**: Dynamic microservice registration and health checks
- **Circuit Breakers**: Fault tolerance and resilience patterns
- **Load Balancing**: Intelligent traffic distribution

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │────│  Load Balancer  │────│  Web Dashboard  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
    ┌────────────────────────────┼────────────────────────────┐
    │                            │                            │
┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐
│Network  │  │Monitor  │  │Policy   │  │Event    │  │Registry │
│Service  │  │Service  │  │Service  │  │Bus      │  │Service  │
└─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘
     │            │            │            │            │
     └────────────┼────────────┼────────────┼────────────┘
                  │            │            │
       ┌──────────┼────────────┼────────────┼──────────┐
       │          │            │            │          │
┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│ PostgreSQL  │  │    Redis    │  │  InfluxDB   │  │ Prometheus  │
│ (Primary)   │  │ (Caching)   │  │ (Metrics)   │  │(Monitoring) │
└─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘
```

## Getting Started

### Prerequisites
- Go 1.21+
- Podman or Docker
- PostgreSQL 14+
- Redis 6+

### Quick Start

1. **Clone the repository**
```bash
git clone https://github.com/your-org/netorchestrator.git
cd netorchestrator
```

2. **Start infrastructure services**
```bash
podman-compose up -d
```

3. **Run the API Gateway**
```bash
go run cmd/api-gateway/main.go
```

4. **Access the API documentation**
- Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- API Health: [http://localhost:8080/api/v1/health](http://localhost:8080/api/v1/health)

## API Endpoints

### Health Check
```http
GET /api/v1/health
```

### Networks
```http
GET    /api/v1/networks       # List all networks
POST   /api/v1/networks       # Create network
GET    /api/v1/networks/{id}  # Get network details
PUT    /api/v1/networks/{id}  # Update network
DELETE /api/v1/networks/{id}  # Delete network
```

### Nodes
```http
GET    /api/v1/nodes          # List all nodes
POST   /api/v1/nodes          # Create node
GET    /api/v1/nodes/{id}     # Get node details
PUT    /api/v1/nodes/{id}     # Update node
DELETE /api/v1/nodes/{id}     # Delete node
```

### Policies
```http
GET    /api/v1/policies       # List all policies
POST   /api/v1/policies       # Create policy
GET    /api/v1/policies/{id}  # Get policy details
PUT    /api/v1/policies/{id}  # Update policy
DELETE /api/v1/policies/{id}  # Delete policy
```

### Metrics
```http
GET    /api/v1/metrics        # Get performance metrics
GET    /api/v1/metrics/summary # Get metrics summary
```

### Events
```http
GET    /api/v1/events         # List system events
POST   /api/v1/events         # Create event
```

## Authentication

The API supports multiple authentication methods:

- **API Key**: Include `Authorization: Bearer <token>` header
- **JWT**: JSON Web Tokens for session-based auth
- **mTLS**: Mutual TLS for service-to-service communication

## Rate Limiting

- **Default**: 1000 requests per minute per client
- **Burst**: Up to 100 requests in 10 seconds
- **Headers**: Rate limit info included in response headers

## Error Handling

All errors follow a consistent format:

```json
{
  "error": "Descriptive error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "validation_error",
    "timestamp": "2024-01-10T12:00:00Z"
  }
}
```

## SDK and Client Libraries

- **Go**: Native Go client library
- **Python**: PyPI package `netorchestrator-client`
- **JavaScript**: NPM package `@netorchestrator/client`
- **CLI**: Command-line interface tool

## Support

- **Documentation**: [https://docs.netorchestrator.io](https://docs.netorchestrator.io)
- **GitHub Issues**: [https://github.com/your-org/netorchestrator/issues](https://github.com/your-org/netorchestrator/issues)
- **Discord**: [https://discord.gg/netorchestrator](https://discord.gg/netorchestrator)
- **Email**: [support@netorchestrator.io](mailto:support@netorchestrator.io)

---

**Version**: 2.0.0 | **Last Updated**: January 2024 | **License**: Apache 2.0
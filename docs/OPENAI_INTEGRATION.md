# OpenAI Integration Guide

## Overview

NetOrchestrator now supports real OpenAI integration for enhanced AI-powered network analysis and natural language processing capabilities.

## Setup Instructions

### 1. Set Your OpenAI API Key

Add your OpenAI API key to your environment:

```bash
export OPENAI_API_KEY="sk-your-api-key-here"
export OPENAI_MODEL="gpt-4"  # Optional, defaults to gpt-4
```

### 2. Features Enhanced with OpenAI

#### AI-Powered Network Analysis
- **Real LLM Analysis**: Network performance analysis now uses GPT-4 instead of mock responses
- **Intelligent Insights**: Root cause analysis and recommendations powered by OpenAI
- **Context-Aware Responses**: Memory and conversation history for better analysis

#### Natural Language Processing (NLP)
- **Enhanced Parsing**: Better understanding of complex network requests
- **Validation**: Automatic validation against supported network components
- **Fallback Support**: Falls back to rule-based parsing if OpenAI is unavailable

### 3. Supported Network Components

The system strictly validates against these supported components:

#### Node Types
- `router` - Network routing devices
- `switch` - Network switching devices  
- `host` - End-user devices/servers
- `firewall` - Security appliances
- `load_balancer` - Load balancing devices
- `gateway` - Network gateways

#### Network Topologies
- `star` - Hub and spoke (min 2 nodes, requires router/switch as hub)
- `mesh` - Full/partial mesh (min 3 nodes)
- `tree` - Hierarchical tree (min 3 nodes)
- `custom` - Custom topology

#### Policy Types
- `traffic` - Traffic management policies
- `security` - Security policies
- `firewall` - Firewall rules (actions: allow, deny, drop)
- `qos` - Quality of Service policies
- `routing` - Routing policies
- `access` - Access control policies

### 4. API Endpoints Enhanced

#### Network Intelligence
- `POST /api/v1/intelligence/analyze/network` - AI-powered network analysis
- `POST /api/v1/intelligence/predict/capacity` - Capacity forecasting
- `POST /api/v1/intelligence/optimize/costs` - Cost optimization
- `POST /api/v1/intelligence/detect/anomalies` - Anomaly detection

#### Natural Language Provisioning
- `POST /api/v1/ai/provision` - Create networks from natural language

### 5. Example NLP Requests

#### Valid Requests
```
"Create a star network with 5 routers and 2 firewalls"
"Build a mesh topology with 4 switches for high availability"  
"Setup a tree network with 1 gateway, 3 routers, and 6 hosts"
"Create a firewall policy that allows HTTP and blocks Telnet"
```

#### Invalid Requests (Will be Rejected)
```
"Create a network with 100 nodes" (exceeds 50 node limit)
"Build a mesh with 1 node" (violates mesh topology constraints)
"Add a Kubernetes cluster" (unsupported node type)
"Create a ring topology" (ring not in supported topologies)
```

### 6. Response Format

Enhanced responses include validation information:

```json
{
  "success": true,
  "network": {...},
  "validation": {
    "valid": true,
    "confidence": 0.92,
    "errors": []
  },
  "parsing_stats": {
    "ai_enhanced": true,
    "topology": "star",
    "nodes_parsed": 7,
    "policies": 1
  }
}
```

### 7. Error Handling

Invalid requests return detailed validation errors:

```json
{
  "error": "Invalid network specification",
  "validation_errors": [
    "Unsupported node type 'kubernetes'. Supported types: router, switch, host, firewall, load_balancer, gateway",
    "Star topology requires at least one router or switch as a hub"
  ],
  "suggestions": "Use supported node types and ensure topology constraints are met"
}
```

### 8. Fallback Behavior

- If OpenAI API key is not configured, the system falls back to mock responses
- If OpenAI API fails, NLP falls back to rule-based parsing
- All validation constraints are enforced regardless of parsing method

### 9. Cost Management

To manage OpenAI API costs:
- Set reasonable token limits in configuration
- Monitor usage through OpenAI dashboard
- Consider using gpt-3.5-turbo for cost savings (set OPENAI_MODEL)
- System includes proper error handling and timeouts

### 10. Security Considerations

- API keys are loaded from environment variables only
- No API keys are logged or exposed in responses
- Requests are validated before sending to OpenAI
- Proper timeout and error handling prevent API abuse

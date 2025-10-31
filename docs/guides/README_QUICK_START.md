# ⚡ NetOrchestrator - 60 Second Setup

## Get NetOrchestrator running in under 60 seconds

### 🏃‍♂️ Ultra Quick Start
```bash
# command to run
./setup-complete.sh
```

# command to stop everything
./stop-and-clean.sh

### 🧪 Test Everything Works  
```bash
# Basic test (30 seconds)
./quick-test.sh

# Full AI test (60 seconds)  
./test-ai-complete.sh
```

### 🛑 Stop Everything
```bash
# Stop all services (keeps data)
./stop-all.sh

# Stop and completely clean up (removes all data)
./stop-and-clean.sh
```

### 🎯 That's it! 

Your NetOrchestrator platform is now running with:

✅ **Enterprise Authentication** - JWT tokens, user management  
✅ **AI Intelligence Suite** - 4 ML-powered analysis engines  
✅ **Network Management** - Create, manage virtual networks  
✅ **Real-time Monitoring** - Alerts, metrics, health checks  
✅ **Production Database** - PostgreSQL + Redis + InfluxDB  
✅ **API Documentation** - Interactive Swagger UI  

---

### 🌐 Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| **API Platform** | http://localhost:8080 | - |
| **API Documentation** | http://localhost:8080/swagger/index.html | - |
| **Health Check** | http://localhost:8080/health | - |
| **Database** | postgresql://netorchestrator:password@localhost:5432/netorchestrator | netorchestrator / password |

### 🔐 Demo Accounts

| Username | Password | Access Level |
|----------|----------|-------------|
| `admin` | `admin123` | Full administrative access |  
| `user` | `user123` | Standard user access |

---

### 🚀 What You Can Do Now

1. **Create Networks**: `POST /api/v1/networks`
2. **AI Network Analysis**: `POST /api/v1/intelligence/analyze/network` 
3. **Capacity Planning**: `POST /api/v1/intelligence/predict/capacity`
4. **Cost Optimization**: `POST /api/v1/intelligence/optimize/costs`
5. **Anomaly Detection**: `POST /api/v1/intelligence/detect/anomalies`

### 📖 Full Documentation

- **Complete Setup Guide**: [`STARTUP.md`](STARTUP.md)
- **Architecture Overview**: [`ARCHITECTURE.md`](ARCHITECTURE.md)
- **Platform Overview**: [`PLATFORM_OVERVIEW.md`](PLATFORM_OVERVIEW.md)

---

**NetOrchestrator: Enterprise AI-Powered Network Orchestration Platform** 🚀

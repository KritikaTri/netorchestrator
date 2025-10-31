# 🧪 Postman Testing Guide for NetOrchestrator API

## 📥 **Step 1: Import the Collection**

1. Open **Postman**
2. Click **Import** (top left)
3. Choose **File** and select: `NetOrchestrator_Postman_Collection.json`
4. Collection will appear in your left sidebar

---

## 🚀 **Step 2: Quick Start - Test in Order**

### **Test Sequence (Follow this order):**

#### **1. Health Check** (No auth needed)
```
GET http://localhost:8080/health
```
**Expected Response:**
```json
{
  "status": "healthy",
  "service": "netorchestrator-api-gateway",
  "timestamp": "2025-10-30T...",
  "version": "2.0.0"
}
```

---

#### **2. Login (Get Token)** ⭐ **START HERE**
```
POST http://localhost:8080/api/v1/auth/login
```
**Body:**
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**Expected Response:**
```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 86400,
  "user": {
    "id": "admin-uuid",
    "username": "admin",
    "email": "admin@netorchestrator.com",
    "roles": ["admin", "user"]
  }
}
```

✅ **The token is automatically saved to collection variables!**

---

#### **3. Get Profile** (Uses saved token)
```
GET http://localhost:8080/api/v1/auth/profile
```

**Expected Response:**
```json
{
  "id": "admin-uuid",
  "username": "admin",
  "email": "admin@netorchestrator.com",
  "roles": ["admin", "user"]
}
```

---

#### **4. Create Network**
```
POST http://localhost:8080/api/v1/networks
```
**Body:**
```json
{
  "name": "postman-test-network",
  "subnet": "10.100.0.0/24",
  "description": "Network created from Postman for testing"
}
```

**Expected Response:**
```json
{
  "network": {
    "id": "7c483659-...",
    "name": "postman-test-network",
    "subnet": "10.100.0.0/24",
    "status": "pending",
    "created_at": "2025-10-30T..."
  }
}
```

✅ **Network ID is automatically saved!**

---

#### **5. List Networks**
```
GET http://localhost:8080/api/v1/networks
```

**Expected Response:**
```json
{
  "networks": [
    {
      "id": "...",
      "name": "postman-test-network",
      "subnet": "10.100.0.0/24",
      "status": "pending"
    }
  ],
  "count": 1
}
```

---

#### **6. Create User**
```
POST http://localhost:8080/api/v1/users
```
**Body:**
```json
{
  "username": "postman-user",
  "email": "postman@example.com",
  "password": "SecurePass123!"
}
```

**Expected Response:**
```json
{
  "message": "User created successfully",
  "user": {
    "id": "generated-uuid",
    "username": "postman-user",
    "email": "postman@example.com",
    "roles": ["user"],
    "status": "active"
  }
}
```

---

#### **7. Create Alert Rule**
```
POST http://localhost:8080/api/v1/monitoring/alert-rules
```
**Body:**
```json
{
  "name": "High CPU Alert",
  "description": "Alert when CPU exceeds 80%",
  "condition": "cpu_usage > 80",
  "severity": "warning",
  "enabled": true
}
```

**Expected Response:**
```json
{
  "message": "Alert rule created successfully",
  "rule": {
    "id": "generated-id",
    "name": "High CPU Alert",
    "enabled": true
  }
}
```

---

#### **8. AI Network Analysis**
```
POST http://localhost:8080/api/v1/intelligence/analyze
```
**Body:**
```json
{
  "metrics": {
    "cpu_usage": 75.5,
    "memory_usage": 60.2,
    "network_throughput": 850,
    "latency": 25
  },
  "time_range": "1h"
}
```

**Expected Response:**
```json
{
  "success": true,
  "data": { ... },
  "insights": [
    "Network latency shows 15% improvement opportunity",
    "Bandwidth utilization indicates potential bottlenecks during peak hours",
    "CPU usage patterns suggest auto-scaling configuration optimization"
  ],
  "confidence": 0.92,
  "processing_time": "234.5ms",
  "metadata": {
    "analysis_type": "network_performance",
    "model_version": "v2.1",
    "features_analyzed": 4
  }
}
```

---

#### **9. AI Cost Optimization**
```
POST http://localhost:8080/api/v1/intelligence/optimize
```
**Body:**
```json
{
  "resources": {
    "instances": 10,
    "storage_gb": 500,
    "bandwidth_gb": 1000
  },
  "constraints": ["maintain_performance"],
  "objectives": ["reduce_cost", "improve_efficiency"]
}
```

**Expected Response:**
```json
{
  "success": true,
  "insights": [
    "Potential monthly savings: $15,000 (32% cost reduction)",
    "Right-size 8 over-provisioned instances → Save $8,500/month",
    "Switch to reserved instances for stable workloads → Save $4,200/month"
  ],
  "confidence": 0.89,
  "metadata": {
    "optimization_type": "multi_objective",
    "roi_estimate": "285%"
  }
}
```

---

#### **10. Get AI Insights**
```
GET http://localhost:8080/api/v1/intelligence/insights?category=cost&limit=5
```

**Expected Response:**
```json
{
  "insights": [
    {
      "title": "Reserved Instance Optimization",
      "description": "Switch to reserved instances for 70% cost savings",
      "impact": "high",
      "confidence": 0.95,
      "actions": ["Purchase 3-year RIs", "Right-size instances"],
      "estimated_benefit": "$180,000/year"
    }
  ],
  "category": "cost",
  "count": 2,
  "generated_at": "2025-10-30T..."
}
```

---

## 🔍 **Step 3: What to Look For**

### **✅ Success Indicators:**

1. **HTTP Status Codes:**
   - `200 OK` - Request successful
   - `201 Created` - Resource created
   - `401 Unauthorized` - Token issue (re-login)

2. **Response Structure:**
   - All responses should have proper JSON
   - No HTML error pages
   - Clear error messages if something fails

3. **Token Auto-Save:**
   - After login, check **Collection Variables**
   - You should see `access_token` filled
   - All subsequent requests use this automatically

4. **Data Persistence:**
   - Created networks should appear in "List Networks"
   - Network IDs should be valid UUIDs
   - Created users should appear in "List Users"

---

## ❌ **Troubleshooting**

### **Issue: 401 Unauthorized**
**Solution:** 
1. Run "1.1 Login (Get Token)" first
2. Check that token is saved in Collection Variables
3. Verify Authorization header is set to `Bearer {{access_token}}`

### **Issue: Connection Refused**
**Solution:**
```bash
# Check if API is running
curl http://localhost:8080/health

# If not running, restart:
cd /Users/kritripa/netorchestrator-1
./api-gateway
```

### **Issue: 404 Not Found**
**Solution:** Verify the endpoint URL matches exactly

### **Issue: Token expired**
**Solution:** Re-run "1.1 Login" to get a fresh token

---

## 🎯 **Essential Tests (Minimum)**

If you're short on time, test these 5 endpoints:

1. **Health Check** - Verify API is running
2. **Login** - Get authentication token
3. **Get Profile** - Verify token works
4. **Create Network** - Test core functionality
5. **AI Insights** - Test new AI features

---

## 📊 **Testing Checklist**

- [ ] Health check returns 200
- [ ] Login returns valid JWT token
- [ ] Profile returns user data
- [ ] Create network returns network ID
- [ ] List networks shows created network
- [ ] Create user succeeds
- [ ] Create alert rule succeeds
- [ ] AI analysis returns insights
- [ ] AI cost optimization returns recommendations
- [ ] Get AI insights returns multiple insights

---

## 💡 **Pro Tips**

1. **Environment Variables:** You can create a Postman environment for different deployments (dev, staging, prod)

2. **Test Scripts:** The collection includes auto-save scripts for tokens and IDs

3. **Categories:** You can test AI insights with different categories:
   - `performance`
   - `cost`
   - `security`
   - `capacity`

4. **Bulk Testing:** Use Postman's Collection Runner to test all endpoints at once

5. **Export Results:** You can export test results from Postman for documentation

---

## 🔐 **Security Notes**

- Default credentials: `admin` / `admin123`
- These are for testing only
- Change passwords before production deployment
- API keys should be rotated regularly

---

## 📝 **Next Steps After Testing**

Once all Postman tests pass:

1. ✅ Confirm all 67 endpoints work
2. ✅ Verify response structures are correct
3. ✅ Check data persistence
4. ✅ Move forward with database layer implementation

---

**Happy Testing! 🚀**


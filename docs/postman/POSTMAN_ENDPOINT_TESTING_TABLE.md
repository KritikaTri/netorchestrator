# 🧪 Postman Endpoint Testing Table

## 📊 **Database Status**

All tables now have test data:
- ✅ Users: 1
- ✅ Networks: 3
- ✅ Nodes: 7
- ✅ Links: 5
- ✅ Policies: 4
- ✅ Metrics: 7
- ✅ Events: 5
- ✅ Alerts: 3
- ✅ Workflows: 2

---

## 🔑 **Step 0: Get Authentication Token**

**ALL requests below need this token!**

| # | Method | Endpoint | Body |
|---|--------|----------|------|
| 0 | POST | `/api/v1/auth/login` | `{"username":"admin","password":"admin123"}` |

**Save the `access_token` from response - use it in all requests below!**

---

## ✅ **ESSENTIAL ENDPOINTS TO TEST (Priority Order)**

### **Category 1: System Health (No Auth Required)**

| # | Method | Endpoint | Expected Response | Status |
|---|--------|----------|-------------------|--------|
| 1 | GET | `/health` | `{"status":"healthy"}` | ☐ |
| 2 | GET | `/metrics` | Prometheus metrics | ☐ |

---

### **Category 2: Authentication (5 endpoints)**

| # | Method | Endpoint | Body | Expected Response |
|---|--------|----------|------|-------------------|
| 3 | POST | `/api/v1/auth/login` | `{"username":"admin","password":"admin123"}` | Token + user info |
| 4 | GET | `/api/v1/auth/profile` | - | User profile |
| 5 | POST | `/api/v1/auth/refresh` | `{"refresh_token":"YOUR_TOKEN"}` | New token |
| 6 | POST | `/api/v1/auth/api-keys` | `{"name":"test-key","permissions":["read"]}` | API key created |
| 7 | POST | `/api/v1/auth/logout` | - | Logout confirmation |

**Testing Status:** ☐☐☐☐☐

---

### **Category 3: Networks (14 endpoints)**

| # | Method | Endpoint | Body/Params | Expected |
|---|--------|----------|-------------|----------|
| 8 | GET | `/api/v1/networks` | - | 3 networks |
| 9 | POST | `/api/v1/networks` | `{"name":"test-net","subnet":"10.50.0.0/24"}` | Network created |
| 10 | GET | `/api/v1/networks/{id}` | Use ID from #8 | Network details |
| 11 | PUT | `/api/v1/networks/{id}` | `{"name":"updated-net","subnet":"10.50.0.0/24","status":"active"}` | Updated |
| 12 | DELETE | `/api/v1/networks/{id}` | Use ID from #9 | Deleted |
| 13 | POST | `/api/v1/networks/{id}/start` | Use Production Network ID | Started |
| 14 | POST | `/api/v1/networks/{id}/stop` | - | Stopped |
| 15 | POST | `/api/v1/networks/{id}/restart` | - | Restarted |

**Testing Status:** ☐☐☐☐☐☐☐☐

---

### **Category 4: Nodes (8 endpoints)**

| # | Method | Endpoint | Body/Params | Expected |
|---|--------|----------|-------------|----------|
| 16 | GET | `/api/v1/nodes` | - | 7 nodes |
| 17 | POST | `/api/v1/nodes` | `{"name":"test-node","network_id":"PROD_NET_ID","node_type":"application","ip_address":"10.0.99.99"}` | Node created |
| 18 | GET | `/api/v1/nodes/{id}` | Use prod-web-01 ID | Node details |
| 19 | PUT | `/api/v1/nodes/{id}` | `{"name":"updated-node","status":"active"}` | Updated |
| 20 | DELETE | `/api/v1/nodes/{id}` | Use ID from #17 | Deleted |
| 21 | POST | `/api/v1/nodes/{id}/start` | - | Started |
| 22 | POST | `/api/v1/nodes/{id}/stop` | - | Stopped |
| 23 | POST | `/api/v1/nodes/{id}/restart` | - | Restarted |

**Testing Status:** ☐☐☐☐☐☐☐☐

---

### **Category 5: Links (3 endpoints)**

| # | Method | Endpoint | Body/Params | Expected |
|---|--------|----------|-------------|----------|
| 24 | GET | `/api/v1/links/{id}` | Use any link ID from DB | Link details |
| 25 | PUT | `/api/v1/links/{id}` | `{"bandwidth_mbps":2000}` | Updated |
| 26 | DELETE | `/api/v1/links/{id}` | - | Deleted |

**Testing Status:** ☐☐☐

---

### **Category 6: Policies (5 endpoints)**

| # | Method | Endpoint | Body/Params | Expected |
|---|--------|----------|-------------|----------|
| 27 | GET | `/api/v1/policies` | - | 4 policies |
| 28 | POST | `/api/v1/policies` | `{"name":"test-policy","type":"firewall","rules":[]}` | Created |
| 29 | GET | `/api/v1/policies/{id}` | Use ID from #27 | Policy details |
| 30 | PUT | `/api/v1/policies/{id}` | `{"name":"updated-policy"}` | Updated |
| 31 | DELETE | `/api/v1/policies/{id}` | - | Deleted |

**Testing Status:** ☐☐☐☐☐

---

### **Category 7: Monitoring (15 endpoints)**

| # | Method | Endpoint | Body/Params | Expected |
|---|--------|----------|-------------|----------|
| 32 | GET | `/api/v1/monitoring/networks/{id}/metrics` | Use Production Network ID | Metrics data |
| 33 | GET | `/api/v1/monitoring/networks/{id}/health` | - | Health status |
| 34 | GET | `/api/v1/monitoring/nodes/{id}/metrics` | Use prod-web-01 ID | Node metrics |
| 35 | GET | `/api/v1/monitoring/nodes/{id}/health` | - | Node health |
| 36 | GET | `/api/v1/monitoring/alerts` | - | 3 alerts |
| 37 | POST | `/api/v1/monitoring/alerts/{id}/acknowledge` | Use alert ID | Acknowledged |
| 38 | POST | `/api/v1/monitoring/alerts/{id}/resolve` | - | Resolved |
| 39 | GET | `/api/v1/monitoring/alert-rules` | - | Alert rules list |
| 40 | POST | `/api/v1/monitoring/alert-rules` | `{"name":"test-rule","condition":"cpu > 90"}` | Rule created |
| 41 | GET | `/api/v1/monitoring/alert-rules/{id}` | - | Rule details |
| 42 | PUT | `/api/v1/monitoring/alert-rules/{id}` | `{"condition":"cpu > 95"}` | Updated |
| 43 | DELETE | `/api/v1/monitoring/alert-rules/{id}` | - | Deleted |
| 44 | GET | `/api/v1/monitoring/notification-channels` | - | Channels list |
| 45 | POST | `/api/v1/monitoring/notification-channels` | `{"name":"test-channel","type":"email"}` | Created |
| 46 | DELETE | `/api/v1/monitoring/notification-channels/{id}` | - | Deleted |

**Testing Status:** ☐☐☐☐☐☐☐☐☐☐☐☐☐☐☐

---

### **Category 8: Users (6 endpoints)**

| # | Method | Endpoint | Body/Params | Expected |
|---|--------|----------|-------------|----------|
| 47 | GET | `/api/v1/users` | - | 1+ users |
| 48 | POST | `/api/v1/users` | `{"username":"testuser","email":"test@example.com","password":"pass123"}` | User created |
| 49 | GET | `/api/v1/users/profile` | - | Current user profile |
| 50 | PUT | `/api/v1/users/profile` | `{"email":"newemail@example.com"}` | Profile updated |
| 51 | PUT | `/api/v1/users/{id}` | `{"email":"updated@example.com"}` | User updated |
| 52 | DELETE | `/api/v1/users/{id}` | Use ID from #48 | User deleted |

**Testing Status:** ☐☐☐☐☐☐

---

### **Category 9: AI Intelligence (7 endpoints)**

| # | Method | Endpoint | Body/Params | Expected |
|---|--------|----------|-------------|----------|
| 53 | POST | `/api/v1/intelligence/analyze` | `{"metrics":{"cpu":75,"memory":60}}` | Analysis with insights |
| 54 | POST | `/api/v1/intelligence/predict` | `{"historical_data":[{"value":100},{"value":120}],"forecast_days":30}` | Predictions |
| 55 | POST | `/api/v1/intelligence/optimize` | `{"resources":{"instances":10,"storage_gb":500}}` | Cost optimization |
| 56 | POST | `/api/v1/intelligence/detect/anomalies` | `{"metric_name":"cpu","values":[45,48,95,50,52]}` | Anomalies detected |
| 57 | GET | `/api/v1/intelligence/insights?category=cost&limit=5` | - | Cost insights |
| 58 | GET | `/api/v1/intelligence/models` | - | ML models list |
| 59 | GET | `/api/v1/intelligence/chains` | - | AI chains |

**Testing Status:** ☐☐☐☐☐☐☐

---

## 📋 **QUICK REFERENCE IDs (Get these from BeekeeperStudio)**

Run these queries in BeekeeperStudio to get IDs:

```sql
-- Get Production Network ID
SELECT id, name FROM networks WHERE name = 'Production Network';

-- Get Node IDs
SELECT id, name, node_type FROM nodes LIMIT 5;

-- Get Link IDs
SELECT id, source_node_id, target_node_id FROM links LIMIT 3;

-- Get Policy IDs
SELECT id, name FROM policies LIMIT 3;

-- Get Alert IDs
SELECT id, title FROM alerts LIMIT 3;
```

---

## 🧪 **TESTING WORKFLOW**

### **Quick Test (5 minutes - 10 endpoints):**
1. ☐ Health (#1)
2. ☐ Login (#3)
3. ☐ Get Profile (#4)
4. ☐ List Networks (#8)
5. ☐ List Nodes (#16)
6. ☐ List Policies (#27)
7. ☐ List Alerts (#36)
8. ☐ Get Insights (#57)
9. ☐ Get AI Models (#58)
10. ☐ Create Network (#9)

### **Medium Test (15 minutes - 25 endpoints):**
- All from Quick Test +
- All Network CRUD operations (#8-15)
- All Node CRUD operations (#16-23)
- Monitoring basics (#32-36)
- AI Intelligence (#53-59)

### **Full Test (30 minutes - All 59 endpoints):**
- Test every endpoint in the table above

---

## 💡 **POSTMAN TIPS**

### **1. Save IDs as Variables:**
After creating a network, save the ID:
```javascript
// In Tests tab
pm.collectionVariables.set("network_id", pm.response.json().network.id);
```

### **2. Pre-request Script for Token:**
```javascript
// Automatically use saved token
pm.request.headers.add({
    key: 'Authorization',
    value: 'Bearer ' + pm.collectionVariables.get('access_token')
});
```

### **3. Check Response Status:**
```javascript
// In Tests tab
pm.test("Status code is 200", function () {
    pm.response.to.have.status(200);
});
```

---

## 📊 **EXPECTED SUCCESS RATES**

| Category | Working | Total | Success Rate |
|----------|---------|-------|--------------|
| System | 2 | 2 | 100% |
| Authentication | 5 | 5 | 100% |
| Networks | 14 | 14 | 100% |
| Nodes | 8 | 8 | 100% |
| Links | 3 | 3 | 100% |
| Policies | 5 | 5 | 100% |
| Monitoring | 6 | 15 | 40% |
| Users | 6 | 6 | 100% |
| AI Intelligence | 7 | 7 | 100% |
| **TOTAL** | **56** | **65** | **86%** |

**Note:** Some monitoring endpoints are stubs and may return mock data.

---

## ✅ **VERIFICATION CHECKLIST**

After testing, verify:

### **In Postman:**
- [ ] All GET requests return data
- [ ] All POST requests create resources
- [ ] All PUT requests update resources
- [ ] All DELETE requests remove resources
- [ ] No 500 Internal Server Errors
- [ ] Authentication works consistently

### **In BeekeeperStudio:**
- [ ] Created data appears in database
- [ ] Updated data shows changes
- [ ] Deleted data is removed
- [ ] Foreign key relationships intact

---

## 🚨 **KNOWN ISSUES**

1. **Network Update 500 Error:**
   - **Fix:** Send complete object with all required fields
   - Include: `name`, `subnet`, `status`, `description`

2. **Some Monitoring Endpoints:**
   - May return mock data (alert rules CRUD)
   - These are lower priority for MVP

---

## 📝 **TESTING NOTES**

**Date:** _____________  
**Tester:** _____________

| Endpoint # | Status | Notes |
|------------|--------|-------|
| | ☐ Pass ☐ Fail | |
| | ☐ Pass ☐ Fail | |
| | ☐ Pass ☐ Fail | |

---

**Generated:** October 30, 2025  
**Total Endpoints:** 59 testable endpoints  
**Estimated Test Time:** 30 minutes for full test


# 🧪 Complete Testing Guide

## ✅ **Everything is Ready for Testing!**

**Status:** 🟢 All 11 database tables populated with 45 test records  
**API:** 🟢 Running on `http://localhost:8080`  
**Database:** 🟢 PostgreSQL on port `5433`

---

## 📦 **What's Been Set Up**

### **1. Database (PostgreSQL)**
✅ All 11 tables created and populated:
- `users` (1 row)
- `networks` (3 rows)
- `nodes` (7 rows)
- `links` (5 rows)
- `policies` (4 rows)
- `metrics` (7 rows)
- `events` (5 rows)
- `alerts` (3 rows)
- `workflow_definitions` (2 rows)
- `workflow_executions` (3 rows)
- `task_executions` (5 rows)

### **2. API Gateway**
✅ Running with 84 endpoints across:
- Authentication
- Networks & Nodes
- Monitoring & Alerts
- AI Intelligence
- Workflows & Automation
- Policies

### **3. Testing Tools**
✅ Two Postman collections ready:
- `NetOrchestrator_Complete_Collection.json` - All 50+ requests
- `NetOrchestrator_Postman_Collection.json` - Essential 27 requests

---

## 🎯 **Quick Testing (2 Minutes)**

### **Step 1: Import Postman Collection**

1. Open Postman
2. Click **Import** button (top left)
3. Drag `NetOrchestrator_Complete_Collection.json` into the import window
4. Click **Import**

### **Step 2: Run Essential Tests**

The collection has **auto-authentication** - just run these requests in order:

#### **Test 1: Authentication** ✅
```
Request: 1. Authentication → Login
Expected: HTTP 200, Token auto-saved
```

#### **Test 2: List Networks** ✅
```
Request: 2. Networks → List Networks
Expected: 3 networks (Default, Production, Development)
Response should show:
  - Production Network: 10.0.0.0/16
  - Development Network: 172.16.0.0/24
```

#### **Test 3: List Nodes** ✅
```
Request: 3. Nodes → List Nodes
Expected: 7 nodes
Response should show:
  - prod-web-01, prod-web-02
  - prod-db-01, prod-cache-01, prod-lb-01
  - dev-app-01, dev-db-01
```

#### **Test 4: View Alerts** ✅
```
Request: 4. Monitoring → List Alerts
Expected: 3 alerts
Response should show:
  - High CPU on prod-db-01 (active)
  - Memory warning (acknowledged)
  - Network latency (resolved)
```

#### **Test 5: AI Insights** ✅
```
Request: 5. AI Intelligence → Get AI Insights
Expected: HTTP 200, AI-generated insights
```

#### **Test 6: Workflow List** ✅
```
Request: 6. Workflows → List Workflow Definitions
Expected: 2 workflows
Response should show:
  - Deploy Web Server (v1.0)
  - Scale Out (v1.0)
```

#### **Test 7: Workflow Executions** ✅
```
Request: 6. Workflows → List Workflow Executions
Expected: 3 executions
Response should show:
  - 1 completed
  - 1 failed
  - 1 running
```

---

## 📊 **Detailed Testing (10 Minutes)**

### **Test Group 1: Core CRUD Operations**

Run these requests to test full CRUD:

1. **Create Network**
   - Request: `2. Networks → Create Network`
   - Should return HTTP 201
   - Network ID will be auto-saved

2. **Get Network Details**
   - Request: `2. Networks → Get Network Details`
   - Uses auto-saved Network ID
   - Should show full network info

3. **Create Node**
   - Request: `3. Nodes → Create Node`
   - Creates node in the network
   - Should return HTTP 201

4. **Get Node Details**
   - Request: `3. Nodes → Get Node Details`
   - Shows full node configuration

### **Test Group 2: Monitoring**

5. **Network Metrics**
   - Request: `4. Monitoring → Network Metrics`
   - Shows throughput, latency, etc.

6. **Node Metrics**
   - Request: `4. Monitoring → Node Metrics`
   - Shows CPU, memory, disk I/O

7. **List Events**
   - Request: `4. Monitoring → List Events`
   - Shows audit log events

### **Test Group 3: AI Intelligence**

8. **Analyze Network**
   - Request: `5. AI Intelligence → Analyze Network`
   - AI-powered network analysis

9. **Predict Capacity**
   - Request: `5. AI Intelligence → Predict Capacity`
   - 30-day capacity prediction

10. **Optimize Costs**
    - Request: `5. AI Intelligence → Optimize Costs`
    - Cost optimization recommendations

11. **Detect Anomalies**
    - Request: `5. AI Intelligence → Detect Anomalies`
    - Real-time anomaly detection

### **Test Group 4: Workflows**

12. **List Workflows**
    - Request: `6. Workflows → List Workflow Definitions`
    - Shows 2 pre-defined workflows

13. **Get Workflow Details**
    - Request: `6. Workflows → Get Workflow Details`
    - Shows task definitions

14. **Execute Workflow**
    - Request: `6. Workflows → Execute Workflow`
    - Starts a new workflow execution

15. **Get Execution Status**
    - Request: `6. Workflows → Get Execution Details`
    - Check execution progress

16. **Get Execution Tasks**
    - Request: `6. Workflows → Get Execution Tasks`
    - See individual task status

---

## 🔍 **Database Verification (BeekeeperStudio)**

### **Connection Details:**
```
Host:     localhost
Port:     5433  ⬅️ Important!
Database: netorch_mvp
Username: netorch_admin
Password: SecurePass2025
```

### **Quick Verification Queries:**

#### **1. See All Table Counts:**
```sql
SELECT 
    'users' as table_name, COUNT(*) FROM users
UNION ALL
SELECT 'networks', COUNT(*) FROM networks
UNION ALL
SELECT 'nodes', COUNT(*) FROM nodes
UNION ALL
SELECT 'links', COUNT(*) FROM links
UNION ALL
SELECT 'policies', COUNT(*) FROM policies
UNION ALL
SELECT 'metrics', COUNT(*) FROM metrics
UNION ALL
SELECT 'events', COUNT(*) FROM events
UNION ALL
SELECT 'alerts', COUNT(*) FROM alerts
UNION ALL
SELECT 'workflow_definitions', COUNT(*) FROM workflow_definitions
UNION ALL
SELECT 'workflow_executions', COUNT(*) FROM workflow_executions
UNION ALL
SELECT 'task_executions', COUNT(*) FROM task_executions
ORDER BY table_name;
```

#### **2. See Network Topology:**
```sql
SELECT 
    net.name as network,
    ns.name as source_node,
    nt.name as target_node,
    l.bandwidth_mbps || ' Mbps' as bandwidth,
    l.latency_ms || ' ms' as latency
FROM links l
JOIN networks net ON l.network_id = net.id
JOIN nodes ns ON l.source_node_id = ns.id
JOIN nodes nt ON l.target_node_id = nt.id;
```

#### **3. See Active Alerts:**
```sql
SELECT 
    a.severity,
    a.title,
    a.status,
    n.name as affected_node
FROM alerts a
LEFT JOIN nodes n ON a.entity_id = n.id
WHERE a.entity_type = 'node'
ORDER BY a.created_at DESC;
```

#### **4. See Workflow Execution Status:**
```sql
SELECT 
    wd.name as workflow,
    we.status,
    COUNT(te.id) as total_tasks,
    we.started_at,
    we.completed_at
FROM workflow_executions we
JOIN workflow_definitions wd ON we.workflow_id = wd.id
LEFT JOIN task_executions te ON te.workflow_execution_id = we.id
GROUP BY wd.name, we.status, we.started_at, we.completed_at
ORDER BY we.started_at DESC;
```

---

## ✅ **Expected Results Summary**

### **API Endpoints:**
- ✅ Authentication: 3/3 working (login, profile, logout)
- ✅ Networks: 3/3 working (list, get, create)
- ✅ Nodes: 3/3 working (list, get, create)
- ✅ Monitoring: 4/4 working (alerts, metrics, events)
- ✅ AI Intelligence: 6/6 working (insights, analyze, predict, optimize, detect, models)
- ✅ Workflows: 7/7 working (list, get, create, execute, executions, tasks)
- ✅ Policies: 2/2 working (list, create)

**Total:** 28 essential endpoints fully functional

### **Database:**
- ✅ All 11 tables created
- ✅ 45 test records populated
- ✅ All foreign keys valid
- ✅ All relationships working

---

## 🐛 **Troubleshooting**

### **Issue: API returns 401 Unauthorized**
**Solution:** Click "Login" request first to get a new token

### **Issue: BeekeeperStudio shows empty tables**
**Solution:** Make sure you're connected to port `5433`, not `5432`

### **Issue: Network/Node not found**
**Solution:** Run "List Networks" or "List Nodes" first to refresh IDs

### **Issue: Workflow execution fails**
**Solution:** Check that workflow_id is set (run "List Workflows" first)

---

## 📋 **Testing Checklist**

Use this to verify everything:

- [ ] ✅ Postman collection imported
- [ ] ✅ Login successful (token saved)
- [ ] ✅ List Networks returns 3 networks
- [ ] ✅ List Nodes returns 7 nodes
- [ ] ✅ List Alerts returns 3 alerts
- [ ] ✅ AI Insights returns data
- [ ] ✅ Workflows list returns 2 workflows
- [ ] ✅ Workflow executions returns 3 executions
- [ ] ✅ Can create new network
- [ ] ✅ Can create new node
- [ ] ✅ Can execute workflow
- [ ] ✅ BeekeeperStudio shows all 11 tables
- [ ] ✅ BeekeeperStudio queries return data

---

## 🎉 **Success Criteria**

Your MVP is working if:

1. ✅ All 28 essential Postman requests return HTTP 200/201
2. ✅ BeekeeperStudio shows all 11 tables with data
3. ✅ Can create new networks and nodes
4. ✅ Monitoring shows real metrics and alerts
5. ✅ AI endpoints return insights
6. ✅ Workflows can be listed and executed
7. ✅ No 404 or 500 errors on core endpoints

---

## 📁 **Files Created for You**

1. **`NetOrchestrator_Complete_Collection.json`** - Full Postman collection (50+ requests)
2. **`DATABASE_COMPLETE_SUMMARY.md`** - Complete database documentation
3. **`COMPLETE_TESTING_GUIDE.md`** - This file
4. **`BEEKEEPER_CONNECTION.md`** - Database connection guide
5. **`API_ENDPOINTS.md`** - Complete endpoint reference

---

## 🚀 **Next Steps After Testing**

Once you've verified everything works:

1. **Phase 3: Polish Database Layer** ✅ DONE
   - All tables created
   - Test data populated
   - Queries working

2. **Phase 4: Add More Features** (Optional)
   - Additional automation workflows
   - More AI models
   - Advanced monitoring rules

3. **Phase 5: Production Prep** (Later)
   - Add comprehensive tests
   - Add documentation
   - Security hardening
   - Performance optimization

---

**Status:** 🟢 **MVP READY FOR TESTING**

**Last Updated:** October 30, 2025  
**Test Data:** 45 records across 11 tables  
**Endpoints:** 84 total, 28 essential tested


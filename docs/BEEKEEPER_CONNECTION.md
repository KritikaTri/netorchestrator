# 🐝 BeekeeperStudio Database Connection Guide

## 📋 **Connection Settings**

Use these exact credentials in BeekeeperStudio:

```
Connection Type: PostgreSQL
Host:            localhost
Port:            5432
Database:        netorchestrator
Username:        netorchestrator
Password:        password
SSL:             Disabled
```

---

## 🚀 **Setup Steps:**

1. **Open BeekeeperStudio**

2. **Click "New Connection"**

3. **Fill in the form:**
   - **Connection Type:** Select "PostgreSQL" from dropdown
   - **Host:** `localhost`
   - **Port:** `5432`
   - **Database:** `netorchestrator`
   - **Username:** `netorchestrator`
   - **Password:** `password`
   - **Connection Name:** `NetOrchestrator Local` (optional, for reference)

4. **Test Connection**
   - Click "Test" button to verify
   - Should show: "✅ Connection successful"

5. **Save & Connect**
   - Click "Connect" or "Save & Connect"

---

## 📊 **Available Tables**

Once connected, you'll see these tables:

| Table | Description | Row Count |
|-------|-------------|-----------|
| `networks` | Network definitions | 3 |
| `nodes` | Network nodes | 0 |
| `links` | Network links/connections | 0 |
| `users` | User accounts | 1 |
| `policies` | Security/routing policies | 0 |
| `alerts` | System alerts | 0 |
| `metrics` | Performance metrics | 0 |
| `events` | Event log | 0 |
| `workflow_definitions` | Automation workflows | 0 |
| `workflow_executions` | Workflow run history | 0 |
| `task_executions` | Task execution logs | 0 |

---

## 🔍 **Useful Queries**

### **View All Networks:**
```sql
SELECT id, name, subnet, status, created_at 
FROM networks 
ORDER BY created_at DESC;
```

### **View All Nodes:**
```sql
SELECT id, name, node_type, ip_address, status, network_id
FROM nodes
ORDER BY created_at DESC;
```

### **View All Links:**
```sql
SELECT l.id, l.link_type, l.bandwidth_mbps, l.status,
       sn.name as source_node, tn.name as target_node
FROM links l
LEFT JOIN nodes sn ON l.source_node_id = sn.id
LEFT JOIN nodes tn ON l.target_node_id = tn.id
ORDER BY l.created_at DESC;
```

### **View Nodes by Network:**
```sql
SELECT n.name as network, 
       COUNT(nd.id) as node_count,
       ARRAY_AGG(nd.name) as node_names
FROM networks n
LEFT JOIN nodes nd ON nd.network_id = n.id
GROUP BY n.name;
```

### **View Network Topology:**
```sql
SELECT n.name as network,
       (SELECT COUNT(*) FROM nodes WHERE network_id = n.id) as nodes,
       (SELECT COUNT(*) FROM links WHERE network_id = n.id) as links
FROM networks n
ORDER BY n.created_at DESC;
```

---

## 🗄️ **Schema Reference**

### **Networks Table:**
```
id           UUID (PK)
name         VARCHAR(255)
subnet       VARCHAR(50)
status       VARCHAR(50)
network_type VARCHAR(100)
gateway      VARCHAR(50)
vlan_id      INTEGER
description  TEXT
metadata     JSONB
config       JSONB
created_at   TIMESTAMP
updated_at   TIMESTAMP
```

### **Nodes Table:**
```
id          UUID (PK)
network_id  UUID (FK)
name        VARCHAR(255)
node_type   VARCHAR(100)
ip_address  VARCHAR(50)
mac_address VARCHAR(17)
status      VARCHAR(50)
cpu_cores   INTEGER
memory_mb   INTEGER
disk_gb     INTEGER
image       VARCHAR(255)
ports       JSONB
environment JSONB
metadata    JSONB
config      JSONB
position    JSONB
created_at  TIMESTAMP
updated_at  TIMESTAMP
```

### **Links Table:**
```
id              UUID (PK)
network_id      UUID (FK)
source_node_id  UUID (FK)
target_node_id  UUID (FK)
link_type       VARCHAR(100)
bandwidth_mbps  INTEGER
latency_ms      INTEGER
packet_loss     NUMERIC
status          VARCHAR(50)
metadata        JSONB
config          JSONB
created_at      TIMESTAMP
updated_at      TIMESTAMP
```

---

## ✅ **Connection Verified**

Your PostgreSQL database is running and accessible!

- **Container:** postgres (Up 17 hours)
- **Port:** 5432 (mapped to localhost)
- **Status:** ✅ Healthy


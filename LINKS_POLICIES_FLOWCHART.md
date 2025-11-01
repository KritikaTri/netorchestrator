# 🔗🛡️ NetOrchestrator: Links and Policies Flow Chart

## **Complete Flow: Virtual Design → Real Container Enforcement**

---

# 🔗 **LINK CREATION FLOW**

```
┌─────────────────────────────────────────────────────────────┐
│                    LINK CREATION FLOW                       │
└─────────────────────────────────────────────────────────────┘

📱 API REQUEST
POST /api/v1/networks/{network_id}/links
{
  "name": "Router-to-Server Link",
  "source_node_id": "776ec99a-18ba-45ef-bdcc-d10c94e96352",
  "target_node_id": "8856b563-ca0c-4ef4-aa49-49f27d48f3f6",
  "config": {
    "bandwidth": 1000,
    "latency": 5,
    "packet_loss": 0.01,
    "protocol": "ethernet"
  }
}
           ↓
┌─────────────────────────────────────────────────────────────┐
│               DATABASE OPERATIONS                           │
└─────────────────────────────────────────────────────────────┘

📋 Step 1: Validate Link Request
   • Check source_node_id exists
   • Check target_node_id exists  
   • Validate both nodes are in same network
   • Check bandwidth/latency values are valid
           ↓
📋 Step 2: Create Link Record
   INSERT INTO "links" (
     id, network_id, source_node_id, target_node_id,
     name, status, config, created_at, updated_at
   ) VALUES (
     'dc3eb451-c7b4-4892-a6b1-3514579ff62b',
     'bd6e4f19-16ba-42ab-9238-548478ac91e1',
     '776ec99a-18ba-45ef-bdcc-d10c94e96352',
     '8856b563-ca0c-4ef4-aa49-49f27d48f3f6',
     'Router-to-Server Link',
     'pending',
     '{"bandwidth":1000,"latency":5,"packet_loss":0.01}',
     '2025-11-01 17:33:34.921',
     '2025-11-01 17:33:34.921'
   )
           ↓
┌─────────────────────────────────────────────────────────────┐
│              CONTAINER PROVISIONING                        │
└─────────────────────────────────────────────────────────────┘

🐳 Step 3: Container Discovery
   • LinkProvisioner.ProvisionLink() called
   • Query database for source node container info
     SELECT * FROM nodes WHERE id = '776ec99a...'
     → Returns: container_name = "node_776ec99a"
   • Query database for target node container info  
     SELECT * FROM nodes WHERE id = '8856b563...'
     → Returns: container_name = "node_8856b563"
           ↓
🐳 Step 4: Container Network Verification
   • Check if source container is running
     podman ps --filter name=node_776ec99a
     → Result: RUNNING ✓
   • Check if target container is running
     podman ps --filter name=node_8856b563
     → Result: RUNNING ✓
   • Verify both containers are in same network
     podman inspect node_776ec99a | grep NetworkID
     podman inspect node_8856b563 | grep NetworkID
     → Result: Both in net_bd6e4f19 ✓
           ↓
🐳 Step 5: Network Connectivity Establishment
   • Containers already in same network → Skip network bridging
   • Apply bandwidth/latency constraints
     podman exec node_776ec99a tc qdisc add dev eth0 root tbf rate 1000mbit
     podman exec node_8856b563 tc qdisc add dev eth0 root tbf rate 1000mbit
   • Apply packet loss simulation (if requested)
     podman exec node_xxx tc qdisc add dev eth0 root netem loss 0.01%
           ↓
🐳 Step 6: Update Link with Container Information
   UPDATE "links" SET
     status = 'active',
     config = '{
       "bandwidth": 1000,
       "latency": 5,
       "custom_attrs": {
         "source_container": "node_776ec99a",
         "target_container": "node_8856b563", 
         "connection_type": "container_bridge",
         "provisioned_at": "2025-11-01T17:33:35+05:30"
       }
     }',
     updated_at = '2025-11-01 17:33:35.744'
   WHERE id = 'dc3eb451-c7b4-4892-a6b1-3514579ff62b'
           ↓
📡 API RESPONSE
{
  "link": {
    "id": "dc3eb451-c7b4-4892-a6b1-3514579ff62b",
    "name": "Router-to-Server Link", 
    "status": "active",
    "config": {
      "custom_attrs": {
        "source_container": "node_776ec99a",
        "target_container": "node_8856b563",
        "connection_type": "container_bridge"
      }
    }
  },
  "message": "Link created and real container networking provisioned"
}
```

---

# 🛡️ **POLICY CREATION FLOW**

```
┌─────────────────────────────────────────────────────────────┐
│                   POLICY CREATION FLOW                      │
└─────────────────────────────────────────────────────────────┘

📱 API REQUEST  
POST /api/v1/demo/firewall
{
  "container_name": "policy-host",
  "rules": [
    {
      "name": "Allow HTTP",
      "protocol": "tcp",
      "port": "80",
      "action": "accept"
    },
    {
      "name": "Block Telnet", 
      "protocol": "tcp",
      "port": "23",
      "action": "drop"
    }
  ]
}
           ↓
┌─────────────────────────────────────────────────────────────┐
│               API VALIDATION & PROCESSING                   │
└─────────────────────────────────────────────────────────────┘

🔐 Step 1: Authentication Check
   • Verify JWT token in Authorization header
   • Extract user permissions and roles
   • Log authentication attempt to audit trail
           ↓
📋 Step 2: Request Validation
   • Validate container_name is provided
   • Check rules array is not empty  
   • Validate rule structure (name, protocol, port, action)
   • Verify container exists and is running
     podman ps --filter name=policy-host
     → Result: Container FOUND and RUNNING ✓
           ↓
┌─────────────────────────────────────────────────────────────┐
│              REAL CONTAINER ENFORCEMENT                     │
└─────────────────────────────────────────────────────────────┘

🛠️ Step 3: Policy Enforcement Preparation
   • PolicyEnforcer.ApplyFirewallPolicy() called
   • Check if container has required capabilities
     podman inspect policy-host | grep CapAdd
     → Result: NET_ADMIN capability available ✓
   • Ensure iptables tools are installed
     podman exec policy-host which iptables
     → Result: iptables AVAILABLE ✓
           ↓
🛠️ Step 4: Clear Previous Rules (Clean Slate)
   • Clear existing iptables INPUT chain
     podman exec policy-host iptables -F INPUT
   • Set default policy to ACCEPT  
     podman exec policy-host iptables -P INPUT ACCEPT
   • Log clearing action
     → Result: Clean firewall state ready for new rules
           ↓
🛠️ Step 5: Apply Real iptables Rules (The Magic!)
   For each rule in policy:
   
   Rule 1: "Allow HTTP"
   • Build iptables command
     cmd = "iptables -A INPUT -p tcp --dport 80 -j ACCEPT -m comment --comment 'NetOrchestrator: Allow HTTP'"
   • Execute on container
     podman exec policy-host iptables -A INPUT -p tcp --dport 80 -j ACCEPT -m comment --comment "NetOrchestrator: Allow HTTP"
   • Result: REAL FIREWALL RULE APPLIED ✅
   
   Rule 2: "Block Telnet"
   • Build iptables command
     cmd = "iptables -A INPUT -p tcp --dport 23 -j DROP -m comment --comment 'NetOrchestrator: Block Telnet'"
   • Execute on container
     podman exec policy-host iptables -A INPUT -p tcp --dport 23 -j DROP -m comment --comment "NetOrchestrator: Block Telnet"
   • Result: REAL FIREWALL RULE APPLIED ✅
           ↓
🛠️ Step 6: Verification and Status Update
   • Get current iptables status
     podman exec policy-host iptables -L INPUT -n --line-numbers
   • Count applied rules
     → Result: 6 total rules, 6 NetOrchestrator rules
   • Update enforcement metadata
     applied_rules = [
       {"rule_index": 1, "command": "iptables -A INPUT...", "status": "applied"},
       {"rule_index": 2, "command": "iptables -A INPUT...", "status": "applied"}
     ]
           ↓
📡 API RESPONSE
{
  "status": "success",
  "container": "policy-host", 
  "applied_rules": [
    {"rule_index": 1, "status": "applied"},
    {"rule_index": 2, "status": "applied"}
  ],
  "iptables_output": "Chain INPUT (policy ACCEPT)\n1 ACCEPT tcp dpt:80 /* NetOrchestrator: Allow HTTP */\n2 DROP tcp dpt:23 /* NetOrchestrator: Block Telnet */",
  "message": "Real firewall policy enforced on container",
  "enforcement_method": "iptables"
}
           ↓
🧪 REAL ENFORCEMENT TESTING
   • Test blocked port (should fail)
     podman exec policy-router nc -z 10.89.2.7 23
     → Result: CONNECTION BLOCKED ✅ (firewall working!)
   • Test allowed port (should succeed)  
     podman exec policy-router nc -z 10.89.2.7 80
     → Result: CONNECTION ALLOWED ✅ (firewall working!)
```

---

# ⚡ **QOS POLICY CREATION FLOW**

```
┌─────────────────────────────────────────────────────────────┐
│                 QOS POLICY CREATION FLOW                    │
└─────────────────────────────────────────────────────────────┘

📱 API REQUEST
POST /api/v1/demo/qos
{
  "container_name": "policy-host",
  "classes": [
    {"classid": "1:10", "rate": "1000mbit"},  # Critical
    {"classid": "1:20", "rate": "500mbit"},   # Normal  
    {"classid": "1:30", "rate": "100mbit"}    # Background
  ]
}
           ↓
🔐 Authentication & Validation (same as firewall)
           ↓
┌─────────────────────────────────────────────────────────────┐
│            TRAFFIC CONTROL ENFORCEMENT                      │
└─────────────────────────────────────────────────────────────┘

⚡ Step 1: Clear Previous QoS Rules
   • Remove existing traffic control
     podman exec policy-host tc qdisc del dev eth0 root
   • Result: Clean traffic control state
           ↓
⚡ Step 2: Create HTB Root Queueing Discipline  
   • Set up hierarchical token bucket
     podman exec policy-host tc qdisc add dev eth0 root handle 1: htb default 30
   • Result: HTB qdisc created for traffic shaping
           ↓
⚡ Step 3: Apply Real Traffic Classes (The Magic!)
   For each class in policy:
   
   Class 1:10 (Critical - 1000 Mbps)
   • Execute tc command
     podman exec policy-host tc class add dev eth0 parent 1: classid 1:10 htb rate 1000mbit
   • Result: REAL TRAFFIC CLASS CREATED ✅
   
   Class 1:20 (Normal - 500 Mbps)  
   • Execute tc command
     podman exec policy-host tc class add dev eth0 parent 1: classid 1:20 htb rate 500mbit
   • Result: REAL TRAFFIC CLASS CREATED ✅
   
   Class 1:30 (Background - 100 Mbps)
   • Execute tc command  
     podman exec policy-host tc class add dev eth0 parent 1: classid 1:30 htb rate 100mbit
   • Result: REAL TRAFFIC CLASS CREATED ✅
           ↓
⚡ Step 4: Verification and Status  
   • Get current tc status
     podman exec policy-host tc qdisc show dev eth0
     → Result: "qdisc htb 1: root refcnt 8 r2q 10 default 0x30"
   • Get traffic classes
     podman exec policy-host tc class show dev eth0
     → Result: 3 HTB classes with correct bandwidth limits
           ↓
📡 API RESPONSE
{
  "status": "success",
  "container": "policy-host",
  "applied_classes": [
    {"class_id": "1:10", "rate": "1000mbit", "status": "applied"},
    {"class_id": "1:20", "rate": "500mbit", "status": "applied"}, 
    {"class_id": "1:30", "rate": "100mbit", "status": "applied"}
  ],
  "qdisc_output": "qdisc htb 1: root refcnt 8...",
  "class_output": "class htb 1:10 root prio 0 rate 1Gbit...",
  "enforcement_method": "tc (traffic control)"
}
```

---

# 🌐 **VIRTUAL NETWORK POLICY INTEGRATION FLOW**

```
┌─────────────────────────────────────────────────────────────┐
│           VIRTUAL NETWORK POLICY INTEGRATION                │
└─────────────────────────────────────────────────────────────┘

📱 VIRTUAL POLICY API REQUEST
POST /api/v1/networks/{network_id}/policies
{
  "name": "Enterprise Security Policy",
  "type": "firewall", 
  "config": {
    "rules": [
      {
        "id": "allow_web",
        "name": "Allow Web Traffic",
        "condition": {"protocol": "tcp", "port": "80"},
        "action": {"action": "accept"}
      }
    ]
  }
}
           ↓
📋 Step 1: Database Policy Creation
   INSERT INTO "policies" (
     id, network_id, name, policy_type, status, config
   ) VALUES (
     '4efef763-904b-4b87-895c-ec5bb31b3ad6',
     'bd6e4f19-16ba-42ab-9238-548478ac91e1', 
     'Enterprise Security Policy',
     'firewall',
     'pending',
     '{"rules":[...]}' 
   )
           ↓
🐳 Step 2: Container Discovery for Enforcement
   • PolicyEnforcer.EnforcePolicy() called
   • Find all containers in network
     SELECT * FROM nodes WHERE network_id = 'bd6e4f19...'
     → Returns: 2 nodes with container info
   
   Node 1: Core Router Demo
   • container_name: "node_776ec99a"
   • container_image: "frrouting/frr:latest"
   
   Node 2: Application Server  
   • container_name: "node_8856b563"
   • container_image: "ubuntu:20.04"
           ↓
🛡️ Step 3: Apply Policy to Each Container
   For each container found:
   
   Container: node_776ec99a (FRRouting)
   • Check if container has NET_ADMIN capability
   • Install iptables if needed (usually present in FRR)
   • Apply firewall rules via iptables commands
     podman exec node_776ec99a iptables -A INPUT -p tcp --dport 80 -j ACCEPT...
   • Result: Policy applied (or logged warning if failed)
   
   Container: node_8856b563 (Ubuntu)  
   • Check if container has NET_ADMIN capability ✓
   • Install iptables if needed (apt-get install iptables)
   • Apply firewall rules via iptables commands
     podman exec node_8856b563 iptables -A INPUT -p tcp --dport 80 -j ACCEPT...
   • Result: POLICY SUCCESSFULLY APPLIED ✅
           ↓
📋 Step 4: Update Database with Enforcement Status
   UPDATE "policies" SET
     status = 'active',
     config = '{
       "rules": [...],
       "custom_attrs": {
         "enforced_at": "2025-11-01T17:37:39+05:30",
         "enforcement_method": "iptables"
       }
     }',
     updated_at = '2025-11-01 17:37:39.791'
   WHERE id = '4efef763-904b-4b87-895c-ec5bb31b3ad6'
           ↓
📡 API RESPONSE
{
  "policy": {
    "id": "4efef763-904b-4b87-895c-ec5bb31b3ad6",
    "name": "Enterprise Security Policy",
    "type": "firewall",
    "status": "active"
  },
  "message": "Policy created and enforced on real container infrastructure",
  "enforcement": "Network policy rules applied to running containers"
}
```

---

# 🔄 **COMPLETE INTEGRATION ARCHITECTURE**

```
┌─────────────────────────────────────────────────────────────┐
│                NETORCHESTRATOR ARCHITECTURE                 │
└─────────────────────────────────────────────────────────────┘

🌐 VIRTUAL LAYER (PostgreSQL Database)
┌─────────────────────────────────────────────────────────────┐
│ networks: "Live Demo Network"                               │
│ ├─ nodes: "Core Router Demo" (router, active)              │  
│ ├─ nodes: "Application Server" (host, active)              │
│ ├─ links: "Router-to-Server Link" (active, 1000 Mbps)      │
│ └─ policies: "Enterprise Security" (active, iptables)      │
└─────────────────────────────────────────────────────────────┘
                            ↕ 
               DATABASE ↔ CONTAINER BRIDGE
                            ↕
🐳 REAL LAYER (Podman Container Runtime)
┌─────────────────────────────────────────────────────────────┐
│ net_bd6e4f19: Container Network (10.89.2.0/24)             │
│ ├─ policy-router: FRRouting (10.89.2.6) + policies        │
│ └─ policy-host: Ubuntu (10.89.2.7) + iptables + tc        │
└─────────────────────────────────────────────────────────────┘

🔗 INTEGRATION COMPONENTS:

ContainerProvisioner:
• Virtual networks → Real container networks
• Virtual nodes → Smart container selection
• API names → Container hostnames

LinkProvisioner:  
• Virtual links → Real container networking
• Bandwidth/latency → Traffic control rules
• Network topology → Container connectivity

PolicyEnforcer:
• Virtual policies → Real iptables rules
• QoS policies → Real tc (traffic control)
• Security rules → Container hardening

API Layer:
• REST endpoints → Database operations → Container commands
• Status queries → Live container inspection
• Enhanced topology → Virtual-to-real mapping
```

---

# 🧪 **VERIFICATION FLOW**

```
┌─────────────────────────────────────────────────────────────┐
│                   VERIFICATION FLOW                         │
└─────────────────────────────────────────────────────────────┘

📊 Policy Status Check
GET /api/v1/demo/containers/policy-host/policy-status
           ↓
🔍 Live Container Inspection
   • Check iptables status
     podman exec policy-host iptables -L INPUT -n --line-numbers
     → Parse output, count NetOrchestrator rules
   
   • Check traffic control status  
     podman exec policy-host tc qdisc show dev eth0
     podman exec policy-host tc class show dev eth0
     → Parse output, count traffic classes
           ↓
📈 Status Analysis
   firewall_status = {
     "status": "available",
     "rules_count": 14,
     "netorchestrator_rules": 6  ← REAL ENFORCEMENT PROOF
   }
   
   qos_status = {  
     "status": "available",
     "classes_count": 3  ← REAL TRAFFIC SHAPING PROOF
   }
           ↓
📡 VERIFICATION RESPONSE
{
  "policy_status": {
    "firewall": {
      "status": "available",
      "netorchestrator_rules": 6,  ← PROOF OF REAL ENFORCEMENT
      "output": "ACCEPT tcp dpt:80 /* NetOrchestrator: Allow HTTP */"
    },
    "qos": {
      "status": "available", 
      "classes_count": 3  ← PROOF OF REAL TRAFFIC CONTROL
    }
  },
  "message": "Real container policy enforcement status"
}
```

---

# 🎯 **FLOW SUMMARY: WHAT ACTUALLY HAPPENS**

## **🔗 Links Flow:**
```
API Call → Database Record → Container Discovery → Network Verification → Metadata Update → Response
```

## **🛡️ Policies Flow:**  
```
API Call → Validation → Container Discovery → Real iptables/tc Commands → Database Update → Verification → Response
```

## **🧪 Testing Flow:**
```
API Status Query → Live Container Inspection → Rule Counting → Network Testing → Proof of Enforcement
```

---

# 🏆 **KEY INTEGRATION POINTS**

### **✅ Database ↔ Container Bridge:**
- **Virtual records** store container metadata  
- **Container operations** update database status
- **API queries** combine virtual and real data

### **✅ Smart Orchestration:**
- **Node types** determine container images
- **Policy types** determine enforcement methods  
- **Container capabilities** determine feature availability

### **✅ Real Enforcement:**
- **iptables commands** applied to live containers
- **tc (traffic control)** for bandwidth management
- **Network testing** proves actual enforcement

---

**This flowchart shows the complete integration between virtual network design and real container infrastructure with actual policy enforcement! 🚀**

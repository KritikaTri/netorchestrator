# NetOrchestrator NLP API - Capabilities & Limitations Analysis

## 📊 **CURRENT NLP API CAPABILITIES**

### ✅ **What the NLP API CAN Do:**

#### **1. Node Creation & Management**
- **Supported Node Types**: `router`, `switch`, `host`, `firewall`, `load_balancer`, `gateway`
- **Smart Parsing**: Handles "2 routers, 1 host, 3 switches" format
- **Container Provisioning**: Creates real Docker/Podman containers
- **Mixed Topologies**: Can create heterogeneous node combinations

#### **2. Network Topology Recognition**  
- **Supported Topologies**: `star`, `mesh`, `tree`, `ring`, `custom`
- **Automatic Detection**: Recognizes topology keywords in natural language
- **Topology Constraints**: 
  - Star: minimum 2 nodes
  - Mesh: minimum 3 nodes  
  - Tree: minimum 3 nodes
  - Ring: minimum 3 nodes
  - Maximum: 50 nodes per network

#### **3. Policy Creation (FIXED TYPES)**
Currently supports **only firewall policies** with these rule types:

##### **SSH Policies**
```json
{
  "name": "Allow SSH" | "Block SSH",
  "action": "allow" | "drop", 
  "port": 22,
  "protocol": "tcp"
}
```

##### **HTTP/HTTPS Policies**  
```json
{
  "name": "Allow HTTP" | "Allow HTTPS",
  "action": "allow",
  "port": 80 | 443,
  "protocol": "tcp"
}
```

##### **General Traffic Policies**
```json
{
  "name": "Allow All Traffic" | "Allow All Incoming/Outgoing Traffic",
  "action": "allow",
  "port": "all",
  "protocol": "all" 
}
```

##### **Protocol Blocking**
```json
{
  "name": "Block FTP" | "Block Telnet",
  "action": "drop", 
  "port": 21 | 23,
  "protocol": "tcp"
}
```

#### **4. Infrastructure Provisioning**
- **Real Containers**: Creates Docker/Podman containers with appropriate images
- **Network Creation**: Establishes isolated container networks
- **Database Integration**: Stores all configurations in PostgreSQL

---

## ❌ **CURRENT LIMITATIONS**

### **1. Link Creation - NOT SUPPORTED BY NLP**
- **Manual Only**: Links must be created via REST API endpoints
- **No NLP Parsing**: Cannot parse "connect router1 to switch1" 
- **Requires**: Explicit POST to `/api/v1/networks/{id}/links`

### **2. Policy Limitations - VERY RESTRICTED**

#### **Policy Types NOT Supported by NLP:**
- ❌ **QoS Policies** (bandwidth, traffic shaping)
- ❌ **Routing Policies** (OSPF, BGP configurations)  
- ❌ **Access Control Policies** (user/role-based access)
- ❌ **Security Policies** (beyond basic firewall)
- ❌ **Traffic Policies** (load balancing, failover)

#### **Firewall Policy Limitations:**
- ❌ **Complex Rules**: Only simple port+protocol combinations
- ❌ **IP Ranges**: Cannot specify source/destination subnets  
- ❌ **Time-based Rules**: No scheduling or temporal policies
- ❌ **Rate Limiting**: No bandwidth or connection limits
- ❌ **Custom Chains**: No iptables chain management

### **3. Network Configuration Limitations**
- ❌ **IP Assignment**: No automatic IP address allocation
- ❌ **VLAN Configuration**: No VLAN segmentation support
- ❌ **Subnet Specification**: Limited subnet configuration
- ❌ **DNS Configuration**: No DNS server setup
- ❌ **Gateway Configuration**: Basic gateway settings only

### **4. Device-Specific Limitations**
- ❌ **Vendor Configs**: No Cisco/Juniper-specific configurations
- ❌ **Protocol Support**: Limited to basic TCP/UDP
- ❌ **Advanced Features**: No MPLS, VPN, or enterprise protocols

---

## 🔧 **ENHANCEMENT POSSIBILITIES**

### **Easy Additions (Low Effort)**
1. **Link Creation Support**: Add NLP parsing for "connect A to B"
2. **Enhanced Firewall Rules**: Add source/destination IP support
3. **Port Ranges**: Support "block ports 8000-9000"
4. **More Protocols**: Add ICMP, UDP-specific rules

### **Medium Effort Enhancements**
1. **QoS Policies**: Add "limit bandwidth to 100Mbps"
2. **IP Address Management**: Add "assign 192.168.1.0/24 subnet"
3. **Advanced Firewall**: Add rate limiting, connection tracking
4. **Routing Policies**: Add "enable OSPF on routers"

### **Complex Additions (High Effort)**
1. **Vendor-Specific Configs**: Cisco IOS, Juniper JUNOS commands
2. **Service Orchestration**: Application deployment on nodes
3. **Multi-Cloud Support**: AWS, Azure, GCP integration
4. **Enterprise Security**: Advanced threat protection, DLP

---

## 🎯 **CURRENT NLP API SPECIFICATION**

### **Request Format**
```bash
POST /api/v1/ai/provision
{
  "text": "Create a star network with 2 routers, 1 host and firewall policy where SSH is blocked"
}
```

### **Supported Natural Language Patterns**

#### **Node Specification**
- ✅ `"2 routers, 1 host, 3 switches"`
- ✅ `"Create mesh network with 4 hosts"`
- ✅ `"Star topology with 1 router and 5 nodes"`

#### **Topology Specification**
- ✅ `"star network"`, `"mesh topology"`, `"tree structure"`, `"ring network"`

#### **Policy Specification**  
- ✅ `"firewall policy where SSH is blocked"`
- ✅ `"SSH access is enabled"`
- ✅ `"all incoming and outgoing traffic is allowed"`
- ✅ `"block FTP and Telnet"`

### **Response Format**
```json
{
  "success": true,
  "network": {
    "id": "uuid",
    "name": "Generated Name",
    "topology": "star"
  },
  "nodes": [...],
  "nodes_created": 3,
  "spec": {
    "nodes": [...],
    "policies": [...],
    "topology": "star"
  }
}
```

---

## 🚀 **RECOMMENDATIONS FOR EXPANSION**

### **Priority 1: Link Support**
Add NLP parsing for link creation:
```javascript
// Example enhancement
if (text.includes("connect") || text.includes("link")) {
  // Parse "connect router1 to switch1"
  // Create LinkSpec objects
}
```

### **Priority 2: Enhanced Policy Framework**
Expand policy types beyond firewall:
```javascript
// QoS Policy Example
{
  "name": "Bandwidth Limit",
  "type": "qos", 
  "rules": [{"bandwidth": "100Mbps", "direction": "ingress"}]
}
```

### **Priority 3: Advanced Firewall Rules**
Add more sophisticated firewall capabilities:
```javascript
// Enhanced Firewall Rule
{
  "name": "Restrict Admin Access",
  "action": "allow",
  "port": 22,
  "protocol": "tcp",
  "source_ip": "192.168.1.0/24",
  "time_restriction": "business_hours"
}
```

---

## 🎯 **SUMMARY**

| Feature | NLP Support | Manual API | Container Provisioning |
|---------|-------------|------------|----------------------|
| **Node Creation** | ✅ Full | ✅ Full | ✅ Yes |
| **Topology Setup** | ✅ Full | ✅ Full | ✅ Yes |  
| **Link Creation** | ❌ None | ✅ Full | ✅ Yes |
| **Firewall Policies** | ✅ Basic | ✅ Full | ⚠️ Partial |
| **QoS Policies** | ❌ None | ✅ Full | ❌ No |
| **Routing Policies** | ❌ None | ✅ Full | ❌ No |
| **Access Policies** | ❌ None | ✅ Full | ❌ No |

**Current Status**: NLP API provides **simplified network creation** with **basic firewall policies** but requires **manual API calls** for advanced features like links and comprehensive policy management.

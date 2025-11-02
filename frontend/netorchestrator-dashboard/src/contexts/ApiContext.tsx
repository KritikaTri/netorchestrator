import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import { useAuth } from './AuthContext';

// Type definitions
interface Network {
  id: string;
  name: string;
  description: string;
  status: string;
  user_id: string;
  config: any;
  created_at: string;
  updated_at: string;
}

interface Node {
  id: string;
  network_id: string;
  name: string;
  type: string;
  ip_address: string;
  mac_address?: string;
  status: string;
  config: any;
  position: any;
  created_at: string;
  updated_at: string;
}

interface Link {
  id: string;
  network_id: string;
  source_node_id: string;
  target_node_id: string;
  name: string;
  status: string;
  config: any;
  created_at: string;
  updated_at: string;
}

interface Policy {
  id: string;
  network_id: string;
  name: string;
  type: string;
  status: string;
  config: any;
  created_at: string;
  updated_at: string;
}

interface Alert {
  id: string;
  title: string;
  description: string;
  severity: string;
  status: string;
  source_type: string;
  source_id: string;
  created_at: string;
  updated_at: string;
}

interface Metric {
  id: string;
  source_type: string;
  source_id: string;
  metric_name: string;
  value: number;
  unit: string;
  timestamp: string;
}

interface Health {
  status: string;
  uptime: number;
  cpu_usage: number;
  memory_usage: number;
  disk_usage: number;
  network_latency: number;
}

interface ApiContextType {
  // Data state
  networks: Network[];
  nodes: Node[];
  links: Link[];
  policies: Policy[];
  alerts: Alert[];
  metrics: Metric[];
  loading: boolean;
  error: string | null;
  
  // Network operations
  fetchNetworks: () => Promise<void>;
  createNetwork: (network: Partial<Network>) => Promise<Network>;
  updateNetwork: (id: string, network: Partial<Network>) => Promise<Network>;
  deleteNetwork: (id: string) => Promise<void>;
  getNetwork: (id: string) => Promise<Network>;
  startNetwork: (id: string) => Promise<void>;
  stopNetwork: (id: string) => Promise<void>;
  restartNetwork: (id: string) => Promise<void>;
  
  // Node operations
  fetchNodes: (networkId?: string) => Promise<void>;
  createNode: (networkId: string, node: Partial<Node>) => Promise<Node>;
  updateNode: (id: string, node: Partial<Node>) => Promise<Node>;
  deleteNode: (id: string) => Promise<void>;
  getNode: (id: string) => Promise<Node>;
  startNode: (id: string) => Promise<void>;
  stopNode: (id: string) => Promise<void>;
  restartNode: (id: string) => Promise<void>;
  
  // Link operations
  fetchLinks: (networkId: string) => Promise<void>;
  createLink: (networkId: string, link: Partial<Link>) => Promise<Link>;
  updateLink: (id: string, link: Partial<Link>) => Promise<Link>;
  deleteLink: (id: string) => Promise<void>;
  getLink: (id: string) => Promise<Link>;
  
  // Policy operations
  fetchPolicies: (networkId: string) => Promise<void>;
  createPolicy: (networkId: string, policy: Partial<Policy>) => Promise<Policy>;
  updatePolicy: (id: string, policy: Partial<Policy>) => Promise<Policy>;
  deletePolicy: (id: string) => Promise<void>;
  getPolicy: (id: string) => Promise<Policy>;
  
  // Monitoring operations
  fetchAlerts: () => Promise<void>;
  acknowledgeAlert: (id: string) => Promise<void>;
  resolveAlert: (id: string) => Promise<void>;
  fetchMetrics: (sourceType?: string, sourceId?: string) => Promise<void>;
  getNetworkMetrics: (id: string) => Promise<Metric[]>;
  getNodeMetrics: (id: string) => Promise<Metric[]>;
  getNetworkHealth: (id: string) => Promise<Health>;
  getNodeHealth: (id: string) => Promise<Health>;
}

const ApiContext = createContext<ApiContextType | undefined>(undefined);

export const ApiProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const { api } = useAuth();
  const [networks, setNetworks] = useState<Network[]>([]);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [links, setLinks] = useState<Link[]>([]);
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleError = (error: any) => {
    const message = error.response?.data?.error || error.message || 'An error occurred';
    setError(message);
    console.error('API Error:', error);
    throw error; // Re-throw to allow callers to handle
  };

  // Network operations
  const fetchNetworks = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get('/api/v1/networks');
      setNetworks(response.data.networks || []);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const createNetwork = useCallback(async (network: Partial<Network>): Promise<Network> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.post('/api/v1/networks', network);
      const newNetwork = response.data.network;
      setNetworks(prev => [...prev, newNetwork]);
      return newNetwork;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const updateNetwork = useCallback(async (id: string, network: Partial<Network>): Promise<Network> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.put(`/api/v1/networks/${id}`, network);
      const updatedNetwork = response.data.network;
      setNetworks(prev => prev.map(n => n.id === id ? updatedNetwork : n));
      return updatedNetwork;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const deleteNetwork = useCallback(async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      await api.delete(`/api/v1/networks/${id}`);
      setNetworks(prev => prev.filter(n => n.id !== id));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const getNetwork = useCallback(async (id: string): Promise<Network> => {
    try {
      const response = await api.get(`/api/v1/networks/${id}`);
      return response.data.network;
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  const startNetwork = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/networks/${id}/start`);
      // Update local state
      setNetworks(prev => prev.map(n => n.id === id ? { ...n, status: 'starting' } : n));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  const stopNetwork = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/networks/${id}/stop`);
      setNetworks(prev => prev.map(n => n.id === id ? { ...n, status: 'stopping' } : n));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  const restartNetwork = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/networks/${id}/restart`);
      setNetworks(prev => prev.map(n => n.id === id ? { ...n, status: 'restarting' } : n));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  // Node operations
  const fetchNodes = useCallback(async (networkId?: string) => {
    try {
      setLoading(true);
      setError(null);
      
      // NOTE: Nodes endpoint not implemented yet, using mock data
      console.warn('Nodes endpoint not available, using mock data');
      setNodes([
        {
          id: 'node-1',
          network_id: networkId || 'network-1',
          name: 'Router Node',
          type: 'router', 
          ip_address: '192.168.1.1',
          mac_address: '00:11:22:33:44:55',
          status: 'active',
          config: {},
          position: { x: 100, y: 100 },
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'node-2',
          network_id: networkId || 'network-1', 
          name: 'Switch Node',
          type: 'switch',
          ip_address: '192.168.1.2',
          mac_address: '00:11:22:33:44:56',
          status: 'inactive',
          config: {},
          position: { x: 200, y: 150 },
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
      ]);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const createNode = useCallback(async (networkId: string, node: Partial<Node>): Promise<Node> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.post(`/api/v1/networks/${networkId}/nodes`, node);
      const newNode = response.data.node;
      setNodes(prev => [...prev, newNode]);
      return newNode;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const updateNode = useCallback(async (id: string, node: Partial<Node>): Promise<Node> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.put(`/api/v1/nodes/${id}`, node);
      const updatedNode = response.data.node;
      setNodes(prev => prev.map(n => n.id === id ? updatedNode : n));
      return updatedNode;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const deleteNode = useCallback(async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      await api.delete(`/api/v1/nodes/${id}`);
      setNodes(prev => prev.filter(n => n.id !== id));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const getNode = useCallback(async (id: string): Promise<Node> => {
    try {
      const response = await api.get(`/api/v1/nodes/${id}`);
      return response.data.node;
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  const startNode = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/nodes/${id}/start`);
      setNodes(prev => prev.map(n => n.id === id ? { ...n, status: 'starting' } : n));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  const stopNode = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/nodes/${id}/stop`);
      setNodes(prev => prev.map(n => n.id === id ? { ...n, status: 'stopping' } : n));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  const restartNode = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/nodes/${id}/restart`);
      setNodes(prev => prev.map(n => n.id === id ? { ...n, status: 'restarting' } : n));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  // Link operations
  const fetchLinks = useCallback(async (networkId: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get(`/api/v1/networks/${networkId}/links`);
      setLinks(response.data.links || []);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const createLink = useCallback(async (networkId: string, link: Partial<Link>): Promise<Link> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.post(`/api/v1/networks/${networkId}/links`, link);
      const newLink = response.data.link;
      setLinks(prev => [...prev, newLink]);
      return newLink;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const updateLink = useCallback(async (id: string, link: Partial<Link>): Promise<Link> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.put(`/api/v1/links/${id}`, link);
      const updatedLink = response.data.link;
      setLinks(prev => prev.map(l => l.id === id ? updatedLink : l));
      return updatedLink;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const deleteLink = useCallback(async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      await api.delete(`/api/v1/links/${id}`);
      setLinks(prev => prev.filter(l => l.id !== id));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const getLink = useCallback(async (id: string): Promise<Link> => {
    try {
      const response = await api.get(`/api/v1/links/${id}`);
      return response.data.link;
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  // Policy operations
  const fetchPolicies = useCallback(async (networkId: string) => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get(`/api/v1/networks/${networkId}/policies`);
      setPolicies(response.data.policies || []);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const createPolicy = useCallback(async (networkId: string, policy: Partial<Policy>): Promise<Policy> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.post(`/api/v1/networks/${networkId}/policies`, policy);
      const newPolicy = response.data.policy;
      setPolicies(prev => [...prev, newPolicy]);
      return newPolicy;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const updatePolicy = useCallback(async (id: string, policy: Partial<Policy>): Promise<Policy> => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.put(`/api/v1/policies/${id}`, policy);
      const updatedPolicy = response.data.policy;
      setPolicies(prev => prev.map(p => p.id === id ? updatedPolicy : p));
      return updatedPolicy;
    } catch (error) {
      handleError(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, [api]);

  const deletePolicy = useCallback(async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      await api.delete(`/api/v1/policies/${id}`);
      setPolicies(prev => prev.filter(p => p.id !== id));
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const getPolicy = useCallback(async (id: string): Promise<Policy> => {
    try {
      const response = await api.get(`/api/v1/policies/${id}`);
      return response.data.policy;
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  // Monitoring operations
  const fetchAlerts = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      // NOTE: Alerts endpoint not implemented yet, using mock data
      console.warn('Alerts endpoint not available, using mock data');
      setAlerts([
        {
          id: 'alert-1',
          title: 'Network Connectivity Issue',
          description: 'Node 2 appears to be offline',
          severity: 'high',
          status: 'active',
          source_type: 'monitoring',
          source_id: 'system',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'alert-2', 
          title: 'High CPU Usage',
          description: 'Router node CPU usage above 80%',
          severity: 'medium',
          status: 'active',
          source_type: 'monitoring',
          source_id: 'system',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
      ]);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const acknowledgeAlert = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/monitoring/alerts/${id}/acknowledge`);
      setAlerts(prev => prev.map(a => a.id === id ? { ...a, status: 'acknowledged' } : a));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  const resolveAlert = useCallback(async (id: string) => {
    try {
      await api.post(`/api/v1/monitoring/alerts/${id}/resolve`);
      setAlerts(prev => prev.map(a => a.id === id ? { ...a, status: 'resolved' } : a));
    } catch (error) {
      handleError(error);
    }
  }, [api]);

  const fetchMetrics = useCallback(async (sourceType?: string, sourceId?: string) => {
    try {
      setLoading(true);
      setError(null);
      let url = '/api/v1/monitoring/metrics';
      const params = new URLSearchParams();
      if (sourceType) params.append('source_type', sourceType);
      if (sourceId) params.append('source_id', sourceId);
      if (params.toString()) url += `?${params.toString()}`;
      
      const response = await api.get(url);
      setMetrics(response.data.metrics || []);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const getNetworkMetrics = useCallback(async (id: string): Promise<Metric[]> => {
    try {
      const response = await api.get(`/api/v1/monitoring/networks/${id}/metrics`);
      return response.data.metrics || [];
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  const getNodeMetrics = useCallback(async (id: string): Promise<Metric[]> => {
    try {
      const response = await api.get(`/api/v1/monitoring/nodes/${id}/metrics`);
      return response.data.metrics || [];
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  const getNetworkHealth = useCallback(async (id: string): Promise<Health> => {
    try {
      const response = await api.get(`/api/v1/monitoring/networks/${id}/health`);
      return response.data.health;
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  const getNodeHealth = useCallback(async (id: string): Promise<Health> => {
    try {
      const response = await api.get(`/api/v1/monitoring/nodes/${id}/health`);
      return response.data.health;
    } catch (error) {
      handleError(error);
      throw error;
    }
  }, [api]);

  const value: ApiContextType = {
    // Data state
    networks,
    nodes,
    links,
    policies,
    alerts,
    metrics,
    loading,
    error,
    
    // Network operations
    fetchNetworks,
    createNetwork,
    updateNetwork,
    deleteNetwork,
    getNetwork,
    startNetwork,
    stopNetwork,
    restartNetwork,
    
    // Node operations
    fetchNodes,
    createNode,
    updateNode,
    deleteNode,
    getNode,
    startNode,
    stopNode,
    restartNode,
    
    // Link operations
    fetchLinks,
    createLink,
    updateLink,
    deleteLink,
    getLink,
    
    // Policy operations
    fetchPolicies,
    createPolicy,
    updatePolicy,
    deletePolicy,
    getPolicy,
    
    // Monitoring operations
    fetchAlerts,
    acknowledgeAlert,
    resolveAlert,
    fetchMetrics,
    getNetworkMetrics,
    getNodeMetrics,
    getNetworkHealth,
    getNodeHealth,
  };

  return (
    <ApiContext.Provider value={value}>
      {children}
    </ApiContext.Provider>
  );
};

export const useApi = (): ApiContextType => {
  const context = useContext(ApiContext);
  if (!context) {
    throw new Error('useApi must be used within an ApiProvider');
  }
  return context;
};
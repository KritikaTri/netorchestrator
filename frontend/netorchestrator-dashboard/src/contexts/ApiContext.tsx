import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import axios, { AxiosInstance } from 'axios';

interface Network {
  id: string;
  name: string;
  description: string;
  type: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface Node {
  id: string;
  network_id: string;
  name: string;
  type: string;
  ip_address: string;
  status: string;
  config: any;
}

interface Metric {
  id: string;
  source_type: string;
  source_id: string;
  metric_name: string;
  value: number;
  timestamp: string;
}

interface ApiContextType {
  networks: Network[];
  nodes: Node[];
  metrics: Metric[];
  loading: boolean;
  error: string | null;
  fetchNetworks: () => Promise<void>;
  fetchNodes: () => Promise<void>;
  fetchMetrics: () => Promise<void>;
  createNetwork: (network: Partial<Network>) => Promise<void>;
  deleteNetwork: (id: string) => Promise<void>;
  api: AxiosInstance;
}

const ApiContext = createContext<ApiContextType | undefined>(undefined);

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080';

export const ApiProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [networks, setNetworks] = useState<Network[]>([]);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [metrics, setMetrics] = useState<Metric[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const api = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
  });

  const handleError = (error: any) => {
    const message = error.response?.data?.error || error.message || 'An error occurred';
    setError(message);
    console.error('API Error:', error);
  };

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

  const fetchNodes = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get('/api/v1/nodes');
      setNodes(response.data.nodes || []);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const fetchMetrics = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get('/api/v1/metrics');
      setMetrics(response.data.metrics || []);
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api]);

  const createNetwork = useCallback(async (network: Partial<Network>) => {
    try {
      setLoading(true);
      setError(null);
      await api.post('/api/v1/networks', network);
      await fetchNetworks(); // Refresh the list
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api, fetchNetworks]);

  const deleteNetwork = useCallback(async (id: string) => {
    try {
      setLoading(true);
      setError(null);
      await api.delete(`/api/v1/networks/${id}`);
      await fetchNetworks(); // Refresh the list
    } catch (error) {
      handleError(error);
    } finally {
      setLoading(false);
    }
  }, [api, fetchNetworks]);

  const value: ApiContextType = {
    networks,
    nodes,
    metrics,
    loading,
    error,
    fetchNetworks,
    fetchNodes,
    fetchMetrics,
    createNetwork,
    deleteNetwork,
    api,
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
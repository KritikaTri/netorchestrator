import React, { useState, useEffect, useCallback } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Alert,
  CircularProgress,
  Chip,
  Paper,
  IconButton,
} from '@mui/material';
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  Node as FlowNode,
  Edge as FlowEdge,
  useNodesState,
  useEdgesState,
  ConnectionMode,
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';
import { useAuth } from '../contexts/AuthContext';
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';
import {
  Refresh as RefreshIcon,
  Warning as WarningIcon,
  CheckCircle as ActiveIcon,
  Error as ErrorIcon,
} from '@mui/icons-material';
import NetworkTopologyCanvas from './NetworkTopologyCanvas';
import TopologyInfoPanel from './TopologyInfoPanel';

interface Network {
  id: string;
  name: string;
  description: string;
  status: string;
  config?: {
    subnet?: string;
    gateway?: string;
  };
}

interface TopologyNode {
  id: string;
  network_id: string;
  name: string;
  type: string;
  ip_address: string;
  status: string;
  config?: {
    cpu?: number;
    memory?: number;
    storage?: number;
  };
  position?: {
    x: number;
    y: number;
  };
  entity_id?: string;
}

interface TopologyLink {
  id: string;
  network_id: string;
  source_node_id: string;
  target_node_id: string;
  link_type: string;
  bandwidth_mbps?: number;
  latency_ms?: number;
  status: string;
}

interface TopologyData {
  network: Network;
  nodes: TopologyNode[];
  links: TopologyLink[];
}

interface NetworkMetrics {
  network_id: string;
  network_name: string;
  containers: {
    total_active: number;
    by_node_type: { [key: string]: number };
  };
  links: {
    total_links: number;
    active_links: number;
    down_links: number;
  };
}

interface NetworkTopologyProps {
  selectedNetworkId?: string;
  onNetworkSelect?: (networkId: string) => void;
}

const NetworkTopology: React.FC<NetworkTopologyProps> = ({ 
  selectedNetworkId, 
  onNetworkSelect 
}) => {
  const { api } = useAuth();
  const [networks, setNetworks] = useState<Network[]>([]);
  const [selectedNetwork, setSelectedNetwork] = useState<string>(selectedNetworkId || '');
  const [topologyData, setTopologyData] = useState<TopologyData | null>(null);
  const [metricsData, setMetricsData] = useState<NetworkMetrics | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch networks on mount
  useEffect(() => {
    fetchNetworks();
  }, []);

  // Fetch topology data when network is selected
  useEffect(() => {
    if (selectedNetwork) {
      fetchTopologyData();
      fetchMetricsData();
      
      // Set up auto-refresh for metrics
      const interval = setInterval(fetchMetricsData, 10000);
      return () => clearInterval(interval);
    }
  }, [selectedNetwork]);

  const fetchNetworks = async () => {
    try {
      const response = await api.get('/api/v1/networks');
      const networkList = response.data.networks || response.data || [];
      setNetworks(networkList);
      
      // Auto-select first network if none selected
      if (!selectedNetwork && networkList.length > 0) {
        setSelectedNetwork(networkList[0].id);
        if (onNetworkSelect) {
          onNetworkSelect(networkList[0].id);
        }
      }
    } catch (err) {
      console.error('Failed to fetch networks:', err);
      setError('Failed to load networks');
    }
  };

  const fetchTopologyData = async () => {
    if (!selectedNetwork) return;
    
    setLoading(true);
    setError(null);
    
    try {
      const response = await api.get(`/api/v1/networks/${selectedNetwork}/enhanced-topology`);
      const data = response.data.topology || response.data;
      setTopologyData(data);
    } catch (err) {
      console.error('Failed to fetch topology:', err);
      setError('Failed to load network topology');
    } finally {
      setLoading(false);
    }
  };

  const fetchMetricsData = async () => {
    if (!selectedNetwork) return;
    
    try {
      const response = await api.get(`/api/v1/metrics/network/${selectedNetwork}`);
      setMetricsData(response.data);
    } catch (err) {
      console.error('Failed to fetch metrics:', err);
      // Don't set error for metrics failure, just log it
    }
  };

  const handleNetworkChange = (networkId: string) => {
    setSelectedNetwork(networkId);
    if (onNetworkSelect) {
      onNetworkSelect(networkId);
    }
  };

  const handleRefresh = () => {
    if (selectedNetwork) {
      fetchTopologyData();
      fetchMetricsData();
    }
  };

  if (loading && !topologyData) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '600px' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      {/* Header Controls */}
      <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, flexWrap: 'wrap' }}>
          <Box sx={{ minWidth: 300, flex: { xs: '1 1 100%', sm: '1 1 300px', md: '0 0 400px' } }}>
            <FormControl fullWidth>
              <InputLabel>Select Network</InputLabel>
              <Select
                value={selectedNetwork}
                onChange={(e) => handleNetworkChange(e.target.value)}
                label="Select Network"
              >
                {networks.map((network) => (
                  <MenuItem key={network.id} value={network.id}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Typography>{network.name}</Typography>
                      <Chip
                        size="small"
                        label={network.status}
                        color={network.status === 'active' ? 'success' : 'default'}
                        variant="outlined"
                      />
                    </Box>
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Box>
          
          <Box>
            <IconButton onClick={handleRefresh} disabled={loading}>
              <RefreshIcon />
            </IconButton>
          </Box>
        </Box>

        {error && (
          <Alert severity="error" sx={{ mt: 2 }}>
            {error}
          </Alert>
        )}
      </Box>

      {/* Main Topology View */}
      {topologyData ? (
        <Box sx={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
          {/* Left Side: Topology Canvas */}
          <Box sx={{ flex: 1, position: 'relative' }}>
            <NetworkTopologyCanvas 
              topologyData={topologyData}
              metricsData={metricsData}
              onNodeSelect={(nodeId) => console.log('Selected node:', nodeId)}
            />
          </Box>

          {/* Right Side: Information Panel */}
          <Box sx={{ width: '350px', borderLeft: 1, borderColor: 'divider' }}>
            <TopologyInfoPanel 
              topologyData={topologyData}
              metricsData={metricsData}
            />
          </Box>
        </Box>
      ) : (
        <Box sx={{ 
          flex: 1, 
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: 'center',
          flexDirection: 'column',
          gap: 2 
        }}>
          <Typography variant="h6" color="text.secondary">
            {selectedNetwork ? 'Loading topology...' : 'Select a network to view topology'}
          </Typography>
          {selectedNetwork && <CircularProgress />}
        </Box>
      )}
    </Box>
  );
};

export default NetworkTopology;
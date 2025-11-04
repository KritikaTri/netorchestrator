import React, { useMemo } from 'react';
import {
  Box,
  Typography,
  Card,
  CardContent,
  Chip,
  Divider,
  LinearProgress,
  IconButton,
} from '@mui/material';
import {
  Error as ErrorIcon,
  Warning as WarningIcon,
  CheckCircle as ActiveIcon,
  Computer as DevicesIcon,
  Link as LinksIcon,
  Storage as HostsIcon,
  Add as AddIcon,
  Remove as RemoveIcon,
} from '@mui/icons-material';
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer } from 'recharts';

interface TopologyNodeData {
  id: string;
  name: string;
  type: string;
  ip_address: string;
  status: string;
  config?: {
    cpu?: number;
    memory?: number;
    storage?: number;
  };
}

interface TopologyData {
  network: {
    id: string;
    name: string;
    description: string;
    status: string;
    config?: {
      subnet?: string;
      gateway?: string;
    };
  };
  nodes: TopologyNodeData[];
  links: any[];
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

interface TopologyInfoPanelProps {
  topologyData: TopologyData;
  metricsData: NetworkMetrics | null;
}

// Mock trend data - in a real app this would come from historical metrics
const generateTrendData = (baseValue: number) => {
  const months = ['JUN', 'JUL', 'AUG', 'SEP', 'OCT'];
  return months.map((month, index) => ({
    month,
    value: baseValue + Math.random() * 20 - 10, // Slight variation around base value
  }));
};

const TopologyInfoPanel: React.FC<TopologyInfoPanelProps> = ({
  topologyData,
  metricsData,
}) => {
  // Calculate summary statistics
  const summaryStats = useMemo(() => {
    const activeNodes = topologyData.nodes.filter(node => node.status === 'active').length;
    const errorNodes = topologyData.nodes.filter(node => node.status === 'error').length;
    const totalLinks = topologyData.links.length;
    const activeLinks = metricsData?.links.active_links ?? topologyData.links.filter(link => link.status === 'active').length;
    
    return {
      devices: topologyData.nodes.length,
      links: totalLinks,
      hosts: activeNodes, // Assuming hosts are active nodes
      activeNodes,
      errorNodes,
      activeLinks,
    };
  }, [topologyData, metricsData]);

  // Get primary network/datacenter name
  const datacenterName = topologyData.network.name.toUpperCase().replace(' NETWORK', ' DATA CENTER');

  // Mock capacity data - in real app would come from aggregated node metrics
  const capacityData = {
    compute: 40, // Percentage
    memory: 15,
    storage: 26,
  };

  const computeTrendData = generateTrendData(capacityData.compute);
  const memoryTrendData = generateTrendData(capacityData.memory);
  const storageTrendData = generateTrendData(capacityData.storage);

  return (
    <Box sx={{ 
      height: '100%', 
      backgroundColor: '#1a1b23', 
      color: 'white',
      display: 'flex',
      flexDirection: 'column',
    }}>
      {/* Header */}
      <Box sx={{ p: 2, borderBottom: '1px solid #2d3748' }}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <Box sx={{ 
            width: 8, 
            height: 8, 
            backgroundColor: '#ef4444', 
            borderRadius: '50%' 
          }} />
          <Box sx={{ 
            width: 8, 
            height: 8, 
            backgroundColor: '#ef4444', 
            borderRadius: '50%' 
          }} />
        </Box>
        <Typography variant="h6" sx={{ fontWeight: 600, mb: 0.5 }}>
          {datacenterName}
        </Typography>
        
        {/* Capacity Section */}
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
          <Chip 
            label="CAPACITY" 
            sx={{ 
              backgroundColor: '#ef4444',
              color: 'white',
              fontSize: '10px',
              height: 20,
              fontWeight: 600,
            }} 
          />
          <Typography variant="body2" sx={{ color: '#a0aec0' }}>
            ALARMS
          </Typography>
          
          {/* Alarms indicators */}
          <Box sx={{ display: 'flex', gap: 0.5, ml: 'auto' }}>
            {summaryStats.errorNodes > 0 && (
              <ErrorIcon sx={{ color: '#ef4444', fontSize: 16 }} />
            )}
            <WarningIcon sx={{ color: '#f59e0b', fontSize: 16 }} />
          </Box>
        </Box>
      </Box>

      {/* Metrics Section */}
      <Box sx={{ flex: 1, p: 2, overflow: 'auto' }}>
        {/* Compute Metrics */}
        <Box sx={{ mb: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
            <Typography variant="body2" sx={{ color: '#a0aec0' }}>
              Compute, mCore
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {capacityData.compute}%
            </Typography>
          </Box>
          
          <Box sx={{ height: 60, mb: 1 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={computeTrendData}>
                <Line 
                  type="monotone" 
                  dataKey="value" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  dot={false}
                />
                <XAxis 
                  dataKey="month" 
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 10, fill: '#6b7280' }}
                />
                <YAxis hide />
              </LineChart>
            </ResponsiveContainer>
          </Box>
          
          <Box sx={{ display: 'flex', justifyContent: 'space-between', fontSize: '10px', color: '#6b7280' }}>
            <span>JUN</span>
            <span>JUL</span>
            <span>AUG</span>
            <span>SEP</span>
            <span>OCT</span>
          </Box>
        </Box>

        {/* Memory Metrics */}
        <Box sx={{ mb: 3 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
            <Typography variant="body2" sx={{ color: '#a0aec0' }}>
              Memory, GiB
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {capacityData.memory}%
            </Typography>
          </Box>
          
          <Box sx={{ height: 60, mb: 1 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={memoryTrendData}>
                <Line 
                  type="monotone" 
                  dataKey="value" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  dot={false}
                />
                <XAxis 
                  dataKey="month" 
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 10, fill: '#6b7280' }}
                />
                <YAxis hide />
              </LineChart>
            </ResponsiveContainer>
          </Box>
          
          <Box sx={{ display: 'flex', justifyContent: 'space-between', fontSize: '10px', color: '#6b7280' }}>
            <span>JUN</span>
            <span>JUL</span>
            <span>AUG</span>
            <span>SEP</span>
            <span>OCT</span>
          </Box>
        </Box>

        {/* Storage Metrics */}
        <Box sx={{ mb: 4 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 1 }}>
            <Typography variant="body2" sx={{ color: '#a0aec0' }}>
              Storage, GiB
            </Typography>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              {capacityData.storage}%
            </Typography>
          </Box>
          
          <Box sx={{ height: 60, mb: 1 }}>
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={storageTrendData}>
                <Line 
                  type="monotone" 
                  dataKey="value" 
                  stroke="#3b82f6" 
                  strokeWidth={2}
                  dot={false}
                />
                <XAxis 
                  dataKey="month" 
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 10, fill: '#6b7280' }}
                />
                <YAxis hide />
              </LineChart>
            </ResponsiveContainer>
          </Box>
          
          <Box sx={{ display: 'flex', justifyContent: 'space-between', fontSize: '10px', color: '#6b7280' }}>
            <span>JUN</span>
            <span>JUL</span>
            <span>AUG</span>
            <span>SEP</span>
            <span>OCT</span>
          </Box>
        </Box>

        {/* Summary Section */}
        <Box sx={{ 
          backgroundColor: '#2d3748', 
          borderRadius: 1, 
          p: 2,
          mt: 2,
        }}>
          <Typography variant="h6" sx={{ mb: 2, fontWeight: 600 }}>
            SUMMARY
          </Typography>
          
          <Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <DevicesIcon sx={{ fontSize: 16, color: '#6b7280' }} />
                <Typography variant="body2" sx={{ color: '#a0aec0' }}>
                  Devices
                </Typography>
              </Box>
              <Typography variant="h4" sx={{ fontWeight: 600, color: '#3b82f6' }}>
                {summaryStats.devices}
              </Typography>
            </Box>
            
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <LinksIcon sx={{ fontSize: 16, color: '#6b7280' }} />
                <Typography variant="body2" sx={{ color: '#a0aec0' }}>
                  Links
                </Typography>
              </Box>
              <Typography variant="h4" sx={{ fontWeight: 600, color: '#3b82f6' }}>
                {summaryStats.links}
              </Typography>
            </Box>
            
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                <HostsIcon sx={{ fontSize: 16, color: '#6b7280' }} />
                <Typography variant="body2" sx={{ color: '#a0aec0' }}>
                  Hosts
                </Typography>
              </Box>
              <Typography variant="h4" sx={{ fontWeight: 600, color: '#3b82f6' }}>
                {summaryStats.hosts}
              </Typography>
            </Box>
          </Box>
        </Box>

        {/* Add/Remove Controls */}
        <Box sx={{ 
          display: 'flex', 
          justifyContent: 'center', 
          gap: 2, 
          mt: 3,
          pb: 2,
        }}>
          <IconButton 
            sx={{ 
              backgroundColor: '#10b981',
              color: 'white',
              '&:hover': { backgroundColor: '#059669' }
            }}
          >
            <AddIcon />
          </IconButton>
          <IconButton 
            sx={{ 
              backgroundColor: '#6b7280',
              color: 'white',
              '&:hover': { backgroundColor: '#4b5563' }
            }}
          >
            <RemoveIcon />
          </IconButton>
        </Box>
      </Box>
    </Box>
  );
};

export default TopologyInfoPanel;

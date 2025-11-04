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


  return (
    <Box sx={{ 
      height: '100%', 
      backgroundColor: '#0f172a', 
      backgroundImage: 'radial-gradient(circle, #1e293b 1px, transparent 1px)',
      backgroundSize: '20px 20px',
      color: 'white',
      display: 'flex',
      flexDirection: 'column',
    }}>
      {/* Header */}
      <Box sx={{ p: 2, borderBottom: '1px solid #1e293b' }}>
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
        
        {/* Network Info Section */}
        <Box sx={{ mb: 1 }}>
          <Typography variant="body2" sx={{ color: '#a0aec0', fontSize: '12px' }}>
            Network Topology Overview
          </Typography>
          <Typography variant="caption" sx={{ color: '#6b7280', fontSize: '10px' }}>
            Real-time network infrastructure visualization
          </Typography>
        </Box>
      </Box>

      {/* Metrics Section */}
      <Box sx={{ flex: 1, p: 2, overflow: 'auto' }}>

        {/* Summary Section */}
        <Box sx={{ 
          backgroundColor: '#1e293b', 
          backgroundImage: 'radial-gradient(circle, #334155 1px, transparent 1px)',
          backgroundSize: '15px 15px',
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

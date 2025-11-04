import React from 'react';
import { Handle, Position } from '@xyflow/react';
import { Box, Typography, Chip } from '@mui/material';
import {
  Router as RouterIcon,
  Hub as SwitchIcon,
  Computer as HostIcon,
  Security as FirewallIcon,
  Storage as StorageIcon,
  AccountTree as GatewayIcon,
} from '@mui/icons-material';

interface TopologyNodeData {
  id: string;
  name: string;
  type: string;
  ip_address: string;
  status: string;
  utilization: number;
  isSelected: boolean;
}

interface TopologyNodeProps {
  data: TopologyNodeData;
}

const getNodeIcon = (type: string) => {
  switch (type.toLowerCase()) {
    case 'router':
      return <RouterIcon sx={{ fontSize: 20 }} />;
    case 'switch':
      return <SwitchIcon sx={{ fontSize: 20 }} />;
    case 'host':
    case 'container':
      return <HostIcon sx={{ fontSize: 20 }} />;
    case 'firewall':
      return <FirewallIcon sx={{ fontSize: 20 }} />;
    case 'storage':
      return <StorageIcon sx={{ fontSize: 20 }} />;
    case 'gateway':
      return <GatewayIcon sx={{ fontSize: 20 }} />;
    default:
      return <HostIcon sx={{ fontSize: 20 }} />;
  }
};

const getStatusColor = (status: string, utilization: number) => {
  if (status === 'error') return '#ef4444';
  if (utilization > 80) return '#ef4444'; // Red for high utilization
  if (utilization > 50) return '#f59e0b'; // Orange for medium utilization
  if (status === 'active') return '#10b981'; // Green for normal
  return '#6b7280'; // Gray for inactive/unknown
};

const getUtilizationColor = (utilization: number) => {
  if (utilization > 80) return '#ef4444';
  if (utilization > 50) return '#f59e0b';
  return '#10b981';
};

const TopologyNode: React.FC<TopologyNodeProps> = ({ data }) => {
  const statusColor = getStatusColor(data.status, data.utilization);
  const utilizationColor = getUtilizationColor(data.utilization);
  
  return (
    <Box
      sx={{
        position: 'relative',
        minWidth: 120,
        minHeight: 90,
        backgroundColor: '#2d3748',
        border: `2px solid ${data.isSelected ? '#3b82f6' : statusColor}`,
        borderRadius: '50%',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: 'pointer',
        transition: 'all 0.2s ease-in-out',
        boxShadow: data.isSelected 
          ? '0 0 20px rgba(59, 130, 246, 0.5)' 
          : '0 4px 8px rgba(0, 0, 0, 0.3)',
        '&:hover': {
          transform: 'scale(1.05)',
          boxShadow: '0 8px 16px rgba(0, 0, 0, 0.4)',
        },
      }}
    >
      {/* Connection Handles */}
      <Handle
        type="target"
        position={Position.Top}
        style={{
          background: statusColor,
          border: '2px solid white',
          width: 8,
          height: 8,
        }}
      />
      <Handle
        type="source"
        position={Position.Bottom}
        style={{
          background: statusColor,
          border: '2px solid white',
          width: 8,
          height: 8,
        }}
      />
      <Handle
        type="target"
        position={Position.Left}
        style={{
          background: statusColor,
          border: '2px solid white',
          width: 8,
          height: 8,
        }}
      />
      <Handle
        type="source"
        position={Position.Right}
        style={{
          background: statusColor,
          border: '2px solid white',
          width: 8,
          height: 8,
        }}
      />

      {/* Node Icon */}
      <Box sx={{ color: 'white', mb: 0.2 }}>
        {getNodeIcon(data.type)}
      </Box>

      {/* Device Type Text */}
      <Box sx={{ 
        color: '#a0aec0', 
        fontSize: '10px', 
        fontWeight: 500,
        textTransform: 'capitalize',
        lineHeight: 1,
        textAlign: 'center',
      }}>
        {data.type}
      </Box>

      {/* Utilization Percentage */}
      <Box
        sx={{
          position: 'absolute',
          top: -10,
          right: -10,
          backgroundColor: utilizationColor,
          color: 'white',
          borderRadius: '50%',
          width: 32,
          height: 32,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '12px',
          fontWeight: 'bold',
          border: '2px solid white',
          boxShadow: '0 2px 4px rgba(0, 0, 0, 0.3)',
        }}
      >
        {data.utilization}%
      </Box>

      {/* Node Name Label */}
      <Box
        sx={{
          position: 'absolute',
          bottom: -40,
          left: '50%',
          transform: 'translateX(-50%)',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          color: 'white',
          padding: '4px 8px',
          borderRadius: 1,
          fontSize: '11px',
          fontWeight: 500,
          whiteSpace: 'nowrap',
          maxWidth: 120,
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        }}
      >
        {data.name}
      </Box>

      {/* Status Indicator */}
      {data.status === 'error' && (
        <Box
          sx={{
            position: 'absolute',
            top: -5,
            left: -5,
            width: 12,
            height: 12,
            backgroundColor: '#ef4444',
            borderRadius: '50%',
            border: '2px solid white',
            animation: 'pulse 2s infinite',
            '@keyframes pulse': {
              '0%': { opacity: 1 },
              '50%': { opacity: 0.5 },
              '100%': { opacity: 1 },
            },
          }}
        />
      )}
    </Box>
  );
};

export default TopologyNode;

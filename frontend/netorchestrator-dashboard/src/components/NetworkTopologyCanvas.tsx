import React, { useMemo, useCallback, useState } from 'react';
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
  MarkerType,
} from '@xyflow/react';
import { Box, Typography } from '@mui/material';
import '@xyflow/react/dist/style.css';
import TopologyNode from './TopologyNode';

interface TopologyNodeData {
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

interface TopologyLinkData {
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
  links: TopologyLinkData[];
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

interface NetworkTopologyCanvasProps {
  topologyData: TopologyData;
  metricsData: NetworkMetrics | null;
  onNodeSelect: (nodeId: string) => void;
}

const nodeTypes = {
  topology: TopologyNode,
};

// Function to calculate utilization percentage for a node
const getNodeUtilization = (nodeId: string, metricsData: NetworkMetrics | null): number => {
  // For now, return mock utilization data based on node position
  // In real implementation, this would come from container metrics
  const mockUtilizations = [9, 20, 10, 2, 5, 75, 18];
  const hash = nodeId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  return mockUtilizations[hash % mockUtilizations.length];
};

// Function to generate automatic layout positions if not provided
const generateLayout = (nodes: TopologyNodeData[]): { [key: string]: { x: number; y: number } } => {
  const positions: { [key: string]: { x: number; y: number } } = {};
  
  // Use a simple circular layout if no positions are provided
  const centerX = 400;
  const centerY = 300;
  const radius = 200;
  
  nodes.forEach((node, index) => {
    if (node.position) {
      positions[node.id] = node.position;
    } else {
      const angle = (index / nodes.length) * 2 * Math.PI;
      positions[node.id] = {
        x: centerX + radius * Math.cos(angle),
        y: centerY + radius * Math.sin(angle),
      };
    }
  });
  
  return positions;
};

const NetworkTopologyCanvas: React.FC<NetworkTopologyCanvasProps> = ({
  topologyData,
  metricsData,
  onNodeSelect,
}) => {
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  // Convert topology data to React Flow format
  const { initialNodes, initialEdges } = useMemo(() => {
    const positions = generateLayout(topologyData.nodes);
    
    const nodes: FlowNode[] = topologyData.nodes.map((node) => ({
      id: node.id,
      type: 'topology',
      position: positions[node.id],
      data: {
        ...node,
        utilization: getNodeUtilization(node.id, metricsData),
        isSelected: selectedNode === node.id,
      },
      draggable: true,
    }));

    const edges: FlowEdge[] = topologyData.links.map((link) => ({
      id: link.id,
      source: link.source_node_id,
      target: link.target_node_id,
      type: 'default', // Changed from 'smoothstep' to 'default'
      animated: link.status === 'active',
      style: {
        stroke: link.status === 'active' ? '#10b981' : '#ef4444',
        strokeWidth: 3, // Increased thickness
      },
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: link.status === 'active' ? '#10b981' : '#ef4444',
        width: 20,
        height: 20,
      },
      label: '', // Remove labels entirely for cleaner look
      labelStyle: {
        fontSize: 10,
        fontWeight: 300,
        fill: '#6b7280',
      },
      labelBgStyle: {
        fill: 'transparent',
        fillOpacity: 0,
      },
    }));

    return { initialNodes: nodes, initialEdges: edges };
  }, [topologyData, metricsData, selectedNode]);

  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

  // Update nodes and edges when topology data or metrics change
  React.useEffect(() => {
    const positions = generateLayout(topologyData.nodes);
    
    const updatedNodes: FlowNode[] = topologyData.nodes.map((node) => ({
      id: node.id,
      type: 'topology',
      position: positions[node.id],
      data: {
        ...node,
        utilization: getNodeUtilization(node.id, metricsData),
        isSelected: selectedNode === node.id,
      },
      draggable: true,
    }));

    const updatedEdges: FlowEdge[] = topologyData.links.map((link) => ({
      id: link.id,
      source: link.source_node_id,
      target: link.target_node_id,
      type: 'default',
      animated: link.status === 'active',
      style: {
        stroke: link.status === 'active' ? '#10b981' : '#ef4444',
        strokeWidth: 3,
      },
      markerEnd: {
        type: MarkerType.ArrowClosed,
        color: link.status === 'active' ? '#10b981' : '#ef4444',
        width: 20,
        height: 20,
      },
      label: '', // Remove labels entirely for cleaner look
      labelStyle: {
        fontSize: 10,
        fontWeight: 300,
        fill: '#6b7280',
      },
      labelBgStyle: {
        fill: 'transparent',
        fillOpacity: 0,
      },
    }));

    setNodes(updatedNodes);
    setEdges(updatedEdges);
  }, [topologyData, metricsData, selectedNode, setNodes, setEdges]);

  const onNodeClick = useCallback(
    (event: React.MouseEvent, node: FlowNode) => {
      setSelectedNode(node.id);
      onNodeSelect(node.id);
    },
    [onNodeSelect]
  );

  const onPaneClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  return (
    <Box sx={{ 
      width: '100%', 
      height: '100%', 
      backgroundColor: '#0f172a',
      backgroundImage: 'radial-gradient(circle, #1e293b 1px, transparent 1px)',
      backgroundSize: '20px 20px',
      position: 'relative',
      '& .react-flow__edge': {
        strokeWidth: '3px !important',
        stroke: '#10b981 !important',
        zIndex: 1000,
      },
      '& .react-flow__edge-path': {
        strokeWidth: '3px !important',
        stroke: '#10b981 !important',
      },
      '& .react-flow__arrowhead': {
        fill: '#10b981 !important',
      },
      '& .react-flow__edge-text': {
        display: 'none !important',
      },
      '& .react-flow__edge-textbg': {
        display: 'none !important',
      },
    }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={onNodeClick}
        onPaneClick={onPaneClick}
        nodeTypes={nodeTypes}
        connectionMode={ConnectionMode.Loose}
        fitView
        fitViewOptions={{
          padding: 0.2,
        }}
        attributionPosition="bottom-left"
      >
        <Background 
          color="#1e293b" 
          size={1} 
          style={{ backgroundColor: '#0f172a' }}
        />
        <Controls 
          style={{
            background: 'rgba(30, 41, 59, 0.8)',
            border: '1px solid #1e293b',
          }}
        />
        <MiniMap 
          style={{
            background: 'rgba(30, 41, 59, 0.8)',
            border: '1px solid #1e293b',
          }}
          nodeColor={(node) => {
            switch (node.data.status) {
              case 'active': return '#10b981';
              case 'error': return '#ef4444';
              case 'warning': return '#f59e0b';
              default: return '#6b7280';
            }
          }}
        />
      </ReactFlow>

      {/* Network Title Overlay */}
      <Box 
        sx={{ 
          position: 'absolute', 
          top: 16, 
          left: 16, 
          zIndex: 10,
          backgroundColor: 'rgba(30, 41, 59, 0.9)',
          padding: '8px 16px',
          borderRadius: 1,
          color: 'white',
        }}
      >
        <Typography variant="h6" sx={{ color: 'white', fontWeight: 600 }}>
          Topology
        </Typography>
        <Typography variant="body2" sx={{ color: '#a0aec0' }}>
          {topologyData.network.name}
        </Typography>
      </Box>
    </Box>
  );
};

export default NetworkTopologyCanvas;

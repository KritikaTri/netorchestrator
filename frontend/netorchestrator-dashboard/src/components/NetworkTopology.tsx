import React from 'react';
import { Card, CardContent, Typography, Box } from '@mui/material';

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

interface NetworkTopologyProps {
  networks: Network[];
  nodes: Node[];
}

const NetworkTopology: React.FC<NetworkTopologyProps> = ({ networks, nodes }) => {
  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>Network Topology</Typography>
        <Box sx={{ height: 400, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Typography color="text.secondary">
            Interactive topology visualization will be implemented here
          </Typography>
        </Box>
        <Typography variant="body2" color="text.secondary">
          Networks: {networks.length} | Nodes: {nodes.length}
        </Typography>
      </CardContent>
    </Card>
  );
};

export default NetworkTopology;
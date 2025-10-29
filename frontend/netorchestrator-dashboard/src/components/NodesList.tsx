import React from 'react';
import { Card, CardContent, Typography, List, ListItem, ListItemText, Chip, Box } from '@mui/material';

interface Node {
  id: string;
  network_id: string;
  name: string;
  type: string;
  ip_address: string;
  status: string;
  config: any;
}

interface NodesListProps {
  nodes: Node[];
  loading: boolean;
}

const NodesList: React.FC<NodesListProps> = ({ nodes, loading }) => {
  if (loading) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>Network Nodes</Typography>
          <Typography>Loading...</Typography>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>Network Nodes</Typography>
        {nodes.length === 0 ? (
          <Typography color="text.secondary">No nodes found</Typography>
        ) : (
          <List>
            {nodes.slice(0, 6).map((node) => (
              <ListItem key={node.id}>
                <ListItemText
                  primary={node.name}
                  secondary={
                    <Box>
                      <Typography variant="body2" color="text.secondary">
                        Type: {node.type}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        IP: {node.ip_address}
                      </Typography>
                    </Box>
                  }
                />
                <Chip
                  label={node.status}
                  color={node.status === 'active' ? 'success' : 'default'}
                  size="small"
                />
              </ListItem>
            ))}
          </List>
        )}
      </CardContent>
    </Card>
  );
};

export default NodesList;
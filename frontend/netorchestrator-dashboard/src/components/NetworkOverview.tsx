import React from 'react';
import { Card, CardContent, Typography, Box, List, ListItem, ListItemText, Chip } from '@mui/material';

interface Network {
  id: string;
  name: string;
  description: string;
  type: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface NetworkOverviewProps {
  networks: Network[];
  loading: boolean;
  error: string | null;
}

const NetworkOverview: React.FC<NetworkOverviewProps> = ({ networks, loading, error }) => {
  if (loading) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>Network Overview</Typography>
          <Typography>Loading...</Typography>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>Network Overview</Typography>
          <Typography color="error">{error}</Typography>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>Network Overview</Typography>
        {networks.length === 0 ? (
          <Typography color="text.secondary">No networks found</Typography>
        ) : (
          <List>
            {networks.slice(0, 5).map((network) => (
              <ListItem key={network.id}>
                <ListItemText
                  primary={network.name}
                  secondary={network.description}
                />
                <Chip
                  label={network.status}
                  color={network.status === 'active' ? 'success' : 'default'}
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

export default NetworkOverview;
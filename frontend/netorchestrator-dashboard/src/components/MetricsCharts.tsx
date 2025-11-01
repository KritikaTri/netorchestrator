import React, { useEffect } from 'react';
import { Card, CardContent, Typography, Box, CircularProgress } from '@mui/material';
import { useApi } from '../contexts/ApiContext';

const MetricsCharts: React.FC = () => {
  const { metrics, fetchMetrics, loading } = useApi();

  useEffect(() => {
    fetchMetrics();
  }, [fetchMetrics]);

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Card>
      <CardContent>
        <Typography variant="h6" gutterBottom>Performance Metrics</Typography>
        <Box sx={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Typography color="text.secondary">
            Real-time metrics charts will be implemented here
          </Typography>
        </Box>
        <Typography variant="body2" color="text.secondary">
          Metrics collected: {metrics.length}
        </Typography>
      </CardContent>
    </Card>
  );
};

export default MetricsCharts;
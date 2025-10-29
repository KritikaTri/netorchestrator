import React from 'react';
import { Card, CardContent, Typography, Box } from '@mui/material';

interface Metric {
  id: string;
  source_type: string;
  source_id: string;
  metric_name: string;
  value: number;
  timestamp: string;
}

interface MetricsChartsProps {
  metrics: Metric[];
}

const MetricsCharts: React.FC<MetricsChartsProps> = ({ metrics }) => {
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
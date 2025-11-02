import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Alert,
  CircularProgress,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Grid,
  Paper,
  FormControlLabel,
  Radio,
  RadioGroup,
  CardHeader,
  Avatar,
  IconButton,
  Chip,
} from '@mui/material';
import {
  Computer as NodeIcon,
  Router as RouterIcon,
  Hub as SwitchIcon,
  Security as FirewallIcon,
  Storage as HostIcon,
  AccountTree as GatewayIcon,
  BarChart as MetricsIcon,
  Refresh as RefreshIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';

interface Network {
  id: string;
  name: string;
  description: string;
  status: string;
}

interface Node {
  id: string;
  network_id: string;
  name: string;
  type: string;
  status: string;
  ip_address: string;
  mac_address: string;
  entity_id: string;
  config: {
    cpu: number;
    memory: number;
    storage: number;
    os: string;
    image: string;
    custom_attrs?: Record<string, string>;
  };
}

interface MetricResponse {
  entity_id: string;
  entity_type: string;
  metric_name: string;
  current_value: number;
  unit: string;
  timestamp: string;
  statistics: {
    count: number;
    average: number;
    minimum: number;
    maximum: number;
    last_10_values: number[];
  };
  status: string;
}

const AVAILABLE_METRICS = [
  { value: 'cpu_percent', label: 'CPU Usage', unit: '%', color: '#FF6384' },
  { value: 'memory_percent', label: 'Memory Usage', unit: '%', color: '#36A2EB' },
  { value: 'memory_usage_mb', label: 'Memory Usage (MB)', unit: 'MB', color: '#FFCE56' },
  { value: 'network_rx_mb', label: 'Network RX', unit: 'MB', color: '#4BC0C0' },
  { value: 'pids', label: 'Process Count', unit: 'count', color: '#9966FF' },
];

const NetworkMetricsView: React.FC = () => {
  const { api } = useAuth();
  const [networks, setNetworks] = useState<Network[]>([]);
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>('');
  const [nodes, setNodes] = useState<Node[]>([]);
  const [selectedNodeId, setSelectedNodeId] = useState<string>('');
  const [selectedMetricType, setSelectedMetricType] = useState<string>('cpu_percent');
  const [metricData, setMetricData] = useState<MetricResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch networks on component mount
  useEffect(() => {
    fetchNetworks();
  }, []);

  // Fetch nodes when network is selected
  useEffect(() => {
    if (selectedNetworkId) {
      fetchNodes();
      setSelectedNodeId(''); // Reset node selection
      setMetricData(null); // Reset metric data
    }
  }, [selectedNetworkId]);

  // Fetch metrics when node and metric type are selected
  useEffect(() => {
    if (selectedNodeId && selectedMetricType) {
      fetchMetricData();
      const interval = setInterval(fetchMetricData, 10000); // Auto-refresh every 10s
      return () => clearInterval(interval);
    }
  }, [selectedNodeId, selectedMetricType]);

  const fetchNetworks = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get('/api/v1/networks');
      setNetworks(response.data.networks || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch networks');
      console.error('Error fetching networks:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchNodes = async () => {
    if (!selectedNetworkId) return;

    try {
      setLoading(true);
      setError(null);
      const response = await api.get(`/api/v1/networks/${selectedNetworkId}/nodes`);
      setNodes(response.data.nodes || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch nodes');
      console.error('Error fetching nodes:', err);
    } finally {
      setLoading(false);
    }
  };

  const fetchMetricData = async () => {
    const selectedNode = nodes.find(n => n.id === selectedNodeId);
    if (!selectedNode?.entity_id || !selectedMetricType) return;

    try {
      setLoading(true);
      setError(null);
      const response = await api.get(`/api/v1/metrics?entity_id=${selectedNode.entity_id}&metric_name=${selectedMetricType}`);
      setMetricData(response.data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch metric data');
      console.error('Error fetching metric data:', err);
    } finally {
      setLoading(false);
    }
  };

  const getNodeIcon = (nodeType: string) => {
    switch (nodeType.toLowerCase()) {
      case 'router': return <RouterIcon />;
      case 'switch': return <SwitchIcon />;
      case 'firewall': return <FirewallIcon />;
      case 'host': return <HostIcon />;
      case 'gateway': return <GatewayIcon />;
      case 'load_balancer': return <GatewayIcon />;
      default: return <NodeIcon />;
    }
  };

  const getNodeColor = (nodeType: string, status: string): string => {
    if (status !== 'active') return '#757575'; // Grey for inactive
    switch (nodeType.toLowerCase()) {
      case 'router': return '#1976d2'; // Blue
      case 'switch': return '#388e3c'; // Green
      case 'firewall': return '#f57c00'; // Orange
      case 'host': return '#7b1fa2'; // Purple
      case 'gateway': return '#c2185b'; // Pink
      case 'load_balancer': return '#00796b'; // Teal
      default: return '#616161'; // Default grey
    }
  };

  const formatValue = (value: number, unit: string): string => {
    if (unit === '%') return `${value.toFixed(1)}%`;
    if (unit === 'MB') return `${value.toFixed(1)} MB`;
    if (unit === 'count') return Math.round(value).toString();
    return value.toFixed(2);
  };

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Interactive Node Metrics
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {/* Step 1: Network Selection */}
      <Card sx={{ mb: 3 }}>
        <CardHeader 
          title="Step 1: Select Network" 
          avatar={<Avatar sx={{ bgcolor: 'primary.main' }}>1</Avatar>}
        />
        <CardContent>
          <FormControl fullWidth>
            <InputLabel>Choose a Network</InputLabel>
            <Select
              value={selectedNetworkId}
              label="Choose a Network"
              onChange={(e) => setSelectedNetworkId(e.target.value)}
              disabled={loading}
            >
              {networks.map((network) => (
                <MenuItem key={network.id} value={network.id}>
                  <Box sx={{ display: 'flex', alignItems: 'center', width: '100%' }}>
                    <Box sx={{ mr: 2 }}>
                      <Typography variant="body1" fontWeight="bold">
                        {network.name}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {network.description}
                      </Typography>
                    </Box>
                    <Box sx={{ ml: 'auto' }}>
                      <Chip 
                        label={network.status} 
                        color={network.status === 'active' ? 'success' : 'default'}
                        size="small"
                      />
                    </Box>
                  </Box>
                </MenuItem>
              ))}
            </Select>
          </FormControl>
        </CardContent>
      </Card>

      {/* Step 2: Node Selection */}
      {selectedNetworkId && (
        <Card sx={{ mb: 3 }}>
          <CardHeader 
            title="Step 2: Select Node" 
            avatar={<Avatar sx={{ bgcolor: 'primary.main' }}>2</Avatar>}
            action={
              <IconButton onClick={fetchNodes} disabled={loading}>
                <RefreshIcon />
              </IconButton>
            }
          />
          <CardContent>
            {loading && nodes.length === 0 ? (
              <CircularProgress />
            ) : nodes.length === 0 ? (
              <Alert severity="info">
                No nodes found in this network. Create some nodes first!
              </Alert>
            ) : (
              <Grid container spacing={2}>
                {nodes.map((node) => (
                  <Grid item xs={12} sm={6} md={4} key={node.id}>
                    <Paper
                      sx={{
                        p: 2,
                        cursor: 'pointer',
                        border: selectedNodeId === node.id ? 2 : 1,
                        borderColor: selectedNodeId === node.id ? 'primary.main' : 'divider',
                        backgroundColor: selectedNodeId === node.id ? 'action.selected' : 'background.paper',
                        '&:hover': {
                          backgroundColor: 'action.hover',
                          borderColor: 'primary.light',
                        },
                        transition: 'all 0.2s ease-in-out',
                      }}
                      onClick={() => setSelectedNodeId(node.id)}
                    >
                      <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                        <Avatar 
                          sx={{ 
                            bgcolor: getNodeColor(node.type, node.status),
                            mr: 2,
                            width: 32, 
                            height: 32 
                          }}
                        >
                          {getNodeIcon(node.type)}
                        </Avatar>
                        <Box>
                          <Typography variant="subtitle1" fontWeight="bold">
                            {node.name}
                          </Typography>
                          <Typography variant="caption" color="text.secondary">
                            {node.type} • {node.status}
                          </Typography>
                        </Box>
                      </Box>
                      <Typography variant="body2" color="text.secondary">
                        IP: {node.ip_address || 'N/A'}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        Entity ID: {node.entity_id ? node.entity_id.substring(0, 12) + '...' : 'None'}
                      </Typography>
                    </Paper>
                  </Grid>
                ))}
              </Grid>
            )}
          </CardContent>
        </Card>
      )}

      {/* Step 3: Metric Selection & Visualization */}
      {selectedNodeId && (
        <Card>
          <CardHeader 
            title="Step 3: Node Metrics" 
            avatar={<Avatar sx={{ bgcolor: 'primary.main' }}>3</Avatar>}
            subheader={`Metrics for: ${nodes.find(n => n.id === selectedNodeId)?.name}`}
            action={
              <IconButton onClick={fetchMetricData} disabled={loading}>
                <RefreshIcon />
              </IconButton>
            }
          />
          <CardContent>
            <Grid container spacing={3}>
              {/* Metric Type Selection */}
              <Grid item xs={12} md={4}>
                <FormControl component="fieldset">
                  <Typography variant="h6" sx={{ mb: 2 }}>
                    Select Metric Type
                  </Typography>
                  <RadioGroup
                    value={selectedMetricType}
                    onChange={(e) => setSelectedMetricType(e.target.value)}
                  >
                    {AVAILABLE_METRICS.map((metric) => (
                      <FormControlLabel
                        key={metric.value}
                        value={metric.value}
                        control={<Radio />}
                        label={
                          <Box sx={{ display: 'flex', alignItems: 'center' }}>
                            <Box
                              sx={{
                                width: 12,
                                height: 12,
                                bgcolor: metric.color,
                                borderRadius: '50%',
                                mr: 1,
                              }}
                            />
                            {metric.label}
                          </Box>
                        }
                      />
                    ))}
                  </RadioGroup>
                </FormControl>
              </Grid>

              {/* Metric Visualization */}
              <Grid item xs={12} md={8}>
                {loading && !metricData ? (
                  <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: 300 }}>
                    <CircularProgress />
                  </Box>
                ) : metricData ? (
                  <Box>
                    <Typography variant="h6" sx={{ mb: 2, textAlign: 'center' }}>
                      {AVAILABLE_METRICS.find(m => m.value === selectedMetricType)?.label} Visualization
                    </Typography>
                    
                    {/* Current Value Display */}
                    <Box sx={{ textAlign: 'center', mb: 3 }}>
                      <Typography variant="h3" color="primary" sx={{ mb: 1 }}>
                        {formatValue(metricData.current_value, metricData.unit)}
                      </Typography>
                      <Typography variant="body2" color="text.secondary">
                        Current Value
                      </Typography>
                      <Typography variant="caption" color="text.secondary" sx={{ display: 'block' }}>
                        Last updated: {new Date(metricData.timestamp).toLocaleTimeString()}
                      </Typography>
                    </Box>

                    {/* Pie Chart for Historical Data */}
                    {metricData.statistics.last_10_values.length > 0 && (
                      <Box sx={{ height: 300, mb: 3 }}>
                        <Typography variant="subtitle2" sx={{ mb: 2, textAlign: 'center' }}>
                          Recent Values Distribution (Last 10 readings)
                        </Typography>
                        <ResponsiveContainer width="100%" height="100%">
                          <PieChart>
                            <Pie
                              data={[
                                {
                                  name: 'Current Value',
                                  value: metricData.current_value,
                                  fill: AVAILABLE_METRICS.find(m => m.value === selectedMetricType)?.color || '#8884d8'
                                },
                                {
                                  name: 'Average',
                                  value: Math.max(0, metricData.statistics.average - metricData.current_value),
                                  fill: '#82ca9d'
                                },
                                {
                                  name: 'Peak (Max - Avg)',
                                  value: Math.max(0, metricData.statistics.maximum - metricData.statistics.average),
                                  fill: '#ffc658'
                                }
                              ].filter(item => item.value > 0)}
                              cx="50%"
                              cy="50%"
                              labelLine={false}
                              label={({ name, value }) => `${name}: ${formatValue(value, metricData.unit)}`}
                              outerRadius={80}
                              fill="#8884d8"
                              dataKey="value"
                            >
                              {[
                                { fill: AVAILABLE_METRICS.find(m => m.value === selectedMetricType)?.color || '#8884d8' },
                                { fill: '#82ca9d' },
                                { fill: '#ffc658' }
                              ].map((entry, index) => (
                                <Cell key={`cell-${index}`} fill={entry.fill} />
                              ))}
                            </Pie>
                            <Tooltip formatter={(value: any) => formatValue(Number(value), metricData.unit)} />
                            <Legend />
                          </PieChart>
                        </ResponsiveContainer>
                      </Box>
                    )}

                    {/* Statistics Table */}
                    <Box sx={{ display: 'flex', justifyContent: 'space-around', textAlign: 'center' }}>
                      <Box>
                        <Typography variant="h6" color="success.main">
                          {formatValue(metricData.statistics.minimum, metricData.unit)}
                        </Typography>
                        <Typography variant="caption">Minimum</Typography>
                      </Box>
                      <Box>
                        <Typography variant="h6" color="primary.main">
                          {formatValue(metricData.statistics.average, metricData.unit)}
                        </Typography>
                        <Typography variant="caption">Average</Typography>
                      </Box>
                      <Box>
                        <Typography variant="h6" color="error.main">
                          {formatValue(metricData.statistics.maximum, metricData.unit)}
                        </Typography>
                        <Typography variant="caption">Maximum</Typography>
                      </Box>
                      <Box>
                        <Typography variant="h6">
                          {metricData.statistics.count}
                        </Typography>
                        <Typography variant="caption">Data Points</Typography>
                      </Box>
                    </Box>
                  </Box>
                ) : (
                  <Alert severity="info">
                    {nodes.find(n => n.id === selectedNodeId)?.entity_id ? 
                      'Loading metrics data...' : 
                      'This node has no container metrics (no entity_id). Create the node with container provisioning enabled.'
                    }
                  </Alert>
                )}
              </Grid>
            </Grid>
          </CardContent>
        </Card>
      )}

      {/* Instructions */}
      {!selectedNetworkId && (
        <Alert severity="info" sx={{ mt: 3 }}>
          <Typography variant="subtitle2" sx={{ mb: 1 }}>
            Welcome to Interactive Node Metrics!
          </Typography>
          <Typography variant="body2">
            1. <strong>Select a network</strong> from the dropdown above<br/>
            2. <strong>Choose a node</strong> from the grid that appears<br/>
            3. <strong>Pick a metric type</strong> to visualize real-time container data<br/>
            4. <strong>View the pie chart</strong> and statistics for deep insights!
          </Typography>
        </Alert>
      )}
    </Box>
  );
};

export default NetworkMetricsView;

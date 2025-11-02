import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  LinearProgress,
  Chip,
  Alert,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  Divider,
} from '@mui/material';
import {
  CheckCircle as ActiveIcon,
  Cancel as InactiveIcon,
  Warning as WarningIcon,
  Speed as LatencyIcon,
  Link as LinkIcon,
  Security as SecurityIcon,
  Storage as ContainerIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import NetworkSelector from './NetworkSelector';

interface NetworkMetrics {
  network_id: string;
  network_name: string;
  timestamp: string;
  containers: {
    total_active: number;
    by_node_type: Record<string, number>;
  };
  links: {
    total_links: number;
    active_links: number;
    down_links: number;
    link_details: Array<{
      link_id: string;
      node_a: string;
      node_b: string;
      status: number;
    }>;
  };
  policies: {
    total_failures: number;
    failures_by_type: Record<string, number>;
  };
  provisioning: {
    avg_latency_seconds: number;
    latency_by_type: Record<string, number>;
  };
}

const NetworkMetricsView: React.FC = () => {
  const { api } = useAuth();
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>('');
  const [metrics, setMetrics] = useState<NetworkMetrics | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (selectedNetworkId) {
      fetchNetworkMetrics();
      const interval = setInterval(fetchNetworkMetrics, 30000); // Refresh every 30s
      return () => clearInterval(interval);
    }
  }, [selectedNetworkId]);

  const fetchNetworkMetrics = async () => {
    if (!selectedNetworkId) return;

    try {
      setLoading(true);
      setError(null);
      const response = await api.get(`/api/v1/metrics/network/${selectedNetworkId}`);
      setMetrics(response.data);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch network metrics');
      console.error('Error fetching network metrics:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleNetworkChange = (networkId: string) => {
    setSelectedNetworkId(networkId);
  };

  const getNodeTypeColor = (nodeType: string): 'primary' | 'secondary' | 'success' | 'warning' => {
    switch (nodeType.toLowerCase()) {
      case 'router': return 'primary';
      case 'switch': return 'secondary';
      case 'host': return 'success';
      case 'server': return 'success';
      default: return 'warning';
    }
  };

  const formatLatency = (seconds: number): string => {
    if (seconds === 0) return 'N/A';
    if (seconds < 1) return `${(seconds * 1000).toFixed(0)}ms`;
    return `${seconds.toFixed(2)}s`;
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ mb: 3, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <Typography variant="h4">Network Metrics</Typography>
        <NetworkSelector
          selectedNetworkId={selectedNetworkId}
          onNetworkChange={handleNetworkChange}
        />
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {loading && !metrics && (
        <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
          <CircularProgress />
        </Box>
      )}

      {!selectedNetworkId && !loading && (
        <Alert severity="info">
          Please select a network to view its metrics
        </Alert>
      )}

      {metrics && (
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 3 }}>
          {/* Containers Overview */}
          <Box sx={{ flex: '1 1 45%', minWidth: 300 }}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <ContainerIcon sx={{ mr: 1 }} color="primary" />
                  <Typography variant="h6">Containers</Typography>
                </Box>
                
                <Typography variant="h3" sx={{ mb: 2 }}>
                  {metrics.containers.total_active}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  Total Active Containers
                </Typography>

                {Object.keys(metrics.containers.by_node_type).length > 0 ? (
                  <>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="subtitle2" sx={{ mb: 1 }}>
                      By Node Type
                    </Typography>
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                      {Object.entries(metrics.containers.by_node_type).map(([type, count]) => (
                        <Chip
                          key={type}
                          label={`${type}: ${count}`}
                          color={getNodeTypeColor(type)}
                          size="small"
                        />
                      ))}
                    </Box>
                  </>
                ) : (
                  <Alert severity="info" sx={{ mt: 2 }}>
                    No containers provisioned yet
                  </Alert>
                )}
              </CardContent>
            </Card>
          </Box>

          {/* Links Overview */}
          <Box sx={{ flex: "1 1 45%", minWidth: 300 }}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <LinkIcon sx={{ mr: 1 }} color="primary" />
                  <Typography variant="h6">Links</Typography>
                </Box>

                <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                  <Box>
                    <Typography variant="h4">{metrics.links.total_links}</Typography>
                    <Typography variant="body2" color="text.secondary">Total</Typography>
                  </Box>
                  <Box sx={{ textAlign: 'center' }}>
                    <Typography variant="h4" color="success.main">
                      {metrics.links.active_links}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">Active</Typography>
                  </Box>
                  <Box sx={{ textAlign: 'right' }}>
                    <Typography variant="h4" color="error.main">
                      {metrics.links.down_links}
                    </Typography>
                    <Typography variant="body2" color="text.secondary">Down</Typography>
                  </Box>
                </Box>

                {metrics.links.link_details.length > 0 ? (
                  <>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="subtitle2" sx={{ mb: 1 }}>
                      Link Details
                    </Typography>
                    <List dense>
                      {metrics.links.link_details.slice(0, 5).map((link) => (
                        <ListItem key={link.link_id}>
                          <ListItemText
                            primary={`${link.node_a} ↔ ${link.node_b}`}
                            secondary={link.link_id.substring(0, 8)}
                          />
                          {link.status === 1 ? (
                            <ActiveIcon color="success" fontSize="small" />
                          ) : (
                            <InactiveIcon color="error" fontSize="small" />
                          )}
                        </ListItem>
                      ))}
                    </List>
                    {metrics.links.link_details.length > 5 && (
                      <Typography variant="caption" color="text.secondary" sx={{ mt: 1 }}>
                        +{metrics.links.link_details.length - 5} more links
                      </Typography>
                    )}
                  </>
                ) : (
                  <Alert severity="info" sx={{ mt: 2 }}>
                    No links created yet
                  </Alert>
                )}
              </CardContent>
            </Card>
          </Box>

          {/* Policy Failures */}
          <Box sx={{ flex: "1 1 45%", minWidth: 300 }}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <SecurityIcon sx={{ mr: 1 }} color={metrics.policies.total_failures > 0 ? 'error' : 'primary'} />
                  <Typography variant="h6">Policy Enforcement</Typography>
                </Box>

                <Typography variant="h3" color={metrics.policies.total_failures > 0 ? 'error' : 'success.main'} sx={{ mb: 2 }}>
                  {metrics.policies.total_failures}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  Total Failures
                </Typography>

                {metrics.policies.total_failures > 0 ? (
                  <>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="subtitle2" sx={{ mb: 1 }}>
                      Failures by Type
                    </Typography>
                    <List dense>
                      {Object.entries(metrics.policies.failures_by_type).map(([type, count]) => (
                        <ListItem key={type}>
                          <ListItemText primary={type} />
                          <Chip label={count} color="error" size="small" />
                        </ListItem>
                      ))}
                    </List>
                  </>
                ) : (
                  <Alert severity="success" sx={{ mt: 2 }}>
                    All policies enforced successfully
                  </Alert>
                )}
              </CardContent>
            </Card>
          </Box>

          {/* Provisioning Latency */}
          <Box sx={{ flex: "1 1 45%", minWidth: 300 }}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
                  <LatencyIcon sx={{ mr: 1 }} color="primary" />
                  <Typography variant="h6">Provisioning Latency</Typography>
                </Box>

                <Typography variant="h3" sx={{ mb: 2 }}>
                  {formatLatency(metrics.provisioning.avg_latency_seconds)}
                </Typography>
                <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                  Average Latency
                </Typography>

                {Object.keys(metrics.provisioning.latency_by_type).length > 0 ? (
                  <>
                    <Divider sx={{ my: 2 }} />
                    <Typography variant="subtitle2" sx={{ mb: 2 }}>
                      Latency by Type
                    </Typography>
                    {Object.entries(metrics.provisioning.latency_by_type).map(([type, latency]) => (
                      <Box key={type} sx={{ mb: 2 }}>
                        <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 0.5 }}>
                          <Typography variant="body2">{type}</Typography>
                          <Typography variant="body2" fontWeight="bold">
                            {formatLatency(latency)}
                          </Typography>
                        </Box>
                        <LinearProgress
                          variant="determinate"
                          value={Math.min((latency / 10) * 100, 100)}
                          color={latency < 2 ? 'success' : latency < 5 ? 'warning' : 'error'}
                          sx={{ height: 6, borderRadius: 3 }}
                        />
                      </Box>
                    ))}
                  </>
                ) : (
                  <Alert severity="info" sx={{ mt: 2 }}>
                    No provisioning operations recorded yet
                  </Alert>
                )}
              </CardContent>
            </Card>
          </Box>

          {/* Network Info Footer */}
          <Box sx={{ flex: "1 1 100%", width: "100%" }}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Network ID: <code>{metrics.network_id}</code>
                    </Typography>
                    <Typography variant="body2" color="text.secondary">
                      Network Name: {metrics.network_name || 'Unnamed'}
                    </Typography>
                  </Box>
                  <Box sx={{ textAlign: 'right' }}>
                    <Typography variant="caption" color="text.secondary">
                      Last updated: {new Date(parseInt(metrics.timestamp) * 1000).toLocaleString()}
                    </Typography>
                    <br />
                    <Typography variant="caption" color="text.secondary">
                      Auto-refresh: 30 seconds
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Box>
        </Box>
      )}
    </Box>
  );
};

export default NetworkMetricsView;

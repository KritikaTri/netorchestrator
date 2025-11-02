import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  CircularProgress,
  Alert,
  Paper,
  Chip,
  LinearProgress,
  Divider,
} from '@mui/material';
import {
  CheckCircle as HealthyIcon,
  Warning as WarningIcon,
  Error as ErrorIcon,
  Speed as PerformanceIcon,
  Memory as MemoryIcon,
  Storage as StorageIcon,
  NetworkCheck as NetworkIcon,
  TrendingUp as TrendingUpIcon,
  TrendingDown as TrendingDownIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';

interface SystemMetrics {
  health_score: number;
  status: string;
  timestamp: string;
  performance: {
    response_time_avg: number;
    response_time_p95: number;
    response_time_p99: number;
    throughput: number;
    error_rate: number;
    success_rate: number;
  };
  resources: {
    cpu: {
      usage_percent: number;
      cores_used: number;
      cores_total: number;
    };
    memory: {
      usage_percent: number;
      used_mb: number;
      total_mb: number;
    };
    storage: {
      usage_percent: number;
      used_gb: number;
      total_gb: number;
    };
    network: {
      bandwidth_usage_percent: number;
      packets_in: number;
      packets_out: number;
      errors: number;
    };
  };
  uptime: string;
  networks_count: number;
  nodes_count: number;
  active_connections: number;
}

const MetricsMonitoring: React.FC = () => {
  const { api } = useAuth();
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchMetrics = async () => {
    try {
      setError(null);
      // Fetch real metrics from the new aggregated metrics API
      const response = await api.get('/api/v1/metrics/summary');
      const data = response.data;

      // Transform backend response to frontend format
      const realMetrics: SystemMetrics = {
        health_score: data.health.score,
        status: data.health.status,
        timestamp: data.timestamp,
        performance: {
          response_time_avg: data.performance.response_time_avg_ms,
          response_time_p95: data.performance.response_time_p95_ms,
          response_time_p99: data.performance.response_time_p99_ms,
          throughput: data.performance.throughput_req_per_sec,
          error_rate: data.performance.error_rate_percent,
          success_rate: data.performance.success_rate_percent,
        },
        resources: {
          cpu: {
            usage_percent: data.resources.cpu.usage_percent,
            cores_used: data.resources.cpu.cores_used,
            cores_total: data.resources.cpu.cores_total,
          },
          memory: {
            usage_percent: data.resources.memory.usage_percent,
            used_mb: data.resources.memory.used_mb,
            total_mb: data.resources.memory.total_mb,
          },
          storage: {
            usage_percent: data.resources.storage.usage_percent,
            used_gb: data.resources.storage.used_gb,
            total_gb: data.resources.storage.total_gb,
          },
          network: {
            bandwidth_usage_percent: data.resources.network.bandwidth_usage_percent,
            packets_in: data.resources.network.packets_in,
            packets_out: data.resources.network.packets_out,
            errors: data.resources.network.errors,
          },
        },
        uptime: data.uptime,
        networks_count: data.infrastructure.networks.total,
        nodes_count: data.infrastructure.nodes.total,
        active_connections: data.resources.network.active_connections,
      };

      setMetrics(realMetrics);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch metrics');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const getHealthColor = (score: number) => {
    if (score >= 90) return 'success';
    if (score >= 70) return 'warning';
    return 'error';
  };

  const getHealthIcon = (score: number) => {
    if (score >= 90) return <HealthyIcon color="success" />;
    if (score >= 70) return <WarningIcon color="warning" />;
    return <ErrorIcon color="error" />;
  };

  const getUsageColor = (percent: number): 'success' | 'warning' | 'error' => {
    if (percent < 70) return 'success';
    if (percent < 85) return 'warning';
    return 'error';
  };

  const formatBytes = (mb: number) => {
    if (mb >= 1024) return `${(mb / 1024).toFixed(2)} GB`;
    return `${mb.toFixed(0)} MB`;
  };

  if (loading && !metrics) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '400px' }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error && !metrics) {
    return (
      <Alert severity="error" sx={{ m: 2 }}>
        {error}
      </Alert>
    );
  }

  if (!metrics) return null;

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        System Metrics & Monitoring
      </Typography>

      {error && (
        <Alert severity="warning" sx={{ mb: 2 }}>
          {error} - Showing cached data
        </Alert>
      )}

      {/* System Health Overview */}
      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, mb: 3 }}>
        <Card sx={{ flex: '1 1 350px', minWidth: 300 }}>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 2 }}>
              <Typography variant="h6">System Health</Typography>
              {getHealthIcon(metrics.health_score)}
            </Box>
            
            <Box sx={{ position: 'relative', display: 'inline-flex', width: '100%', justifyContent: 'center', mb: 2 }}>
              <Box sx={{ position: 'relative' }}>
                <CircularProgress
                  variant="determinate"
                  value={metrics.health_score}
                  size={120}
                  thickness={6}
                  color={getHealthColor(metrics.health_score) as any}
                />
                <Box
                  sx={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    bottom: 0,
                    right: 0,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexDirection: 'column',
                  }}
                >
                  <Typography variant="h4" component="div">
                    {metrics.health_score.toFixed(0)}
                  </Typography>
                  <Typography variant="caption" color="text.secondary">
                    Health Score
                  </Typography>
                </Box>
              </Box>
            </Box>

            <Divider sx={{ my: 2 }} />

            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
              <Chip label={`Status: ${metrics.status}`} color="success" size="small" />
              <Chip label={`Uptime: ${metrics.uptime}`} variant="outlined" size="small" />
              <Chip label={`Networks: ${metrics.networks_count}`} variant="outlined" size="small" />
              <Chip label={`Nodes: ${metrics.nodes_count}`} variant="outlined" size="small" />
            </Box>
          </CardContent>
        </Card>

        {/* Performance Metrics */}
        <Card sx={{ flex: '1 1 350px', minWidth: 300 }}>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <PerformanceIcon sx={{ mr: 1 }} color="primary" />
              <Typography variant="h6">Performance</Typography>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2" color="text.secondary">Avg Response Time</Typography>
                <Typography variant="body2" fontWeight="bold">
                  {metrics.performance.response_time_avg.toFixed(1)} ms
                </Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2" color="text.secondary">P95 Response Time</Typography>
                <Typography variant="body2">
                  {metrics.performance.response_time_p95.toFixed(1)} ms
                </Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2" color="text.secondary">P99 Response Time</Typography>
                <Typography variant="body2">
                  {metrics.performance.response_time_p99.toFixed(1)} ms
                </Typography>
              </Box>
            </Box>

            <Divider sx={{ my: 2 }} />

            <Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2" color="text.secondary">Throughput</Typography>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <TrendingUpIcon fontSize="small" color="success" sx={{ mr: 0.5 }} />
                  <Typography variant="body2" fontWeight="bold" color="success.main">
                    {metrics.performance.throughput} req/s
                  </Typography>
                </Box>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2" color="text.secondary">Success Rate</Typography>
                <Typography variant="body2" color="success.main" fontWeight="bold">
                  {metrics.performance.success_rate.toFixed(2)}%
                </Typography>
              </Box>
              <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
                <Typography variant="body2" color="text.secondary">Error Rate</Typography>
                <Typography variant="body2" color={metrics.performance.error_rate > 1 ? 'error.main' : 'text.primary'}>
                  {metrics.performance.error_rate.toFixed(2)}%
                </Typography>
              </Box>
            </Box>
          </CardContent>
        </Card>
      </Box>

      {/* Resource Utilization */}
      <Typography variant="h5" gutterBottom sx={{ mt: 3, mb: 2 }}>
        Resource Utilization
      </Typography>

      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
        {/* CPU Usage */}
        <Card sx={{ flex: '1 1 45%', minWidth: 300 }}>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <MemoryIcon sx={{ mr: 1 }} color="primary" />
              <Typography variant="h6">CPU Usage</Typography>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">
                  {metrics.resources.cpu.usage_percent.toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {metrics.resources.cpu.cores_used.toFixed(1)} / {metrics.resources.cpu.cores_total} cores
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={Math.min(metrics.resources.cpu.usage_percent, 100)}
                color={getUsageColor(metrics.resources.cpu.usage_percent)}
                sx={{ height: 8, borderRadius: 4 }}
              />
            </Box>

            <Typography variant="caption" color="text.secondary">
              {metrics.resources.cpu.usage_percent < 70 ? '✓ Normal load' :
                metrics.resources.cpu.usage_percent < 85 ? '⚠ High load' : '⚠ Critical load'}
            </Typography>
          </CardContent>
        </Card>

        {/* Memory Usage */}
        <Card sx={{ flex: '1 1 45%', minWidth: 300 }}>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <StorageIcon sx={{ mr: 1 }} color="primary" />
              <Typography variant="h6">Memory Usage</Typography>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">
                  {metrics.resources.memory.usage_percent.toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {formatBytes(metrics.resources.memory.used_mb)} / {formatBytes(metrics.resources.memory.total_mb)}
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={Math.min(metrics.resources.memory.usage_percent, 100)}
                color={getUsageColor(metrics.resources.memory.usage_percent)}
                sx={{ height: 8, borderRadius: 4 }}
              />
            </Box>

            <Typography variant="caption" color="text.secondary">
              {metrics.resources.memory.usage_percent < 70 ? '✓ Sufficient memory available' :
                metrics.resources.memory.usage_percent < 85 ? '⚠ Memory pressure' : '⚠ Critical memory usage'}
            </Typography>
          </CardContent>
        </Card>

        {/* Storage Usage */}
        <Card sx={{ flex: '1 1 45%', minWidth: 300 }}>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <StorageIcon sx={{ mr: 1 }} color="primary" />
              <Typography variant="h6">Storage Usage</Typography>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">
                  {metrics.resources.storage.usage_percent.toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {metrics.resources.storage.used_gb.toFixed(1)} / {metrics.resources.storage.total_gb} GB
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={Math.min(metrics.resources.storage.usage_percent, 100)}
                color={getUsageColor(metrics.resources.storage.usage_percent)}
                sx={{ height: 8, borderRadius: 4 }}
              />
            </Box>

            <Typography variant="caption" color="text.secondary">
              {metrics.resources.storage.usage_percent < 70 ? '✓ Plenty of storage available' :
                metrics.resources.storage.usage_percent < 85 ? '⚠ Storage filling up' : '⚠ Low storage'}
            </Typography>
          </CardContent>
        </Card>

        {/* Network Usage */}
        <Card sx={{ flex: '1 1 45%', minWidth: 300 }}>
          <CardContent>
            <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
              <NetworkIcon sx={{ mr: 1 }} color="primary" />
              <Typography variant="h6">Network Activity</Typography>
            </Box>

            <Box sx={{ mb: 2 }}>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                <Typography variant="body2">
                  Bandwidth: {metrics.resources.network.bandwidth_usage_percent.toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  {metrics.active_connections} active
                </Typography>
              </Box>
              <LinearProgress
                variant="determinate"
                value={Math.min(metrics.resources.network.bandwidth_usage_percent, 100)}
                color={getUsageColor(metrics.resources.network.bandwidth_usage_percent)}
                sx={{ height: 8, borderRadius: 4 }}
              />
            </Box>

            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
              <Typography variant="caption" color="text.secondary">Packets In</Typography>
              <Typography variant="caption">{metrics.resources.network.packets_in.toLocaleString()}/s</Typography>
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
              <Typography variant="caption" color="text.secondary">Packets Out</Typography>
              <Typography variant="caption">{metrics.resources.network.packets_out.toLocaleString()}/s</Typography>
            </Box>
            <Box sx={{ display: 'flex', justifyContent: 'space-between' }}>
              <Typography variant="caption" color="text.secondary">Errors</Typography>
              <Typography variant="caption" color={metrics.resources.network.errors > 10 ? 'error' : 'inherit'}>
                {metrics.resources.network.errors}
              </Typography>
            </Box>
          </CardContent>
        </Card>
      </Box>

      {/* Footer */}
      <Box sx={{ mt: 3, textAlign: 'center' }}>
        <Typography variant="caption" color="text.secondary">
          Last updated: {new Date(metrics.timestamp).toLocaleString()} • Auto-refresh: 30s
        </Typography>
      </Box>
    </Box>
  );
};

export default MetricsMonitoring;

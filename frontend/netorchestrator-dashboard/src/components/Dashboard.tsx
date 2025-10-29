import React, { useState, useEffect } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Container,
  Grid,
  Card,
  CardContent,
  Box,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  Badge,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Hub as NetworkIcon,
  Storage as NodeIcon,
  Security as PolicyIcon,
  Analytics as MetricsIcon,
  Notifications as NotificationsIcon,
  Settings as SettingsIcon,
  Person as PersonIcon,
} from '@mui/icons-material';
import { useApi } from '../contexts/ApiContext';
import { useWebSocket } from '../contexts/WebSocketContext';
import NetworkOverview from './NetworkOverview';
import NetworkTopology from './NetworkTopology';
import MetricsCharts from './MetricsCharts';
import NodesList from './NodesList';

const Dashboard: React.FC = () => {
  const { networks, nodes, metrics, loading, error, fetchNetworks, fetchNodes, fetchMetrics } = useApi();
  const { connected, lastMessage } = useWebSocket();
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [notifications, setNotifications] = useState<any[]>([]);

  useEffect(() => {
    // Initial data fetch
    fetchNetworks();
    fetchNodes();
    fetchMetrics();
  }, [fetchNetworks, fetchNodes, fetchMetrics]);

  useEffect(() => {
    // Handle real-time updates
    if (lastMessage) {
      switch (lastMessage.type) {
        case 'network_update':
          fetchNetworks();
          break;
        case 'metric_update':
          fetchMetrics();
          break;
        case 'alert':
          setNotifications(prev => [...prev, lastMessage.data]);
          break;
        case 'node_status':
          fetchNodes();
          break;
      }
    }
  }, [lastMessage, fetchNetworks, fetchNodes, fetchMetrics]);

  const handleNotificationClick = (event: React.MouseEvent<HTMLElement>) => {
    setAnchorEl(event.currentTarget);
  };

  const handleNotificationClose = () => {
    setAnchorEl(null);
  };

  const activeNetworks = networks.filter(n => n.status === 'active').length;
  const activeNodes = nodes.filter(n => n.status === 'active').length;
  const totalNetworks = networks.length;
  const totalNodes = nodes.length;

  return (
    <Box sx={{ flexGrow: 1 }}>
      {/* Header */}
      <AppBar position="static" sx={{ backgroundColor: 'rgba(26, 29, 54, 0.95)', backdropFilter: 'blur(10px)' }}>
        <Toolbar>
          <NetworkIcon sx={{ mr: 2 }} />
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            NetOrchestrator - NaaS Platform
          </Typography>
          
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Chip
              label={connected ? 'Connected' : 'Disconnected'}
              color={connected ? 'success' : 'error'}
              size="small"
              variant="outlined"
            />
            
            <IconButton color="inherit">
              <Badge badgeContent={notifications.length} color="error">
                <NotificationsIcon onClick={handleNotificationClick} />
              </Badge>
            </IconButton>
            
            <IconButton color="inherit">
              <SettingsIcon />
            </IconButton>
            
            <IconButton color="inherit">
              <PersonIcon />
            </IconButton>
          </Box>
        </Toolbar>
      </AppBar>

      {/* Notifications Menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleNotificationClose}
      >
        {notifications.length === 0 ? (
          <MenuItem onClick={handleNotificationClose}>No new notifications</MenuItem>
        ) : (
          notifications.map((notification, index) => (
            <MenuItem key={index} onClick={handleNotificationClose}>
              {notification.message || JSON.stringify(notification)}
            </MenuItem>
          ))
        )}
      </Menu>

      <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
        {/* Stats Cards */}
        <Grid container spacing={3} sx={{ mb: 4 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <DashboardIcon sx={{ mr: 2, color: 'primary.main' }} />
                  <Box>
                    <Typography variant="h4" component="div">
                      {totalNetworks}
                    </Typography>
                    <Typography color="text.secondary">
                      Total Networks
                    </Typography>
                    <Typography variant="body2" color="success.main">
                      {activeNetworks} Active
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <NodeIcon sx={{ mr: 2, color: 'secondary.main' }} />
                  <Box>
                    <Typography variant="h4" component="div">
                      {totalNodes}
                    </Typography>
                    <Typography color="text.secondary">
                      Total Nodes
                    </Typography>
                    <Typography variant="body2" color="success.main">
                      {activeNodes} Active
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <PolicyIcon sx={{ mr: 2, color: 'info.main' }} />
                  <Box>
                    <Typography variant="h4" component="div">
                      12
                    </Typography>
                    <Typography color="text.secondary">
                      Active Policies
                    </Typography>
                    <Typography variant="body2" color="warning.main">
                      2 Warnings
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
          
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <MetricsIcon sx={{ mr: 2, color: 'success.main' }} />
                  <Box>
                    <Typography variant="h4" component="div">
                      98.5%
                    </Typography>
                    <Typography color="text.secondary">
                      Uptime
                    </Typography>
                    <Typography variant="body2" color="success.main">
                      All systems operational
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Main Content */}
        <Grid container spacing={3}>
          {/* Network Overview */}
          <Grid item xs={12} lg={8}>
            <NetworkOverview networks={networks} loading={loading} error={error} />
          </Grid>
          
          {/* Nodes List */}
          <Grid item xs={12} lg={4}>
            <NodesList nodes={nodes} loading={loading} />
          </Grid>
          
          {/* Network Topology */}
          <Grid item xs={12} lg={8}>
            <NetworkTopology networks={networks} nodes={nodes} />
          </Grid>
          
          {/* Metrics Charts */}
          <Grid item xs={12} lg={4}>
            <MetricsCharts metrics={metrics} />
          </Grid>
        </Grid>
      </Container>
    </Box>
  );
};

export default Dashboard;
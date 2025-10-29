import React, { useState, useEffect } from 'react';
import {
  AppBar,
  Toolbar,
  Typography,
  Container,
  Card,
  CardContent,
  Box,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  Button,
  Alert,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Hub as NetworkIcon,
  Storage as NodeIcon,
  Notifications as NotificationsIcon,
  Person as PersonIcon,
  Logout as LogoutIcon,
  Add as AddIcon,
  PlayArrow as StartIcon,
  Stop as StopIcon,
  CheckCircle as ActiveIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { useApi } from '../contexts/ApiContext';
import { useWebSocket } from '../contexts/WebSocketContext';

const Dashboard: React.FC = () => {
  const { user, logout } = useAuth();
  const { 
    networks, 
    nodes, 
    alerts, 
    loading, 
    error, 
    fetchNetworks, 
    fetchNodes, 
    fetchAlerts, 
    createNetwork, 
    startNetwork, 
    stopNetwork 
  } = useApi();
  const { connected, lastMessage } = useWebSocket();
  
  const [userMenuAnchor, setUserMenuAnchor] = useState<null | HTMLElement>(null);
  const [createNetworkOpen, setCreateNetworkOpen] = useState(false);
  const [networkForm, setNetworkForm] = useState({
    name: '',
    description: '',
  });

  useEffect(() => {
    // Initial data fetch
    console.log('Fetching initial data...');
    fetchNetworks();
    fetchNodes();
    fetchAlerts();
  }, [fetchNetworks, fetchNodes, fetchAlerts]);

  useEffect(() => {
    // Handle real-time updates
    if (lastMessage) {
      console.log('Received WebSocket message:', lastMessage);
      // Refresh data based on message type
      fetchNetworks();
      fetchNodes();
      fetchAlerts();
    }
  }, [lastMessage, fetchNetworks, fetchNodes, fetchAlerts]);

  const handleUserMenuClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setUserMenuAnchor(event.currentTarget);
  };

  const handleUserMenuClose = () => {
    setUserMenuAnchor(null);
  };

  const handleLogout = () => {
    handleUserMenuClose();
    logout();
  };

  const handleCreateNetwork = async () => {
    try {
      await createNetwork(networkForm);
      setCreateNetworkOpen(false);
      setNetworkForm({ name: '', description: '' });
    } catch (error) {
      console.error('Failed to create network:', error);
    }
  };

  const handleNetworkAction = async (networkId: string, action: 'start' | 'stop') => {
    try {
      if (action === 'start') {
        await startNetwork(networkId);
      } else {
        await stopNetwork(networkId);
      }
    } catch (error) {
      console.error(`Failed to ${action} network:`, error);
    }
  };

  const activeNetworks = networks.filter(n => n.status === 'active').length;
  const activeNodes = nodes.filter(n => n.status === 'active').length;
  const totalNetworks = networks.length;
  const totalNodes = nodes.length;
  const activeAlerts = alerts.filter(a => a.status !== 'resolved').length;

  return (
    <Box sx={{ flexGrow: 1 }}>
      {/* App Bar */}
      <AppBar position="static">
        <Toolbar>
          <DashboardIcon sx={{ mr: 2 }} />
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            NetOrchestrator Dashboard
          </Typography>
          
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <Chip 
              label={connected ? 'Connected' : 'Disconnected'} 
              color={connected ? 'success' : 'error'} 
              size="small" 
            />
            
            <IconButton color="inherit">
              <NotificationsIcon />
            </IconButton>
            
            <IconButton color="inherit" onClick={handleUserMenuClick}>
              <PersonIcon />
            </IconButton>
          </Box>
        </Toolbar>
      </AppBar>

      {/* User Menu */}
      <Menu
        anchorEl={userMenuAnchor}
        open={Boolean(userMenuAnchor)}
        onClose={handleUserMenuClose}
      >
        <MenuItem onClick={handleUserMenuClose}>
          <PersonIcon sx={{ mr: 1 }} />
          {user?.username || 'User'}
        </MenuItem>
        <MenuItem onClick={handleLogout}>
          <LogoutIcon sx={{ mr: 1 }} />
          Logout
        </MenuItem>
      </Menu>

      {/* Main Content */}
      <Container maxWidth="xl" sx={{ mt: 4, mb: 4 }}>
        {/* Error Display */}
        {error && (
          <Alert severity="error" sx={{ mb: 2 }}>
            {error}
          </Alert>
        )}

        {/* Stats Cards */}
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, mb: 3 }}>
          <Card sx={{ minWidth: 250, flex: '1 1 250px' }}>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
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
                <NetworkIcon sx={{ fontSize: 40, color: 'primary.main' }} />
              </Box>
            </CardContent>
          </Card>
          
          <Card sx={{ minWidth: 250, flex: '1 1 250px' }}>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
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
                <NodeIcon sx={{ fontSize: 40, color: 'primary.main' }} />
              </Box>
            </CardContent>
          </Card>
          
          <Card sx={{ minWidth: 250, flex: '1 1 250px' }}>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="h4" component="div">
                    {activeAlerts}
                  </Typography>
                  <Typography color="text.secondary">
                    Active Alerts
                  </Typography>
                </Box>
                <WarningIcon sx={{ fontSize: 40, color: 'warning.main' }} />
              </Box>
            </CardContent>
          </Card>
          
          <Card sx={{ minWidth: 250, flex: '1 1 250px' }}>
            <CardContent>
              <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <Box>
                  <Typography variant="h4" component="div">
                    {connected ? 'OK' : 'DOWN'}
                  </Typography>
                  <Typography color="text.secondary">
                    Connection Status
                  </Typography>
                </Box>
                {connected ? (
                  <ActiveIcon sx={{ fontSize: 40, color: 'success.main' }} />
                ) : (
                  <ErrorIcon sx={{ fontSize: 40, color: 'error.main' }} />
                )}
              </Box>
            </CardContent>
          </Card>
        </Box>

        {/* Networks Section */}
        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
          <Card sx={{ flex: '1 1 600px', minWidth: 600 }}>
            <CardContent>
              <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
                <Typography variant="h6">Networks</Typography>
                <Button
                  variant="contained"
                  startIcon={<AddIcon />}
                  onClick={() => setCreateNetworkOpen(true)}
                >
                  Create Network
                </Button>
              </Box>
              
              {loading ? (
                <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}>
                  <CircularProgress />
                </Box>
              ) : (
                <List>
                  {networks.length === 0 ? (
                    <ListItem>
                      <ListItemText primary="No networks found" />
                    </ListItem>
                  ) : (
                    networks.map((network) => (
                      <ListItem key={network.id}>
                        <ListItemText
                          primary={network.name}
                          secondary={network.description}
                        />
                        <Chip
                          label={network.status}
                          color={network.status === 'active' ? 'success' : 'default'}
                          size="small"
                          sx={{ mr: 1 }}
                        />
                        <IconButton
                          onClick={() => handleNetworkAction(network.id, network.status === 'active' ? 'stop' : 'start')}
                          size="small"
                        >
                          {network.status === 'active' ? <StopIcon /> : <StartIcon />}
                        </IconButton>
                      </ListItem>
                    ))
                  )}
                </List>
              )}
            </CardContent>
          </Card>
          
          <Card sx={{ flex: '1 1 300px', minWidth: 300 }}>
            <CardContent>
              <Typography variant="h6" sx={{ mb: 2 }}>Recent Alerts</Typography>
              <List>
                {alerts.slice(0, 5).map((alert) => (
                  <ListItem key={alert.id}>
                    <ListItemText
                      primary={alert.title}
                      secondary={alert.description || 'No description'}
                    />
                    <Chip
                      label={alert.severity}
                      color={alert.severity === 'high' ? 'error' : alert.severity === 'medium' ? 'warning' : 'info'}
                      size="small"
                    />
                  </ListItem>
                ))}
                {alerts.length === 0 && (
                  <ListItem>
                    <ListItemText primary="No alerts" />
                  </ListItem>
                )}
              </List>
            </CardContent>
          </Card>
        </Box>
      </Container>

      {/* Create Network Dialog */}
      <Dialog open={createNetworkOpen} onClose={() => setCreateNetworkOpen(false)}>
        <DialogTitle>Create New Network</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Network Name"
            fullWidth
            variant="outlined"
            value={networkForm.name}
            onChange={(e) => setNetworkForm({ ...networkForm, name: e.target.value })}
            sx={{ mb: 2 }}
          />
          <TextField
            margin="dense"
            label="Description"
            fullWidth
            variant="outlined"
            multiline
            rows={3}
            value={networkForm.description}
            onChange={(e) => setNetworkForm({ ...networkForm, description: e.target.value })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateNetworkOpen(false)}>Cancel</Button>
          <Button onClick={handleCreateNetwork} variant="contained">Create</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Dashboard;
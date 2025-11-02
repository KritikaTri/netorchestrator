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
  Drawer,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Divider,
  Tabs,
  Tab,
  CircularProgress,
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Hub as NetworkIcon,
  Storage as NodeIcon,
  Link as LinkIcon,
  Security as SecurityIcon,
  Timeline as TimelineIcon,
  Person as PersonIcon,
  Logout as LogoutIcon,
  Menu as MenuIcon,
  Notifications as NotificationsIcon,
  CheckCircle as ActiveIcon,
  Error as ErrorIcon,
  Warning as WarningIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { useApi } from '../contexts/ApiContext';
import NetworkManagement from './NetworkManagement';
import NodeManagement from './NodeManagement';
import LinkManagement from './LinkManagement';
import PolicyManagement from './PolicyManagement';
import MetricsMonitoring from './MetricsMonitoring';
import NetworkMetricsView from './NetworkMetricsView';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;
  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`tabpanel-${index}`}
      aria-labelledby={`tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ py: 3 }}>{children}</Box>}
    </div>
  );
}

const MainDashboard: React.FC = () => {
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
  } = useApi();
  
  const [userMenuAnchor, setUserMenuAnchor] = useState<null | HTMLElement>(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [activeView, setActiveView] = useState('overview');

  useEffect(() => {
    // Initial data fetch
    fetchNetworks();
    fetchNodes();
    fetchAlerts();
  }, []);

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

  const activeNetworks = networks.filter(n => n.status === 'active').length;
  const activeNodes = nodes.filter(n => n.status === 'active').length;
  const totalNetworks = networks.length;
  const totalNodes = nodes.length;
  const activeAlerts = alerts.filter(a => a.status !== 'resolved').length;

  const menuItems = [
    { id: 'overview', label: 'Overview', icon: <DashboardIcon /> },
    { id: 'networks', label: 'Networks', icon: <NetworkIcon /> },
    { id: 'nodes', label: 'Nodes', icon: <NodeIcon /> },
    { id: 'links', label: 'Links', icon: <LinkIcon /> },
    { id: 'policies', label: 'Policies', icon: <SecurityIcon /> },
    { id: 'monitoring', label: 'System Monitoring', icon: <TimelineIcon /> },
    { id: 'network-metrics', label: 'Network Metrics', icon: <NetworkIcon /> },
  ];

  const renderContent = () => {
    switch (activeView) {
      case 'overview':
        return (
          <Box>
            <Typography variant="h4" gutterBottom>
              Network Orchestration Dashboard
            </Typography>

            {/* Stats Cards */}
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2, mb: 3, mt: 3 }}>
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
            </Box>

            {/* Recent Networks */}
            <Card sx={{ mb: 2 }}>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Recent Networks
                </Typography>
                {loading ? (
                  <Box sx={{ display: 'flex', justifyContent: 'center', p: 2 }}>
                    <CircularProgress />
                  </Box>
                ) : networks.length === 0 ? (
                  <Typography color="textSecondary">
                    No networks found. Create your first network to get started.
                  </Typography>
                ) : (
                  <List>
                    {networks.slice(0, 5).map((network) => (
                      <ListItemButton key={network.id}>
                        <ListItemIcon>
                          <NetworkIcon />
                        </ListItemIcon>
                        <ListItemText
                          primary={network.name}
                          secondary={network.description}
                        />
                        <Chip
                          label={network.status}
                          color={network.status === 'active' ? 'success' : 'default'}
                          size="small"
                        />
                      </ListItemButton>
                    ))}
                  </List>
                )}
                <Box sx={{ mt: 2 }}>
                  <Button
                    variant="contained"
                    fullWidth
                    onClick={() => setActiveView('networks')}
                  >
                    Manage Networks
                  </Button>
                </Box>
              </CardContent>
            </Card>

            {/* Recent Alerts */}
            {activeAlerts > 0 && (
              <Card>
                <CardContent>
                  <Typography variant="h6" gutterBottom>
                    Recent Alerts
                  </Typography>
                  <List>
                    {alerts.slice(0, 5).map((alert) => (
                      <ListItemButton key={alert.id}>
                        <ListItemIcon>
                          {alert.severity === 'high' ? <ErrorIcon color="error" /> : <WarningIcon color="warning" />}
                        </ListItemIcon>
                        <ListItemText
                          primary={alert.title}
                          secondary={alert.description}
                        />
                        <Chip
                          label={alert.severity}
                          color={alert.severity === 'high' ? 'error' : alert.severity === 'medium' ? 'warning' : 'info'}
                          size="small"
                        />
                      </ListItemButton>
                    ))}
                  </List>
                </CardContent>
              </Card>
            )}
          </Box>
        );
      
      case 'networks':
        return <NetworkManagement />;
      
      case 'nodes':
        return <NodeManagement />;
      
      case 'links':
        return <LinkManagement />;
      
      case 'policies':
        return <PolicyManagement />;
      
      case 'monitoring':
        return <MetricsMonitoring />;
      
      case 'network-metrics':
        return <NetworkMetricsView />;
      
      default:
        return null;
    }
  };

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      {/* App Bar */}
      <AppBar position="static">
        <Toolbar>
          <IconButton
            color="inherit"
            edge="start"
            onClick={() => setDrawerOpen(!drawerOpen)}
            sx={{ mr: 2 }}
          >
            <MenuIcon />
          </IconButton>
          
          <DashboardIcon sx={{ mr: 2 }} />
          
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            NetOrchestrator
          </Typography>
          
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <IconButton color="inherit">
              <NotificationsIcon />
              {activeAlerts > 0 && (
                <Box
                  sx={{
                    position: 'absolute',
                    top: 8,
                    right: 8,
                    width: 8,
                    height: 8,
                    bgcolor: 'error.main',
                    borderRadius: '50%',
                  }}
                />
              )}
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
        <MenuItem disabled>
          <PersonIcon sx={{ mr: 1 }} />
          {user?.username || 'User'}
        </MenuItem>
        <Divider />
        <MenuItem onClick={handleLogout}>
          <LogoutIcon sx={{ mr: 1 }} />
          Logout
        </MenuItem>
      </Menu>

      {/* Navigation Drawer */}
      <Drawer
        anchor="left"
        open={drawerOpen}
        onClose={() => setDrawerOpen(false)}
      >
        <Box sx={{ width: 250 }} role="presentation">
          <Box sx={{ p: 2 }}>
            <Typography variant="h6">Navigation</Typography>
          </Box>
          <Divider />
          <List>
            {menuItems.map((item) => (
              <ListItemButton
                key={item.id}
                selected={activeView === item.id}
                onClick={() => {
                  setActiveView(item.id);
                  setDrawerOpen(false);
                }}
              >
                <ListItemIcon>{item.icon}</ListItemIcon>
                <ListItemText primary={item.label} />
              </ListItemButton>
            ))}
          </List>
        </Box>
      </Drawer>

      {/* Main Content */}
      <Container maxWidth="xl" sx={{ mt: 4, mb: 4, flexGrow: 1 }}>
        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => {}}>
            {error}
          </Alert>
        )}
        
        {renderContent()}
      </Container>

      {/* Footer */}
      <Box
        component="footer"
        sx={{
          py: 2,
          px: 2,
          mt: 'auto',
          backgroundColor: (theme) =>
            theme.palette.mode === 'light'
              ? theme.palette.grey[200]
              : theme.palette.grey[800],
        }}
      >
        <Container maxWidth="xl">
          <Typography variant="body2" color="text.secondary" align="center">
            NetOrchestrator v2.0 - Network Orchestration Platform
          </Typography>
        </Container>
      </Box>
    </Box>
  );
};

export default MainDashboard;

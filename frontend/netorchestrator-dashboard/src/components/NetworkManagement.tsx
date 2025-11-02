import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Typography,
  Chip,
  IconButton,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Tabs,
  Tab,
  Stepper,
  Step,
  StepLabel,
  Alert,
  CircularProgress,
  Tooltip,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  PlayArrow as PlayIcon,
  Stop as StopIcon,
  Visibility as ViewIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';

interface NetworkConfig {
  subnet: string;
  gateway: string;
  dns_servers: string[];
  vlan_id?: number;
  mtu?: number;
  driver?: string;
}

interface Network {
  id: string;
  name: string;
  description: string;
  status: string;
  user_id: string;
  config: NetworkConfig;
  created_at: string;
  updated_at: string;
}

const NetworkManagement: React.FC = () => {
  const { api } = useAuth();
  const [networks, setNetworks] = useState<Network[]>([]);
  const [loading, setLoading] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [selectedNetwork, setSelectedNetwork] = useState<Network | null>(null);
  const [activeStep, setActiveStep] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    subnet: '10.0.0.0/24',
    gateway: '10.0.0.1',
    dns_servers: '8.8.8.8,8.8.4.4',
    vlan_id: '',
    mtu: '1500',
    driver: 'bridge',
  });

  const steps = ['Basic Info', 'Network Configuration', 'Review'];

  useEffect(() => {
    fetchNetworks();
  }, []);

  const fetchNetworks = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.get('/api/v1/networks');
      setNetworks(response.data.networks || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch networks');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (network?: Network) => {
    if (network) {
      setEditMode(true);
      setSelectedNetwork(network);
      setFormData({
        name: network.name,
        description: network.description,
        subnet: network.config.subnet || '10.0.0.0/24',
        gateway: network.config.gateway || '10.0.0.1',
        dns_servers: network.config.dns_servers?.join(',') || '8.8.8.8,8.8.4.4',
        vlan_id: network.config.vlan_id?.toString() || '',
        mtu: network.config.mtu?.toString() || '1500',
        driver: network.config.driver || 'bridge',
      });
    } else {
      setEditMode(false);
      setSelectedNetwork(null);
      setFormData({
        name: '',
        description: '',
        subnet: '10.0.0.0/24',
        gateway: '10.0.0.1',
        dns_servers: '8.8.8.8,8.8.4.4',
        vlan_id: '',
        mtu: '1500',
        driver: 'bridge',
      });
    }
    setActiveStep(0);
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setError(null);
    setSuccess(null);
  };

  const handleNext = () => {
    setActiveStep((prev) => prev + 1);
  };

  const handleBack = () => {
    setActiveStep((prev) => prev - 1);
  };

  const handleSubmit = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const payload = {
        name: formData.name,
        description: formData.description,
        config: {
          subnet: formData.subnet,
          gateway: formData.gateway,
          dns_servers: formData.dns_servers.split(',').map(s => s.trim()),
          ...(formData.vlan_id && { vlan_id: parseInt(formData.vlan_id) }),
          mtu: parseInt(formData.mtu),
          driver: formData.driver,
        },
      };

      if (editMode && selectedNetwork) {
        await api.put(`/api/v1/networks/${selectedNetwork.id}`, payload);
        setSuccess('Network updated successfully!');
      } else {
        await api.post('/api/v1/networks', payload);
        setSuccess('Network created successfully!');
      }

      await fetchNetworks();
      setTimeout(() => {
        handleCloseDialog();
      }, 1500);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save network');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (networkId: string) => {
    if (!window.confirm('Are you sure you want to delete this network?')) {
      return;
    }

    setLoading(true);
    try {
      await api.delete(`/api/v1/networks/${networkId}`);
      setSuccess('Network deleted successfully!');
      await fetchNetworks();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete network');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
        return 'success';
      case 'inactive':
        return 'default';
      case 'provisioning':
        return 'warning';
      case 'error':
        return 'error';
      default:
        return 'default';
    }
  };

  const renderStepContent = (step: number) => {
    switch (step) {
      case 0:
        return (
          <Box sx={{ mt: 2 }}>
            <TextField
              fullWidth
              label="Network Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              margin="normal"
              required
              helperText="Enter a unique name for your network"
            />
            <TextField
              fullWidth
              label="Description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              margin="normal"
              multiline
              rows={3}
              helperText="Describe the purpose of this network"
            />
          </Box>
        );
      case 1:
        return (
          <Box sx={{ mt: 2 }}>
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
              <Box sx={{ flex: '1 1 45%', minWidth: '200px' }}>
                <TextField
                  fullWidth
                  label="Subnet (CIDR)"
                  value={formData.subnet}
                  onChange={(e) => setFormData({ ...formData, subnet: e.target.value })}
                  margin="normal"
                  required
                  helperText="e.g., 10.0.0.0/24"
                />
              </Box>
              <Box sx={{ flex: '1 1 45%', minWidth: '200px' }}>
                <TextField
                  fullWidth
                  label="Gateway"
                  value={formData.gateway}
                  onChange={(e) => setFormData({ ...formData, gateway: e.target.value })}
                  margin="normal"
                  required
                  helperText="e.g., 10.0.0.1"
                />
              </Box>
              <Box sx={{ flex: '1 1 100%' }}>
                <TextField
                  fullWidth
                  label="DNS Servers"
                  value={formData.dns_servers}
                  onChange={(e) => setFormData({ ...formData, dns_servers: e.target.value })}
                  margin="normal"
                  helperText="Comma-separated list, e.g., 8.8.8.8,8.8.4.4"
                />
              </Box>
              <Box sx={{ flex: '1 1 30%', minWidth: '150px' }}>
                <TextField
                  fullWidth
                  label="VLAN ID (Optional)"
                  type="number"
                  value={formData.vlan_id}
                  onChange={(e) => setFormData({ ...formData, vlan_id: e.target.value })}
                  margin="normal"
                  helperText="e.g., 100"
                />
              </Box>
              <Box sx={{ flex: '1 1 30%', minWidth: '150px' }}>
                <TextField
                  fullWidth
                  label="MTU"
                  type="number"
                  value={formData.mtu}
                  onChange={(e) => setFormData({ ...formData, mtu: e.target.value })}
                  margin="normal"
                  helperText="Default: 1500"
                />
              </Box>
              <Box sx={{ flex: '1 1 30%', minWidth: '150px' }}>
                <FormControl fullWidth margin="normal">
                  <InputLabel>Network Driver</InputLabel>
                  <Select
                    value={formData.driver}
                    label="Network Driver"
                    onChange={(e) => setFormData({ ...formData, driver: e.target.value })}
                  >
                    <MenuItem value="bridge">Bridge</MenuItem>
                    <MenuItem value="overlay">Overlay</MenuItem>
                    <MenuItem value="macvlan">Macvlan</MenuItem>
                    <MenuItem value="ipvlan">IPvlan</MenuItem>
                  </Select>
                </FormControl>
              </Box>
            </Box>
          </Box>
        );
      case 2:
        return (
          <Box sx={{ mt: 2 }}>
            <Typography variant="h6" gutterBottom>
              Review Network Configuration
            </Typography>
            <TableContainer component={Paper} sx={{ mt: 2 }}>
              <Table size="small">
                <TableBody>
                  <TableRow>
                    <TableCell><strong>Name</strong></TableCell>
                    <TableCell>{formData.name}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Description</strong></TableCell>
                    <TableCell>{formData.description}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Subnet</strong></TableCell>
                    <TableCell>{formData.subnet}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Gateway</strong></TableCell>
                    <TableCell>{formData.gateway}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>DNS Servers</strong></TableCell>
                    <TableCell>{formData.dns_servers}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>VLAN ID</strong></TableCell>
                    <TableCell>{formData.vlan_id || 'None'}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>MTU</strong></TableCell>
                    <TableCell>{formData.mtu}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Driver</strong></TableCell>
                    <TableCell>{formData.driver}</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </TableContainer>
          </Box>
        );
      default:
        return null;
    }
  };

  return (
    <Box sx={{ p: 3 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">Network Management</Typography>
        <Box>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={fetchNetworks}
            sx={{ mr: 1 }}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => handleOpenDialog()}
          >
            Create Network
          </Button>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" onClose={() => setError(null)} sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" onClose={() => setSuccess(null)} sx={{ mb: 2 }}>
          {success}
        </Alert>
      )}

      {loading && networks.length === 0 ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Description</TableCell>
                <TableCell>Subnet</TableCell>
                <TableCell>Gateway</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Created</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {networks.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} align="center">
                    <Typography color="textSecondary" sx={{ py: 3 }}>
                      No networks found. Create your first network to get started.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                networks.map((network) => (
                  <TableRow key={network.id} hover>
                    <TableCell>{network.name}</TableCell>
                    <TableCell>{network.description}</TableCell>
                    <TableCell>{network.config?.subnet || 'N/A'}</TableCell>
                    <TableCell>{network.config?.gateway || 'N/A'}</TableCell>
                    <TableCell>
                      <Chip
                        label={network.status}
                        color={getStatusColor(network.status) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      {new Date(network.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Tooltip title="Edit">
                        <IconButton
                          size="small"
                          onClick={() => handleOpenDialog(network)}
                        >
                          <EditIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleDelete(network.id)}
                        >
                          <DeleteIcon />
                        </IconButton>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      {/* Create/Edit Dialog */}
      <Dialog
        open={openDialog}
        onClose={handleCloseDialog}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          {editMode ? 'Edit Network' : 'Create New Network'}
        </DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          {success && (
            <Alert severity="success" sx={{ mb: 2 }}>
              {success}
            </Alert>
          )}

          <Stepper activeStep={activeStep} sx={{ mt: 2, mb: 3 }}>
            {steps.map((label) => (
              <Step key={label}>
                <StepLabel>{label}</StepLabel>
              </Step>
            ))}
          </Stepper>

          {renderStepContent(activeStep)}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          {activeStep > 0 && (
            <Button onClick={handleBack}>Back</Button>
          )}
          {activeStep < steps.length - 1 ? (
            <Button
              variant="contained"
              onClick={handleNext}
              disabled={!formData.name || !formData.subnet || !formData.gateway}
            >
              Next
            </Button>
          ) : (
            <Button
              variant="contained"
              onClick={handleSubmit}
              disabled={loading}
            >
              {loading ? <CircularProgress size={24} /> : editMode ? 'Update' : 'Create'}
            </Button>
          )}
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default NetworkManagement;

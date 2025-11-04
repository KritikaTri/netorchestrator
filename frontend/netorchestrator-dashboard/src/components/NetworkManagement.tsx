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
  Psychology as AIIcon,
  Settings as AdvancedIcon,
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
  const [creationMode, setCreationMode] = useState<'advanced' | 'nlp'>('advanced');
  const [nlpText, setNlpText] = useState('');
  const [nlpLoading, setNlpLoading] = useState(false);
  const [nlpPreview, setNlpPreview] = useState<any>(null);

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
    setCreationMode('advanced');
    setNlpText('');
    setNlpPreview(null);
    setError(null);
    setSuccess(null);
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setCreationMode('advanced');
    setNlpText('');
    setNlpPreview(null);
    setError(null);
    setSuccess(null);
  };

  const handleNlpPreview = async () => {
    if (!nlpText.trim()) {
      setError('Please enter a network description');
      return;
    }

    setNlpLoading(true);
    setError(null);
    try {
      const response = await api.post('/api/v1/ai/provision', {
        text: nlpText
      });
      
      setNlpPreview(response.data);
      setError(null);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to parse network description');
      setNlpPreview(null);
    } finally {
      setNlpLoading(false);
    }
  };

  const handleNlpCreate = async () => {
    if (!nlpPreview) {
      setError('Please preview the network first');
      return;
    }

    setLoading(true);
    try {
      // The NLP API already creates the network, so we just need to refresh the list
      await fetchNetworks();
      setSuccess('Network created successfully from natural language!');
      setOpenDialog(false);
    } catch (err: any) {
      setError('Network creation completed but failed to refresh list');
    } finally {
      setLoading(false);
    }
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

          {!editMode && (
            <Tabs 
              value={creationMode} 
              onChange={(_, newValue) => setCreationMode(newValue)}
              sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}
            >
              <Tab 
                icon={<AIIcon />} 
                label="Natural Language" 
                value="nlp"
                sx={{ textTransform: 'none' }}
              />
              <Tab 
                icon={<AdvancedIcon />} 
                label="Advanced Configuration" 
                value="advanced"
                sx={{ textTransform: 'none' }}
              />
            </Tabs>
          )}

          {creationMode === 'nlp' && !editMode ? (
            <Box sx={{ mt: 2 }}>
              <Typography variant="h6" gutterBottom>
                Describe Your Network
              </Typography>
              <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                Describe the network you want to create in natural language. For example:
                "Create a star network with 2 routers, 1 host and firewall policy where SSH access is enabled"
              </Typography>
              
              <Alert severity="info" sx={{ mb: 2 }}>
                <Typography variant="body2">
                  <strong>Supported:</strong> Basic topologies (star, mesh, tree), node types (router, switch, host, firewall), 
                  and simple firewall policies (SSH, HTTP, HTTPS)
                </Typography>
              </Alert>

              <TextField
                fullWidth
                multiline
                rows={4}
                label="Network Description"
                placeholder="Create a star network with 2 routers, 1 host and firewall policy where SSH access is enabled"
                value={nlpText}
                onChange={(e) => setNlpText(e.target.value)}
                sx={{ mb: 2 }}
              />

              <Box sx={{ display: 'flex', gap: 2, mb: 2 }}>
                <Button
                  variant="outlined"
                  onClick={handleNlpPreview}
                  disabled={!nlpText.trim() || nlpLoading}
                  startIcon={nlpLoading ? <CircularProgress size={16} /> : <AIIcon />}
                >
                  {nlpLoading ? 'Analyzing...' : 'Preview Network'}
                </Button>
              </Box>

              {nlpPreview && (
                <Card variant="outlined" sx={{ mt: 2 }}>
                  <CardContent>
                    <Typography variant="h6" gutterBottom>
                      Preview: {nlpPreview.network?.name}
                    </Typography>
                    <Typography variant="body2" color="text.secondary" paragraph>
                      {nlpPreview.network?.description}
                    </Typography>

                    <Box sx={{ mb: 2 }}>
                      <Typography variant="subtitle2" gutterBottom>
                        Network Details:
                      </Typography>
                      <Chip label={`${nlpPreview.spec?.topology} topology`} size="small" sx={{ mr: 1 }} />
                      <Chip label={`${nlpPreview.nodes_created} nodes`} size="small" sx={{ mr: 1 }} />
                      <Chip 
                        label={`Confidence: ${Math.round((nlpPreview.validation?.confidence || 0) * 100)}%`} 
                        size="small" 
                        color={nlpPreview.validation?.confidence > 0.7 ? 'success' : 'warning'}
                      />
                    </Box>

                    {nlpPreview.nodes && (
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="subtitle2" gutterBottom>
                          Nodes:
                        </Typography>
                        {nlpPreview.nodes.map((node: any, index: number) => (
                          <Chip 
                            key={index}
                            label={`${node.name} (${node.type})`}
                            variant="outlined"
                            size="small"
                            sx={{ mr: 1, mb: 1 }}
                          />
                        ))}
                      </Box>
                    )}

                    {nlpPreview.spec?.policies && nlpPreview.spec.policies.length > 0 && (
                      <Box sx={{ mb: 2 }}>
                        <Typography variant="subtitle2" gutterBottom>
                          Policies:
                        </Typography>
                        {nlpPreview.spec.policies.map((policy: any, index: number) => (
                          <Box key={index} sx={{ mb: 1 }}>
                            <Typography variant="body2">
                              <strong>{policy.name}</strong> ({policy.type})
                            </Typography>
                            {policy.rules && policy.rules.map((rule: any, ruleIndex: number) => (
                              <Chip
                                key={ruleIndex}
                                label={`${rule.name || rule.action} ${rule.protocol}/${rule.port}`}
                                variant="outlined"
                                size="small"
                                sx={{ mr: 1, mt: 0.5 }}
                              />
                            ))}
                          </Box>
                        ))}
                      </Box>
                    )}

                    {nlpPreview.validation?.errors && nlpPreview.validation.errors.length > 0 && (
                      <Alert severity="warning" sx={{ mt: 2 }}>
                        <Typography variant="body2">
                          <strong>Validation Issues:</strong>
                        </Typography>
                        {nlpPreview.validation.errors.map((error: string, index: number) => (
                          <Typography key={index} variant="body2">
                            • {error}
                          </Typography>
                        ))}
                      </Alert>
                    )}
                  </CardContent>
                </Card>
              )}
            </Box>
          ) : (
            <>
              <Stepper activeStep={activeStep} sx={{ mt: 2, mb: 3 }}>
                {steps.map((label) => (
                  <Step key={label}>
                    <StepLabel>{label}</StepLabel>
                  </Step>
                ))}
              </Stepper>

              {renderStepContent(activeStep)}
            </>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          
          {creationMode === 'nlp' && !editMode ? (
            <Button
              variant="contained"
              onClick={handleNlpCreate}
              disabled={!nlpPreview || loading}
              startIcon={loading ? <CircularProgress size={16} /> : <AIIcon />}
            >
              {loading ? 'Creating...' : 'Create Network'}
            </Button>
          ) : (
            <>
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
            </>
          )}
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default NetworkManagement;

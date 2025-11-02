import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
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
  Stepper,
  Step,
  StepLabel,
  Alert,
  CircularProgress,
  Tooltip,
  Card,
  CardContent,
  Divider,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  PlayArrow as PlayIcon,
  Stop as StopIcon,
  Computer as ComputerIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';

interface Port {
  container: number;
  host: number;
  protocol: string;
}

interface NodeConfig {
  cpu: number;
  memory: number;
  storage: number;
  os: string;
  image: string;
  ports?: Port[];
  environment?: { [key: string]: string };
  volumes?: string[];
  commands?: string[];
}

interface Node {
  id: string;
  network_id: string;
  name: string;
  type: string;
  ip_address: string;
  mac_address?: string;
  status: string;
  config: NodeConfig;
  position: any;
  created_at: string;
  updated_at: string;
}

interface Network {
  id: string;
  name: string;
}

interface NodeManagementProps {
  networkId?: string;
}

const NodeManagement: React.FC<NodeManagementProps> = ({ networkId: propNetworkId }) => {
  const { api } = useAuth();
  const [nodes, setNodes] = useState<Node[]>([]);
  const [networks, setNetworks] = useState<Network[]>([]);
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>(propNetworkId || '');
  const [loading, setLoading] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [activeStep, setActiveStep] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    type: 'router',
    ip_address: '',
    mac_address: '',
    cpu: '2',
    memory: '2048',
    storage: '20',
    os: 'linux',
    image: '',
    ports: [] as Port[],
    environment: {} as { [key: string]: string },
    volumes: [] as string[],
    commands: [] as string[],
  });

  // Port form state
  const [portForm, setPortForm] = useState({ container: '', host: '', protocol: 'tcp' });
  const [envKey, setEnvKey] = useState('');
  const [envValue, setEnvValue] = useState('');
  const [volume, setVolume] = useState('');
  const [command, setCommand] = useState('');

  const steps = ['Basic Info', 'Resources', 'Advanced', 'Review'];

  const nodeTypes = [
    { value: 'router', label: 'Router', image: 'frrouting/frr:latest' },
    { value: 'switch', label: 'Switch', image: 'alpine:latest' },
    { value: 'host', label: 'Host', image: 'ubuntu:20.04' },
    { value: 'server', label: 'Server', image: 'nginx:latest' },
    { value: 'firewall', label: 'Firewall', image: 'alpine:latest' },
  ];

  useEffect(() => {
    fetchNetworks();
  }, []);

  useEffect(() => {
    if (selectedNetworkId) {
      fetchNodes(selectedNetworkId);
    }
  }, [selectedNetworkId]);

  const fetchNetworks = async () => {
    try {
      const response = await api.get('/api/v1/networks');
      setNetworks(response.data.networks || []);
    } catch (err: any) {
      setError('Failed to fetch networks');
    }
  };

  const fetchNodes = async (networkId: string) => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.get(`/api/v1/networks/${networkId}/nodes`);
      setNodes(response.data.nodes || []);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to fetch nodes');
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (node?: Node) => {
    if (node) {
      setEditMode(true);
      setSelectedNode(node);
      setFormData({
        name: node.name,
        type: node.type,
        ip_address: node.ip_address,
        mac_address: node.mac_address || '',
        cpu: node.config.cpu?.toString() || '2',
        memory: node.config.memory?.toString() || '2048',
        storage: node.config.storage?.toString() || '20',
        os: node.config.os || 'linux',
        image: node.config.image || '',
        ports: node.config.ports || [],
        environment: node.config.environment || {},
        volumes: node.config.volumes || [],
        commands: node.config.commands || [],
      });
    } else {
      setEditMode(false);
      setSelectedNode(null);
      const defaultType = nodeTypes[0];
      setFormData({
        name: '',
        type: defaultType.value,
        ip_address: '',
        mac_address: '',
        cpu: '2',
        memory: '2048',
        storage: '20',
        os: 'linux',
        image: defaultType.image,
        ports: [],
        environment: {},
        volumes: [],
        commands: [],
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

  const handleTypeChange = (type: string) => {
    const nodeType = nodeTypes.find(nt => nt.value === type);
    setFormData({ ...formData, type, image: nodeType?.image || '' });
  };

  const addPort = () => {
    if (portForm.container && portForm.host) {
      setFormData({
        ...formData,
        ports: [...formData.ports, {
          container: parseInt(portForm.container),
          host: parseInt(portForm.host),
          protocol: portForm.protocol,
        }],
      });
      setPortForm({ container: '', host: '', protocol: 'tcp' });
    }
  };

  const removePort = (index: number) => {
    setFormData({
      ...formData,
      ports: formData.ports.filter((_, i) => i !== index),
    });
  };

  const addEnvironment = () => {
    if (envKey && envValue) {
      setFormData({
        ...formData,
        environment: { ...formData.environment, [envKey]: envValue },
      });
      setEnvKey('');
      setEnvValue('');
    }
  };

  const removeEnvironment = (key: string) => {
    const newEnv = { ...formData.environment };
    delete newEnv[key];
    setFormData({ ...formData, environment: newEnv });
  };

  const addVolume = () => {
    if (volume) {
      setFormData({ ...formData, volumes: [...formData.volumes, volume] });
      setVolume('');
    }
  };

  const removeVolume = (index: number) => {
    setFormData({
      ...formData,
      volumes: formData.volumes.filter((_, i) => i !== index),
    });
  };

  const addCommand = () => {
    if (command) {
      setFormData({ ...formData, commands: [...formData.commands, command] });
      setCommand('');
    }
  };

  const removeCommand = (index: number) => {
    setFormData({
      ...formData,
      commands: formData.commands.filter((_, i) => i !== index),
    });
  };

  const handleSubmit = async () => {
    if (!selectedNetworkId) {
      setError('Please select a network first');
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const payload = {
        name: formData.name,
        type: formData.type,
        ip_address: formData.ip_address,
        ...(formData.mac_address && { mac_address: formData.mac_address }),
        config: {
          cpu: parseInt(formData.cpu),
          memory: parseInt(formData.memory),
          storage: parseInt(formData.storage),
          os: formData.os,
          image: formData.image,
          ...(formData.ports.length > 0 && { ports: formData.ports }),
          ...(Object.keys(formData.environment).length > 0 && { environment: formData.environment }),
          ...(formData.volumes.length > 0 && { volumes: formData.volumes }),
          ...(formData.commands.length > 0 && { commands: formData.commands }),
        },
      };

      if (editMode && selectedNode) {
        await api.put(`/api/v1/nodes/${selectedNode.id}`, payload);
        setSuccess('Node updated successfully!');
      } else {
        await api.post(`/api/v1/networks/${selectedNetworkId}/nodes`, payload);
        setSuccess('Node created successfully and container provisioned!');
      }

      await fetchNodes(selectedNetworkId);
      setTimeout(() => {
        handleCloseDialog();
      }, 1500);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save node');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (nodeId: string) => {
    if (!window.confirm('Are you sure you want to delete this node and its container?')) {
      return;
    }

    setLoading(true);
    try {
      await api.delete(`/api/v1/nodes/${nodeId}`);
      setSuccess('Node deleted successfully!');
      if (selectedNetworkId) {
        await fetchNodes(selectedNetworkId);
      }
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete node');
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
      case 'running':
        return 'success';
      case 'stopped':
      case 'inactive':
        return 'default';
      case 'provisioning':
      case 'starting':
        return 'warning';
      case 'error':
      case 'failed':
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
              label="Node Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              margin="normal"
              required
              helperText="Enter a unique name for this node"
            />
            <FormControl fullWidth margin="normal" required>
              <InputLabel>Node Type</InputLabel>
              <Select
                value={formData.type}
                label="Node Type"
                onChange={(e) => handleTypeChange(e.target.value)}
              >
                {nodeTypes.map((type) => (
                  <MenuItem key={type.value} value={type.value}>
                    {type.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            <TextField
              fullWidth
              label="IP Address"
              value={formData.ip_address}
              onChange={(e) => setFormData({ ...formData, ip_address: e.target.value })}
              margin="normal"
              required
              helperText="e.g., 10.0.0.10"
            />
            <TextField
              fullWidth
              label="MAC Address (Optional)"
              value={formData.mac_address}
              onChange={(e) => setFormData({ ...formData, mac_address: e.target.value })}
              margin="normal"
              helperText="e.g., 00:11:22:33:44:55"
            />
          </Box>
        );
      case 1:
        return (
          <Box sx={{ mt: 2 }}>
            <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
              <Box sx={{ flex: '1 1 30%', minWidth: '150px' }}>
                <TextField
                  fullWidth
                  label="CPU Cores"
                  type="number"
                  value={formData.cpu}
                  onChange={(e) => setFormData({ ...formData, cpu: e.target.value })}
                  margin="normal"
                  required
                  helperText="Number of CPU cores"
                />
              </Box>
              <Box sx={{ flex: '1 1 30%', minWidth: '150px' }}>
                <TextField
                  fullWidth
                  label="Memory (MB)"
                  type="number"
                  value={formData.memory}
                  onChange={(e) => setFormData({ ...formData, memory: e.target.value })}
                  margin="normal"
                  required
                  helperText="Memory in MB"
                />
              </Box>
              <Box sx={{ flex: '1 1 30%', minWidth: '150px' }}>
                <TextField
                  fullWidth
                  label="Storage (GB)"
                  type="number"
                  value={formData.storage}
                  onChange={(e) => setFormData({ ...formData, storage: e.target.value })}
                  margin="normal"
                  required
                  helperText="Storage in GB"
                />
              </Box>
              <Box sx={{ flex: '1 1 45%', minWidth: '200px' }}>
                <TextField
                  fullWidth
                  label="Operating System"
                  value={formData.os}
                  onChange={(e) => setFormData({ ...formData, os: e.target.value })}
                  margin="normal"
                  required
                />
              </Box>
              <Box sx={{ flex: '1 1 45%', minWidth: '200px' }}>
                <TextField
                  fullWidth
                  label="Container Image"
                  value={formData.image}
                  onChange={(e) => setFormData({ ...formData, image: e.target.value })}
                  margin="normal"
                  required
                  helperText="Docker/Podman image"
                />
              </Box>
            </Box>
          </Box>
        );
      case 2:
        return (
          <Box sx={{ mt: 2 }}>
            <Typography variant="h6" gutterBottom>Port Mappings</Typography>
            <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
              <TextField
                label="Container Port"
                type="number"
                value={portForm.container}
                onChange={(e) => setPortForm({ ...portForm, container: e.target.value })}
                size="small"
              />
              <TextField
                label="Host Port"
                type="number"
                value={portForm.host}
                onChange={(e) => setPortForm({ ...portForm, host: e.target.value })}
                size="small"
              />
              <Select
                value={portForm.protocol}
                onChange={(e) => setPortForm({ ...portForm, protocol: e.target.value })}
                size="small"
              >
                <MenuItem value="tcp">TCP</MenuItem>
                <MenuItem value="udp">UDP</MenuItem>
              </Select>
              <Button onClick={addPort} variant="outlined" size="small">Add</Button>
            </Box>
            <Box sx={{ mb: 3 }}>
              {formData.ports.map((port, index) => (
                <Chip
                  key={index}
                  label={`${port.container}:${port.host}/${port.protocol}`}
                  onDelete={() => removePort(index)}
                  sx={{ mr: 1, mb: 1 }}
                />
              ))}
            </Box>

            <Divider sx={{ my: 2 }} />

            <Typography variant="h6" gutterBottom>Environment Variables</Typography>
            <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
              <TextField
                label="Key"
                value={envKey}
                onChange={(e) => setEnvKey(e.target.value)}
                size="small"
                sx={{ flex: 1 }}
              />
              <TextField
                label="Value"
                value={envValue}
                onChange={(e) => setEnvValue(e.target.value)}
                size="small"
                sx={{ flex: 1 }}
              />
              <Button onClick={addEnvironment} variant="outlined" size="small">Add</Button>
            </Box>
            <Box sx={{ mb: 3 }}>
              {Object.entries(formData.environment).map(([key, value]) => (
                <Chip
                  key={key}
                  label={`${key}=${value}`}
                  onDelete={() => removeEnvironment(key)}
                  sx={{ mr: 1, mb: 1 }}
                />
              ))}
            </Box>

            <Divider sx={{ my: 2 }} />

            <Typography variant="h6" gutterBottom>Volumes</Typography>
            <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
              <TextField
                label="Volume"
                value={volume}
                onChange={(e) => setVolume(e.target.value)}
                placeholder="/host/path:/container/path"
                size="small"
                fullWidth
              />
              <Button onClick={addVolume} variant="outlined" size="small">Add</Button>
            </Box>
            <Box sx={{ mb: 3 }}>
              {formData.volumes.map((vol, index) => (
                <Chip
                  key={index}
                  label={vol}
                  onDelete={() => removeVolume(index)}
                  sx={{ mr: 1, mb: 1 }}
                />
              ))}
            </Box>

            <Divider sx={{ my: 2 }} />

            <Typography variant="h6" gutterBottom>Startup Commands</Typography>
            <Box sx={{ display: 'flex', gap: 1, mb: 2 }}>
              <TextField
                label="Command"
                value={command}
                onChange={(e) => setCommand(e.target.value)}
                size="small"
                fullWidth
              />
              <Button onClick={addCommand} variant="outlined" size="small">Add</Button>
            </Box>
            <Box>
              {formData.commands.map((cmd, index) => (
                <Chip
                  key={index}
                  label={cmd}
                  onDelete={() => removeCommand(index)}
                  sx={{ mr: 1, mb: 1 }}
                />
              ))}
            </Box>
          </Box>
        );
      case 3:
        return (
          <Box sx={{ mt: 2 }}>
            <Typography variant="h6" gutterBottom>Review Node Configuration</Typography>
            <TableContainer component={Paper} sx={{ mt: 2 }}>
              <Table size="small">
                <TableBody>
                  <TableRow>
                    <TableCell><strong>Name</strong></TableCell>
                    <TableCell>{formData.name}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Type</strong></TableCell>
                    <TableCell>{formData.type}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>IP Address</strong></TableCell>
                    <TableCell>{formData.ip_address}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>CPU</strong></TableCell>
                    <TableCell>{formData.cpu} cores</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Memory</strong></TableCell>
                    <TableCell>{formData.memory} MB</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Storage</strong></TableCell>
                    <TableCell>{formData.storage} GB</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Image</strong></TableCell>
                    <TableCell>{formData.image}</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Ports</strong></TableCell>
                    <TableCell>{formData.ports.length} mappings</TableCell>
                  </TableRow>
                  <TableRow>
                    <TableCell><strong>Environment</strong></TableCell>
                    <TableCell>{Object.keys(formData.environment).length} variables</TableCell>
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
        <Typography variant="h4">Node Management</Typography>
        <Box sx={{ display: 'flex', gap: 1 }}>
          <FormControl sx={{ minWidth: 200 }}>
            <InputLabel>Select Network</InputLabel>
            <Select
              value={selectedNetworkId}
              label="Select Network"
              onChange={(e) => setSelectedNetworkId(e.target.value)}
              size="small"
            >
              {networks.map((network) => (
                <MenuItem key={network.id} value={network.id}>
                  {network.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <Button
            variant="outlined"
            startIcon={<RefreshIcon />}
            onClick={() => selectedNetworkId && fetchNodes(selectedNetworkId)}
          >
            Refresh
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => handleOpenDialog()}
            disabled={!selectedNetworkId}
          >
            Add Node
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

      {!selectedNetworkId ? (
        <Card>
          <CardContent>
            <Typography color="textSecondary" align="center">
              Please select a network to view and manage nodes
            </Typography>
          </CardContent>
        </Card>
      ) : loading && nodes.length === 0 ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Name</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>IP Address</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Image</TableCell>
                <TableCell>Resources</TableCell>
                <TableCell>Created</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {nodes.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} align="center">
                    <Typography color="textSecondary" sx={{ py: 3 }}>
                      No nodes found. Add your first node to get started.
                    </Typography>
                  </TableCell>
                </TableRow>
              ) : (
                nodes.map((node) => (
                  <TableRow key={node.id} hover>
                    <TableCell>{node.name}</TableCell>
                    <TableCell>
                      <Chip label={node.type} size="small" />
                    </TableCell>
                    <TableCell>{node.ip_address}</TableCell>
                    <TableCell>
                      <Chip
                        label={node.status}
                        color={getStatusColor(node.status) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>{node.config?.image || 'N/A'}</TableCell>
                    <TableCell>
                      {node.config?.cpu}C / {node.config?.memory}MB
                    </TableCell>
                    <TableCell>
                      {new Date(node.created_at).toLocaleDateString()}
                    </TableCell>
                    <TableCell>
                      <Tooltip title="Edit">
                        <IconButton
                          size="small"
                          onClick={() => handleOpenDialog(node)}
                        >
                          <EditIcon />
                        </IconButton>
                      </Tooltip>
                      <Tooltip title="Delete">
                        <IconButton
                          size="small"
                          color="error"
                          onClick={() => handleDelete(node.id)}
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
          {editMode ? 'Edit Node' : 'Create New Node'}
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
              disabled={!formData.name || !formData.ip_address || !formData.image}
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

export default NodeManagement;

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
  Alert,
  CircularProgress,
  Card,
  CardContent,
  Divider,
  Stack,
  Switch,
  FormControlLabel,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  Security as SecurityIcon,
  CheckCircle as ActiveIcon,
  Error as ErrorIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { useApi } from '../contexts/ApiContext';

interface PolicyConfig {
  priority?: number;
  action?: string;
  enabled?: boolean;
  rules?: any[];
  custom_attrs?: { [key: string]: any };
}

interface Policy {
  id: string;
  network_id: string;
  type: string;  // Changed from policy_type to match backend
  name: string;
  description?: string;
  status: string;
  config: PolicyConfig;
  created_at: string;
  updated_at: string;
}

interface Network {
  id: string;
  name: string;
}

const PolicyManagement: React.FC = () => {
  const { api } = useAuth();
  const { createPolicy, deletePolicy } = useApi();
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [networks, setNetworks] = useState<Network[]>([]);
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [selectedPolicy, setSelectedPolicy] = useState<Policy | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    type: 'firewall',  // Changed from policy_type to type
    priority: 100,
    enabled: true,
    rules: [{
      id: '',
      name: '',
      protocol: 'tcp',
      port: '22',
      action: 'allow'
    }]
  });

  useEffect(() => {
    fetchNetworks();
  }, []);

  useEffect(() => {
    if (selectedNetworkId) {
      fetchPolicies();
    }
  }, [selectedNetworkId]);

  const fetchNetworks = async () => {
    try {
      const response = await api.get('/api/v1/networks');
      setNetworks(response.data.networks || []);
      if (response.data.networks && response.data.networks.length > 0) {
        setSelectedNetworkId(response.data.networks[0].id);
      }
    } catch (err: any) {
      setError('Failed to fetch networks');
    }
  };

  const fetchPolicies = async () => {
    if (!selectedNetworkId) return;
    
    setLoading(true);
    try {
      const response = await api.get(`/api/v1/networks/${selectedNetworkId}/policies`);
      setPolicies(response.data.policies || []);
      setError(null);
    } catch (err: any) {
      setError('Failed to fetch policies');
      setPolicies([]);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = (policy?: Policy) => {
    if (policy) {
      setEditMode(true);
      setSelectedPolicy(policy);
      setFormData({
        name: policy.name,
        description: policy.description || '',
        type: policy.type,  // Changed from policy_type to type
        priority: policy.config.priority || 100,
        enabled: policy.status === 'active',
        rules: policy.config.rules?.map(rule => ({
          id: rule.id || '',
          name: rule.name || '',
          protocol: rule.condition?.protocol || 'tcp',
          port: rule.condition?.port || '22',
          action: rule.action?.action || 'allow'
        })) || [{
          id: '',
          name: '',
          protocol: 'tcp',
          port: '22',
          action: 'allow'
        }]
      });
    } else {
      setEditMode(false);
      setSelectedPolicy(null);
      setFormData({
        name: '',
        description: '',
        type: 'firewall',  // Changed from policy_type to type
        priority: 100,
        enabled: true,
        rules: [{
          id: '',
          name: '',
          protocol: 'tcp',
          port: '22',
          action: 'allow'
        }]
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setEditMode(false);
    setSelectedPolicy(null);
    setError(null);
  };

  const handleSubmit = async () => {
    if (!selectedNetworkId) {
      setError('Please select a network first');
      return;
    }

    if (!formData.name) {
      setError('Please enter a policy name');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const policyData = {
        name: formData.name,
        type: formData.type,  // Changed from policy_type to type
        config: {
          rules: formData.rules.map((rule, index) => ({
            id: rule.id || `rule-${index + 1}`,
            name: rule.name || `Rule ${index + 1}`,
            condition: {
              protocol: rule.protocol,
              port: rule.port
            },
            action: {
              action: rule.action
            },
            priority: formData.priority + index * 10  // Increment priority for multiple rules
          })),
          priority: formData.priority
        },
      };
      
      // Add description if provided
      if (formData.description) {
        policyData.description = formData.description;
      }

      if (editMode && selectedPolicy) {
        // Update policy
        await api.put(`/api/v1/networks/${selectedNetworkId}/policies/${selectedPolicy.id}`, policyData);
        setSuccess('Policy updated successfully');
      } else {
        // Create new policy
        await api.post(`/api/v1/networks/${selectedNetworkId}/policies`, policyData);
        setSuccess('Policy created successfully');
      }

      handleCloseDialog();
      fetchPolicies();
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save policy');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (policyId: string) => {
    if (!window.confirm('Are you sure you want to delete this policy?')) {
      return;
    }

    if (!selectedNetworkId) {
      setError('No network selected');
      return;
    }

    setLoading(true);
    try {
      await api.delete(`/api/v1/networks/${selectedNetworkId}/policies/${policyId}`);
      setSuccess('Policy deleted successfully');
      fetchPolicies();
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError('Failed to delete policy');
    } finally {
      setLoading(false);
    }
  };

  const getPolicyTypeColor = (type: string) => {
    const colors: { [key: string]: any } = {
      firewall: 'error',
      qos: 'warning',
      routing: 'info',
      security: 'secondary',
      traffic: 'success',
      access: 'primary',
    };
    return colors[type] || 'default';
  };

  const getStatusIcon = (status: string) => {
    return status === 'active' ? (
      <ActiveIcon color="success" fontSize="small" />
    ) : (
      <ErrorIcon color="error" fontSize="small" />
    );
  };

  return (
    <Box>
      <Card>
        <CardContent>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
            <Typography variant="h5">
              <SecurityIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
              Policy Management
            </Typography>
            <Box sx={{ display: 'flex', gap: 2 }}>
              <FormControl sx={{ minWidth: 250 }}>
                <InputLabel>Network</InputLabel>
                <Select
                  value={selectedNetworkId}
                  label="Network"
                  onChange={(e) => setSelectedNetworkId(e.target.value)}
                >
                  {networks.map((network) => (
                    <MenuItem key={network.id} value={network.id}>
                      {network.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={() => handleOpenDialog()}
                disabled={!selectedNetworkId}
              >
                Add Policy
              </Button>
              <IconButton onClick={fetchPolicies} disabled={loading}>
                <RefreshIcon />
              </IconButton>
            </Box>
          </Box>

          {success && (
            <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess(null)}>
              {success}
            </Alert>
          )}

          {error && (
            <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
              {error}
            </Alert>
          )}

          {!selectedNetworkId ? (
            <Alert severity="info">Please select a network to view policies</Alert>
          ) : loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
              <CircularProgress />
            </Box>
          ) : policies.length === 0 ? (
            <Alert severity="info">
              No policies found. Click "Add Policy" to create security and network policies.
            </Alert>
          ) : (
            <TableContainer component={Paper} variant="outlined">
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Status</TableCell>
                    <TableCell>Name</TableCell>
                    <TableCell>Type</TableCell>
                    <TableCell>Description</TableCell>
                    <TableCell>Priority</TableCell>
                    <TableCell>Action</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {policies.map((policy) => (
                    <TableRow key={policy.id}>
                      <TableCell>{getStatusIcon(policy.status)}</TableCell>
                      <TableCell>
                        <Typography variant="body2" fontWeight="medium">
                          {policy.name}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={policy.type}
                          color={getPolicyTypeColor(policy.type)}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {policy.description || 'N/A'}
                        </Typography>
                      </TableCell>
                      <TableCell>{policy.config.priority || 'N/A'}</TableCell>
                      <TableCell>
                        <Chip
                          label={policy.config.action || 'N/A'}
                          color={policy.config.action === 'allow' ? 'success' : 'error'}
                          size="small"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          onClick={() => handleOpenDialog(policy)}
                          color="primary"
                        >
                          <EditIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => handleDelete(policy.id)}
                          color="error"
                        >
                          <DeleteIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </CardContent>
      </Card>

      {/* Add/Edit Policy Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <DialogTitle>
          {editMode ? 'Edit Policy' : 'Create New Policy'}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ pt: 2 }}>
            {error && (
              <Alert severity="error">
                {error}
              </Alert>
            )}

            <TextField
              label="Policy Name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              fullWidth
              required
            />

            <TextField
              label="Description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              fullWidth
              multiline
              rows={2}
            />

            <FormControl fullWidth>
              <InputLabel>Policy Type</InputLabel>
              <Select
                value={formData.type}
                label="Policy Type"
                onChange={(e) => setFormData({ ...formData, type: e.target.value })}
              >
                <MenuItem value="firewall">Firewall</MenuItem>
                <MenuItem value="qos">Quality of Service (QoS)</MenuItem>
                <MenuItem value="routing">Routing</MenuItem>
                <MenuItem value="security">Security</MenuItem>
                <MenuItem value="traffic">Traffic</MenuItem>
                <MenuItem value="access">Access Control</MenuItem>
              </Select>
            </FormControl>

            <TextField
              label="Priority"
              type="number"
              value={formData.priority}
              onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) || 100 })}
              fullWidth
              helperText="Lower numbers = higher priority (1-1000)"
              inputProps={{ min: 1, max: 1000 }}
            />

            <Divider>
              <Typography variant="body2" color="text.secondary">
                Policy Rules
              </Typography>
            </Divider>

            {formData.rules.map((rule, index) => (
              <Card key={index} variant="outlined" sx={{ p: 2 }}>
                <Stack spacing={2}>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Typography variant="subtitle2">Rule {index + 1}</Typography>
                    {formData.rules.length > 1 && (
                      <IconButton 
                        size="small" 
                        onClick={() => {
                          const newRules = formData.rules.filter((_, i) => i !== index);
                          setFormData({ ...formData, rules: newRules });
                        }}
                      >
                        <DeleteIcon />
                      </IconButton>
                    )}
                  </Box>
                  
                  <Box sx={{ display: 'flex', gap: 2 }}>
                    <TextField
                      label="Rule ID"
                      value={rule.id}
                      onChange={(e) => {
                        const newRules = [...formData.rules];
                        newRules[index] = { ...rule, id: e.target.value };
                        setFormData({ ...formData, rules: newRules });
                      }}
                      fullWidth
                      placeholder="e.g., ssh-allow, http-allow"
                    />
                    
                    <TextField
                      label="Rule Name"
                      value={rule.name}
                      onChange={(e) => {
                        const newRules = [...formData.rules];
                        newRules[index] = { ...rule, name: e.target.value };
                        setFormData({ ...formData, rules: newRules });
                      }}
                      fullWidth
                      placeholder="e.g., Allow SSH, Allow HTTP"
                    />
                  </Box>
                  
                  <Box sx={{ display: 'flex', gap: 2 }}>
                    <FormControl fullWidth>
                      <InputLabel>Protocol</InputLabel>
                      <Select
                        value={rule.protocol}
                        label="Protocol"
                        onChange={(e) => {
                          const newRules = [...formData.rules];
                          newRules[index] = { ...rule, protocol: e.target.value };
                          setFormData({ ...formData, rules: newRules });
                        }}
                      >
                        <MenuItem value="tcp">TCP</MenuItem>
                        <MenuItem value="udp">UDP</MenuItem>
                        <MenuItem value="icmp">ICMP</MenuItem>
                        <MenuItem value="any">Any</MenuItem>
                      </Select>
                    </FormControl>
                    
                    <TextField
                      label="Port"
                      value={rule.port}
                      onChange={(e) => {
                        const newRules = [...formData.rules];
                        newRules[index] = { ...rule, port: e.target.value };
                        setFormData({ ...formData, rules: newRules });
                      }}
                      fullWidth
                      placeholder="e.g., 22, 80, 443"
                    />
                    
                    <FormControl fullWidth>
                      <InputLabel>Action</InputLabel>
                      <Select
                        value={rule.action}
                        label="Action"
                        onChange={(e) => {
                          const newRules = [...formData.rules];
                          newRules[index] = { ...rule, action: e.target.value };
                          setFormData({ ...formData, rules: newRules });
                        }}
                      >
                        <MenuItem value="allow">Allow</MenuItem>
                        <MenuItem value="deny">Deny</MenuItem>
                        <MenuItem value="drop">Drop</MenuItem>
                      </Select>
                    </FormControl>
                  </Box>
                </Stack>
              </Card>
            ))}
            
            <Button
              variant="outlined"
              startIcon={<AddIcon />}
              onClick={() => {
                const newRule = {
                  id: '',
                  name: '',
                  protocol: 'tcp',
                  port: '',
                  action: 'allow'
                };
                setFormData({ ...formData, rules: [...formData.rules, newRule] });
              }}
            >
              Add Another Rule
            </Button>

            <FormControlLabel
              control={
                <Switch
                  checked={formData.enabled}
                  onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })}
                />
              }
              label="Enable Policy"
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button
            onClick={handleSubmit}
            variant="contained"
            disabled={loading}
          >
            {loading ? <CircularProgress size={24} /> : editMode ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default PolicyManagement;

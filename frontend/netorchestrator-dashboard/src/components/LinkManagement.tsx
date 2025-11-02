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
  Slider,
  Stack,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Refresh as RefreshIcon,
  Link as LinkIcon,
  CheckCircle as ActiveIcon,
  Error as ErrorIcon,
} from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';
import { useApi } from '../contexts/ApiContext';

interface LinkConfig {
  bandwidth?: number;
  latency?: number;
  packet_loss?: number;
  jitter?: number;
  enabled?: boolean;
  custom_attrs?: { [key: string]: any };
}

interface Link {
  id: string;
  network_id: string;
  source_node_id: string;
  target_node_id: string;
  link_type: string;
  status: string;
  config: LinkConfig;
  created_at: string;
  updated_at: string;
  source_node?: any;
  target_node?: any;
}

interface Node {
  id: string;
  name: string;
  type: string;
  network_id: string;
}

interface Network {
  id: string;
  name: string;
}

const LinkManagement: React.FC = () => {
  const { api } = useAuth();
  const { createLink, deleteLink } = useApi();
  const [links, setLinks] = useState<Link[]>([]);
  const [nodes, setNodes] = useState<Node[]>([]);
  const [networks, setNetworks] = useState<Network[]>([]);
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [openDialog, setOpenDialog] = useState(false);
  const [editMode, setEditMode] = useState(false);
  const [selectedLink, setSelectedLink] = useState<Link | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Form state
  const [formData, setFormData] = useState({
    source_node_id: '',
    target_node_id: '',
    link_type: 'ethernet',
    bandwidth: 1000, // Mbps
    latency: 1, // ms
    packet_loss: 0, // percentage
    jitter: 0, // ms
  });

  useEffect(() => {
    fetchNetworks();
  }, []);

  useEffect(() => {
    if (selectedNetworkId) {
      fetchLinks();
      fetchNodes();
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

  const fetchLinks = async () => {
    if (!selectedNetworkId) return;
    
    setLoading(true);
    try {
      const response = await api.get(`/api/v1/networks/${selectedNetworkId}/links`);
      setLinks(response.data.links || []);
      setError(null);
    } catch (err: any) {
      setError('Failed to fetch links');
      setLinks([]);
    } finally {
      setLoading(false);
    }
  };

  const fetchNodes = async () => {
    if (!selectedNetworkId) return;
    
    try {
      const response = await api.get(`/api/v1/networks/${selectedNetworkId}/nodes`);
      setNodes(response.data.nodes || []);
    } catch (err: any) {
      console.error('Failed to fetch nodes:', err);
      setNodes([]);
    }
  };

  const handleOpenDialog = (link?: Link) => {
    if (link) {
      setEditMode(true);
      setSelectedLink(link);
      setFormData({
        source_node_id: link.source_node_id,
        target_node_id: link.target_node_id,
        link_type: link.link_type,
        bandwidth: link.config.bandwidth || 1000,
        latency: link.config.latency || 1,
        packet_loss: link.config.packet_loss || 0,
        jitter: link.config.jitter || 0,
      });
    } else {
      setEditMode(false);
      setSelectedLink(null);
      setFormData({
        source_node_id: '',
        target_node_id: '',
        link_type: 'ethernet',
        bandwidth: 1000,
        latency: 1,
        packet_loss: 0,
        jitter: 0,
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setEditMode(false);
    setSelectedLink(null);
    setError(null);
  };

  const handleSubmit = async () => {
    if (!selectedNetworkId) {
      setError('Please select a network first');
      return;
    }

    if (!formData.source_node_id || !formData.target_node_id) {
      setError('Please select both source and target nodes');
      return;
    }

    if (formData.source_node_id === formData.target_node_id) {
      setError('Source and target nodes cannot be the same');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const linkData = {
        source_node_id: formData.source_node_id,
        target_node_id: formData.target_node_id,
        link_type: formData.link_type,
        status: 'active',
        config: {
          bandwidth: formData.bandwidth,
          latency: formData.latency,
          packet_loss: formData.packet_loss,
          jitter: formData.jitter,
          enabled: true,
        },
      };

      if (editMode && selectedLink) {
        // Update link
        await api.put(`/api/v1/networks/${selectedNetworkId}/links/${selectedLink.id}`, linkData);
        setSuccess('Link updated successfully');
      } else {
        // Create new link
        await api.post(`/api/v1/networks/${selectedNetworkId}/links`, linkData);
        setSuccess('Link created successfully');
      }

      handleCloseDialog();
      fetchLinks();
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save link');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (linkId: string) => {
    if (!window.confirm('Are you sure you want to delete this link?')) {
      return;
    }

    if (!selectedNetworkId) {
      setError('No network selected');
      return;
    }

    setLoading(true);
    try {
      await api.delete(`/api/v1/networks/${selectedNetworkId}/links/${linkId}`);
      setSuccess('Link deleted successfully');
      fetchLinks();
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError('Failed to delete link');
    } finally {
      setLoading(false);
    }
  };

  const getLinkTypeColor = (type: string) => {
    const colors: { [key: string]: any } = {
      ethernet: 'primary',
      fiber: 'success',
      wireless: 'info',
      virtual: 'warning',
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
              <LinkIcon sx={{ mr: 1, verticalAlign: 'middle' }} />
              Link Management
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
                disabled={!selectedNetworkId || nodes.length < 2}
              >
                Add Link
              </Button>
              <IconButton onClick={fetchLinks} disabled={loading}>
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
            <Alert severity="info">Please select a network to view links</Alert>
          ) : nodes.length < 2 ? (
            <Alert severity="warning">
              You need at least 2 nodes in this network to create links. Go to Nodes to add more nodes.
            </Alert>
          ) : loading ? (
            <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
              <CircularProgress />
            </Box>
          ) : links.length === 0 ? (
            <Alert severity="info">
              No links found. Click "Add Link" to create connections between nodes.
            </Alert>
          ) : (
            <TableContainer component={Paper} variant="outlined">
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Status</TableCell>
                    <TableCell>Source Node</TableCell>
                    <TableCell>Target Node</TableCell>
                    <TableCell>Type</TableCell>
                    <TableCell>Bandwidth (Mbps)</TableCell>
                    <TableCell>Latency (ms)</TableCell>
                    <TableCell>Packet Loss (%)</TableCell>
                    <TableCell>Jitter (ms)</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {links.map((link) => (
                    <TableRow key={link.id}>
                      <TableCell>{getStatusIcon(link.status)}</TableCell>
                      <TableCell>
                        {link.source_node?.name || link.source_node_id}
                      </TableCell>
                      <TableCell>
                        {link.target_node?.name || link.target_node_id}
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={link.link_type}
                          color={getLinkTypeColor(link.link_type)}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>{link.config.bandwidth || 'N/A'}</TableCell>
                      <TableCell>{link.config.latency || 'N/A'}</TableCell>
                      <TableCell>{link.config.packet_loss || 0}</TableCell>
                      <TableCell>{link.config.jitter || 0}</TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          onClick={() => handleOpenDialog(link)}
                          color="primary"
                        >
                          <EditIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => handleDelete(link.id)}
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

      {/* Add/Edit Link Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <DialogTitle>
          {editMode ? 'Edit Link' : 'Create New Link'}
        </DialogTitle>
        <DialogContent>
          <Stack spacing={3} sx={{ pt: 2 }}>
            {error && (
              <Alert severity="error">
                {error}
              </Alert>
            )}

            <Box sx={{ display: 'flex', gap: 2 }}>
              <FormControl fullWidth>
                <InputLabel>Source Node</InputLabel>
                <Select
                  value={formData.source_node_id}
                  label="Source Node"
                  onChange={(e) => setFormData({ ...formData, source_node_id: e.target.value })}
                  disabled={editMode}
                >
                  {nodes.map((node) => (
                    <MenuItem key={node.id} value={node.id}>
                      {node.name} ({node.type})
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>

              <FormControl fullWidth>
                <InputLabel>Target Node</InputLabel>
                <Select
                  value={formData.target_node_id}
                  label="Target Node"
                  onChange={(e) => setFormData({ ...formData, target_node_id: e.target.value })}
                  disabled={editMode}
                >
                  {nodes
                    .filter((node) => node.id !== formData.source_node_id)
                    .map((node) => (
                      <MenuItem key={node.id} value={node.id}>
                        {node.name} ({node.type})
                      </MenuItem>
                    ))}
                </Select>
              </FormControl>
            </Box>

            <FormControl fullWidth>
              <InputLabel>Link Type</InputLabel>
              <Select
                value={formData.link_type}
                label="Link Type"
                onChange={(e) => setFormData({ ...formData, link_type: e.target.value })}
              >
                <MenuItem value="ethernet">Ethernet</MenuItem>
                <MenuItem value="fiber">Fiber Optic</MenuItem>
                <MenuItem value="wireless">Wireless</MenuItem>
                <MenuItem value="virtual">Virtual</MenuItem>
              </Select>
            </FormControl>

            <Divider sx={{ my: 1 }}>
              <Typography variant="body2" color="text.secondary">
                Performance Configuration
              </Typography>
            </Divider>

            <Box>
              <Typography gutterBottom>
                Bandwidth: {formData.bandwidth} Mbps
              </Typography>
              <Slider
                value={formData.bandwidth}
                onChange={(_, value) => setFormData({ ...formData, bandwidth: value as number })}
                min={1}
                max={10000}
                step={10}
                marks={[
                  { value: 10, label: '10M' },
                  { value: 100, label: '100M' },
                  { value: 1000, label: '1G' },
                  { value: 10000, label: '10G' },
                ]}
              />
            </Box>

            <Box>
              <Typography gutterBottom>
                Latency: {formData.latency} ms
              </Typography>
              <Slider
                value={formData.latency}
                onChange={(_, value) => setFormData({ ...formData, latency: value as number })}
                min={0}
                max={1000}
                step={1}
                marks={[
                  { value: 0, label: '0ms' },
                  { value: 10, label: '10ms' },
                  { value: 100, label: '100ms' },
                  { value: 1000, label: '1s' },
                ]}
              />
            </Box>

            <Box sx={{ display: 'flex', gap: 2 }}>
              <Box sx={{ flex: 1 }}>
                <Typography gutterBottom>
                  Packet Loss: {formData.packet_loss}%
                </Typography>
                <Slider
                  value={formData.packet_loss}
                  onChange={(_, value) => setFormData({ ...formData, packet_loss: value as number })}
                  min={0}
                  max={100}
                  step={0.1}
                  marks={[
                    { value: 0, label: '0%' },
                    { value: 5, label: '5%' },
                    { value: 10, label: '10%' },
                  ]}
                />
              </Box>

              <Box sx={{ flex: 1 }}>
                <Typography gutterBottom>
                  Jitter: {formData.jitter} ms
                </Typography>
                <Slider
                  value={formData.jitter}
                  onChange={(_, value) => setFormData({ ...formData, jitter: value as number })}
                  min={0}
                  max={100}
                  step={1}
                  marks={[
                    { value: 0, label: '0ms' },
                    { value: 10, label: '10ms' },
                    { value: 50, label: '50ms' },
                  ]}
                />
              </Box>
            </Box>
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

export default LinkManagement;

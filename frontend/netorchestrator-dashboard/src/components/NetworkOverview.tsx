import React, { useState, useEffect } from 'react';
import {
  Card,
  CardContent,
  Typography,
  Box,
  Button,
  Grid,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Tabs,
  Tab,
  CircularProgress,
  Alert,
} from '@mui/material';
import {
  Add as AddIcon,
  Delete as DeleteIcon,
  Storage as NodeIcon,
  Link as LinkIcon,
  Security as PolicyIcon,
} from '@mui/icons-material';
import { useApi } from '../contexts/ApiContext';

interface Network {
  id: string;
  name: string;
  description: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface Node {
  id: string;
  network_id: string;
  name: string;
  type: string;
  ip_address: string;
  status: string;
  created_at: string;
}

interface Link {
  id: string;
  network_id: string;
  source_node_id: string;
  target_node_id: string;
  bandwidth_mbps?: number;
  status: string;
  created_at: string;
}

interface Policy {
  id: string;
  network_id: string;
  name: string;
  type: string;
  status: string;
  created_at: string;
}

interface NetworkOverviewProps {
  networks: Network[];
  loading: boolean;
  error: string | null;
}

const NetworkOverview: React.FC<NetworkOverviewProps> = ({ networks, loading, error }) => {
  const { nodes: allNodes, links: allLinks, policies: allPolicies, createNode, fetchNodes, fetchLinks, createLink, fetchPolicies, createPolicy } = useApi();
  const [selectedNetwork, setSelectedNetwork] = useState<Network | null>(null);
  const [activeTab, setActiveTab] = useState(0);
  const [showNodeDialog, setShowNodeDialog] = useState(false);
  const [showLinkDialog, setShowLinkDialog] = useState(false);
  const [showPolicyDialog, setShowPolicyDialog] = useState(false);
  const [loadingDetails, setLoadingDetails] = useState(false);
  
  // Filter nodes, links, and policies for selected network
  const nodes = allNodes.filter(n => !selectedNetwork || n.network_id === selectedNetwork.id);
  const links = allLinks.filter(l => !selectedNetwork || l.network_id === selectedNetwork.id);
  const policies = allPolicies.filter(p => !selectedNetwork || p.network_id === selectedNetwork.id);

  const [nodeForm, setNodeForm] = useState({
    name: '',
    type: 'router',
    ip_address: '',
  });

  const [linkForm, setLinkForm] = useState({
    source_node_id: '',
    target_node_id: '',
    bandwidth: '1000',
  });

  const [policyForm, setPolicyForm] = useState({
    name: '',
    type: 'firewall',
    description: '',
  });

  // Fetch details when network is selected
  useEffect(() => {
    if (selectedNetwork) {
      fetchNetworkDetails();
    }
  }, [selectedNetwork]);

  const fetchNetworkDetails = async () => {
    if (!selectedNetwork) return;
    setLoadingDetails(true);
    try {
      // Fetch nodes for this network
      await fetchNodes(selectedNetwork.id);
      
      // Fetch links and policies for this network
      await fetchLinks(selectedNetwork.id);
      await fetchPolicies(selectedNetwork.id);
    } catch (error) {
      console.error('Failed to fetch network details:', error);
    } finally {
      setLoadingDetails(false);
    }
  };

  const handleAddNode = async () => {
    if (!selectedNetwork) return;
    try {
      await createNode(selectedNetwork.id, nodeForm);
      setShowNodeDialog(false);
      setNodeForm({ name: '', type: 'router', ip_address: '' });
      fetchNetworkDetails();
    } catch (error) {
      console.error('Failed to create node:', error);
      alert('Failed to create node');
    }
  };

  const handleAddLink = async () => {
    if (!selectedNetwork) return;
    try {
      await createLink(selectedNetwork.id, {
        source_node_id: linkForm.source_node_id,
        target_node_id: linkForm.target_node_id,
        bandwidth_mbps: parseInt(linkForm.bandwidth),
      });
      setShowLinkDialog(false);
      setLinkForm({ source_node_id: '', target_node_id: '', bandwidth: '1000' });
      fetchNetworkDetails();
    } catch (error) {
      console.error('Failed to create link:', error);
      alert('Failed to create link');
    }
  };

  const handleAddPolicy = async () => {
    if (!selectedNetwork) return;
    try {
      await createPolicy(selectedNetwork.id, {
        name: policyForm.name,
        type: policyForm.type,
        description: policyForm.description,
      });
      setShowPolicyDialog(false);
      setPolicyForm({ name: '', type: 'firewall', description: '' });
      fetchNetworkDetails();
    } catch (error) {
      console.error('Failed to create policy:', error);
      alert('Failed to create policy');
    }
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert severity="error" sx={{ mb: 2 }}>
        {error}
      </Alert>
    );
  }

  if (networks.length === 0) {
    return (
      <Card>
        <CardContent>
          <Typography color="text.secondary">No networks found</Typography>
        </CardContent>
      </Card>
    );
  }

  return (
    <Box>
      {/* Network Selection */}
      <Card sx={{ mb: 2 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>Select Network to Manage</Typography>
          <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap' }}>
            {networks.map((network) => (
              <Button
                key={network.id}
                variant={selectedNetwork?.id === network.id ? 'contained' : 'outlined'}
                onClick={() => setSelectedNetwork(network)}
                sx={{ textTransform: 'none' }}
              >
                {network.name}
                <Chip
                  label={network.status}
                  size="small"
                  color={network.status === 'active' ? 'success' : 'default'}
                  sx={{ ml: 1 }}
                />
              </Button>
            ))}
          </Box>
        </CardContent>
      </Card>

      {/* Network Details */}
      {selectedNetwork && (
        <Box>
          <Card sx={{ mb: 2 }}>
            <CardContent>
              <Typography variant="h6" gutterBottom>{selectedNetwork.name}</Typography>
              <Typography variant="body2" color="text.secondary" gutterBottom>
                {selectedNetwork.description}
              </Typography>
              <Box sx={{ mt: 2, display: 'flex', gap: 1 }}>
                <Chip label={`Status: ${selectedNetwork.status}`} size="small" />
                <Chip label={`Created: ${new Date(selectedNetwork.created_at).toLocaleDateString()}`} size="small" variant="outlined" />
              </Box>
            </CardContent>
          </Card>

          {/* Tabs for Resources */}
          <Card>
            <CardContent>
              <Tabs value={activeTab} onChange={(e, v) => setActiveTab(v)} sx={{ mb: 2 }}>
                <Tab icon={<NodeIcon />} label={`Nodes (${nodes.length})`} />
                <Tab icon={<LinkIcon />} label={`Links (${links.length})`} />
                <Tab icon={<PolicyIcon />} label={`Policies (${policies.length})`} />
              </Tabs>

              {/* Tab Content */}
              <Box sx={{ minHeight: 300 }}>
                {activeTab === 0 && (
                  <Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                      <Typography variant="subtitle1">Network Nodes</Typography>
                      <Button
                        variant="contained"
                        size="small"
                        startIcon={<AddIcon />}
                        onClick={() => setShowNodeDialog(true)}
                      >
                        Add Node
                      </Button>
                    </Box>
                    {loadingDetails ? (
                      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
                        <CircularProgress />
                      </Box>
                    ) : nodes.length === 0 ? (
                      <Typography color="text.secondary">No nodes in this network</Typography>
                    ) : (
                      <TableContainer>
                        <Table>
                          <TableHead>
                            <TableRow>
                              <TableCell>Name</TableCell>
                              <TableCell>Type</TableCell>
                              <TableCell>IP Address</TableCell>
                              <TableCell>Status</TableCell>
                              <TableCell>Actions</TableCell>
                            </TableRow>
                          </TableHead>
                          <TableBody>
                            {nodes.map((node) => (
                              <TableRow key={node.id}>
                                <TableCell>{node.name}</TableCell>
                                <TableCell>{node.type}</TableCell>
                                <TableCell>{node.ip_address}</TableCell>
                                <TableCell>
                                  <Chip
                                    label={node.status}
                                    size="small"
                                    color={node.status === 'active' ? 'success' : 'default'}
                                  />
                                </TableCell>
                                <TableCell>
                                  <IconButton size="small" color="error">
                                    <DeleteIcon />
                                  </IconButton>
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </TableContainer>
                    )}
                  </Box>
                )}

                {activeTab === 1 && (
                  <Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                      <Typography variant="subtitle1">Network Links</Typography>
                      <Button
                        variant="contained"
                        size="small"
                        startIcon={<AddIcon />}
                        onClick={() => setShowLinkDialog(true)}
                        disabled={nodes.length < 2}
                      >
                        Add Link
                      </Button>
                    </Box>
                    {links.length === 0 ? (
                      <Typography color="text.secondary">No links in this network</Typography>
                    ) : (
                      <TableContainer>
                        <Table>
                          <TableHead>
                            <TableRow>
                              <TableCell>Source</TableCell>
                              <TableCell>Target</TableCell>
                              <TableCell>Bandwidth (Mbps)</TableCell>
                              <TableCell>Status</TableCell>
                              <TableCell>Actions</TableCell>
                            </TableRow>
                          </TableHead>
                          <TableBody>
                            {links.map((link) => (
                              <TableRow key={link.id}>
                                <TableCell>{link.source_node_id}</TableCell>
                                <TableCell>{link.target_node_id}</TableCell>
                                <TableCell>{link.bandwidth_mbps || 'N/A'}</TableCell>
                                <TableCell>
                                  <Chip label={link.status} size="small" />
                                </TableCell>
                                <TableCell>
                                  <IconButton size="small" color="error">
                                    <DeleteIcon />
                                  </IconButton>
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </TableContainer>
                    )}
                  </Box>
                )}

                {activeTab === 2 && (
                  <Box>
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 2 }}>
                      <Typography variant="subtitle1">Network Policies</Typography>
                      <Button
                        variant="contained"
                        size="small"
                        startIcon={<AddIcon />}
                        onClick={() => setShowPolicyDialog(true)}
                      >
                        Add Policy
                      </Button>
                    </Box>
                    {policies.length === 0 ? (
                      <Typography color="text.secondary">No policies in this network</Typography>
                    ) : (
                      <TableContainer>
                        <Table>
                          <TableHead>
                            <TableRow>
                              <TableCell>Name</TableCell>
                              <TableCell>Type</TableCell>
                              <TableCell>Status</TableCell>
                              <TableCell>Actions</TableCell>
                            </TableRow>
                          </TableHead>
                          <TableBody>
                            {policies.map((policy) => (
                              <TableRow key={policy.id}>
                                <TableCell>{policy.name}</TableCell>
                                <TableCell>{policy.type}</TableCell>
                                <TableCell>
                                  <Chip label={policy.status} size="small" />
                                </TableCell>
                                <TableCell>
                                  <IconButton size="small" color="error">
                                    <DeleteIcon />
                                  </IconButton>
                                </TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      </TableContainer>
                    )}
                  </Box>
                )}
              </Box>
            </CardContent>
          </Card>
        </Box>
      )}

      {/* Add Node Dialog */}
      <Dialog open={showNodeDialog} onClose={() => setShowNodeDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Node to {selectedNetwork?.name}</DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label="Node Name"
            fullWidth
            value={nodeForm.name}
            onChange={(e) => setNodeForm({ ...nodeForm, name: e.target.value })}
            sx={{ mb: 2, mt: 1 }}
          />
          <FormControl fullWidth sx={{ mb: 2 }}>
            <InputLabel>Node Type</InputLabel>
            <Select
              value={nodeForm.type}
              label="Node Type"
              onChange={(e) => setNodeForm({ ...nodeForm, type: e.target.value })}
            >
              <MenuItem value="router">Router</MenuItem>
              <MenuItem value="switch">Switch</MenuItem>
              <MenuItem value="host">Host</MenuItem>
              <MenuItem value="firewall">Firewall</MenuItem>
            </Select>
          </FormControl>
          <TextField
            margin="dense"
            label="IP Address"
            fullWidth
            value={nodeForm.ip_address}
            onChange={(e) => setNodeForm({ ...nodeForm, ip_address: e.target.value })}
            placeholder="192.168.1.1"
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowNodeDialog(false)}>Cancel</Button>
          <Button onClick={handleAddNode} variant="contained" disabled={!nodeForm.name}>
            Create
          </Button>
        </DialogActions>
      </Dialog>

      {/* Add Link Dialog */}
      <Dialog open={showLinkDialog} onClose={() => setShowLinkDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Link to {selectedNetwork?.name}</DialogTitle>
        <DialogContent>
          <FormControl fullWidth sx={{ mb: 2, mt: 1 }}>
            <InputLabel>Source Node</InputLabel>
            <Select
              value={linkForm.source_node_id}
              label="Source Node"
              onChange={(e) => setLinkForm({ ...linkForm, source_node_id: e.target.value })}
            >
              {nodes.map((node) => (
                <MenuItem key={node.id} value={node.id}>
                  {node.name} ({node.ip_address})
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <FormControl fullWidth sx={{ mb: 2 }}>
            <InputLabel>Target Node</InputLabel>
            <Select
              value={linkForm.target_node_id}
              label="Target Node"
              onChange={(e) => setLinkForm({ ...linkForm, target_node_id: e.target.value })}
            >
              {nodes.map((node) => (
                <MenuItem key={node.id} value={node.id}>
                  {node.name} ({node.ip_address})
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <TextField
            margin="dense"
            label="Bandwidth (Mbps)"
            fullWidth
            type="number"
            value={linkForm.bandwidth}
            onChange={(e) => setLinkForm({ ...linkForm, bandwidth: e.target.value })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowLinkDialog(false)}>Cancel</Button>
          <Button onClick={handleAddLink} variant="contained" disabled={!linkForm.source_node_id || !linkForm.target_node_id}>
            Create
          </Button>
        </DialogActions>
      </Dialog>

      {/* Add Policy Dialog */}
      <Dialog open={showPolicyDialog} onClose={() => setShowPolicyDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Add Policy to {selectedNetwork?.name}</DialogTitle>
        <DialogContent>
          <TextField
            margin="dense"
            label="Policy Name"
            fullWidth
            value={policyForm.name}
            onChange={(e) => setPolicyForm({ ...policyForm, name: e.target.value })}
            sx={{ mb: 2, mt: 1 }}
          />
          <FormControl fullWidth sx={{ mb: 2 }}>
            <InputLabel>Policy Type</InputLabel>
            <Select
              value={policyForm.type}
              label="Policy Type"
              onChange={(e) => setPolicyForm({ ...policyForm, type: e.target.value })}
            >
              <MenuItem value="firewall">Firewall</MenuItem>
              <MenuItem value="routing">Routing</MenuItem>
              <MenuItem value="qos">Quality of Service</MenuItem>
              <MenuItem value="security">Security</MenuItem>
            </Select>
          </FormControl>
          <TextField
            margin="dense"
            label="Description"
            fullWidth
            multiline
            rows={3}
            value={policyForm.description}
            onChange={(e) => setPolicyForm({ ...policyForm, description: e.target.value })}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setShowPolicyDialog(false)}>Cancel</Button>
          <Button onClick={handleAddPolicy} variant="contained" disabled={!policyForm.name}>
            Create
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default NetworkOverview;
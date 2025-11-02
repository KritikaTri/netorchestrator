import React, { useState, useEffect } from 'react';
import {
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Box,
  Chip,
  CircularProgress,
  SelectChangeEvent,
} from '@mui/material';
import { NetworkCheck as NetworkIcon } from '@mui/icons-material';
import { useAuth } from '../contexts/AuthContext';

interface Network {
  id: string;
  name: string;
  status: string;
  description: string;
}

interface NetworkSelectorProps {
  selectedNetworkId: string;
  onNetworkChange: (networkId: string) => void;
}

const NetworkSelector: React.FC<NetworkSelectorProps> = ({
  selectedNetworkId,
  onNetworkChange,
}) => {
  const { api } = useAuth();
  const [networks, setNetworks] = useState<Network[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchNetworks();
  }, []);

  const fetchNetworks = async () => {
    try {
      setLoading(true);
      const response = await api.get('/api/v1/networks');
      const networkList = response.data.networks || [];
      setNetworks(networkList);

      // Auto-select first network if none selected
      if (networkList.length > 0 && !selectedNetworkId) {
        onNetworkChange(networkList[0].id);
      }
    } catch (error) {
      console.error('Failed to fetch networks:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (event: SelectChangeEvent<string>) => {
    onNetworkChange(event.target.value);
  };

  const selectedNetwork = networks.find(n => n.id === selectedNetworkId);

  if (loading) {
    return (
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <CircularProgress size={20} />
        <span>Loading networks...</span>
      </Box>
    );
  }

  if (networks.length === 0) {
    return (
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
        <NetworkIcon />
        <span>No networks available</span>
      </Box>
    );
  }

  return (
    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
      <NetworkIcon color="primary" />
      <FormControl sx={{ minWidth: 250 }} size="small">
        <InputLabel id="network-selector-label">Select Network</InputLabel>
        <Select
          labelId="network-selector-label"
          id="network-selector"
          value={selectedNetworkId}
          label="Select Network"
          onChange={handleChange}
        >
          {networks.map((network) => (
            <MenuItem key={network.id} value={network.id}>
              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%' }}>
                <span>{network.name || 'Unnamed Network'}</span>
                <Chip
                  label={network.status}
                  size="small"
                  color={network.status === 'active' ? 'success' : 'default'}
                  sx={{ ml: 'auto' }}
                />
              </Box>
            </MenuItem>
          ))}
        </Select>
      </FormControl>
      {selectedNetwork && selectedNetwork.description && (
        <Box sx={{ color: 'text.secondary', fontSize: '0.875rem' }}>
          {selectedNetwork.description}
        </Box>
      )}
    </Box>
  );
};

export default NetworkSelector;

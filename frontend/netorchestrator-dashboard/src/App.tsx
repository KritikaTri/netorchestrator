import React from 'react';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import { CssBaseline, Box } from '@mui/material';
import Dashboard from './components/Dashboard';
import { ApiProvider } from './contexts/ApiContext';
import { WebSocketProvider } from './contexts/WebSocketContext';

const darkTheme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
    background: {
      default: '#0a0e27',
      paper: '#1a1d36',
    },
  },
  typography: {
    fontFamily: '"Roboto", "Helvetica", "Arial", sans-serif',
    h1: {
      fontSize: '2.5rem',
      fontWeight: 300,
    },
    h2: {
      fontSize: '2rem',
      fontWeight: 300,
    },
  },
  components: {
    MuiCard: {
      styleOverrides: {
        root: {
          background: 'linear-gradient(135deg, rgba(26, 29, 54, 0.9) 0%, rgba(26, 29, 54, 0.7) 100%)',
          backdropFilter: 'blur(10px)',
          border: '1px solid rgba(255, 255, 255, 0.1)',
        },
      },
    },
  },
});

function App() {
  return (
    <ThemeProvider theme={darkTheme}>
      <CssBaseline />
      <ApiProvider>
        <WebSocketProvider>
          <Box sx={{ 
            minHeight: '100vh',
            background: 'linear-gradient(135deg, #0a0e27 0%, #1a1d36 50%, #2d1b69 100%)',
            backgroundAttachment: 'fixed',
          }}>
            <Dashboard />
          </Box>
        </WebSocketProvider>
      </ApiProvider>
    </ThemeProvider>
  );
}

export default App;

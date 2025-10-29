import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';

interface WebSocketContextType {
  socket: WebSocket | null;
  connected: boolean;
  lastMessage: any;
  sendMessage: (data: any) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

const WS_URL = process.env.REACT_APP_WS_URL || 'ws://localhost:8080/ws';

export const WebSocketProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<any>(null);

  useEffect(() => {
    // Initialize WebSocket connection
    const newSocket = new WebSocket(WS_URL);

    newSocket.onopen = () => {
      console.log('WebSocket connected');
      setConnected(true);
    };

    newSocket.onclose = () => {
      console.log('WebSocket disconnected');
      setConnected(false);
    };

    newSocket.onerror = (error) => {
      console.error('WebSocket error:', error);
      setConnected(false);
    };

    newSocket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data);
        console.log('WebSocket message received:', message);
        setLastMessage(message);
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    };

    setSocket(newSocket);

    return () => {
      if (newSocket.readyState === WebSocket.OPEN) {
        newSocket.close();
      }
    };
  }, []);

  const sendMessage = (data: any) => {
    if (socket && connected && socket.readyState === WebSocket.OPEN) {
      socket.send(JSON.stringify(data));
    } else {
      console.warn('WebSocket not connected, cannot send message');
    }
  };

  const value: WebSocketContextType = {
    socket,
    connected,
    lastMessage,
    sendMessage,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
};

export const useWebSocket = (): WebSocketContextType => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};
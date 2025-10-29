import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { io, Socket } from 'socket.io-client';

interface WebSocketContextType {
  socket: Socket | null;
  connected: boolean;
  lastMessage: any;
  sendMessage: (event: string, data: any) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(undefined);

const WS_URL = process.env.REACT_APP_WS_URL || 'http://localhost:8080';

export const WebSocketProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [socket, setSocket] = useState<Socket | null>(null);
  const [connected, setConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<any>(null);

  useEffect(() => {
    // Initialize socket connection
    const newSocket = io(WS_URL, {
      transports: ['websocket', 'polling'],
      autoConnect: true,
    });

    newSocket.on('connect', () => {
      console.log('WebSocket connected');
      setConnected(true);
    });

    newSocket.on('disconnect', () => {
      console.log('WebSocket disconnected');
      setConnected(false);
    });

    newSocket.on('network_update', (data) => {
      console.log('Network update received:', data);
      setLastMessage({ type: 'network_update', data });
    });

    newSocket.on('metric_update', (data) => {
      console.log('Metric update received:', data);
      setLastMessage({ type: 'metric_update', data });
    });

    newSocket.on('alert', (data) => {
      console.log('Alert received:', data);
      setLastMessage({ type: 'alert', data });
    });

    newSocket.on('node_status', (data) => {
      console.log('Node status update:', data);
      setLastMessage({ type: 'node_status', data });
    });

    setSocket(newSocket);

    return () => {
      newSocket.close();
    };
  }, []);

  const sendMessage = (event: string, data: any) => {
    if (socket && connected) {
      socket.emit(event, data);
    } else {
      console.warn('Socket not connected, cannot send message');
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
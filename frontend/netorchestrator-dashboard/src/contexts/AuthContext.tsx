import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import axios, { AxiosInstance } from 'axios';

interface User {
  id: string;
  username: string;
  email: string;
  roles: string[];
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  error: string | null;
  login: (username: string, password: string) => Promise<boolean>;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
  isAuthenticated: boolean;
  api: AxiosInstance;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080';

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(localStorage.getItem('auth_token'));
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Create axios instance with interceptors
  const api = axios.create({
    baseURL: API_BASE_URL,
    timeout: 10000,
  });

  // Request interceptor to add auth token
  api.interceptors.request.use(
    (config) => {
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // Response interceptor to handle auth errors
  api.interceptors.response.use(
    (response) => response,
    async (error) => {
      if (error.response?.status === 401 && token) {
        // Try to refresh token
        const refreshed = await refreshToken();
        if (!refreshed) {
          logout();
        }
      }
      return Promise.reject(error);
    }
  );

  const login = async (username: string, password: string): Promise<boolean> => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await api.post('/api/v1/auth/login', {
        username,
        password,
      });

      const { access_token, user: userData } = response.data;
      
      setToken(access_token);
      setUser(userData);
      localStorage.setItem('auth_token', access_token);
      
      return true;
    } catch (error: any) {
      const message = error.response?.data?.error || 'Login failed';
      setError(message);
      return false;
    } finally {
      setLoading(false);
    }
  };

  const logout = () => {
    setUser(null);
    setToken(null);
    setError(null);
    localStorage.removeItem('auth_token');
  };

  const refreshToken = async (): Promise<boolean> => {
    try {
      const response = await api.post('/api/v1/auth/refresh');
      const { access_token } = response.data;
      
      setToken(access_token);
      localStorage.setItem('auth_token', access_token);
      
      return true;
    } catch (error) {
      return false;
    }
  };

  // Check if user is authenticated on app load
  useEffect(() => {
    if (token && !user) {
      // Validate token and get user info
      api.get('/api/v1/auth/profile')
        .then(response => {
          setUser(response.data.user);
        })
        .catch(() => {
          logout();
        });
    }
  }, [token, user]);

  const value: AuthContextType = {
    user,
    token,
    loading,
    error,
    login,
    logout,
    refreshToken,
    isAuthenticated: !!token && !!user,
    api,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
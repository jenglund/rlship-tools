import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { AuthContextType, User } from '../types/auth';

// Create a mock set of dev users for development
const DEV_USERS: Record<string, User> = {
  'dev_user1@gonetribal.com': {
    id: 'dev_user1',
    email: 'dev_user1@gonetribal.com',
    displayName: 'Dev User 1',
  },
  'dev_user2@gonetribal.com': {
    id: 'dev_user2',
    email: 'dev_user2@gonetribal.com',
    displayName: 'Dev User 2',
  },
  'dev_user3@gonetribal.com': {
    id: 'dev_user3',
    email: 'dev_user3@gonetribal.com',
    displayName: 'Dev User 3',
  },
  'dev_user4@gonetribal.com': {
    id: 'dev_user4',
    email: 'dev_user4@gonetribal.com',
    displayName: 'Dev User 4',
  },
};

// Default context values
const defaultAuthContext: AuthContextType = {
  isAuthenticated: false,
  user: null,
  isLoading: true,
  login: async () => {},
  logout: async () => {},
};

// Create the context
const AuthContext = createContext<AuthContextType>(defaultAuthContext);

// Custom hook to use the auth context
export const useAuth = () => useContext(AuthContext);

// Provider component
export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [authState, setAuthState] = useState({
    isAuthenticated: false,
    user: null as User | null,
    isLoading: true,
  });

  // Check for existing login on mount
  useEffect(() => {
    const checkLoginStatus = async () => {
      // In a real app, we'd check local storage / async storage here
      // For now, we'll just set isLoading to false
      setAuthState(prevState => ({
        ...prevState,
        isLoading: false,
      }));
    };

    checkLoginStatus();
  }, []);

  // Login function
  const login = async (email: string): Promise<void> => {
    try {
      // In a real app, we would make an API call here
      // For now, we'll just simulate a login with our mock data
      const user = DEV_USERS[email];
      
      if (!user) {
        throw new Error('User not found');
      }

      // Simulate a network delay
      await new Promise(resolve => setTimeout(resolve, 300));
      
      // Update the auth state
      setAuthState({
        isAuthenticated: true,
        user,
        isLoading: false,
      });

      // In a real app, we'd store the token/user info in local storage here
      
      return Promise.resolve();
    } catch (error) {
      console.error('Login error:', error);
      return Promise.reject(error);
    }
  };

  // Logout function
  const logout = async (): Promise<void> => {
    try {
      // In a real app, we would make an API call here
      // Simulate a network delay
      await new Promise(resolve => setTimeout(resolve, 300));
      
      // Update the auth state
      setAuthState({
        isAuthenticated: false,
        user: null,
        isLoading: false,
      });

      // In a real app, we'd clear the token/user info from local storage here
      
      return Promise.resolve();
    } catch (error) {
      console.error('Logout error:', error);
      return Promise.reject(error);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        ...authState,
        login,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}; 
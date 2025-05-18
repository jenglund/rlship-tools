import React, { createContext, useState, useContext, useEffect } from 'react';

// Dev users for testing
const DEV_USERS = [
  { id: 1, name: 'Test User 1', email: 'test1@example.com' },
  { id: 2, name: 'Test User 2', email: 'test2@example.com' },
  { id: 3, name: 'Test User 3', email: 'test3@example.com' },
  { id: 4, name: 'Test User 4', email: 'test4@example.com' },
];

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for existing user in localStorage
    const storedUser = localStorage.getItem('tribeUser');
    if (storedUser) {
      setCurrentUser(JSON.parse(storedUser));
    }
    setLoading(false);
  }, []);

  // Login with a dev user
  const login = (userId) => {
    const user = DEV_USERS.find(user => user.id === userId);
    if (user) {
      setCurrentUser(user);
      localStorage.setItem('tribeUser', JSON.stringify(user));
      return true;
    }
    return false;
  };

  // Logout
  const logout = () => {
    setCurrentUser(null);
    localStorage.removeItem('tribeUser');
  };

  const value = {
    currentUser,
    login,
    logout,
    devUsers: DEV_USERS,
    loading
  };

  return (
    <AuthContext.Provider value={value}>
      {!loading && children}
    </AuthContext.Provider>
  );
};

export default AuthContext; 
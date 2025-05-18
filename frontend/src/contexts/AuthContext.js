import React, { createContext, useState, useContext, useEffect } from 'react';
import authService from '../services/authService';

const AuthContext = createContext();

export const useAuth = () => useContext(AuthContext);

export const AuthProvider = ({ children }) => {
  const [currentUser, setCurrentUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    // Check for existing user in localStorage
    const user = authService.getCurrentUser();
    if (user) {
      setCurrentUser(user);
    }
    setLoading(false);
  }, []);

  // Check if email exists
  const checkEmailExists = async (email) => {
    try {
      setError(null);
      const response = await authService.checkEmail(email);
      return response.data.exists;
    } catch (err) {
      setError('Error checking email');
      throw err;
    }
  };

  // Register a new user
  const register = async (userData) => {
    try {
      setError(null);
      const user = await authService.register(userData);
      setCurrentUser(user);
      return user;
    } catch (err) {
      setError('Registration failed');
      throw err;
    }
  };

  // Login with email
  const login = async (email) => {
    try {
      setError(null);
      const user = await authService.login(email);
      setCurrentUser(user);
      return user;
    } catch (err) {
      setError('Login failed');
      throw err;
    }
  };

  // Logout
  const logout = () => {
    authService.logout();
    setCurrentUser(null);
  };

  // Update user profile
  const updateProfile = async (userData) => {
    try {
      setError(null);
      const updatedUser = await authService.updateProfile(userData);
      setCurrentUser(updatedUser);
      return updatedUser;
    } catch (err) {
      setError('Profile update failed');
      throw err;
    }
  };

  const value = {
    currentUser,
    login,
    logout,
    register,
    checkEmailExists,
    updateProfile,
    loading,
    error
  };

  return (
    <AuthContext.Provider value={value}>
      {!loading && children}
    </AuthContext.Provider>
  );
};

export default AuthContext; 
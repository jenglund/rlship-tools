import React from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const ProtectedRoute = ({ children }) => {
  const { currentUser } = useAuth();
  
  if (!currentUser) {
    // Redirect to home page if not logged in
    return <Navigate to="/" />;
  }

  return children;
};

export default ProtectedRoute; 
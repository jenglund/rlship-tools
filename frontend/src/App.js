import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import SplashScreen from './components/SplashScreen';
import LoginScreen from './components/LoginScreen';
import RegistrationScreen from './components/RegistrationScreen';
import TribesListScreen from './components/TribesListScreen';
import TribeDetailScreen from './components/TribeDetailScreen';
import ProtectedRoute from './components/ProtectedRoute';

function App() {
  return (
    <Router>
      <AuthProvider>
        <Routes>
          <Route path="/" element={<SplashScreen />} />
          <Route path="/login" element={<LoginScreen />} />
          <Route path="/register" element={<RegistrationScreen />} />
          <Route 
            path="/tribes" 
            element={
              <ProtectedRoute>
                <TribesListScreen />
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/tribes/:id" 
            element={
              <ProtectedRoute>
                <TribeDetailScreen />
              </ProtectedRoute>
            } 
          />
        </Routes>
      </AuthProvider>
    </Router>
  );
}

export default App; 
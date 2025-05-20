import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ListProvider } from './contexts/ListContext';
import SplashScreen from './components/SplashScreen';
import LoginScreen from './components/LoginScreen';
import RegistrationScreen from './components/RegistrationScreen';
import TribesListScreen from './components/TribesListScreen';
import TribeDetailScreen from './components/TribeDetailScreen';
import ProfileScreen from './components/ProfileScreen';
import AdminPanel from './components/AdminPanel';
import AdminTribeDetail from './components/AdminTribeDetail';
import Navigation from './components/Navigation';
import ProtectedRoute from './components/ProtectedRoute';
import TribeMemberProfileScreen from './components/TribeMemberProfileScreen';
import ListsScreen from './components/ListsScreen';
import ListDetailScreen from './components/ListDetailScreen';
import ListItemDetailScreen from './components/ListItemDetailScreen';

function App() {
  return (
    <Router>
      <AuthProvider>
        <ListProvider>
          <Navigation />
          <Routes>
            <Route path="/" element={<SplashScreen />} />
            <Route path="/login" element={<LoginScreen />} />
            <Route path="/register" element={<RegistrationScreen />} />
            <Route 
              path="/profile" 
              element={
                <ProtectedRoute>
                  <ProfileScreen />
                </ProtectedRoute>
              } 
            />
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
            <Route 
              path="/tribes/:tribeId/members/:memberId" 
              element={
                <ProtectedRoute>
                  <TribeMemberProfileScreen />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/lists" 
              element={
                <ProtectedRoute>
                  <ListsScreen />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/lists/:id" 
              element={
                <ProtectedRoute>
                  <ListDetailScreen />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/lists/:listID/items/:itemID" 
              element={
                <ProtectedRoute>
                  <ListItemDetailScreen />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/admin" 
              element={
                <ProtectedRoute>
                  <AdminPanel />
                </ProtectedRoute>
              } 
            />
            <Route 
              path="/admin/tribes/:tribeId" 
              element={
                <ProtectedRoute>
                  <AdminTribeDetail />
                </ProtectedRoute>
              } 
            />
          </Routes>
        </ListProvider>
      </AuthProvider>
    </Router>
  );
}

export default App; 
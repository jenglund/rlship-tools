import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Container, Form, Button, Alert } from 'react-bootstrap';

const RegistrationScreen = () => {
  const { currentUser, register } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  
  const [formData, setFormData] = useState({
    email: '',
    name: '',
    provider: 'email' // Default provider
  });
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Redirect if already logged in
  useEffect(() => {
    if (currentUser) {
      navigate('/tribes');
    }
  }, [currentUser, navigate]);

  // Get email from location state if available
  useEffect(() => {
    if (location.state && location.state.email) {
      setFormData(prev => ({ ...prev, email: location.state.email }));
    }
  }, [location.state]);

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // Validate form
    if (!formData.name.trim()) {
      setError('Display name is required');
      return;
    }

    if (!formData.email || !formData.email.includes('@')) {
      setError('Please enter a valid email address');
      return;
    }

    try {
      setLoading(true);
      setError('');
      
      // Add firebase_uid field for all users in development mode
      const userData = {...formData};
      userData.firebase_uid = `dev-${userData.email}`;
      console.debug(`Setting firebase_uid: dev-${userData.email}`);
      
      // Register the user
      await register(userData);
      
      // Redirect to tribes page on success
      navigate('/tribes');
    } catch (err) {
      console.error('Registration error:', err);
      setError('An error occurred during registration');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container className="py-5">
      <div className="registration-container mx-auto" style={{ maxWidth: '500px' }}>
        <h2 className="mb-4 text-center">Create Your Account</h2>
        {error && <Alert variant="danger">{error}</Alert>}
        
        <Form onSubmit={handleSubmit}>
          <Form.Group className="mb-3">
            <Form.Label htmlFor="email">Email Address</Form.Label>
            <Form.Control
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              placeholder="Enter your email"
              required
              readOnly={!!location.state?.email}
            />
            <Form.Text className="text-muted">
              Enter any email address to register (development mode)
            </Form.Text>
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Label htmlFor="name">Display Name</Form.Label>
            <Form.Control
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              placeholder="Enter your display name"
              required
            />
          </Form.Group>
          
          <Button 
            type="submit" 
            variant="primary"
            className="w-100"
            disabled={loading}
          >
            {loading ? 'Creating Account...' : 'Create Account'}
          </Button>
        </Form>
        
        <div className="mt-3 text-center">
          <p>Already have an account? <Link to="/login">Log in</Link></p>
        </div>
      </div>
    </Container>
  );
};

export default RegistrationScreen; 
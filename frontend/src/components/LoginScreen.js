import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Container, Form, Button, Alert } from 'react-bootstrap';

const LoginScreen = () => {
  const { currentUser, login } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Redirect if already logged in
  useEffect(() => {
    if (currentUser) {
      navigate('/tribes');
    }
  }, [currentUser, navigate]);

  const handleEmailChange = (e) => {
    setEmail(e.target.value);
    setError('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email || !email.includes('@')) {
      setError('Please enter a valid email address');
      return;
    }

    try {
      setLoading(true);
      setError('');
      
      // Login directly without checking if email exists for development purposes
      await login(email);
      navigate('/tribes');
    } catch (err) {
      console.error('Login error:', err);
      setError('An error occurred during login');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container className="py-5">
      <div className="login-container mx-auto" style={{ maxWidth: '500px' }}>
        <h2 className="mb-4 text-center">Log in to Tribe</h2>
        <p className="text-muted text-center mb-4">Development Mode</p>
        
        {error && <Alert variant="danger">{error}</Alert>}
        
        <Form onSubmit={handleSubmit}>
          <Form.Group className="mb-3">
            <Form.Label htmlFor="email">Email Address</Form.Label>
            <Form.Control
              type="email"
              id="email"
              value={email}
              onChange={handleEmailChange}
              placeholder="Enter your email"
              required
            />
            <Form.Text className="text-muted">
              Enter any email address to log in (development mode)
            </Form.Text>
          </Form.Group>
          
          <Button 
            type="submit" 
            variant="primary"
            className="w-100"
            disabled={loading}
          >
            {loading ? 'Processing...' : 'Login'}
          </Button>
        </Form>
      </div>
    </Container>
  );
};

export default LoginScreen; 
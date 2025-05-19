import React, { useEffect, useState } from 'react';
import { Container, Card, Row, Col, Image, ListGroup } from 'react-bootstrap';
import { useAuth } from '../contexts/AuthContext';
import axios from 'axios';
import config from '../config';

const ProfileScreen = () => {
  const { currentUser } = useAuth();
  const [userTribes, setUserTribes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchUserTribes = async () => {
      try {
        const response = await axios.get(`${config.API_URL}/tribes/my`);
        setUserTribes(response.data.data || []);
        setLoading(false);
      } catch (err) {
        console.error('Error fetching user tribes:', err);
        setError('Failed to load your tribes. Please try again later.');
        setLoading(false);
      }
    };

    if (currentUser) {
      fetchUserTribes();
    } else {
      setLoading(false);
    }
  }, [currentUser]);

  if (!currentUser) {
    return (
      <Container className="mt-5">
        <h2>User Profile</h2>
        <div className="text-center mt-5">
          <p>Please log in to view your profile.</p>
        </div>
      </Container>
    );
  }

  if (loading) {
    return (
      <Container className="mt-5">
        <h2>User Profile</h2>
        <div className="text-center mt-5">
          <p>Loading your profile...</p>
        </div>
      </Container>
    );
  }

  return (
    <Container className="mt-5">
      <h2>My Profile</h2>
      
      <Card className="mt-4">
        <Card.Body>
          <Row>
            <Col md={3} className="text-center mb-3 mb-md-0">
              {currentUser.avatar_url ? (
                <Image 
                  src={currentUser.avatar_url} 
                  roundedCircle 
                  style={{ width: '150px', height: '150px', objectFit: 'cover' }} 
                />
              ) : (
                <div 
                  className="bg-secondary rounded-circle d-flex align-items-center justify-content-center"
                  style={{ width: '150px', height: '150px', margin: '0 auto' }}
                >
                  <span className="text-white" style={{ fontSize: '50px' }}>
                    {currentUser.name?.charAt(0) || '?'}
                  </span>
                </div>
              )}
            </Col>
            <Col md={9}>
              <h3>{currentUser.name}</h3>
              <p className="text-muted">{currentUser.email}</p>
              <p><strong>User ID:</strong> {currentUser.id}</p>
            </Col>
          </Row>
        </Card.Body>
      </Card>

      <Card className="mt-4">
        <Card.Header as="h5">My Tribes</Card.Header>
        <Card.Body>
          {error && <p className="text-danger">{error}</p>}
          
          {userTribes.length === 0 ? (
            <p>You are not a member of any tribes yet.</p>
          ) : (
            <ListGroup>
              {userTribes.map(tribe => (
                <ListGroup.Item key={tribe.id} className="d-flex justify-content-between align-items-center">
                  <div>
                    <h5>{tribe.name}</h5>
                    <small className="text-muted">Type: {tribe.type}</small>
                    {tribe.description && <p className="mb-0">{tribe.description}</p>}
                  </div>
                  <div>
                    <small className="text-muted">
                      Display Name: {currentUser.name}
                    </small>
                  </div>
                </ListGroup.Item>
              ))}
            </ListGroup>
          )}
        </Card.Body>
      </Card>
    </Container>
  );
};

export default ProfileScreen; 
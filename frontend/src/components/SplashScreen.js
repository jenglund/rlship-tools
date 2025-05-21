import React from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { Button, Container, Row, Col } from 'react-bootstrap';

const SplashScreen = () => {
  const { currentUser } = useAuth();
  const navigate = useNavigate();

  const handleGoToTribes = () => {
    navigate('/tribes');
  };

  return (
    <Container className="splash-screen py-5">
      <Row className="justify-content-center">
        <Col md={8} className="text-center">
          <h1 className="mb-4">Welcome to Tribe</h1>
          <p className="mb-4">
            A place for groups to share and plan activities together. Connect with your tribes and organize your shared experiences.
          </p>
          
          {currentUser ? (
            <div>
              <p className="mb-3">You are logged in as <strong>{currentUser.name || currentUser.email}</strong></p>
              <Button variant="primary" onClick={handleGoToTribes}>
                Go to My Tribes
              </Button>
            </div>
          ) : (
            <div className="d-flex justify-content-center gap-3">
              <Link to="/login">
                <Button variant="primary">Log In</Button>
              </Link>
              <Link to="/register">
                <Button variant="outline-primary">Register</Button>
              </Link>
            </div>
          )}
        </Col>
      </Row>
    </Container>
  );
};

export default SplashScreen; 
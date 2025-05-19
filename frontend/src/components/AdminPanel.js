import React, { useState, useEffect } from 'react';
import { Container, Card, Table, Button, Alert } from 'react-bootstrap';
import { Link, Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { isAdmin } from '../utils/adminUtils';
import adminService from '../services/adminService';
import { formatDateString } from '../utils/dateUtils';

const AdminPanel = () => {
  const { currentUser } = useAuth();
  const [tribes, setTribes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchTribes = async () => {
      try {
        const data = await adminService.getRecentTribes();
        setTribes(data);
        setLoading(false);
      } catch (err) {
        console.error('Error fetching tribes:', err);
        setError('Failed to load tribes. Please try again later.');
        setLoading(false);
      }
    };

    if (currentUser && isAdmin(currentUser)) {
      fetchTribes();
    } else {
      setLoading(false);
    }
  }, [currentUser]);

  // Redirect non-admin users
  if (!currentUser || !isAdmin(currentUser)) {
    return <Navigate to="/" />;
  }

  if (loading) {
    return (
      <Container className="mt-5">
        <h2>Admin Panel</h2>
        <div className="text-center mt-5">
          <p>Loading data...</p>
        </div>
      </Container>
    );
  }

  return (
    <Container className="mt-5">
      <h2>Admin Panel</h2>
      <p className="text-muted">Access granted to {currentUser.email}</p>
      
      <Card className="mt-4">
        <Card.Header as="h5">Recent Tribes</Card.Header>
        <Card.Body>
          {error && <Alert variant="danger">{error}</Alert>}
          
          {tribes.length === 0 ? (
            <p>No tribes found.</p>
          ) : (
            <Table responsive striped hover>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Type</th>
                  <th>Created At</th>
                  <th>Visibility</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {tribes.map(tribe => (
                  <tr key={tribe.id}>
                    <td>{tribe.name}</td>
                    <td>{tribe.type}</td>
                    <td>{formatDateString(tribe.created_at)}</td>
                    <td>{tribe.visibility}</td>
                    <td>
                      <Link to={`/admin/tribes/${tribe.id}`}>
                        <Button variant="primary" size="sm">View Details</Button>
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          )}
        </Card.Body>
      </Card>
    </Container>
  );
};

export default AdminPanel; 
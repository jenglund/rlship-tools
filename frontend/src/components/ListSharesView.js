import React, { useState, useEffect } from 'react';
import { Card, ListGroup, Button, Spinner, Alert } from 'react-bootstrap';
import { useList } from '../contexts/ListContext';
import listService from '../services/listService';

const ListSharesView = ({ listId }) => {
  const [shares, setShares] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  const { getListShares } = useList();
  
  useEffect(() => {
    const fetchListShares = async () => {
      try {
        setLoading(true);
        setError(null);
        
        // Use the context method for mock data instead of direct service call
        const data = await getListShares(listId);
        setShares(data || []);
      } catch (err) {
        console.error('Error fetching list shares:', err);
        setError('Failed to load list shares');
      } finally {
        setLoading(false);
      }
    };
    
    fetchListShares();
  }, [listId, getListShares]);
  
  const handleUnshare = async (tribeID) => {
    try {
      setError(null);
      await listService.unshareListWithTribe(listId, tribeID);
      
      // Update the local state
      setShares(prev => prev.filter(tribe => tribe.id !== tribeID));
    } catch (err) {
      console.error('Error unsharing list:', err);
      setError('Failed to unshare list');
    }
  };
  
  if (loading) {
    return (
      <div className="text-center my-3">
        <Spinner animation="border" size="sm" />
      </div>
    );
  }
  
  if (error) {
    return <Alert variant="danger">{error}</Alert>;
  }
  
  if (shares.length === 0) {
    return (
      <Card className="mb-4">
        <Card.Header>Shared With</Card.Header>
        <Card.Body>
          <p className="text-muted mb-0">This list is not shared with any tribes.</p>
        </Card.Body>
      </Card>
    );
  }
  
  return (
    <Card className="mb-4">
      <Card.Header>Shared With</Card.Header>
      <ListGroup variant="flush">
        {shares.map(tribe => (
          <ListGroup.Item key={tribe.id} className="d-flex justify-content-between align-items-center">
            <div>
              <strong>{tribe.name}</strong>
              {tribe.description && <p className="text-muted mb-0">{tribe.description}</p>}
            </div>
            <Button 
              variant="outline-danger" 
              size="sm" 
              onClick={() => handleUnshare(tribe.id)}
            >
              Unshare
            </Button>
          </ListGroup.Item>
        ))}
      </ListGroup>
    </Card>
  );
};

export default ListSharesView; 
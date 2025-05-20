import React, { useState, useEffect } from 'react';
import { Card, ListGroup, Button, Spinner, Alert, Row, Col } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import listService from '../services/listService';
import ListCard from './ListCard';

const TribeSharedLists = ({ tribeId }) => {
  const [sharedLists, setSharedLists] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    fetchSharedLists();
  }, [tribeId]);
  
  const fetchSharedLists = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await listService.getSharedListsForTribe(tribeId);
      setSharedLists(data || []);
    } catch (err) {
      console.error('Error fetching shared lists:', err);
      setError('Failed to load shared lists');
    } finally {
      setLoading(false);
    }
  };
  
  if (loading) {
    return (
      <div className="text-center my-3">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Loading shared lists...</span>
        </Spinner>
      </div>
    );
  }
  
  if (error) {
    return <Alert variant="danger">{error}</Alert>;
  }
  
  if (sharedLists.length === 0) {
    return (
      <Card>
        <Card.Header>Shared Lists</Card.Header>
        <Card.Body>
          <p className="text-muted mb-0">No lists have been shared with this tribe yet.</p>
        </Card.Body>
      </Card>
    );
  }
  
  return (
    <Card>
      <Card.Header>Shared Lists</Card.Header>
      <Card.Body>
        <Row>
          {sharedLists.map(list => (
            <Col md={4} key={list.id} className="mb-3">
              <ListCard list={list} />
            </Col>
          ))}
        </Row>
      </Card.Body>
    </Card>
  );
};

export default TribeSharedLists; 
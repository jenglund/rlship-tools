import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Spinner, Alert } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import ListCard from './ListCard';
import { useList } from '../contexts/ListContext';

const UserListsView = ({ userId }) => {
  const { userLists, fetchUserLists, loading, error } = useList();
  
  useEffect(() => {
    if (userId) {
      fetchUserLists(userId);
    }
  }, [userId, fetchUserLists]);
  
  if (loading) {
    return (
      <div className="text-center my-3">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Loading your lists...</span>
        </Spinner>
      </div>
    );
  }
  
  if (error) {
    return <Alert variant="danger">{error}</Alert>;
  }
  
  if (userLists.length === 0) {
    return (
      <Card className="mt-4">
        <Card.Header as="h5">My Lists</Card.Header>
        <Card.Body>
          <p className="text-muted mb-0">You haven't created any lists yet.</p>
          <div className="mt-3">
            <Link to="/lists" className="btn btn-primary">
              Create Your First List
            </Link>
          </div>
        </Card.Body>
      </Card>
    );
  }
  
  return (
    <Card className="mt-4">
      <Card.Header as="h5">My Lists</Card.Header>
      <Card.Body>
        <Row>
          {userLists.map(list => (
            <Col md={4} key={list.id} className="mb-3">
              <ListCard list={list} />
            </Col>
          ))}
        </Row>
        <div className="mt-3 text-end">
          <Link to="/lists" className="btn btn-outline-primary">
            View All Lists
          </Link>
        </div>
      </Card.Body>
    </Card>
  );
};

export default UserListsView; 
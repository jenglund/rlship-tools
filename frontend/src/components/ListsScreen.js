import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Button, Spinner } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useList } from '../contexts/ListContext';
import ListCard from './ListCard';
import CreateListModal from './CreateListModal';

const ListsScreen = () => {
  const [showCreateModal, setShowCreateModal] = useState(false);
  const { currentUser } = useAuth();
  const { 
    userLists, 
    sharedLists, 
    loading, 
    error, 
    fetchUserLists, 
    fetchSharedLists,
    deleteList
  } = useList();
  
  const navigate = useNavigate();

  useEffect(() => {
    const loadLists = async () => {
      await fetchUserLists();
      await fetchSharedLists();
    };

    if (currentUser && currentUser.id) {
      loadLists();
    }
  }, [currentUser, fetchUserLists, fetchSharedLists]);

  const handleCreateList = () => {
    setShowCreateModal(true);
  };

  const handleCloseCreateModal = () => {
    setShowCreateModal(false);
  };

  const handleShareList = (list) => {
    // This will be implemented later
    console.log('Share list clicked', list.id);
  };

  const handleDeleteList = async (list) => {
    try {
      if (window.confirm(`Are you sure you want to delete the list "${list.name}"? This action cannot be undone.`)) {
        await deleteList(list.id);
      }
    } catch (err) {
      console.error('Error deleting list:', err);
    }
  };

  if (loading) {
    return (
      <Container className="mt-4 text-center">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Loading...</span>
        </Spinner>
      </Container>
    );
  }

  if (error) {
    return (
      <Container className="mt-4">
        <div className="alert alert-danger" role="alert">
          {error}
        </div>
      </Container>
    );
  }

  return (
    <Container className="mt-4">
      <Row className="mb-4">
        <Col>
          <h1>My Lists</h1>
        </Col>
        <Col className="text-end">
          <Button variant="success" onClick={handleCreateList}>Create New List</Button>
        </Col>
      </Row>

      <Row className="mb-4">
        <Col>
          <h2>My Lists</h2>
          {userLists.length === 0 ? (
            <p>You don't have any lists yet. Create one to get started!</p>
          ) : (
            userLists.map(list => (
              <ListCard 
                key={list.id} 
                list={list} 
                onShare={handleShareList}
                onDelete={handleDeleteList}
              />
            ))
          )}
        </Col>
      </Row>

      <Row>
        <Col>
          <h2>Shared With Me</h2>
          {sharedLists.length === 0 ? (
            <p>No lists have been shared with you yet.</p>
          ) : (
            sharedLists.map(list => (
              <ListCard 
                key={list.id} 
                list={{...list, shared: true}} 
                onShare={null}
                onDelete={null}
              />
            ))
          )}
        </Col>
      </Row>

      <CreateListModal 
        show={showCreateModal} 
        onHide={handleCloseCreateModal} 
      />
    </Container>
  );
};

export default ListsScreen; 
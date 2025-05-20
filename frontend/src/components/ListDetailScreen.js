import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Button, Spinner, Badge } from 'react-bootstrap';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useList } from '../contexts/ListContext';
import ListItemCard from './ListItemCard';
import CreateListItemModal from './CreateListItemModal';
import EditListItemModal from './EditListItemModal';
import ShareListModal from './ShareListModal';
import ListSharesView from './ListSharesView';

const ListDetailScreen = () => {
  const [showAddItemModal, setShowAddItemModal] = useState(false);
  const [showEditItemModal, setShowEditItemModal] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);
  const [selectedItem, setSelectedItem] = useState(null);
  
  const { id } = useParams();
  const { currentUser } = useAuth();
  const { 
    currentList, 
    currentListItems, 
    loading, 
    error, 
    fetchListDetails,
    deleteList,
    deleteListItem
  } = useList();
  
  const navigate = useNavigate();

  useEffect(() => {
    if (id) {
      fetchListDetails(id);
    }
  }, [id, fetchListDetails]);

  const handleAddItem = () => {
    setShowAddItemModal(true);
  };

  const handleCloseAddItemModal = () => {
    setShowAddItemModal(false);
  };

  const handleEditList = () => {
    // This will be implemented later
    console.log('Edit list clicked');
  };

  const handleShareList = () => {
    setShowShareModal(true);
  };

  const handleCloseShareModal = () => {
    setShowShareModal(false);
    // Refresh list details to get updated shares
    fetchListDetails(id);
  };

  const handleDeleteList = async () => {
    try {
      if (window.confirm('Are you sure you want to delete this list? This action cannot be undone.')) {
        await deleteList(id);
        navigate('/lists');
      }
    } catch (err) {
      console.error('Error deleting list:', err);
    }
  };

  const handleEditItem = (item) => {
    setSelectedItem(item);
    setShowEditItemModal(true);
  };

  const handleCloseEditItemModal = () => {
    setShowEditItemModal(false);
    setSelectedItem(null);
  };

  const handleDeleteItem = async (item) => {
    try {
      if (window.confirm(`Are you sure you want to delete "${item.name}"? This action cannot be undone.`)) {
        await deleteListItem(id, item.id);
      }
    } catch (err) {
      console.error('Error deleting item:', err);
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
      {currentList && (
        <>
          <Row className="mb-4">
            <Col>
              <h1>{currentList.name}</h1>
              <p className="text-muted">{currentList.description}</p>
              <Badge bg="info">{currentList.type}</Badge>
              <Badge bg={currentList.visibility === 'private' ? 'secondary' : 'success'} className="ms-2">
                {currentList.visibility}
              </Badge>
            </Col>
            <Col className="text-end">
              <Button variant="primary" className="me-2" onClick={handleEditList}>
                Edit List
              </Button>
              <Button variant="success" className="me-2" onClick={handleShareList}>
                Share List
              </Button>
              <Button variant="danger" onClick={handleDeleteList}>
                Delete List
              </Button>
            </Col>
          </Row>

          {/* Sharing information */}
          <Row className="mb-4">
            <Col>
              <ListSharesView listId={id} />
            </Col>
          </Row>

          <Row className="mb-4">
            <Col>
              <h2>Items</h2>
            </Col>
            <Col className="text-end">
              <Button variant="success" onClick={handleAddItem}>
                Add Item
              </Button>
            </Col>
          </Row>

          <Row>
            {currentListItems.length === 0 ? (
              <Col>
                <p>This list doesn't have any items yet. Add some to get started!</p>
              </Col>
            ) : (
              currentListItems.map(item => (
                <Col md={4} key={item.id} className="mb-4">
                  <ListItemCard
                    item={{...item, type: currentList.type}}
                    onEdit={() => handleEditItem(item)}
                    onDelete={() => handleDeleteItem(item)}
                  />
                </Col>
              ))
            )}
          </Row>

          {currentList && (
            <>
              <CreateListItemModal
                show={showAddItemModal}
                onHide={handleCloseAddItemModal}
                listId={id}
                listType={currentList.type}
              />
              
              {selectedItem && (
                <EditListItemModal
                  show={showEditItemModal}
                  onHide={handleCloseEditItemModal}
                  listId={id}
                  listType={currentList.type}
                  item={selectedItem}
                />
              )}
              
              <ShareListModal
                show={showShareModal}
                onHide={handleCloseShareModal}
                listId={id}
              />
            </>
          )}
        </>
      )}
    </Container>
  );
};

export default ListDetailScreen; 
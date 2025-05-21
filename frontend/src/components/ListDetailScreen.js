import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Button, Spinner, Badge, ButtonGroup, Card, Placeholder, Nav } from 'react-bootstrap';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useList } from '../contexts/ListContext';
import ListItemCard from './ListItemCard';
import CreateListItemModal from './CreateListItemModal';
import EditListItemModal from './EditListItemModal';
import ShareListModal from './ShareListModal';
import ListSharesView from './ListSharesView';
import LocationMapView from './LocationMapView';
import GenerateSelectionModal from './GenerateSelectionModal';
import { FaList, FaMap, FaRandom } from 'react-icons/fa';

// Skeleton loader for list details
const ListDetailSkeleton = () => {
  return (
    <>
      <Row className="mb-4">
        <Col xs={12} lg={8} className="mb-3 mb-lg-0">
          <Placeholder as="h1" animation="glow">
            <Placeholder xs={6} />
          </Placeholder>
          <Placeholder as="p" animation="glow">
            <Placeholder xs={10} />
            <Placeholder xs={8} />
          </Placeholder>
          <Placeholder.Button variant="info" xs={2} />
          <Placeholder.Button variant="secondary" xs={2} className="ms-2" />
        </Col>
        <Col xs={12} lg={4} className="d-flex justify-content-lg-end align-items-start">
          <ButtonGroup className="w-100 w-lg-auto">
            <Placeholder.Button variant="primary" xs={4} />
            <Placeholder.Button variant="success" xs={4} />
            <Placeholder.Button variant="danger" xs={4} />
          </ButtonGroup>
        </Col>
      </Row>
      <Row className="mb-4">
        <Col xs={12}>
          <Card className="p-3">
            <Placeholder as="h5" animation="glow">
              <Placeholder xs={4} />
            </Placeholder>
            <Placeholder as="p" animation="glow">
              <Placeholder xs={7} />
            </Placeholder>
          </Card>
        </Col>
      </Row>
    </>
  );
};

// Skeleton loader for list items
const ListItemSkeleton = () => {
  return (
    <Card className="h-100 shadow-sm">
      <Card.Body className="d-flex flex-column">
        <Placeholder as={Card.Title} animation="glow">
          <Placeholder xs={8} />
        </Placeholder>
        <Placeholder as={Card.Text} animation="glow" className="flex-grow-1">
          <Placeholder xs={12} />
          <Placeholder xs={10} />
          <Placeholder xs={8} />
        </Placeholder>
        <div className="d-flex justify-content-end gap-2 mt-3">
          <Placeholder.Button variant="outline-primary" xs={3} />
          <Placeholder.Button variant="outline-danger" xs={3} />
        </div>
      </Card.Body>
    </Card>
  );
};

const ListDetailScreen = () => {
  const [showAddItemModal, setShowAddItemModal] = useState(false);
  const [showEditItemModal, setShowEditItemModal] = useState(false);
  const [showShareModal, setShowShareModal] = useState(false);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [selectedItem, setSelectedItem] = useState(null);
  const [viewMode, setViewMode] = useState('list'); // 'list' or 'map'
  
  const { id } = useParams();
  const { currentUser } = useAuth();
  const { 
    currentList, 
    currentListItems, 
    loading, 
    error, 
    fetchListDetails,
    deleteList,
    deleteListItem,
    refreshUserLists
  } = useList();
  
  const navigate = useNavigate();

  useEffect(() => {
    if (id) {
      fetchListDetails(id);
    }
  }, [id, fetchListDetails]);

  // Debug log to check currentListItems
  useEffect(() => {
    console.log('ListDetailScreen - currentListItems:', currentListItems);
    if (currentListItems && !Array.isArray(currentListItems)) {
      console.error('ListDetailScreen - currentListItems is not an array:', currentListItems);
    }
  }, [currentListItems]);

  // Reset view mode when list type changes
  useEffect(() => {
    if (currentList && currentList.type === 'location') {
      // Default to list view unless it's a location list
      setViewMode('list');
    } else {
      // Force list view for non-location lists
      setViewMode('list');
    }
  }, [currentList]);

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

  const handleBackToLists = () => {
    // Make sure user lists will be refreshed when navigating back
    refreshUserLists();
    navigate('/lists');
  };

  const handleDeleteList = async () => {
    try {
      if (window.confirm('Are you sure you want to delete this list? This action cannot be undone.')) {
        await deleteList(id);
        handleBackToLists();
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

  const handleMapItemClick = (item) => {
    setSelectedItem(item);
    setShowEditItemModal(true);
  };

  const handleGenerateSelection = () => {
    setShowGenerateModal(true);
  };

  const handleCloseGenerateModal = () => {
    setShowGenerateModal(false);
  };

  const renderItemSkeletons = () => {
    return Array(3).fill().map((_, index) => (
      <Col key={`item-skeleton-${index}`}>
        <ListItemSkeleton />
      </Col>
    ));
  };

  // Check if should show map view option
  const canShowMapView = currentList && currentList.type === 'location';

  if (error) {
    return (
      <Container className="mt-4">
        <div className="alert alert-danger" role="alert">
          <h4 className="alert-heading">Error Loading List</h4>
          <p>{error}</p>
          <hr />
          <div className="d-flex gap-2">
            <Button 
              variant="outline-danger" 
              onClick={() => fetchListDetails(id)}
            >
              Retry
            </Button>
            <Button 
              variant="outline-secondary" 
              onClick={handleBackToLists}
            >
              Back to Lists
            </Button>
          </div>
        </div>
      </Container>
    );
  }

  return (
    <Container className="mt-4">
      {loading ? (
        <ListDetailSkeleton />
      ) : currentList && (
        <>
          <Row className="mb-4">
            <Col xs={12} lg={8} className="mb-3 mb-lg-0">
              <h1>{currentList.name}</h1>
              <p className="text-muted">{currentList.description}</p>
              <div className="mb-2">
                <Badge bg="info">{currentList.type}</Badge>
                <Badge bg={currentList.visibility === 'private' ? 'secondary' : 'success'} className="ms-2">
                  {currentList.visibility}
                </Badge>
              </div>
            </Col>
            <Col xs={12} lg={4} className="d-flex justify-content-lg-end align-items-start">
              <ButtonGroup className="w-100 w-lg-auto">
                <Button variant="primary" onClick={handleEditList}>
                  Edit List
                </Button>
                <Button variant="success" onClick={handleShareList}>
                  Share List
                </Button>
                <Button variant="danger" onClick={handleDeleteList}>
                  Delete List
                </Button>
              </ButtonGroup>
            </Col>
          </Row>

          {/* Sharing information */}
          <Row className="mb-4">
            <Col xs={12}>
              <ListSharesView listId={id} />
            </Col>
          </Row>

          <Row className="mb-4 align-items-center">
            <Col xs={6} sm={3}>
              <h2>Items</h2>
            </Col>
            
            {/* View mode toggle for location lists */}
            {canShowMapView && (
              <Col xs={6} sm={3} className="text-center">
                <ButtonGroup className="w-100">
                  <Button 
                    variant={viewMode === 'list' ? 'primary' : 'outline-primary'} 
                    onClick={() => setViewMode('list')}
                  >
                    <FaList className="me-1" /> List
                  </Button>
                  <Button 
                    variant={viewMode === 'map' ? 'primary' : 'outline-primary'} 
                    onClick={() => setViewMode('map')}
                  >
                    <FaMap className="me-1" /> Map
                  </Button>
                </ButtonGroup>
              </Col>
            )}
            
            <Col xs={12} sm={6} className="text-sm-end mt-3 mt-sm-0 d-flex justify-content-end gap-2">
              {currentListItems && currentListItems.length > 0 && (
                <Button 
                  variant="outline-primary" 
                  onClick={handleGenerateSelection}
                >
                  <FaRandom className="me-1" /> Generate
                </Button>
              )}
              <Button variant="success" onClick={handleAddItem}>
                Add Item
              </Button>
            </Col>
          </Row>

          {/* Map view for location lists */}
          {canShowMapView && viewMode === 'map' && (
            <Row className="mb-4">
              <Col xs={12}>
                <LocationMapView 
                  items={Array.isArray(currentListItems) ? currentListItems : []} 
                  onItemClick={handleMapItemClick} 
                />
              </Col>
            </Row>
          )}

          {/* List view */}
          {viewMode === 'list' && (
            <Row className="row-cols-1 row-cols-md-2 row-cols-lg-3 g-4">
              {loading ? (
                renderItemSkeletons()
              ) : !Array.isArray(currentListItems) || currentListItems.length === 0 ? (
                <Col xs={12}>
                  <p>This list doesn't have any items yet. Add some to get started!</p>
                </Col>
              ) : (
                Array.isArray(currentListItems) && currentListItems.map(item => (
                  <Col key={item.id}>
                    <ListItemCard
                      item={{...item, type: currentList.type, listId: id}}
                      onEdit={() => handleEditItem(item)}
                      onDelete={() => handleDeleteItem(item)}
                    />
                  </Col>
                ))
              )}
            </Row>
          )}

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

              <GenerateSelectionModal
                show={showGenerateModal}
                onHide={handleCloseGenerateModal}
                listId={id}
                listItems={Array.isArray(currentListItems) ? currentListItems : []}
              />
            </>
          )}
        </>
      )}
    </Container>
  );
};

export default ListDetailScreen; 
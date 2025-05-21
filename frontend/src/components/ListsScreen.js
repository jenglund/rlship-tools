import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Button, Spinner, Card, Placeholder, Form, InputGroup, Badge, Dropdown } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useList } from '../contexts/ListContext';
import ListCard from './ListCard';
import CreateListModal from './CreateListModal';
import { FaSearch, FaFilter } from 'react-icons/fa';

// List type options
const LIST_TYPES = [
  { value: 'all', label: 'All Types' },
  { value: 'location', label: 'Locations' },
  { value: 'activity', label: 'Activities' },
  { value: 'media', label: 'Media' },
  { value: 'food', label: 'Food & Recipes' },
  { value: 'general', label: 'General' }
];

// Skeleton loader for list cards
const ListCardSkeleton = () => {
  return (
    <Card className="mb-3 h-100">
      <Card.Body className="d-flex flex-column">
        <Placeholder as={Card.Title} animation="glow">
          <Placeholder xs={8} />
        </Placeholder>
        <Placeholder as={Card.Text} animation="glow" className="mt-2 flex-grow-1">
          <Placeholder xs={12} />
          <Placeholder xs={8} />
        </Placeholder>
        <div className="d-flex justify-content-between align-items-center mt-3">
          <Placeholder.Button variant="primary" xs={2} />
          <Placeholder.Button variant="outline-primary" xs={2} />
        </div>
      </Card.Body>
    </Card>
  );
};

const ListsScreen = () => {
  // Modal state
  const [showCreateModal, setShowCreateModal] = useState(false);
  
  // Search and filters state
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedType, setSelectedType] = useState('all');
  const [showShared, setShowShared] = useState(true);
  const [showOwned, setShowOwned] = useState(true);
  
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

  // Add debug logging
  useEffect(() => {
    if (userLists) {
      console.log("Debug - userLists:", userLists);
      console.log("Debug - filteredUserLists:", filteredUserLists);
    }
    if (sharedLists) {
      console.log("Debug - sharedLists:", sharedLists);
      console.log("Debug - filteredSharedLists:", filteredSharedLists);
    }
  }, [userLists, sharedLists, filteredUserLists, filteredSharedLists]);

  // Filter functions
  const filterBySearch = (list) => {
    if (!searchTerm) return true;
    
    const term = searchTerm.toLowerCase();
    return (
      list.name.toLowerCase().includes(term) || 
      (list.description && list.description.toLowerCase().includes(term))
    );
  };

  const filterByType = (list) => {
    if (selectedType === 'all') return true;
    return list.type === selectedType;
  };

  // Apply filters to get filtered lists
  const filteredUserLists = Array.isArray(userLists)
    ? userLists.filter(list => filterBySearch(list) && filterByType(list))
    : [];
  
  const filteredSharedLists = Array.isArray(sharedLists)
    ? sharedLists.filter(list => filterBySearch(list) && filterByType(list))
    : [];

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

  const handleSearchChange = (e) => {
    setSearchTerm(e.target.value);
  };

  const handleTypeChange = (type) => {
    setSelectedType(type);
  };

  const clearFilters = () => {
    setSearchTerm('');
    setSelectedType('all');
    setShowShared(true);
    setShowOwned(true);
  };

  const renderSkeletonLoaders = (count) => {
    return Array(count).fill().map((_, index) => (
      <Col key={`skeleton-${index}`} xs={12} md={6} lg={4} className="mb-3">
        <ListCardSkeleton />
      </Col>
    ));
  };

  // Debug function to trace Row components
  const traceRowRender = (component, items) => {
    console.log(`Debug - Rendering Row with ${items?.length || 0} items:`, items);
    return component;
  };

  // Determine if any filters are active
  const hasActiveFilters = searchTerm || selectedType !== 'all' || !showShared || !showOwned;

  return (
    <Container className="mt-4">
      <Row className="mb-4 align-items-center">
        <Col xs={12} sm={6}>
          <h1>My Lists</h1>
        </Col>
        <Col xs={12} sm={6} className="text-sm-end mt-2 mt-sm-0">
          <Button variant="success" onClick={handleCreateList}>Create New List</Button>
        </Col>
      </Row>

      {/* Search and filter section */}
      <Row className="mb-4">
        <Col xs={12} md={6} className="mb-3 mb-md-0">
          <InputGroup>
            <InputGroup.Text>
              <FaSearch />
            </InputGroup.Text>
            <Form.Control
              type="text"
              placeholder="Search lists..."
              value={searchTerm}
              onChange={handleSearchChange}
            />
            {searchTerm && (
              <Button 
                variant="outline-secondary" 
                onClick={() => setSearchTerm('')}
              >
                Clear
              </Button>
            )}
          </InputGroup>
        </Col>
        <Col xs={12} md={6}>
          <div className="d-flex justify-content-md-end gap-2 flex-wrap">
            <Dropdown className="mb-2 mb-md-0">
              <Dropdown.Toggle variant="outline-secondary" id="dropdown-type">
                <FaFilter className="me-1" />
                {LIST_TYPES.find(t => t.value === selectedType)?.label || 'All Types'}
              </Dropdown.Toggle>

              <Dropdown.Menu>
                {LIST_TYPES.map(type => (
                  <Dropdown.Item 
                    key={type.value} 
                    active={selectedType === type.value}
                    onClick={() => handleTypeChange(type.value)}
                  >
                    {type.label}
                  </Dropdown.Item>
                ))}
              </Dropdown.Menu>
            </Dropdown>

            <Form.Check 
              type="switch"
              id="show-owned-switch"
              label="My Lists"
              checked={showOwned}
              onChange={() => setShowOwned(!showOwned)}
              className="me-2"
            />

            <Form.Check 
              type="switch"
              id="show-shared-switch"
              label="Shared"
              checked={showShared}
              onChange={() => setShowShared(!showShared)}
            />

            {hasActiveFilters && (
              <Button 
                variant="link" 
                className="text-decoration-none p-0 ms-2"
                onClick={clearFilters}
              >
                Reset All
              </Button>
            )}
          </div>
        </Col>
      </Row>

      {/* Filter status badges */}
      {hasActiveFilters && (
        <Row className="mb-3">
          <Col>
            <div className="d-flex flex-wrap gap-2 align-items-center">
              <small className="text-muted me-2">Active filters:</small>
              {searchTerm && (
                <Badge bg="info" className="d-flex align-items-center">
                  Search: {searchTerm}
                </Badge>
              )}
              {selectedType !== 'all' && (
                <Badge bg="info" className="d-flex align-items-center">
                  Type: {LIST_TYPES.find(t => t.value === selectedType)?.label}
                </Badge>
              )}
              {!showOwned && (
                <Badge bg="info" className="d-flex align-items-center">
                  Hidden: My Lists
                </Badge>
              )}
              {!showShared && (
                <Badge bg="info" className="d-flex align-items-center">
                  Hidden: Shared Lists
                </Badge>
              )}
            </div>
          </Col>
        </Row>
      )}

      {showOwned && (
        <Row className="mb-4">
          <Col xs={12}>
            <h2>My Lists</h2>
            {loading ? (
              <Row>
                {renderSkeletonLoaders(3)}
              </Row>
            ) : error ? (
              <div className="alert alert-danger" role="alert">
                <p>{error}</p>
                <Button 
                  variant="outline-danger" 
                  size="sm" 
                  onClick={() => fetchUserLists()}
                >
                  Retry
                </Button>
              </div>
            ) : filteredUserLists.length === 0 ? (
              searchTerm || selectedType !== 'all' ? (
                <p>No lists match your search criteria.</p>
              ) : (
                <p>You don't have any lists yet. Create one to get started!</p>
              )
            ) : (
              <Row>
                {filteredUserLists.map(list => {
                  // Ensure list has an ID for the key prop
                  const listId = list.id || `user-list-${Math.random()}`;
                  console.log(`Debug - Rendering user list:`, list);
                  
                  return (
                    <Col key={listId} xs={12} md={6} lg={4} className="mb-3">
                      <ListCard 
                        list={list} 
                        onShare={handleShareList}
                        onDelete={handleDeleteList}
                      />
                    </Col>
                  );
                })}
              </Row>
            )}
          </Col>
        </Row>
      )}

      {showShared && (
        <Row>
          <Col xs={12}>
            <h2>Shared With Me</h2>
            {loading ? (
              <Row>
                {renderSkeletonLoaders(2)}
              </Row>
            ) : error ? (
              <div className="alert alert-danger" role="alert">
                <p>{error}</p>
                <Button 
                  variant="outline-danger" 
                  size="sm" 
                  onClick={() => fetchSharedLists()}
                >
                  Retry
                </Button>
              </div>
            ) : filteredSharedLists.length === 0 ? (
              searchTerm || selectedType !== 'all' ? (
                <p>No shared lists match your search criteria.</p>
              ) : (
                <p>No lists have been shared with you yet.</p>
              )
            ) : (
              <Row>
                {filteredSharedLists.map(list => {
                  // Ensure list has an ID for the key prop
                  const listId = list.id || `shared-list-${Math.random()}`;
                  console.log(`Debug - Rendering shared list:`, list);
                  
                  return (
                    <Col key={listId} xs={12} md={6} lg={4} className="mb-3">
                      <ListCard 
                        list={{...list, shared: true}} 
                        onShare={null}
                        onDelete={null}
                      />
                    </Col>
                  );
                })}
              </Row>
            )}
          </Col>
        </Row>
      )}

      <CreateListModal 
        show={showCreateModal} 
        onHide={handleCloseCreateModal} 
      />
    </Container>
  );
};

export default ListsScreen; 
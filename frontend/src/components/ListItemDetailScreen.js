import React, { useState, useEffect } from 'react';
import { Container, Row, Col, Button, Card, Spinner, Badge } from 'react-bootstrap';
import { useParams, useNavigate } from 'react-router-dom';
import { useList } from '../contexts/ListContext';
import LocationMapView from './LocationMapView';

const ListItemDetailScreen = () => {
  const { listID, itemID } = useParams();
  const navigate = useNavigate();
  const { 
    currentList, 
    currentListItems, 
    loading, 
    error, 
    fetchListDetails, 
    updateListItem, 
    deleteListItem 
  } = useList();

  const [item, setItem] = useState(null);
  const [isEditing, setIsEditing] = useState(false);
  const [formData, setFormData] = useState({});

  useEffect(() => {
    // Fetch list details to get current list and items
    if (listID) {
      fetchListDetails(listID);
    }
  }, [listID, fetchListDetails]);

  // Find the current item when list items are loaded
  useEffect(() => {
    if (currentListItems && currentListItems.length > 0 && itemID) {
      const foundItem = currentListItems.find(item => item.id === itemID);
      if (foundItem) {
        setItem(foundItem);
        // Initialize form data with item values
        setFormData({
          name: foundItem.name,
          description: foundItem.description,
          ...foundItem.metadata
        });
      }
    }
  }, [currentListItems, itemID]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSave = async () => {
    try {
      // Extract metadata fields from formData
      const { name, description, ...metadata } = formData;
      
      const updatedItem = await updateListItem(listID, itemID, { 
        name, 
        description, 
        metadata 
      });
      
      setItem(updatedItem);
      setIsEditing(false);
    } catch (err) {
      console.error('Error updating item:', err);
    }
  };

  const handleDelete = async () => {
    try {
      if (window.confirm(`Are you sure you want to delete "${item.name}"? This action cannot be undone.`)) {
        await deleteListItem(listID, itemID);
        // Navigate back to list detail screen
        navigate(`/lists/${listID}`);
      }
    } catch (err) {
      console.error('Error deleting item:', err);
    }
  };

  const handleCancel = () => {
    // Reset form data to original values
    if (item) {
      setFormData({
        name: item.name,
        description: item.description,
        ...item.metadata
      });
    }
    setIsEditing(false);
  };

  const renderEditForm = () => {
    return (
      <Card className="mb-4">
        <Card.Body>
          <div className="mb-3">
            <label htmlFor="name" className="form-label">Name</label>
            <input
              type="text"
              className="form-control"
              id="name"
              name="name"
              value={formData.name || ''}
              onChange={handleInputChange}
            />
          </div>
          <div className="mb-3">
            <label htmlFor="description" className="form-label">Description</label>
            <textarea
              className="form-control"
              id="description"
              name="description"
              rows="3"
              value={formData.description || ''}
              onChange={handleInputChange}
            />
          </div>

          {/* Render location-specific fields if list type is location */}
          {currentList && currentList.type === 'location' && (
            <>
              <div className="mb-3">
                <label htmlFor="address" className="form-label">Address</label>
                <input
                  type="text"
                  className="form-control"
                  id="address"
                  name="address"
                  value={formData.address || ''}
                  onChange={handleInputChange}
                />
              </div>
              <div className="row mb-3">
                <div className="col">
                  <label htmlFor="latitude" className="form-label">Latitude</label>
                  <input
                    type="number"
                    step="0.000001"
                    className="form-control"
                    id="latitude"
                    name="latitude"
                    value={formData.latitude || ''}
                    onChange={handleInputChange}
                  />
                </div>
                <div className="col">
                  <label htmlFor="longitude" className="form-label">Longitude</label>
                  <input
                    type="number"
                    step="0.000001"
                    className="form-control"
                    id="longitude"
                    name="longitude"
                    value={formData.longitude || ''}
                    onChange={handleInputChange}
                  />
                </div>
              </div>
            </>
          )}

          {/* Render media-specific fields if list type is media */}
          {currentList && currentList.type === 'media' && (
            <>
              <div className="mb-3">
                <label htmlFor="mediaType" className="form-label">Media Type</label>
                <select
                  className="form-select"
                  id="mediaType"
                  name="mediaType"
                  value={formData.mediaType || ''}
                  onChange={handleInputChange}
                >
                  <option value="">Select media type</option>
                  <option value="movie">Movie</option>
                  <option value="tv">TV Show</option>
                  <option value="book">Book</option>
                  <option value="music">Music</option>
                  <option value="other">Other</option>
                </select>
              </div>
              <div className="mb-3">
                <label htmlFor="creator" className="form-label">Creator/Author</label>
                <input
                  type="text"
                  className="form-control"
                  id="creator"
                  name="creator"
                  value={formData.creator || ''}
                  onChange={handleInputChange}
                />
              </div>
              <div className="mb-3">
                <label htmlFor="releaseYear" className="form-label">Release Year</label>
                <input
                  type="number"
                  className="form-control"
                  id="releaseYear"
                  name="releaseYear"
                  value={formData.releaseYear || ''}
                  onChange={handleInputChange}
                />
              </div>
            </>
          )}

          {/* Render activity-specific fields if list type is activity */}
          {currentList && currentList.type === 'activity' && (
            <>
              <div className="mb-3">
                <label htmlFor="duration" className="form-label">Duration (minutes)</label>
                <input
                  type="number"
                  className="form-control"
                  id="duration"
                  name="duration"
                  value={formData.duration || ''}
                  onChange={handleInputChange}
                />
              </div>
              <div className="mb-3">
                <label htmlFor="cost" className="form-label">Estimated Cost</label>
                <input
                  type="number"
                  step="0.01"
                  className="form-control"
                  id="cost"
                  name="cost"
                  value={formData.cost || ''}
                  onChange={handleInputChange}
                />
              </div>
              <div className="mb-3">
                <label htmlFor="participants" className="form-label">Recommended Participants</label>
                <input
                  type="text"
                  className="form-control"
                  id="participants"
                  name="participants"
                  value={formData.participants || ''}
                  onChange={handleInputChange}
                />
              </div>
            </>
          )}

          <div className="d-flex justify-content-end gap-2">
            <Button variant="secondary" onClick={handleCancel}>
              Cancel
            </Button>
            <Button variant="primary" onClick={handleSave}>
              Save Changes
            </Button>
          </div>
        </Card.Body>
      </Card>
    );
  };

  const renderItemDetails = () => {
    if (!item) return null;

    return (
      <Card className="mb-4">
        <Card.Body>
          <div className="d-flex justify-content-between align-items-start mb-3">
            <h3>{item.name}</h3>
            <div>
              <Badge bg="info">{currentList?.type || 'Unknown'}</Badge>
            </div>
          </div>
          <p className="text-muted">{item.description}</p>

          {/* Display type-specific metadata */}
          {currentList && currentList.type === 'location' && item.metadata && (
            <div className="mt-4">
              <h4>Location Details</h4>
              {item.metadata.address && (
                <p><strong>Address:</strong> {item.metadata.address}</p>
              )}
              {(item.metadata.latitude && item.metadata.longitude) && (
                <>
                  <p>
                    <strong>Coordinates:</strong> {item.metadata.latitude}, {item.metadata.longitude}
                  </p>
                  <div className="mt-3" style={{ height: '300px' }}>
                    <LocationMapView 
                      items={[item]} 
                      center={{ lat: Number(item.metadata.latitude), lng: Number(item.metadata.longitude) }}
                      zoom={15}
                    />
                  </div>
                </>
              )}
            </div>
          )}

          {currentList && currentList.type === 'media' && item.metadata && (
            <div className="mt-4">
              <h4>Media Details</h4>
              {item.metadata.mediaType && (
                <p><strong>Media Type:</strong> {item.metadata.mediaType}</p>
              )}
              {item.metadata.creator && (
                <p><strong>Creator/Author:</strong> {item.metadata.creator}</p>
              )}
              {item.metadata.releaseYear && (
                <p><strong>Release Year:</strong> {item.metadata.releaseYear}</p>
              )}
            </div>
          )}

          {currentList && currentList.type === 'activity' && item.metadata && (
            <div className="mt-4">
              <h4>Activity Details</h4>
              {item.metadata.duration && (
                <p><strong>Duration:</strong> {item.metadata.duration} minutes</p>
              )}
              {item.metadata.cost && (
                <p><strong>Estimated Cost:</strong> ${item.metadata.cost}</p>
              )}
              {item.metadata.participants && (
                <p><strong>Recommended Participants:</strong> {item.metadata.participants}</p>
              )}
            </div>
          )}
        </Card.Body>
      </Card>
    );
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
          <h4 className="alert-heading">Error Loading Item</h4>
          <p>{error}</p>
          <hr />
          <Button 
            variant="outline-secondary" 
            onClick={() => navigate(`/lists/${listID}`)}
          >
            Back to List
          </Button>
        </div>
      </Container>
    );
  }

  if (!currentList || !item) {
    return (
      <Container className="mt-4">
        <div className="alert alert-warning" role="alert">
          <h4 className="alert-heading">Item Not Found</h4>
          <p>The requested item could not be found. It may have been deleted.</p>
          <hr />
          <Button 
            variant="outline-secondary" 
            onClick={() => navigate(`/lists/${listID}`)}
          >
            Back to List
          </Button>
        </div>
      </Container>
    );
  }

  return (
    <Container className="mt-4">
      <Row className="mb-3">
        <Col>
          <div className="d-flex justify-content-between align-items-center">
            <h2>Item Details</h2>
            <div className="d-flex gap-2">
              {!isEditing ? (
                <>
                  <Button variant="primary" onClick={() => setIsEditing(true)}>
                    Edit
                  </Button>
                  <Button variant="danger" onClick={handleDelete}>
                    Delete
                  </Button>
                  <Button 
                    variant="outline-secondary" 
                    onClick={() => navigate(`/lists/${listID}`)}
                  >
                    Back to List
                  </Button>
                </>
              ) : null}
            </div>
          </div>
        </Col>
      </Row>

      {isEditing ? renderEditForm() : renderItemDetails()}
    </Container>
  );
};

export default ListItemDetailScreen; 
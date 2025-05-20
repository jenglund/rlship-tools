import React, { useState, useEffect } from 'react';
import { Modal, Button, Form, Alert, Spinner } from 'react-bootstrap';
import { useList } from '../contexts/ListContext';
import LocationPicker from './LocationPicker';

const CreateListItemModal = ({ show, onHide, listId, listType }) => {
  const { addListItem, operations } = useList();
  
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [metadata, setMetadata] = useState({});
  const [error, setError] = useState(null);
  const [validated, setValidated] = useState(false);
  
  // Use operations state for loading indicator
  const isCreating = operations.creating;

  // Initialize type-specific metadata fields
  useEffect(() => {
    if (listType) {
      let initialMetadata = {};
      
      switch (listType) {
        case 'location':
          initialMetadata = {
            latitude: '',
            longitude: '',
            address: '',
            location: '',
            city: '',
            cuisine: '',
            price: 'medium'
          };
          break;
        case 'media':
          initialMetadata = {
            category: '',
            length: '',
            platform: ''
          };
          break;
        case 'activity':
          initialMetadata = {
            duration: '',
            category: '',
            participants: ''
          };
          break;
        case 'food':
          initialMetadata = {
            cuisine: '',
            preparationTime: '',
            difficulty: 'medium'
          };
          break;
        default:
          initialMetadata = {};
      }
      
      setMetadata(initialMetadata);
    }
  }, [listType, show]);

  const resetForm = () => {
    setName('');
    setDescription('');
    setError(null);
    setValidated(false);
    
    // Reset metadata based on list type
    if (listType) {
      let initialMetadata = {};
      
      switch (listType) {
        case 'location':
          initialMetadata = {
            latitude: '',
            longitude: '',
            address: '',
            location: '',
            city: '',
            cuisine: '',
            price: 'medium'
          };
          break;
        case 'media':
          initialMetadata = {
            category: '',
            length: '',
            platform: ''
          };
          break;
        case 'activity':
          initialMetadata = {
            duration: '',
            category: '',
            participants: ''
          };
          break;
        case 'food':
          initialMetadata = {
            cuisine: '',
            preparationTime: '',
            difficulty: 'medium'
          };
          break;
        default:
          initialMetadata = {};
      }
      
      setMetadata(initialMetadata);
    }
  };

  const handleClose = () => {
    resetForm();
    onHide();
  };

  const handleMetadataChange = (field, value) => {
    setMetadata(prev => ({
      ...prev,
      [field]: value
    }));
  };

  const handleLocationChange = (locationData) => {
    // Update all location fields at once
    setMetadata(prev => ({
      ...prev,
      latitude: locationData.latitude,
      longitude: locationData.longitude,
      address: locationData.address,
      location: locationData.location
    }));
    
    // If we have a location name and the item name is empty, use the location name
    if (locationData.location && !name) {
      setName(locationData.location);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const form = e.currentTarget;
    
    if (form.checkValidity() === false) {
      e.stopPropagation();
      setValidated(true);
      return;
    }

    try {
      setError(null);
      
      const itemData = {
        name,
        description,
        metadata
      };

      await addListItem(listId, itemData);
      handleClose();
    } catch (err) {
      console.error('Error creating list item:', err);
      setError('Failed to add item to list. Please try again.');
    }
  };

  // Render metadata fields based on list type
  const renderMetadataFields = () => {
    if (!listType) return null;

    switch (listType) {
      case 'location':
        return (
          <>
            <LocationPicker 
              value={{
                latitude: metadata.latitude,
                longitude: metadata.longitude,
                address: metadata.address,
                location: metadata.location
              }}
              onChange={handleLocationChange}
            />
            
            <Form.Group className="mb-3 mt-3">
              <Form.Label>City</Form.Label>
              <Form.Control
                type="text"
                placeholder="City"
                value={metadata.city || ''}
                onChange={(e) => handleMetadataChange('city', e.target.value)}
                disabled={isCreating}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Cuisine</Form.Label>
              <Form.Control
                type="text"
                placeholder="Type of cuisine"
                value={metadata.cuisine || ''}
                onChange={(e) => handleMetadataChange('cuisine', e.target.value)}
                disabled={isCreating}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Price Range</Form.Label>
              <Form.Select
                value={metadata.price || 'medium'}
                onChange={(e) => handleMetadataChange('price', e.target.value)}
                disabled={isCreating}
              >
                <option value="low">$ - Budget</option>
                <option value="medium">$$ - Moderate</option>
                <option value="high">$$$ - Expensive</option>
                <option value="very-high">$$$$ - Very Expensive</option>
              </Form.Select>
            </Form.Group>
          </>
        );
        
      case 'media':
        return (
          <>
            <Form.Group className="mb-3">
              <Form.Label>Category</Form.Label>
              <Form.Select
                value={metadata.category || ''}
                onChange={(e) => handleMetadataChange('category', e.target.value)}
                disabled={isCreating}
              >
                <option value="">Select a category</option>
                <option value="movie">Movie</option>
                <option value="tv-show">TV Show</option>
                <option value="book">Book</option>
                <option value="game">Game</option>
                <option value="music">Music</option>
                <option value="other">Other</option>
              </Form.Select>
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Length</Form.Label>
              <Form.Control
                type="text"
                placeholder="Duration or length"
                value={metadata.length || ''}
                onChange={(e) => handleMetadataChange('length', e.target.value)}
                disabled={isCreating}
              />
              <Form.Text className="text-muted">
                e.g. "2.5 hours", "30 minutes", "350 pages"
              </Form.Text>
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Platform</Form.Label>
              <Form.Control
                type="text"
                placeholder="Where to watch/read/listen"
                value={metadata.platform || ''}
                onChange={(e) => handleMetadataChange('platform', e.target.value)}
                disabled={isCreating}
              />
              <Form.Text className="text-muted">
                e.g. "Netflix", "Spotify", "Amazon"
              </Form.Text>
            </Form.Group>
          </>
        );
        
      case 'activity':
        return (
          <>
            <Form.Group className="mb-3">
              <Form.Label>Category</Form.Label>
              <Form.Select
                value={metadata.category || ''}
                onChange={(e) => handleMetadataChange('category', e.target.value)}
                disabled={isCreating}
              >
                <option value="">Select a category</option>
                <option value="outdoor">Outdoor</option>
                <option value="indoor">Indoor</option>
                <option value="sport">Sport</option>
                <option value="entertainment">Entertainment</option>
                <option value="education">Educational</option>
                <option value="social">Social</option>
                <option value="other">Other</option>
              </Form.Select>
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Duration</Form.Label>
              <Form.Control
                type="text"
                placeholder="How long does it take?"
                value={metadata.duration || ''}
                onChange={(e) => handleMetadataChange('duration', e.target.value)}
                disabled={isCreating}
              />
              <Form.Text className="text-muted">
                e.g. "2 hours", "30 minutes", "All day"
              </Form.Text>
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Participants</Form.Label>
              <Form.Control
                type="text"
                placeholder="How many people?"
                value={metadata.participants || ''}
                onChange={(e) => handleMetadataChange('participants', e.target.value)}
                disabled={isCreating}
              />
              <Form.Text className="text-muted">
                e.g. "2-4 people", "Group of 6"
              </Form.Text>
            </Form.Group>
          </>
        );
        
      case 'food':
        return (
          <>
            <Form.Group className="mb-3">
              <Form.Label>Cuisine</Form.Label>
              <Form.Control
                type="text"
                placeholder="Type of cuisine"
                value={metadata.cuisine || ''}
                onChange={(e) => handleMetadataChange('cuisine', e.target.value)}
                disabled={isCreating}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Preparation Time</Form.Label>
              <Form.Control
                type="text"
                placeholder="How long to prepare"
                value={metadata.preparationTime || ''}
                onChange={(e) => handleMetadataChange('preparationTime', e.target.value)}
                disabled={isCreating}
              />
              <Form.Text className="text-muted">
                e.g. "30 minutes", "1 hour", "15 min prep + 45 min cook"
              </Form.Text>
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Difficulty</Form.Label>
              <Form.Select
                value={metadata.difficulty || 'medium'}
                onChange={(e) => handleMetadataChange('difficulty', e.target.value)}
                disabled={isCreating}
              >
                <option value="easy">Easy</option>
                <option value="medium">Medium</option>
                <option value="hard">Challenging</option>
              </Form.Select>
            </Form.Group>
          </>
        );
        
      default:
        return null;
    }
  };

  return (
    <Modal show={show} onHide={handleClose} centered size={listType === 'location' ? 'lg' : 'md'}>
      <Modal.Header closeButton>
        <Modal.Title>Add Item to List</Modal.Title>
      </Modal.Header>
      
      <Form noValidate validated={validated} onSubmit={handleSubmit}>
        <Modal.Body>
          {error && (
            <Alert variant="danger">{error}</Alert>
          )}
          
          <Form.Group className="mb-3">
            <Form.Label>Name</Form.Label>
            <Form.Control
              type="text"
              placeholder="Enter item name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              maxLength={100}
              disabled={isCreating}
            />
            <Form.Control.Feedback type="invalid">
              Please provide an item name.
            </Form.Control.Feedback>
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Label>Description</Form.Label>
            <Form.Control
              as="textarea"
              rows={3}
              placeholder="Enter item description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={500}
              disabled={isCreating}
            />
          </Form.Group>
          
          {renderMetadataFields()}
        </Modal.Body>
        
        <Modal.Footer>
          <Button variant="secondary" onClick={handleClose} disabled={isCreating}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={isCreating}>
            {isCreating ? (
              <>
                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" className="me-2" />
                Adding...
              </>
            ) : 'Add Item'}
          </Button>
        </Modal.Footer>
      </Form>
    </Modal>
  );
};

export default CreateListItemModal; 
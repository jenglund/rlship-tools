import React, { useState, useEffect } from 'react';
import { Modal, Button, Form, Alert } from 'react-bootstrap';
import { useList } from '../contexts/ListContext';

const CreateListItemModal = ({ show, onHide, listId, listType }) => {
  const { addListItem } = useList();
  
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [metadata, setMetadata] = useState({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [validated, setValidated] = useState(false);

  // Initialize type-specific metadata fields
  useEffect(() => {
    if (listType) {
      let initialMetadata = {};
      
      switch (listType) {
        case 'location':
          initialMetadata = {
            address: '',
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
            address: '',
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

  const handleSubmit = async (e) => {
    e.preventDefault();
    const form = e.currentTarget;
    
    if (form.checkValidity() === false) {
      e.stopPropagation();
      setValidated(true);
      return;
    }

    try {
      setLoading(true);
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
    } finally {
      setLoading(false);
    }
  };

  // Render metadata fields based on list type
  const renderMetadataFields = () => {
    if (!listType) return null;

    switch (listType) {
      case 'location':
        return (
          <>
            <Form.Group className="mb-3">
              <Form.Label>Address</Form.Label>
              <Form.Control
                type="text"
                placeholder="Street address"
                value={metadata.address || ''}
                onChange={(e) => handleMetadataChange('address', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>City</Form.Label>
              <Form.Control
                type="text"
                placeholder="City"
                value={metadata.city || ''}
                onChange={(e) => handleMetadataChange('city', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Cuisine</Form.Label>
              <Form.Control
                type="text"
                placeholder="Type of cuisine"
                value={metadata.cuisine || ''}
                onChange={(e) => handleMetadataChange('cuisine', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Price Range</Form.Label>
              <Form.Select
                value={metadata.price || 'medium'}
                onChange={(e) => handleMetadataChange('price', e.target.value)}
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
              <Form.Control
                type="text"
                placeholder="Movie, TV Show, Book, etc."
                value={metadata.category || ''}
                onChange={(e) => handleMetadataChange('category', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Length</Form.Label>
              <Form.Control
                type="text"
                placeholder="Duration or pages"
                value={metadata.length || ''}
                onChange={(e) => handleMetadataChange('length', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Platform</Form.Label>
              <Form.Control
                type="text"
                placeholder="Netflix, HBO, Kindle, etc."
                value={metadata.platform || ''}
                onChange={(e) => handleMetadataChange('platform', e.target.value)}
              />
            </Form.Group>
          </>
        );
        
      case 'activity':
        return (
          <>
            <Form.Group className="mb-3">
              <Form.Label>Duration</Form.Label>
              <Form.Control
                type="text"
                placeholder="How long it takes"
                value={metadata.duration || ''}
                onChange={(e) => handleMetadataChange('duration', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Category</Form.Label>
              <Form.Control
                type="text"
                placeholder="Indoor, Outdoor, Game, etc."
                value={metadata.category || ''}
                onChange={(e) => handleMetadataChange('category', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Participants</Form.Label>
              <Form.Control
                type="text"
                placeholder="Number of people needed"
                value={metadata.participants || ''}
                onChange={(e) => handleMetadataChange('participants', e.target.value)}
              />
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
                placeholder="Italian, Mexican, etc."
                value={metadata.cuisine || ''}
                onChange={(e) => handleMetadataChange('cuisine', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Preparation Time</Form.Label>
              <Form.Control
                type="text"
                placeholder="How long to make"
                value={metadata.preparationTime || ''}
                onChange={(e) => handleMetadataChange('preparationTime', e.target.value)}
              />
            </Form.Group>
            
            <Form.Group className="mb-3">
              <Form.Label>Difficulty</Form.Label>
              <Form.Select
                value={metadata.difficulty || 'medium'}
                onChange={(e) => handleMetadataChange('difficulty', e.target.value)}
              >
                <option value="easy">Easy</option>
                <option value="medium">Medium</option>
                <option value="hard">Hard</option>
              </Form.Select>
            </Form.Group>
          </>
        );
        
      default:
        return null;
    }
  };

  return (
    <Modal show={show} onHide={handleClose} centered>
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
            />
            <Form.Control.Feedback type="invalid">
              Please provide an item name.
            </Form.Control.Feedback>
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Label>Description</Form.Label>
            <Form.Control
              as="textarea"
              rows={2}
              placeholder="Enter item description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={500}
            />
          </Form.Group>
          
          {renderMetadataFields()}
        </Modal.Body>
        
        <Modal.Footer>
          <Button variant="secondary" onClick={handleClose}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={loading}>
            {loading ? 'Adding...' : 'Add Item'}
          </Button>
        </Modal.Footer>
      </Form>
    </Modal>
  );
};

export default CreateListItemModal; 
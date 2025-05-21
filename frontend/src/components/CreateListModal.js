import React, { useState, useEffect } from 'react';
import { Modal, Button, Form, Alert, Spinner, Card, ListGroup } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { useList } from '../contexts/ListContext';
import { getListTypeOptions } from '../utils/listTypeUtils';

const CreateListModal = ({ show, onHide }) => {
  const { createList, operations, error: contextError, userLists, fetchUserLists } = useList();
  
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [type, setType] = useState('general');
  const [visibility, setVisibility] = useState('private');
  const [localError, setLocalError] = useState(null);
  const [validated, setValidated] = useState(false);
  const [existingLists, setExistingLists] = useState([]);
  const [checkingExisting, setCheckingExisting] = useState(false);
  
  // Use the creating operation status from context
  const isCreating = operations.creating;
  // Combine local and context errors
  const error = localError || contextError;
  // Get list type options from utility
  const listTypeOptions = getListTypeOptions();
  
  // Fetch user lists when modal opens
  useEffect(() => {
    if (show) {
      fetchUserLists();
    }
  }, [show, fetchUserLists]);

  const resetForm = () => {
    setName('');
    setDescription('');
    setType('general');
    setVisibility('private');
    setLocalError(null);
    setValidated(false);
    setExistingLists([]);
  };

  const handleClose = () => {
    resetForm();
    onHide();
  };
  
  // Check if a list with the same name already exists
  const checkExistingLists = () => {
    if (!name.trim()) return [];
    
    const matchingLists = userLists.filter(list => 
      list.name.toLowerCase() === name.toLowerCase() ||
      list.name.toLowerCase().includes(name.toLowerCase())
    );
    
    return matchingLists;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const form = e.currentTarget;
    
    if (form.checkValidity() === false) {
      e.stopPropagation();
      setValidated(true);
      return;
    }
    
    // Check for existing lists with similar names
    setCheckingExisting(true);
    const matches = checkExistingLists();
    setCheckingExisting(false);
    
    // If we found exact name matches, show them instead of creating a new list
    const exactMatches = matches.filter(list => 
      list.name.toLowerCase() === name.toLowerCase()
    );
    
    if (exactMatches.length > 0) {
      setExistingLists(exactMatches);
      setLocalError(`A list with the name "${name}" already exists. Please use the existing list or choose a different name.`);
      return;
    }
    
    // If we found similar lists, show them as a warning but still allow creation
    if (matches.length > 0) {
      setExistingLists(matches);
      // Continue with list creation anyway
    }

    try {
      setLocalError(null);
      
      const listData = {
        name,
        description,
        type,
        visibility,
        metadata: {}
      };

      await createList(listData);
      handleClose();
    } catch (err) {
      console.error('Error creating list:', err);
      
      // Check if it's a 409 Conflict (list already exists)
      if (err.response && err.response.status === 409) {
        // Check if the server returned the duplicate list in the error response data
        const duplicateList = err.response.data && err.response.data.data;
        
        if (duplicateList && duplicateList.id) {
          // We received the duplicate list from the server, add it to existingLists
          setExistingLists([duplicateList]);
          setLocalError(`A list with the name "${name}" already exists. Please use the existing list or choose a different name.`);
        } else {
          // Refresh the list of user lists to make sure we have the latest data
          try {
            await fetchUserLists();
            
            // Check for existing lists again
            const updatedMatches = checkExistingLists();
            if (updatedMatches.length > 0) {
              setExistingLists(updatedMatches);
              setLocalError(`A list with the name "${name}" already exists. Please use the existing list or choose a different name.`);
            } else {
              setLocalError('A list with this name already exists but cannot be retrieved. Please try a different name.');
            }
          } catch (refreshError) {
            console.error('Error refreshing user lists:', refreshError);
            setLocalError('A list with this name already exists. Please try a different name or refresh the page.');
          }
        }
      } else {
        setLocalError('Failed to create list. Please try again.');
      }
    }
  };

  const TypeOption = ({ option, isSelected }) => (
    <Card 
      className={`mb-2 type-option ${isSelected ? 'border-primary' : ''}`}
      onClick={() => setType(option.value)}
      style={{ 
        cursor: 'pointer', 
        borderWidth: isSelected ? '2px' : '1px',
        backgroundColor: isSelected ? 'rgba(0, 123, 255, 0.05)' : 'white'
      }}
    >
      <Card.Body className="d-flex align-items-center p-2">
        <div className="me-3" style={{ color: option.icon.props ? option.icon.props.color : '' }}>
          {option.icon}
        </div>
        <div>
          <div className="fw-bold">{option.label}</div>
          <small className="text-muted">{option.description}</small>
        </div>
      </Card.Body>
    </Card>
  );

  return (
    <Modal show={show} onHide={handleClose} centered size="lg">
      <Modal.Header closeButton>
        <Modal.Title>Create New List</Modal.Title>
      </Modal.Header>
      
      <Form noValidate validated={validated} onSubmit={handleSubmit}>
        <Modal.Body>
          {error && (
            <Alert variant="danger">{error}</Alert>
          )}
          
          {existingLists.length > 0 && (
            <Alert variant="warning">
              <Alert.Heading>Similar Lists Found</Alert.Heading>
              <p>We found the following lists with similar names:</p>
              <ListGroup className="mb-3">
                {existingLists.map(list => (
                  <ListGroup.Item key={list.id} className="d-flex justify-content-between align-items-center">
                    <div>
                      <strong>{list.name}</strong>
                      {list.description && <p className="mb-0 small">{list.description}</p>}
                    </div>
                    <Link 
                      to={`/lists/${list.id}`} 
                      className="btn btn-sm btn-primary"
                      onClick={handleClose}
                    >
                      View List
                    </Link>
                  </ListGroup.Item>
                ))}
              </ListGroup>
            </Alert>
          )}
          
          <Form.Group className="mb-3">
            <Form.Label>Name</Form.Label>
            <Form.Control
              type="text"
              placeholder="Enter list name"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                setExistingLists([]);
                setLocalError(null);
              }}
              required
              maxLength={100}
              disabled={isCreating || checkingExisting}
            />
            <Form.Control.Feedback type="invalid">
              Please provide a list name.
            </Form.Control.Feedback>
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Label>Description</Form.Label>
            <Form.Control
              as="textarea"
              rows={3}
              placeholder="Enter list description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              maxLength={500}
              disabled={isCreating || checkingExisting}
            />
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Label>List Type</Form.Label>
            
            <div className="mt-2 mb-2">
              {listTypeOptions.map((option) => (
                <TypeOption 
                  key={option.value} 
                  option={option} 
                  isSelected={type === option.value} 
                />
              ))}
            </div>
            
            <Form.Text className="text-muted">
              Select a type that best describes the items in this list.
            </Form.Text>
          </Form.Group>
          
          <Form.Group className="mb-3">
            <Form.Label>Visibility</Form.Label>
            <Form.Select
              value={visibility}
              onChange={(e) => setVisibility(e.target.value)}
              required
              disabled={isCreating || checkingExisting}
            >
              <option value="private">Private</option>
              <option value="public">Public</option>
            </Form.Select>
            <Form.Text className="text-muted">
              Private lists are only visible to you and tribes you share with.
            </Form.Text>
          </Form.Group>
        </Modal.Body>
        
        <Modal.Footer>
          <Button variant="secondary" onClick={handleClose} disabled={isCreating || checkingExisting}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={isCreating || checkingExisting}>
            {isCreating || checkingExisting ? (
              <>
                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" className="me-2" />
                {checkingExisting ? 'Checking...' : 'Creating...'}
              </>
            ) : 'Create List'}
          </Button>
        </Modal.Footer>
      </Form>
    </Modal>
  );
};

export default CreateListModal; 
import React, { useState } from 'react';
import { Modal, Button, Form, Alert, Spinner, Card } from 'react-bootstrap';
import { useList } from '../contexts/ListContext';
import { getListTypeOptions } from '../utils/listTypeUtils';

const CreateListModal = ({ show, onHide }) => {
  const { createList, operations, error: contextError } = useList();
  
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [type, setType] = useState('general');
  const [visibility, setVisibility] = useState('private');
  const [localError, setLocalError] = useState(null);
  const [validated, setValidated] = useState(false);
  
  // Use the creating operation status from context
  const isCreating = operations.creating;
  // Combine local and context errors
  const error = localError || contextError;
  // Get list type options from utility
  const listTypeOptions = getListTypeOptions();

  const resetForm = () => {
    setName('');
    setDescription('');
    setType('general');
    setVisibility('private');
    setLocalError(null);
    setValidated(false);
  };

  const handleClose = () => {
    resetForm();
    onHide();
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
      setLocalError('Failed to create list. Please try again.');
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
          
          <Form.Group className="mb-3">
            <Form.Label>Name</Form.Label>
            <Form.Control
              type="text"
              placeholder="Enter list name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              required
              maxLength={100}
              disabled={isCreating}
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
              disabled={isCreating}
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
              disabled={isCreating}
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
          <Button variant="secondary" onClick={handleClose} disabled={isCreating}>
            Cancel
          </Button>
          <Button type="submit" variant="primary" disabled={isCreating}>
            {isCreating ? (
              <>
                <Spinner as="span" animation="border" size="sm" role="status" aria-hidden="true" className="me-2" />
                Creating...
              </>
            ) : 'Create List'}
          </Button>
        </Modal.Footer>
      </Form>
    </Modal>
  );
};

export default CreateListModal; 
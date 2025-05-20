import React from 'react';
import { Card, Button, Badge } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';

const ListItemCard = ({ item, onEdit, onDelete }) => {
  const navigate = useNavigate();
  
  // Get appropriate icon based on item type or parent list type
  const getTypeIcon = (type) => {
    switch (type) {
      case 'location':
        return '📍';
      case 'media':
        return '🎬';
      case 'activity':
        return '🎮';
      case 'food':
        return '🍽️';
      default:
        return '📋';
    }
  };

  const handleClick = () => {
    // Navigate to the detail view
    if (item.id && item.listId) {
      navigate(`/lists/${item.listId}/items/${item.id}`);
    }
  };

  const handleEdit = (e) => {
    e.stopPropagation();
    if (onEdit) {
      onEdit(item);
    }
  };

  const handleDelete = (e) => {
    e.stopPropagation();
    if (onDelete) {
      onDelete(item);
    }
  };

  const handleViewDetails = (e) => {
    e.stopPropagation();
    if (item.id && item.listId) {
      navigate(`/lists/${item.listId}/items/${item.id}`);
    }
  };

  const renderMetadata = () => {
    if (!item.metadata) return null;

    return (
      <div className="mt-2">
        {Object.entries(item.metadata).map(([key, value]) => {
          // Skip empty values or internal properties (those starting with _)
          if (!value || key.startsWith('_')) return null;
          
          return (
            <div key={key} className="text-muted small">
              <strong>{key.charAt(0).toUpperCase() + key.slice(1)}: </strong>
              <span className="text-break">{value}</span>
            </div>
          );
        })}
      </div>
    );
  };

  return (
    <Card 
      className="h-100 shadow-sm item-card"
      onClick={handleClick}
      style={{ cursor: 'pointer' }}
    >
      <Card.Body className="d-flex flex-column">
        <Card.Title className="text-break">
          {getTypeIcon(item.type || 'general')} {item.name}
        </Card.Title>
        {item.description && (
          <Card.Text className="flex-grow-1 text-break">{item.description}</Card.Text>
        )}
        
        <div className="flex-grow-1">
          {renderMetadata()}
        </div>
        
        <div className="mt-3 d-flex flex-wrap gap-2 justify-content-end">
          <Button 
            variant="outline-secondary" 
            size="sm"
            onClick={handleViewDetails}
          >
            View Details
          </Button>
          {onEdit && (
            <Button 
              variant="outline-primary" 
              size="sm"
              onClick={handleEdit}
            >
              Edit
            </Button>
          )}
          {onDelete && (
            <Button 
              variant="outline-danger" 
              size="sm"
              onClick={handleDelete}
            >
              Delete
            </Button>
          )}
        </div>
      </Card.Body>
    </Card>
  );
};

export default ListItemCard; 
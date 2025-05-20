import React from 'react';
import { Card, Button, Badge } from 'react-bootstrap';
import { useNavigate } from 'react-router-dom';

const ListCard = ({ list, onShare, onDelete }) => {
  const navigate = useNavigate();

  const handleViewList = () => {
    navigate(`/lists/${list.id}`);
  };

  const handleShareList = (e) => {
    e.stopPropagation();
    if (onShare) {
      onShare(list);
    }
  };

  const handleDeleteList = (e) => {
    e.stopPropagation();
    if (onDelete) {
      onDelete(list);
    }
  };

  // Get appropriate icon based on list type
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

  return (
    <Card className="mb-3 list-card" style={{ cursor: 'pointer' }} onClick={handleViewList}>
      <Card.Body>
        <div className="d-flex justify-content-between align-items-start">
          <div>
            <Card.Title>
              {getTypeIcon(list.type)} {list.name}
            </Card.Title>
            {list.shared && (
              <Badge bg="success" className="me-1">Shared</Badge>
            )}
            <Badge bg="info">{list.type}</Badge>
          </div>
        </div>
        
        <Card.Text className="text-muted small mt-2">{list.description}</Card.Text>
        
        <div className="d-flex justify-content-between align-items-center mt-3">
          <span className="text-muted small">
            {list.itemCount || 0} items
          </span>
          <div>
            {onShare && (
              <Button 
                variant="outline-success" 
                size="sm" 
                className="me-2"
                onClick={handleShareList}
              >
                Share
              </Button>
            )}
            {onDelete && (
              <Button 
                variant="outline-danger" 
                size="sm"
                onClick={handleDeleteList}
              >
                Delete
              </Button>
            )}
            <Button 
              variant="primary" 
              size="sm" 
              className="ms-2"
              onClick={handleViewList}
            >
              View
            </Button>
          </div>
        </div>
      </Card.Body>
    </Card>
  );
};

export default ListCard; 
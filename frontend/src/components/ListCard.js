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
    <Card className="mb-3 list-card h-100" style={{ cursor: 'pointer' }} onClick={handleViewList}>
      <Card.Body className="d-flex flex-column">
        <div className="d-flex justify-content-between align-items-start">
          <div>
            <Card.Title className="text-break">
              {getTypeIcon(list.type)} {list.name}
            </Card.Title>
            <div>
              {list.shared && (
                <Badge bg="success" className="me-1">Shared</Badge>
              )}
              <Badge bg="info">{list.type}</Badge>
            </div>
          </div>
        </div>
        
        <Card.Text className="text-muted small mt-2 flex-grow-1">{list.description}</Card.Text>
        
        <div className="d-flex flex-column flex-sm-row justify-content-between align-items-sm-center mt-3 gap-2">
          <span className="text-muted small">
            {list.itemCount || 0} items
          </span>
          <div className="d-flex gap-2 justify-content-end flex-wrap">
            {onShare && (
              <Button 
                variant="outline-success" 
                size="sm"
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
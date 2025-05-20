import React from 'react';
import { Card, Badge } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { renderListTypeIcon, getListTypeColor } from '../utils/listTypeUtils';

const ListCard = ({ list, className = '' }) => {
  const { id, name, description, type, itemCount, visibility } = list;

  // Get the type-specific color for styling
  const typeColor = getListTypeColor(type);
  
  return (
    <Card 
      className={`h-100 shadow-sm list-card ${className}`}
      as={Link} 
      to={`/lists/${id}`}
      style={{ 
        textDecoration: 'none',
        color: 'inherit',
        borderLeft: `4px solid ${typeColor}`
      }}
    >
      <Card.Body className="d-flex flex-column">
        <Card.Title className="d-flex align-items-center mb-2">
          {renderListTypeIcon(type)} <span className="ms-2 text-break">{name}</span>
        </Card.Title>
        
        <div className="mb-2 d-flex gap-2">
          <Badge bg="info" className="text-uppercase small">
            {type}
          </Badge>
          <Badge 
            bg={visibility === 'private' ? 'secondary' : 'success'} 
            className="text-uppercase small"
          >
            {visibility}
          </Badge>
          <Badge bg="primary" className="text-uppercase small">
            {itemCount} item{itemCount !== 1 ? 's' : ''}
          </Badge>
        </div>
        
        {description && (
          <Card.Text className="flex-grow-1 small text-muted text-break">
            {description.length > 100 
              ? `${description.substring(0, 100)}...` 
              : description
            }
          </Card.Text>
        )}
      </Card.Body>
    </Card>
  );
};

export default ListCard; 
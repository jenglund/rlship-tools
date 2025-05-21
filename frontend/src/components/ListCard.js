import React, { useEffect } from 'react';
import { Card, Badge } from 'react-bootstrap';
import { Link } from 'react-router-dom';
import { renderListTypeIcon, getListTypeColor } from '../utils/listTypeUtils';

const ListCard = ({ list, className = '' }) => {
  // Debug validation for list prop
  useEffect(() => {
    if (!list) {
      console.error('ListCard: list prop is undefined or null');
      return;
    }
    
    // Check for API response structure and extract data if needed
    if (list.success === true && list.data) {
      console.warn('ListCard: Received API response object instead of list data', list);
      list = list.data;
    }
    
    if (!list.id) {
      console.error('ListCard: list.id is missing', list);
    }
    
    if (!list.type) {
      console.warn('ListCard: list.type is missing', list);
    }
  }, [list]);
  
  // Guard against null/undefined list
  if (!list) {
    console.error('ListCard: Rendering with null/undefined list');
    return null;
  }
  
  // Extract API response data if the wrong format was passed
  let listData = list;
  if (list.success === true && list.data) {
    console.warn('ListCard: Received API response object, extracting data field');
    listData = list.data;
  }
  
  // Destructure with default values to handle missing properties
  const { 
    id = null, 
    name = 'Unnamed List', 
    description = '', 
    type = 'general', 
    itemCount = 0, 
    visibility = 'private' 
  } = listData;

  // Skip rendering if we don't have an ID
  if (!id) {
    console.error('ListCard: Cannot render list without ID', listData);
    return null;
  }

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
          {renderListTypeIcon(type)} <span className="ms-2 text-break">{name || 'Unnamed List'}</span>
        </Card.Title>
        
        <div className="mb-2 d-flex gap-2">
          <Badge bg="info" className="text-uppercase small">
            {type || 'unknown'}
          </Badge>
          <Badge 
            bg={visibility === 'private' ? 'secondary' : 'success'} 
            className="text-uppercase small"
          >
            {visibility || 'unknown'}
          </Badge>
          <Badge bg="primary" className="text-uppercase small">
            {typeof itemCount === 'number' ? `${itemCount} item${itemCount !== 1 ? 's' : ''}` : '0 items'}
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
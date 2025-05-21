import React, { useState, useEffect } from 'react';
import { Card, ListGroup, Button, Spinner, Alert, Badge } from 'react-bootstrap';
import { useList } from '../contexts/ListContext';
import { useAuth } from '../contexts/AuthContext';
import listService from '../services/listService';

const ListSharesView = ({ listId }) => {
  const [shares, setShares] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [isOwner, setIsOwner] = useState(false);
  const [ownerInfo, setOwnerInfo] = useState(null);
  
  const { getListShares, currentList } = useList();
  const { currentUser } = useAuth();
  
  useEffect(() => {
    const fetchListOwnership = async () => {
      if (!currentList) return;
      
      try {
        // Check if the current user is the owner
        const isCurrentUserOwner = currentList.ownerId === currentUser?.id && 
                                  currentList.ownerType === 'user';
        setIsOwner(isCurrentUserOwner);
        
        // Set owner info for display
        if (currentList.ownerType === 'user') {
          setOwnerInfo({
            type: 'User',
            id: currentList.ownerId,
            displayName: currentList.ownerId === currentUser?.id ? 'You' : 'Another User'
          });
        } else if (currentList.ownerType === 'tribe') {
          // In a real app, we would fetch the tribe name
          setOwnerInfo({
            type: 'Tribe',
            id: currentList.ownerId,
            displayName: 'Tribe'
          });
        }
      } catch (err) {
        console.error('Error determining list ownership:', err);
        // Don't set error state for this, just log it
      }
    };
    
    const fetchListShares = async () => {
      try {
        setLoading(true);
        setError(null);
        
        // Only try to get shares if we are the owner
        if (isOwner) {
          // Use the context method to get shares
          const data = await getListShares(listId);
          
          // Always ensure we have an array, even if the API returns null/undefined
          setShares(Array.isArray(data) ? data : []);
        } else {
          // Not the owner, so don't try to fetch shares
          setShares([]);
        }
      } catch (err) {
        console.error('Error fetching list shares:', err);
        // Handle common errors
        if (err?.response?.status === 403) {
          // Permission denied - just set shares to empty
          setShares([]);
        } else if (err?.response?.status !== 404) {
          // For any error other than 404, show an error message
          setError('Failed to load list shares');
        }
        // Initialize with empty array rather than null
        setShares([]);
      } finally {
        setLoading(false);
      }
    };
    
    fetchListOwnership();
    // Only fetch shares after we've determined ownership
    if (currentList) {
      fetchListShares();
    }
  }, [listId, getListShares, currentList, currentUser, isOwner]);
  
  const handleUnshare = async (tribeID) => {
    try {
      setError(null);
      await listService.unshareListWithTribe(listId, tribeID);
      
      // Update the local state
      setShares(prev => prev.filter(share => share.tribeId !== tribeID));
    } catch (err) {
      console.error('Error unsharing list:', err);
      setError('Failed to unshare list');
    }
  };
  
  if (loading) {
    return (
      <div className="text-center my-3">
        <Spinner animation="border" size="sm" />
      </div>
    );
  }
  
  if (error) {
    return <Alert variant="danger">{error}</Alert>;
  }

  // Always show the ownership info
  return (
    <>
      {/* Ownership information */}
      <Card className="mb-4">
        <Card.Header>List Ownership</Card.Header>
        <Card.Body>
          {ownerInfo ? (
            <div className="d-flex justify-content-between align-items-center">
              <div>
                <strong>Owner:</strong> {ownerInfo.displayName}
                <Badge bg="info" className="ms-2">{ownerInfo.type}</Badge>
              </div>
              {isOwner && (
                <Badge bg="success">You own this list</Badge>
              )}
            </div>
          ) : (
            <p className="text-muted mb-0">Ownership information unavailable</p>
          )}
        </Card.Body>
      </Card>
      
      {/* Sharing information - only visible to the owner */}
      {isOwner && (
        <Card className="mb-4">
          <Card.Header>Shared With</Card.Header>
          {shares.length === 0 ? (
            <Card.Body>
              <p className="text-muted mb-0">This list is not shared with any tribes.</p>
            </Card.Body>
          ) : (
            <ListGroup variant="flush">
              {shares.map(share => (
                <ListGroup.Item key={share.tribeId} className="d-flex justify-content-between align-items-center">
                  <div>
                    <strong>{share.name || share.tribeName || 'Unknown Tribe'}</strong>
                    {share.description && <p className="text-muted mb-0">{share.description}</p>}
                  </div>
                  <Button 
                    variant="outline-danger" 
                    size="sm" 
                    onClick={() => handleUnshare(share.tribeId)}
                  >
                    Unshare
                  </Button>
                </ListGroup.Item>
              ))}
            </ListGroup>
          )}
        </Card>
      )}
    </>
  );
};

export default ListSharesView; 
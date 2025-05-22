import React, { useState, useEffect, useRef } from 'react';
import { Modal, Button, Form, ListGroup, Badge, Spinner, Alert } from 'react-bootstrap';
import { useAuth } from '../contexts/AuthContext';
import { useList } from '../contexts/ListContext';
import listService from '../services/listService';
import tribeService from '../services/tribeService';

const ShareListModal = ({ show, onHide, listId }) => {
  const [tribes, setTribes] = useState([]);
  const [sharedWith, setSharedWith] = useState([]);
  const [loading, setLoading] = useState(false);
  const [sharesLoading, setSharesLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const fetchedRef = useRef(false);
  
  const { currentUser } = useAuth();
  const { getListShares } = useList();
  
  // Load user tribes when modal opens
  useEffect(() => {
    // Only fetch data when the modal is shown and we haven't already fetched
    if (show && !fetchedRef.current) {
      fetchedRef.current = true;
      
      const fetchUserTribes = async () => {
        try {
          setLoading(true);
          console.log('ShareListModal: Fetching user tribes');
          const data = await tribeService.getUserTribes();
          console.log('ShareListModal: Tribes fetched successfully', data);
          setTribes(data || []);
        } catch (err) {
          console.error('Error fetching tribes:', err);
          setError('Failed to load your tribes. Please try again.');
        } finally {
          setLoading(false);
        }
      };
      
      const fetchListShares = async () => {
        try {
          setSharesLoading(true);
          console.log('ShareListModal: Fetching list shares for list ID', listId);
          // Use context method instead of direct service call
          const data = await getListShares(listId);
          console.log('ShareListModal: Shares fetched successfully', data);
          setSharedWith(data || []);
        } catch (err) {
          console.error('Error fetching list shares:', err);
          setError('Failed to load list shares. Please try again.');
        } finally {
          setSharesLoading(false);
        }
      };
      
      fetchUserTribes();
      fetchListShares();
    }
    
    // Reset the fetch flag when modal is closed
    if (!show) {
      fetchedRef.current = false;
    }
  }, [show, listId]);
  
  const handleShare = async (tribeID) => {
    try {
      setError(null);
      setSuccess(null);
      
      console.log(`ShareListModal: Sharing list ${listId} with tribe ${tribeID}`);
      await listService.shareListWithTribe(listId, tribeID);
      
      // Update the shared with list
      setSharedWith(prev => {
        // Avoid duplicates
        if (!prev.find(tribe => tribe.id === tribeID)) {
          const tribe = tribes.find(t => t.id === tribeID);
          return [...prev, tribe];
        }
        return prev;
      });
      
      setSuccess(`List shared successfully!`);
    } catch (err) {
      console.error('Error sharing list:', err);
      setError('Failed to share list. Please try again.');
    }
  };
  
  const handleUnshare = async (tribeID) => {
    try {
      setError(null);
      setSuccess(null);
      
      console.log(`ShareListModal: Unsharing list ${listId} from tribe ${tribeID}`);
      await listService.unshareListWithTribe(listId, tribeID);
      
      // Update the shared with list
      setSharedWith(prev => prev.filter(tribe => tribe.id !== tribeID));
      
      setSuccess(`List unshared successfully!`);
    } catch (err) {
      console.error('Error unsharing list:', err);
      setError('Failed to unshare list. Please try again.');
    }
  };
  
  const isShared = (tribeID) => {
    return sharedWith.some(tribe => tribe.id === tribeID);
  };
  
  return (
    <Modal show={show} onHide={onHide} centered>
      <Modal.Header closeButton>
        <Modal.Title>Share List</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {error && <Alert variant="danger">{error}</Alert>}
        {success && <Alert variant="success">{success}</Alert>}
        
        <h5>Currently shared with:</h5>
        {sharesLoading ? (
          <div className="text-center my-3">
            <Spinner animation="border" size="sm" />
          </div>
        ) : sharedWith.length === 0 ? (
          <p className="text-muted">This list is not shared with any tribes.</p>
        ) : (
          <ListGroup className="mb-3">
            {sharedWith.map(tribe => (
              <ListGroup.Item key={tribe.id} className="d-flex justify-content-between align-items-center">
                {tribe.name || tribe.tribeName || 'Unknown Tribe'}
                <Button 
                  variant="outline-danger" 
                  size="sm" 
                  onClick={() => handleUnshare(tribe.id)}
                >
                  Unshare
                </Button>
              </ListGroup.Item>
            ))}
          </ListGroup>
        )}
        
        <h5>Your tribes:</h5>
        {loading ? (
          <div className="text-center my-3">
            <Spinner animation="border" size="sm" />
          </div>
        ) : tribes.length === 0 ? (
          <p className="text-muted">You don't belong to any tribes yet.</p>
        ) : (
          <ListGroup>
            {tribes.map(tribe => (
              <ListGroup.Item key={tribe.id} className="d-flex justify-content-between align-items-center">
                {tribe.name}
                {isShared(tribe.id) ? (
                  <Badge bg="success">Shared</Badge>
                ) : (
                  <Button 
                    variant="outline-primary" 
                    size="sm" 
                    onClick={() => handleShare(tribe.id)}
                  >
                    Share
                  </Button>
                )}
              </ListGroup.Item>
            ))}
          </ListGroup>
        )}
      </Modal.Body>
      <Modal.Footer>
        <Button variant="secondary" onClick={onHide}>
          Close
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default ShareListModal; 
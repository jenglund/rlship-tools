import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Navigate } from 'react-router-dom';
import { Container, Card, Table, Button, Alert, Modal } from 'react-bootstrap';
import { useAuth } from '../contexts/AuthContext';
import { isAdmin } from '../utils/adminUtils';
import adminService from '../services/adminService';
import { formatDateString } from '../utils/dateUtils';

const AdminTribeDetail = () => {
  const { tribeId } = useParams();
  const navigate = useNavigate();
  const { currentUser } = useAuth();
  
  const [tribe, setTribe] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  // Modal states
  const [showDeleteTribeModal, setShowDeleteTribeModal] = useState(false);
  const [showRemoveMemberModal, setShowRemoveMemberModal] = useState(false);
  const [selectedMember, setSelectedMember] = useState(null);
  const [actionInProgress, setActionInProgress] = useState(false);

  useEffect(() => {
    const fetchTribeDetails = async () => {
      try {
        const data = await adminService.getTribeDetailsAdmin(tribeId);
        setTribe(data);
        setLoading(false);
      } catch (err) {
        console.error('Error fetching tribe details:', err);
        setError('Failed to load tribe details. Please try again later.');
        setLoading(false);
      }
    };

    if (currentUser && isAdmin(currentUser) && tribeId) {
      fetchTribeDetails();
    } else {
      setLoading(false);
    }
  }, [currentUser, tribeId]);

  // Redirect non-admin users
  if (!currentUser || !isAdmin(currentUser)) {
    return <Navigate to="/" />;
  }

  if (loading) {
    return (
      <Container className="mt-5">
        <h2>Tribe Administration</h2>
        <div className="text-center mt-5">
          <p>Loading tribe details...</p>
        </div>
      </Container>
    );
  }

  if (error) {
    return (
      <Container className="mt-5">
        <h2>Tribe Administration</h2>
        <Alert variant="danger" className="mt-4">{error}</Alert>
        <Button variant="secondary" onClick={() => navigate('/admin')} className="mt-3">
          Back to Admin Panel
        </Button>
      </Container>
    );
  }

  if (!tribe) {
    return (
      <Container className="mt-5">
        <h2>Tribe Administration</h2>
        <Alert variant="warning" className="mt-4">Tribe not found</Alert>
        <Button variant="secondary" onClick={() => navigate('/admin')} className="mt-3">
          Back to Admin Panel
        </Button>
      </Container>
    );
  }

  const handleDeleteTribe = async () => {
    setActionInProgress(true);
    try {
      await adminService.deleteTribe(tribe.id);
      setShowDeleteTribeModal(false);
      navigate('/admin', { state: { message: `Tribe "${tribe.name}" was deleted successfully.` } });
    } catch (err) {
      console.error('Error deleting tribe:', err);
      setError(`Failed to delete tribe: ${err.message}`);
      setShowDeleteTribeModal(false);
      setActionInProgress(false);
    }
  };

  const handleRemoveMember = async () => {
    if (!selectedMember) return;
    
    setActionInProgress(true);
    try {
      await adminService.removeMemberFromTribe(tribe.id, selectedMember.user_id);
      
      // Update the local tribe state to remove the member
      setTribe({
        ...tribe,
        members: tribe.members.filter(m => m.user_id !== selectedMember.user_id)
      });
      
      setShowRemoveMemberModal(false);
      setSelectedMember(null);
      setActionInProgress(false);
    } catch (err) {
      console.error('Error removing member:', err);
      setError(`Failed to remove member: ${err.message}`);
      setShowRemoveMemberModal(false);
      setSelectedMember(null);
      setActionInProgress(false);
    }
  };

  return (
    <Container className="mt-5">
      <h2>Tribe Administration</h2>
      <Button variant="secondary" onClick={() => navigate('/admin')} className="mb-4">
        Back to Admin Panel
      </Button>
      
      {error && <Alert variant="danger" className="mt-3">{error}</Alert>}
      
      <Card className="mt-4">
        <Card.Header as="h5">Tribe Details</Card.Header>
        <Card.Body>
          <h3>{tribe.name}</h3>
          <p><strong>Type:</strong> {tribe.type}</p>
          <p><strong>Visibility:</strong> {tribe.visibility}</p>
          <p><strong>Description:</strong> {tribe.description || 'No description'}</p>
          <p><strong>Created:</strong> {formatDateString(tribe.created_at)}</p>
          <p><strong>Last Updated:</strong> {formatDateString(tribe.updated_at)}</p>
          <Button 
            variant="danger" 
            onClick={() => setShowDeleteTribeModal(true)}
            disabled={actionInProgress}
          >
            Delete Tribe
          </Button>
        </Card.Body>
      </Card>
      
      <Card className="mt-4">
        <Card.Header as="h5">Tribe Members</Card.Header>
        <Card.Body>
          {tribe.members && tribe.members.length > 0 ? (
            <Table responsive striped hover>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>User ID</th>
                  <th>Membership Type</th>
                  <th>Joined</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {tribe.members.map(member => (
                  <tr key={member.id}>
                    <td>{member.user?.name || member.display_name}</td>
                    <td>{member.user?.email || 'N/A'}</td>
                    <td>{member.user_id}</td>
                    <td>{member.membership_type}</td>
                    <td>{formatDateString(member.created_at)}</td>
                    <td>
                      <Button 
                        variant="danger" 
                        size="sm"
                        onClick={() => {
                          setSelectedMember(member);
                          setShowRemoveMemberModal(true);
                        }}
                        disabled={actionInProgress}
                      >
                        Remove
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </Table>
          ) : (
            <p>No members found.</p>
          )}
        </Card.Body>
      </Card>
      
      {/* Delete Tribe Confirmation Modal */}
      <Modal show={showDeleteTribeModal} onHide={() => setShowDeleteTribeModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Confirm Delete</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          Are you sure you want to delete the tribe "{tribe.name}"?
          This action cannot be undone.
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowDeleteTribeModal(false)} disabled={actionInProgress}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleDeleteTribe} disabled={actionInProgress}>
            {actionInProgress ? 'Deleting...' : 'Delete Tribe'}
          </Button>
        </Modal.Footer>
      </Modal>
      
      {/* Remove Member Confirmation Modal */}
      <Modal show={showRemoveMemberModal} onHide={() => setShowRemoveMemberModal(false)}>
        <Modal.Header closeButton>
          <Modal.Title>Confirm Member Removal</Modal.Title>
        </Modal.Header>
        <Modal.Body>
          Are you sure you want to remove {selectedMember?.user?.name || selectedMember?.display_name || 'this member'} from the tribe?
        </Modal.Body>
        <Modal.Footer>
          <Button variant="secondary" onClick={() => setShowRemoveMemberModal(false)} disabled={actionInProgress}>
            Cancel
          </Button>
          <Button variant="danger" onClick={handleRemoveMember} disabled={actionInProgress}>
            {actionInProgress ? 'Removing...' : 'Remove Member'}
          </Button>
        </Modal.Footer>
      </Modal>
    </Container>
  );
};

export default AdminTribeDetail; 
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import tribeService from '../services/tribeService';
import InviteMemberModal from './InviteMemberModal';
import PendingInvitationView from './PendingInvitationView';
import { useAuth } from '../contexts/AuthContext';

const DeleteConfirmationModal = ({ show, onClose, onConfirm, tribeName }) => {
  if (!show) return null;

  return (
    <div className="modal-backdrop">
      <div className="modal-content">
        <div className="modal-header">
          <h3>Delete Tribe</h3>
          <button onClick={onClose} className="close-btn">&times;</button>
        </div>
        
        <div className="modal-body">
          <p>Are you sure you want to delete the tribe "{tribeName}"?</p>
          <p className="text-danger">This action cannot be undone.</p>
        </div>
        
        <div className="modal-footer">
          <button type="button" onClick={onClose} className="btn btn-secondary">
            Cancel
          </button>
          <button type="button" onClick={onConfirm} className="btn btn-danger">
            Delete Tribe
          </button>
        </div>
      </div>
    </div>
  );
};

const TribeDetailScreen = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { currentUser } = useAuth();
  
  const [tribe, setTribe] = useState(null);
  const [members, setMembers] = useState([]);
  const [currentUserMember, setCurrentUserMember] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviter, setInviter] = useState(null);

  useEffect(() => {
    const fetchTribeData = async () => {
      try {
        setLoading(true);
        setError('');
        
        // Get tribe details
        const tribeData = await tribeService.getTribeById(id);
        setTribe(tribeData);
        
        // Get tribe members
        const membersData = await tribeService.getTribeMembers(id);
        setMembers(membersData);
        
        // Find the current user's membership
        if (currentUser) {
          const userMember = membersData.find(member => 
            member.user && member.user.firebase_uid === currentUser.uid
          );
          
          setCurrentUserMember(userMember);
          
          // If the user is a pending member, fetch the inviter's information
          if (userMember && userMember.membership_type === 'pending' && userMember.invited_by) {
            // Find the inviter in the members list
            const inviterMember = membersData.find(member => 
              member.user_id === userMember.invited_by
            );
            
            if (inviterMember) {
              setInviter(inviterMember.user);
            }
          }
        }
      } catch (err) {
        console.error('Error fetching tribe data:', err);
        setError('Failed to load tribe information');
      } finally {
        setLoading(false);
      }
    };

    fetchTribeData();
  }, [id, currentUser]);

  const handleDeleteClick = () => {
    setShowDeleteModal(true);
  };

  const handleDeleteConfirm = async () => {
    try {
      await tribeService.deleteTribe(id);
      navigate('/tribes');
    } catch (err) {
      console.error('Error deleting tribe:', err);
      setError('Failed to delete tribe');
      setShowDeleteModal(false);
    }
  };

  const handleBackClick = () => {
    navigate('/tribes');
  };

  const handleInviteClick = () => {
    setShowInviteModal(true);
  };

  const handleInviteMember = async (email) => {
    try {
      // In a real application, you would:
      // 1. Search for a user by email
      // 2. Get their ID
      // 3. Add them to the tribe
      // For now, we'll assume a dummy ID for demonstration
      const dummyUserId = "00000000-0000-0000-0000-000000000000"; // Replace with actual API call
      
      // Add the user to the tribe
      await tribeService.addTribeMember(id, dummyUserId);
      
      // Refresh the member list
      const membersData = await tribeService.getTribeMembers(id);
      setMembers(membersData);
      
      return dummyUserId;
    } catch (err) {
      console.error('Error inviting member:', err);
      throw err;
    }
  };

  if (loading) {
    return <div className="container">Loading tribe details...</div>;
  }

  if (error) {
    return (
      <div className="container">
        <div className="alert alert-danger">{error}</div>
        <button className="btn btn-primary" onClick={handleBackClick}>
          Back to Tribes
        </button>
      </div>
    );
  }

  if (!tribe) {
    return (
      <div className="container">
        <div className="alert alert-warning">Tribe not found</div>
        <button className="btn btn-primary" onClick={handleBackClick}>
          Back to Tribes
        </button>
      </div>
    );
  }

  // If the user is a pending member, show the invitation view
  if (currentUserMember && currentUserMember.membership_type === 'pending') {
    return (
      <div className="container tribe-detail-container">
        <PendingInvitationView 
          tribe={tribe} 
          invitedBy={inviter}
          invitedAt={currentUserMember.invited_at}
        />
      </div>
    );
  }

  // Otherwise show the regular tribe view
  return (
    <div className="container tribe-detail-container">
      <div className="tribe-header">
        <button className="btn btn-back" onClick={handleBackClick}>
          &larr; Back
        </button>
        <h1>{tribe.name}</h1>
        <div className="tribe-actions">
          <button className="btn btn-primary" onClick={handleInviteClick}>
            Invite Member
          </button>
          <button className="btn btn-danger" onClick={handleDeleteClick}>
            Delete Tribe
          </button>
        </div>
      </div>

      {tribe.description && (
        <div className="tribe-description">
          <p>{tribe.description}</p>
        </div>
      )}

      <div className="tribe-members-section">
        <h2>Members</h2>
        {members.length === 0 ? (
          <p>No members yet.</p>
        ) : (
          <div className="members-list">
            {members.map(member => (
              <div key={member.id} className="member-card">
                <div className="member-info">
                  <h3>{member.user.name || member.display_name}</h3>
                  <p>{member.user.email}</p>
                  {member.membership_type === 'pending' && (
                    <span className="badge pending-badge">Pending</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <DeleteConfirmationModal 
        show={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        onConfirm={handleDeleteConfirm}
        tribeName={tribe.name}
      />

      <InviteMemberModal
        show={showInviteModal}
        onClose={() => setShowInviteModal(false)}
        onInvite={handleInviteMember}
        tribeId={id}
      />
    </div>
  );
};

export default TribeDetailScreen; 
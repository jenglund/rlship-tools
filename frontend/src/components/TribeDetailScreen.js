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

        // If the user is pending, try to get inviter information
        if (tribeData.current_user_membership_type === 'pending' || tribeData.pending_invitation) {
          console.log('Pending invitation data:', tribeData);
          
          // First check if inviter info is directly in the tribe data
          if (tribeData.inviter) {
            console.log('Found inviter in tribe data:', tribeData.inviter);
            setInviter(tribeData.inviter);
          } else if (tribeData.invited_by) {
            // If we have invited_by ID but not the full inviter object, try to fetch the user
            console.log('Found invited_by ID, fetching user details:', tribeData.invited_by);
            try {
              // Get tribe members to find the inviter
              const membersData = await tribeService.getTribeMembers(id);
              const inviterMember = membersData.find(m => m.user_id === tribeData.invited_by);
              
              if (inviterMember && inviterMember.user) {
                console.log('Found inviter from members list:', inviterMember.user);
                setInviter({
                  id: tribeData.invited_by,
                  name: inviterMember.user.name || inviterMember.display_name,
                  ...inviterMember.user
                });
              }
            } catch (err) {
              console.error('Error fetching inviter details:', err);
            }
          } else {
            console.log('No inviter information found in tribe data');
            setInviter(null);
          }
          
          setMembers([]);
          setCurrentUserMember(null);
          return;
        }
        
        // Rest of the function unchanged for non-pending users
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
              console.log('Found inviter for pending member:', inviterMember.user);
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
  
  // We'll refresh the member list when the invite modal is closed
  const handleInviteModalClose = async () => {
    try {
      // Refresh the member list
      const membersData = await tribeService.getTribeMembers(id);
      setMembers(membersData);
    } catch (err) {
      console.error('Error refreshing members:', err);
    }
    setShowInviteModal(false);
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

  // If the user is a pending member, show the invitation view (use new backend contract)
  if (
    (tribe.current_user_membership_type === 'pending' || tribe.pending_invitation)
  ) {
    // Find inviter data in the tribe response
    let inviterData = inviter;
    
    // If inviter data didn't come from backend directly, try to find it in members list
    if (!inviterData && tribe.invited_by) {
      // Find inviter in members list if available
      if (members && members.length > 0) {
        const foundInviter = members.find(m => m.user_id === tribe.invited_by);
        if (foundInviter) {
          inviterData = {
            id: foundInviter.user_id,
            name: foundInviter.user?.name || foundInviter.display_name,
            ...foundInviter.user
          };
        }
      }
    }
    
    console.log('Rendering PendingInvitationView with:', {
      tribe: tribe,
      invitedBy: inviterData,
      invitedAt: tribe.invited_at,
      invitedById: tribe.invited_by
    });
    
    return (
      <div className="container tribe-detail-container">
        <PendingInvitationView 
          tribe={tribe} 
          invitedBy={inviterData}
          invitedAt={tribe.invited_at || (currentUserMember && currentUserMember.invited_at)}
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
                  <h3>
                    <a href={`/tribes/${tribe.id}/members/${member.user_id}`}>{member.user.name || member.display_name}</a>
                  </h3>
                  <p>{member.user.email}</p>
                  <p>
                    <strong>Invited by:</strong> {member.invited_by ? (
                      <a href={`/tribes/${tribe.id}/members/${member.invited_by}`}>
                        {(() => {
                          // Find the inviter in the members list
                          const inviter = members.find(m => m.user_id === member.invited_by);
                          return inviter ? (inviter.user?.name || inviter.display_name || 'View Profile') : 'View Profile';
                        })()}
                      </a>
                    ) : 'Unknown'}
                  </p>
                  <p>
                    <strong>Invited at:</strong> {member.invited_at ? new Date(member.invited_at).toLocaleDateString() : 'Unknown'}
                  </p>
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
        onClose={handleInviteModalClose}
        tribeId={id}
      />
    </div>
  );
};

export default TribeDetailScreen; 
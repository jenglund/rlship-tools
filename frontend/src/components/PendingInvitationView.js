import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import tribeService from '../services/tribeService';

const PendingInvitationView = ({ tribe, invitedBy, invitedAt }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [inviter, setInviter] = useState(invitedBy || tribe.inviter);
  const navigate = useNavigate();

  // Log all the invitation data we have
  useEffect(() => {
    console.log('Tribe data in PendingInvitationView:', {
      tribe: tribe,
      invitedBy: invitedBy,
      invitedAt: invitedAt,
      tribeInviter: tribe.inviter,
      tribeInvitedBy: tribe.invited_by
    });
    
    // Use tribe.inviter if available and we don't have invitedBy
    if (!invitedBy && tribe && tribe.inviter) {
      console.log('Using inviter from tribe object:', tribe.inviter);
      setInviter(tribe.inviter);
    }
  }, [tribe, invitedBy, invitedAt]);

  // If invitedBy is not provided, try to get it from the tribe object
  useEffect(() => {
    const fetchInviterIfNeeded = async () => {
      if (!inviter && tribe) {
        console.log('No inviter found, checking tribe for invited_by:', tribe);
        
        if (tribe.invited_by) {
          try {
            console.log('Tribe has invited_by ID, trying to fetch members:', tribe.invited_by);
            // Try to fetch members to find inviter
            const members = await tribeService.getTribeMembers(tribe.id);
            const foundInviter = members.find(m => m.user_id === tribe.invited_by);
            
            if (foundInviter) {
              console.log('Found inviter in members list:', foundInviter);
              setInviter({
                id: tribe.invited_by,
                name: foundInviter.user?.name || foundInviter.display_name,
                ...foundInviter.user
              });
            }
          } catch (err) {
            console.error('Error fetching inviter details:', err);
          }
        }
      }
    };
    
    fetchInviterIfNeeded();
  }, [tribe, inviter]);

  const handleAccept = async () => {
    try {
      setLoading(true);
      setError('');
      await tribeService.acceptInvitation(tribe.id);
      // Refresh the page to show the full tribe view
      window.location.reload();
    } catch (err) {
      console.error('Error accepting invitation:', err);
      setError('Failed to accept invitation. Please try again later.');
      setLoading(false);
    }
  };

  const handleReject = async () => {
    try {
      setLoading(true);
      setError('');
      await tribeService.rejectInvitation(tribe.id);
      // Navigate back to tribes list
      navigate('/tribes');
    } catch (err) {
      console.error('Error rejecting invitation:', err);
      setError('Failed to reject invitation. Please try again later.');
      setLoading(false);
    }
  };

  const handleReportAbuse = () => {
    alert('This feature is not yet implemented.');
  };

  // Format the invitation date
  const formattedDate = invitedAt ? new Date(invitedAt).toLocaleDateString() : 'Unknown date';
  
  // Get display name for inviter - try all possible sources
  const inviterName = 
    inviter?.name || 
    inviter?.display_name || 
    tribe?.inviter?.name || 
    'Unknown User';
  
  // Get the inviter ID from any available source
  const inviterId = 
    inviter?.id || 
    tribe?.inviter?.id || 
    tribe?.invited_by;

  return (
    <div className="pending-invitation-container">
      <div className="invitation-header">
        <h2>You have been invited to {tribe.name}!</h2>
      </div>

      {error && <div className="alert alert-danger">{error}</div>}

      <div className="invitation-details">
        <p>
          <strong>Invited by:</strong> {inviterId ? (
            <a href={`/tribes/${tribe.id}/members/${inviterId}`}>
              {inviterName}
            </a>
          ) : 'Unknown'}
        </p>
        <p>
          <strong>Invitation date:</strong> {formattedDate}
        </p>
      </div>

      <div className="invitation-actions">
        <button
          className="btn btn-success"
          onClick={handleAccept}
          disabled={loading}
        >
          {loading ? 'Processing...' : 'Accept'}
        </button>
        <button
          className="btn btn-danger"
          onClick={handleReject}
          disabled={loading}
        >
          {loading ? 'Processing...' : 'Reject'}
        </button>
        <button
          className="btn btn-secondary abuse-report-btn"
          onClick={handleReportAbuse}
          disabled={true}
        >
          Report Abuse
        </button>
      </div>
    </div>
  );
};

export default PendingInvitationView; 
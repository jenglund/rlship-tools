import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import tribeService from '../services/tribeService';

const PendingInvitationView = ({ tribe, invitedBy, invitedAt }) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const navigate = useNavigate();

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

  return (
    <div className="pending-invitation-container">
      <div className="invitation-header">
        <h2>You have been invited to {tribe.name}!</h2>
      </div>

      {error && <div className="alert alert-danger">{error}</div>}

      <div className="invitation-details">
        <p>
          <strong>Invited by:</strong> {invitedBy ? (
            <a href={`/tribes/${tribe.id}/members/${invitedBy.id}`}>
              {invitedBy.name || 'View Profile'}
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
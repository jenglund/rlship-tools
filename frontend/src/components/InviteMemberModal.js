import React, { useState } from 'react';
import userService from '../services/userService';
import tribeService from '../services/tribeService';

const InviteMemberModal = ({ show, onClose, tribeId }) => {
  const [email, setEmail] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!email.trim()) {
      setError('Email is required');
      return;
    }

    try {
      setLoading(true);
      setError('');
      
      // First search for user by email
      const user = await userService.searchUserByEmail(email);
      
      if (!user) {
        setError('No user found with this email address');
        return;
      }
      
      try {
        // Add user to tribe
        await tribeService.addTribeMember(tribeId, user.id);
        setSuccess(true);
        setTimeout(() => {
          setSuccess(false);
          setEmail('');
          onClose();
        }, 2000);
      } catch (addErr) {
        // Check if the error is because the user is already a member of the tribe
        if (addErr.response && addErr.response.status === 400) {
          setError('This user is already a member of the tribe');
        } else {
          setError(addErr.response?.data?.message || 'Failed to add member to tribe');
        }
      }
    } catch (err) {
      console.error('Error inviting member:', err);
      setError(err.response?.data?.message || 'Failed to invite member');
    } finally {
      setLoading(false);
    }
  };

  if (!show) return null;

  return (
    <div className="modal-backdrop">
      <div className="modal-content">
        <div className="modal-header">
          <h3>Invite Member</h3>
          <button onClick={onClose} className="close-btn">&times;</button>
        </div>
        
        <form onSubmit={handleSubmit}>
          {error && <div className="alert alert-danger">{error}</div>}
          {success && <div className="alert alert-success">Invitation sent successfully!</div>}
          
          <div className="form-group">
            <label htmlFor="email">Email Address</label>
            <input
              type="email"
              id="email"
              className="form-control"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Enter email address"
              required
            />
            <small className="form-text text-muted">
              Enter the email address of the person you want to invite.
            </small>
          </div>
          
          <div className="modal-footer">
            <button type="button" onClick={onClose} className="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading || success}>
              {loading ? 'Sending...' : 'Send Invitation'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default InviteMemberModal; 
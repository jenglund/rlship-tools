import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import tribeService from '../services/tribeService';

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
  
  const [tribe, setTribe] = useState(null);
  const [members, setMembers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showDeleteModal, setShowDeleteModal] = useState(false);

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
      } catch (err) {
        console.error('Error fetching tribe data:', err);
        setError('Failed to load tribe information');
      } finally {
        setLoading(false);
      }
    };

    fetchTribeData();
  }, [id]);

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

  return (
    <div className="container tribe-detail-container">
      <div className="tribe-header">
        <button className="btn btn-back" onClick={handleBackClick}>
          &larr; Back
        </button>
        <h1>{tribe.name}</h1>
        <button className="btn btn-danger" onClick={handleDeleteClick}>
          Delete Tribe
        </button>
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
                <h3>{member.user.name || member.display_name}</h3>
                <p>{member.user.email}</p>
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
    </div>
  );
};

export default TribeDetailScreen; 
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import tribeService from '../services/tribeService';

const CreateTribeModal = ({ show, onClose, onCreate }) => {
  const [tribeName, setTribeName] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (show) {
      setTribeName('');
      setError('');
    }
  }, [show]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!tribeName.trim()) {
      setError('Tribe name is required');
      return;
    }

    try {
      setLoading(true);
      await onCreate(tribeName);
      onClose();
    } catch (err) {
      setError('Failed to create tribe');
    } finally {
      setLoading(false);
    }
  };

  if (!show) return null;

  return (
    <div className="modal-backdrop">
      <div className="modal-content">
        <div className="modal-header">
          <h3>Create New Tribe</h3>
          <button onClick={onClose} className="close-btn">&times;</button>
        </div>
        
        <form onSubmit={handleSubmit}>
          {error && <div className="alert alert-danger">{error}</div>}
          
          <div className="form-group">
            <label htmlFor="tribeName">Tribe Name</label>
            <input
              type="text"
              id="tribeName"
              className="form-control"
              value={tribeName}
              onChange={(e) => setTribeName(e.target.value)}
              placeholder="Enter tribe name"
              required
            />
          </div>
          
          <div className="modal-footer">
            <button type="button" onClick={onClose} className="btn btn-secondary">
              Cancel
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? 'Creating...' : 'Create Tribe'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

const TribesListScreen = () => {
  const [tribes, setTribes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const { currentUser, logout } = useAuth();
  const navigate = useNavigate();

  const fetchTribes = async () => {
    try {
      setLoading(true);
      const tribesData = await tribeService.getUserTribes();
      setTribes(tribesData);
    } catch (error) {
      console.error('Error fetching tribes:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTribes();
  }, []);

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  const handleCreateTribe = async (name) => {
    const tribeData = {
      name: name,
      type: 'custom',
      visibility: 'private',
      description: ''
    };
    
    const newTribe = await tribeService.createTribe(tribeData);
    await fetchTribes(); // Refresh the list
    
    // Navigate to the new tribe's detail page
    navigate(`/tribes/${newTribe.id}`);
  };

  const handleTribeClick = (tribeId) => {
    navigate(`/tribes/${tribeId}`);
  };

  return (
    <div className="container tribes-container">
      <div className="tribes-header">
        <h1>My Tribes</h1>
        <div className="header-actions">
          <button 
            className="btn btn-primary" 
            onClick={() => setShowCreateModal(true)}
          >
            Create Tribe
          </button>
          <button className="btn btn-logout" onClick={handleLogout}>
            Log Out
          </button>
        </div>
      </div>

      {loading ? (
        <div className="loading">Loading tribes...</div>
      ) : (
        <>
          {tribes.length === 0 ? (
            <div className="empty-state">
              <p>You don't have any tribes yet.</p>
              <button 
                className="btn btn-primary" 
                onClick={() => setShowCreateModal(true)}
              >
                Create Your First Tribe
              </button>
            </div>
          ) : (
            <div className="tribes-list">
              {tribes.map(tribe => (
                <div 
                  key={tribe.id} 
                  className="tribe-card"
                  onClick={() => handleTribeClick(tribe.id)}
                >
                  <h3>{tribe.name}</h3>
                  {tribe.description && <p>{tribe.description}</p>}
                  <div className="tribe-meta">
                    <small>
                      {tribe.members ? tribe.members.length : 0} member{(tribe.members && tribe.members.length !== 1) ? 's' : ''}
                    </small>
                  </div>
                </div>
              ))}
            </div>
          )}
        </>
      )}

      <CreateTribeModal 
        show={showCreateModal} 
        onClose={() => setShowCreateModal(false)} 
        onCreate={handleCreateTribe}
      />
    </div>
  );
};

export default TribesListScreen; 
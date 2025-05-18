import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { getUserTribes } from '../services/tribeService';

const TribesListScreen = () => {
  const [tribes, setTribes] = useState([]);
  const [loading, setLoading] = useState(true);
  const { currentUser, logout } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    const fetchTribes = async () => {
      try {
        const tribesData = await getUserTribes(currentUser.id);
        setTribes(tribesData);
      } catch (error) {
        console.error('Error fetching tribes:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchTribes();
  }, [currentUser]);

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  if (loading) {
    return <div className="container">Loading tribes...</div>;
  }

  return (
    <div className="container tribes-container">
      <div className="tribes-header">
        <h1>My Tribes</h1>
        <div>
          <button className="btn btn-logout" onClick={handleLogout}>
            Log Out
          </button>
        </div>
      </div>

      {tribes.length === 0 ? (
        <p>You don't have any tribes yet.</p>
      ) : (
        <div className="tribes-list">
          {tribes.map(tribe => (
            <div key={tribe.id} className="tribe-card">
              <h3>{tribe.name}</h3>
              <p>{tribe.description}</p>
              <div>
                <small>Members: {tribe.memberCount}</small>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default TribesListScreen; 
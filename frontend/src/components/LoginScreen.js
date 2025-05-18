import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const LoginScreen = () => {
  const { login, devUsers } = useAuth();
  const navigate = useNavigate();

  const handleLogin = (userId) => {
    const success = login(userId);
    if (success) {
      navigate('/tribes');
    }
  };

  return (
    <div className="container">
      <div className="login-container">
        <h2>Log in to Tribe</h2>
        <div className="btn-container">
          {devUsers.map(user => (
            <button
              key={user.id}
              className="btn btn-secondary"
              onClick={() => handleLogin(user.id)}
            >
              Login as {user.name}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
};

export default LoginScreen; 
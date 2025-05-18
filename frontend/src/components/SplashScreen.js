import React from 'react';
import { Link } from 'react-router-dom';

const SplashScreen = () => {
  return (
    <div className="splash-screen">
      <h1>Welcome to Tribe</h1>
      <p>
        A place for groups to share and plan activities together. Connect with your tribes and organize your shared experiences.
      </p>
      <Link to="/login">
        <button className="btn btn-primary">Log In</button>
      </Link>
    </div>
  );
};

export default SplashScreen; 
import axios from 'axios';
import config from '../config';

const API_URL = config.API_URL;

// Configure axios to include the dev email header for all requests in development
axios.interceptors.request.use(config => {
  const user = localStorage.getItem('user');
  if (user) {
    const userData = JSON.parse(user);
    if (userData.email && userData.email.match(/^dev_user\d+@gonetribal\.com$/)) {
      config.headers['X-Dev-Email'] = userData.email;
    }
  }
  return config;
});

const authService = {
  // Check if email exists in the system
  checkEmail: async (email) => {
    try {
      const response = await axios.post(`${API_URL}/users/check-email`, { email });
      return response.data;
    } catch (error) {
      console.error('Error checking email:', error);
      throw error;
    }
  },

  // Register a new user
  register: async (userData) => {
    try {
      const response = await axios.post(`${API_URL}/users/register`, userData);
      // Store user data in localStorage
      localStorage.setItem('user', JSON.stringify(response.data.data));
      return response.data.data;
    } catch (error) {
      console.error('Error registering user:', error);
      throw error;
    }
  },

  // Login existing user
  login: async (email) => {
    try {
      // For dev users, just fetch user info with the dev email header
      if (email.match(/^dev_user\d+@gonetribal\.com$/)) {
        // Set dev email header for this request
        const headers = { 'X-Dev-Email': email };
        const response = await axios.get(`${API_URL}/users/me`, { headers });
        
        // Generate a dev token and add it to the user data
        const userData = {
          ...response.data.data,
          token: `dev:${email}` // Add a dev token format that backend can recognize
        };
        
        // Store user data in localStorage
        localStorage.setItem('user', JSON.stringify(userData));
        return userData;
      } else {
        // Normal login flow for non-dev users
        const response = await axios.post(`${API_URL}/users/login`, { email });
        localStorage.setItem('user', JSON.stringify(response.data.data));
        return response.data.data;
      }
    } catch (error) {
      console.error('Error logging in:', error);
      throw error;
    }
  },

  // Logout user
  logout: () => {
    localStorage.removeItem('user');
  },

  // Get current user
  getCurrentUser: () => {
    const user = localStorage.getItem('user');
    return user ? JSON.parse(user) : null;
  },

  // Update user profile
  updateProfile: async (userData) => {
    try {
      const currentUser = authService.getCurrentUser();
      const token = currentUser ? currentUser.token : null;
      
      const response = await axios.put(
        `${API_URL}/users/me`,
        userData
      );
      
      // Update stored user data
      const updatedUser = { ...currentUser, ...response.data.data };
      localStorage.setItem('user', JSON.stringify(updatedUser));
      
      return updatedUser;
    } catch (error) {
      console.error('Error updating profile:', error);
      throw error;
    }
  }
};

export default authService; 
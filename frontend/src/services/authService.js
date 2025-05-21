import axios from 'axios';
import config from '../config';

const API_URL = config.API_URL;

// Configure axios to include auth headers for all requests 
axios.interceptors.request.use(config => {
  const user = localStorage.getItem('user');
  if (user) {
    try {
      const userData = JSON.parse(user);
      
      // Add token for authorization if available
      if (userData.token) {
        config.headers['Authorization'] = `Bearer ${userData.token}`;
        console.debug(`Adding Authorization header: Bearer ${userData.token.substring(0, 10)}...`);
      }
      
      // Add email header for all users in development mode
      if (userData.email) {
        config.headers['X-Dev-Email'] = userData.email;
        console.debug(`Adding X-Dev-Email header: ${userData.email}`);
      }
    } catch (error) {
      console.error('Error parsing user data from localStorage:', error);
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
      console.debug(`Logging in user: ${email}`);
      
      // Set email header for this request
      const headers = { 'X-Dev-Email': email };
      
      try {
        // Try to get user info directly first
        const response = await axios.get(`${API_URL}/users/me`, { headers });
        
        // Generate a token and add it to the user data
        const userData = {
          ...response.data.data,
          token: `dev:${email}` // Add a token format that backend can recognize
        };
        
        // Store user data in localStorage
        localStorage.setItem('user', JSON.stringify(userData));
        console.debug(`Stored user data: ${JSON.stringify(userData, null, 2)}`);
        return userData;
      } catch (err) {
        // If user doesn't exist yet, fall back to login endpoint
        console.debug(`User not found, using login endpoint: ${err}`);
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
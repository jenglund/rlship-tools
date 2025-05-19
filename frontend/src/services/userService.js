import axios from 'axios';
import authService from './authService';
import config from '../config';

const API_URL = config.API_URL;

// Helper function to get the authorization header
const getAuthHeader = () => {
  const user = authService.getCurrentUser();
  return user ? { Authorization: `Bearer ${user.token}` } : {};
};

const userService = {
  // Search for user by email
  searchUserByEmail: async (email) => {
    try {
      const response = await axios.post(
        `${API_URL}/users/search`,
        { email },
        { headers: getAuthHeader() }
      );
      return response.data.data;
    } catch (error) {
      // Return null if user not found (instead of throwing an error)
      if (error.response && error.response.status === 404) {
        return null;
      }
      console.error(`Error searching for user by email:`, error);
      throw error;
    }
  },

  // Get user by ID
  getUserById: async (userId) => {
    try {
      const response = await axios.get(`${API_URL}/users/${userId}`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error getting user ${userId}:`, error);
      throw error;
    }
  }
};

export default userService; 
import axios from 'axios';
import authService from './authService';
import config from '../config';

const API_URL = config.API_URL;

// Helper function to get the authorization header
const getAuthHeader = () => {
  const user = authService.getCurrentUser();
  return user ? { Authorization: `Bearer ${user.token}` } : {};
};

const tribeService = {
  // Get all tribes for the current user
  getUserTribes: async () => {
    try {
      const response = await axios.get(`${API_URL}/tribes/my`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error('Error fetching user tribes:', error);
      throw error;
    }
  },

  // Get a single tribe by ID
  getTribeById: async (tribeId) => {
    try {
      const response = await axios.get(`${API_URL}/tribes/${tribeId}`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching tribe ${tribeId}:`, error);
      throw error;
    }
  },

  // Create a new tribe
  createTribe: async (tribeData) => {
    try {
      const response = await axios.post(`${API_URL}/tribes`, tribeData, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error('Error creating tribe:', error);
      throw error;
    }
  },

  // Update a tribe
  updateTribe: async (tribeId, tribeData) => {
    try {
      const response = await axios.put(`${API_URL}/tribes/${tribeId}`, tribeData, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error updating tribe ${tribeId}:`, error);
      throw error;
    }
  },

  // Delete a tribe
  deleteTribe: async (tribeId) => {
    try {
      await axios.delete(`${API_URL}/tribes/${tribeId}`, {
        headers: getAuthHeader()
      });
      return true;
    } catch (error) {
      console.error(`Error deleting tribe ${tribeId}:`, error);
      throw error;
    }
  },

  // Get all members of a tribe
  getTribeMembers: async (tribeId) => {
    try {
      const response = await axios.get(`${API_URL}/tribes/${tribeId}/members`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching members for tribe ${tribeId}:`, error);
      throw error;
    }
  },

  // Add a member to a tribe
  addTribeMember: async (tribeId, userId) => {
    try {
      const response = await axios.post(
        `${API_URL}/tribes/${tribeId}/members`,
        { user_id: userId },
        { headers: getAuthHeader() }
      );
      return response.data.data;
    } catch (error) {
      console.error(`Error adding member to tribe ${tribeId}:`, error);
      throw error;
    }
  },

  // Remove a member from a tribe
  removeTribeMember: async (tribeId, userId) => {
    try {
      await axios.delete(`${API_URL}/tribes/${tribeId}/members/${userId}`, {
        headers: getAuthHeader()
      });
      return true;
    } catch (error) {
      console.error(`Error removing member from tribe ${tribeId}:`, error);
      throw error;
    }
  }
};

export default tribeService; 
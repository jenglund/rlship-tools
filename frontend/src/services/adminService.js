import axios from 'axios';
import config from '../config';

const API_URL = config.API_URL;

const adminService = {
  // Get recent tribes for admin panel (most recent 20)
  getRecentTribes: async () => {
    try {
      const response = await axios.get(`${API_URL}/tribes?limit=20&sort=created_at:desc`);
      return response.data.data;
    } catch (error) {
      console.error('Error fetching recent tribes:', error);
      throw error;
    }
  },

  // Get tribe details including members for admin
  getTribeDetailsAdmin: async (tribeId) => {
    try {
      // Get tribe details
      const tribeResponse = await axios.get(`${API_URL}/tribes/${tribeId}`);
      
      // Get tribe members with user details
      const membersResponse = await axios.get(`${API_URL}/tribes/${tribeId}/members?include_users=true`);
      
      // Combine the data
      return {
        ...tribeResponse.data.data,
        members: membersResponse.data.data || []
      };
    } catch (error) {
      console.error('Error fetching tribe details for admin:', error);
      throw error;
    }
  },

  // Delete a tribe (admin function)
  deleteTribe: async (tribeId) => {
    try {
      await axios.delete(`${API_URL}/tribes/${tribeId}`);
      return true;
    } catch (error) {
      console.error('Error deleting tribe:', error);
      throw error;
    }
  },

  // Remove a member from a tribe (admin function)
  removeMemberFromTribe: async (tribeId, userId) => {
    try {
      await axios.delete(`${API_URL}/tribes/${tribeId}/members/${userId}`);
      return true;
    } catch (error) {
      console.error('Error removing member from tribe:', error);
      throw error;
    }
  }
};

export default adminService; 
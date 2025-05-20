import axios from 'axios';
import authService from './authService';
import config from '../config';

const API_URL = config.API_URL;

// Helper function to get the authorization header
const getAuthHeader = () => {
  const user = authService.getCurrentUser();
  return user ? { Authorization: `Bearer ${user.token}` } : {};
};

const listService = {
  // Core list management methods
  
  // Create a new list
  createList: async (listData) => {
    try {
      const response = await axios.post(`${API_URL}/v1/lists`, listData, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error('Error creating list:', error);
      throw error;
    }
  },

  // Get lists owned by the current user
  getUserLists: async (userID) => {
    try {
      const response = await axios.get(`${API_URL}/v1/lists/user/${userID}`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching user lists for user ${userID}:`, error);
      throw error;
    }
  },

  // Get a single list by ID
  getListDetails: async (id) => {
    try {
      const response = await axios.get(`${API_URL}/v1/lists/${id}`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching list ${id}:`, error);
      throw error;
    }
  },

  // Update an existing list
  updateList: async (id, listData) => {
    try {
      const response = await axios.put(`${API_URL}/v1/lists/${id}`, listData, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error updating list ${id}:`, error);
      throw error;
    }
  },

  // Delete a list
  deleteList: async (id) => {
    try {
      await axios.delete(`${API_URL}/v1/lists/${id}`, {
        headers: getAuthHeader()
      });
      return true;
    } catch (error) {
      console.error(`Error deleting list ${id}:`, error);
      throw error;
    }
  },

  // List item methods
  
  // Get items in a list
  getListItems: async (listID) => {
    try {
      const response = await axios.get(`${API_URL}/v1/lists/${listID}/items`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching items for list ${listID}:`, error);
      throw error;
    }
  },

  // Add a new item to a list
  addListItem: async (listID, itemData) => {
    try {
      const response = await axios.post(`${API_URL}/v1/lists/${listID}/items`, itemData, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error adding item to list ${listID}:`, error);
      throw error;
    }
  },

  // Update an existing list item
  updateListItem: async (listID, itemID, itemData) => {
    try {
      const response = await axios.put(`${API_URL}/v1/lists/${listID}/items/${itemID}`, itemData, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error updating item ${itemID} in list ${listID}:`, error);
      throw error;
    }
  },

  // Delete an item from a list
  deleteListItem: async (listID, itemID) => {
    try {
      await axios.delete(`${API_URL}/v1/lists/${listID}/items/${itemID}`, {
        headers: getAuthHeader()
      });
      return true;
    } catch (error) {
      console.error(`Error removing item ${itemID} from list ${listID}:`, error);
      throw error;
    }
  },

  // List sharing methods
  
  // Get all shares for a list
  getListShares: async (listID) => {
    try {
      const response = await axios.get(`${API_URL}/v1/lists/${listID}/shares`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching shares for list ${listID}:`, error);
      throw error;
    }
  },

  // Share a list with a tribe
  shareListWithTribe: async (listID, tribeID) => {
    try {
      const response = await axios.post(`${API_URL}/v1/lists/${listID}/share/${tribeID}`, {}, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error sharing list ${listID} with tribe ${tribeID}:`, error);
      throw error;
    }
  },

  // Unshare a list with a tribe
  unshareListWithTribe: async (listID, tribeID) => {
    try {
      await axios.delete(`${API_URL}/v1/lists/${listID}/share/${tribeID}`, {
        headers: getAuthHeader()
      });
      return true;
    } catch (error) {
      console.error(`Error unsharing list ${listID} from tribe ${tribeID}:`, error);
      throw error;
    }
  },

  // Get lists shared with a tribe
  getSharedListsForTribe: async (tribeID) => {
    try {
      const response = await axios.get(`${API_URL}/v1/lists/shared/${tribeID}`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching shared lists for tribe ${tribeID}:`, error);
      throw error;
    }
  }
};

export default listService; 
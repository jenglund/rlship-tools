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
      console.log(`Creating list with data:`, listData);
      console.log(`API URL: ${API_URL}/v1/lists`);
      const response = await axios.post(`${API_URL}/v1/lists`, listData, {
        headers: getAuthHeader()
      });
      console.log(`Successfully created list:`, response.data);
      return response.data.data;
    } catch (error) {
      console.error('Error creating list:', error);
      console.error('Request details:', {
        url: `${API_URL}/v1/lists`,
        data: listData,
        headers: getAuthHeader()
      });
      if (error.response) {
        console.error('Response status:', error.response.status);
        console.error('Response data:', error.response.data);
      }
      throw error;
    }
  },

  // Get lists owned by the current user
  getUserLists: async (userID) => {
    try {
      console.log(`Fetching lists for user: ${userID}`);
      console.log(`API URL: ${API_URL}/v1/lists/user/${userID}`);
      const response = await axios.get(`${API_URL}/v1/lists/user/${userID}`, {
        headers: getAuthHeader()
      });
      console.log(`Successfully fetched user lists:`, response.data);
      
      // Make sure we always return an array
      if (response.data && response.data.data) {
        return Array.isArray(response.data.data) ? response.data.data : [];
      }
      return [];
    } catch (error) {
      console.error(`Error fetching user lists for user ${userID}:`, error);
      console.error('Request details:', {
        url: `${API_URL}/v1/lists/user/${userID}`,
        headers: getAuthHeader()
      });
      if (error.response) {
        console.error('Response status:', error.response.status);
        console.error('Response data:', error.response.data);
      }
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

  // Get list owners
  getListOwners: async (listID) => {
    try {
      const response = await axios.get(`${API_URL}/v1/lists/${listID}/owners`, {
        headers: getAuthHeader()
      });
      return response.data.data;
    } catch (error) {
      console.error(`Error fetching owners for list ${listID}:`, error);
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
      console.log(`Fetching shares for list: ${listID}`);
      console.log(`API URL: ${API_URL}/v1/lists/${listID}/shares`);
      console.log(`Auth headers:`, getAuthHeader());
      
      const response = await axios.get(`${API_URL}/v1/lists/${listID}/shares`, {
        headers: getAuthHeader()
      });
      
      console.log(`Successfully fetched list shares:`, response.data);
      
      // If we have share data and it doesn't already include tribe names
      if (response.data.data && response.data.data.length > 0) {
        // Fetch tribe details for each share to get the tribe names
        const sharesWithNames = await Promise.all(
          response.data.data.map(async (share) => {
            try {
              // Only fetch tribe details if we don't already have the name
              if (!share.name && !share.tribeName && share.tribeId) {
                const tribeResponse = await axios.get(`${API_URL}/tribes/${share.tribeId}`, {
                  headers: getAuthHeader()
                });
                
                if (tribeResponse.data && tribeResponse.data.data) {
                  // Add the tribe name to the share object
                  return {
                    ...share,
                    name: tribeResponse.data.data.name,
                    tribeName: tribeResponse.data.data.name
                  };
                }
              }
              return share;
            } catch (err) {
              console.warn(`Couldn't fetch tribe details for ${share.tribeId}:`, err);
              return share;
            }
          })
        );
        
        return sharesWithNames;
      }
      
      // Handle case where data is returned but might be empty
      return response.data.data || [];
    } catch (error) {
      console.error(`Error fetching shares for list ${listID}:`, error);
      console.error('Request details:', {
        url: `${API_URL}/v1/lists/${listID}/shares`,
        headers: getAuthHeader()
      });
      
      if (error.response) {
        console.error('Response status:', error.response.status);
        console.error('Response data:', error.response.data);
        
        // If we get a 403 or 404, throw the error rather than returning an empty array
        // This will prevent infinite loops of retries
        if (error.response.status === 403 || error.response.status === 404) {
          throw error;
        }
      }
      
      // For other errors, return empty array to avoid breaking the UI
      return [];
    }
  },

  // Share a list with a tribe
  shareListWithTribe: async (listID, tribeID) => {
    try {
      console.log(`Sharing list ${listID} with tribe ${tribeID}`);
      console.log(`API URL: ${API_URL}/v1/lists/${listID}/share/${tribeID}`);
      const response = await axios.post(`${API_URL}/v1/lists/${listID}/share/${tribeID}`, {}, {
        headers: getAuthHeader()
      });
      console.log(`Successfully shared list:`, response.data);
      return response.data.data;
    } catch (error) {
      console.error(`Error sharing list ${listID} with tribe ${tribeID}:`, error);
      console.error('Request details:', {
        url: `${API_URL}/v1/lists/${listID}/share/${tribeID}`,
        headers: getAuthHeader()
      });
      if (error.response) {
        console.error('Response status:', error.response.status);
        console.error('Response data:', error.response.data);
      }
      throw error;
    }
  },

  // Unshare a list with a tribe
  unshareListWithTribe: async (listID, tribeID) => {
    try {
      console.log(`Unsharing list ${listID} from tribe ${tribeID}`);
      console.log(`API URL: ${API_URL}/v1/lists/${listID}/share/${tribeID}`);
      await axios.delete(`${API_URL}/v1/lists/${listID}/share/${tribeID}`, {
        headers: getAuthHeader()
      });
      console.log(`Successfully unshared list`);
      return true;
    } catch (error) {
      console.error(`Error unsharing list ${listID} from tribe ${tribeID}:`, error);
      console.error('Request details:', {
        url: `${API_URL}/v1/lists/${listID}/share/${tribeID}`,
        headers: getAuthHeader()
      });
      if (error.response) {
        console.error('Response status:', error.response.status);
        console.error('Response data:', error.response.data);
      }
      throw error;
    }
  },

  // Get lists shared with a tribe
  getSharedListsForTribe: async (tribeID) => {
    try {
      const response = await axios.get(`${API_URL}/v1/lists/shared/${tribeID}`, {
        headers: getAuthHeader()
      });
      
      // Make sure we always return an array
      if (response.data && response.data.data) {
        return Array.isArray(response.data.data) ? response.data.data : [];
      }
      return [];
    } catch (error) {
      console.error(`Error fetching shared lists for tribe ${tribeID}:`, error);
      throw error;
    }
  }
};

export default listService; 
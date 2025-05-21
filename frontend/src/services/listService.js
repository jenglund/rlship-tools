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
      console.log(`Successfully created list. Full response:`, response.data);
      
      // Ensure we're returning just the data object, not the entire success response
      if (response.data && response.data.success === true && response.data.data) {
        console.log(`Extracted list data from API response:`, response.data.data);
        return response.data.data;
      } else if (response.data && response.data.id) {
        // Some APIs return the object directly
        console.log(`API returned list directly:`, response.data);
        return response.data;
      }
      
      // If we can't clearly identify a list object, log a warning
      console.warn(`Unexpected API response format from createList:`, response.data);
      
      // If we have a data property, return it as a last resort
      if (response.data && response.data.data) {
        return response.data.data;
      }
      
      // Fall back to the response.data in case the API structure changes
      return response.data;
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
  getUserLists: async (userID, timestamp = null) => {
    try {
      console.log(`Fetching lists for user: ${userID}${timestamp ? ' (with cache busting)' : ''}`);
      
      // Add timestamp for cache busting if provided
      const cacheBuster = timestamp ? `?_t=${timestamp}` : '';
      const url = `${API_URL}/v1/lists/user/${userID}${cacheBuster}`;
      console.log(`API URL: ${url}`);
      
      const response = await axios.get(url, {
        headers: getAuthHeader()
      });
      
      console.log(`Successfully fetched user lists:`, response.data);
      
      // Based on curl output, the API uses the structure:
      // { data: { success: true, data: [ ...lists ] } }
      
      if (response.data && response.data.data) {
        const innerResponse = response.data.data;
        
        if (innerResponse && innerResponse.success === true && Array.isArray(innerResponse.data)) {
          console.log(`Found nested lists array with ${innerResponse.data.length} items`);
          return innerResponse.data;
        }
        
        if (innerResponse && Array.isArray(innerResponse)) {
          console.log(`Found array directly in data with ${innerResponse.length} items`);
          return innerResponse;
        }
      }
      
      // If above doesn't extract lists, try other common patterns
      if (response.data && response.data.success === true && Array.isArray(response.data.data)) {
        console.log(`Found lists array in response.data with ${response.data.data.length} items`);
        return response.data.data;
      }
      
      if (Array.isArray(response.data)) {
        console.log(`Response is a direct array with ${response.data.length} items`);
        return response.data;
      }
      
      // Log the full structure for diagnosis
      console.warn('Could not extract lists array from API response with expected structure');
      console.log('Response structure:', JSON.stringify(response.data, null, 2));
      
      return [];
    } catch (error) {
      console.error(`Error fetching user lists for user ${userID}:`, error);
      if (error.response) {
        console.error('Error response status:', error.response.status);
        console.error('Error response data:', error.response.data);
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
      console.log(`Fetching items for list: ${listID}`);
      console.log(`API URL: ${API_URL}/v1/lists/${listID}/items`);
      
      const response = await axios.get(`${API_URL}/v1/lists/${listID}/items`, {
        headers: getAuthHeader()
      });
      
      console.log(`Successfully fetched list items:`, response.data);
      
      // Handle both cases where API returns the array directly or as response.data.data
      if (response.data && response.data.data) {
        // API returned {data: [...]} structure
        return Array.isArray(response.data.data) ? response.data.data : [];
      } else if (Array.isArray(response.data)) {
        // API returned the array directly
        return response.data;
      }
      
      // Fallback to empty array
      console.warn('Unexpected response format from list items API:', response.data);
      return [];
    } catch (error) {
      console.error(`Error fetching items for list ${listID}:`, error);
      if (error.response) {
        console.error('Response status:', error.response.status);
        console.error('Response data:', error.response.data);
      }
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
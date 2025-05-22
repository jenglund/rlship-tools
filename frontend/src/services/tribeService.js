import axios from 'axios';
import authService from './authService';
import config from '../config';
import { getAuthHeader } from '../utils/authUtils';

const API_URL = config.API_URL;

// Simple cache to prevent excessive identical API calls
const apiCache = {
  getUserTribes: {
    data: null,
    timestamp: 0,
    inProgress: false
  }
};

// Cache expiration in milliseconds (5 seconds)
const CACHE_EXPIRY = 5000;

// Helper function to get the authorization header
const getAuthHeader = () => {
  const user = authService.getCurrentUser();
  const headers = user ? { Authorization: `Bearer ${user.token}` } : {};
  console.debug('Auth headers for API request:', headers);
  return headers;
};

const tribeService = {
  // Get all tribes for the current user
  getUserTribes: async () => {
    // Check if we already have a request in progress
    if (apiCache.getUserTribes.inProgress) {
      console.log('getUserTribes: Request already in progress, returning cached data');
      return apiCache.getUserTribes.data || [];
    }
    
    // Check if we have valid cached data
    const now = new Date().getTime();
    if (
      apiCache.getUserTribes.data && 
      now - apiCache.getUserTribes.timestamp < CACHE_EXPIRY
    ) {
      console.log('getUserTribes: Returning cached data');
      return apiCache.getUserTribes.data;
    }
    
    try {
      console.log('getUserTribes: Fetching from API');
      apiCache.getUserTribes.inProgress = true;
      
      // Add cache buster to prevent browser caching
      const cacheBuster = new Date().getTime();
      const response = await axios.get(`${API_URL}/tribes/my?t=${cacheBuster}`, {
        headers: getAuthHeader()
      });
      
      // Save to cache
      apiCache.getUserTribes.data = response.data.data || [];
      apiCache.getUserTribes.timestamp = now;
      
      console.log(`getUserTribes: Successfully fetched ${apiCache.getUserTribes.data.length} tribes`);
      return apiCache.getUserTribes.data;
    } catch (error) {
      console.error('Error fetching user tribes:', error);
      throw error;
    } finally {
      apiCache.getUserTribes.inProgress = false;
    }
  },

  // Reset the cache (useful after creating/deleting tribes)
  resetTribesCache: () => {
    console.log('Resetting tribes cache');
    apiCache.getUserTribes.data = null;
    apiCache.getUserTribes.timestamp = 0;
    apiCache.getUserTribes.inProgress = false;
  },

  // Get a single tribe by ID
  getTribeById: async (tribeId) => {
    try {
      const response = await axios.get(`${API_URL}/tribes/${tribeId}`, {
        headers: getAuthHeader()
      });
      console.log('Tribe data response:', response.data);
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
  addTribeMember: async (tribeId, userId, force = false) => {
    try {
      const response = await axios.post(
        `${API_URL}/tribes/${tribeId}/members`,
        { user_id: userId, force },
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
  },

  // Respond to a tribe invitation (accept or reject)
  respondToInvitation: async (tribeId, action) => {
    try {
      const response = await axios.post(
        `${API_URL}/tribes/${tribeId}/respond`,
        { action },
        { headers: getAuthHeader() }
      );
      return response.data.data;
    } catch (error) {
      console.error(`Error responding to invitation for tribe ${tribeId}:`, error);
      throw error;
    }
  },

  // Accept a tribe invitation
  acceptInvitation: async (tribeId) => {
    return tribeService.respondToInvitation(tribeId, 'accept');
  },

  // Reject a tribe invitation
  rejectInvitation: async (tribeId) => {
    return tribeService.respondToInvitation(tribeId, 'reject');
  }
};

export default tribeService; 
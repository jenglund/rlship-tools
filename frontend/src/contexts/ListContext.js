import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { useAuth } from './AuthContext';
import listService from '../services/listService';
import tribeService from '../services/tribeService';

const ListContext = createContext();

export const useList = () => useContext(ListContext);

export const ListProvider = ({ children }) => {
  const [userLists, setUserLists] = useState([]);
  const [sharedLists, setSharedLists] = useState([]);
  const [currentList, setCurrentList] = useState(null);
  const [currentListItems, setCurrentListItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [operations, setOperations] = useState({
    creating: false,
    updating: false,
    deleting: false,
    sharing: false
  });
  const { currentUser } = useAuth();

  // Helper function to handle errors consistently
  const handleError = useCallback((errorMessage, operation, err) => {
    console.error(`Error ${operation}:`, err);
    setError(errorMessage);
    return err;
  }, []);

  // Helper function to set operation loading state
  const setOperationLoading = useCallback((operation, isLoading) => {
    setOperations(prev => ({
      ...prev,
      [operation]: isLoading
    }));
  }, []);

  // Fetch user's lists
  const fetchUserLists = useCallback(async () => {
    if (!currentUser || !currentUser.id) {
      console.log("ListContext - fetchUserLists: No current user or user ID, skipping fetch");
      return;
    }

    try {
      setLoading(true);
      setError(null);
      console.log("ListContext - fetchUserLists: Fetching lists for user ID:", currentUser.id);
      const lists = await listService.getUserLists(currentUser.id);
      console.log("ListContext - Raw user lists from API:", lists);
      
      // Process the lists to ensure proper format
      let processedLists = [];
      
      // Verify list structure and IDs
      if (Array.isArray(lists)) {
        console.log("ListContext - Lists is an array with length:", lists.length);
        processedLists = lists.map((list, index) => {
          // Check if item is an API response rather than a list
          if (list && list.success === true && list.data) {
            console.warn(`List at index ${index} is an API response object, extracting data:`, list);
            return list.data;
          }
          
          // Check for missing ID
          if (!list.id) {
            console.warn(`List at index ${index} is missing an ID:`, list);
          } else {
            console.log(`List at index ${index} has ID: ${list.id}, name: ${list.name}`);
          }
          
          return list;
        }).filter(list => list && list.id); // Only keep lists with IDs
        
        console.log("ListContext - Processed user lists:", processedLists);
      } else {
        console.warn("Lists is not an array:", lists);
      }
      
      setUserLists(processedLists);
      console.log("ListContext - userLists after setting:", processedLists);
    } catch (err) {
      console.error("ListContext - Error in fetchUserLists:", err);
      handleError('Failed to load your lists. Please try again later.', 'fetching user lists', err);
      setUserLists([]);
    } finally {
      setLoading(false);
    }
  }, [currentUser, handleError]);

  // Fetch lists shared with the user's tribes
  const fetchSharedLists = useCallback(async () => {
    if (!currentUser || !currentUser.id) return;

    try {
      setLoading(true);
      setError(null);
      
      // Add a debounce/throttle mechanism to prevent excessive calls
      console.log("ListContext - fetchSharedLists: Fetching shared lists");
      
      // First, get the user's tribes using tribeService
      const userTribes = await tribeService.getUserTribes();
      console.log("ListContext - User tribes:", userTribes);
      
      if (!Array.isArray(userTribes) || userTribes.length === 0) {
        console.log("ListContext - No tribes found or invalid tribes data");
        setSharedLists([]);
        return;
      }
      
      // Create a Set of tribe IDs to track which tribes we've processed
      const processedTribeIds = new Set();
      const allSharedLists = [];
      
      // Process each tribe but avoid duplicates
      for (const tribe of userTribes) {
        if (!tribe.id || processedTribeIds.has(tribe.id)) {
          continue; // Skip invalid tribes or ones we've already processed
        }
        
        processedTribeIds.add(tribe.id);
        
        try {
          console.log(`ListContext - Fetching shared lists for tribe ${tribe.id}`);
          const tribeLists = await listService.getSharedListsForTribe(tribe.id);
          console.log(`ListContext - Received ${tribeLists?.length || 0} shared lists for tribe ${tribe.id}`);
          
          if (Array.isArray(tribeLists) && tribeLists.length > 0) {
            // Track processed list IDs to avoid duplicates
            const processedListIds = new Set(allSharedLists.map(list => list.id));
            
            for (const list of tribeLists) {
              // Extract data if needed and validate
              let processedList = list;
              if (list && list.success === true && list.data) {
                processedList = list.data;
              }
              
              if (!processedList.id) {
                console.warn(`ListContext - Shared list is missing an ID:`, processedList);
                continue;
              }
              
              // Only add if we haven't seen this list ID before
              if (!processedListIds.has(processedList.id)) {
                processedListIds.add(processedList.id);
                allSharedLists.push(processedList);
              }
            }
          }
        } catch (err) {
          console.warn(`ListContext - Error fetching shared lists for tribe ${tribe.id}:`, err);
          // Continue with other tribes even if one fails
        }
      }
      
      console.log(`ListContext - Setting shared lists array with ${allSharedLists.length} lists`);
      setSharedLists(allSharedLists);
    } catch (err) {
      console.error("ListContext - Error in fetchSharedLists:", err);
      handleError('Failed to load shared lists. Please try again later.', 'fetching shared lists', err);
      setSharedLists([]);
    } finally {
      setLoading(false);
    }
  }, [currentUser, handleError]);

  // Fetch a specific list and its items
  const fetchListDetails = useCallback(async (listId) => {
    try {
      setLoading(true);
      setError(null);
      
      // Get the list details from the API
      const list = await listService.getListDetails(listId);
      setCurrentList(list);
      
      // Get the list items from the API
      try {
        const items = await listService.getListItems(listId);
        // Ensure items is always an array, even if the API returns a different structure
        if (Array.isArray(items)) {
          setCurrentListItems(items);
        } else if (items && typeof items === 'object' && items.data && Array.isArray(items.data)) {
          // Handle case where API might return {data: [...]} structure
          console.log('ListContext: Received API response object for list items, extracting data array', items);
          setCurrentListItems(items.data);
        } else {
          console.warn('ListContext: List items from API is not an array, defaulting to empty array:', items);
          setCurrentListItems([]);
        }
      } catch (itemErr) {
        console.error(`Error fetching items for list ${listId}:`, itemErr);
        setCurrentListItems([]);
      }
    } catch (err) {
      handleError('Failed to load list details. Please try again later.', 'fetching list details', err);
      // Reset state to avoid null references
      setCurrentList(null);
      setCurrentListItems([]);
    } finally {
      setLoading(false);
    }
  }, [handleError]);

  // Create a new list
  const createList = useCallback(async (listData) => {
    try {
      setOperationLoading('creating', true);
      setError(null);
      const newList = await listService.createList(listData);
      
      // Check if the response is in the API response format and extract data if needed
      let listToAdd = newList;
      if (newList && newList.success === true && newList.data) {
        console.log('ListContext: Received API response object from listService, extracting data:', newList);
        listToAdd = newList.data;
      }
      
      console.log('ListContext: Adding new list to state:', listToAdd);
      
      // Verify the list has an ID before adding to state
      if (!listToAdd || !listToAdd.id) {
        console.error('ListContext: Cannot add list without ID to state:', listToAdd);
        throw new Error('Invalid list data returned from API');
      }
      
      // Update the local state with the new list
      setUserLists(prevLists => [...prevLists, listToAdd]);
      return listToAdd;
    } catch (err) {
      throw handleError('Failed to create list. Please try again later.', 'creating list', err);
    } finally {
      setOperationLoading('creating', false);
    }
  }, [handleError, setOperationLoading]);

  // Update a list
  const updateList = useCallback(async (listId, listData) => {
    try {
      setOperationLoading('updating', true);
      setError(null);
      const updatedList = await listService.updateList(listId, listData);
      
      // Update the local state with the updated list
      setUserLists(prevLists => 
        prevLists.map(list => list.id === listId ? updatedList : list)
      );
      
      // If this is the current list, update it
      if (currentList && currentList.id === listId) {
        setCurrentList(updatedList);
      }
      
      return updatedList;
    } catch (err) {
      throw handleError('Failed to update list. Please try again later.', 'updating list', err);
    } finally {
      setOperationLoading('updating', false);
    }
  }, [currentList, handleError, setOperationLoading]);

  // Delete a list
  const deleteList = useCallback(async (listId) => {
    try {
      setOperationLoading('deleting', true);
      setError(null);
      await listService.deleteList(listId);
      
      // Update the local state by removing the deleted list
      setUserLists(prevLists => prevLists.filter(list => list.id !== listId));
      setSharedLists(prevLists => prevLists.filter(list => list.id !== listId));
      
      // If this is the current list, clear it
      if (currentList && currentList.id === listId) {
        setCurrentList(null);
        setCurrentListItems([]);
      }
      
      return true;
    } catch (err) {
      throw handleError('Failed to delete list. Please try again later.', 'deleting list', err);
    } finally {
      setOperationLoading('deleting', false);
    }
  }, [currentList, handleError, setOperationLoading]);

  // Add an item to a list
  const addListItem = useCallback(async (listId, itemData) => {
    try {
      setLoading(true);
      setError(null);
      const newItem = await listService.addListItem(listId, itemData);
      
      // Update the local state with the new item
      if (currentList && currentList.id === listId) {
        setCurrentListItems(prevItems => [...prevItems, newItem]);
      }
      
      // Update the item count in the list
      setUserLists(prevLists => 
        prevLists.map(list => 
          list.id === listId 
            ? { ...list, itemCount: (list.itemCount || 0) + 1 } 
            : list
        )
      );
      
      setSharedLists(prevLists => 
        prevLists.map(list => 
          list.id === listId 
            ? { ...list, itemCount: (list.itemCount || 0) + 1 } 
            : list
        )
      );
      
      return newItem;
    } catch (err) {
      throw handleError('Failed to add item to list. Please try again later.', 'adding list item', err);
    } finally {
      setLoading(false);
    }
  }, [currentList, handleError]);

  // Update a list item
  const updateListItem = useCallback(async (listId, itemId, itemData) => {
    try {
      setLoading(true);
      setError(null);
      const updatedItem = await listService.updateListItem(listId, itemId, itemData);
      
      // Update the local state with the updated item
      if (currentList && currentList.id === listId) {
        setCurrentListItems(prevItems => 
          prevItems.map(item => item.id === itemId ? updatedItem : item)
        );
      }
      
      return updatedItem;
    } catch (err) {
      throw handleError('Failed to update item. Please try again later.', 'updating list item', err);
    } finally {
      setLoading(false);
    }
  }, [currentList, handleError]);

  // Delete a list item
  const deleteListItem = useCallback(async (listId, itemId) => {
    try {
      setLoading(true);
      setError(null);
      await listService.deleteListItem(listId, itemId);
      
      // Update the local state by removing the deleted item
      if (currentList && currentList.id === listId) {
        setCurrentListItems(prevItems => prevItems.filter(item => item.id !== itemId));
      }
      
      // Update the item count in the list
      setUserLists(prevLists => 
        prevLists.map(list => 
          list.id === listId && list.itemCount > 0
            ? { ...list, itemCount: list.itemCount - 1 } 
            : list
        )
      );
      
      setSharedLists(prevLists => 
        prevLists.map(list => 
          list.id === listId && list.itemCount > 0
            ? { ...list, itemCount: list.itemCount - 1 } 
            : list
        )
      );
      
      return true;
    } catch (err) {
      throw handleError('Failed to delete item. Please try again later.', 'deleting list item', err);
    } finally {
      setLoading(false);
    }
  }, [currentList, handleError]);

  // Get list shares
  const getListShares = useCallback(async (listId) => {
    if (!listId) {
      console.warn('ListContext - getListShares called with invalid listId');
      return [];
    }
    
    try {
      console.log(`ListContext - Getting shares for list ${listId}`);
      const shares = await listService.getListShares(listId);
      console.log(`ListContext - Retrieved ${shares?.length || 0} shares for list ${listId}`);
      return shares || [];
    } catch (err) {
      console.error(`ListContext - Error getting list shares for ${listId}:`, err);
      
      // Check for specific error types that should be handled differently
      if (err.response && err.response.status === 403) {
        console.warn(`ListContext - Permission denied to view shares for list ${listId}`);
      }
      
      // Don't throw the error, just return empty array
      return [];
    }
  }, []);

  // Load user lists when the component mounts or user changes
  useEffect(() => {
    if (currentUser) {
      console.log('ListContext - Initial load of user lists and shared lists');
      
      // Load user's own lists first
      fetchUserLists();
      
      // Only fetch shared lists if user has their own lists loaded
      // This helps reduce concurrent API load
      const timer = setTimeout(() => {
        fetchSharedLists();
      }, 300);
      
      return () => clearTimeout(timer);
    }
  }, [currentUser, fetchUserLists, fetchSharedLists]);

  // Add a function to refresh user lists
  const refreshUserLists = useCallback(async () => {
    console.log('ListContext - refreshUserLists called');
    // Clear lists first to avoid stale data
    setUserLists([]);
    // Force fetch from server
    try {
      if (currentUser && currentUser.id) {
        console.log('ListContext - Forcing fresh fetch of user lists');
        const timestamp = new Date().getTime(); // Add cache busting
        const lists = await listService.getUserLists(currentUser.id, timestamp);
        console.log('ListContext - Refreshed user lists:', lists);
        
        // Process the lists
        if (Array.isArray(lists)) {
          setUserLists(lists.filter(list => list && list.id));
        } else {
          console.warn('ListContext - Refreshed lists is not an array:', lists);
        }
      } else {
        console.log('ListContext - No user ID available for refreshUserLists');
      }
    } catch (err) {
      console.error('ListContext - Error refreshing user lists:', err);
    }
  }, [currentUser, listService]);

  const value = {
    userLists,
    sharedLists,
    currentList,
    currentListItems,
    loading,
    error,
    operations,
    fetchUserLists,
    fetchSharedLists,
    fetchListDetails,
    createList,
    updateList,
    deleteList,
    addListItem,
    updateListItem,
    deleteListItem,
    getListShares,
    refreshUserLists
  };

  return (
    <ListContext.Provider value={value}>
      {children}
    </ListContext.Provider>
  );
};

export default ListContext; 
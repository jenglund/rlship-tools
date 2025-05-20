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
    if (!currentUser || !currentUser.id) return;

    try {
      setLoading(true);
      setError(null);
      const lists = await listService.getUserLists(currentUser.id);
      setUserLists(lists);
    } catch (err) {
      handleError('Failed to load your lists. Please try again later.', 'fetching user lists', err);
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
      
      // First, get the user's tribes using tribeService
      const userTribes = await tribeService.getUserTribes();
      
      // Fetch shared lists for each tribe the user belongs to
      const allSharedLists = [];
      
      // Process each of the user's tribes
      for (const tribe of userTribes) {
        try {
          const tribeLists = await listService.getSharedListsForTribe(tribe.id);
          if (tribeLists && tribeLists.length > 0) {
            allSharedLists.push(...tribeLists);
          }
        } catch (err) {
          console.warn(`Error fetching shared lists for tribe ${tribe.id}:`, err);
          // Continue with other tribes even if one fails
        }
      }
      
      setSharedLists(allSharedLists);
    } catch (err) {
      handleError('Failed to load shared lists. Please try again later.', 'fetching shared lists', err);
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
      const items = await listService.getListItems(listId);
      setCurrentListItems(items);
    } catch (err) {
      handleError('Failed to load list details. Please try again later.', 'fetching list details', err);
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
      
      // Update the local state with the new list
      setUserLists(prevLists => [...prevLists, newList]);
      return newList;
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
    try {
      setLoading(true);
      setError(null);
      const shares = await listService.getListShares(listId);
      return shares;
    } catch (err) {
      throw handleError('Failed to get list shares. Please try again later.', 'getting list shares', err);
    } finally {
      setLoading(false);
    }
  }, [handleError]);

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
    getListShares
  };

  return (
    <ListContext.Provider value={value}>
      {children}
    </ListContext.Provider>
  );
};

export default ListContext; 
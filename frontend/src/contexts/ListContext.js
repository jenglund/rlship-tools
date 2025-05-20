import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { useAuth } from './AuthContext';
import listService from '../services/listService';

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
      // Here we would normally call the API, but for now we'll use placeholder data
      // const lists = await listService.getUserLists(currentUser.id);
      const lists = [
        { id: '1', name: 'Restaurants to Try', description: 'Places we want to try', type: 'location', itemCount: 12 },
        { id: '2', name: 'Movies to Watch', description: 'Our watch list', type: 'media', itemCount: 8 }
      ];
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
      // Here we would collect all lists shared with all of the user's tribes
      // This is a placeholder implementation
      const sharedLists = [
        { id: '3', name: 'Game Night Ideas', description: 'Board games to play', type: 'activity', itemCount: 5 }
      ];
      setSharedLists(sharedLists);
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
      // Here we would normally call the API, but for now we'll use placeholder data based on the ID
      // const list = await listService.getListDetails(listId);
      
      // Use hardcoded mock data for specific list IDs
      let list;
      let items;
      
      // Determine which list data to return based on ID
      switch(listId) {
        case '1':
          list = {
            id: '1',
            name: 'Restaurants to Try',
            description: 'Places we want to try for date night',
            type: 'location',
            ownerType: 'user',
            ownerId: '1',
            visibility: 'private'
          };
          
          items = [
            { id: '1', name: 'Sushi Place', description: 'Great sushi restaurant downtown', metadata: { location: 'Downtown', cuisine: 'Japanese' } },
            { id: '2', name: 'Italian Bistro', description: 'Authentic Italian food', metadata: { location: 'Westside', cuisine: 'Italian' } },
            { id: '3', name: 'Taco Shop', description: 'Best tacos in town', metadata: { location: 'Eastside', cuisine: 'Mexican' } }
          ];
          break;
          
        case '2':
          list = {
            id: '2',
            name: 'Movies to Watch',
            description: 'Our watch list for movie nights',
            type: 'media',
            ownerType: 'user',
            ownerId: '1',
            visibility: 'private'
          };
          
          items = [
            { id: '4', name: 'The Matrix', description: 'Sci-fi classic', metadata: { genre: 'Sci-fi', year: '1999' } },
            { id: '5', name: 'Inception', description: 'Mind-bending thriller', metadata: { genre: 'Thriller', year: '2010' } }
          ];
          break;
          
        case '3':
          list = {
            id: '3',
            name: 'Game Night Ideas',
            description: 'Board games to play with friends',
            type: 'activity',
            ownerType: 'user',
            ownerId: '1',
            visibility: 'shared'
          };
          
          items = [
            { id: '6', name: 'Catan', description: 'Classic strategy game', metadata: { players: '3-4', duration: '1-2 hours' } },
            { id: '7', name: 'Codenames', description: 'Word association game', metadata: { players: '4+', duration: '15-30 minutes' } },
            { id: '8', name: 'Pandemic', description: 'Cooperative game', metadata: { players: '2-4', duration: '45 minutes' } }
          ];
          break;
          
        default:
          // For any other ID, create a generic list with the ID
          list = {
            id: listId,
            name: `List ${listId}`,
            description: `This is list ${listId}`,
            type: 'general',
            ownerType: 'user',
            ownerId: '1',
            visibility: 'private'
          };
          
          items = [
            { id: 'default-1', name: 'Sample Item 1', description: 'This is a sample item', metadata: {} },
            { id: 'default-2', name: 'Sample Item 2', description: 'This is another sample item', metadata: {} }
          ];
      }
      
      setCurrentList(list);
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
      // Here we would normally call the API, but for now we'll use a mock
      // const newList = await listService.createList(listData);
      const newList = {
        id: Date.now().toString(), // Generate a fake ID
        ...listData,
        itemCount: 0
      };
      
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
      // Here we would normally call the API, but for now we'll use a mock
      // const updatedList = await listService.updateList(listId, listData);
      const updatedList = {
        id: listId,
        ...listData
      };
      
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
      // Here we would normally call the API, but for now we'll use a mock
      // await listService.deleteList(listId);
      
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
      // Here we would normally call the API, but for now we'll use a mock
      // const newItem = await listService.addListItem(listId, itemData);
      const newItem = {
        id: Date.now().toString(), // Generate a fake ID
        listId,
        ...itemData
      };
      
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
      // Here we would normally call the API, but for now we'll use a mock
      // const updatedItem = await listService.updateListItem(listId, itemId, itemData);
      const updatedItem = {
        id: itemId,
        listId,
        ...itemData
      };
      
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
      // Here we would normally call the API, but for now we'll use a mock
      // await listService.deleteListItem(listId, itemId);
      
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

  // Get list shares (mocked implementation)
  const getListShares = useCallback(async (listId) => {
    try {
      setLoading(true);
      setError(null);
      // Here we would normally call the API, but for now we'll use placeholder data
      // const shares = await listService.getListShares(listId);
      
      // Provide different share data based on list ID
      switch (listId) {
        case '1':
          // Return mock data for Restaurants list
          return [
            { 
              id: '101', 
              listId, 
              tribeId: '201',
              name: 'Family',
              tribeName: 'Family', 
              sharedAt: new Date().toISOString() 
            },
            { 
              id: '102', 
              listId, 
              tribeId: '202', 
              name: 'Friends',
              tribeName: 'Friends', 
              sharedAt: new Date().toISOString() 
            }
          ];
          
        case '2':
          // Return mock data for Movies list
          return [
            { 
              id: '103', 
              listId, 
              tribeId: '203',
              name: 'Movie Club',
              tribeName: 'Movie Club', 
              sharedAt: new Date().toISOString() 
            }
          ];
          
        case '3':
          // Return mock data for Game Night list
          return [
            { 
              id: '104', 
              listId, 
              tribeId: '202',
              name: 'Friends',
              tribeName: 'Friends', 
              sharedAt: new Date().toISOString() 
            },
            { 
              id: '105', 
              listId, 
              tribeId: '204', 
              name: 'Game Group',
              tribeName: 'Game Group', 
              sharedAt: new Date().toISOString() 
            }
          ];
          
        default:
          // Return empty array for any other list ID
          return [];
      }
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
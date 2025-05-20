import React, { createContext, useContext, useState, useEffect } from 'react';
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
  const { currentUser } = useAuth();

  // Fetch user's lists
  const fetchUserLists = async () => {
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
      console.error('Error fetching user lists:', err);
      setError('Failed to load your lists. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  // Fetch lists shared with the user's tribes
  const fetchSharedLists = async () => {
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
      console.error('Error fetching shared lists:', err);
      setError('Failed to load shared lists. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  // Fetch a specific list and its items
  const fetchListDetails = async (listId) => {
    try {
      setLoading(true);
      setError(null);
      // Here we would normally call the API, but for now we'll use placeholder data
      // const list = await listService.getListDetails(listId);
      const list = {
        id: listId,
        name: 'Restaurants to Try',
        description: 'Places we want to try for date night',
        type: 'location',
        ownerType: 'user',
        ownerId: '1',
        visibility: 'private'
      };
      setCurrentList(list);

      // Fetch items for this list
      // const items = await listService.getListItems(listId);
      const items = [
        { id: '1', name: 'Sushi Place', description: 'Great sushi restaurant downtown', metadata: { location: 'Downtown', cuisine: 'Japanese' } },
        { id: '2', name: 'Italian Bistro', description: 'Authentic Italian food', metadata: { location: 'Westside', cuisine: 'Italian' } },
        { id: '3', name: 'Taco Shop', description: 'Best tacos in town', metadata: { location: 'Eastside', cuisine: 'Mexican' } }
      ];
      setCurrentListItems(items);
    } catch (err) {
      console.error('Error fetching list details:', err);
      setError('Failed to load list details. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  // Create a new list
  const createList = async (listData) => {
    try {
      setLoading(true);
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
      console.error('Error creating list:', err);
      setError('Failed to create list. Please try again later.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Update a list
  const updateList = async (listId, listData) => {
    try {
      setLoading(true);
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
      console.error('Error updating list:', err);
      setError('Failed to update list. Please try again later.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Delete a list
  const deleteList = async (listId) => {
    try {
      setLoading(true);
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
      console.error('Error deleting list:', err);
      setError('Failed to delete list. Please try again later.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Add an item to a list
  const addListItem = async (listId, itemData) => {
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
      console.error('Error adding list item:', err);
      setError('Failed to add item to list. Please try again later.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Update a list item
  const updateListItem = async (listId, itemId, itemData) => {
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
      console.error('Error updating list item:', err);
      setError('Failed to update item. Please try again later.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  // Delete a list item
  const deleteListItem = async (listId, itemId) => {
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
      console.error('Error deleting list item:', err);
      setError('Failed to delete item. Please try again later.');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  const value = {
    userLists,
    sharedLists,
    currentList,
    currentListItems,
    loading,
    error,
    fetchUserLists,
    fetchSharedLists,
    fetchListDetails,
    createList,
    updateList,
    deleteList,
    addListItem,
    updateListItem,
    deleteListItem
  };

  return (
    <ListContext.Provider value={value}>
      {children}
    </ListContext.Provider>
  );
};

export default ListContext; 
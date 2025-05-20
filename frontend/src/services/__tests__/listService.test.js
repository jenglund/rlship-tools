import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';
import listService from '../listService';
import authService from '../authService';
import config from '../../config';

// Create axios mock
const mock = new MockAdapter(axios);
const API_URL = config.API_URL;

// Mock the authService.getCurrentUser method
jest.mock('../authService', () => ({
  getCurrentUser: jest.fn()
}));

describe('listService', () => {
  // Setup before each test
  beforeEach(() => {
    // Reset mock adapter
    mock.reset();
    // Reset mocked functions
    jest.clearAllMocks();
    // Mock user authentication
    authService.getCurrentUser.mockReturnValue({ token: 'fake-token' });
  });

  // Test createList method
  describe('createList', () => {
    const listData = {
      name: 'Test List',
      description: 'Test Description',
      type: 'activity'
    };
    const mockResponse = {
      data: {
        id: '123',
        name: 'Test List',
        description: 'Test Description',
        type: 'activity'
      }
    };

    test('should successfully create a list', async () => {
      // Arrange
      mock.onPost(`${API_URL}/v1/lists`).reply(200, {
        status: 'success',
        data: mockResponse.data
      });

      // Act
      const result = await listService.createList(listData);

      // Assert
      expect(result).toEqual(mockResponse.data);
      expect(mock.history.post[0].url).toBe(`${API_URL}/v1/lists`);
      expect(JSON.parse(mock.history.post[0].data)).toEqual(listData);
      expect(mock.history.post[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle validation errors', async () => {
      // Arrange
      const errorResponse = {
        status: 'error',
        message: 'Validation failed',
        errors: ['Name is required']
      };
      mock.onPost(`${API_URL}/v1/lists`).reply(400, errorResponse);

      // Act & Assert
      await expect(listService.createList({})).rejects.toThrow();
      expect(mock.history.post[0].url).toBe(`${API_URL}/v1/lists`);
    });

    test('should handle network failure', async () => {
      // Arrange
      mock.onPost(`${API_URL}/v1/lists`).networkError();

      // Act & Assert
      await expect(listService.createList(listData)).rejects.toThrow();
    });
  });

  // Test getUserLists method
  describe('getUserLists', () => {
    const userID = '123';
    const mockLists = [
      { id: '1', name: 'List 1', description: 'Description 1', type: 'activity' },
      { id: '2', name: 'List 2', description: 'Description 2', type: 'place' }
    ];

    test('should successfully retrieve user lists', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/user/${userID}`).reply(200, {
        status: 'success',
        data: mockLists
      });

      // Act
      const result = await listService.getUserLists(userID);

      // Assert
      expect(result).toEqual(mockLists);
      expect(mock.history.get[0].url).toBe(`${API_URL}/v1/lists/user/${userID}`);
      expect(mock.history.get[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle empty list response', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/user/${userID}`).reply(200, {
        status: 'success',
        data: []
      });

      // Act
      const result = await listService.getUserLists(userID);

      // Assert
      expect(result).toEqual([]);
    });

    test('should handle unauthorized access', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/user/${userID}`).reply(403, {
        status: 'error',
        message: 'Unauthorized'
      });

      // Act & Assert
      await expect(listService.getUserLists(userID)).rejects.toThrow();
    });
  });

  // Test getListDetails method
  describe('getListDetails', () => {
    const listID = '123';
    const mockList = {
      id: listID,
      name: 'Test List',
      description: 'Test Description',
      type: 'activity',
      items: []
    };

    test('should successfully retrieve list details', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}`).reply(200, {
        status: 'success',
        data: mockList
      });

      // Act
      const result = await listService.getListDetails(listID);

      // Assert
      expect(result).toEqual(mockList);
      expect(mock.history.get[0].url).toBe(`${API_URL}/v1/lists/${listID}`);
      expect(mock.history.get[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle list not found', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}`).reply(404, {
        status: 'error',
        message: 'List not found'
      });

      // Act & Assert
      await expect(listService.getListDetails(listID)).rejects.toThrow();
    });
  });
  
  // Test updateList method
  describe('updateList', () => {
    const listID = '123';
    const updateData = {
      name: 'Updated List Name',
      description: 'Updated Description'
    };
    const mockResponse = {
      id: listID,
      name: 'Updated List Name',
      description: 'Updated Description',
      type: 'activity'
    };

    test('should successfully update a list', async () => {
      // Arrange
      mock.onPut(`${API_URL}/v1/lists/${listID}`).reply(200, {
        status: 'success',
        data: mockResponse
      });

      // Act
      const result = await listService.updateList(listID, updateData);

      // Assert
      expect(result).toEqual(mockResponse);
      expect(mock.history.put[0].url).toBe(`${API_URL}/v1/lists/${listID}`);
      expect(JSON.parse(mock.history.put[0].data)).toEqual(updateData);
      expect(mock.history.put[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle validation errors', async () => {
      // Arrange
      const errorResponse = {
        status: 'error',
        message: 'Validation failed',
        errors: ['Name cannot be empty']
      };
      mock.onPut(`${API_URL}/v1/lists/${listID}`).reply(400, errorResponse);

      // Act & Assert
      await expect(listService.updateList(listID, { name: '' })).rejects.toThrow();
    });

    test('should handle list not found on update', async () => {
      // Arrange
      mock.onPut(`${API_URL}/v1/lists/${listID}`).reply(404, {
        status: 'error',
        message: 'List not found'
      });

      // Act & Assert
      await expect(listService.updateList(listID, updateData)).rejects.toThrow();
    });
  });
  
  // Test deleteList method
  describe('deleteList', () => {
    const listID = '123';

    test('should successfully delete a list', async () => {
      // Arrange
      mock.onDelete(`${API_URL}/v1/lists/${listID}`).reply(200, {
        status: 'success',
        data: null
      });

      // Act
      const result = await listService.deleteList(listID);

      // Assert
      expect(result).toBe(true);
      expect(mock.history.delete[0].url).toBe(`${API_URL}/v1/lists/${listID}`);
      expect(mock.history.delete[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle list not found on delete', async () => {
      // Arrange
      mock.onDelete(`${API_URL}/v1/lists/${listID}`).reply(404, {
        status: 'error',
        message: 'List not found'
      });

      // Act & Assert
      await expect(listService.deleteList(listID)).rejects.toThrow();
    });
  });

  // Test getListItems method
  describe('getListItems', () => {
    const listID = '123';
    const mockItems = [
      { id: '1', name: 'Item 1', description: 'Description 1' },
      { id: '2', name: 'Item 2', description: 'Description 2' }
    ];

    test('should successfully retrieve list items', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}/items`).reply(200, {
        status: 'success',
        data: mockItems
      });

      // Act
      const result = await listService.getListItems(listID);

      // Assert
      expect(result).toEqual(mockItems);
      expect(mock.history.get[0].url).toBe(`${API_URL}/v1/lists/${listID}/items`);
      expect(mock.history.get[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle empty items response', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}/items`).reply(200, {
        status: 'success',
        data: []
      });

      // Act
      const result = await listService.getListItems(listID);

      // Assert
      expect(result).toEqual([]);
    });

    test('should handle list not found', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}/items`).reply(404, {
        status: 'error',
        message: 'List not found'
      });

      // Act & Assert
      await expect(listService.getListItems(listID)).rejects.toThrow();
    });
  });

  // Test addListItem method
  describe('addListItem', () => {
    const listID = '123';
    const itemData = {
      name: 'New Item',
      description: 'New Item Description'
    };
    const mockResponse = {
      id: '456',
      name: 'New Item',
      description: 'New Item Description',
      listID: listID
    };

    test('should successfully add an item to a list', async () => {
      // Arrange
      mock.onPost(`${API_URL}/v1/lists/${listID}/items`).reply(200, {
        status: 'success',
        data: mockResponse
      });

      // Act
      const result = await listService.addListItem(listID, itemData);

      // Assert
      expect(result).toEqual(mockResponse);
      expect(mock.history.post[0].url).toBe(`${API_URL}/v1/lists/${listID}/items`);
      expect(JSON.parse(mock.history.post[0].data)).toEqual(itemData);
      expect(mock.history.post[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle validation errors', async () => {
      // Arrange
      const errorResponse = {
        status: 'error',
        message: 'Validation failed',
        errors: ['Name is required']
      };
      mock.onPost(`${API_URL}/v1/lists/${listID}/items`).reply(400, errorResponse);

      // Act & Assert
      await expect(listService.addListItem(listID, {})).rejects.toThrow();
    });

    test('should handle list not found when adding item', async () => {
      // Arrange
      mock.onPost(`${API_URL}/v1/lists/${listID}/items`).reply(404, {
        status: 'error',
        message: 'List not found'
      });

      // Act & Assert
      await expect(listService.addListItem(listID, itemData)).rejects.toThrow();
    });
  });

  // Test updateListItem method
  describe('updateListItem', () => {
    const listID = '123';
    const itemID = '456';
    const updateData = {
      name: 'Updated Item Name',
      description: 'Updated Item Description'
    };
    const mockResponse = {
      id: itemID,
      name: 'Updated Item Name',
      description: 'Updated Item Description',
      listID: listID
    };

    test('should successfully update an item', async () => {
      // Arrange
      mock.onPut(`${API_URL}/v1/lists/${listID}/items/${itemID}`).reply(200, {
        status: 'success',
        data: mockResponse
      });

      // Act
      const result = await listService.updateListItem(listID, itemID, updateData);

      // Assert
      expect(result).toEqual(mockResponse);
      expect(mock.history.put[0].url).toBe(`${API_URL}/v1/lists/${listID}/items/${itemID}`);
      expect(JSON.parse(mock.history.put[0].data)).toEqual(updateData);
      expect(mock.history.put[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle validation errors', async () => {
      // Arrange
      const errorResponse = {
        status: 'error',
        message: 'Validation failed',
        errors: ['Name cannot be empty']
      };
      mock.onPut(`${API_URL}/v1/lists/${listID}/items/${itemID}`).reply(400, errorResponse);

      // Act & Assert
      await expect(listService.updateListItem(listID, itemID, { name: '' })).rejects.toThrow();
    });

    test('should handle item not found on update', async () => {
      // Arrange
      mock.onPut(`${API_URL}/v1/lists/${listID}/items/${itemID}`).reply(404, {
        status: 'error',
        message: 'Item not found'
      });

      // Act & Assert
      await expect(listService.updateListItem(listID, itemID, updateData)).rejects.toThrow();
    });
  });

  // Test deleteListItem method
  describe('deleteListItem', () => {
    const listID = '123';
    const itemID = '456';

    test('should successfully delete an item', async () => {
      // Arrange
      mock.onDelete(`${API_URL}/v1/lists/${listID}/items/${itemID}`).reply(200, {
        status: 'success',
        data: null
      });

      // Act
      const result = await listService.deleteListItem(listID, itemID);

      // Assert
      expect(result).toBe(true);
      expect(mock.history.delete[0].url).toBe(`${API_URL}/v1/lists/${listID}/items/${itemID}`);
      expect(mock.history.delete[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle item not found on delete', async () => {
      // Arrange
      mock.onDelete(`${API_URL}/v1/lists/${listID}/items/${itemID}`).reply(404, {
        status: 'error',
        message: 'Item not found'
      });

      // Act & Assert
      await expect(listService.deleteListItem(listID, itemID)).rejects.toThrow();
    });
  });

  // Test getListShares method
  describe('getListShares', () => {
    const listID = '123';
    const mockShares = [
      { id: '1', tribeID: '101', tribeName: 'Family', sharedAt: '2023-01-01T00:00:00Z' },
      { id: '2', tribeID: '102', tribeName: 'Friends', sharedAt: '2023-01-02T00:00:00Z' }
    ];

    test('should successfully retrieve list shares', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}/shares`).reply(200, {
        status: 'success',
        data: mockShares
      });

      // Act
      const result = await listService.getListShares(listID);

      // Assert
      expect(result).toEqual(mockShares);
      expect(mock.history.get[0].url).toBe(`${API_URL}/v1/lists/${listID}/shares`);
      expect(mock.history.get[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle empty shares response', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}/shares`).reply(200, {
        status: 'success',
        data: []
      });

      // Act
      const result = await listService.getListShares(listID);

      // Assert
      expect(result).toEqual([]);
    });

    test('should handle list not found', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/${listID}/shares`).reply(404, {
        status: 'error',
        message: 'List not found'
      });

      // Act & Assert
      await expect(listService.getListShares(listID)).rejects.toThrow();
    });
  });

  // Test shareListWithTribe method
  describe('shareListWithTribe', () => {
    const listID = '123';
    const tribeID = '101';
    const mockResponse = {
      id: '201',
      listID: listID,
      tribeID: tribeID,
      sharedAt: '2023-01-01T00:00:00Z'
    };

    test('should successfully share a list with a tribe', async () => {
      // Arrange
      mock.onPost(`${API_URL}/v1/lists/${listID}/share/${tribeID}`).reply(200, {
        status: 'success',
        data: mockResponse
      });

      // Act
      const result = await listService.shareListWithTribe(listID, tribeID);

      // Assert
      expect(result).toEqual(mockResponse);
      expect(mock.history.post[0].url).toBe(`${API_URL}/v1/lists/${listID}/share/${tribeID}`);
      expect(mock.history.post[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle duplicate sharing error', async () => {
      // Arrange
      const errorResponse = {
        status: 'error',
        message: 'List is already shared with this tribe'
      };
      mock.onPost(`${API_URL}/v1/lists/${listID}/share/${tribeID}`).reply(400, errorResponse);

      // Act & Assert
      await expect(listService.shareListWithTribe(listID, tribeID)).rejects.toThrow();
    });

    test('should handle list not found when sharing', async () => {
      // Arrange
      mock.onPost(`${API_URL}/v1/lists/${listID}/share/${tribeID}`).reply(404, {
        status: 'error',
        message: 'List not found'
      });

      // Act & Assert
      await expect(listService.shareListWithTribe(listID, tribeID)).rejects.toThrow();
    });

    test('should handle tribe not found when sharing', async () => {
      // Arrange
      mock.onPost(`${API_URL}/v1/lists/${listID}/share/${tribeID}`).reply(404, {
        status: 'error',
        message: 'Tribe not found'
      });

      // Act & Assert
      await expect(listService.shareListWithTribe(listID, tribeID)).rejects.toThrow();
    });
  });

  // Test unshareListWithTribe method
  describe('unshareListWithTribe', () => {
    const listID = '123';
    const tribeID = '101';

    test('should successfully unshare a list with a tribe', async () => {
      // Arrange
      mock.onDelete(`${API_URL}/v1/lists/${listID}/share/${tribeID}`).reply(200, {
        status: 'success',
        data: null
      });

      // Act
      const result = await listService.unshareListWithTribe(listID, tribeID);

      // Assert
      expect(result).toBe(true);
      expect(mock.history.delete[0].url).toBe(`${API_URL}/v1/lists/${listID}/share/${tribeID}`);
      expect(mock.history.delete[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle share not found when unsharing', async () => {
      // Arrange
      mock.onDelete(`${API_URL}/v1/lists/${listID}/share/${tribeID}`).reply(404, {
        status: 'error',
        message: 'Share not found'
      });

      // Act & Assert
      await expect(listService.unshareListWithTribe(listID, tribeID)).rejects.toThrow();
    });
  });

  // Test getSharedListsForTribe method
  describe('getSharedListsForTribe', () => {
    const tribeID = '101';
    const mockLists = [
      { id: '1', name: 'Shared List 1', description: 'Description 1', type: 'activity' },
      { id: '2', name: 'Shared List 2', description: 'Description 2', type: 'place' }
    ];

    test('should successfully retrieve shared lists for tribe', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/shared/${tribeID}`).reply(200, {
        status: 'success',
        data: mockLists
      });

      // Act
      const result = await listService.getSharedListsForTribe(tribeID);

      // Assert
      expect(result).toEqual(mockLists);
      expect(mock.history.get[0].url).toBe(`${API_URL}/v1/lists/shared/${tribeID}`);
      expect(mock.history.get[0].headers).toHaveProperty('Authorization', 'Bearer fake-token');
    });

    test('should handle empty shared lists response', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/shared/${tribeID}`).reply(200, {
        status: 'success',
        data: []
      });

      // Act
      const result = await listService.getSharedListsForTribe(tribeID);

      // Assert
      expect(result).toEqual([]);
    });

    test('should handle tribe not found', async () => {
      // Arrange
      mock.onGet(`${API_URL}/v1/lists/shared/${tribeID}`).reply(404, {
        status: 'error',
        message: 'Tribe not found'
      });

      // Act & Assert
      await expect(listService.getSharedListsForTribe(tribeID)).rejects.toThrow();
    });
  });
}); 
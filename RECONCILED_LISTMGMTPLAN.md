# Reconciled Action Plan for Frontend List Management

## 1. Introduction and Current State Analysis

This document synthesizes previous analyses to create a comprehensive, consistent action plan for implementing list management features in the Tribe application frontend. This functionality will allow users to create, manage, and share lists (restaurants, activities, vacation spots, etc.) with their tribes.

### Current State
- **Backend**: A robust set of list-related APIs exists, covering CRUD operations for lists and list items, sharing mechanisms, and retrieval operations.
- **Frontend**: Currently lacks dedicated list management components or services.

## 2. API Structure Clarifications

A critical finding from Gemini's analysis is the proper API structure:

- **API Prefix**: All endpoints must be prefixed with `/api/v1`
- **Parameter Naming**: There are inconsistencies in the backend API parameter naming that must be addressed
  - Some endpoints use `{id}` while others use `{listID}`
  - Some use `{tribeId}` while others use `{tribeID}`
  
Frontend services should account for these variations while maintaining internal consistency. The reconciled API paths are:

### List Management
- `GET /api/v1/lists/user/{userID}` - Get user-owned lists
- `GET /api/v1/lists/shared/{tribeID}` - Get lists shared with a tribe
- `POST /api/v1/lists` - Create a new list
- `GET /api/v1/lists/{id}` - Get a specific list
- `PUT /api/v1/lists/{id}` - Update a list
- `DELETE /api/v1/lists/{id}` - Delete a list

### List Items Management
- `GET /api/v1/lists/{id}/items` - Get all items in a list
- `POST /api/v1/lists/{id}/items` - Add an item to a list
- `PUT /api/v1/lists/{id}/items/{itemID}` - Update a list item
- `DELETE /api/v1/lists/{id}/items/{itemID}` - Remove an item from a list

### List Sharing
- `GET /api/v1/lists/{id}/shares` - Get all shares for a list
- `POST /api/v1/lists/{id}/share/{tribeId}` - Share a list with a tribe
- `DELETE /api/v1/lists/{id}/share/{tribeId}` - Unshare a list from a tribe

### Additional Endpoints
- `GET /api/v1/lists` - General list endpoint (potentially for admin view)
- `GET /api/v1/lists/tribe/{tribeId}` - Get lists associated with a tribe

## 3. Component Architecture

The reconciled component architecture maintains a consistent naming convention and aligns with existing frontend patterns:

### Screens (Page-Level Components)
1. **`ListsScreen.js`** - Main screen showing both user-owned and shared lists
2. **`ListDetailScreen.js`** - View and manage a single list and its items
3. **`ListItemDetailScreen.js`** - View/edit details of a single list item (for detailed item view/editing)

### Modals
4. **`CreateListModal.js`** - Modal for creating a new list
5. **`EditListModal.js`** - Modal for editing list properties
6. **`CreateListItemModal.js`** - Modal for adding a new item to a list
7. **`EditListItemModal.js`** - Modal for editing an item
8. **`ShareListModal.js`** - Modal for sharing a list with tribes

### Reusable Components
9. **`ListCard.js`** - Card representation of a list for use in lists/grids
10. **`ListItemCard.js`** - Card representation of a list item
11. **`ListSharesView.js`** - Component showing which tribes a list is shared with

### Existing Components to Update
1. **`Navigation.js`** - Add a "Lists" link
2. **`App.js`** (or router) - Add routes for new list screens
3. **`TribeDetailScreen.js`** - (Phase 2) Add a "Shared Lists" tab
4. **`ProfileScreen.js`** - (Phase 2) Add a "My Lists" section

## 4. Services and State Management

### List Service
Create a new `listService.js` with functions corresponding to all API endpoints:

```javascript
// listService.js - All endpoints prefixed with /api/v1
const listService = {
  // List Management
  createList: (listData) => axios.post('/api/v1/lists', listData),
  getUserLists: (userID) => axios.get(`/api/v1/lists/user/${userID}`),
  getSharedListsForTribe: (tribeID) => axios.get(`/api/v1/lists/shared/${tribeID}`),
  getListDetails: (id) => axios.get(`/api/v1/lists/${id}`),
  updateList: (id, listData) => axios.put(`/api/v1/lists/${id}`, listData),
  deleteList: (id) => axios.delete(`/api/v1/lists/${id}`),
  
  // List Items Management
  getListItems: (listID) => axios.get(`/api/v1/lists/${listID}/items`),
  addListItem: (listID, itemData) => axios.post(`/api/v1/lists/${listID}/items`, itemData),
  updateListItem: (listID, itemID, itemData) => 
    axios.put(`/api/v1/lists/${listID}/items/${itemID}`, itemData),
  deleteListItem: (listID, itemID) => 
    axios.delete(`/api/v1/lists/${listID}/items/${itemID}`),
  
  // List Sharing
  getListShares: (listID) => axios.get(`/api/v1/lists/${listID}/shares`),
  shareListWithTribe: (listID, tribeId) => 
    axios.post(`/api/v1/lists/${listID}/share/${tribeId}`),
  unshareListWithTribe: (listID, tribeId) => 
    axios.delete(`/api/v1/lists/${listID}/share/${tribeId}`),
  
  // Additional Endpoints
  getAllLists: (offset, limit) => 
    axios.get(`/api/v1/lists?offset=${offset}&limit=${limit}`),
  getTribeLists: (tribeId) => axios.get(`/api/v1/lists/tribe/${tribeId}`)
};
```

### State Management
Create a React Context (`ListContext.js`) to manage list-related state. For more complex requirements, consider adding SWR or React Query in a later phase.

## 5. Phased Implementation Plan

### Phase 1: Core List Management (MVP)
1. **Service Implementation**
   - Create `listService.js` with the necessary API functions
   - Ensure proper auth header handling consistent with other services
   - Verify API paths and parameter formats

2. **Navigation & Routing**
   - Add "Lists" link to Navigation
   - Create routes in App.js for list screens

3. **ListsScreen Implementation**
   - Create screen with two sections: "My Lists" and "Shared With Me"
   - Implement list fetching using `getUserLists` and `getSharedListsForTribe`
   - Create `ListCard` component for display
   - Add button to open `CreateListModal`

4. **List Creation & Editing**
   - Implement `CreateListModal` with form for list properties
   - Implement `EditListModal` for updating lists
   - Connect to `listService` functions

5. **List Detail View**
   - Implement `ListDetailScreen` showing list properties and items
   - Add item display using `ListItemCard` components
   - Add buttons for editing/deleting the list

### Phase 2: List Item Management
1. **List Item CRUD Operations**
   - Implement `CreateListItemModal` for adding items
   - Implement `EditListItemModal` for editing items
   - Add item deletion functionality
   - Implement `ListItemDetailScreen` for detailed item viewing/editing

2. **List Types Support**
   - Add list type selection in creation/editing
   - Create type-specific item forms and displays

### Phase 3: List Sharing
1. **Share Management**
   - Implement `ShareListModal` for tribe selection
   - Create `ListSharesView` to display current shares
   - Integrate with sharing endpoints
   - Add unsharing functionality

2. **Integration with Tribe Views**
   - Update `TribeDetailScreen` to show shared lists
   - Add share/unshare capabilities from tribe view

### Phase 4: Polish and Advanced Features
1. **UI/UX Refinements**
   - Responsive design improvements
   - Loading states and error handling
   - Search and filtering

2. **Advanced Features**
   - Google Maps integration for location lists
   - Menu generation based on lists
   - List templates and recommendations

## 6. UI/UX Considerations

### List Card Design
- Type-specific icon
- Item count indicator
- Shared status indicator
- Consistent with existing card components

### List Detail Layout
- Tabs for Items, Sharing, Settings
- Clear actions (add, edit, delete, share)
- Responsive layout that works well on mobile

### General UI Guidelines
- Follow existing color scheme and style
- Use consistent spacing and typography
- Ensure accessibility compliance
- Provide clear feedback for user actions

## 7. Code Structure and Organization

```
frontend/
  src/
    components/
      lists/
        ListsScreen.js
        ListDetailScreen.js
        ListItemDetailScreen.js
        ListCard.js
        ListItemCard.js
        modals/
          CreateListModal.js
          EditListModal.js
          CreateListItemModal.js
          EditListItemModal.js
          ShareListModal.js
        ListSharesView.js
    services/
      listService.js
    contexts/
      ListContext.js
    hooks/
      useList.js (optional, for custom list-related hooks)
```

## 8. Potential Backend Gaps and Considerations

1. **API Consistency**: Address parameter naming inconsistencies (`id` vs `listID` and `tribeId` vs `tribeID`)
2. **List Filtering/Sorting**: May need enhanced endpoints for filtering by name, date, type
3. **List Item Schema**: Ensure backend can handle different item types with varying fields
4. **Permissions Model**: Clarify ownership vs. editing rights for shared lists
5. **Bulk Operations**: Consider future support for batch actions

## 9. Next Steps

1. **API Verification**
   - Test all identified endpoints with Postman/curl to verify paths and parameters
   - Document any discrepancies found

2. **Core Component Development**
   - Begin with `listService.js` implementation
   - Develop `ListsScreen.js` and `ListCard.js` components
   - Implement basic list creation

3. **Integration Testing**
   - Test list creation and retrieval end-to-end
   - Verify data flow and state management

This reconciled plan addresses the differences in the previous analyses while ensuring a consistent, maintainable approach aligned with the existing codebase. 
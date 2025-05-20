# Frontend Mock Data Replacement Strategy

## Current State Analysis

After analyzing the codebase, I've identified the following key issues:

1. **Mock Data in ListContext.js**: The `ListContext.js` file contains hardcoded mock data for lists, list items, and sharing instead of fetching real data from the backend.

2. **Commented API Calls**: The service layer has real API calls implemented but they're commented out in favor of hardcoded mock data.

3. **Authentication Integration**: The authentication service is already configured to work with development users, but it's not fully utilized for data operations.

## Action Plan

### 1. Restore API Calls in ListContext.js

Replace all hardcoded mock data with real API calls that are currently commented out:

#### High Priority Changes:

- **fetchUserLists**: Uncomment the `listService.getUserLists(currentUser.id)` call
- **fetchListDetails**: Uncomment the `listService.getListDetails(listId)` call
- **createList**: Uncomment the `listService.createList(listData)` call
- **updateList**: Uncomment the `listService.updateList(listId, listData)` call
- **deleteList**: Uncomment the `listService.deleteList(listId)` call
- **addListItem**: Uncomment the `listService.addListItem(listId, itemData)` call
- **updateListItem**: Uncomment the `listService.updateListItem(listId, itemId, itemData)` call
- **deleteListItem**: Uncomment the `listService.deleteListItem(listId, itemId)` call
- **getListShares**: Uncomment the `listService.getListShares(listId)` call

### 2. Verify and Fix Backend API Endpoints

Ensure all necessary endpoints exist and work correctly in the backend:

| Frontend Service Method | Backend Endpoint | Status |
|------------------------|-----------------|--------|
| getUserLists | GET `/v1/lists/user/:userID` | Verify |
| getListDetails | GET `/v1/lists/:id` | Verify |
| createList | POST `/v1/lists` | Verify |
| updateList | PUT `/v1/lists/:id` | Verify |
| deleteList | DELETE `/v1/lists/:id` | Verify |
| getListItems | GET `/v1/lists/:listID/items` | Verify |
| addListItem | POST `/v1/lists/:listID/items` | Verify |
| updateListItem | PUT `/v1/lists/:listID/items/:itemID` | Verify |
| deleteListItem | DELETE `/v1/lists/:listID/items/:itemID` | Verify |
| getListShares | GET `/v1/lists/:listID/shares` | Verify |
| shareListWithTribe | POST `/v1/lists/:listID/share/:tribeID` | Verify |
| unshareListWithTribe | DELETE `/v1/lists/:listID/share/:tribeID` | Verify |
| getSharedListsForTribe | GET `/v1/lists/shared/:tribeID` | Verify |

### 3. Fix Backend Sharing Implementation

Ensure the backend correctly implements permissions and sharing functionality:

1. Verify that lists are private by default and only visible to their owners
2. Ensure sharing with tribes works correctly:
   - When a list is shared with a tribe, all tribe members should be able to see it
   - When a list is unshared from a tribe, tribe members can no longer see it
3. Implement proper authorization checks on all list endpoints

### 4. Enhance Error Handling

1. Improve error handling in the frontend to provide meaningful feedback when API calls fail
2. Add retry mechanisms for transient failures
3. Implement proper loading states during API calls

### 5. User Authentication Integration

1. Ensure all API requests include proper authentication headers
2. Implement token refresh mechanism if necessary
3. Verify that dev user authentication works as expected with the backend

### 6. Update UI Components

1. Update List components to handle real data structures from the API
2. Add proper loading indicators during API calls
3. Implement error states for failed requests
4. Ensure list ownership and permissions are correctly displayed

### 7. Testing Strategy

1. Verify all API endpoints with real dev user accounts
2. Test sharing functionality with multiple dev users
3. Verify that permissions are enforced correctly
4. Test edge cases (empty lists, large lists, etc.)

## Backend Implementation Checklist

If necessary, implement or fix the following backend functionality:

1. List Management:
   - [ ] CRUD operations for lists
   - [ ] CRUD operations for list items
   
2. Sharing Implementation:
   - [ ] Share lists with tribes
   - [ ] Unshare lists from tribes
   - [ ] Fetch lists shared with a tribe
   - [ ] Proper permissions enforcement

3. User Authentication:
   - [ ] Development authentication support
   - [ ] User ID integration with lists and tribes
   - [ ] Permission validation

## Implementation Timeline

1. **Phase 1 (Backend Verification)**:
   - Verify all required backend endpoints exist and work correctly
   - Fix any missing or incorrect backend functionality
   - Implement proper permission checking

2. **Phase 2 (Mock Data Removal)**:
   - Replace hardcoded mock data with real API calls in ListContext.js
   - Update other components relying on mock data

3. **Phase 3 (Testing and Refinement)**:
   - Test with multiple dev users
   - Fix any issues found during testing
   - Refine UI to handle all possible states

## Conclusion

By following this plan, we will transition from using mock data to using real backend data with proper ownership and sharing permissions. This will provide a more realistic user experience and allow for proper testing of the application's functionality. 
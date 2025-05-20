# Gemini's Action Plan for Frontend List Management

## 1. Introduction

This document outlines a comprehensive action plan for implementing list management, sharing, and curation features in the Tribe application's frontend. It builds upon previous analysis (e.g., `FE_LIST_MGMT.md`) and incorporates findings from a direct review of the existing backend API capabilities.

The core goal is to provide users with the ability to create, manage, and share lists of various types (e.g., restaurants, activities, vacation spots) with their tribes, aligning with the application's focus on collaborative planning.

## 2. Current State Analysis

*   **Backend**: A significant set of list-related APIs already exists, covering CRUD operations for lists and list items, sharing mechanisms, user-specific list retrieval, and tribe-specific list retrieval. Some advanced features like sync, menu generation, and conflict resolution are also present in the backend.
*   **Frontend**: Currently, there are no dedicated list management components, services, or contexts. This will be a new feature area for the frontend.
*   **Claude's Plan (`FE_LIST_MGMT.md`)**: Provided a solid foundation, identifying key components and API endpoints. This plan will refine and expand upon it, particularly regarding API endpoint specifics and potential phasing.

## 3. Key Discrepancies with `FE_LIST_MGMT.md` and API Clarifications

While Claude's plan was largely accurate, direct backend analysis revealed:

*   **API Endpoint Variations**:
    *   The backend uses `/api/v1` as a prefix for all routes, which was not explicitly mentioned in Claude's API list. All frontend calls must include this.
    *   Path parameter naming for list and tribe identifiers shows some variation (e.g., `{listID}` vs. `{id}`, `{tribeID}` vs. `{tribeId}`).
        *   For creating new frontend features, we will **standardize on using `listID` and `tribeID`** for clarity and consistency where possible, but will adapt to the backend's specific requirements for each endpoint.
        *   `GET /lists/{id}/shares` should be called as `GET /api/v1/lists/{listID}/shares` by the frontend (assuming `id` is a placeholder for `listID`).
        *   `POST /lists/{id}/share/{tribeId}` should be called as `POST /api/v1/lists/{listID}/share/{tribeID}`.
        *   `DELETE /lists/{id}/share/{tribeId}` should be called as `DELETE /api/v1/lists/{listID}/share/{tribeID}`.
*   **Additional Endpoints Not in Claude's Plan (for consideration)**:
    *   `GET /api/v1/lists/`: General `ListLists` endpoint. The frontend will need to decide how to use this, perhaps for an admin view or if user-specific filtering is handled by query parameters not immediately visible in the route definition. For user-specific lists, `GET /api/v1/lists/user/{userID}` is clearer.
    *   `GET /api/v1/lists/tribe/{tribeID}`: `GetTribeLists` - Potentially useful for displaying lists directly associated with or "owned" by a tribe, distinct from lists *shared with* a tribe.
*   **Confirmation of API Prefix**: All backend routes are typically prefixed (e.g. `/api/v1`). The handler registration `r.Route("/lists", ...)` implies these list routes will be under `/api/v1/lists/...`. This needs to be consistently used in the frontend service.

## 4. Proposed Frontend Components

This largely aligns with Claude's plan, with minor clarifications.

### New Components to Create:
1.  **`ListsListScreen.js`**: Main screen.
    *   Sections: "My Lists" (user-owned) and "Shared With My Tribes".
    *   Functionality: Create new list, navigate to list details, basic search/filter.
2.  **`ListDetailScreen.js`**: View/manage a single list.
    *   Displays list properties (name, description).
    *   Manages list items (view, add, edit, remove).
    *   Access to sharing management.
    *   Access to edit list properties / delete list.
3.  **`ListItemDisplay.js`** (or `ListItemCard.js`): Component to display a single list item within `ListDetailScreen`.
    *   Could include quick actions like edit/delete.
4.  **`CreateListModal.js`**: Modal for new list creation.
    *   Inputs: Name, description, list type (e.g., "restaurants", "activities", "places").
5.  **`EditListModal.js`**: Modal for editing existing list properties.
6.  **`CreateListItemModal.js`**: Modal for adding a new item to a list.
    *   Inputs will vary based on list type (e.g., for a restaurant: name, address, notes; for an activity: description, estimated cost).
7.  **`EditListItemModal.js`**: Modal for editing an existing list item.
8.  **`ShareListModal.js`**: Modal for sharing a list.
    *   Allow selecting tribes the user is a member of.
    *   Show currently shared tribes.
9.  **`ListCard.js`**: Reusable component to display a list summary in `ListsListScreen`.
    *   Shows: Name, item count, shared status, type icon.

### Existing Components to Update:
1.  **`Navigation.js`**: Add a "Lists" link.
2.  **`App.js` (or router configuration)**: Add routes for the new list screens.
3.  **`TribeDetailScreen.js`**: (Phase 2/3) Consider adding a "Shared Lists" tab or section showing lists shared with that specific tribe.
4.  **`ProfileScreen.js`**: (Phase 2/3) Consider adding a "My Lists" section.

## 5. Frontend Services and State Management

### New Services:
1.  **`listService.js`**:
    *   Encapsulates all API calls to the backend list endpoints.
    *   Handles request/response structures, authentication headers.
    *   **Crucially, all endpoint paths must be prefixed with `/api/v1`**.

    **Key functions in `listService.js` (mapping to backend and Claude's plan):**

    *   `createList(listData)`: `POST /api/v1/lists/`
    *   `getUserLists(userID)`: `GET /api/v1/lists/user/{userID}`
    *   `getSharedListsForTribe(tribeID)`: `GET /api/v1/lists/shared/{tribeID}` (for "Shared with me" view)
    *   `getListDetails(listID)`: `GET /api/v1/lists/{listID}`
    *   `updateList(listID, listData)`: `PUT /api/v1/lists/{listID}`
    *   `deleteList(listID)`: `DELETE /api/v1/lists/{listID}`
    *   `getListItems(listID)`: `GET /api/v1/lists/{listID}/items`
    *   `addListItem(listID, itemData)`: `POST /api/v1/lists/{listID}/items`
    *   `updateListItem(listID, itemID, itemData)`: `PUT /api/v1/lists/{listID}/items/{itemID}`
    *   `deleteListItem(listID, itemID)`: `DELETE /api/v1/lists/{listID}/items/{itemID}`
    *   `getListShares(listID)`: `GET /api/v1/lists/{listID}/shares` (using corrected path param)
    *   `shareListWithTribe(listID, tribeID)`: `POST /api/v1/lists/{listID}/share/{tribeID}` (using corrected path params)
    *   `unshareListWithTribe(listID, tribeID)`: `DELETE /api/v1/lists/{listID}/share/{tribeID}` (using corrected path params)
    *   _(Optional, for future or admin)_ `getAllLists(offset, limit)`: `GET /api/v1/lists/`

### State Management:
*   **React Context (`ListContext.js`)**: As suggested by Claude, this is a good starting point for managing list-related state, especially for data fetched and shared across multiple components (e.g., the user's lists, currently viewed list details).
*   **Local Component State**: For form inputs, modal visibility, etc.
*   Consider libraries like SWR or React Query for more advanced data fetching, caching, and synchronization if `ListContext` becomes complex.

## 6. Detailed Phased Implementation Plan

### Phase 1: Core List & Item Viewing and Basic CRUD (MVP)

1.  **Setup `listService.js`**:
    *   Implement basic structure and auth header handling.
    *   Add functions for: `createList`, `getUserLists`, `getListDetails`, `updateList`, `deleteList`.
2.  **Navigation & Routing**:
    *   Add "Lists" link in `Navigation.js`.
    *   Define routes for `ListsListScreen` and `ListDetailScreen` in `App.js`.
3.  **`ListsListScreen.js` (View User's Lists)**:
    *   Fetch and display lists owned by the current user using `listService.getUserLists`.
    *   Use `ListCard.js` for display.
    *   Button to trigger `CreateListModal`.
4.  **`CreateListModal.js` & `EditListModal.js`**:
    *   Forms for list name, description. (Initially, list type can be a simple text field or deferred).
    *   Integrate with `listService.createList` and `listService.updateList`.
    *   Refresh list on `ListsListScreen` after creation/update.
5.  **`ListDetailScreen.js` (View & Basic List Item Management)**:
    *   Fetch and display list details using `listService.getListDetails`.
    *   Fetch and display items using `listService.getListItems` (using `ListItemDisplay.js`).
    *   Buttons to trigger `CreateListItemModal`, `EditListItemModal`.
    *   Implement delete list functionality with `listService.deleteList`.
6.  **`CreateListItemModal.js` & `EditListItemModal.js`**:
    *   Basic forms for item details (e.g., a 'name' or 'content' field).
    *   Integrate with `listService.addListItem`, `listService.updateListItem`.
    *   Implement item deletion with `listService.deleteListItem`.
    *   Refresh items on `ListDetailScreen`.
7.  **Basic `ListContext.js`**:
    *   To hold user's lists and potentially the currently active list being viewed/edited, to reduce prop drilling.

### Phase 2: List Sharing & Enhanced Viewing

1.  **Extend `listService.js`**:
    *   Add functions: `getSharedListsForTribe`, `getListShares`, `shareListWithTribe`, `unshareListWithTribe`.
2.  **`ListsListScreen.js` (Display Shared Lists)**:
    *   Fetch and display lists shared with the user's tribes (requires fetching user's tribes first, then iterating or a dedicated backend endpoint if available). `getSharedListsForTribe` will be key.
3.  **`ShareListModal.js`**:
    *   Allow user to select one of their tribes.
    *   Call `listService.shareListWithTribe`.
    *   Display current shares (using `listService.getListShares`) and allow unsharing via `listService.unshareListWithTribe`.
4.  **Integrate Sharing into `ListDetailScreen.js`**:
    *   Button to open `ShareListModal`.
    *   Visually indicate if a list is shared and with whom/which tribes.
5.  **Refine List Types**:
    *   Introduce a `list_type` field in `CreateListModal` (e.g., dropdown: 'Generic', 'Locations', 'To-Do').
    *   `CreateListItemModal` and `EditListItemModal` forms adapt to `list_type` (e.g., 'Locations' type would have fields for address, link). This requires `models.ListItem` in backend to support flexible data or different item tables per type.

### Phase 3: Advanced Features & UI Polish

1.  **Support for List Types (UI refinement)**:
    *   Implement distinct UI rendering in `ListItemDisplay.js` based on `list_type`.
    *   Type-specific icons and actions.
2.  **Integration with `TribeDetailScreen.js`**:
    *   Add a "Shared Lists" tab showing lists shared *with that specific tribe*.
3.  **Profile Integration (`ProfileScreen.js`)**:
    *   Section showing "My Lists".
4.  **UI/UX Enhancements**:
    *   Responsive design improvements.
    *   Loading states, error handling, and user feedback.
    *   Search/filter capabilities on `ListsListScreen`.
5.  **Consider Backend Sync/Menu Features**:
    *   If `SyncList`, `GenerateMenu` backend features are mature, plan frontend integration. This was noted as "Advanced Features" in Claude's plan and seems appropriate to defer.

## 7. UI/UX Considerations (Aligns with Claude)

1.  **`ListCard.js` Design**: Name, item count, shared status, type icon.
2.  **`ListDetailScreen.js` Layout**: Clear separation of info, items, sharing, settings. Tabs are a good approach.
3.  **Responsive Design**: Mobile-first.
4.  **User Feedback**: Clear loading indicators, success/error messages for all operations.

## 8. Code Structure and Organization (Aligns with Claude)

```
frontend/
  src/
    components/
      lists/  # All new list-related components here
        ListsListScreen.js
        ListDetailScreen.js
        ListItemDisplay.js
        CreateListModal.js
        EditListModal.js
        ListCard.js
        CreateListItemModal.js
        EditListItemModal.js
        ShareListModal.js
        # Potentially ListSharesView.js if sharing display becomes complex
    services/
      listService.js
    contexts/
      ListContext.js
    # Potentially hooks/ for list-specific hooks
    # Potentially utils/ for list-related utility functions
```

## 9. Potential Backend Gaps/Clarifications (from Claude's plan, still relevant)

*   **Filtering/Sorting for `GetUserLists` and `GetSharedLists`**: As lists grow, the frontend will need ways to filter or sort them (e.g., by name, date created, type). The backend might need to support these query parameters.
*   **List Item Schema for Different Types**: The `models.ListItem` backend structure needs to be flexible enough (e.g., using a JSONB `data` field) to accommodate different fields for different list item types, or the backend needs type-specific item tables/models.
*   **Bulk Operations**: For managing many list items (e.g., bulk delete, bulk add from CSV) - a future enhancement.
*   **Permissions**: Further clarity on list ownership vs. editing rights vs. viewing rights when shared. The current backend owner routes (`/{listID}/owners`) might play a role here beyond simple sharing.

## 10. Next Steps (Immediate)

1.  Verify the standard API prefix (e.g., `/api/v1/`) with backend router configuration.
2.  Begin implementation of **Phase 1**:
    *   Create `listService.js` with initial CRUD functions (using the confirmed API prefix).
    *   Develop `ListsListScreen.js` to display user's own lists.
    *   Implement `CreateListModal.js`.

This plan provides a detailed roadmap. Flexibility will be needed as development progresses and new insights are gained. 
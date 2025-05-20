# Future Work and Implementation Plans

This document outlines the roadmap for future development of the Tribe project, including planned features, implementation details, and testing strategies.

## Authentication and Tribe Management Implementation Plan

### Overview
This implementation plan outlines the steps needed to connect the frontend and backend for basic user authentication, user registration, and tribe management features.

### User Flow
1. **Authentication Flow**
   - When a user attempts to log in, we collect their email
   - If the email exists in our system, log the user in
   - If the email doesn't exist, prompt for additional information (display name, etc.)
   - Create a new user account with the provided information
   - Log the new user in

2. **Tribe Management Flow**
   - After login, user sees a list of tribes they belong to (empty for new users)
   - User can create a new tribe by clicking a "Create Tribe" button
   - When creating a tribe, prompt for tribe name
   - After tribe creation, user is automatically added as a member
   - Display the tribe overview page showing all members 
   - Add ability to delete tribes with confirmation prompt
   - After deletion, return to the tribes list screen

### Implementation Tasks

#### Backend Tasks
1. **Authentication Endpoints**
   - Implement `/api/auth/check-email` endpoint to verify if an email exists
   - Create `/api/auth/register` endpoint for new user registration
   - Implement `/api/auth/login` endpoint for user login
   - Add proper validation for user input

2. **User Management**
   - Implement user service for creating and managing users
   - Add validation for required user fields (email, display name)
   - Create repository functions for user operations

3. **Tribe Management Endpoints**
   - Implement `/api/tribes` endpoint with GET (list) and POST (create) methods
   - Create `/api/tribes/:id` endpoint with GET (details) and DELETE methods
   - Ensure proper authentication and authorization checks

4. **Error Handling**
   - Implement consistent error responses
   - Add validation error handling
   - Create middleware for authentication verification

#### Frontend Tasks
1. **Authentication Components**
   - Update `LoginScreen` to handle email input
   - Create `RegistrationScreen` for new user information collection
   - Implement `AuthService` for API communication
   - Update `AuthContext` to manage user state

2. **Tribe Management Components**
   - Enhance `TribesListScreen` to display user's tribes
   - Add "Create Tribe" button and associated modal/form
   - Create `TribeDetailScreen` to show tribe members
   - Implement delete functionality with confirmation dialog
   - Add proper navigation between screens

3. **Services and API Integration**
   - Enhance `tribeService.js` to communicate with backend API
   - Create `userService.js` for user-related API calls
   - Implement proper error handling and loading states

4. **UI Improvements**
   - Add form validation for user inputs
   - Implement loading indicators during API calls
   - Create error notifications for failed operations
   - Add confirmation dialogs for destructive actions

### Testing Strategy
1. **Unit Tests**
   - Test user registration and login logic
   - Test tribe creation and deletion functionality
   - Verify proper state management

2. **Integration Tests**
   - Test API integration with backend
   - Verify proper error handling

3. **E2E Tests**
   - Test complete user flows from login to tribe management

### Implementation Order
1. Backend authentication endpoints
2. Frontend authentication components
3. Backend tribe management endpoints
4. Frontend tribe management components
5. Error handling and validation
6. Testing and refinement

## User Story Implementation Plans

### User Story #1: Basic Authentication and Tribe Viewing Flow

**User Story Description**:
As a user, I want to land on a welcome page for Tribe, have the ability to log in, view my tribes, and log out.

**Implementation Plan**:
   
**Frontend Changes**:
1. **Splash Screen (Landing Page)**:
   - Create a responsive welcome page with Tribe branding
   - Add a prominent login button
   - Ensure the page loads quickly with optimized assets

2. **Authentication**:
   - Create an AuthContext for managing authentication state
   - Implement AuthService for API communication
   - Build a LoginScreen with four dev user login buttons
   - Add logout functionality on the tribes list page
   - (Future) Add OAuth integration for Google login

3. **Tribes List Screen**:
   - Create a TribesListScreen component to display all tribes a user belongs to
   - Implement tribe card components with basic tribe information
   - Add navigation to view individual tribe details

4. **Navigation**:
   - Set up React Router for navigation between screens
   - Implement protected routes for authenticated users
   - Create a navigation layout with consistent UI elements

5. **State Management**:
   - Use React Context for authentication state
   - Implement proper loading and error states
   - Add data fetching hooks for API integration

**Backend Support Needed**:
1. **Authentication Endpoints**:
   - Implement `/api/auth/login` endpoint for dev user authentication
   - Create `/api/auth/logout` endpoint
   - Add authentication middleware for protected routes

2. **Tribe Data Endpoints**:
   - Ensure `/api/tribes` endpoint returns all tribes for the authenticated user
   - Create `/api/tribes/:id` endpoint for individual tribe details

**Implementation Steps**:
1. Set up required dependencies (React Router, etc.)
2. Create authentication context and service
3. Build splash screen and login components
4. Implement tribes list view
5. Add navigation between components
6. Connect to backend API endpoints
7. Add proper error handling and loading states
8. Implement logout functionality
9. Write tests for all components and flows

## Tribe Membership Status Implementation

### Overview
Recently implemented a "pending" status for tribe members who have been invited but not yet accepted their invitation. This involved:

1. **Model Updates**
   - Added `MembershipPending` constant to the membership types
   - Updated validation logic to handle pending members appropriately
   - Expanded test coverage for the new status

2. **Repository Layer**
   - Updated `AddMember` method to add members with pending status by default
   - Implemented functionality in `UpdateMember` to change status when invitation is accepted/rejected

3. **Handler Layer**
   - Created `RespondToInvitation` handler to process invitation responses
   - Updated existing member-related handlers to account for pending status

### Future Improvements
1. **Test Coverage Expansion**
   - Add more comprehensive tests for the membership status transitions
   - Improve repository mock test coverage (currently at 0% for many methods)
   - Ensure edge cases are properly tested (e.g., trying to downgrade a full member to pending)

2. **User Experience Enhancements**
   - Implement notifications for pending invitations
   - Add ability to see pending invitations on user dashboard
   - Create invitation email functionality

3. **Security Improvements**
   - Add rate limiting for invitation responses
   - Implement expiration for pending invitations
   - Add audit logging for membership status changes

4. **Frontend Integration**
   - Create UI components for displaying and managing invitations
   - Implement invitation acceptance/rejection flows
   - Show pending status in tribe member lists

### Implementation Timeline
1. Short-term (1-2 weeks):
   - Complete test coverage improvements
   - Fix any remaining bugs in the invitation flow

2. Medium-term (2-4 weeks):
   - Implement frontend components for invitation management
   - Add notification system for pending invitations

3. Long-term (1-2 months):
   - Add invitation expiration functionality
   - Implement email notifications for invitations
   - Create comprehensive audit logging

## Backend Development Priorities

1. **✅ Database Schema Isolation and Test Stability (COMPLETED)**
   - Fixed issues with null constraint violations:
     - Updated tribe creation in tests to include required metadata field
     - Added proper JSON objects for metadata in test fixtures
     - Ensured consistency between model structs and database schema
   - Fixed context handling in repository methods:
     - Updated methods to use the schema context properly
     - Fixed context cancellation issues in transaction management
   - All repository and API tests now pass successfully

2. **✅ Fixed Tribe Member Invitation Data (COMPLETED)**
   - Fixed issue with the `invited_by` field not being properly set when inviting new members to a tribe
   - Updated the `AddMember` handler to correctly extract the user ID from the context using the `getUserIDFromContext` helper
   - Ensured consistency in how the user ID is accessed and used across different handlers
   - This fixes the frontend display of "Invited by" information for all tribe members, not just tribe creators

3. **Improve Backend Test Coverage**
   - Add tests for edge cases in repository methods
   - Improve context-based testing approach
   - Standardize test fixtures and setup/teardown
   - Add integration tests for common workflows
   - Create performance benchmarks for database operations
   - Implement test coverage tracking and reporting
   - Add race condition detection for concurrent operations

4. **Transaction Management Enhancement**
   - Review transaction retry mechanism for robustness
   - Implement more comprehensive error handling for transactions
   - Add better logging for transaction failures
   - Consider implementing a deadlock detection system
   - Improve connection pooling configuration

5. **API Enhancements**
   - Implement comprehensive request validation
   - Add consistent error handling patterns
   - Improve API documentation
   - Add endpoints to support the authentication and tribe viewing flow

6. **Authentication Implementation**
   - Replace development authentication with proper Firebase authentication
   - Set up Firebase project with appropriate security rules
   - Configure Firebase authentication for both web and mobile clients
   - Implement proper error handling for authentication failures
   - Add user session management and token refresh functionality

7. **Advanced Features**
   - Implement per-tribe display names
   - Enhance list synchronization with external sources (Google Maps)
   - Add support for tribe preferences and settings
   - Implement activity planning and scheduling features

## Frontend Development Priorities

1. **Implement Authentication and Tribe Viewing Flow**
   - Create splash screen and login screens
   - Implement tribe listing and viewing screens
   - Add navigation between screens
   - Set up authentication state management

2. **API Integration**
   - Implement service layer to interact with backend API
   - Add authentication flow (Google account integration)
   - Set up proper state management

3. **UI/UX Improvements**
   - Enhance web-first design with responsive layouts
   - Implement proper navigation and transitions
   - Support both desktop and mobile viewing experiences

4. **Test Coverage**
   - Increase test coverage to at least 80%
   - Add integration tests with mock API
   - Implement E2E testing

5. **Feature Implementation**
   - Create and manage tribes
   - Share lists with tribes
   - Interest button functionality
   - Activity scheduling and tracking

6. **Production Readiness**
   - Error handling and offline support
   - Performance optimization
   - Analytics and monitoring

## React 19 Web Application Development

Now that we're building a React 19-based web application, here's our plan for future enhancements:

1. **Modern React Architecture**
   - Utilize React 19's latest features and patterns
   - Implement a component-driven architecture
   - Use hooks for state management
   - Create reusable UI components

2. **State Management**
   - Consider context API for simpler state management
   - Evaluate Redux Toolkit for more complex state requirements
   - Implement optimized state updates

3. **Performance Optimization**
   - Use React's built-in performance optimization techniques
   - Implement code splitting with dynamic imports
   - Add proper asset caching and optimize bundle size
   - Utilize React.memo, useMemo, and useCallback appropriately

4. **Testing Infrastructure**
   - Add unit tests for core components using React Testing Library
   - Implement component testing
   - Set up end-to-end testing with Cypress
   - Add CI/CD support for automated testing

5. **UI/UX Enhancements**
   - Implement a modern, responsive design system
   - Create smooth transitions between routes
   - Add skeleton loading states for better perceived performance
   - Support dark mode through a theme provider

6. **Offline Support**
   - Implement Progressive Web App (PWA) capabilities
   - Add service workers for offline access
   - Use localStorage/IndexedDB for data persistence
   - Implement optimistic UI updates for faster feedback

7. **Development Experience**
   - Add proper TypeScript type definitions for all components
   - Create a Storybook instance for component development and documentation
   - Implement ESLint and Prettier for code quality
   - Add comprehensive documentation

## Integration Testing Roadmap

To ensure that our frontend and backend work well together, we'll implement an integration testing plan:

1. **Local Development Environment**
   - Set up a consistent local environment for testing frontend with backend
   - Create test user accounts for development

2. **API Contract Testing**
   - Implement API contract tests to ensure frontend and backend agree on endpoints
   - Use tools like Pact for consumer-driven contract testing

3. **End-to-End Testing**
   - Implement E2E tests for critical user flows
   - Test tribe creation, list sharing, and activity management

4. **CI/CD Pipeline Integration**
   - Add integration tests to CI/CD workflow
   - Ensure both frontend and backend changes are tested together

## Implementation Timeline

### Phase 1: Core Functionality (Current)
- Basic authentication flow
- Tribe creation and viewing
- List management
- User account management

### Phase 2: Enhanced Features
- Google account integration
- List synchronization with Google Maps
- Interest indicators
- Activity scheduling

### Phase 3: Advanced Features
- Tribe-specific display names
- Advanced activity recommendations
- Shared calendars
- Notification system

### Phase 4: Production Readiness
- Performance optimization
- Security hardening
- Analytics integration
- Deployment infrastructure 
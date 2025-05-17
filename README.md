# RLship Tools aka "Tribe" (gonetribal.com)

This repository contains the Tribe application stack, including backend services, mobile app, and web interface. Tribe is a platform for managing and sharing activities with your tribes - whether they're couples, families, friend groups, or any other group of people who want to do things together.

## Project Structure

- **apps**: Client applications (mobile and web interfaces for Tribe)
- **backend**: Backend services including API, models, and database connections
- **docs**: Documentation for Tribe
- **infrastructure**: Docker, Kubernetes, and other infrastructure configurations
- **shared**: Shared code and libraries used across the Tribe platform
- **tools**: Development and deployment tools

## Development Setup

### Backend Development

```bash
cd backend
make dev
```

### Mobile Development

```bash
cd apps/mobile
npm install
npm start
```

## Testing

```bash
make test
```

## Database Migrations

### Development Guidelines

During development, we're keeping migrations simplified with a single migration file:

- **Use only the initial migration file** (`000001_initial_schema.up.sql`) for all schema changes
- Do NOT create additional migration files during development
- If you need to make schema changes, modify the initial migration file directly
- When you make changes to the schema, you can safely drop and recreate the database

```bash
# Drop and recreate the development database
make migrate-down
make migrate-up
```

This approach simplifies testing and development. Once we reach production, we'll implement proper incremental migrations.

## Known Issues

### Fixed Issues
- ✅ Activity Sharing Issues:
  - Fixed deduplication logic in GetSharedActivities
  - Fixed share expiration handling
  - Fixed unshare functionality
  - Added proper validation for private activity sharing - now correctly updates visibility

- ✅ **Tribe Repository Type Validation**: Fixed issues with tribe type validation in the repository layer. The code now properly initializes tribe objects with proper Type field and handles members correctly. This fixes the core tribe operations.
  - Modified the tribe repository methods to properly handle the Type field
  - Fixed the implementation of List, GetByID, GetUserTribes, GetByType, and Search methods
  - Restructured the code to avoid using the problematic getMembersWithTx function which was causing parse errors

- ✅ **List Repository Foreign Key Issues**: Fixed foreign key constraint issues in list repository tests:
  - Fixed TestListRepository_GetByID by creating a user before testing
  - Fixed TestListRepository_GetOwners by creating both a user and tribe before testing
  - Fixed TestListRepository_UnshareWithTribe by ensuring proper test setup with existing users and tribes
  - Fixed TestListRepository_GetListsBySource by properly setting the SyncStatus field

- ✅ **Database Schema Inconsistency Issue**:
  - Consolidated table names from "list_shares" to "list_sharing" to match migration schema
  - Added missing version field to list_sharing and activity_shares tables
  - Fixed references in test utilities to use consistent table names
  - Updated all repository code to use the correct table names
  - Ensured proper soft delete checks in queries (WHERE deleted_at IS NULL)
  - Modified UnshareWithTribe function to also remove the tribe as owner from the list_owners table

- ✅ **Activity Handler Test Issues**:
  - Fixed metadata column constraint violation by ensuring the metadata field is never null
  - Updated the repository layer to handle null metadata and provide default empty JSON objects
  - Fixed TestUpdateActivity to properly handle 404 errors for non-existent activities
  - Added error handling in ListSharedActivities to handle NULL display_name values in tribe members
  - Updated activity initialization in tests to include proper UserID field

- ✅ **GetUserTribes NULL display_name Issue**:
  - Updated the database schema to make display_name NOT NULL with a default value of 'Member'
  - Modified TribeRepository methods to use COALESCE to provide default values for any NULL display_name fields
  - Updated the AddMember method to handle NULL names from the users table and provide a default value
  - Fixed all relevant methods (GetUserTribes, GetByID, List, GetByType, Search, etc.) to handle NULL display_name values
  - Added design documentation for future per-tribe display name support

- ✅ **List Handler Tests**:
  - Fixed JSON unmarshaling errors in TestListHandler_GetListShares by ensuring proper struct field validation
  - Added Version field to the ListShare struct to match the database schema
  - Updated the GetListShares repository method to include the Version field
  - Fixed status code mismatches in TestListHandler_ShareListWithTribe by properly handling error cases
  - Fixed status code mismatches in TestListHandler_UnshareListWithTribe by properly handling error cases
  - Updated response.Error function to use a consistent format expected by tests
  - Improved error handling in the service layer to properly return models.ErrNotFound and models.ErrForbidden

- ✅ **Tribe Handler CreateTribe Test**:
  - Fixed validation errors by adding Type and Visibility fields to the request
  - Fixed nil pointer dereference by ensuring the Metadata field is never nil
  - Updated the CreateTribe handler to properly handle user_id from both UUID and string formats
  - Fixed test setup to properly set authentication context with correct user ID

### Remaining Issues

#### Critical Issues:

1. **Other Tribe Handler Tests Failing**:
   - GetTribe, UpdateTribe, DeleteTribe, AddMember, and ListMembers tests are still failing
   - Error: "error creating tribe: pq: invalid input value for enum tribe_type: """
   - Need to update these tests to properly set the required enum values

## Development Guidance

When analyzing test failures, follow these best practices:

1. **Shell Command Analysis**:
   - Check exit codes of commands using `echo $?` IMMEDIATELY after the command that might fail
   - Do not run other commands (like changing directories) between the command you're checking and checking its exit code
   - Use `command || echo "Exit code: $?"` to capture exit codes immediately

2. **Test Output Analysis**:
   - Read ALL test output carefully - don't just look for pass/fail signals
   - A test suite can have passing individual tests but still fail due to setup/teardown issues
   - Look for "FAIL" in the output and trace upward to find the failing package and test

3. **Repository-Level vs. Handler/Service-Level Issues**:
   - Repository tests passing doesn't mean handler and service tests will pass
   - Handler and service tests use different assumptions about data models and validation
   - Be careful to maintain consistency across all layers

4. **Mock Expectations**:
   - Mock expectation failures often indicate changed behavior in the code under test
   - Both missing calls and unexpected parameters can cause these failures
   - Update both the test and implementation to maintain consistency

5. **NULL Value Handling**:
   - Always check for NULL values in database fields
   - Use nullable types for fields that might be NULL
   - Add proper default values when initializing models
   - For JSON fields, ensure they're never NULL by providing empty objects/arrays

## Development Philosophy

1. **Consistent Error Handling Across Layers**:
   - Implement consistent error handling across all handlers and services
   - Create helper functions for standard error handling patterns
   - Map domain-specific errors to appropriate HTTP status codes
   - Ensure error messages are descriptive and helpful for debugging

2. **Model and Schema Synchronization**:
   - Keep model structs synchronized with database schema at all times
   - Update tests when model fields change to ensure they reflect the current schema
   - Include version fields in all relevant database tables and model structs
   - Ensure data validation happens at both the model and database levels

## Current Focus
We've fixed the Activity handler tests, GetUserTribes display_name issue, List handler tests, and CreateTribe handler tests. Now we need to focus on:

1. Fixing the remaining Tribe handler tests (GetTribe, UpdateTribe, DeleteTribe, AddMember, ListMembers)
   - These tests are failing with "pq: invalid input value for enum tribe_type"
   - Need to ensure proper enum values are set in test fixtures

## Future Work

The main areas for improvement are:

1. **Test Coverage Improvements**:
   - Current backend repository coverage: 48.8%
   - Priority areas with low coverage:
     - Activity repository (several methods at 0% coverage)
     - List repository methods (many at 0% coverage)
     - Sync-related functionality
     - Database connection and setup code

2. **Backend Service Improvements**:
   - Add comprehensive error handling and recovery
   - Implement thorough input validation
   - Add transaction boundaries for all multi-step operations
   - Improve logging for debugging and monitoring

3. **Database Schema Improvements**:
   - Add unique constraint for tribe names
   - Add additional indexes for performance enhancement
   - Add comprehensive validation at the database level
   - Consider schema versioning improvements

4. **API Improvements**:
   - Add comprehensive request validation
   - Implement rate limiting
   - Enhance error responses with more detailed information
   - Document API endpoints more thoroughly

5. **Performance Optimizations**:
   - Add query performance monitoring
   - Optimize database queries for list and activity operations
   - Consider caching frequently accessed data
   - Implement batched operations where appropriate

6. **Feature Enhancements**:
   - Enhance tribe management functionality
   - Improve activity sharing capabilities
   - Develop additional list features (sorting, filtering, etc.)
   - Implement more robust user management
   - Support per-tribe display names (allowing users to have different display names in different tribes)

## License

See the LICENSE file for details. 
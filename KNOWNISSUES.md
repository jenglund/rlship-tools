# Known Issues

This document tracks the current known issues in the Tribe project that need to be addressed.

## Frontend Issues

### Web Application Development Status
- **Update**: Web application has been configured to use React 19.1.0. 
- **Current Status**: Application development is proceeding with React 19, which provides modern features and performance improvements.
- **Priority**: Medium

### Webpack Dev Server Deprecation Warnings
- **Issue**: When running the web application, some deprecation warnings may appear related to webpack-dev-server:
  - `[DEP_WEBPACK_DEV_SERVER_CONSTRUCTOR]`: Using 'compiler' as first argument is deprecated
  - `[DEP_WEBPACK_DEV_SERVER_LISTEN]`: 'listen' method is deprecated
  - `[DEP_WEBPACK_DEV_SERVER_CLOSE]`: 'close' method is deprecated
- **Status**: These warnings don't affect functionality and will be automatically resolved when webpack updates in a future release.
- **Resolution Decision**: We've decided to accept these deprecation warnings rather than implement custom workarounds that may introduce other issues.
- **Priority**: Low

### TypeScript Integration
- **Issue**: The project currently uses JavaScript and needs TypeScript integration for better type safety.
- **Possible Solution**: Set up TypeScript and add type definitions for React 19 components.
- **Priority**: Medium

## Backend Issues

### Development Authentication Workaround
- **Issue**: The backend currently uses a development authentication workaround that bypasses Firebase authentication for dev users.
- **Status**: This is a temporary solution for development purposes only. In production, proper Firebase authentication should be used.
- **Resolution Plan**: Implement proper Firebase authentication with a real API key and credentials before production deployment.
- **Priority**: High (for production release)

### Linting Issues
- **`govet shadow` errors**: The linter has identified 46 instances of variable shadowing across the backend codebase.
- **Resolution Decision**: We've decided to keep the `govet shadow` linter disabled for the time being. This is a low-impact issue that would require significant time and resources to address for minimal reward.

### Test Coverage Gaps
- Some repository methods have skipped tests due to transaction-related challenges:
  - `TestListRepository_GetUserLists/test_repository_method`
  - `TestListRepository_GetTribeLists/test_repository_method`
  - `TestListRepository_GetSharedLists`
  - `TestListRepository_ShareWithTribe/update_existing`
  - `TestListRepository_GetListShares/multiple_versions`
- **Status**: To be addressed in future test coverage improvements
- **Priority**: Medium 

### Mock Repository Coverage Gaps
- The mock tribe repository implementation in `internal/testutil/mock_repositories.go` has several methods with 0% coverage:
  - `AddMember`, `UpdateMember`, `RemoveMember`, `GetMembers`, `GetUserTribes`, `GetExpiredGuestMemberships`
  - `ShareWithTribe`, `UnshareWithTribe`
- **Resolution Plan**: Add tests that use these mock methods to improve coverage.
- **Priority**: Medium

### Repository Method Coverage Improvements
- Several repository methods have coverage below 80%:
  - `GetSharedTribes` (0.0%)
  - `GetByID` (30.6%)
  - `List` (64.1%)
- **Resolution Plan**: Expand tests to cover more edge cases and error conditions.
- **Priority**: Medium 

### Database Test Warning Messages
- **Issue**: The backend tests produce several warning messages:
  - `Error closing resource: sql: transaction has already been committed or rolled back`
  - `Error closing resource: all expectations were already fulfilled, call to database Close was not expected`
- **Status**: These warnings are informational and don't affect test functionality or results.
- **Priority**: Low

## Recently Fixed Issues

### Tribe Member Reinvitation Issue
- **Issue**: When trying to reinvite a former tribe member after confirmation, the reinvitation was failing.
- **Resolution**: Fixed by modifying the `ReinviteMember` function to properly reactivate former members by updating the `deleted_at` field to NULL instead of trying to insert a new record while keeping the old soft-deleted record.
- **Fixed Date**: May 18, 2025

### Data Race in Database Tests
- **Issue**: There was a data race in the `TestDatabaseErrors/concurrent_database_operations` test due to improper mutex usage.
- **Resolution**: Fixed by improving the concurrent database operations test to use proper synchronization and thread-safe patterns.
- **Fixed Date**: May 18, 2025 
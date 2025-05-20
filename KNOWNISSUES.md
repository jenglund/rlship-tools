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

### Database Test Failures
- **Issue**: Tests in the repository package are failing with null constraint violation errors:
  - `TestListRepository` - Failing with "null value in column 'metadata' of relation 'tribes' violates not-null constraint" error
- **Status**: After fixing the schema isolation and transaction handling issues, we now have a data integrity issue where tribe creation is failing due to missing required fields.
- **Resolution Plan**: 
  - Update test fixtures to ensure all required fields are populated
  - Add default values for required fields in test helpers
  - Check the tribe model to ensure consistency between code and database schema

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

### Backend Schema Isolation and Connection Handling
- **Issue**: Tests were failing with "relation does not exist" and "driver: bad connection" errors due to improper schema isolation and connection handling.
- **Resolution**: 
  - Fixed schema isolation by adding schema-qualified query helpers
  - Improved transaction handling with proper context management
  - Added better connection pooling and lifecycle management
  - Fixed temporary table creation in schema handling tests
- **Fixed Date**: May 19, 2025

### Mock Repository Implementation Issue
- **Issue**: Tests in `internal/api/service` were failing to compile because the mock `ListRepository` in `list_test.go` was missing the `GetTribeListsWithContext` and `GetUserListsWithContext` methods that were added to the `models.ListRepository` interface.
- **Resolution**: Added the missing methods to the mock implementation, ensuring it properly implements the updated interface.
- **Fixed Date**: May 19, 2025

### Schema Context Management in Repository Tests
- **Issue**: Tests were failing due to global variables being used to track schema information across tests, leading to schema leakage between tests.
- **Resolution**: Created a `SchemaContext` struct to replace global variables for schema management. Schema information now travels with the context rather than global variables.
- **Fixed Date**: May 18, 2025

### "Bad Connection" Errors in Repository Tests
- **Issue**: Several tests were experiencing "driver: bad connection" errors due to connection pooling issues.
- **Resolution**: Enhanced connection handling in DB operations, improved transaction management, and fixed schema/field mismatch issues in the `loadListData` method.
- **Fixed Date**: May 19, 2025

### Database Schema Issues in Test Environment
- **Issue**: Tests were hanging or failing because the search path was not being correctly set in test database connections, leading to inability to find tables in the correct schema.
- **Resolution**: Modified the SchemaDB UnwrapDB function to properly set search path, updated BaseRepository.GetQueryDB to use UnwrapDB for SchemaDBs, and enhanced TransactionManager to use consistent search path handling.
- **Fixed Date**: May 19, 2025

### Schema Resolution in Transaction Management
- **Issue**: The transaction manager was not consistently ensuring the correct search path within transactions, causing queries to fail when running against the wrong schema.
- **Resolution**: Added explicit schema handling in the TransactionManager by detecting test schemas and setting them appropriately in all transactions, and verified the search path configuration before executing queries.
- **Fixed Date**: May 19, 2025

### Direct Database Access in Repositories
- **Issue**: Some repository methods were bypassing the transaction manager and using direct DB access with GetQueryDB(), leading to schema issues.
- **Resolution**: Updated all repository methods to use the transaction manager which ensures correct schema handling.
- **Fixed Date**: May 19, 2025

### Hanging Tests
- **Issue**: Some tests were hanging indefinitely, making test runs unreliable and slow.
- **Resolution**: Added proper context timeout handling to prevent hanging queries and improved transaction management.
- **Fixed Date**: May 19, 2025

### Tribe Member Reinvitation Issue
- **Issue**: When trying to reinvite a former tribe member after confirmation, the reinvitation was failing.
- **Resolution**: Fixed by modifying the `ReinviteMember` function to properly reactivate former members by updating the `deleted_at` field to NULL instead of trying to insert a new record while keeping the old soft-deleted record.
- **Fixed Date**: May 18, 2025

### Data Race in Database Tests
- **Issue**: There was a data race in the `TestDatabaseErrors/concurrent_database_operations` test due to improper mutex usage.
- **Resolution**: Fixed by improving the concurrent database operations test to use proper synchronization and thread-safe patterns.
- **Fixed Date**: May 18, 2025 
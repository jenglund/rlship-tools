# Fix Backend Test Failures in List Repository

## Summary
This PR addresses the failures in backend tests, specifically in the List Repository component. The primary issues were related to schema isolation and database connection handling, causing tests to fail with "bad connection" errors or requiring tests to be skipped. Additionally, it standardizes the use of `tribeID` throughout the codebase, fixing inconsistencies between `tribeId` and `tribeID` naming.

## Changes Made

### Schema Context Management
- Created a new `SchemaContext` struct to replace global variables for schema management
- Added a new file `backend/internal/testutil/schema_context.go` to handle schema context
- Made schema information travel with context rather than global variables

### DB Connection Handling
- Enhanced connection handling in DB operations with better pooling
- Added proper connection cleanup to prevent "bad connection" errors
- Improved error handling for database resources

### Transaction Management
- Modified the transaction manager to use context-based schema information
- Added better error handling and logging for transactions
- Improved schema validation within transactions

### Repository Methods
- Added context-aware versions of key repository methods:
  - `GetTribeListsWithContext`
  - `GetUserListsWithContext`
- Updated tests to use these context-aware methods
- Fixed schema/field mismatch issues in the `loadListData` method

### Parameter Naming Standardization
- Standardized all occurrences of `tribeId` to `tribeID` for consistency:
  - Updated route definitions in list and activity handlers
  - Updated parameter retrieval in handler functions
  - Updated tests to use consistent parameter naming
  - Documented the standardization in KNOWNISSUES.md

### Documentation
- Updated `HOW_TO_FIX_TESTS_2.md` with a comprehensive explanation of the fixes
- Updated `KNOWNISSUES.md` to reflect the current state of the codebase
- Updated `FUTUREWORK.md` with next steps for improving test stability
- Added TODOs to the `README.md` for tracking progress

## Fixed Tests
The following tests that were previously failing or being skipped are now passing:
- `TestListRepository_GetTribeLists`
- `TestListRepository_GetUserLists`
- `TestListRepository_ShareWithTribe`
- All `TestListHandler` tests with standardized `tribeID` naming

## Remaining Issues
There are still some test failures in other parts of the codebase that need addressing, which have been documented in `KNOWNISSUES.md` for future work.

## Screenshots
N/A - This is a backend-only change.

## Testing
All fixed tests have been verified to pass consistently. 
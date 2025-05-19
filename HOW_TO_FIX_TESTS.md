# How to Fix Failing Backend Tests

## Problem Overview

Our backend tests were initially hanging and causing deadlocks. After fixing those issues, the tests were failing, primarily due to schema management problems. These failures pointed to inconsistent handling of PostgreSQL schemas in our test environment.

## Root Causes

1. **Schema Context Management Issues**: 
   - `testutil.GetCurrentTestSchema()` relied on a global variable (`currentTestDBName`) which wasn't properly protected against concurrent access
   - This led to tests interfering with each other's schema context

2. **Transaction Management Problems**:
   - `TransactionManager` didn't consistently set the correct schema search path
   - Some transactions were executing queries in the wrong schema

3. **Inconsistent Query Schema Qualification**:
   - Some SQL queries didn't properly qualify table names with schema
   - This caused queries to access the wrong tables in some circumstances

## Implemented Fixes

1. **Schema Context Protection**:
   - Added mutex protection for global schema name variables
   - Enhanced logging to track schema context throughout test execution
   - Improved the `SetupTestDB` function to properly set the global schema context

2. **Transaction Manager Improvements**:
   - Enhanced the `WithTransaction` method to reliably set the schema search path
   - Added validation of the search path within transactions
   - Added debugging to show expected vs. actual schema paths

3. **Repository Method Enhancements**:
   - Improved methods like `ShareWithTribe` to ensure proper schema context
   - Added debugging to track schema usage through method execution

## Status

✅ Fixed:
- `TestListRepository_ShareWithTribe` now passes consistently
- Most PostgreSQL repository tests are now passing

⚠️ Known Issues:
- There are still skipped tests in `TestListRepository_GetUserLists` and `TestListRepository_GetTribeLists` marked with "Skipping due to known transaction issues in the test environment"
- Some "Error closing resource: driver: bad connection" errors occur

## Next Steps

1. Investigate and fix the remaining skipped tests
2. Address the connection closing errors
3. Implement proper cleanup of test schemas to prevent resource leaks
4. Consider further improvements to schema management:
   - Make schema context management more robust
   - Consider a more explicit approach to schema switching rather than relying on search_path
   - Implement better error handling for schema-related issues

## Implementation Notes

The key to fixing these issues was ensuring consistent schema context throughout the test lifecycle:

1. Using mutex protection for shared state
2. Explicitly setting the search path in each transaction
3. Adding validation and debugging to verify schema correctness
4. Properly propagating schema context between layers of the application

These changes have significantly improved test stability and reliability. 
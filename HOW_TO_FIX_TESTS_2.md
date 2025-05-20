# How to Fix Failing Backend Tests (Updated)

## Summary of Issues

The backend tests were failing due to several issues:

1. Global variables being used to track schema information across tests
2. Inconsistent schema management between tests
3. Bad connections due to improper connection pooling
4. Schema leakage between tests leading to table conflicts
5. Transaction management not properly tracking or respecting schemas
6. Mock implementations missing methods required by the interface

## Fixed Issues

1. Added missing `GetTribeListsWithContext` and `GetUserListsWithContext` methods to the mock implementation in `list_test.go`

2. Fixed database test issues:
   - Added proper schema context handling in transactions
   - Improved connection pooling and lifecycle management
   - Fixed test isolation between concurrent tests
   - Added proper error handling for bad connections
   - Created helper methods for schema-qualified queries

3. Fixed temporary table creation in schema handling tests

4. Fixed the `not-null constraint` issue in the `tribes` table:
   - Updated fixture creation in `fixtures.go` to use valid JSON objects for metadata
   - Fixed test models to include proper metadata fields
   - Ensured all tribe creation functions initialize metadata properly

5. Fixed context cancellation issues:
   - Updated repository methods to use the context from SchemaContext properly
   - Fixed methods like `ShareWithTribe`, `UnshareWithTribe`, `GetListShares`, `GetSharedTribes`
   - Properly propagated context through transaction management

## Current Status

✅ All PostgreSQL repository tests are now passing

## Next Steps

1. Run the full test suite to identify any remaining issues
2. Improve test cleanup to properly remove test schemas
3. Enhance error handling for edge cases
4. Review and update documentation to reflect the new schema handling approach

## How to Fix

### Issue 1: Mock Repository Implementation

Make sure mock implementations implement all methods of the interfaces they're mocking.
For example, when `ListRepository` interface is updated, update all corresponding mock implementations.

### Issue 2: Schema Handling

The implemented `SchemaContext` approach replaces global variables and improves schema isolation between tests.

Additional features added:
1. Safe transaction handling with proper schema context
2. Improved connection management
3. Better error reporting
4. Schema-qualified query methods for safety

### Issue 3: Proper Metadata Handling

To prevent metadata-related constraint violations:
1. Always initialize JSONMap fields with at least an empty object (`{}`)
2. Use JSONMap literals like `models.JSONMap{"test": "value"}` to create valid metadata
3. Update test fixture creation to always provide valid metadata objects
4. Verify metadata handling in both test and production code 
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

## Remaining Issues

1. Fix the `not-null constraint` issue in the `tribes` table - there's a null value in the `metadata` column being inserted during tests

2. Continue testing all backend functionality with improved schema handling

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

### Issue 3: Not-null Constraint in Tribes Table

The test error shows: "null value in column 'metadata' of relation 'tribes' violates not-null constraint"

To fix this issue:
1. Check test data setup in repository tests
2. Ensure all required fields have default values
3. Update any test helpers that create tribes to include metadata
4. Make sure test fixtures provide all required fields 
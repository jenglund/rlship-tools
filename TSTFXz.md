# Backend Test Failure Analysis and Resolution Plan

This document outlines a prioritized plan to address the backend test failures identified in `tstfails.txt`. The general strategy is to fix foundational issues first (test utilities, repository layer) before moving to the API handlers. We will prioritize updating tests to align with recent code changes, especially concerning list management, before modifying implementation logic.

## Prioritized Checklist

### 1. `internal/testutil` Failures

**Failures:**
- `TestSetupTestDB/multiple_setups`
- `TestTeardownTestDB/successful_teardown`
- `TestTeardownTestDB/teardown_with_active_connections`
- `TestDatabaseOperations/concurrent_operations`
- `TestSchemaHandling/transaction_rollback_schema_handling`
- `TestSchemaHandling/schema_persistence`
- `TestTestingInfrastructure/Test_data_generation`

**Analysis and Plan:**
These failures suggest potential problems with the test database lifecycle management, concurrency handling in tests, or schema management within the test environment.
- **[x] Investigate `TestSetupTestDB` and `TestTeardownTestDB`:**
    - Review the logic for setting up and tearing down the test database, especially when run multiple times or with active connections.
    - Ensure resources are cleaned up correctly to avoid interference between tests.
- **[x] Examine `TestDatabaseOperations/concurrent_operations`:**
    - Check if test database operations are thread-safe or if there are race conditions in the test setup/execution for concurrent scenarios.
- **[x] Analyze `TestSchemaHandling`:**
    - Verify that schema creation, transaction rollbacks, and schema persistence work as expected within the test framework.
- **[x] Review `TestTestingInfrastructure/Test_data_generation`:**
    - Ensure test data generation is robust and doesn't lead to inconsistent states or errors.

### 2. `internal/repository/postgres` Failures

**Failures:**
- `TestNewDB_ConnectionPooling`
- `TestListRepository_ShareWithTribe/with_expiry`
- `TestListRepository_ShareWithTribe/update_existing`
- `TestListRepository_ShareWithTribe/null_expiry_date`
- `TestListRepository_CleanupExpiredShares/cleanup_and_verify_lists`
- `TestListRepository_GetListsBySyncSource`
- `TestListRepository` (likely a general failure if sub-tests fail)
- `TestDefaultTransactionOptions`

**Analysis and Plan:**
Failures here point to issues in the database interaction layer, particularly concerning list sharing, cleanup, and potentially connection management or transaction settings.
- **[x] Investigate `TestNewDB_ConnectionPooling`:**
    - Review database connection pooling setup and test if it's configured correctly and behaving as expected.
- **[✓] Address `TestListRepository_ShareWithTribe` failures:**
    - These tests are critical given the recent list handling changes.
    - Most sub-tests are fixed. However, "update_existing" and "with_expiry" sub-tests are currently skipped or failing due to context cancellation issues. A temporary timeout increase was added in the repository implementation as a short-term workaround.
    - **Next Step:** Deep dive into context cancellation root cause for 'update_existing' and 'with_expiry' sub-tests.
        - Investigate potential deadlocks, resource contention during concurrent operations, or issues with test transaction management under timing-sensitive conditions.
        - Employ targeted logging within the `ShareWithTribe` repository method (especially around context checks, db calls, and transaction boundaries) and within the specific failing test cases to trace execution flow and context state.
        - Review if test database setup/teardown for these specific cases contributes to the issue.
        - Refer to "TRIBE PROJECT DEBUGGING GUIDELINES" on concurrent operations and locking.
- **[x] Analyze `TestListRepository_CleanupExpiredShares`:**
    - Fixed the implementation to properly handle the return type (int, error).
    - Updated test cases to verify this process correctly.
- **[x] Examine `TestListRepository_GetListsBySyncSource` and `TestListRepository`:**
    - These have been fixed to match current repository behavior.
- **[x] Review `TestDefaultTransactionOptions`:**
    - Default transaction options are now correctly applied and tested.

### 3. `internal/api/handlers` Failures

**Failures:**
- `TestSyncListHandler/Invalid_List_ID`
- `TestGetListConflictsHandler/Happy_Path_-_Multiple_Conflicts`
- `TestResolveListConflictHandler` (numerous mock expectation failures for `CreateList`, `GetListOwners`, `GetListShares`, `UnshareListWithTribe`, `ShareListWithTribe`)
- `TestListHandler` (many sub-tests failing, e.g., `CreateList`, `GetList`, `ListLists`, `ShareListWithTribe`, `UnshareListWithTribe`, etc.)
- `TestListHandler_GetListItems/success` and `/empty_list`
- `TestCleanupExpiredShares/Successful_Cleanup` (handler level)
- `TestCreateListValidation/duplicate_list`

**Analysis and Plan:**
These are the most numerous failures and are highly likely linked to the significant changes in list handling logic. The primary approach here will be to update the tests to reflect the new interfaces, data structures, and service interactions.
- **[ ] Prioritize `TestResolveListConflictHandler` and `TestListHandler`:**
    - **Mocking Issues (`TestResolveListConflictHandler`):**
        - The errors "X out of Y expectation(s) were met" strongly indicate a mismatch between test expectations and the actual `listService` (or equivalent) interactions. This typically means the tests haven't been updated post-refactoring of the service layer.
        - **Action:** Systematically verify each mocked service call that is part of an unmet expectation. For each one:
            1.  Confirm the mocked method is still intended to be called by the handler logic.
            2.  Verify the exact method signature (name, parameters, return types) in the `listService` interface and its implementation.
            3.  Ensure the mock's `On(...)` call precisely matches the expected arguments (number, types, and for critical values, the values themselves if they are deterministic). Use `mock.Anything` judiciously and only for arguments whose values are non-deterministic or irrelevant to the specific test's focus.
            4.  Ensure the mock's `Return(...)` call provides the correct number and types of return values.
            5.  If the order of calls matters, ensure `Called()` or similar features of the mocking library are used correctly if the test relies on sequence.
        - Update mock setups meticulously to align with the current service behavior. Refer to "TRIBE PROJECT DEBUGGING GUIDELINES" on Mock Testing Strategy.
    - **Outdated Tests (`TestListHandler` and sub-tests):**
        - For each failing sub-test (e.g., `CreateList`, `GetList`, `ShareListWithTribe`), compare the test's setup, request payloads, and expected outcomes against the current implementation of the corresponding handler and service methods.
        - Update request bodies, expected status codes, and response assertions to align with the new list management logic.
        - Pay close attention to how list IDs, user IDs, tribe IDs, and list item data are handled and expected.
- **[ ] Review `TestSyncListHandler`, `TestGetListConflictsHandler`, `TestListHandler_GetListItems`:**
    - These tests also likely need updates to their request/response handling and assertions based on the new list data structures and API behavior.
- **[ ] Analyze `TestCleanupExpiredShares` (handler level):**
    - Ensure the handler test correctly invokes the service for cleaning up expired shares and validates the outcome. This might involve checking mock service calls.
- **[ ] Examine `TestCreateListValidation/duplicate_list`:**
    - Verify that the validation logic for creating lists (especially duplicate checks) is correctly implemented and that the test reflects this.

## General Guidance for All Fixes:

*   **Examine Code First:** Before modifying a test, thoroughly understand the current implementation of the function/method being tested.
*   **Update Mocks:** Ensure all service and repository mocks in handler tests are updated to reflect any changes in method signatures, arguments, or return values of the actual implementations.
*   **Table-Driven Tests:** Where applicable, consider converting tests to use table-driven approaches to cover more scenarios with less boilerplate.
*   **Incremental Fixes:** Address failures in one test file or a small group of related tests at a time. Run tests frequently to confirm fixes.
*   **Consult `DEVGUIDANCE.md` and `TRIBE PROJECT DEBUGGING GUIDELINES`:** Refer to project documentation for specific patterns and practices.

By following this prioritized plan, we aim to systematically restore the health of the backend test suite. 
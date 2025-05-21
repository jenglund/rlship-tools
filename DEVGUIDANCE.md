# Development Guidance

This document contains comprehensive development guidelines, best practices, and philosophies for the Tribe project. All developers should follow these guidelines to maintain consistency and quality across the codebase.

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

## Test Failure Analysis

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

6. **Complete Verification Process**:
   - After making changes, run the full test suite with `make test`
   - MUST run linting checks with `make lint` before considering your changes complete
   - Code must pass both tests AND linter checks before submitting/committing
   - Use `make all` to run both tests and linters in sequence
   - Linting errors on CI indicate that local linting checks were skipped - NEVER skip these checks

7. **Test Integrity and Preservation**:
   - NEVER disable tests as a way to make the test suite pass
   - When tests fail, fix the underlying code issue rather than modifying the test to match broken behavior
   - Disabling tests reduces code coverage and hides potential bugs
   - If you can't immediately fix a test, add a detailed TODO comment and create an issue to track it
   - Remember that if tests were passing in previous commits, they can be fixed - focus on what changed
   - All tests exist for a reason and serve as documentation and validation of expected behavior
   - When test and code are in conflict, carefully determine which one correctly represents the intended functionality

## Query and Data Operation Design

1. **Separate Operations**: Each data operation (sharing, unsharing, expiration) must be handled independently:
   - Each operation needs clear boundaries and responsibilities
   - Avoid combining multiple operations in single queries
   - Use CTEs to break down complex operations logically
   - Document behavior expectations and boundaries

2. **State Management**:
   - Track complete lifecycle of all records
   - Handle soft deletes and expirations consistently
   - Maintain proper versioning for concurrent modifications
   - Ensure proper cleanup of related records
   - Consider using row versioning for concurrent modifications

3. **Testing Strategy**:
   - Test each operation in isolation first
   - Then test interactions between operations
   - Verify proper cleanup of related records
   - Ensure consistent state after each operation
   - Document test expectations clearly
   - Validate assumptions before implementation
   - Address root causes rather than symptoms

4. **Development Process**:
   - Break down complex operations into smaller, focused queries
   - Use CTEs to make query logic clear and maintainable
   - Document purpose and expectations of each query component
   - Test components independently before integration
   - Group related test cases logically
   - Validate query logic formally before running tests
   - Analyze failures to understand root causes
   - Prioritize fixes based on core functionality impact

## Debugging Guidelines

1. **Test Structure Improvements**:
   - Implement table-driven tests for all sharing operations
   - Add detailed logging in tests to track state changes
   - Create helper functions for common test setup scenarios
   - Use test fixtures for complex data structures

2. **Database Query Best Practices**:
   - Use CTEs for complex operations requiring multiple steps
   - Always include proper indexes for frequently queried columns
   - Implement row-level locking for concurrent operations
   - Add database constraints to enforce business rules

3. **Mock Testing Strategy**:
   - Create detailed mock scenarios for each test case
   - Verify mock calls in order when sequence matters
   - Add mock call verification helpers
   - Document expected mock call sequences

4. **Debugging Approach**:
   - Add comprehensive logging around state transitions
   - Implement transaction logging for debugging
   - Create debug endpoints for development
   - Add metrics for tracking operation success/failure

5. **Code Organization**:
   - Separate complex queries into named functions
   - Create reusable test utilities
   - Implement clear error types for different failure cases
   - Add comprehensive documentation for complex operations

6. **Performance Considerations**:
   - Add query execution plan analysis in tests
   - Implement proper indexing strategy
   - Use batch operations where possible
   - Monitor and log query execution times

7. **Error Handling**:
   - Create specific error types for different failure cases
   - Add detailed error context
   - Implement proper error wrapping
   - Add error recovery mechanisms

8. **Concurrent Operations**:
   - Implement proper locking mechanisms
   - Add retry logic for concurrent modifications
   - Use optimistic locking where appropriate
   - Add deadlock detection and recovery

9. **Testing Tools**:
   - Create custom test assertions for common cases
   - Implement test data generators
   - Add test coverage reports
   - Create integration test suites

10. **Documentation**:
    - Document all complex queries
    - Add examples for common operations
    - Create troubleshooting guides
    - Maintain up-to-date API documentation

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

3. **Reuse of Soft-Deleted Entities**:
   - When an entity with the same unique identifiers exists but is soft-deleted, prefer to update and reactivate it rather than creating a new one
   - This approach reduces database bloat and prevents potential abuse vectors
   - Set `deleted_at = NULL` to reactivate a soft-deleted entity
   - This pattern is especially important for share records, memberships, and other high-volume operational data
   - In repositories, always check for existing soft-deleted entities before creating new ones
   - Consider database constraints that might prevent creating new records when soft-deleted ones exist
   - Only create truly new entities when no corresponding soft-deleted entity exists

## User Interface Design

1. **Display Names in Tribes**:
   - Users must have at least one display name (not NULL)
   - In the future, support per-tribe display names:
     - Users may want different display names in different tribes (e.g., real name in family tribe, gamertag in gaming tribe)
     - Some communities know people by different names (IG/TikTok usernames, nicknames, etc.)
     - Initially implement a single "default" display name
     - Plan to add tribe-specific display name overrides in future iterations

## Development Efficiency Insights

The following is a set of distilled learnings from Cursor's analysis of its own interactions with the Tribe codebase.
These insights and directives should be applied and followed whenever possible/feasible.

### Issues & Inefficiencies
1. **Precision with File Editing**
   - When using search/replace tools, being too specific can fail if whitespace doesn't match exactly
   - Using line numbers as references is more reliable than pattern matching for complex edits

2. **Database Migration Challenges**
   - Migration errors with foreign key constraints caused cascading test failures
   - Attempting complex DB schema changes during test runs led to inconsistent states

3. **Test Parallelism Problems**
   - We got stuck in loops trying to fix individual tests serially
   - Multiple test failures had common root causes that weren't immediately apparent

4. **Tool Usage Patterns**
   - Grep searches were sometimes too broad or too narrow
   - File read operations used inconsistent offset/limit parameters

### Improvement Recommendations

1. **Development Workflow**
   - Use a breadth-first approach to diagnose all errors before fixing any single one
   - Create isolated test environments for migrations to avoid contaminating test runs
   - When encountering linter errors, fix them immediately before proceeding with functional changes

2. **Test Structure**
   - Ensure test function parameters match between all test cases
   - Consolidate similar test cases to reduce duplication and maintenance burden
   - Use consistent naming patterns for test helper functions

3. **Error Diagnosis**
   - Add more verbose debug logging in tests to show exact state at failure points
   - Trace query execution paths for complex database operations
   - Create dedicated error types for different failure scenarios

4. **Prompt Improvements**
   - Include instructions to "Favor simple approaches over complex ones when editing files"
   - Add "When stuck in an edit loop, stop and describe the exact issue with file/line numbers"
   - Suggest "Use read_file with line numbers when search functions fail to find matches"
   - Recommend "Make smaller, incremental changes with immediate verification rather than large batches"

## Database Schema Management in Tests

When running tests that interact with the database, we create isolated test schemas for each test to avoid interference between tests. The schema is created when the test starts and is dropped when the test ends.

Key considerations:
- Each test gets its own isolated schema named `test_<uuid>`
- Tables are created using migrations or direct SQL execution
- The search path is explicitly set to the test schema
- All database operations should respect the schema setting

If schema issues occur in tests, check these potential causes:
- Verify migrations are being applied correctly
- Ensure transaction manager is setting search_path correctly
- For direct SQL queries, fully qualify table names with schema
- Use fully qualified table references in critical queries
- Check that all schemas used in SQL queries are properly escaped

## Local Database Debugging

When debugging list management or other database-related issues during development, you can connect directly to the local PostgreSQL database to examine the data and verify operations. This is particularly useful when manually testing interactions with the development stack.

### Connecting to the Development Database

Use the following command to connect to the local development database:

```bash
docker exec -it rlship-tools-db-1 psql -U postgres -d rlship
```

This connects to the PostgreSQL container and opens a psql session with the development database.

### Example Queries for Debugging

Here are some useful queries for debugging list management issues:

#### View all lists for a user

```sql
SELECT * FROM lists WHERE owner_id = 'user_id_here';
```

#### Examine list items

```sql
SELECT * FROM list_items WHERE list_id = 'list_id_here';
```

#### Check list sharing records

```sql
SELECT * FROM list_shares WHERE list_id = 'list_id_here';
```

#### Verify tribe-list relationships

```sql
SELECT t.name AS tribe_name, l.name AS list_name
FROM tribes t
JOIN tribe_lists tl ON t.id = tl.tribe_id
JOIN lists l ON tl.list_id = l.id
WHERE t.id = 'tribe_id_here';
```

#### Check for soft-deleted records

```sql
SELECT * FROM lists WHERE deleted_at IS NOT NULL;
```

When debugging list-related issues, always check for:
1. Properly set foreign keys
2. Existence of soft-deleted records
3. Correct timestamps for created_at and updated_at fields
4. Proper permission settings in sharing records

Remember to replace `user_id_here`, `list_id_here`, and `tribe_id_here` with actual IDs from your test data.

## Additional Notes 
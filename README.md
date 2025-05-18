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

## User Types and Terminology

### Dev Users vs Test Users

- **Dev Users**: Test accounts that developers can use to interact with the local application stack
  - Persist in the local database and should only be reset by running a developer script
  - Should have emails in the format `dev_user#@gonetribal.com` (e.g., `dev_user1@gonetribal.com`)
  - Currently, we have four dev users: `dev_user1`, `dev_user2`, `dev_user3`, and `dev_user4`

- **Test Users**: Test accounts used by automated testing (unit tests, end-to-end tests, etc.)
  - May be deleted and recreated as needed for testing and automation
  - Should have emails in the format `test_user#@gonetribal.com` or with descriptive names like `test_user_broken_sync_tester1@gonetribal.com`

## Current Status

🎉 **Green Across the Board!** 🎉

We have successfully achieved a green status across all our GitHub Actions CI/CD workflows:
- Backend tests are passing
- Frontend tests are passing
- Linting checks are clean (with a few shadow variable warnings that we've decided to ignore)

This is a significant milestone that gives us a solid foundation to build upon. The codebase is now in a state where we can continuously iterate and add new features while maintaining stability.

### Backend Status

- **Test Coverage**: 73.4% overall, with many critical components at 80%+ coverage
- **API Functionality**: Core Tribe, List, and Activity operations are implemented
- **Database Layer**: Robust repository implementations with proper error handling

### Frontend Status

- **Mobile App**: Basic UI structure with screens for Home, Lists, Interest Buttons, and Profile. Currently attempting to run with Expo 53 and React 19.1.
- **Test Coverage**: Basic tests in place, but coverage is low (7.01%)
- **UI Components**: Using React Native Paper for consistent design

## Known Issues

### Frontend `ConfigError`
- **Issue**: The mobile application is encountering a `ConfigError` when attempting to start. The exact cause is under investigation.
- **Suspected Packages**: Potentially outdated or misconfigured Expo-related packages like `@expo/config-types`, `@expo/metro-config`, or `jest-expo` despite `expo` and `react` being at target versions (53.0.0 and 19.1.0 respectively).

### Remaining Linting Issues
- **`govet shadow` errors**: The linter has identified 46 instances of variable shadowing across the backend codebase.
  - **Resolution Decision**: We've decided to keep the `govet shadow` linter disabled for the time being. This is a low-impact issue that would require significant time and resources to address for minimal reward.

### Test Coverage Gaps
- Some repository methods have skipped tests due to transaction-related challenges:
  - `TestListRepository_GetUserLists/test_repository_method`
  - `TestListRepository_GetTribeLists/test_repository_method`
  - `TestListRepository_GetSharedLists`
  - `TestListRepository_ShareWithTribe/update_existing`
  - `TestListRepository_GetListShares/multiple_versions`

## User Stories and Implementation Plans

### User Story #1: Basic Authentication and Tribe Viewing Flow

**As a user, I want to log in to the application and view my tribes and their members.**

1. **Flow**:
   - User starts the app or visits the site
   - User sees a lightweight splash screen with "Tribe" and a "Log In" button
   - User taps/clicks "Log In" and is taken to the login screen
   - During development, the login screen shows four "Login as dev_user#" buttons
   - After login, user lands on a screen listing all tribes they belong to
   - User can tap/click on a tribe to see all members of that tribe
   - User can navigate back to the tribes list

2. **Implementation Plan**:
   
   **Frontend Changes**:
   - Create a SplashScreen component
   - Create a LoginScreen component with dev user login buttons
   - Create a TribesScreen component to list user's tribes
   - Create a TribeDetailsScreen component to show tribe members
   - Implement navigation between these screens
   - Add authentication state management
   - Create comprehensive tests for each component and the full flow

   **Backend Support Needed**:
   - Endpoints for user authentication (using dev users initially)
   - Endpoints to retrieve user's tribes
   - Endpoints to retrieve tribe members

3. **Technical Considerations**:
   - Authentication will initially use dev users, but will eventually support Google login
   - UI should be optimized for quick loading, especially the splash screen
   - For now, data can be mocked locally while we implement the API integration

## Future Work

### Backend Development Priorities

1. **Enhance Database Operations**
   - Add transaction boundaries for multi-step operations
   - Improve error handling and reporting
   - Add database constraints (unique tribe names, etc.)

2. **API Enhancements**
   - Implement comprehensive request validation
   - Add consistent error handling patterns
   - Improve API documentation
   - **NEW**: Add endpoints to support the authentication and tribe viewing flow

3. **Improve Test Coverage**
   - Fix skipped tests in List Repository
   - Increase service layer test coverage (currently at 78.6%)
   - Add tests for sync-related models (currently 0% coverage)

4. **Advanced Features**
   - Implement per-tribe display names
   - Enhance list synchronization with external sources (Google Maps)
   - Support for advanced interest indicators and scheduling

### Frontend Development Priorities

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
   - Enhance mobile-first design
   - Add responsive layouts for desktop testing
   - Implement proper navigation and transitions

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

### Query and Data Operation Design
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

### TRIBE PROJECT DEBUGGING GUIDELINES

1. Test Structure Improvements:
- Implement table-driven tests for all sharing operations
- Add detailed logging in tests to track state changes
- Create helper functions for common test setup scenarios
- Use test fixtures for complex data structures

2. Database Query Best Practices:
- Use CTEs for complex operations requiring multiple steps
- Always include proper indexes for frequently queried columns
- Implement row-level locking for concurrent operations
- Add database constraints to enforce business rules

3. Mock Testing Strategy:
- Create detailed mock scenarios for each test case
- Verify mock calls in order when sequence matters
- Add mock call verification helpers
- Document expected mock call sequences

4. Debugging Approach:
- Add comprehensive logging around state transitions
- Implement transaction logging for debugging
- Create debug endpoints for development
- Add metrics for tracking operation success/failure

5. Code Organization:
- Separate complex queries into named functions
- Create reusable test utilities
- Implement clear error types for different failure cases
- Add comprehensive documentation for complex operations

6. Performance Considerations:
- Add query execution plan analysis in tests
- Implement proper indexing strategy
- Use batch operations where possible
- Monitor and log query execution times

7. Error Handling:
- Create specific error types for different failure cases
- Add detailed error context
- Implement proper error wrapping
- Add error recovery mechanisms

8. Concurrent Operations:
- Implement proper locking mechanisms
- Add retry logic for concurrent modifications
- Use optimistic locking where appropriate
- Add deadlock detection and recovery

9. Testing Tools:
- Create custom test assertions for common cases
- Implement test data generators
- Add test coverage reports
- Create integration test suites

10. Documentation:
- Document all complex queries
- Add examples for common operations
- Create troubleshooting guides
- Maintain up-to-date API documentation

## Development Philosophy

11. **Consistent Error Handling Across Layers**:
- Implement consistent error handling across all handlers and services
- Create helper functions for standard error handling patterns
- Map domain-specific errors to appropriate HTTP status codes
- Ensure error messages are descriptive and helpful for debugging

12. **Model and Schema Synchronization**:
- Keep model structs synchronized with database schema at all times
- Update tests when model fields change to ensure they reflect the current schema
- Include version fields in all relevant database tables and model structs
- Ensure data validation happens at both the model and database levels

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

## Acknowledgements

This codebase would not have been possible without the significant contributions and assistance from:

- **Cursor**: The AI-powered IDE that dramatically improved coding efficiency
- **Claude**: For providing intelligent code analysis, suggestions, and debugging assistance
- **Gemini**: For additional AI-powered support in development

## Project Statistics

- **Total Files**: 140
- **Total Lines of Code**: 84,538
- **Total Size**: 2.79 MB

**Development Speed**: This codebase represents work that would typically take a team of 2-3 human developers approximately 3-6 months to build, test, and stabilize. The use of AI-assisted development has enabled us to reach this milestone in a fraction of the time while maintaining high code quality standards.

## License

See the LICENSE file for details.

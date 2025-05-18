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

## Current Status

🎉 **Green Across the Board!** 🎉

We have successfully achieved a green status across all our GitHub Actions CI/CD workflows:
- Backend tests are passing
- Frontend tests are passing
- Linting checks are clean (with a few shadow variable warnings remaining)

This is a significant milestone that gives us a solid foundation to build upon. The codebase is now in a state where we can continuously iterate and add new features while maintaining stability.

### Backend Status

- **Test Coverage**: 73.4% overall, with many critical components at 80%+ coverage
- **API Functionality**: Core Tribe, List, and Activity operations are implemented
- **Database Layer**: Robust repository implementations with proper error handling

### Frontend Status

- **Mobile App**: Basic UI structure with screens for Home, Lists, Interest Buttons, and Profile
- **Test Coverage**: Basic tests in place, but coverage is low (7.01%)
- **UI Components**: Using React Native Paper for consistent design

## Known Issues

### Remaining Linting Issues
- **`govet shadow` errors**: The linter has identified 46 instances of variable shadowing across the backend codebase, primarily in repository layer files.
  - **Resolution Plan**: Systematically resolve shadowing issues in each file, starting with highest impact files.

### Test Coverage Gaps
- Some repository methods have skipped tests due to transaction-related challenges:
  - `TestListRepository_GetUserLists/test_repository_method`
  - `TestListRepository_GetTribeLists/test_repository_method`
  - `TestListRepository_GetSharedLists`
  - `TestListRepository_ShareWithTribe/update_existing`
  - `TestListRepository_GetListShares/multiple_versions`

## Future Work

### Backend Development Priorities

1. **Resolve Shadowing Linting Issues**
   - Fix all 46 identified `govet shadow` errors
   - Implement a style guide for variable naming to prevent future shadowing issues

2. **Improve Test Coverage**
   - Fix skipped tests in List Repository
   - Increase service layer test coverage (currently at 78.6%)
   - Add tests for sync-related models (currently 0% coverage)

3. **Enhance Database Operations**
   - Add transaction boundaries for multi-step operations
   - Improve error handling and reporting
   - Add database constraints (unique tribe names, etc.)

4. **API Enhancements**
   - Implement comprehensive request validation
   - Add consistent error handling patterns
   - Improve API documentation

5. **Advanced Features**
   - Implement per-tribe display names
   - Enhance list synchronization with external sources (Google Maps)
   - Support for advanced interest indicators and scheduling

### Frontend Development Priorities

1. **API Integration**
   - Implement service layer to interact with backend API
   - Add authentication flow (Google account integration)
   - Set up proper state management

2. **UI/UX Improvements**
   - Enhance mobile-first design
   - Add responsive layouts for desktop testing
   - Implement proper navigation and transitions

3. **Test Coverage**
   - Increase test coverage to at least 80%
   - Add integration tests with mock API
   - Implement E2E testing

4. **Feature Implementation**
   - Create and manage tribes
   - Share lists with tribes
   - Interest button functionality
   - Activity scheduling and tracking

5. **Production Readiness**
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

## User Interface Design

1. **Display Names in Tribes**:
- Users must have at least one display name (not NULL)
- In the future, support per-tribe display names:
  - Users may want different display names in different tribes (e.g., real name in family tribe, gamertag in gaming tribe)
  - Some communities know people by different names (IG/TikTok usernames, nicknames, etc.)
  - Initially implement a single "default" display name
  - Plan to add tribe-specific display name overrides in future iterations

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
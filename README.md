# Tribe - Social Activity Planning App

Tribe is a social activity planning application designed to help groups of people (tribes) share and plan activities together. Whether it's couples planning date nights, roommates coordinating house activities, or friend groups organizing outings, Tribe makes it easy to collaborate on and track shared activities.

## Core Features

- **Tribes**: Groups of people who share activities and plans together
- **Lists**: Shared collections of activities, places, or plans
- **Location Integration**: Sync with Google Maps lists for place suggestions
- **Availability Management**: Track when tribe members are free/interested in activities
- **Smart Suggestions**: Get activity recommendations based on history and preferences
- **Federation** (Future): Support for federated instances allowing users to connect across different Tribe clusters while maintaining control over data sharing and privacy

## Project Structure

- `apps/`
  - `mobile/`: React Native mobile application
  - `web/`: Web application frontend
- `backend/`: Go backend service
  - `api/`: API definitions and handlers
  - `internal/`: Core business logic and models
  - `migrations/`: Database schema migrations
- `infrastructure/`: Deployment and infrastructure configuration
- `shared/`: Shared types and utilities
- `docs/`: Project documentation

## Development Setup

1. Clone the repository
2. Install dependencies:
   ```bash
   make setup
   ```
3. Set up environment variables (see .env.example)
4. Start the development environment:
   ```bash
   make dev
   ```

## Development Guidance

### Database Schema Changes
1. **Early Stage Flexibility**: During early development, prefer modifying the initial schema directly rather than creating new migrations. This is appropriate while we don't have production data to maintain.
   - This guideline exists to prevent inefficient stacking of migrations and infinite fix-verify loops
   - Modifying the initial schema is faster and cleaner during early development
   - The development guidelines here will be adjusted as needed based on observed problems or performance issues
2. **Schema First**: Focus on getting the schema right in the initial migration rather than accumulating migrations early.
3. **When to Use Migrations**: Only introduce new migrations when:
   - We have production data that needs to be preserved
   - The schema change is part of a new feature iteration
   - The change requires careful data transformation

### Error Handling
1. **Always Prefer Graceful Handling**: Never use panics for error handling. All failure modes should be gracefully handled with proper error returns and cleanup.
2. **Comprehensive Error States**: Design functions to handle all possible error states gracefully, including:
   - Database connection failures
   - Constraint violations
   - Concurrent access conflicts
   - Resource cleanup failures

### Test-Driven Development
1. **Coverage Requirements**: Maintain at least 90% test coverage, ideally 95% or more.
2. **Parallel Problem Resolution**: Address test failures and issues using a breadth-first approach rather than depth-first:
   - Identify all related issues before starting fixes
   - Fix common root causes that affect multiple tests
   - Avoid getting stuck in fix-verify loops for individual tests
3. **Debugging Practices**:
   - Add console logs to surface problems
   - Print data structures to verify properties/fields/columns
   - Use descriptive logging statements for error tracking
4. **Test Design**:
   - Tests should verify graceful error handling, not expect panics
   - Focus on behavior verification rather than implementation details
   - Include positive and negative test cases
   - Test cleanup and resource management

## Testing

Run tests with:
```bash
make test
```

Ensure test coverage remains above 90% for all new code.

## Known Issues

1. Database Constraint Violations:
   - [ ] `user_id` NOT NULL constraint violation in activities table
   - [ ] `owner_id` NOT NULL constraint violation in lists table
   - [ ] `metadata` NOT NULL constraint violation in tribes table
   - [ ] Foreign key constraint violation in lists table for owner_id
   - [ ] Missing `list_owners` table

2. Test Coverage Issues:
   - [ ] Low coverage in internal/config package (0.0%)
   - [x] Improve coverage in internal/models package (71.6% coverage achieved)
   - [x] Improve coverage in internal/middleware package (87.8% coverage achieved)
   - [ ] Low coverage in internal/testutil package (65.5%)
   - [ ] Low coverage in internal/api/service package (69.6%)
   - [ ] Build failures in postgres repository package

3. Test Failures:
   - [ ] Activity handler tests failing with panic
   - [ ] List service sharing tests failing
   - [ ] Sync service test failures:
     - Invalid state transitions not handled correctly
     - Concurrent operations not properly managed
     - External errors not properly propagated
     - Conflict resolution logic issues

## Future Work

1. Database Improvements:
   - [ ] Add proper indexes for performance optimization
   - [ ] Add proper foreign key constraints for data integrity
   - [ ] Add proper cascading deletes
   - [ ] Add proper default values for required fields
   - [ ] Add proper enum type constraints
   - [ ] Create missing tables (list_owners)

2. Test Improvements:
   - [ ] Fix activity handler test panics
   - [ ] Fix list service sharing tests
   - [ ] Fix sync service state transition tests
   - [ ] Fix sync service concurrent operation tests
   - [ ] Fix sync service external error handling
   - [ ] Fix sync service conflict resolution
   - [ ] Improve test coverage in internal/config package
   - [ ] Improve test coverage in internal/testutil package
   - [ ] Improve test coverage in internal/api/service package
   - [ ] Fix postgres repository build failures
   - [ ] Add integration tests for sync functionality
   - [ ] Add performance tests for database operations
   - [ ] Add stress tests for concurrent operations
   - [ ] Add end-to-end tests for critical user flows
   - [ ] Add more model validation tests
   - [ ] Add handler validation tests
   - [ ] Add configuration validation tests
   - [ ] Add database connection retry tests
   - [ ] Add server shutdown tests
   - [ ] Add middleware error handling tests
   - [ ] Fix transaction management tests
   - [ ] Improve test data generation
   - [ ] Add proper test database role setup

3. Code Quality:
   - [ ] Improve error handling consistency
   - [ ] Add logging for better debugging
   - [ ] Add metrics for monitoring
   - [ ] Add tracing for request flows
   - [ ] Add validation for all models
   - [ ] Add consistent error types and messages
   - [ ] Add graceful shutdown handling
   - [ ] Add health check endpoints
   - [ ] Add rate limiting
   - [ ] Add request logging
   - [ ] Add panic recovery middleware
   - [ ] Add request ID tracking
   - [ ] Implement API versioning

4. Documentation:
   - [ ] Add API documentation
   - [ ] Add deployment guide
   - [ ] Add development setup guide
   - [ ] Add troubleshooting guide
   - [ ] Add model validation documentation
   - [ ] Add handler validation documentation
   - [ ] Add configuration documentation
   - [ ] Add monitoring and metrics documentation
   - [ ] Add API versioning documentation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details. 
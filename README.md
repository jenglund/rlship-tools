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

## Testing

Run tests with:
```bash
make test
```

Ensure test coverage remains above 90% for all new code.

## Known Issues

1. Database Constraint Violations:
   - [x] JSONMap implementation complete (Scan and Value methods implemented)
   - [x] Metadata type standardized to use JSONMap
   - [x] Database constraints for nullable fields added
   - [x] Default values for boolean and numeric fields added
   - [x] Type mismatches between models and database layer (UUID and OwnerType)
   - [x] Type mismatch between ListConflict and SyncConflict in service and repository layers
   - [ ] `user_id` NOT NULL constraint violation in activities table
   - [ ] `owner_id` NOT NULL constraint violation in lists table
   - [ ] `metadata` NOT NULL constraint violation in tribes table
   - [ ] Foreign key constraint violation in lists table for owner_id
   - [ ] Missing `list_owners` table
   - [ ] Invalid enum values being passed for visibility_type and tribe_type

2. Test Coverage Issues:
   - [x] Low coverage in cmd/api package (added comprehensive tests for configuration, database, auth, and server initialization)
   - [x] Low coverage in cmd/migrate package (83.8% coverage achieved)
   - [ ] Low coverage in internal/config package (0.0%)
   - [ ] Low coverage in internal/testutil package (6.8%)
   - [x] Improve coverage in internal/models package (activity validation tests added)
   - [x] Improve coverage in internal/middleware package (added tests for repository errors, middleware chaining, and Firebase auth initialization)
   - [ ] Overall low coverage in postgres repository package (20.8%)

3. Transaction Management Issues:
   - [ ] Role "user" does not exist for transaction tests
   - [ ] Deadlock retry mechanism not properly tested
   - [ ] Lock timeout retry mechanism not properly tested

4. Test Infrastructure Issues:
   - [ ] Test data generation failing due to foreign key constraints
   - [ ] Inconsistent test database cleanup
   - [ ] Missing test setup for required database roles

## Future Work

1. Database Improvements:
   - [ ] Squash migrations and recreate clean migration history
   - [ ] Add proper indexes for performance optimization
   - [ ] Add foreign key constraints for data integrity
   - [ ] Add proper cascading deletes
   - [ ] Add proper default values for required fields
   - [ ] Add proper enum type constraints
   - [ ] Create missing tables (list_owners)
   - [ ] Review and fix all NOT NULL constraints

2. Test Improvements:
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
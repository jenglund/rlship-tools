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

1. Type System and Model Issues:
   
   a. Type Consistency Problems:
   - [x] JSONMap implementation complete (Scan and Value methods implemented)
   - [x] Metadata type standardized to use JSONMap
   - [x] Database constraints for nullable fields added
   - [x] Default values for boolean and numeric fields added
   - [x] Type mismatches between models and database layer (UUID and OwnerType)
   
   b. Validation Layer Issues:
   - [x] Cross-field validations added (location data, seasonal dates, guest membership)
   - [x] Consistent validation between service and repository layers
   - [x] Incomplete validation for sync-related fields
   - [ ] No validation for maximum field lengths or content format
   
   c. Model Relationship Issues:
   - [x] Foreign key constraints properly handled with cascade rules
   - [x] Clear ownership cascade rules for related entities
   - [x] Proper cleanup handling for related records
   - [x] Soft-delete propagation implemented with triggers and constraints

   d. Data Integrity Issues:
   - [x] List sync status validation failures
   - [x] Consistent timestamp handling across models
   - [x] Transaction boundaries implemented for multi-table operations
   - [x] Optimistic locking implemented for concurrent modifications
   - [x] Proper transaction management for concurrent operations

2. Sync System Issues:
   
   a. Error Handling:
   - [x] Specific error types for sync-related issues
   - [x] Improved error messages for better debugging
   - [x] Proper error handling in handlers with appropriate HTTP status codes
   - [x] Error type checking utilities

   b. Testing Coverage:
   - [x] Comprehensive sync service tests
   - [x] Tests for edge cases in sync transitions
   - [x] Tests for conflict resolution scenarios

   c. Documentation:
   - [x] Detailed documentation about sync workflows
   - [x] Documentation for conflict resolution strategies

## Future Work

1. Test Infrastructure Improvements:
   - [ ] Implement comprehensive test coverage for cmd/* packages
   - [ ] Improve middleware test coverage to exceed 90%
   - [x] Create robust test data setup/teardown infrastructure
   - [x] Standardize mock usage patterns and expectations
   - [x] Add integration tests for database constraints
   - [x] Add tests for soft delete propagation and cascade rules
   - [ ] Add tests for transaction boundaries and rollbacks
   - [ ] Add tests for optimistic locking and concurrent modifications

2. Code Quality and Safety:
   - [x] Standardize error handling patterns across all layers
   - [x] Implement proper null safety checks in handlers
   - [x] Align model validation with database constraints
   - [ ] Add comprehensive input validation
   - [x] Add transaction boundaries for multi-table operations
   - [x] Implement optimistic locking for concurrent modifications
   - [ ] Add deadlock detection and prevention
   - [x] Implement retry logic for transient failures
   - [ ] Add audit logging for critical operations

3. Documentation and Standards:
   - [x] Document error handling patterns
   - [x] Create mock usage guidelines
   - [ ] Update API documentation with validation rules
   - [x] Document test data setup procedures
   - [x] Create contribution guidelines for tests
   - [x] Document cascade rules and soft delete behavior
   - [x] Document transaction boundaries and isolation levels
   - [x] Document optimistic locking and version handling

4. Infrastructure and Tooling:
   - [ ] Add automated test coverage checks
   - [ ] Implement database migration tests
   - [x] Add linting for consistent error handling
   - [x] Create test data generation tools
   - [ ] Add performance benchmarks
   - [ ] Add monitoring for database constraint violations
   - [ ] Add monitoring for transaction performance
   - [ ] Add monitoring for deadlocks and lock contention
   - [ ] Add monitoring for optimistic locking failures

5. Federation Support (Future):
   - Design federation protocol for inter-instance communication
   - Implement selective federation controls
     - Allow instance owners to control which other instances they federate with
     - Support granular control over what data/connections are shared
     - Enable federation transitivity controls (who can share federation connections)
   - Add federation-aware data models
     - Support distributed unique identifiers
     - Handle data ownership across instances
     - Implement conflict resolution for federated data
   - Implement privacy controls
     - Allow users to control data visibility across federation
     - Support instance-level privacy policies
     - Enable selective sharing of tribe connections
   - Add federation management tools
     - Federation status monitoring
     - Connection health checks
     - Federation audit logs
   - Design security measures
     - Instance verification
     - Secure communication channels
     - Rate limiting and abuse prevention
   - Create federation documentation
     - Protocol specifications
     - Implementation guidelines
     - Security best practices
     - Privacy considerations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details. 
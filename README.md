# Tribe - Social Activity Planning App

Tribe is a social activity planning application designed to help groups of people (tribes) share and plan activities together. Whether it's couples planning date nights, roommates coordinating house activities, or friend groups organizing outings, Tribe makes it easy to collaborate on and track shared activities.

## Core Features

- **Tribes**: Groups of people who share activities and plans together
- **Lists**: Shared collections of activities, places, or plans
- **Location Integration**: Sync with Google Maps lists for place suggestions
- **Availability Management**: Track when tribe members are free/interested in activities
- **Smart Suggestions**: Get activity recommendations based on history and preferences

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
   - [ ] Type mismatches between models and database layer (UUID and OwnerType)
   
   b. Validation Layer Issues:
   - [x] Cross-field validations added (location data, seasonal dates, guest membership)
   - [ ] Inconsistent validation between service and repository layers
   - [ ] Incomplete validation for sync-related fields
   - [ ] No validation for maximum field lengths or content format
   
   c. Model Relationship Issues:
   - [x] Foreign key constraints properly handled with cascade rules
   - [x] Clear ownership cascade rules for related entities
   - [x] Proper cleanup handling for related records
   - [x] Soft-delete propagation implemented with triggers and constraints

   d. Data Integrity Issues:
   - [ ] List sync status validation failures
   - [ ] Inconsistent timestamp handling across models
   - [ ] Missing transaction boundaries for multi-table operations
   - [ ] No handling of concurrent modifications

Plan of Attack for Type System and Model Issues:

1. Type System Cleanup (Priority: High)
   - [x] Create comprehensive type mapping documentation
   - [x] Implement proper JSONMap scanning/value methods
   - [x] Standardize nullable field handling
   - [ ] Add strong typing for all enum-like fields

2. Validation Enhancement (Priority: High)
   - [x] Implement cross-field validation helpers
   - [x] Add database constraint validations
   - [ ] Create validation middleware for API layer
   - [ ] Add field length and format validations

3. Model Relationship Refactoring (Priority: Medium)
   - [x] Document entity relationship diagrams
   - [x] Implement proper cascade rules
   - [x] Add relationship integrity checks
   - [x] Create cleanup utilities for testing

4. Data Integrity Implementation (Priority: High)
   - [ ] Add optimistic locking for concurrent modifications
   - [ ] Implement proper transaction boundaries
   - [ ] Create sync status state machine
   - [ ] Add audit logging for critical operations

Next Steps:
1. [x] Implement database enum types and constraints for string-based enums (VisibilityType, OwnerType, etc.)
2. [x] Standardize nullable field handling across all models
3. [x] Add proper foreign key constraints and cascade rules
4. [ ] Implement transaction boundaries for multi-table operations

## Future Work

1. Test Infrastructure Improvements:
   - Implement comprehensive test coverage for cmd/* packages
   - Improve middleware test coverage to exceed 90%
   - Create robust test data setup/teardown infrastructure
   - Standardize mock usage patterns and expectations
   - Add integration tests for database constraints
   - Add tests for soft delete propagation and cascade rules

2. Code Quality and Safety:
   - Standardize error handling patterns across all layers
   - Implement proper null safety checks in handlers
   - Align model validation with database constraints
   - Add comprehensive input validation
   - Add transaction boundaries for multi-table operations
   - Implement optimistic locking for concurrent modifications

3. Documentation and Standards:
   - Document error handling patterns
   - Create mock usage guidelines
   - Update API documentation with validation rules
   - Document test data setup procedures
   - Create contribution guidelines for tests
   - Document cascade rules and soft delete behavior

4. Infrastructure and Tooling:
   - Add automated test coverage checks
   - Implement database migration tests
   - Add linting for consistent error handling
   - Create test data generation tools
   - Add performance benchmarks
   - Add monitoring for database constraint violations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details. 
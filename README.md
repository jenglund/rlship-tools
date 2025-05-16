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
   - Type mismatches between models and database layer (UUID and OwnerType)
   - Inconsistent validation between service and repository layers
   - List sync status validation failures
   - Foreign key constraints not properly handled in test data setup

2. Test Coverage and Quality:
   - Zero coverage in cmd/api and cmd/migrate packages
   - Low coverage (78.6%) in middleware package
   - Poor coverage (31.5%) in testutil package
   - Mock expectations not properly aligned with actual service behavior
   - Test data setup/teardown infrastructure issues

3. Error Handling and Safety:
   - Inconsistent error handling patterns across layers
   - Nil pointer safety issues in activity handler
   - Improper error propagation in service layer
   - Inconsistent error type matching in tests

4. Database and Infrastructure:
   - Foreign key constraint violations in test data
   - Missing proper database constraint handling
   - Undefined repository methods (e.g., MarkItemChosen)
   - Test database setup/cleanup issues

## Future Work

1. Test Infrastructure Improvements:
   - Implement comprehensive test coverage for cmd/* packages
   - Improve middleware test coverage to exceed 90%
   - Create robust test data setup/teardown infrastructure
   - Standardize mock usage patterns and expectations
   - Add integration tests for database constraints

2. Code Quality and Safety:
   - Standardize error handling patterns across all layers
   - Implement proper null safety checks in handlers
   - Align model validation with database constraints
   - Review and update foreign key relationships
   - Add comprehensive input validation

3. Documentation and Standards:
   - Document error handling patterns
   - Create mock usage guidelines
   - Update API documentation with validation rules
   - Document test data setup procedures
   - Create contribution guidelines for tests

4. Infrastructure and Tooling:
   - Add automated test coverage checks
   - Implement database migration tests
   - Add linting for consistent error handling
   - Create test data generation tools
   - Add performance benchmarks

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details. 
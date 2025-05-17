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
   - [x] Type mismatch between ListConflict and SyncConflict in service and repository layers
   
   b. Validation Layer Issues:
   - [x] Cross-field validations added (location data, seasonal dates, guest membership)
   - [x] Consistent validation between service and repository layers
   - [x] Incomplete validation for sync-related fields
   - [x] Type assertion needed for Metadata in activity handlers
   - [x] Invalid operation on JSONMap in activity_photos_test.go
   - [x] UUID array type mismatch in list_test.go
   - [x] OwnerType string to pointer conversion in list_test.go
   - [x] Activity validation added for required fields (user_id, visibility, etc.)
   - [x] Tribe metadata handling added for tribe creation
   - [x] List owner foreign key validation added
   
   c. Database Migration Issues:
   - [ ] Migration failure: metadata column does not exist
   - [ ] JSONB type conversion issues
   - [ ] Missing NOT NULL constraints
   - [ ] Missing check constraints for JSON objects

2. Test Coverage Issues:
   - [ ] Low coverage in cmd/api package (0.0%)
   - [ ] Low coverage in cmd/migrate package (0.0%)
   - [ ] Low coverage in internal/config package (0.0%)
   - [ ] Low coverage in internal/testutil package (6.8%)
   - [x] Improve coverage in internal/models package (activity validation tests added)
   - [ ] Improve coverage in internal/middleware package (78.6%)

## Future Work

1. Database Improvements:
   - [ ] Squash migrations and recreate clean migration history
   - [ ] Add proper indexes for performance optimization
   - [ ] Add foreign key constraints for data integrity
   - [ ] Add proper cascading deletes

2. Test Improvements:
   - [ ] Add integration tests for sync functionality
   - [ ] Add performance tests for database operations
   - [ ] Add stress tests for concurrent operations
   - [ ] Add end-to-end tests for critical user flows
   - [ ] Add more model validation tests
   - [ ] Add handler validation tests

3. Code Quality:
   - [ ] Improve error handling consistency
   - [ ] Add logging for better debugging
   - [ ] Add metrics for monitoring
   - [ ] Add tracing for request flows
   - [ ] Add validation for all models
   - [ ] Add consistent error types and messages

4. Documentation:
   - [ ] Add API documentation
   - [ ] Add deployment guide
   - [ ] Add development setup guide
   - [ ] Add troubleshooting guide
   - [ ] Add model validation documentation
   - [ ] Add handler validation documentation

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests
5. Submit a pull request

## License

See [LICENSE](LICENSE) file for details. 